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

package job_manager // nolint:revive,stylecheck

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"

	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/config"
	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/utils"
)

type Manager struct {
	jobNames          []string
	config            *config.Config
	logger            *logrus.Entry
	createRunnersFunc createJobFunc
}

func New(jobNames []string, config *config.Config, logger *logrus.Entry, createRunnersFunc createJobFunc) *Manager {
	return &Manager{
		jobNames:          jobNames,
		config:            config,
		logger:            logger,
		createRunnersFunc: createRunnersFunc,
	}
}

func (m *Manager) Run(sourceType utils.SourceType, source string) (map[string]Result, error) {
	nameToResultChan := make(map[string]chan Result, len(m.jobNames))

	// create jobs
	jobs := make([]Job, len(m.jobNames))
	for i, name := range m.jobNames {
		nameToResultChan[name] = make(chan Result, 10) // nolint:gomnd
		jobs[i] = m.createRunnersFunc(name, m.config, m.logger, nameToResultChan[name])
	}

	// start jobs
	for _, j := range jobs {
		err := j.Run(sourceType, source)
		if err != nil {
			return nil, fmt.Errorf("failed to run job: %v", err)
		}
	}

	// wait for results
	var resultError error
	totalSuccessfulResultsCount := 0
	results := make(map[string]Result, len(m.jobNames))
	for name, channel := range nameToResultChan {
		// TODO: maybe need a timeout while waiting for a specific job result?
		result := <-channel

		if err := result.GetError(); err != nil {
			errStr := fmt.Errorf("%q job failed: %v", name, err)
			m.logger.Warning(errStr)
			resultError = multierror.Append(resultError, errStr)
		} else {
			m.logger.Infof("Got result for job %q", name)
			results[name] = result
			totalSuccessfulResultsCount++
		}
	}

	// Return error if all jobs failed to return results.
	// TODO: should it be configurable? allow the user to decide failure threshold?
	if totalSuccessfulResultsCount == 0 {
		return nil, resultError // nolint:wrapcheck
	}

	return results, nil
}
