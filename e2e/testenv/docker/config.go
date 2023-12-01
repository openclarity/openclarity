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

package docker

import (
	"context"

	envtypes "github.com/openclarity/vmclarity/e2e/testenv/types"
)

type ContainerImages = envtypes.ContainerImages[string]

//nolint:containedctx
type Config struct {
	// EnvName the name of the environment to be created
	EnvName string `mapstructure:"env_name"`
	// Images used for docker deployment
	Images ContainerImages `mapstructure:",squash"`
	// ComposeFiles contains the list of docker compose files used for deployment
	ComposeFiles []string `mapstructure:"compose_files"`
	// WorkDir absolute path to the directory where the deployment files prior performing actions
	WorkDir string `mapstructure:"work_dir"`
	// SkipAssetInstall defines whether to deploy test assets to environment or not
	SkipAssetInstall bool `mapstructure:"skip_asset_install"`

	// ctx used during project initialization
	ctx context.Context
}

// ConfigOptFn defines transformer function for Config.
type ConfigOptFn func(*Config) error

var applyConfigWithOpts = envtypes.WithOpts[Config, ConfigOptFn]

func WithContext(ctx context.Context) ConfigOptFn {
	return func(config *Config) error {
		config.ctx = ctx

		return nil
	}
}

// WithWorkDir set workDir for Config.
func WithWorkDir(dir string) ConfigOptFn {
	return func(config *Config) error {
		config.WorkDir = dir

		return nil
	}
}
