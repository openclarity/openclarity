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

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/openclarity/vmclarity/runtime_scan/pkg/config/aws"
)

const (
	ScannerAWSRegion          = "SCANNER_AWS_REGION"
	defaultScannerAWSRegion   = "us-east-1"
	JobResultTimeout          = "JOB_RESULT_TIMEOUT"
	JobResultsPollingInterval = "JOB_RESULT_POLLING_INTERVAL"
	DeleteJobPolicy           = "DELETE_JOB_POLICY"
)

type OrchestratorConfig struct {
	AWSConfig       *aws.Config
	BackendAddress  string
	BackendRestPort int
	BackendBaseURL  string
	ScannerConfig
}

type ScannerConfig struct {
	Region                    string // scanner region TODO: why do we need this???
	JobResultTimeout          time.Duration
	JobResultsPollingInterval time.Duration
	DeleteJobPolicy           DeleteJobPolicyType
	ScannerImage              string
}

func setConfigDefaults() {
	viper.SetDefault(ScannerAWSRegion, defaultScannerAWSRegion)
	viper.SetDefault(JobResultTimeout, "120m")
	viper.SetDefault(JobResultsPollingInterval, "30s")
	viper.SetDefault(DeleteJobPolicy, DeleteJobPolicySuccessful)

	viper.AutomaticEnv()
}

func LoadConfig(backendAddress string, backendPort int, baseURL string) (*OrchestratorConfig, error) {
	setConfigDefaults()

	config := &OrchestratorConfig{
		AWSConfig:       aws.LoadConfig(),
		BackendRestPort: backendPort,
		BackendAddress:  backendAddress,
		BackendBaseURL:  baseURL,
		ScannerConfig: ScannerConfig{
			Region:                    viper.GetString(ScannerAWSRegion),
			JobResultTimeout:          viper.GetDuration(JobResultTimeout),
			JobResultsPollingInterval: viper.GetDuration(JobResultsPollingInterval),
			DeleteJobPolicy:           getDeleteJobPolicyType(viper.GetString(DeleteJobPolicy)),
			ScannerImage:              "what???", // TODO: Set the image
		},
	}

	return config, nil
}

func getDeleteJobPolicyType(policyType string) DeleteJobPolicyType {
	deleteJobPolicy := DeleteJobPolicyType(policyType)
	if !deleteJobPolicy.IsValid() {
		log.Warnf("Invalid %s type - using default `%s`", DeleteJobPolicy, DeleteJobPolicySuccessful)
		deleteJobPolicy = DeleteJobPolicySuccessful
	}

	return deleteJobPolicy
}
