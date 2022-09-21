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
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	ScannerDependencyTrackAPIKey                         = "SCANNER_DEPENDENCY_TRACK_API_KEY"
	ScannerDependencyTrackHost                           = "SCANNER_DEPENDENCY_TRACK_HOST"
	ScannerDependencyTrackProjectName                    = "SCANNER_DEPENDENCY_TRACK_PROJECT_NAME"
	ScannerDependencyTrackProjectVersion                 = "SCANNER_DEPENDENCY_TRACK_PROJECT_VERSION"
	ScannerDependencyTrackShouldDeleteProject            = "SCANNER_DEPENDENCY_TRACK_SHOULD_DELETE_PROJECT"
	ScannerDependencyTrackDisableTLS                     = "SCANNER_DEPENDENCY_TRACK_DISABLE_TLS"
	ScannerDependencyTrackInsecureSkipVerify             = "SCANNER_DEPENDENCY_TRACK_INSECURE_SKIP_VERIFY"
	ScannerDependencyTrackFetchVulnerabilitiesRetryCount = "SCANNER_DEPENDENCY_TRACK_FETCH_VULNERABILITIES_RETRY_COUNT"
	ScannerDependencyTrackFetchVulnerabilitiesRetrySleep = "SCANNER_DEPENDENCY_TRACK_FETCH_VULNERABILITIES_RETRY_SLEEP"
)

type DependencyTrackConfig struct {
	APIKey                         string        `json:"-"`
	Host                           string        `json:"host"`
	ProjectName                    string        `json:"project-name"`
	ProjectVersion                 string        `json:"project-version"`
	ShouldDeleteProject            bool          `json:"should-delete-project"`
	DisableTLS                     bool          `json:"disable-tls"`
	InsecureSkipVerify             bool          `json:"insecure-skip-verify"`
	FetchVulnerabilitiesRetryCount int           `json:"fetch-vulnerabilities-retry-count"`
	FetchVulnerabilitiesRetrySleep time.Duration `json:"fetch-vulnerabilities-retry-sleep"`
}

func ConvertToDependencyTrackConfig(scanner *Scanner, logger *logrus.Entry) DependencyTrackConfig {
	dependencyTrackConfig := scanner.DependencyTrackConfig
	if logrus.IsLevelEnabled(logrus.InfoLevel) {
		configB, err := json.Marshal(dependencyTrackConfig)
		if err == nil {
			logger.Infof("DependencyTrack config: %s", configB)
		} else {
			logger.Errorf("Failed to marshal dependency track config: %v", err)
		}
	}
	return dependencyTrackConfig
}

func LoadDependencyTrackConfig() DependencyTrackConfig {
	setDependencyTrackConfigDefaults()
	return DependencyTrackConfig{
		APIKey:                         viper.GetString(ScannerDependencyTrackAPIKey),
		Host:                           viper.GetString(ScannerDependencyTrackHost),
		ProjectName:                    viper.GetString(ScannerDependencyTrackProjectName),
		ProjectVersion:                 viper.GetString(ScannerDependencyTrackProjectVersion),
		ShouldDeleteProject:            viper.GetBool(ScannerDependencyTrackShouldDeleteProject),
		DisableTLS:                     viper.GetBool(ScannerDependencyTrackDisableTLS),
		InsecureSkipVerify:             viper.GetBool(ScannerDependencyTrackInsecureSkipVerify),
		FetchVulnerabilitiesRetryCount: viper.GetInt(ScannerDependencyTrackFetchVulnerabilitiesRetryCount),
		FetchVulnerabilitiesRetrySleep: viper.GetDuration(ScannerDependencyTrackFetchVulnerabilitiesRetrySleep),
	}
}

func setDependencyTrackConfigDefaults() {
	viper.SetDefault(ScannerDependencyTrackAPIKey, "")
	viper.SetDefault(ScannerDependencyTrackHost, "dependency-track-apiserver.dependency-track")
	viper.SetDefault(ScannerDependencyTrackProjectName, "")    // will be auto generated if empty
	viper.SetDefault(ScannerDependencyTrackProjectVersion, "") // will be auto generated if empty
	viper.SetDefault(ScannerDependencyTrackShouldDeleteProject, true)
	viper.SetDefault(ScannerDependencyTrackDisableTLS, false)
	viper.SetDefault(ScannerDependencyTrackInsecureSkipVerify, true)
	viper.SetDefault(ScannerDependencyTrackFetchVulnerabilitiesRetryCount, 5)              // nolint:gomnd
	viper.SetDefault(ScannerDependencyTrackFetchVulnerabilitiesRetrySleep, 30*time.Second) // nolint:gomnd
}
