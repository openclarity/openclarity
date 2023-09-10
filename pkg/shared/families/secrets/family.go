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

package secrets

import (
	"context"
	"fmt"
	"time"

	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
	"github.com/openclarity/kubeclarity/shared/pkg/utils"

	"github.com/openclarity/vmclarity/pkg/shared/families/interfaces"
	familiesresults "github.com/openclarity/vmclarity/pkg/shared/families/results"
	"github.com/openclarity/vmclarity/pkg/shared/families/secrets/common"
	"github.com/openclarity/vmclarity/pkg/shared/families/secrets/job"
	"github.com/openclarity/vmclarity/pkg/shared/families/types"
	familiesutils "github.com/openclarity/vmclarity/pkg/shared/families/utils"
	"github.com/openclarity/vmclarity/pkg/shared/log"
)

type Secrets struct {
	conf Config
}

func (s Secrets) Run(ctx context.Context, _ *familiesresults.Results) (interfaces.IsResults, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithField("family", "secrets")
	logger.Info("Secrets Run...")

	manager := job_manager.New(s.conf.ScannersList, s.conf.ScannersConfig, logger, job.Factory)
	mergedResults := NewMergedResults()

	var secretsResults Results
	for _, input := range s.conf.Inputs {
		startTime := time.Now()
		results, err := manager.Run(utils.SourceType(input.InputType), input.Input)
		if err != nil {
			return nil, fmt.Errorf("failed to scan input %q for secrets: %w", input.Input, err)
		}
		endTime := time.Now()
		inputSize, err := familiesutils.GetInputSize(input)
		if err != nil {
			logger.Warnf("Failed to calculate input %v size: %v", input, err)
		}

		// Merge results.
		for name, result := range results {
			secretResult := result.(*common.Results) // nolint:forcetypeassert
			if familiesutils.ShouldStripInputPath(input.StripPathFromResult, s.conf.StripInputPaths) {
				secretResult = StripPathFromResult(secretResult, input.Input)
			}
			logger.Infof("Merging result from %q", name)
			mergedResults = mergedResults.Merge(secretResult)
		}
		secretsResults.Metadata.InputScans = append(secretsResults.Metadata.InputScans, types.CreateInputScanMetadata(startTime, endTime, inputSize, input))
	}

	logger.Info("Secrets Done...")
	secretsResults.MergedResults = mergedResults
	return &secretsResults, nil
}

// StripPathFromResult strip input path from results wherever it is found.
func StripPathFromResult(result *common.Results, path string) *common.Results {
	for i := range result.Findings {
		result.Findings[i].File = familiesutils.TrimMountPath(result.Findings[i].File, path)
		result.Findings[i].Fingerprint = familiesutils.RemoveMountPathSubStringIfNeeded(result.Findings[i].Fingerprint, path)
	}
	return result
}

func (s Secrets) GetType() types.FamilyType {
	return types.Secrets
}

// ensure types implement the requisite interfaces.
var _ interfaces.Family = &Secrets{}

func New(conf Config) *Secrets {
	return &Secrets{
		conf: conf,
	}
}
