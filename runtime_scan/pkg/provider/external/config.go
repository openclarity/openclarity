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

package external

import (
	"fmt"

	"github.com/spf13/viper"
)

const (
	DefaultEnvPrefix = "VMCLARITY_EXTERNAL"
)

type Config struct {
	ProviderPluginAddress string `mapstructure:"provider_plugin_address"`
}

func (c *Config) Validate() error {
	if c.ProviderPluginAddress == "" {
		return fmt.Errorf("parameter ProviderPluginAddress must be provided")
	}

	return nil
}

func NewConfig() (*Config, error) {
	// Avoid modifying the global instance
	v := viper.New()

	v.SetEnvPrefix(DefaultEnvPrefix)
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	_ = v.BindEnv("provider_plugin_address")

	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to parse provider configuration. Provider=External: %w", err)
	}

	return config, nil
}
