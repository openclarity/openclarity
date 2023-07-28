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

package sbom

import (
	"context"
	"fmt"

	"github.com/openclarity/kubeclarity/cli/pkg"
	cliutils "github.com/openclarity/kubeclarity/cli/pkg/utils"
	sharedanalyzer "github.com/openclarity/kubeclarity/shared/pkg/analyzer"
	"github.com/openclarity/kubeclarity/shared/pkg/analyzer/job"
	"github.com/openclarity/kubeclarity/shared/pkg/converter"
	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
	"github.com/openclarity/kubeclarity/shared/pkg/utils"

	"github.com/openclarity/vmclarity/pkg/shared/families/interfaces"
	familiesresults "github.com/openclarity/vmclarity/pkg/shared/families/results"
	"github.com/openclarity/vmclarity/pkg/shared/families/types"
	"github.com/openclarity/vmclarity/pkg/shared/log"
)

type SBOM struct {
	conf Config
}

func (s SBOM) Run(ctx context.Context, _ *familiesresults.Results) (interfaces.IsResults, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithField("family", "sbom")
	logger.Info("SBOM Run...")

	if len(s.conf.Inputs) == 0 {
		return nil, fmt.Errorf("inputs list is empty")
	}

	// TODO: move the logic from cli utils to shared utils
	// TODO: now that we support multiple inputs,
	//  we need to change the fact the the MergedResults assumes it is only for 1 input?
	hash, err := cliutils.GenerateHash(utils.SourceType(s.conf.Inputs[0].InputType), s.conf.Inputs[0].Input)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hash for source %s: %v", s.conf.Inputs[0].Input, err)
	}

	manager := job_manager.New(s.conf.AnalyzersList, s.conf.AnalyzersConfig, logger, job.Factory)
	mergedResults := sharedanalyzer.NewMergedResults(utils.SourceType(s.conf.Inputs[0].InputType), hash)

	for _, input := range s.conf.Inputs {
		results, err := manager.Run(utils.SourceType(input.InputType), input.Input)
		if err != nil {
			return nil, fmt.Errorf("failed to analyzer input %q: %v", s.conf.Inputs[0].Input, err)
		}

		// Merge results.
		for name, result := range results {
			logger.Infof("Merging result from %q", name)
			mergedResults = mergedResults.Merge(result.(*sharedanalyzer.Results)) // nolint:forcetypeassert
		}
	}

	for i, with := range s.conf.MergeWith {
		name := fmt.Sprintf("merge_with_%d", i)
		cdxBOMBytes, err := converter.GetCycloneDXSBOMFromFile(with.SbomPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get CDX SBOM from path=%s: %v", with.SbomPath, err)
		}
		results := sharedanalyzer.CreateResults(cdxBOMBytes, name, with.SbomPath, utils.SBOM)
		logger.Infof("Merging result from %q", with.SbomPath)
		mergedResults = mergedResults.Merge(results)
	}

	// TODO(sambetts) Expose CreateMergedSBOM as well as
	// CreateMergedSBOMBytes so that we don't need to re-convert it
	mergedSBOMBytes, err := mergedResults.CreateMergedSBOMBytes("cyclonedx-json", pkg.GitRevision)
	if err != nil {
		return nil, fmt.Errorf("failed to create merged output: %v", err)
	}

	cdxBom, err := converter.GetCycloneDXSBOMFromBytes(mergedSBOMBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to load merged output to CDX bom: %w", err)
	}

	logger.Info("SBOM Done...")

	return &Results{
		SBOM: cdxBom,
	}, nil
}

func (s SBOM) GetType() types.FamilyType {
	return types.SBOM
}

// ensure types implement the requisite interfaces.
var _ interfaces.Family = &SBOM{}

func New(conf Config) *SBOM {
	return &SBOM{
		conf: conf,
	}
}
