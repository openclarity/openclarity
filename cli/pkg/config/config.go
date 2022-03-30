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
	"github.com/spf13/viper"

	shared "wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/config"
)

const (
	LogLevel       = "LOG_LEVEL"
	EnableJSONLog  = "ENABLE_JSON_LOG"
	LocalImageScan = "LOCAL_IMAGE_SCAN"
)

type Config struct {
	LogLevel       string
	EnableJSONLog  bool
	LocalImageScan bool
	Backend        *Backend
	SharedConfig   *shared.Config
}

func LoadConfig() *Config {
	setConfigDefaults()
	return &Config{
		LogLevel:       viper.GetString(LogLevel),
		EnableJSONLog:  viper.GetBool(EnableJSONLog),
		LocalImageScan: viper.GetBool(LocalImageScan),
		Backend:        loadBackendConfig(),
		SharedConfig: &shared.Config{
			Registry: loadRegistryConfig(),
			Analyzer: shared.LoadAnalyzerConfig(),
			Scanner:  shared.LoadScannerConfig(),
		},
	}
}

func setConfigDefaults() {
	viper.SetDefault(LogLevel, "info")
	viper.SetDefault(EnableJSONLog, false)
}
