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

package gcp

import (
	"context"

	envtypes "github.com/openclarity/vmclarity/testenv/types"
)

//nolint:containedctx
type Config struct {
	WorkDir        string `mapstructure:"work_dir"`
	EnvName        string `mapstructure:"env_name"`
	PublicKeyFile  string `mapstructure:"public_key_file"`
	PrivateKeyFile string `mapstructure:"private_key_file"`

	ctx context.Context
}

type ConfigOptFn func(*Config) error

var applyConfigWithOpts = envtypes.WithOpts[Config, ConfigOptFn]

func WithContext(ctx context.Context) ConfigOptFn {
	return func(config *Config) error {
		config.ctx = ctx

		return nil
	}
}

func WithWorkDir(dir string) ConfigOptFn {
	return func(config *Config) error {
		config.WorkDir = dir

		return nil
	}
}
