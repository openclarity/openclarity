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

package testenv

import (
	"fmt"

	"github.com/openclarity/vmclarity/e2e/testenv/docker"
	"github.com/openclarity/vmclarity/e2e/testenv/kubernetes"
	"github.com/openclarity/vmclarity/e2e/testenv/types"
)

// New returns an object implementing the types.Environment interface from Config.
func New(config *Config, opts ...ConfigOptFn) (types.Environment, error) {
	opts = append(opts,
		withResolvedWorkDirPath(),
		withDefaultWorkDir(),
	)
	if err := applyConfigWithOpts(config, opts...); err != nil {
		return nil, fmt.Errorf("failed to apply options to ProviderConfig: %w", err)
	}

	var env types.Environment
	var err error
	switch config.Platform {
	case types.EnvironmentTypeDocker:
		env, err = docker.New(config.Docker,
			docker.WithContext(config.ctx),
			docker.WithWorkDir(config.WorkDir),
		)
	case types.EnvironmentTypeKubernetes:
		env, err = kubernetes.New(config.Kubernetes,
			kubernetes.WithLogger(config.logger),
			kubernetes.WithWorkDir(config.WorkDir),
		)
	case types.EnvironmentTypeAWS, types.EnvironmentTypeAzure, types.EnvironmentTypeGCP:
		fallthrough
	default:
		err = fmt.Errorf("unsupported Environment: %s", config.Platform)
	}

	return env, err
}
