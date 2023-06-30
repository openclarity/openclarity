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

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/openclarity/vmclarity/backend/pkg/backend"
	"github.com/openclarity/vmclarity/backend/pkg/config"
	databaseTypes "github.com/openclarity/vmclarity/backend/pkg/database/types"
	"github.com/openclarity/vmclarity/backend/pkg/version"
	"github.com/openclarity/vmclarity/shared/pkg/log"
)

const (
	LogLevelFlag         = "log-level"
	LogLevelDefaultValue = "warning"
	ExecutableName       = "vmclarity-backend"
)

var (
	logLevel = LogLevelDefaultValue
	rootCmd  = &cobra.Command{
		Use:   ExecutableName,
		Short: "VMClarity Backend",
		Long:  "VMClarity Backend",
		Version: fmt.Sprintf("Version: %s \nCommit: %s\nBuild Time: %s",
			version.Version, version.CommitHash, version.BuildTimestamp),
		SilenceUsage: true,
	}
)

func init() {
	viper.SetDefault(config.HealthCheckAddress, ":8081")
	viper.SetDefault(config.BackendRestPort, "8888")
	viper.SetDefault(config.DatabaseDriver, databaseTypes.DBDriverTypeLocal)
	viper.SetDefault(config.DisableOrchestrator, "false")
	viper.SetDefault(config.UISitePath, "/app/site")
	viper.AutomaticEnv()

	cmdRun := cobra.Command{
		Use:     "run",
		Run:     runCommand,
		Short:   "Starts the server",
		Long:    "Starts the VMClarity backend server",
		Example: ExecutableName + " run",
	}
	cmdRun.PersistentFlags().StringVar(&logLevel,
		LogLevelFlag,
		LogLevelDefaultValue,
		"Set log level [panic fatal error warning info debug trace]")

	cmdVersion := cobra.Command{
		Use:     "version",
		Run:     versionCommand,
		Short:   "Displays the version",
		Long:    "Displays the version of the VMClarity backend server",
		Example: ExecutableName + " version",
	}

	rootCmd.AddCommand(&cmdRun)
	rootCmd.AddCommand(&cmdVersion)
}

func main() {
	cobra.CheckErr(rootCmd.Execute())
}

// Main entry point for the backend, triggered by the
// `run` command in the CLI.
func runCommand(_ *cobra.Command, _ []string) {
	log.InitLogger(logLevel, os.Stderr)

	ctx := context.Background()
	logger := logrus.WithContext(ctx)
	ctx = log.SetLoggerForContext(ctx, logger)
	backend.Run(ctx)
}

// Command to display the version.
func versionCommand(_ *cobra.Command, _ []string) {
	fmt.Printf("Version: %s \nCommit: %s\nBuild Time: %s",
		version.Version, version.CommitHash, version.BuildTimestamp)
}
