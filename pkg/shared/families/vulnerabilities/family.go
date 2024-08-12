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

package vulnerabilities

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/openclarity/kubeclarity/shared/pkg/config"
	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
	sharedscanner "github.com/openclarity/kubeclarity/shared/pkg/scanner"
	"github.com/openclarity/kubeclarity/shared/pkg/scanner/job"
	"github.com/openclarity/kubeclarity/shared/pkg/utils"

	"github.com/openclarity/vmclarity/pkg/shared/families/interfaces"
	"github.com/openclarity/vmclarity/pkg/shared/families/results"
	"github.com/openclarity/vmclarity/pkg/shared/families/sbom"
	"github.com/openclarity/vmclarity/pkg/shared/families/types"
	familiesutils "github.com/openclarity/vmclarity/pkg/shared/families/utils"
	"github.com/openclarity/vmclarity/pkg/shared/log"
)

const (
	sbomTempFilePath = "/tmp/sbom"
)

type Vulnerabilities struct {
	conf           Config
	ScannersConfig config.Config
}

func (v Vulnerabilities) Run(ctx context.Context, res *results.Results) (interfaces.IsResults, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithField("family", "vulnerabilities")
	logger.Info("Vulnerabilities Run...")

	manager := job_manager.New(v.conf.ScannersList, v.conf.ScannersConfig, logger, job.Factory)
	mergedResults := sharedscanner.NewMergedResults()

	if v.conf.InputFromSbom {
		logger.Infof("Using input from SBOM results")

		sbomResults, err := results.GetResult[*sbom.Results](res)
		if err != nil {
			return nil, fmt.Errorf("failed to get sbom results: %w", err)
		}

		sbomBytes, err := sbomResults.EncodeToBytes("cyclonedx-json")
		if err != nil {
			return nil, fmt.Errorf("failed to encode sbom results to bytes: %w", err)
		}

		// TODO: need to avoid writing sbom to file
		if err := os.WriteFile(sbomTempFilePath, sbomBytes, 0o600 /* read & write */); err != nil { // nolint:gomnd,gofumpt
			return nil, fmt.Errorf("failed to write sbom to file: %w", err)
		}

		v.conf.Inputs = append(v.conf.Inputs, types.Input{
			Input:     sbomTempFilePath,
			InputType: "sbom",
		})
	}

	if len(v.conf.Inputs) == 0 {
		return nil, fmt.Errorf("inputs list is empty")
	}

	var vulResults Results
	for _, input := range v.conf.Inputs {
		startTime := time.Now()
		runResults, err := manager.Run(utils.SourceType(input.InputType), input.Input)
		if err != nil {
			return nil, fmt.Errorf("failed to run for input %v of type %v: %w", input.Input, input.InputType, err)
		}
		endTime := time.Now()
		inputSize, err := familiesutils.GetInputSize(input)
		if err != nil {
			logger.Warnf("Failed to calculate input %v size: %v", input, err)
		}

		// Merge results.
		for name, result := range runResults {
			logger.Infof("Merging result from %q", name)
			mergedResults = mergedResults.Merge(result.(*sharedscanner.Results)) // nolint:forcetypeassert
		}
		vulResults.Metadata.InputScans = append(vulResults.Metadata.InputScans, types.CreateInputScanMetadata(startTime, endTime, inputSize, input))

		// TODO:
		// // Set source values.
		// mergedResults.SetSource(sharedscanner.Source{
		//	Type: "image",
		//	Name: config.ImageIDToScan,
		//	Hash: config.ImageHashToScan,
		// })
	}

	logger.Info("Vulnerabilities Done...")

	vulResults.MergedResults = mergedResults
	return &vulResults, nil
}

func (v Vulnerabilities) GetType() types.FamilyType {
	return types.Vulnerabilities
}

// ensure types implement the requisite interfaces.
var _ interfaces.Family = &Vulnerabilities{}

func New(conf Config) *Vulnerabilities {
	return &Vulnerabilities{
		conf: conf,
	}
}
