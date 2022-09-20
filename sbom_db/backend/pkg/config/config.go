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

package config

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	BackendRestPort    = "BACKEND_REST_PORT"
	HealthCheckAddress = "HEALTH_CHECK_ADDRESS"

	EnableDBInfoLogs = "ENABLE_DB_INFO_LOGS"

	FakeDataEnvVar = "FAKE_DATA"
)

type Config struct {
	BackendRestPort    int
	HealthCheckAddress string

	// database config
	EnableDBInfoLogs bool
	EnableFakeData   bool
}

func LoadConfig() (*Config, error) {
	config := &Config{}

	config.BackendRestPort = viper.GetInt(BackendRestPort)
	config.HealthCheckAddress = viper.GetString(HealthCheckAddress)

	config.EnableDBInfoLogs = viper.GetBool(EnableDBInfoLogs)
	config.EnableFakeData = viper.GetBool(FakeDataEnvVar)

	configB, err := json.Marshal(config)
	if err == nil {
		log.Infof("\n\nconfig=%s\n\n", configB)
	} else {
		log.Warningf("Failed to marshal config. %v", err)
	}

	return config, nil
}
