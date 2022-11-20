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

	"github.com/openclarity/vmclarity/runtime_scan/pkg/config/aws"
)

const (
	ScannerAWSRegion                  = "SCANNER_AWS_REGION"
	defaultScannerAWSRegion           = "us-east-1"
	ScannerJobResultListenPort        = "SCANNER_JOB_RESULT_LISTEN_PORT"
	defaultScannerJobResultListenPort = 8888
)

type Config struct {
	ScannerJobResultListenPort int
	Region                     string // scanner region
	AWSConfig                  *aws.Config
}

func setConfigDefaults() {
	viper.SetDefault(ScannerAWSRegion, defaultScannerAWSRegion)
	viper.SetDefault(ScannerJobResultListenPort, defaultScannerJobResultListenPort)

	viper.AutomaticEnv()
}

func LoadConfig() (*Config, error) {
	setConfigDefaults()

	config := &Config{
		ScannerJobResultListenPort: viper.GetInt(ScannerJobResultListenPort),
		Region:                     viper.GetString(ScannerAWSRegion),
		AWSConfig:                  aws.LoadConfig(),
	}

	return config, nil
}
