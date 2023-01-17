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

package secrets

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
	"github.com/openclarity/kubeclarity/shared/pkg/utils"

	"github.com/openclarity/vmclarity/shared/pkg/families/interfaces"
	familiesresults "github.com/openclarity/vmclarity/shared/pkg/families/results"
	"github.com/openclarity/vmclarity/shared/pkg/families/secrets/common"
	"github.com/openclarity/vmclarity/shared/pkg/families/secrets/job"
)

type Secrets struct {
	conf   Config
	logger *log.Entry
}

func (s Secrets) Run(res *familiesresults.Results) (interfaces.IsResults, error) {
	s.logger.Info("Secrets Run...")

	manager := job_manager.New(s.conf.ScannersList, s.conf.ScannersConfig, s.logger, job.Factory)
	mergedResults := NewMergedResults()

	for _, input := range s.conf.Inputs {
		results, err := manager.Run(utils.SourceType(input.InputType), input.Input)
		if err != nil {
			return nil, fmt.Errorf("failed to scan input %q for secrets: %v", input.Input, err)
		}

		// Merge results.
		for name, result := range results {
			s.logger.Infof("Merging result from %q", name)
			mergedResults = mergedResults.Merge(result.(*common.Results)) // nolint:forcetypeassert
		}
	}

	s.logger.Info("Secrets Done...")
	return &Results{
		MergedResults: mergedResults,
	}, nil
}

// ensure types implement the requisite interfaces.
var _ interfaces.Family = &Secrets{}

func New(logger *log.Entry, conf Config) *Secrets {
	return &Secrets{
		conf:   conf,
		logger: logger.Dup().WithField("family", "secrets"),
	}
}
