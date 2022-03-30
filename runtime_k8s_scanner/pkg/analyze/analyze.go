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

package analyze

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"wwwin-github.cisco.com/eti/scan-gazr/runtime_k8s_scanner/pkg/config"
	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/analyzer"
	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/analyzer/job"
	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/job_manager"
	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/utils"
)

type Analyzer interface {
	Analyze(config *config.Config) (*analyzer.MergedResults, error)
}

type AnalyzerImpl struct {
	logger *logrus.Entry
}

func Create(logger *logrus.Entry) Analyzer {
	return &AnalyzerImpl{
		logger: logger.Dup().WithField("component", "analyzer"),
	}
}

func (a *AnalyzerImpl) Analyze(config *config.Config) (*analyzer.MergedResults, error) {
	manager := job_manager.New(config.SharedConfig.Analyzer.AnalyzerList, config.SharedConfig, a.logger,
		job.CreateAnalyzerJob)

	results, err := manager.Run(utils.IMAGE, config.ImageIDToScan)
	if err != nil {
		return nil, fmt.Errorf("failed to run job manager: %v", err)
	}

	// Merge results.
	mergedResults := analyzer.NewMergedResults(utils.IMAGE, config.ImageHashToScan)
	for name, result := range results {
		a.logger.Infof("Merging result from %q", name)
		mergedResults = mergedResults.Merge(result.(*analyzer.Results), config.SharedConfig.Analyzer.OutputFormat)
	}

	return mergedResults, nil
}
