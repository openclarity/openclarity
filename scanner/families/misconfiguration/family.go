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

package misconfiguration

import (
	"context"
	"fmt"
	"time"

	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/scanner/families/interfaces"
	"github.com/openclarity/vmclarity/scanner/families/misconfiguration/job"
	misconfigurationTypes "github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
	"github.com/openclarity/vmclarity/scanner/families/results"
	"github.com/openclarity/vmclarity/scanner/families/types"
	familiesutils "github.com/openclarity/vmclarity/scanner/families/utils"
	"github.com/openclarity/vmclarity/scanner/job_manager"
	"github.com/openclarity/vmclarity/scanner/utils"
)

type Misconfiguration struct {
	conf misconfigurationTypes.Config
}

func (m Misconfiguration) Run(ctx context.Context, _ *results.Results) (interfaces.IsResults, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithField("family", "misconfiguration")
	logger.Info("Misconfiguration Run...")

	misConfigResults := NewResults()

	manager := job_manager.New(m.conf.ScannersList, m.conf.ScannersConfig, logger, job.Factory)
	for _, input := range m.conf.Inputs {
		startTime := time.Now()
		managerResults, err := manager.Run(ctx, utils.SourceType(input.InputType), input.Input)
		if err != nil {
			return nil, fmt.Errorf("failed to scan input %q for misconfigurations: %w", input.Input, err)
		}
		endTime := time.Now()
		inputSize, err := familiesutils.GetInputSize(input)
		if err != nil {
			logger.Warnf("Failed to calculate input %v size: %v", input, err)
		}

		// Merge results.
		for name, result := range managerResults {
			logger.Infof("Merging result from %q", name)
			if assetScan, ok := result.(misconfigurationTypes.ScannerResult); ok {
				if familiesutils.ShouldStripInputPath(input.StripPathFromResult, m.conf.StripInputPaths) {
					assetScan = StripPathFromResult(assetScan, input.Input)
				}
				misConfigResults.AddScannerResult(assetScan)
			} else {
				return nil, fmt.Errorf("received bad scanner result type %T, expected misconfigurationTypes.ScannerResult", result)
			}
		}
		misConfigResults.Metadata.InputScans = append(misConfigResults.Metadata.InputScans, types.CreateInputScanMetadata(startTime, endTime, inputSize, input))
	}

	logger.Info("Misconfiguration Done...")

	return misConfigResults, nil
}

// StripPathFromResult strip input path from results wherever it is found.
func StripPathFromResult(result misconfigurationTypes.ScannerResult, path string) misconfigurationTypes.ScannerResult {
	for i := range result.Misconfigurations {
		result.Misconfigurations[i].Location = familiesutils.TrimMountPath(result.Misconfigurations[i].Location, path)
	}
	return result
}

func (m Misconfiguration) GetType() types.FamilyType {
	return types.Misconfiguration
}

// ensure types implement the requisite interfaces.
var _ interfaces.Family = &Misconfiguration{}

func New(conf misconfigurationTypes.Config) *Misconfiguration {
	return &Misconfiguration{
		conf: conf,
	}
}
