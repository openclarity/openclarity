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

package scan

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cli "github.com/openclarity/vmclarity/cli"
	"github.com/openclarity/vmclarity/cli/cmd/logutil"

	"github.com/openclarity/vmclarity/cli/state"

	"github.com/openclarity/vmclarity/cli/presenter"

	apiclient "github.com/openclarity/vmclarity/api/client"
	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/scanner/families"
)

const (
	DefaultWatcherInterval = 2 * time.Minute
	DefaultMountTimeout    = 10 * time.Minute
)

// ScanCmd represents the scan command.
var ScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan",
	Long:  `Run scanner families`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logutil.Logger.Infof("Running...")

		// Main context which remains active even if the scan is aborted allowing post-processing operations
		// like updating asset scan state
		ctx := log.SetLoggerForContext(cmd.Context(), logutil.Logger)

		cfgFile, err := cmd.Flags().GetString("config")
		if err != nil {
			logutil.Logger.Fatalf("Unable to get config file name: %v", err)
		}
		server, err := cmd.Flags().GetString("server")
		if err != nil {
			logutil.Logger.Fatalf("Unable to get VMClarity server address: %v", err)
		}
		output, err := cmd.Flags().GetString("output")
		if err != nil {
			logutil.Logger.Fatalf("Unable to get output file name: %v", err)
		}
		assetScanID, err := cmd.Flags().GetString("asset-scan-id")
		if err != nil {
			logutil.Logger.Fatalf("Unable to get asset scan ID: %v", err)
		}
		mountVolume, err := cmd.Flags().GetBool("mount-attached-volume")
		if err != nil {
			logutil.Logger.Fatalf("Unable to get mount attached volume flag: %v", err)
		}

		config := loadConfig(cfgFile)
		cli, err := newCli(config, server, assetScanID, output)
		if err != nil {
			return fmt.Errorf("failed to initialize CLI: %w", err)
		}

		// Create context used to signal to operations that the scan is aborted
		abortCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		// Start watching for abort event
		cli.WatchForAbort(ctx, cancel, DefaultWatcherInterval)

		if err := cli.WaitForReadyState(abortCtx); err != nil {
			err = fmt.Errorf("failed to wait for AssetScan being ready to scan: %w", err)
			if e := cli.MarkFailed(ctx, err.Error()); e != nil {
				logutil.Logger.Errorf("Failed to update AssetScan status to failed: %v", e)
			}
			return err
		}

		if mountVolume {
			// Set timeout for mounting volumes
			mountCtx, mountCancel := context.WithTimeout(abortCtx, DefaultMountTimeout)
			defer mountCancel()

			mountPoints, err := cli.MountVolumes(mountCtx)
			if err != nil {
				err = fmt.Errorf("failed to mount attached volume: %w", err)
				if e := cli.MarkFailed(ctx, err.Error()); e != nil {
					logutil.Logger.Errorf("Failed to update asset scan stat to failed: %v", e)
				}
				return err
			}
			families.SetMountPointsForFamiliesInput(mountPoints, config)
		}

		err = cli.MarkInProgress(ctx, config)
		if err != nil {
			return fmt.Errorf("failed to inform server %v scan has started: %w", server, err)
		}

		logutil.Logger.Infof("Running scanners...")
		runErrors := families.New(config).Run(abortCtx, cli)

		if len(runErrors) > 0 {
			logutil.Logger.Errorf("Errors when running families: %+v", runErrors)
			err := cli.MarkFailed(ctx, errors.Join(runErrors...).Error())
			if err != nil {
				return fmt.Errorf("failed to inform the server %v that scan failed: %w", server, err)
			}
			return nil
		}

		err = cli.MarkDone(ctx)
		if err != nil {
			return fmt.Errorf("failed to inform the server %v the scan was completed: %w", server, err)
		}

		return nil
	},
}

// nolint: gochecknoinits
func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	ScanCmd.Flags().String("config", "", "config file (default is $HOME/.vmclarity.yaml)")
	ScanCmd.Flags().String("output", "", "set output directory path. Stdout is used if not set.")
	ScanCmd.Flags().String("server", "", "VMClarity server to export asset scans to, for example: http://localhost:9999/api")
	ScanCmd.Flags().String("asset-scan-id", "", "the AssetScan ID to monitor and report results to")
	ScanCmd.Flags().Bool("mount-attached-volume", false, "discover for an attached volume and mount it before the scan")

	// TODO(sambetts) we may have to change this to our own validation when
	// we add the CI/CD scenario and there isn't an existing asset-scan-id
	// in the backend to PATCH
	ScanCmd.MarkFlagsRequiredTogether("server", "asset-scan-id")
}

// loadConfig reads in config file and ENV variables if set.
func loadConfig(cfgFile string) *families.Config {
	logutil.Logger.Infof("Initializing configuration...")
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
	config := &families.Config{}
	err = viper.Unmarshal(config)
	cobra.CheckErr(err)

	if logrus.IsLevelEnabled(logrus.InfoLevel) {
		configB, err := yaml.Marshal(config)
		cobra.CheckErr(err)
		logutil.Logger.Infof("Using config file (%s):\n%s", viper.ConfigFileUsed(), string(configB))
	}

	return config
}

func newCli(config *families.Config, server, assetScanID, output string) (*cli.CLI, error) {
	var manager state.Manager
	var presenters []presenter.Presenter
	var err error

	if config == nil {
		return nil, errors.New("families config must not be nil")
	}

	if server != "" {
		var client *apiclient.Client
		var p presenter.Presenter

		client, err = apiclient.New(server)
		if err != nil {
			return nil, fmt.Errorf("failed to create VMClarity API client: %w", err)
		}

		manager, err = state.NewVMClarityState(client, assetScanID)
		if err != nil {
			return nil, fmt.Errorf("failed to create VMClarity state manager: %w", err)
		}

		p, err = presenter.NewVMClarityPresenter(client, assetScanID)
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
