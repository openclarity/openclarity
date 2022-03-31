// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cisco-open/kubei/shared/pkg/converter"
	"github.com/cisco-open/kubei/shared/pkg/formatter"
)

// convertCmd represents the convert command.
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert SBOM",
	Long:  `Currently, only supports converting SBOM from cyclondx format to syft-json format`,
	Run: func(cmd *cobra.Command, args []string) {
		convertSBOM(cmd, args)
	},
	Hidden: true,
}

// nolint: gochecknoinits
func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.Flags().StringP("input-format", "", "cyclonedx", "Define input format")
	convertCmd.Flags().StringP("input-sboms", "i", "", "Define input filename")
	convertCmd.Flags().StringP("output-format", "", "json", "Define output format")
	convertCmd.Flags().StringP("output-sbom", "o", "", "Define output filename")
}

func convertSBOM(cmd *cobra.Command, _ []string) {
	inputFormat, err := cmd.Flags().GetString("input-format")
	if err != nil {
		logrus.Fatalf("Unable to get input format: %v", err)
	}

	outputFormat, err := cmd.Flags().GetString("output-format")
	if err != nil {
		logrus.Fatalf("Unable to get output format: %v", err)
	}

	inputSBOMFile, err := cmd.Flags().GetString("input-sboms")
	if err != nil {
		logrus.Fatalf("Unable to get input sbom: %v", err)
	}

	outputSBOMFile, err := cmd.Flags().GetString("output-sbom")
	if err != nil {
		logrus.Fatalf("Unable to get output sbom: %v", err)
	}

	if !isSupportedConversion(inputFormat, outputFormat) {
		logger.Fatalf("Unsupported SBOM conversion from %s to %s", inputFormat, outputFormat)
	}

	if err = converter.ConvertCycloneDXToSyftJSONFromFile(inputSBOMFile, outputSBOMFile); err != nil {
		logger.Fatalf("Failed to convert: %v", err)
	}
}

// check supported conversion.
func isSupportedConversion(inputFormat, outputFormat string) bool {
	// check CycloneDXNameJSON because syft version 0.32.2 can generate cyclondx-json output
	if (inputFormat == formatter.CycloneDXFormat || inputFormat == formatter.CycloneDXJSONFormat) && outputFormat == formatter.SyftFormat {
		return true
	}

	return false
}
