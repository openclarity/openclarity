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
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/openclarity/kubeclarity/cli/pkg/utils"
	"github.com/openclarity/kubeclarity/shared/pkg/converter"
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
	convertCmd.Flags().StringP("input-format", "", "cyclonedx", "Deprecated and ignored")
	convertCmd.Flags().StringP("input-sboms", "i", "", "Define input filename")
	convertCmd.Flags().StringP("output-format", "", "json", "Define output format")
	convertCmd.Flags().StringP("output-sbom", "o", "", "Define output filename")
}

func convertSBOM(cmd *cobra.Command, _ []string) {
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

	outputSbomFormat, err := converter.StringToSbomFormat(outputFormat)
	if err != nil {
		logger.Fatalf("Unsupported SBOM conversion: %v", err)
	}

	if err = ConvertSBOMFile(inputSBOMFile, outputSBOMFile, outputSbomFormat); err != nil {
		logger.Fatalf("Failed to convert: %v", err)
	}
}

func ConvertSBOMFile(inputSBOMFile string, outputSBOMFile string, outputFormat converter.SbomFormat) error {
	cdxBom, err := converter.GetCycloneDXSBOMFromFile(inputSBOMFile)
	if err != nil {
		return fmt.Errorf("failed to get CycloneDX SBOM from file: %v", err)
	}

	bomBytes, err := converter.CycloneDxToBytes(cdxBom, outputFormat)
	if err != nil {
		return fmt.Errorf("failed to convert CycloneDX to %v format: %v", outputFormat, err)
	}

	if err := utils.WriteSBOM(bomBytes, outputSBOMFile); err != nil {
		return fmt.Errorf("failed to write SBOM to file %s: %v", outputSBOMFile, err)
	}

	return nil
}
