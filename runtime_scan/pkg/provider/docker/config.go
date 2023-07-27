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
	"fmt"

	"github.com/spf13/viper"
)

const (
	DefaultEnvPrefix   = "VMCLARITY_DOCKER"
	DefaultHelperImage = "alpine:3.18.2"
	DefaultNetworkName = "vmclarity"
)

type Config struct {
	// HelperImage defines helper container image that performs init tasks.
	HelperImage string `mapstructure:"helper_image"`
	// NetworkName defines the user defined bridge network where we attach the scanner container.
	NetworkName string `mapstructure:"network_name"`
}

func NewConfig() (*Config, error) {
	// Avoid modifying the global instance
	v := viper.New()

	v.SetEnvPrefix(DefaultEnvPrefix)
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	_ = v.BindEnv("helper_image")
	v.SetDefault("helper_image", DefaultHelperImage)

	_ = v.BindEnv("network_name")
	v.SetDefault("network_name", DefaultNetworkName)

	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to parse provider configuration. Provider=Docker: %w", err)
	}

	return config, nil
}
