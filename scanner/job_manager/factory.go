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

	"github.com/sirupsen/logrus"
)

type Factory struct {
	createJobFuncs map[string]CreateJobFunc // scanner name to CreateJobFunc
}

func NewJobFactory() *Factory {
	return &Factory{createJobFuncs: make(map[string]CreateJobFunc)}
}

type CreateJobFunc func(name string, conf IsConfig, logger *logrus.Entry, resultChan chan Result) Job

func (f *Factory) Register(name string, createJobFunc CreateJobFunc) {
	if f.createJobFuncs == nil {
		f.createJobFuncs = make(map[string]CreateJobFunc)
	}

	if _, ok := f.createJobFuncs[name]; ok {
		logrus.Fatalf("%q already registered", name)
	}

	f.createJobFuncs[name] = createJobFunc
}

func (f *Factory) CreateJob(name string, conf IsConfig, logger *logrus.Entry, resultChan chan Result) (Job, error) {
	createFunc, ok := f.createJobFuncs[name]
	if !ok {
		return nil, fmt.Errorf("%v not a registered job", name)
	}

	return createFunc(name, conf, logger, resultChan), nil
}
