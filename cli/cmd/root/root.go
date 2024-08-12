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

package root

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/openclarity/vmclarity/cli/cmd/asset"
	"github.com/openclarity/vmclarity/cli/cmd/logutil"
	"github.com/openclarity/vmclarity/cli/cmd/scan"
	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/core/version"
)

// RootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:          "vmclarity",
	Short:        "VMClarity",
	Long:         `VMClarity`,
	Version:      version.String(),
	SilenceUsage: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Create os.Signal aware context.Context which will trigger context cancellation
	// upon receiving any of the listed signals.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGHUP, syscall.SIGTERM)
	defer func() {
		cancel()
	}()

	cobra.CheckErr(rootCmd.ExecuteContext(ctx))
}

// nolint: gochecknoinits
func init() {
	cobra.OnInitialize(
		initLogger,
	)
	rootCmd.AddCommand(scan.ScanCmd)
	rootCmd.AddCommand(asset.AssetCreateCmd)
	rootCmd.AddCommand(asset.AssetScanCreateCmd)
}

func initLogger() {
	log.InitLogger(logrus.InfoLevel.String(), os.Stderr)
	logutil.Logger = logrus.WithField("app", "vmclarity")
}
