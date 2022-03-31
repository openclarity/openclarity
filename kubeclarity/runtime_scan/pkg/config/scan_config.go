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
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
)

const (
	MaxParallelism               = "MAX_PARALLELISM"
	TargetNamespace              = "TARGET_NAMESPACE"
	IgnoreNamespaces             = "IGNORE_NAMESPACES"
	JobResultTimeout             = "JOB_RESULT_TIMEOUT"
	DeleteJobPolicy              = "DELETE_JOB_POLICY"
	ShouldScanCISDockerBenchmark = "SHOULD_SCAN_CIS_DOCKER_BENCHMARK"
	RegistryInsecure             = "REGISTRY_INSECURE"
)

type ScanConfig struct {
	MaxScanParallelism           int
	TargetNamespaces             []string
	IgnoredNamespaces            []string
	JobResultTimeout             time.Duration
	DeleteJobPolicy              DeleteJobPolicyType
	ShouldScanCISDockerBenchmark bool
}

func setScanConfigDefaults() {
	viper.SetDefault(MaxParallelism, "10")
	viper.SetDefault(TargetNamespace, corev1.NamespaceAll) // Scan all namespaces by default
	viper.SetDefault(IgnoreNamespaces, "")
	viper.SetDefault(JobResultTimeout, "10m")
	viper.SetDefault(DeleteJobPolicy, DeleteJobPolicySuccessful)
	viper.SetDefault(ShouldScanCISDockerBenchmark, "false")

	viper.AutomaticEnv()
}

func LoadScanConfig() *ScanConfig {
	setScanConfigDefaults()

	shouldScanCISDockerBenchmark := viper.GetBool(ShouldScanCISDockerBenchmark)
	registryInsecure, _ := strconv.ParseBool(viper.GetString(RegistryInsecure))
	// Disable CIS docker benchmark scan if insecure registry is set - currently not supported
	if registryInsecure {
		shouldScanCISDockerBenchmark = false
	}

	return &ScanConfig{
		MaxScanParallelism:           viper.GetInt(MaxParallelism),
		IgnoredNamespaces:            strings.Split(viper.GetString(IgnoreNamespaces), ","),
		JobResultTimeout:             viper.GetDuration(JobResultTimeout),
		DeleteJobPolicy:              getDeleteJobPolicyType(viper.GetString(DeleteJobPolicy)),
		ShouldScanCISDockerBenchmark: shouldScanCISDockerBenchmark,
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
