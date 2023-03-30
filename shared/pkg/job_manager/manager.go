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
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
)

type Manager[I any, R any] struct {
	jobs   map[string]Job[I, R]
	logger *logrus.Entry
}

func New[C any, I any, R any](jobNames []string, config C, logger *logrus.Entry, factory *Factory[C, I, R]) (*Manager[I, R], error) {
	// create jobs
	jobs := make(map[string]Job[I, R], len(jobNames))
	for _, name := range jobNames {
		job, err := factory.CreateJob(name, config, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create job: %v", err)
		}

		jobs[name] = job
	}

	return &Manager[I, R]{
		jobs:   jobs,
		logger: logger,
	}, nil
}

func (m *Manager[I, R]) Run(input I) (map[string]R, error) {
	results := make(map[string]R, len(m.jobs))
	errors := make(map[string]error, len(m.jobs))

	var wg sync.WaitGroup

	// run jobs in parallel
	for name, j := range m.jobs {
		jobName := name
		job := j
		wg.Add(1)
		go func() {
			defer wg.Done()

			// TODO: maybe need a ctx + timeout for each job to
			// prevent them locking up forever
			result, err := job.Run(input)
			if err != nil {
				err = fmt.Errorf("%q job failed: %v", jobName, err)
				m.logger.Warning(err)
			}
			results[jobName] = result
			errors[jobName] = err
		}()
	}

	wg.Wait()

	// Merge errors together and count them
	errCount := 0
	var resultError error
	for _, err := range errors {
		if err != nil {
			resultError = multierror.Append(resultError, err)
			errCount++
		}
	}

	// Return error if all jobs failed to return results.
	// TODO: should it be configurable? allow the user to decide failure threshold?
	if errCount == len(m.jobs) {
		return nil, resultError // nolint:wrapcheck
	}

	return results, nil
}
