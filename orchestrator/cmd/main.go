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
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/core/version"
	"github.com/openclarity/vmclarity/orchestrator"
)

const (
	LogLevelFlag         = "log-level"
	LogLevelDefaultValue = "warning"
	ExecutableName       = "vmclarity-orchestrator"
)

var (
	logLevel = LogLevelDefaultValue
	rootCmd  = &cobra.Command{
		Use:   ExecutableName,
		Short: "VMClarity Orchestrator",
		Long:  "VMClarity Orchestrator",
		Version: fmt.Sprintf("Version: %s \nCommit: %s\nBuild Time: %s",
			version.Version, version.CommitHash, version.BuildTimestamp),
		SilenceUsage: true,
	}
)

func init() {
	cmdRun := cobra.Command{
		Use:     "run",
		Run:     runCommand,
		Short:   "Starts the orchestrator",
		Long:    "Starts the VMClarity Orchestrator",
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
		Long:    "Displays the version of the VMClarity Orchestrator",
		Example: ExecutableName + " version",
	}

	rootCmd.AddCommand(&cmdRun)
	rootCmd.AddCommand(&cmdVersion)
}

func main() {
	cobra.CheckErr(rootCmd.Execute())
}

// Main entry point for the orchestrator, triggered by the `run` command.
func runCommand(cmd *cobra.Command, _ []string) {
	ctx, cancel := context.WithCancel(cmd.Context())

	log.InitLogger(logLevel, os.Stderr)
	logger := logrus.WithContext(ctx)
	ctx = log.SetLoggerForContext(ctx, logger)

	config, err := orchestrator.NewConfig()
	if err != nil {
		logger.Fatalf("failed to load Orchestrator config: %v", err)
	}

	err = orchestrator.Run(ctx, config)
	if err != nil {
		logger.Fatalf("failed to run Orchestrator: %v", err)
	}

	// Wait for deactivation
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	s := <-sig
	cancel()
	logger.Warningf("Received a termination signal: %v", s)
}

// Command to display the version.
func versionCommand(_ *cobra.Command, _ []string) {
	fmt.Printf("Version: %s \nCommit: %s\nBuild Time: %s",
		version.Version, version.CommitHash, version.BuildTimestamp)
}
