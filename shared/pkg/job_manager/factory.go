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

type CreateJobFunc[C any, I any, R any] func(conf C, logger *logrus.Entry) (Job[I, R], error)

type Factory[C any, I any, R any] struct {
	createJobFuncs map[string]CreateJobFunc[C, I, R] // scanner name to CreateJobFunc
}

func NewJobFactory[C any, I any, R any]() *Factory[C, I, R] {
	return &Factory[C, I, R]{createJobFuncs: make(map[string]CreateJobFunc[C, I, R])}
}

func (f *Factory[C, I, R]) Register(name string, createJobFunc CreateJobFunc[C, I, R]) {
	if f.createJobFuncs == nil {
		f.createJobFuncs = make(map[string]CreateJobFunc[C, I, R])
	}

	if _, ok := f.createJobFuncs[name]; ok {
		logrus.Fatalf("%q already registered", name)
	}

	f.createJobFuncs[name] = createJobFunc
}

func (f *Factory[C, I, R]) CreateJob(name string, conf C, logger *logrus.Entry) (Job[I, R], error) {
	createFunc, ok := f.createJobFuncs[name]
	if !ok {
		return nil, fmt.Errorf("%v not a registered job", name)
	}

	return createFunc(conf, logger)
}
