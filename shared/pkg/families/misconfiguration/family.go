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
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
	"github.com/openclarity/kubeclarity/shared/pkg/utils"

	"github.com/openclarity/vmclarity/shared/pkg/families/interfaces"
	"github.com/openclarity/vmclarity/shared/pkg/families/misconfiguration/job"
	misconfigurationTypes "github.com/openclarity/vmclarity/shared/pkg/families/misconfiguration/types"
	"github.com/openclarity/vmclarity/shared/pkg/families/results"
	"github.com/openclarity/vmclarity/shared/pkg/families/types"
)

type Misconfiguration struct {
	conf   misconfigurationTypes.Config
	logger *log.Entry
}

func (m Misconfiguration) Run(res *results.Results) (interfaces.IsResults, error) {
	m.logger.Info("Misconfiguration Run...")

	results := NewResults()

	manager := job_manager.New(m.conf.ScannersList, m.conf.ScannersConfig, m.logger, job.Factory)
	for _, input := range m.conf.Inputs {
		managerResults, err := manager.Run(utils.SourceType(input.InputType), input.Input)
		if err != nil {
			return nil, fmt.Errorf("failed to scan input %q for misconfigurations: %v", input.Input, err)
		}

		// Merge results.
		for name, result := range managerResults {
			m.logger.Infof("Merging result from %q", name)
			if scanResult, ok := result.(misconfigurationTypes.ScannerResult); ok {
				results.AddScannerResult(scanResult)
			} else {
				return nil, fmt.Errorf("received bad scanner result type %T, expected misconfigurationTypes.ScannerResult", result)
			}
		}
	}

	m.logger.Info("Misconfiguration Done...")

	return results, nil
}

func (m Misconfiguration) GetType() types.FamilyType {
	return types.Misconfiguration
}

// ensure types implement the requisite interfaces.
var _ interfaces.Family = &Misconfiguration{}

func New(logger *log.Entry, conf misconfigurationTypes.Config) *Misconfiguration {
	return &Misconfiguration{
		conf:   conf,
		logger: logger.Dup().WithField("family", "misconfiguration"),
	}
}
