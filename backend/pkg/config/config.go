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

package config

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	BackendRestAddress    = "BACKEND_REST_ADDRESS"
	BackendRestDisableTLS = "BACKEND_REST_DISABLE_TLS"
	BackendRestPort       = "BACKEND_REST_PORT"
	HealthCheckAddress    = "HEALTH_CHECK_ADDRESS"

	DBNameEnvVar     = "DB_NAME"
	DBUserEnvVar     = "DB_USER"
	DBPasswordEnvVar = "DB_PASS"
	DBHostEnvVar     = "DB_HOST"
	DBPortEnvVar     = "DB_PORT_NUMBER"
	DatabaseDriver   = "DATABASE_DRIVER"
	EnableDBInfoLogs = "ENABLE_DB_INFO_LOGS"

	LocalDBPath = "LOCAL_DB_PATH"

	FakeDataEnvVar           = "FAKE_DATA"
	FakeRuntimeScannerEnvVar = "FAKE_RUNTIME_SCANNER"
)

type Config struct {
	BackendRestAddress string `json:"backend-rest-address,omitempty"`
	BackendRestPort    int    `json:"backend-rest-port,omitempty"`
	HealthCheckAddress string `json:"health-check-address,omitempty"`

	// database config
	DatabaseDriver           string `json:"database-driver,omitempty"`
	DBName                   string `json:"db-name,omitempty"`
	DBUser                   string `json:"db-user,omitempty"`
	DBPassword               string `json:"-"`
	DBHost                   string `json:"db-host,omitempty"`
	DBPort                   string `json:"db-port,omitempty"`
	EnableDBInfoLogs         bool   `json:"enable-db-info-logs"`
	EnableFakeData           bool   `json:"enable-fake-data"`
	EnableFakeRuntimeScanner bool   `json:"enable-fake-runtime-scanner"`

	LocalDBPath string `json:"local-db-path,omitempty"`
}

func LoadConfig() (*Config, error) {
	config := &Config{}

	config.BackendRestAddress = viper.GetString(BackendRestAddress)
	config.BackendRestPort = viper.GetInt(BackendRestPort)
	config.HealthCheckAddress = viper.GetString(HealthCheckAddress)

	config.DatabaseDriver = viper.GetString(DatabaseDriver)
	config.DBPassword = viper.GetString(DBPasswordEnvVar)
	config.DBUser = viper.GetString(DBUserEnvVar)
	config.DBHost = viper.GetString(DBHostEnvVar)
	config.DBPort = viper.GetString(DBPortEnvVar)
	config.DBName = viper.GetString(DBNameEnvVar)
	config.EnableDBInfoLogs = viper.GetBool(EnableDBInfoLogs)
	config.EnableFakeData = viper.GetBool(FakeDataEnvVar)
	config.EnableFakeRuntimeScanner = viper.GetBool(FakeRuntimeScannerEnvVar)

	config.LocalDBPath = viper.GetString(LocalDBPath)

	configB, err := json.Marshal(config)
	if err == nil {
		log.Infof("\n\nconfig=%s\n\n", configB)
	} else {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	return config, nil
}
