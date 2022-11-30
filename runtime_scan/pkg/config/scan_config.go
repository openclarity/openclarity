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

	"github.com/openclarity/vmclarity/runtime_scan/pkg/types"
)

const (
	JobResultTimeout = "JOB_RESULT_TIMEOUT"
	MaxParallelism   = "MAX_PARALLELISM"
	DeleteJobPolicy  = "DELETE_JOB_POLICY"
)

type ScanConfig struct {
	MaxScanParallelism int
	// instances to scan
	Instances []types.Instance
	// per provider scan scope
	ScanScope        types.ScanScope
	JobResultTimeout time.Duration
	DeleteJobPolicy  DeleteJobPolicyType
	ScannerConfig    *types.ScannerConfig
}

func setScanConfigDefaults() {
	viper.SetDefault(MaxParallelism, "5")
	viper.SetDefault(JobResultTimeout, "120m")
	viper.SetDefault(DeleteJobPolicy, DeleteJobPolicySuccessful)

	viper.AutomaticEnv()
}

func LoadScanConfig() *ScanConfig {
	setScanConfigDefaults()

	return &ScanConfig{
		MaxScanParallelism: viper.GetInt(MaxParallelism),
		JobResultTimeout:   viper.GetDuration(JobResultTimeout),
		DeleteJobPolicy:    getDeleteJobPolicyType(viper.GetString(DeleteJobPolicy)),
	}
}

func getDeleteJobPolicyType(policyType string) DeleteJobPolicyType {
	deleteJobPolicy := DeleteJobPolicyType(policyType)
	if !deleteJobPolicy.IsValid() {
		log.Warnf("Invalid %s type - using default `%s`", DeleteJobPolicy, DeleteJobPolicySuccessful)
		deleteJobPolicy = DeleteJobPolicySuccessful
	}

	return deleteJobPolicy
}
