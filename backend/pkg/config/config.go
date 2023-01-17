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
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	BackendRestPort    = "BACKEND_REST_PORT"
	HealthCheckAddress = "HEALTH_CHECK_ADDRESS"

	PrometheusRefreshIntervalSeconds = "PROMETHEUS_REFRESH_INTERVAL_SECONDS"

	DBNameEnvVar              = "DB_NAME"
	DBUserEnvVar              = "DB_USER"
	DBPasswordEnvVar          = "DB_PASS"
	DBHostEnvVar              = "DB_HOST"
	DBPortEnvVar              = "DB_PORT_NUMBER"
	DatabaseDriver            = "DATABASE_DRIVER"
	EnableDBInfoLogs          = "ENABLE_DB_INFO_LOGS"
	ViewRefreshIntervalEnvVar = "DB_VIEW_REFRESH_INTERVAL"

	FakeDataEnvVar           = "FAKE_DATA"
	FakeRuntimeScannerEnvVar = "FAKE_RUNTIME_SCANNER"
)

type Config struct {
	BackendRestPort    int
	HealthCheckAddress string

	// How often to refresh Prometheus data in seconds
	PrometheusRefreshIntervalSeconds int

	// database config
	DatabaseDriver            string
	DBName                    string
	DBUser                    string
	DBPassword                string
	DBHost                    string
	DBPort                    string
	EnableDBInfoLogs          bool
	EnableFakeData            bool
	EnableFakeRuntimeScanner  bool
	ViewRefreshIntervalSecond int
}

func LoadConfig() (*Config, error) {
	config := &Config{}

	config.BackendRestPort = viper.GetInt(BackendRestPort)
	config.HealthCheckAddress = viper.GetString(HealthCheckAddress)

	config.PrometheusRefreshIntervalSeconds = viper.GetInt(PrometheusRefreshIntervalSeconds)

	config.DatabaseDriver = viper.GetString(DatabaseDriver)
	config.DBPassword = viper.GetString(DBPasswordEnvVar)
	config.DBUser = viper.GetString(DBUserEnvVar)
	config.DBHost = viper.GetString(DBHostEnvVar)
	config.DBPort = viper.GetString(DBPortEnvVar)
	config.DBName = viper.GetString(DBNameEnvVar)
	config.ViewRefreshIntervalSecond = viper.GetInt(ViewRefreshIntervalEnvVar)
	config.EnableDBInfoLogs = viper.GetBool(EnableDBInfoLogs)
	config.EnableFakeData = viper.GetBool(FakeDataEnvVar)
	config.EnableFakeRuntimeScanner = viper.GetBool(FakeRuntimeScannerEnvVar)

	configB, err := json.Marshal(config)
	if err == nil {
		log.Infof("\n\nconfig=%s\n\n", configB)
	} else {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	return config, nil
}
