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

package server

import (
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"

	dbtypes "github.com/openclarity/vmclarity/api/server/database/types"

	"github.com/spf13/viper"
)

const (
	DefaultEnvPrefix          = "VMCLARITY_APISERVER"
	DefaultListenAddress      = "0.0.0.0:8888"
	DefaultHealthCheckAddress = "0.0.0.0:8081"
	DefaultDatabaseDriver     = dbtypes.DBDriverTypeLocal
)

type Config struct {
	ListenAddress      string `json:"listen-address,omitempty" mapstructure:"listen_address"`
	HealthCheckAddress string `json:"healthcheck-address,omitempty" mapstructure:"healthcheck_address"`

	// database config
	DatabaseDriver   string `json:"database-driver,omitempty" mapstructure:"database_driver"`
	DBName           string `json:"db-name,omitempty" mapstructure:"db_name"`
	DBUser           string `json:"db-user,omitempty" mapstructure:"db_user"`
	DBPassword       string `json:"-" mapstructure:"db_pass"`
	DBHost           string `json:"db-host,omitempty" mapstructure:"db_host"`
	DBPort           string `json:"db-port,omitempty" mapstructure:"db_port"`
	EnableDBInfoLogs bool   `json:"enable-db-info-logs" mapstructure:"enable_db_info_logs"`

	LocalDBPath string `json:"local-db-path,omitempty" mapstructure:"local_db_path"`
}

func NewConfig() (*Config, error) {
	v := viper.NewWithOptions(
		viper.KeyDelimiter("."),
		viper.EnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_")),
	)

	v.SetEnvPrefix(DefaultEnvPrefix)
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	_ = v.BindEnv("listen_address")
	v.SetDefault("listen_address", DefaultListenAddress)

	_ = v.BindEnv("healthcheck_address")
	v.SetDefault("healthcheck_address", DefaultHealthCheckAddress)

	_ = v.BindEnv("database_driver")
	v.SetDefault("database_driver", DefaultDatabaseDriver)

	_ = v.BindEnv("db_name")

	_ = v.BindEnv("db_user")

	_ = v.BindEnv("db_pass")

	_ = v.BindEnv("db_host")

	_ = v.BindEnv("db_port")

	_ = v.BindEnv("enable_db_info_logs")

	_ = v.BindEnv("local_db_path")

	decodeHooks := mapstructure.ComposeDecodeHookFunc(
		// TextUnmarshallerHookFunc is needed to decode custom types
		mapstructure.TextUnmarshallerHookFunc(),
		// Default decoders
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	)

	config := &Config{}
	if err := v.Unmarshal(config, viper.DecodeHook(decodeHooks)); err != nil {
		return nil, fmt.Errorf("failed to load API Server configuration: %w", err)
	}

	return config, nil
}
