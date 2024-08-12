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

package azure

import (
	envtypes "github.com/openclarity/vmclarity/testenv/types"
)

const (
	// DefaultLocation the default Azure region to be used.
	DefaultLocation = "eastus"
)

type ContainerImages = envtypes.ContainerImages[string]

type Config struct {
	// EnvName the name of the environment to be created
	EnvName string `mapstructure:"env_name"`
	// WorkDir absolute path to the directory where the deployment files prior performing actions
	WorkDir string `mapstructure:"work_dir"`
	// Region the Azure region to be used
	Region string `mapstructure:"region"`
	// PublicKeyFile the public key file to be used for the key pair
	PublicKeyFile string `mapstructure:"public_key_file"`
	// PrivateKeyFile the private key file to be used for the key pair
	PrivateKeyFile string `mapstructure:"private_key_file"`
}

// ConfigOptFn defines transformer function for Config.
type ConfigOptFn func(*Config) error

var applyConfigWithOpts = envtypes.WithOpts[Config, ConfigOptFn]

// WithWorkDir set workDir for Config.
func WithWorkDir(dir string) ConfigOptFn {
	return func(config *Config) error {
		config.WorkDir = dir

		return nil
	}
}
