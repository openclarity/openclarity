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

package infofinder

import (
	"context"
	"fmt"
	"time"

	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/scanner/families/infofinder/job"
	infofinderTypes "github.com/openclarity/vmclarity/scanner/families/infofinder/types"
	"github.com/openclarity/vmclarity/scanner/families/interfaces"
	"github.com/openclarity/vmclarity/scanner/families/results"
	"github.com/openclarity/vmclarity/scanner/families/types"
	familiesutils "github.com/openclarity/vmclarity/scanner/families/utils"
	"github.com/openclarity/vmclarity/scanner/job_manager"
	"github.com/openclarity/vmclarity/scanner/utils"
)

type InfoFinder struct {
	conf infofinderTypes.Config
}

func (i InfoFinder) Run(ctx context.Context, _ *results.Results) (interfaces.IsResults, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithField("family", "info finder")
	logger.Info("InfoFinder Run...")

	infoFinderResults := NewResults()

	manager := job_manager.New(i.conf.ScannersList, i.conf.ScannersConfig, logger, job.Factory)
	for _, input := range i.conf.Inputs {
		startTime := time.Now()
		managerResults, err := manager.Run(ctx, utils.SourceType(input.InputType), input.Input)
		if err != nil {
			return nil, fmt.Errorf("failed to scan input %q for info: %w", input.Input, err)
		}
		endTime := time.Now()
		inputSize, err := familiesutils.GetInputSize(input)
		if err != nil {
			logger.Warnf("Failed to calculate input %v size: %v", input, err)
		}

		// Merge results.
		for name, result := range managerResults {
			logger.Infof("Merging result from %q", name)
			if assetScan, ok := result.(*infofinderTypes.ScannerResult); ok {
				if familiesutils.ShouldStripInputPath(input.StripPathFromResult, i.conf.StripInputPaths) {
					assetScan = stripPathFromResult(assetScan, input.Input)
				}
				infoFinderResults.AddScannerResult(assetScan)
			} else {
				return nil, fmt.Errorf("received bad scanner result type %T, expected infofinderTypes.ScannerResult", result)
			}
		}
		infoFinderResults.Metadata.InputScans = append(infoFinderResults.Metadata.InputScans, types.CreateInputScanMetadata(startTime, endTime, inputSize, input))
	}

	logger.Info("InfoFinder Done...")

	return infoFinderResults, nil
}

// stripPathFromResult strip input path from results wherever it is found.
func stripPathFromResult(result *infofinderTypes.ScannerResult, path string) *infofinderTypes.ScannerResult {
	for i := range result.Infos {
		result.Infos[i].Path = familiesutils.TrimMountPath(result.Infos[i].Path, path)
	}
	return result
}

func (i InfoFinder) GetType() types.FamilyType {
	return types.InfoFinder
}

// ensure types implement the requisite interfaces.
var _ interfaces.Family = &InfoFinder{}

func New(conf infofinderTypes.Config) *InfoFinder {
	return &InfoFinder{
		conf: conf,
	}
}
