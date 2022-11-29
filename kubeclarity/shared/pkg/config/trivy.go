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
	"time"

	"github.com/spf13/viper"
)

const AnalyzerTrivyTimeoutSecondsDefault = 300

const AnalyzerTrivyTimeoutSeconds = "ANALYZER_TRIVY_TIMEOUT_SECONDS"

type AnalyzerTrivyConfig struct {
	Timeout int `yaml:"timeout" mapstructure:"timeout"`
}

func setAnalyzerTrivyConfigDefaults() {
	viper.SetDefault(AnalyzerTrivyTimeoutSeconds, AnalyzerTrivyTimeoutSecondsDefault)
}

func LoadAnalyzerTrivyConfig() AnalyzerTrivyConfig {
	setAnalyzerTrivyConfigDefaults()

	return AnalyzerTrivyConfig{
		Timeout: viper.GetInt(AnalyzerTrivyTimeoutSeconds),
	}
}

type AnalyzerTrivyConfigEx struct {
	Timeout  time.Duration
	Registry *Registry
}

func CreateAnalyzerTrivyConfigEx(analyzer *Analyzer, registry *Registry) AnalyzerTrivyConfigEx {
	return AnalyzerTrivyConfigEx{
		Timeout:  time.Duration(analyzer.TrivyConfig.Timeout) * time.Second,
		Registry: registry,
	}
}

const ScannerTrivyTimeoutSecondsDefault = 300

const (
	ScannerTrivyTimeoutSeconds = "SCANNER_TRIVY_TIMEOUT_SECONDS"
	ScannerTrivyServerAddress  = "SCANNER_TRIVY_SERVER_ADDRESS"
	ScannerTrivyServerToken    = "SCANNER_TRIVY_SERVER_TOKEN" // nolint:gosec
)

type ScannerTrivyConfig struct {
	Timeout     int
	ServerAddr  string
	ServerToken string
}

func setScannerTrivyConfigDefaults() {
	viper.SetDefault(ScannerTrivyTimeoutSeconds, ScannerTrivyTimeoutSecondsDefault)
}

func LoadScannerTrivyConfig() ScannerTrivyConfig {
	setScannerTrivyConfigDefaults()

	return ScannerTrivyConfig{
		Timeout:     viper.GetInt(ScannerTrivyTimeoutSeconds),
		ServerAddr:  viper.GetString(ScannerTrivyServerAddress),
		ServerToken: viper.GetString(ScannerTrivyServerToken),
	}
}

type ScannerTrivyConfigEx struct {
	Timeout     time.Duration
	ServerAddr  string
	ServerToken string
	Registry    *Registry
}

func CreateScannerTrivyConfigEx(scanner *Scanner, registry *Registry) ScannerTrivyConfigEx {
	return ScannerTrivyConfigEx{
		Timeout:     time.Duration(scanner.TrivyConfig.Timeout) * time.Second,
		Registry:    registry,
		ServerAddr:  scanner.TrivyConfig.ServerAddr,
		ServerToken: scanner.TrivyConfig.ServerToken,
	}
}
