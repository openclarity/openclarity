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

package orchestrator

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/spf13/viper"

	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/discovery"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/scanconfigwatcher"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/scanresultprocessor"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/scanresultwatcher"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/scanwatcher"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider/aws"
)

const (
	ScannerAWSRegion              = "SCANNER_AWS_REGION"
	DeleteJobPolicy               = "DELETE_JOB_POLICY"
	ScannerContainerImage         = "SCANNER_CONTAINER_IMAGE"
	ScannerKeyPairName            = "SCANNER_KEY_PAIR_NAME"
	GitleaksBinaryPath            = "GITLEAKS_BINARY_PATH"
	ClamBinaryPath                = "CLAM_BINARY_PATH"
	FreshclamBinaryPath           = "FRESHCLAM_BINARY_PATH"
	AlternativeFreshclamMirrorURL = "ALTERNATIVE_FRESHCLAM_MIRROR_URL"
	LynisInstallPath              = "LYNIS_INSTALL_PATH"
	AttachedVolumeDeviceName      = "ATTACHED_VOLUME_DEVICE_NAME"
	ScannerBackendAddress         = "SCANNER_VMCLARITY_BACKEND_ADDRESS"
	ScanConfigWatchInterval       = "SCAN_CONFIG_WATCH_INTERVAL"
	ExploitDBAddress              = "EXPLOIT_DB_ADDRESS"
	TrivyServerAddress            = "TRIVY_SERVER_ADDRESS"
	TrivyServerTimeout            = "TRIVY_SERVER_TIMEOUT"
	GrypeServerAddress            = "GRYPE_SERVER_ADDRESS"
	GrypeServerTimeout            = "GRYPE_SERVER_TIMEOUT"
	ChkrootkitBinaryPath          = "CHKROOTKIT_BINARY_PATH"

	ScanConfigPollingInterval  = "SCAN_CONFIG_POLLING_INTERVAL"
	ScanConfigReconcileTimeout = "SCAN_CONFIG_RECONCILE_TIMEOUT"

	ScanPollingInterval  = "SCAN_POLLING_INTERVAL"
	ScanReconcileTimeout = "SCAN_RECONCILE_TIMEOUT"
	ScanTimeout          = "SCAN_TIMEOUT"

	ScanResultPollingInterval  = "SCAN_RESULT_POLLING_INTERVAL"
	ScanResultReconcileTimeout = "SCAN_RESULT_RECONCILE_TIMEOUT"

	ScanResultProcessorPollingInterval  = "SCAN_RESULT_PROCESSOR_POLLING_INTERVAL"
	ScanResultProcessorReconcileTimeout = "SCAN_RESULT_PROCESSOR_RECONCILE_TIMEOUT"

	DiscoveryInterval = "DISCOVERY_INTERVAL"
)

const (
	DefaultScannerAWSRegion         = "us-east-1"
	DefaultAttachedVolumeDeviceName = "xvdh"

	DefaultTrivyServerTimeout = 5 * time.Minute
	DefaultGrypeServerTimeout = 2 * time.Minute
)

type Config struct {
	AWSConfig             *aws.Config
	ScannerBackendAddress string

	DiscoveryConfig           discovery.Config
	ScanConfigWatcherConfig   scanconfigwatcher.Config
	ScanWatcherConfig         scanwatcher.Config
	ScanResultWatcherConfig   scanresultwatcher.Config
	ScanResultProcessorConfig scanresultprocessor.Config
}

func setConfigDefaults(backendHost string, backendPort int, backendBaseURL string) {
	viper.SetDefault(ScannerAWSRegion, DefaultScannerAWSRegion)
	viper.SetDefault(ScanConfigWatchInterval, "30s")
	viper.SetDefault(DeleteJobPolicy, string(scanresultwatcher.DeleteJobPolicyAlways))
	viper.SetDefault(ScannerBackendAddress, fmt.Sprintf("http://%s%s", net.JoinHostPort(backendHost, strconv.Itoa(backendPort)), backendBaseURL))
	// https://github.com/openclarity/vmclarity-tools-base/blob/main/Dockerfile#L33
	viper.SetDefault(GitleaksBinaryPath, "/artifacts/gitleaks")
	// https://github.com/openclarity/vmclarity-tools-base/blob/main/Dockerfile#L35
	viper.SetDefault(LynisInstallPath, "/artifacts/lynis")
	// https://github.com/openclarity/vmclarity-tools-base/blob/main/Dockerfile
	viper.SetDefault(ChkrootkitBinaryPath, "/artifacts/chkrootkit")
	viper.SetDefault(ExploitDBAddress, fmt.Sprintf("http://%s", net.JoinHostPort(backendHost, "1326")))
	viper.SetDefault(AttachedVolumeDeviceName, DefaultAttachedVolumeDeviceName)
	viper.SetDefault(ClamBinaryPath, "clamscan")
	viper.SetDefault(FreshclamBinaryPath, "freshclam")
	viper.SetDefault(TrivyServerTimeout, DefaultTrivyServerTimeout)
	viper.SetDefault(GrypeServerTimeout, DefaultGrypeServerTimeout)
	viper.SetDefault(ScanConfigPollingInterval, scanconfigwatcher.DefaultPollInterval.String())
	viper.SetDefault(ScanConfigReconcileTimeout, scanconfigwatcher.DefaultReconcileTimeout.String())
	viper.SetDefault(ScanPollingInterval, scanwatcher.DefaultPollInterval.String())
	viper.SetDefault(ScanReconcileTimeout, scanwatcher.DefaultReconcileTimeout.String())
	viper.SetDefault(ScanTimeout, scanwatcher.DefaultScanTimeout.String())
	viper.SetDefault(ScanResultPollingInterval, scanresultwatcher.DefaultPollInterval.String())
	viper.SetDefault(ScanResultReconcileTimeout, scanresultwatcher.DefaultReconcileTimeout.String())
	viper.SetDefault(ScanResultProcessorPollingInterval, scanresultprocessor.DefaultPollInterval.String())
	viper.SetDefault(ScanResultProcessorReconcileTimeout, scanresultprocessor.DefaultReconcileTimeout.String())
	viper.SetDefault(DiscoveryInterval, discovery.DefaultInterval.String())

	viper.AutomaticEnv()
}

func LoadConfig(backendHost string, backendPort int, baseURL string) (*Config, error) {
	setConfigDefaults(backendHost, backendPort, baseURL)

	c := &Config{
		AWSConfig: aws.LoadConfig(),
		DiscoveryConfig: discovery.Config{
			DiscoveryInterval: viper.GetDuration(DiscoveryInterval),
		},
		ScannerBackendAddress: viper.GetString(ScannerBackendAddress),
		ScanConfigWatcherConfig: scanconfigwatcher.Config{
			PollPeriod:       viper.GetDuration(ScanConfigPollingInterval),
			ReconcileTimeout: viper.GetDuration(ScanConfigReconcileTimeout),
		},
		ScanWatcherConfig: scanwatcher.Config{
			PollPeriod:       viper.GetDuration(ScanPollingInterval),
			ReconcileTimeout: viper.GetDuration(ScanReconcileTimeout),
			ScanTimeout:      viper.GetDuration(ScanTimeout),
		},
		ScanResultWatcherConfig: scanresultwatcher.Config{
			PollPeriod:       viper.GetDuration(ScanResultPollingInterval),
			ReconcileTimeout: viper.GetDuration(ScanResultReconcileTimeout),
			ScannerConfig: scanresultwatcher.ScannerConfig{
				Region:                        viper.GetString(ScannerAWSRegion),
				DeleteJobPolicy:               scanresultwatcher.GetDeleteJobPolicyType(viper.GetString(DeleteJobPolicy)),
				ScannerImage:                  viper.GetString(ScannerContainerImage),
				ScannerBackendAddress:         viper.GetString(ScannerBackendAddress),
				ScannerKeyPairName:            viper.GetString(ScannerKeyPairName),
				GitleaksBinaryPath:            viper.GetString(GitleaksBinaryPath),
				LynisInstallPath:              viper.GetString(LynisInstallPath),
				DeviceName:                    viper.GetString(AttachedVolumeDeviceName),
				ExploitsDBAddress:             viper.GetString(ExploitDBAddress),
				ClamBinaryPath:                viper.GetString(ClamBinaryPath),
				FreshclamBinaryPath:           viper.GetString(FreshclamBinaryPath),
				AlternativeFreshclamMirrorURL: viper.GetString(AlternativeFreshclamMirrorURL),
				TrivyServerAddress:            viper.GetString(TrivyServerAddress),
				TrivyServerTimeout:            viper.GetDuration(TrivyServerTimeout),
				GrypeServerAddress:            viper.GetString(GrypeServerAddress),
				GrypeServerTimeout:            viper.GetDuration(GrypeServerTimeout),
				ChkrootkitBinaryPath:          viper.GetString(ChkrootkitBinaryPath),
			},
		},
		ScanResultProcessorConfig: scanresultprocessor.Config{
			PollPeriod:       viper.GetDuration(ScanResultProcessorPollingInterval),
			ReconcileTimeout: viper.GetDuration(ScanResultProcessorReconcileTimeout),
		},
	}

	return c, nil
}
