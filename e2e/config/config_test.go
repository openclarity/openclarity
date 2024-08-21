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

package config

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/openclarity/openclarity/e2e/benchmark"
	"github.com/openclarity/openclarity/testenv"
	awsenv "github.com/openclarity/openclarity/testenv/aws"
	azureenv "github.com/openclarity/openclarity/testenv/azure"
	dockerenv "github.com/openclarity/openclarity/testenv/docker"
	gcpenv "github.com/openclarity/openclarity/testenv/gcp"
	k8senv "github.com/openclarity/openclarity/testenv/kubernetes"
	"github.com/openclarity/openclarity/testenv/kubernetes/helm"
	k8senvtypes "github.com/openclarity/openclarity/testenv/kubernetes/types"
	testenvtypes "github.com/openclarity/openclarity/testenv/types"
)

func TestConfig(t *testing.T) {
	kubernetesFamiliesConfig := FullScanFamiliesConfig
	kubernetesFamiliesConfig.Sbom.Analyzers = &[]string{"trivy", "windows"}

	tests := []struct {
		Name    string
		EnvVars map[string]string

		ExpectedNewErrorMatcher types.GomegaMatcher
		ExpectedConfig          *Config
	}{
		{
			Name: "Custom config",
			EnvVars: map[string]string{
				// E2E configuration
				"OPENCLARITY_E2E_USE_EXISTING":      "true",
				"OPENCLARITY_E2E_ENV_SETUP_TIMEOUT": "45m",
				// testenv configuration
				"OPENCLARITY_E2E_PLATFORM":                  "kuBERnetES",
				"OPENCLARITY_E2E_ENV_NAME":                  "openclarity-e2e-test",
				"OPENCLARITY_E2E_APISERVER_IMAGE":           "openclarity/openclarity-api-server:test",
				"OPENCLARITY_E2E_ORCHESTRATOR_IMAGE":        "openclarity/openclarity-orchestrator:test",
				"OPENCLARITY_E2E_UI_IMAGE":                  "openclarity/openclarity-ui:test",
				"OPENCLARITY_E2E_UIBACKEND_IMAGE":           "openclarity/openclarity-uibackend:test",
				"OPENCLARITY_E2E_SCANNER_IMAGE":             "openclarity/openclarity-cli:test",
				"OPENCLARITY_E2E_CR_DISCOVERY_SERVER_IMAGE": "openclarity/openclarity-cr-discovery-server:test",
				"OPENCLARITY_E2E_PLUGIN_KICS_IMAGE":         "openclarity/openclarity-plugin-kics:test",
				// testenv.docker
				"OPENCLARITY_E2E_DOCKER_COMPOSE_FILES": "docker-compose.yml,docker-compose-override.yml",
				// testenv.kubernetes
				"OPENCLARITY_E2E_KUBERNETES_NAMESPACE":                "openclarity-e2e-namespace",
				"OPENCLARITY_E2E_KUBERNETES_PROVIDER":                 "eXTERnal",
				"OPENCLARITY_E2E_KUBERNETES_CLUSTER_NAME":             "openclarity-e2e-cluster",
				"OPENCLARITY_E2E_KUBERNETES_CLUSTER_CREATION_TIMEOUT": "1h",
				"OPENCLARITY_E2E_KUBERNETES_VERSION":                  "1.28",
				"OPENCLARITY_E2E_KUBERNETES_KUBECONFIG":               "kubeconfig/default.yaml",
				// testenv.aws
				"OPENCLARITY_E2E_AWS_REGION":           "us-west-2",
				"OPENCLARITY_E2E_AWS_PRIVATE_KEY_FILE": "/home/openclarity/.ssh/id_rsa",
				"OPENCLARITY_E2E_AWS_PUBLIC_KEY_FILE":  "/home/openclarity/.ssh/id_rsa.pub",
				// testenv.gcp
				"OPENCLARITY_E2E_GCP_PRIVATE_KEY_FILE": "/home/openclarity/.ssh/id_rsa",
				"OPENCLARITY_E2E_GCP_PUBLIC_KEY_FILE":  "/home/openclarity/.ssh/id_rsa.pub",
				// testenv.azure
				"OPENCLARITY_E2E_AZURE_REGION":           "polandcentral",
				"OPENCLARITY_E2E_AZURE_PRIVATE_KEY_FILE": "/home/openclarity/.ssh/id_rsa",
				"OPENCLARITY_E2E_AZURE_PUBLIC_KEY_FILE":  "/home/openclarity/.ssh/id_rsa.pub",
				// Benchmark configuration
				"OPENCLARITY_E2E_BENCHMARK_ENABLED":          "false",
				"OPENCLARITY_E2E_BENCHMARK_OUTPUT_FILE_PATH": "/home/openclarity/scanner-benchmark.md",
			},
			ExpectedNewErrorMatcher: Not(HaveOccurred()),
			ExpectedConfig: &Config{
				ReuseEnv:        true,
				EnvSetupTimeout: 45 * time.Minute,
				TestEnvConfig: testenv.Config{
					Platform: testenvtypes.EnvironmentTypeKubernetes,
					EnvName:  "openclarity-e2e-test",
					Images: testenvtypes.ContainerImages[string]{
						APIServer:         "openclarity/openclarity-api-server:test",
						Orchestrator:      "openclarity/openclarity-orchestrator:test",
						UI:                "openclarity/openclarity-ui:test",
						UIBackend:         "openclarity/openclarity-uibackend:test",
						Scanner:           "openclarity/openclarity-cli:test",
						CRDiscoveryServer: "openclarity/openclarity-cr-discovery-server:test",
						PluginKics:        "openclarity/openclarity-plugin-kics:test",
					},
					Docker: &dockerenv.Config{
						EnvName: "openclarity-e2e-test",
						Images: testenvtypes.ContainerImages[string]{
							APIServer:         "openclarity/openclarity-api-server:test",
							Orchestrator:      "openclarity/openclarity-orchestrator:test",
							UI:                "openclarity/openclarity-ui:test",
							UIBackend:         "openclarity/openclarity-uibackend:test",
							Scanner:           "openclarity/openclarity-cli:test",
							CRDiscoveryServer: "openclarity/openclarity-cr-discovery-server:test",
							PluginKics:        "openclarity/openclarity-plugin-kics:test",
						},
						ComposeFiles: []string{
							"docker-compose.yml",
							"docker-compose-override.yml",
						},
					},
					Kubernetes: &k8senv.Config{
						EnvName:   "openclarity-e2e-test",
						Namespace: "openclarity-e2e-namespace",
						Provider:  k8senvtypes.ProviderTypeExternal,
						ProviderConfig: k8senvtypes.ProviderConfig{
							ClusterName:            "openclarity-e2e-cluster",
							ClusterCreationTimeout: time.Hour,
							KubeConfigPath:         "kubeconfig/default.yaml",
							KubernetesVersion:      "1.28",
						},
						Installer: k8senvtypes.InstallerTypeHelm,
						HelmConfig: &helm.Config{
							Namespace:      "openclarity-e2e-namespace",
							ReleaseName:    "",
							ChartPath:      "",
							StorageDriver:  "secret",
							KubeConfigPath: "",
						},
						Images: testenvtypes.ContainerImages[testenvtypes.ImageRef]{
							APIServer:         testenvtypes.NewImageRef("docker.io/openclarity/openclarity-api-server", "docker.io", "openclarity/openclarity-api-server", "test", ""),
							Orchestrator:      testenvtypes.NewImageRef("docker.io/openclarity/openclarity-orchestrator", "docker.io", "openclarity/openclarity-orchestrator", "test", ""),
							UI:                testenvtypes.NewImageRef("docker.io/openclarity/openclarity-ui", "docker.io", "openclarity/openclarity-ui", "test", ""),
							UIBackend:         testenvtypes.NewImageRef("docker.io/openclarity/openclarity-uibackend", "docker.io", "openclarity/openclarity-uibackend", "test", ""),
							Scanner:           testenvtypes.NewImageRef("docker.io/openclarity/openclarity-cli", "docker.io", "openclarity/openclarity-cli", "test", ""),
							CRDiscoveryServer: testenvtypes.NewImageRef("docker.io/openclarity/openclarity-cr-discovery-server", "docker.io", "openclarity/openclarity-cr-discovery-server", "test", ""),
							PluginKics:        testenvtypes.NewImageRef("docker.io/openclarity/openclarity-plugin-kics", "docker.io", "openclarity/openclarity-plugin-kics", "test", ""),
						},
					},
					AWS: &awsenv.Config{
						EnvName:        "openclarity-e2e-test",
						Region:         "us-west-2",
						PrivateKeyFile: "/home/openclarity/.ssh/id_rsa",
						PublicKeyFile:  "/home/openclarity/.ssh/id_rsa.pub",
					},
					GCP: &gcpenv.Config{
						EnvName:        "openclarity-e2e-test",
						PrivateKeyFile: "/home/openclarity/.ssh/id_rsa",
						PublicKeyFile:  "/home/openclarity/.ssh/id_rsa.pub",
					},
					Azure: &azureenv.Config{
						EnvName:        "openclarity-e2e-test",
						Region:         "polandcentral",
						PrivateKeyFile: "/home/openclarity/.ssh/id_rsa",
						PublicKeyFile:  "/home/openclarity/.ssh/id_rsa.pub",
					},
				},
				TestSuiteParams: &TestSuiteParams{
					ServicesReadyTimeout: 5 * time.Minute,
					ScanTimeout:          5 * time.Minute,
					Scope:                "assetInfo/labels/any(t: t/key eq 'scanconfig' and t/value eq 'test') and assetInfo/containerName eq 'alpine'",
					FamiliesConfig:       kubernetesFamiliesConfig,
				},
				BenchmarkConfig: benchmark.Config{
					Enabled:        false,
					OutputFilePath: "/home/openclarity/scanner-benchmark.md",
				},
			},
		},
		{
			Name: "Default config",
			EnvVars: map[string]string{
				"HOME": "/home/openclarity",
			},
			ExpectedNewErrorMatcher: Not(HaveOccurred()),
			ExpectedConfig: &Config{
				ReuseEnv:        false,
				EnvSetupTimeout: DefaultEnvSetupTimeout,
				TestEnvConfig: testenv.Config{
					Platform: testenv.DefaultPlatform,
					EnvName:  testenv.DefaultEnvName,
					Images: testenvtypes.ContainerImages[string]{
						APIServer:         "ghcr.io/openclarity/openclarity-api-server:latest",
						Orchestrator:      "ghcr.io/openclarity/openclarity-orchestrator:latest",
						UI:                "ghcr.io/openclarity/openclarity-ui:latest",
						UIBackend:         "ghcr.io/openclarity/openclarity-ui-backend:latest",
						Scanner:           "ghcr.io/openclarity/openclarity-cli:latest",
						CRDiscoveryServer: "ghcr.io/openclarity/openclarity-cr-discovery-server:latest",
						PluginKics:        "ghcr.io/openclarity/openclarity-plugin-kics:latest",
					},
					Docker: &dockerenv.Config{
						EnvName: testenv.DefaultEnvName,
						Images: testenvtypes.ContainerImages[string]{
							APIServer:         "ghcr.io/openclarity/openclarity-api-server:latest",
							Orchestrator:      "ghcr.io/openclarity/openclarity-orchestrator:latest",
							UI:                "ghcr.io/openclarity/openclarity-ui:latest",
							UIBackend:         "ghcr.io/openclarity/openclarity-ui-backend:latest",
							Scanner:           "ghcr.io/openclarity/openclarity-cli:latest",
							CRDiscoveryServer: "ghcr.io/openclarity/openclarity-cr-discovery-server:latest",
							PluginKics:        "ghcr.io/openclarity/openclarity-plugin-kics:latest",
						},
						ComposeFiles: nil,
					},
					Kubernetes: &k8senv.Config{
						EnvName:   testenv.DefaultEnvName,
						Namespace: k8senv.DefaultNamespace,
						Provider:  k8senvtypes.ProviderTypeKind,
						ProviderConfig: k8senvtypes.ProviderConfig{
							ClusterName:            testenv.DefaultEnvName,
							ClusterCreationTimeout: k8senv.DefaultClusterCreationTimeout,
							KubeConfigPath:         "/home/openclarity/.kube/config",
							KubernetesVersion:      k8senv.DefaultKubernetesVersion,
						},
						Installer: k8senvtypes.InstallerTypeHelm,
						HelmConfig: &helm.Config{
							Namespace:      "default",
							ReleaseName:    "",
							ChartPath:      "",
							StorageDriver:  "secret",
							KubeConfigPath: "",
						},
						Images: testenvtypes.ContainerImages[testenvtypes.ImageRef]{
							APIServer:         testenvtypes.NewImageRef("ghcr.io/openclarity/openclarity-api-server", "ghcr.io", "openclarity/openclarity-api-server", "latest", ""),
							Orchestrator:      testenvtypes.NewImageRef("ghcr.io/openclarity/openclarity-orchestrator", "ghcr.io", "openclarity/openclarity-orchestrator", "latest", ""),
							UI:                testenvtypes.NewImageRef("ghcr.io/openclarity/openclarity-ui", "ghcr.io", "openclarity/openclarity-ui", "latest", ""),
							UIBackend:         testenvtypes.NewImageRef("ghcr.io/openclarity/openclarity-ui-backend", "ghcr.io", "openclarity/openclarity-ui-backend", "latest", ""),
							Scanner:           testenvtypes.NewImageRef("ghcr.io/openclarity/openclarity-cli", "ghcr.io", "openclarity/openclarity-cli", "latest", ""),
							CRDiscoveryServer: testenvtypes.NewImageRef("ghcr.io/openclarity/openclarity-cr-discovery-server", "ghcr.io", "openclarity/openclarity-cr-discovery-server", "latest", ""),
							PluginKics:        testenvtypes.NewImageRef("ghcr.io/openclarity/openclarity-plugin-kics", "ghcr.io", "openclarity/openclarity-plugin-kics", "latest", ""),
						},
					},
					AWS: &awsenv.Config{
						EnvName:        "openclarity-testenv",
						Region:         "eu-central-1",
						PrivateKeyFile: "",
						PublicKeyFile:  "",
					},
					GCP: &gcpenv.Config{
						EnvName:        "openclarity-testenv",
						PrivateKeyFile: "",
						PublicKeyFile:  "",
					},
					Azure: &azureenv.Config{
						EnvName:        "openclarity-testenv",
						Region:         "eastus",
						PrivateKeyFile: "",
						PublicKeyFile:  "",
					},
				},
				TestSuiteParams: &TestSuiteParams{
					ServicesReadyTimeout: 5 * time.Minute,
					ScanTimeout:          5 * time.Minute,
					Scope:                "assetInfo/labels/any(t: t/key eq 'scanconfig' and t/value eq 'test')",
					FamiliesConfig:       FullScanFamiliesConfig,
				},
				BenchmarkConfig: benchmark.Config{
					Enabled:        true,
					OutputFilePath: "/tmp/scanner-benchmark.md",
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
