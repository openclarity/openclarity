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

package scan

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/cisco-open/kubei/runtime_k8s_scanner/pkg/config"
	"github.com/cisco-open/kubei/shared/pkg/job_manager"
	sharedscanner "github.com/cisco-open/kubei/shared/pkg/scanner"
	"github.com/cisco-open/kubei/shared/pkg/scanner/job"
	"github.com/cisco-open/kubei/shared/pkg/utils"
)

type Scanner interface {
	Scan(config *config.Config, sbomFilePath string, src sharedscanner.Source) (*sharedscanner.MergedResults, error)
}

type ScannerImpl struct {
	logger *logrus.Entry
}

func Create(logger *logrus.Entry) *ScannerImpl {
	return &ScannerImpl{
		logger: logger.Dup().WithField("component", "scanner"),
	}
}

func (s *ScannerImpl) Scan(config *config.Config, sbomFilePath string) (*sharedscanner.MergedResults, error) {
	manager := job_manager.New(config.SharedConfig.Scanner.ScannersList, config.SharedConfig, s.logger, job.CreateJob)
	results, err := manager.Run(utils.SBOM, sbomFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to run job manager: %v", err)
	}

	// Merge results.
	mergedResults := sharedscanner.NewMergedResults()
	for name, result := range results {
		s.logger.Infof("Merging result from %q", name)
		mergedResults = mergedResults.Merge(result.(*sharedscanner.Results))
	}

	// Set source values.
	mergedResults.SetSource(sharedscanner.Source{
		Type: "image",
		Name: config.ImageIDToScan,
		Hash: config.ImageHashToScan,
	})

	return mergedResults, nil
}
