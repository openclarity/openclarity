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
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
	"github.com/openclarity/kubeclarity/shared/pkg/utils"

	familiesinterface "github.com/openclarity/vmclarity/shared/pkg/families/interfaces"
	familiesresults "github.com/openclarity/vmclarity/shared/pkg/families/results"
	"github.com/openclarity/vmclarity/shared/pkg/families/rootkits/common"
	"github.com/openclarity/vmclarity/shared/pkg/families/rootkits/job"
	"github.com/openclarity/vmclarity/shared/pkg/families/types"
)

type Rootkits struct {
	conf   Config
	logger *log.Entry
}

func (r Rootkits) Run(res *familiesresults.Results) (familiesinterface.IsResults, error) {
	r.logger.Info("Rootkits Run...")

	manager := job_manager.New(r.conf.ScannersList, r.conf.ScannersConfig, r.logger, job.Factory)
	mergedResults := NewMergedResults()

	for _, input := range r.conf.Inputs {
		results, err := manager.Run(utils.SourceType(input.InputType), input.Input)
		if err != nil {
			return nil, fmt.Errorf("failed to scan input %q for rootkits: %v", input.Input, err)
		}

		// Merge results.
		for name, result := range results {
			r.logger.Infof("Merging result from %q", name)
			mergedResults = mergedResults.Merge(result.(*common.Results)) // nolint:forcetypeassert
		}
	}

	r.logger.Info("Rootkits Done...")
	return &Results{
		MergedResults: mergedResults,
	}, nil
}

func (r Rootkits) GetType() types.FamilyType {
	return types.Rootkits
}

// ensure types implement the requisite interfaces.
var _ familiesinterface.Family = &Rootkits{}

func New(logger *log.Entry, conf Config) *Rootkits {
	return &Rootkits{
		conf:   conf,
		logger: logger.Dup().WithField("family", "rootkits"),
	}
}
