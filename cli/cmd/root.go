// Copyright © 2023 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ghodss/yaml"
	kubeclarityutils "github.com/openclarity/kubeclarity/shared/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/openclarity/vmclarity/cli/pkg"
	"github.com/openclarity/vmclarity/cli/pkg/cli"
	"github.com/openclarity/vmclarity/cli/pkg/presenter"
	"github.com/openclarity/vmclarity/cli/pkg/state"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
	"github.com/openclarity/vmclarity/shared/pkg/backendclient"
	"github.com/openclarity/vmclarity/shared/pkg/families"
	"github.com/openclarity/vmclarity/shared/pkg/families/malware"
	misconfigurationTypes "github.com/openclarity/vmclarity/shared/pkg/families/misconfiguration/types"
	"github.com/openclarity/vmclarity/shared/pkg/families/rootkits"
	"github.com/openclarity/vmclarity/shared/pkg/families/sbom"
	"github.com/openclarity/vmclarity/shared/pkg/families/secrets"
	"github.com/openclarity/vmclarity/shared/pkg/families/vulnerabilities"
)

const DefaultWatcherInterval = 2 * time.Minute

var (
	cfgFile string
	config  *families.Config
	logger  *logrus.Entry
	output  string

	server                string
	scanResultID          string
	mountVolume           bool
	waitForServerAttached bool
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:          "vmclarity",
	Short:        "VMClarity",
	Long:         `VMClarity`,
	Version:      pkg.GitRevision,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Infof("Running...")

		// Main context which remains active even if the scan is aborted allowing post-processing operations
		// like updating scan result state
		ctx := cmd.Context()

		cli, err := newCli()
		if err != nil {
			return fmt.Errorf("failed to initialize CLI: %w", err)
		}

		// Create context used to signal to operations that the scan is aborted
		abortCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		// Start watching for abort event
		cli.WatchForAbort(ctx, cancel, DefaultWatcherInterval)

		if waitForServerAttached {
			if err := cli.WaitForVolumeAttachment(abortCtx); err != nil {
				err = fmt.Errorf("failed to wait for block device being attached: %w", err)
				if e := cli.MarkDone(ctx, []error{err}); e != nil {
					logger.Errorf("Failed to update scan result stat to completed with errors: %v", e)
				}
				return err
			}
		}

		if mountVolume {
			mountPoints, err := cli.MountVolumes(abortCtx)
			if err != nil {
				err = fmt.Errorf("failed to mount attached volume: %w", err)
				if e := cli.MarkDone(ctx, []error{err}); e != nil {
					logger.Errorf("Failed to update scan result stat to completed with errors: %v", e)
				}
				return err
			}
			setMountPointsForFamiliesInput(mountPoints, config)
		}

		err = cli.MarkInProgress(ctx)
		if err != nil {
			return fmt.Errorf("failed to inform server %v scan has started: %w", server, err)
		}

		logger.Infof("Running scanners...")
		res, familiesErr := families.New(logger, config).Run(abortCtx)

		logger.Infof("Exporting results...")
		errs := cli.ExportResults(abortCtx, res, familiesErr)

		if len(familiesErr) > 0 {
			errs = append(errs, fmt.Errorf("at least one family failed to run"))
		}

		err = cli.MarkDone(ctx, errs)
		if err != nil {
			return fmt.Errorf("failed to inform the server %v the scan was completed: %w", server, err)
		}

		if len(familiesErr) > 0 {
			return fmt.Errorf("failed to run families: %+v", familiesErr)
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

// nolint: gochecknoinits
func init() {
	cobra.OnInitialize(
		initLogger,
		initConfig,
	)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.vmclarity.yaml)")
	rootCmd.PersistentFlags().StringVar(&output, "output", "", "set output directory path. Stdout is used if not set.")
	rootCmd.PersistentFlags().StringVar(&server, "server", "", "VMClarity server to export scan results to, for example: http://localhost:9999/api")
	rootCmd.PersistentFlags().StringVar(&scanResultID, "scan-result-id", "", "the ScanResult ID to export the scan results to")
	rootCmd.PersistentFlags().BoolVar(&mountVolume, "mount-attached-volume", false, "discover for an attached volume and mount it before the scan")
	rootCmd.PersistentFlags().BoolVar(&waitForServerAttached, "wait-for-server-attached", false, "wait for the VMClarity server to attach the volume")

	// TODO(sambetts) we may have to change this to our own validation when
	// we add the CI/CD scenario and there isn't an existing scan-result-id
	// in the backend to PATCH
	rootCmd.MarkFlagsRequiredTogether("server", "scan-result-id")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	logrus.Infof("init config")
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory OR current directory with name ".families" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".families")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	cobra.CheckErr(err)

	// Load config
	config = &families.Config{}
	err = viper.Unmarshal(config)
	cobra.CheckErr(err)

	if logrus.IsLevelEnabled(logrus.InfoLevel) {
		configB, err := yaml.Marshal(config)
		cobra.CheckErr(err)
		logrus.Infof("Using config file (%s):\n%s", viper.ConfigFileUsed(), string(configB))
	}
}

func initLogger() {
	log := logrus.New()
	log.SetLevel(logrus.InfoLevel)
	logger = log.WithField("app", "vmclarity")
}

func newCli() (*cli.CLI, error) {
	var manager state.Manager
	var presenters []presenter.Presenter
	var err error

	if config == nil {
		return nil, errors.New("families config must not be nil")
	}

	if server != "" {
		var client *backendclient.BackendClient
		var p presenter.Presenter

		client, err = backendclient.Create(server)
		if err != nil {
			return nil, fmt.Errorf("failed to create VMClarity API client: %w", err)
		}

		manager, err = state.NewVMClarityState(client, scanResultID)
		if err != nil {
			return nil, fmt.Errorf("failed to create VMClarity state manager: %w", err)
		}

		p, err = presenter.NewVMClarityPresenter(client, scanResultID)
		if err != nil {
			return nil, fmt.Errorf("failed to create VMClarity presenter: %w", err)
		}
		presenters = append(presenters, p)
	} else {
		manager, err = state.NewLocalState()
		if err != nil {
			return nil, fmt.Errorf("failed to create local state: %w", err)
		}
	}

	if output != "" {
		presenters = append(presenters, presenter.NewFilePresenter(output, config))
	} else {
		presenters = append(presenters, presenter.NewConsolePresenter(os.Stdout, config))
	}

	var p presenter.Presenter
	if len(presenters) == 1 {
		p = presenters[0]
	} else {
		p = &presenter.MultiPresenter{Presenters: presenters}
	}

	return &cli.CLI{Manager: manager, Presenter: p, FamiliesConfig: config}, nil
}

func setMountPointsForFamiliesInput(mountPoints []string, familiesConfig *families.Config) *families.Config {
	// update families inputs with the mount point as rootfs
	for _, mountDir := range mountPoints {
		if familiesConfig.SBOM.Enabled {
			familiesConfig.SBOM.Inputs = append(familiesConfig.SBOM.Inputs, sbom.Input{
				Input:     mountDir,
				InputType: string(kubeclarityutils.ROOTFS),
			})
		}

		if familiesConfig.Vulnerabilities.Enabled {
			if familiesConfig.SBOM.Enabled {
				familiesConfig.Vulnerabilities.InputFromSbom = true
			} else {
				familiesConfig.Vulnerabilities.Inputs = append(familiesConfig.Vulnerabilities.Inputs, vulnerabilities.Input{
					Input:     mountDir,
					InputType: string(kubeclarityutils.ROOTFS),
				})
			}
		}

		if familiesConfig.Secrets.Enabled {
			familiesConfig.Secrets.Inputs = append(familiesConfig.Secrets.Inputs, secrets.Input{
				StripPathFromResult: utils.PointerTo(true),
				Input:               mountDir,
				InputType:           string(kubeclarityutils.ROOTFS),
			})
		}

		if familiesConfig.Malware.Enabled {
			familiesConfig.Malware.Inputs = append(familiesConfig.Malware.Inputs, malware.Input{
				StripPathFromResult: utils.PointerTo(true),
				Input:               mountDir,
				InputType:           string(kubeclarityutils.ROOTFS),
			})
		}

		if familiesConfig.Rootkits.Enabled {
			familiesConfig.Rootkits.Inputs = append(familiesConfig.Rootkits.Inputs, rootkits.Input{
				StripPathFromResult: utils.PointerTo(true),
				Input:               mountDir,
				InputType:           string(kubeclarityutils.ROOTFS),
			})
		}

		if familiesConfig.Misconfiguration.Enabled {
			familiesConfig.Misconfiguration.Inputs = append(
				familiesConfig.Misconfiguration.Inputs,
				misconfigurationTypes.Input{
					StripPathFromResult: utils.PointerTo(true),
					Input:               mountDir,
					InputType:           string(kubeclarityutils.ROOTFS),
				},
			)
		}
	}
	return familiesConfig
}
