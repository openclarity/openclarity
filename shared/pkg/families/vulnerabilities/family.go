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

package vulnerabilities

import (
	"fmt"
	"os"

	"github.com/openclarity/kubeclarity/shared/pkg/config"
	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
	sharedscanner "github.com/openclarity/kubeclarity/shared/pkg/scanner"
	"github.com/openclarity/kubeclarity/shared/pkg/scanner/job"
	"github.com/openclarity/kubeclarity/shared/pkg/utils"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/shared/pkg/families/interfaces"
	"github.com/openclarity/vmclarity/shared/pkg/families/results"
	"github.com/openclarity/vmclarity/shared/pkg/families/sbom"
)

const (
	sbomTempFilePath = "/tmp/sbom"
)

type Vulnerabilities struct {
	logger         *log.Entry
	conf           Config
	ScannersConfig config.Config
}

func (v Vulnerabilities) Run(res *results.Results) (interfaces.IsResults, error) {
	v.logger.Info("Vulnerabilities Run...")

	manager := job_manager.New(v.conf.ScannersList, v.conf.ScannersConfig, v.logger, job.Factory)
	mergedResults := sharedscanner.NewMergedResults()

	if v.conf.InputFromSbom {
		v.logger.Infof("Using input from SBOM results")

		sbomResults, err := results.GetResult[*sbom.Results](res)
		if err != nil {
			return nil, fmt.Errorf("failed to get sbom results: %v", err)
		}

		sbomBytes, err := sbomResults.EncodeToBytes("cyclonedx-json")
		if err != nil {
			return nil, fmt.Errorf("failed to encode sbom results to bytes: %w", err)
		}

		// TODO: need to avoid writing sbom to file
		if err := os.WriteFile(sbomTempFilePath, sbomBytes, 0600 /* read & write */); err != nil { // nolint:gomnd,gofumpt
			return nil, fmt.Errorf("failed to write sbom to file: %v", err)
		}

		v.conf.Inputs = append(v.conf.Inputs, Inputs{
			Input:     sbomTempFilePath,
			InputType: "sbom",
		})
	}

	if len(v.conf.Inputs) == 0 {
		return nil, fmt.Errorf("inputs list is empty")
	}

	for _, input := range v.conf.Inputs {
		runResults, err := manager.Run(utils.SourceType(input.InputType), input.Input)
		if err != nil {
			return nil, fmt.Errorf("failed to run for input %v of type %v: %w", input.Input, input.InputType, err)
		}

		// Merge results.
		for name, result := range runResults {
			v.logger.Infof("Merging result from %q", name)
			mergedResults = mergedResults.Merge(result.(*sharedscanner.Results)) // nolint:forcetypeassert
		}

		// TODO:
		//// Set source values.
		//mergedResults.SetSource(sharedscanner.Source{
		//	Type: "image",
		//	Name: config.ImageIDToScan,
		//	Hash: config.ImageHashToScan,
		//})
	}

	v.logger.Info("Vulnerabilities Done...")

	return &Results{
		MergedResults: mergedResults,
	}, nil
}

// ensure types implement the requisite interfaces.
var _ interfaces.Family = &Vulnerabilities{}

func New(logger *log.Entry, conf Config) *Vulnerabilities {
	return &Vulnerabilities{
		logger: logger.Dup().WithField("family", "vulnerabilities"),
		conf:   conf,
	}
}
