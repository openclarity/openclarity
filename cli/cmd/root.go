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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ghodss/yaml"
	kubeclarityutils "github.com/openclarity/kubeclarity/shared/pkg/utils"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/cli/pkg"
	"github.com/openclarity/vmclarity/cli/pkg/mount"
	"github.com/openclarity/vmclarity/shared/pkg/families"
	"github.com/openclarity/vmclarity/shared/pkg/families/exploits"
	"github.com/openclarity/vmclarity/shared/pkg/families/malware"
	misconfigurationTypes "github.com/openclarity/vmclarity/shared/pkg/families/misconfiguration/types"
	"github.com/openclarity/vmclarity/shared/pkg/families/results"
	"github.com/openclarity/vmclarity/shared/pkg/families/sbom"
	"github.com/openclarity/vmclarity/shared/pkg/families/secrets"
	"github.com/openclarity/vmclarity/shared/pkg/families/vulnerabilities"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

var (
	cfgFile string
	config  *families.Config
	logger  *logrus.Entry
	output  string

	server       string
	scanResultID string
	mountVolume  bool
)

const (
	fsTypeExt4 = "ext4"
	fsTypeXFS  = "xfs"
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

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var exporter *Exporter
		if server != "" {
			exp, err := CreateExporter()
			if err != nil {
				return fmt.Errorf("failed to create a result exporter: %w", err)
			}
			exporter = exp
		}

		if mountVolume && exporter != nil {
			// wait for volume to be attached.
			if err := waitForAttached(ctx, exporter); err != nil {
				return fmt.Errorf("failed to wait for volume attached: %v", err)
			}
			logger.Infof("got volume attached state")

			mountPoints, err := mountAttachedVolume()
			if err != nil {
				return fmt.Errorf("failed to mount attached volume: %v", err)
			}
			setMountPointsForFamiliesInput(mountPoints, config)
		}

		if exporter != nil {
			// TODO ideally we want to mark the state of each family, and not just general.
			err := exporter.MarkScanResultInProgress(ctx)
			if err != nil {
				return fmt.Errorf("failed to inform server %v scan has started: %w", server, err)
			}
		}

		res, famerr := families.New(logger, config).Run()

		if exporter != nil {
			logger.Infof("Exporting results to the backend...")
			var generalErrors []error

			exportErrors := exporter.ExportResults(ctx, res, famerr)
			generalErrors = append(generalErrors, exportErrors...)

			if len(famerr) > 0 {
				generalErrors = append(generalErrors, fmt.Errorf("at least one family failed to run"))
			}

			err := exporter.MarkScanResultDone(ctx, generalErrors)
			if err != nil {
				return fmt.Errorf("failed to inform the server %v the scan was completed: %w", server, err)
			}
		}

		if len(famerr) > 0 {
			return fmt.Errorf("failed to run families: %+v", famerr)
		}

		if err := outputResults(config, res); err != nil {
			return fmt.Errorf("failed to output results: %v", err)
		}

		return nil
	},
}

func waitForAttached(ctx context.Context, exporter *Exporter) error {
	if exporter == nil {
		return errors.New("the Exporter parameter must not be nil")
	}

	// nolint:govet
	ctx, cancel := context.WithTimeout(ctx, utils.DefaultResourceReadyWaitTimeoutMin*time.Minute)
	defer cancel()

	for {
		select {
		case <-time.After(utils.DefaultResourceReadyCheckIntervalSec * time.Second):
			status, err := exporter.client.GetScanResultStatus(ctx, scanResultID)
			if err != nil {
				return fmt.Errorf("failed to get scan result status: %v", err)
			}
			// wait for status attached (meaning volume was attached and can be mounted).
			if *status.General.State == models.ATTACHED {
				return nil
			}
		case <-ctx.Done():
			return fmt.Errorf("waiting for volume ready was canceled: %v", ctx.Err())
		}
	}
}

// nolint:cyclop
func outputResults(config *families.Config, res *results.Results) error {
	if config.SBOM.Enabled {
		if err := outputSBOMResults(config, res); err != nil {
			return err
		}
	}

	if config.Vulnerabilities.Enabled {
		if err := outputVulnerabilitiesResults(res); err != nil {
			return err
		}
	}

	if config.Secrets.Enabled {
		if err := outputSecretsResults(res); err != nil {
			return err
		}
	}

	if config.Exploits.Enabled {
		if err := outputExploitsResults(res); err != nil {
			return err
		}
	}

	if config.Malware.Enabled {
		if err := outputMalwareResults(res); err != nil {
			return err
		}
	}

	return nil
}

func outputSBOMResults(config *families.Config, res *results.Results) error {
	sbomResults, err := results.GetResult[*sbom.Results](res)
	if err != nil {
		return fmt.Errorf("failed to get sbom results: %v", err)
	}

	outputFormat := config.SBOM.AnalyzersConfig.Analyzer.OutputFormat
	sbomBytes, err := sbomResults.EncodeToBytes(outputFormat)
	if err != nil {
		return fmt.Errorf("failed to encode sbom results to bytes: %w", err)
	}

	// TODO: Need to implement a better presenter
	err = Output(sbomBytes, "sbom")
	if err != nil {
		return fmt.Errorf("failed to output sbom results: %v", err)
	}
	return nil
}

func outputVulnerabilitiesResults(res *results.Results) error {
	vulnerabilitiesResults, err := results.GetResult[*vulnerabilities.Results](res)
	if err != nil {
		return fmt.Errorf("failed to get sbom results: %v", err)
	}

	bytes, err := json.Marshal(vulnerabilitiesResults.MergedResults)
	if err != nil {
		return fmt.Errorf("failed to output vulnerabilities results: %v", err)
	}
	err = Output(bytes, "vulnerabilities")
	if err != nil {
		return fmt.Errorf("failed to output vulnerabilities results: %v", err)
	}
	return nil
}

func outputSecretsResults(res *results.Results) error {
	secretsResults, err := results.GetResult[*secrets.Results](res)
	if err != nil {
		return fmt.Errorf("failed to get secrets results: %v", err)
	}

	bytes, err := json.Marshal(secretsResults)
	if err != nil {
		return fmt.Errorf("failed to output secrets results: %v", err)
	}
	err = Output(bytes, "secrets")
	if err != nil {
		return fmt.Errorf("failed to output secrets results: %v", err)
	}
	return nil
}

func outputExploitsResults(res *results.Results) error {
	exploitsResults, err := results.GetResult[*exploits.Results](res)
	if err != nil {
		return fmt.Errorf("failed to get exploits results: %v", err)
	}

	bytes, err := json.Marshal(exploitsResults)
	if err != nil {
		return fmt.Errorf("failed to marshal exploits results: %v", err)
	}
	err = Output(bytes, "exploits")
	if err != nil {
		return fmt.Errorf("failed to output exploits results: %v", err)
	}
	return nil
}

func outputMalwareResults(res *results.Results) error {
	malwareResults, err := results.GetResult[*malware.MergedResults](res)
	if err != nil {
		return fmt.Errorf("failed to get malware results: %v", err)
	}

	bytes, err := json.Marshal(malwareResults)
	if err != nil {
		return fmt.Errorf("failed to marshal malware results: %v", err)
	}
	err = Output(bytes, "malware")
	if err != nil {
		return fmt.Errorf("failed to output  %v", err)
	}
	return nil
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
	rootCmd.PersistentFlags().StringVar(&output, "output", "", "set file path output (default: stdout)")
	rootCmd.PersistentFlags().StringVar(&server, "server", "", "VMClarity server to export scan results to, for example: http://localhost:9999/api")
	rootCmd.PersistentFlags().StringVar(&scanResultID, "scan-result-id", "", "the ScanResult ID to export the scan results to")
	rootCmd.PersistentFlags().BoolVar(&mountVolume, "mount-attached-volume", false, "discover for an attached volume and mount it before the scan")

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

func Output(bytes []byte, outputPrefix string) error {
	if output == "" {
		os.Stdout.Write([]byte(fmt.Sprintf("%s results:\n", outputPrefix)))
		os.Stdout.Write(bytes)
		os.Stdout.Write([]byte("\n=================================================\n"))
		return nil
	}

	filePath := outputPrefix + "." + output
	logger.Infof("Writing results to %v...", filePath)

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666) // nolint:gomnd,gofumpt
	if err != nil {
		return fmt.Errorf("failed open file %s: %v", filePath, err)
	}
	defer file.Close()

	_, err = file.Write(bytes)
	if err != nil {
		return fmt.Errorf("failed to write bytes to file %s: %v", filePath, err)
	}

	return nil
}

func isSupportedFS(fs string) bool {
	switch fs {
	case fsTypeExt4, fsTypeXFS:
		return true
	}
	return false
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
				Input:     mountDir,
				InputType: string(kubeclarityutils.ROOTFS),
			})
		}
		if familiesConfig.Malware.Enabled {
			familiesConfig.Malware.Inputs = append(familiesConfig.Malware.Inputs, malware.Input{
				Input:     mountDir,
				InputType: string(kubeclarityutils.ROOTFS),
			})
		}
		if familiesConfig.Misconfiguration.Enabled {
			familiesConfig.Misconfiguration.Inputs = append(
				familiesConfig.Misconfiguration.Inputs,
				misconfigurationTypes.Input{
					Input:     mountDir,
					InputType: string(kubeclarityutils.ROOTFS),
				},
			)
		}
	}
	return familiesConfig
}

func mountAttachedVolume() ([]string, error) {
	var mountPoints []string

	devices, err := mount.ListBlockDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to list block devices: %v", err)
	}
	for _, device := range devices {
		// if the device is not mounted and of a supported filesystem type,
		// we assume it belongs to the attached volume, so we mount it.
		if device.MountPoint == "" && isSupportedFS(device.FilesystemType) {
			mountDir := "/mnt/snapshot" + uuid.NewV4().String()

			if err := device.Mount(mountDir); err != nil {
				return nil, fmt.Errorf("failed to mount device: %v", err)
			}
			logger.Infof("Mounted device %v on %v", device.DeviceName, mountDir)
			mountPoints = append(mountPoints, mountDir)
		}
	}
	return mountPoints, nil
}
