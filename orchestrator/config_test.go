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
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	apitypes "github.com/openclarity/vmclarity/api/types"

	"github.com/openclarity/vmclarity/orchestrator/discoverer"
	assetscanprocessor "github.com/openclarity/vmclarity/orchestrator/processor/assetscan"
	assetscanwatcher "github.com/openclarity/vmclarity/orchestrator/watcher/assetscan"
	assetscanestimationwatcher "github.com/openclarity/vmclarity/orchestrator/watcher/assetscanestimation"
	scanwatcher "github.com/openclarity/vmclarity/orchestrator/watcher/scan"
	scanconfigwatcher "github.com/openclarity/vmclarity/orchestrator/watcher/scanconfig"
	scanestimationwatcher "github.com/openclarity/vmclarity/orchestrator/watcher/scanestimation"
)

func TestUnmarshalCloudProvider(t *testing.T) {
	tests := []struct {
		Name              string
		CloudProviderText string

		ExpectedNewErrorMatcher types.GomegaMatcher
		ExpectedCloudProvider   apitypes.CloudProvider
	}{
		{
			Name:                    "AWS",
			CloudProviderText:       "aWs",
			ExpectedNewErrorMatcher: Not(HaveOccurred()),
			ExpectedCloudProvider:   apitypes.AWS,
		},
		{
			Name:                    "Azure",
			CloudProviderText:       "aZuRe",
			ExpectedNewErrorMatcher: Not(HaveOccurred()),
			ExpectedCloudProvider:   apitypes.Azure,
		},
		{
			Name:                    "Docker",
			CloudProviderText:       "dOcKer",
			ExpectedNewErrorMatcher: Not(HaveOccurred()),
			ExpectedCloudProvider:   apitypes.Docker,
		},
		{
			Name:                    "External",
			CloudProviderText:       "eXterNal",
			ExpectedNewErrorMatcher: Not(HaveOccurred()),
			ExpectedCloudProvider:   apitypes.External,
		},
		{
			Name:                    "GCP",
			CloudProviderText:       "GCP",
			ExpectedNewErrorMatcher: Not(HaveOccurred()),
			ExpectedCloudProvider:   apitypes.GCP,
		},
		{
			Name:                    "Kubernetes",
			CloudProviderText:       "kuBERnetes",
			ExpectedNewErrorMatcher: Not(HaveOccurred()),
			ExpectedCloudProvider:   apitypes.Kubernetes,
		},
		{
			Name:                    "Invalid",
			CloudProviderText:       "super awesome provider",
			ExpectedNewErrorMatcher: HaveOccurred(),
			ExpectedCloudProvider:   apitypes.CloudProvider("invalid"),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			provider := apitypes.CloudProvider("invalid")
			err := provider.UnmarshalText([]byte(test.CloudProviderText))

			g.Expect(err).Should(test.ExpectedNewErrorMatcher)
			g.Expect(provider).Should(BeEquivalentTo(test.ExpectedCloudProvider))
		})
	}
}

func TestConfig(t *testing.T) {
	tests := []struct {
		Name    string
		EnvVars map[string]string

		ExpectedNewErrorMatcher types.GomegaMatcher
		ExpectedConfig          *Config
	}{
		{
			Name: "Custom config",
			EnvVars: map[string]string{
				"VMCLARITY_ORCHESTRATOR_PROVIDER":                                           "docker",
				"VMCLARITY_ORCHESTRATOR_APISERVER_ADDRESS":                                  "http://example.com:8484/api",
				"VMCLARITY_ORCHESTRATOR_HEALTHCHECK_ADDRESS":                                "example.com:18888",
				"VMCLARITY_ORCHESTRATOR_CONTROLLER_STARTUP_DELAY":                           "15s",
				"VMCLARITY_ORCHESTRATOR_DISCOVERY_INTERVAL":                                 "60s",
				"VMCLARITY_ORCHESTRATOR_SCANCONFIG_WATCHER_POLL_PERIOD":                     "45s",
				"VMCLARITY_ORCHESTRATOR_SCANCONFIG_WATCHER_RECONCILE_TIMEOUT":               "9m",
				"VMCLARITY_ORCHESTRATOR_SCAN_WATCHER_POLL_PERIOD":                           "46s",
				"VMCLARITY_ORCHESTRATOR_SCAN_WATCHER_RECONCILE_TIMEOUT":                     "8m",
				"VMCLARITY_ORCHESTRATOR_SCAN_WATCHER_SCAN_TIMEOUT":                          "24h",
				"VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_POLL_PERIOD":                      "32s",
				"VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_RECONCILE_TIMEOUT":                "7m",
				"VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_ABORT_TIMEOUT":                    "12h",
				"VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_DELETE_POLICY":                    "never",
				"VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_APISERVER_ADDRESS":        "http://alternative.example.com:8484/api",
				"VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_EXPLOITSDB_ADDRESS":       "exploitsdb.example.com",
				"VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_TRIVY_SERVER_ADDRESS":     "trivy.example.com",
				"VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_TRIVY_SCAN_TIMEOUT":       "14m",
				"VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_GRYPE_SERVER_ADDRESS":     "grype.example.com",
				"VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_GRYPE_SERVER_TIMEOUT":     "17m",
				"VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_YARA_RULE_SERVER_ADDRESS": "yara.example.com",
				"VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_CONTAINER_IMAGE":          "super-scanner:latest",
				"VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_FRESHCLAM_MIRROR":         "freshclam.mirror.org",
				"VMCLARITY_ORCHESTRATOR_ASSETSCAN_ESTIMATION_WATCHER_POLL_PERIOD":           "44s",
				"VMCLARITY_ORCHESTRATOR_ASSETSCAN_ESTIMATION_WATCHER_RECONCILE_TIMEOUT":     "6m",
				"VMCLARITY_ORCHESTRATOR_SCAN_ESTIMATION_WATCHER_POLL_PERIOD":                "59s",
				"VMCLARITY_ORCHESTRATOR_SCAN_ESTIMATION_WATCHER_RECONCILE_TIMEOUT":          "5m",
				"VMCLARITY_ORCHESTRATOR_SCAN_ESTIMATION_WATCHER_ESTIMATION_TIMEOUT":         "23h",
				"VMCLARITY_ORCHESTRATOR_ASSETSCAN_PROCESSOR_POLL_PERIOD":                    "23s",
				"VMCLARITY_ORCHESTRATOR_ASSETSCAN_PROCESSOR_RECONCILE_TIMEOUT":              "4m",
			},
			ExpectedNewErrorMatcher: Not(HaveOccurred()),
			ExpectedConfig: &Config{
				ProviderKind:           apitypes.Docker,
				APIServerAddress:       "http://example.com:8484/api",
				HealthCheckAddress:     "example.com:18888",
				ControllerStartupDelay: 15 * time.Second,
				DiscoveryConfig: discoverer.Config{
					DiscoveryInterval: time.Minute,
				},
				ScanConfigWatcherConfig: scanconfigwatcher.Config{
					PollPeriod:       45 * time.Second,
					ReconcileTimeout: 9 * time.Minute,
				},
				ScanWatcherConfig: scanwatcher.Config{
					PollPeriod:       46 * time.Second,
					ReconcileTimeout: 8 * time.Minute,
					ScanTimeout:      24 * time.Hour,
				},
				AssetScanWatcherConfig: assetscanwatcher.Config{
					PollPeriod:       32 * time.Second,
					ReconcileTimeout: 7 * time.Minute,
					AbortTimeout:     12 * time.Hour,
					DeleteJobPolicy:  assetscanwatcher.DeleteJobPolicyNever,
					ScannerConfig: assetscanwatcher.ScannerConfig{
						APIServerAddress:              "http://alternative.example.com:8484/api",
						ExploitsDBAddress:             "exploitsdb.example.com",
						TrivyServerAddress:            "trivy.example.com",
						TrivyScanTimeout:              14 * time.Minute,
						GrypeServerAddress:            "grype.example.com",
						GrypeServerTimeout:            17 * time.Minute,
						YaraRuleServerAddress:         "yara.example.com",
						ScannerImage:                  "super-scanner:latest",
						AlternativeFreshclamMirrorURL: "freshclam.mirror.org",
					},
				},
				AssetScanEstimationWatcherConfig: assetscanestimationwatcher.Config{
					PollPeriod:       44 * time.Second,
					ReconcileTimeout: 6 * time.Minute,
				},
				ScanEstimationWatcherConfig: scanestimationwatcher.Config{
					PollPeriod:            59 * time.Second,
					ReconcileTimeout:      5 * time.Minute,
					ScanEstimationTimeout: 23 * time.Hour,
				},
				AssetScanProcessorConfig: assetscanprocessor.Config{
					PollPeriod:       23 * time.Second,
					ReconcileTimeout: 4 * time.Minute,
				},
			},
		},
		{
			Name:                    "Default config",
			EnvVars:                 map[string]string{},
			ExpectedNewErrorMatcher: Not(HaveOccurred()),
			ExpectedConfig: &Config{
				ProviderKind:           DefaultProviderKind,
				APIServerAddress:       "",
				HealthCheckAddress:     DefaultHealthCheckAddress,
				ControllerStartupDelay: DefaultControllerStartupDelay,
				DiscoveryConfig: discoverer.Config{
					DiscoveryInterval: discoverer.DefaultInterval,
				},
				ScanConfigWatcherConfig: scanconfigwatcher.Config{
					PollPeriod:       scanconfigwatcher.DefaultPollInterval,
					ReconcileTimeout: scanconfigwatcher.DefaultReconcileTimeout,
				},
				ScanWatcherConfig: scanwatcher.Config{
					PollPeriod:       scanwatcher.DefaultPollInterval,
					ReconcileTimeout: scanwatcher.DefaultReconcileTimeout,
					ScanTimeout:      scanwatcher.DefaultScanTimeout,
				},
				AssetScanWatcherConfig: assetscanwatcher.Config{
					PollPeriod:       assetscanwatcher.DefaultPollInterval,
					ReconcileTimeout: assetscanwatcher.DefaultReconcileTimeout,
					AbortTimeout:     assetscanwatcher.DefaultAbortTimeout,
					DeleteJobPolicy:  assetscanwatcher.DeleteJobPolicyAlways,
					ScannerConfig: assetscanwatcher.ScannerConfig{
						APIServerAddress:              "",
						ExploitsDBAddress:             "",
						TrivyServerAddress:            "",
						TrivyScanTimeout:              assetscanwatcher.DefaultTrivyScanTimeout,
						GrypeServerAddress:            "",
						GrypeServerTimeout:            assetscanwatcher.DefaultGrypeServerTimeout,
						YaraRuleServerAddress:         "",
						ScannerImage:                  "",
						AlternativeFreshclamMirrorURL: "",
					},
				},
				AssetScanEstimationWatcherConfig: assetscanestimationwatcher.Config{
					PollPeriod:       assetscanestimationwatcher.DefaultPollInterval,
					ReconcileTimeout: assetscanestimationwatcher.DefaultReconcileTimeout,
				},
				ScanEstimationWatcherConfig: scanestimationwatcher.Config{
					PollPeriod:            scanestimationwatcher.DefaultPollInterval,
					ReconcileTimeout:      scanestimationwatcher.DefaultReconcileTimeout,
					ScanEstimationTimeout: scanestimationwatcher.DefaultScanEstimationTimeout,
				},
				AssetScanProcessorConfig: assetscanprocessor.Config{
					PollPeriod:       assetscanprocessor.DefaultPollInterval,
					ReconcileTimeout: assetscanprocessor.DefaultReconcileTimeout,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			os.Clearenv()
			for k, v := range test.EnvVars {
				err := os.Setenv(k, v)
				g.Expect(err).Should(Not(HaveOccurred()))
			}

			config, err := NewConfig()

			g.Expect(err).Should(test.ExpectedNewErrorMatcher)
			g.Expect(config).Should(BeEquivalentTo(test.ExpectedConfig))
		})
	}
}
