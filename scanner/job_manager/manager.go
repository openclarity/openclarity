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
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/scanner/utils"
)

type Manager struct {
	jobNames   []string
	config     IsConfig
	logger     *logrus.Entry
	jobFactory *Factory
}

func New(jobNames []string, config IsConfig, logger *logrus.Entry, factory *Factory) *Manager {
	return &Manager{
		jobNames:   jobNames,
		config:     config,
		logger:     logger,
		jobFactory: factory,
	}
}

func (m *Manager) Run(ctx context.Context, sourceType utils.SourceType, userInput string) (map[string]Result, error) {
	nameToResultChan := make(map[string]chan Result, len(m.jobNames))

	// create jobs
	jobs := make([]Job, len(m.jobNames))
	for i, name := range m.jobNames {
		nameToResultChan[name] = make(chan Result, 10) // nolint:mnd
		job, err := m.jobFactory.CreateJob(name, m.config, m.logger, nameToResultChan[name])
		if err != nil {
			return nil, fmt.Errorf("failed to create job: %w", err)
		}

		jobs[i] = job
	}

	// start jobs
	for _, j := range jobs {
		err := j.Run(ctx, sourceType, userInput)
		if err != nil {
			return nil, fmt.Errorf("failed to run job: %w", err)
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
			errStr := fmt.Errorf("%q job failed: %w", name, err)
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
