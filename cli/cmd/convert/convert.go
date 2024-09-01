// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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

package convert

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/openclarity/openclarity/cli/cmd/logutil"
	"github.com/openclarity/openclarity/cli/presenter"
	"github.com/openclarity/openclarity/scanner/utils/converter"
)

var ConvertCmd = &cobra.Command{
	Use:    "convert",
	Short:  "Convert SBOM",
	Long:   `Currently, only supports converting SBOM from cyclondx format to syft-json format`,
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		logutil.Logger.Infof("Converting SBOM...")

		outputFormat, err := cmd.Flags().GetString("output-format")
		if err != nil {
			logutil.Logger.Fatalf("Unable to get output format: %v", err)
		}

		inputSBOMFile, err := cmd.Flags().GetString("input-sboms")
		if err != nil {
			logutil.Logger.Fatalf("Unable to get input sbom: %v", err)
		}

		var outputPath string
		outputPath, err = cmd.Flags().GetString("output-path")
		if err != nil {
			outputPath, err = os.Getwd()
			if err != nil {
				outputPath = "."
			}
		}

		outputSBOMFile, err := cmd.Flags().GetString("output-sbom")
		if err != nil {
			logutil.Logger.Fatalf("Unable to get output sbom: %v", err)
		}

		outputSbomFormat, err := converter.StringToSbomFormat(outputFormat)
		if err != nil {
			logutil.Logger.Fatalf("Unsupported SBOM conversion: %v", err)
		}

		bomBytes, err := convertSBOMFile(inputSBOMFile, outputSbomFormat)
		if err != nil {
			logutil.Logger.Fatalf("Failed to convert SBOM: %v", err)
		}

		writer := presenter.FileWriter{Path: outputPath}
		err = writer.Write(bomBytes, outputSBOMFile)
		if err != nil {
			logutil.Logger.Fatalf("Failed to write SBOM: %v", err)
		}
	},
}

// nolint: gochecknoinits
func init() {
	ConvertCmd.Flags().StringP("input-format", "", "cyclonedx", "Deprecated and ignored")
	ConvertCmd.Flags().StringP("input-sbom", "i", "", "Define input filename")
	ConvertCmd.Flags().StringP("output-path", "p", "", "Define output path")
	ConvertCmd.Flags().StringP("output-format", "", "json", "Define output format")
	ConvertCmd.Flags().StringP("output-sbom", "o", "", "Define output filename")
}

func convertSBOMFile(inputSBOMFile string, outputFormat converter.SbomFormat) ([]byte, error) {
	cdxBom, err := converter.GetCycloneDXSBOMFromFile(inputSBOMFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get CycloneDX SBOM from file: %w", err)
	}

	bomBytes, err := converter.CycloneDxToBytes(cdxBom, outputFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to convert CycloneDX to %v format: %w", outputFormat, err)
	}

	return bomBytes, nil
}
