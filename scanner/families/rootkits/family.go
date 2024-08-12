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

package rootkits

import (
	"context"
	"fmt"
	"time"

	"github.com/openclarity/vmclarity/core/log"
	familiesinterface "github.com/openclarity/vmclarity/scanner/families/interfaces"
	familiesresults "github.com/openclarity/vmclarity/scanner/families/results"
	"github.com/openclarity/vmclarity/scanner/families/rootkits/common"
	"github.com/openclarity/vmclarity/scanner/families/rootkits/job"
	"github.com/openclarity/vmclarity/scanner/families/types"
	familiesutils "github.com/openclarity/vmclarity/scanner/families/utils"
	"github.com/openclarity/vmclarity/scanner/job_manager"
	"github.com/openclarity/vmclarity/scanner/utils"
)

type Rootkits struct {
	conf Config
}

func (r Rootkits) Run(ctx context.Context, _ *familiesresults.Results) (familiesinterface.IsResults, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithField("family", "rootkits")
	logger.Info("Rootkits Run...")

	manager := job_manager.New(r.conf.ScannersList, r.conf.ScannersConfig, logger, job.Factory)
	mergedResults := NewMergedResults()

	var rootkitsResults Results
	for _, input := range r.conf.Inputs {
		startTime := time.Now()
		results, err := manager.Run(ctx, utils.SourceType(input.InputType), input.Input)
		if err != nil {
			return nil, fmt.Errorf("failed to scan input %q for rootkits: %w", input.Input, err)
		}
		endTime := time.Now()
		inputSize, err := familiesutils.GetInputSize(input)
		if err != nil {
			logger.Warnf("Failed to calculate input %v size: %v", input, err)
		}

		// Merge results.
		for name, result := range results {
			logger.Infof("Merging result from %q", name)
			scannerResult := result.(*common.Results) // nolint:forcetypeassert
			if familiesutils.ShouldStripInputPath(input.StripPathFromResult, r.conf.StripInputPaths) {
				scannerResult = StripPathFromResult(scannerResult, input.Input)
			}
			mergedResults = mergedResults.Merge(scannerResult)
		}
		rootkitsResults.Metadata.InputScans = append(rootkitsResults.Metadata.InputScans, types.CreateInputScanMetadata(startTime, endTime, inputSize, input))
	}

	logger.Info("Rootkits Done...")
	rootkitsResults.MergedResults = mergedResults
	return &rootkitsResults, nil
}

// StripPathFromResult strip input path from results wherever it is found.
func StripPathFromResult(result *common.Results, path string) *common.Results {
	for i := range result.Rootkits {
		result.Rootkits[i].Message = familiesutils.RemoveMountPathSubStringIfNeeded(result.Rootkits[i].Message, path)
	}
	return result
}

func (r Rootkits) GetType() types.FamilyType {
	return types.Rootkits
}

// ensure types implement the requisite interfaces.
var _ familiesinterface.Family = &Rootkits{}

func New(conf Config) *Rootkits {
	return &Rootkits{
		conf: conf,
	}
}
