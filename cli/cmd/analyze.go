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

	"github.com/spf13/cobra"

	"wwwin-github.cisco.com/eti/scan-gazr/cli/pkg"
	_export "wwwin-github.cisco.com/eti/scan-gazr/cli/pkg/analyzer/export"
	"wwwin-github.cisco.com/eti/scan-gazr/cli/pkg/utils"
	sharedanalyzer "wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/analyzer"
	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/analyzer/job"
	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/formatter"
	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/job_manager"
	sharedutils "wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/utils"
)

// analyzeCmd represents the analyze command.
var analyzeCmd = &cobra.Command{
	Use:   "analyze [SOURCE]",
	Short: "Content analyzer",
	Long:  `KubeClarity content analyzer.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("an image/directory argument is required")
		}

		return cobra.MaximumNArgs(1)(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		analyzeContent(cmd, args)
	},
}

// nolint: gochecknoinits
func init() {
	rootCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().StringP("output", "o", "",
		"set output (default: stdout)")
	analyzeCmd.Flags().StringP("input-type", "i", "",
		fmt.Sprintf("set input type (input type can be %s,%s,%s default:%s)",
			sharedutils.DIR, sharedutils.FILE, sharedutils.IMAGE, sharedutils.IMAGE))
	analyzeCmd.Flags().String("application-id", "",
		"ID of a defined application to associate the exported analysis")
	analyzeCmd.Flags().BoolP("export", "e", false,
		"export analysis results to the backend")
}

// nolint:cyclop
func analyzeContent(cmd *cobra.Command, args []string) {
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		logger.Fatalf("Unable to get output filename: %v", err)
	}

	inputType, err := cmd.Flags().GetString("input-type")
	if err != nil {
		logger.Fatalf("Unable to get input type: %v", err)
	}

	sourceType, err := sharedutils.ValidateInputType(inputType)
	if err != nil {
		logger.Fatalf("Failed to validate input type: %v", err)
	}

	export, err := cmd.Flags().GetBool("export")
	if err != nil {
		logger.Fatalf("Unable to get export flag: %v", err)
	}

	appID, err := cmd.Flags().GetString("application-id")
	if err != nil {
		logger.Fatalf("Unable to get application ID: %v", err)
	}

	manager := job_manager.New(appConfig.SharedConfig.Analyzer.AnalyzerList, appConfig.SharedConfig, logger, job.CreateAnalyzerJob)
	src := utils.SetSource(appConfig.LocalImageScan, sourceType, args[0])
	results, err := manager.Run(sourceType, src)
	if err != nil {
		logger.Fatalf("Failed to run job manager: %v", err)
	}

	hash, err := utils.GenerateHash(sourceType, src)
	if err != nil {
		logger.Fatalf("Failed to generate hash for source %s: %v", src, err)
	}

	// Merge results
	mergedResults := sharedanalyzer.NewMergedResults(sourceType, hash)
	outputFormat := appConfig.SharedConfig.Analyzer.OutputFormat

	for _, result := range results {
		mergedResults = mergedResults.Merge(result.(*sharedanalyzer.Results), outputFormat)
	}

	mergedSboms, err := mergedResults.CreateMergedSBOMBytes(outputFormat, pkg.GitRevision)
	if err != nil {
		logger.Fatalf("Failed to create merged output: %v", err)
	}

	if err := formatter.WriteSBOM(mergedSboms, output); err != nil {
		logger.Fatalf("Failed to write results to file %v: %v ", output, err)
	}

	if export {
		logger.Infof("Exporting analysis results to the backend: %s", appConfig.Backend.Host)
		apiClient := utils.NewHTTPClient(appConfig.Backend)
		// TODO generate application ID
		if err := _export.Export(apiClient, mergedResults, appID); err != nil {
			logger.Errorf("Failed to export analysis results to the backend: %v", err)
		}
	}
}
