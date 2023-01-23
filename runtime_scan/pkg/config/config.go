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
	"fmt"
	"net"
	"strconv"
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
	ScannerContainerImage     = "SCANNER_CONTAINER_IMAGE"
	ScannerBackendAddress     = "SCANNER_VMCLARITY_BACKEND_ADDRESS"
	ScanConfigWatchInterval   = "SCAN_CONFIG_WATCH_INTERVAL"
)

type OrchestratorConfig struct {
	AWSConfig             *aws.Config
	ScannerBackendAddress string
	ScannerConfig
}

type ScannerConfig struct {
	// We need to know where the VMClarity scanner is running so that we
	// can boot the scanner jobs in the same region, there isn't a
	// mechanism to discover this right now so its passed in as a config
	// value.
	Region string

	// Address that the Scanner should use to talk to the VMClarity backend
	// We use a configuration variable for this instead of discovering it
	// automatically in case VMClarity backend has multiple IPs (internal
	// traffic and external traffic for example) so we need the specific
	// address to use.
	ScannerBackendAddress string

	JobResultTimeout          time.Duration
	JobResultsPollingInterval time.Duration
	ScanConfigWatchInterval   time.Duration
	DeleteJobPolicy           DeleteJobPolicyType

	// The container image to use once we've booted the scanner virtual
	// machine, that contains the VMClarity CLI plus all the required
	// tools.
	ScannerImage string
}

func setConfigDefaults(backendAddress string, backendPort int, backendBaseURL string) {
	viper.SetDefault(ScannerAWSRegion, defaultScannerAWSRegion)
	viper.SetDefault(JobResultTimeout, "120m")
	viper.SetDefault(JobResultsPollingInterval, "30s")
	viper.SetDefault(ScanConfigWatchInterval, "30s")
	viper.SetDefault(DeleteJobPolicy, string(DeleteJobPolicySuccessful))
	viper.SetDefault(ScannerBackendAddress, fmt.Sprintf("http://%s%s", net.JoinHostPort(backendAddress, strconv.Itoa(backendPort)), backendBaseURL))

	viper.AutomaticEnv()
}

func LoadConfig(backendAddress string, backendPort int, baseURL string) (*OrchestratorConfig, error) {
	setConfigDefaults(backendAddress, backendPort, baseURL)

	config := &OrchestratorConfig{
		AWSConfig:             aws.LoadConfig(),
		ScannerBackendAddress: viper.GetString(ScannerBackendAddress),
		ScannerConfig: ScannerConfig{
			Region:                    viper.GetString(ScannerAWSRegion),
			JobResultTimeout:          viper.GetDuration(JobResultTimeout),
			JobResultsPollingInterval: viper.GetDuration(JobResultsPollingInterval),
			ScanConfigWatchInterval:   viper.GetDuration(ScanConfigWatchInterval),
			DeleteJobPolicy:           getDeleteJobPolicyType(viper.GetString(DeleteJobPolicy)),
			ScannerImage:              viper.GetString(ScannerContainerImage),
			ScannerBackendAddress:     viper.GetString(ScannerBackendAddress),
		},
	}

	return config, nil
}

func getDeleteJobPolicyType(policyType string) DeleteJobPolicyType {
	deleteJobPolicy := DeleteJobPolicyType(policyType)
	if !deleteJobPolicy.IsValid() {
		log.Warnf("Invalid %s type (%s) - using default `%s`", DeleteJobPolicy, policyType, DeleteJobPolicySuccessful)
		deleteJobPolicy = DeleteJobPolicySuccessful
	}

	return deleteJobPolicy
}
