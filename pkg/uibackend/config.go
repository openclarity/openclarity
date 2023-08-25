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

package uibackend

import (
	"github.com/spf13/viper"
)

const (
	ListenAddress       = "LISTEN_ADDRESS"
	APIServerHost       = "APISERVER_HOST"
	APIServerDisableTLS = "APISERVER_DISABLE_TLS"
	APIServerPort       = "APISERVER_PORT"
	HealthCheckAddress  = "HEALTH_CHECK_ADDRESS"
)

type Config struct {
	ListenAddress      string `json:"listen-address,omitempty"`
	APIServerHost      string `json:"apiserver-host,omitempty"`
	APIServerPort      int    `json:"apiserver-port,omitempty"`
	HealthCheckAddress string `json:"health-check-address,omitempty"`
}

func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()

	viper.SetDefault(ListenAddress, ":8890")
	viper.SetDefault(HealthCheckAddress, ":8083")

	c := &Config{
		ListenAddress:      viper.GetString(ListenAddress),
		APIServerHost:      viper.GetString(APIServerHost),
		APIServerPort:      viper.GetInt(APIServerPort),
		HealthCheckAddress: viper.GetString(HealthCheckAddress),
	}
	return c, nil
}
