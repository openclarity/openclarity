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

	"github.com/spf13/viper"

	"github.com/openclarity/vmclarity/e2e/testenv/docker"
	"github.com/openclarity/vmclarity/e2e/testenv/types"
)

const (
	DefaultEnvPrefix = "VMCLARITY_E2E"
	DefaultPlatform  = types.Docker
	DefaultReuseFlag = false
)

func NewConfig() (*types.Config, error) {
	// Avoid modifying the global instance
	v := viper.New()

	v.SetEnvPrefix(DefaultEnvPrefix)
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	_ = v.BindEnv("platform")
	v.SetDefault("platform", DefaultPlatform)
	_ = v.BindEnv("use_existing")
	v.SetDefault("use_existing", DefaultReuseFlag)

	config := &types.Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to parse testenv configuration: %w", err)
	}

	return config, nil
}

func New(config *types.Config) (types.Environment, error) {
	var env types.Environment
	var err error

	switch config.Platform {
	case types.Docker:
		env, err = docker.New(config)
	case types.AWS, types.Azure, types.GCP, types.Kubernetes:
		fallthrough
	default:
		err = fmt.Errorf("platform is not supported: %s", config.Platform)
	}

	return env, err
}
