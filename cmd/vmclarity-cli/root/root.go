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
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/openclarity/vmclarity/cmd/vmclarity-cli/asset"
	"github.com/openclarity/vmclarity/cmd/vmclarity-cli/logutil"
	"github.com/openclarity/vmclarity/cmd/vmclarity-cli/scan"
	"github.com/openclarity/vmclarity/pkg/shared/log"
	"github.com/openclarity/vmclarity/pkg/version"
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
	cobra.CheckErr(rootCmd.Execute())
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
