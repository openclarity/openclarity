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

	"github.com/openclarity/vmclarity/testenv"
	awsenv "github.com/openclarity/vmclarity/testenv/aws"
	azureenv "github.com/openclarity/vmclarity/testenv/azure"
	dockerenv "github.com/openclarity/vmclarity/testenv/docker"
	gcpenv "github.com/openclarity/vmclarity/testenv/gcp"
	k8senv "github.com/openclarity/vmclarity/testenv/kubernetes"
	"github.com/openclarity/vmclarity/testenv/kubernetes/helm"
	k8senvtypes "github.com/openclarity/vmclarity/testenv/kubernetes/types"
	testenvtypes "github.com/openclarity/vmclarity/testenv/types"
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
				"VMCLARITY_E2E_USE_EXISTING":      "true",
				"VMCLARITY_E2E_ENV_SETUP_TIMEOUT": "45m",
				// testenv configuration
				"VMCLARITY_E2E_PLATFORM":                  "kuBERnetES",
				"VMCLARITY_E2E_ENV_NAME":                  "vmclarity-e2e-test",
				"VMCLARITY_E2E_APISERVER_IMAGE":           "openclarity/vmclarity-apiserver:test",
				"VMCLARITY_E2E_ORCHESTRATOR_IMAGE":        "openclarity/vmclarity-orchestrator:test",
				"VMCLARITY_E2E_UI_IMAGE":                  "openclarity/vmclarity-ui:test",
				"VMCLARITY_E2E_UIBACKEND_IMAGE":           "openclarity/vmclarity-uibackend:test",
				"VMCLARITY_E2E_SCANNER_IMAGE":             "openclarity/vmclarity-cli:test",
				"VMCLARITY_E2E_CR_DISCOVERY_SERVER_IMAGE": "openclarity/vmclarity-cr-discovery-server:test",
				"VMCLARITY_E2E_PLUGIN_KICS_IMAGE":         "openclarity/vmclarity-plugin-kics:test",
				// testenv.docker
				"VMCLARITY_E2E_DOCKER_COMPOSE_FILES": "docker-compose.yml,docker-compose-override.yml",
				// testenv.kubernetes
				"VMCLARITY_E2E_KUBERNETES_NAMESPACE":                "vmclarity-e2e-namespace",
				"VMCLARITY_E2E_KUBERNETES_PROVIDER":                 "eXTERnal",
				"VMCLARITY_E2E_KUBERNETES_CLUSTER_NAME":             "vmclarity-e2e-cluster",
				"VMCLARITY_E2E_KUBERNETES_CLUSTER_CREATION_TIMEOUT": "1h",
				"VMCLARITY_E2E_KUBERNETES_VERSION":                  "1.28",
				"VMCLARITY_E2E_KUBERNETES_KUBECONFIG":               "kubeconfig/default.yaml",
				// testenv.aws
				"VMCLARITY_E2E_AWS_REGION":           "us-west-2",
				"VMCLARITY_E2E_AWS_PRIVATE_KEY_FILE": "/home/vmclarity/.ssh/id_rsa",
				"VMCLARITY_E2E_AWS_PUBLIC_KEY_FILE":  "/home/vmclarity/.ssh/id_rsa.pub",
				// testenv.gcp
				"VMCLARITY_E2E_GCP_PRIVATE_KEY_FILE": "/home/vmclarity/.ssh/id_rsa",
				"VMCLARITY_E2E_GCP_PUBLIC_KEY_FILE":  "/home/vmclarity/.ssh/id_rsa.pub",
				// testenv.azure
				"VMCLARITY_E2E_AZURE_REGION":           "polandcentral",
				"VMCLARITY_E2E_AZURE_PRIVATE_KEY_FILE": "/home/vmclarity/.ssh/id_rsa",
				"VMCLARITY_E2E_AZURE_PUBLIC_KEY_FILE":  "/home/vmclarity/.ssh/id_rsa.pub",
			},
			ExpectedNewErrorMatcher: Not(HaveOccurred()),
			ExpectedConfig: &Config{
				ReuseEnv:        true,
				EnvSetupTimeout: 45 * time.Minute,
				TestEnvConfig: testenv.Config{
					Platform: testenvtypes.EnvironmentTypeKubernetes,
					EnvName:  "vmclarity-e2e-test",
					Images: testenvtypes.ContainerImages[string]{
						APIServer:         "openclarity/vmclarity-apiserver:test",
						Orchestrator:      "openclarity/vmclarity-orchestrator:test",
						UI:                "openclarity/vmclarity-ui:test",
						UIBackend:         "openclarity/vmclarity-uibackend:test",
						Scanner:           "openclarity/vmclarity-cli:test",
						CRDiscoveryServer: "openclarity/vmclarity-cr-discovery-server:test",
						PluginKics:        "openclarity/vmclarity-plugin-kics:test",
					},
					Docker: &dockerenv.Config{
						EnvName: "vmclarity-e2e-test",
						Images: testenvtypes.ContainerImages[string]{
							APIServer:         "openclarity/vmclarity-apiserver:test",
							Orchestrator:      "openclarity/vmclarity-orchestrator:test",
							UI:                "openclarity/vmclarity-ui:test",
							UIBackend:         "openclarity/vmclarity-uibackend:test",
							Scanner:           "openclarity/vmclarity-cli:test",
							CRDiscoveryServer: "openclarity/vmclarity-cr-discovery-server:test",
							PluginKics:        "openclarity/vmclarity-plugin-kics:test",
						},
						ComposeFiles: []string{
							"docker-compose.yml",
							"docker-compose-override.yml",
						},
					},
					Kubernetes: &k8senv.Config{
						EnvName:   "vmclarity-e2e-test",
						Namespace: "vmclarity-e2e-namespace",
						Provider:  k8senvtypes.ProviderTypeExternal,
						ProviderConfig: k8senvtypes.ProviderConfig{
							ClusterName:            "vmclarity-e2e-cluster",
							ClusterCreationTimeout: time.Hour,
							KubeConfigPath:         "kubeconfig/default.yaml",
							KubernetesVersion:      "1.28",
						},
						Installer: k8senvtypes.InstallerTypeHelm,
						HelmConfig: &helm.Config{
							Namespace:      "vmclarity-e2e-namespace",
							ReleaseName:    "",
							ChartPath:      "",
							StorageDriver:  "secret",
							KubeConfigPath: "",
						},
						Images: testenvtypes.ContainerImages[testenvtypes.ImageRef]{
							APIServer:         testenvtypes.NewImageRef("docker.io/openclarity/vmclarity-apiserver", "docker.io", "openclarity/vmclarity-apiserver", "test", ""),
							Orchestrator:      testenvtypes.NewImageRef("docker.io/openclarity/vmclarity-orchestrator", "docker.io", "openclarity/vmclarity-orchestrator", "test", ""),
							UI:                testenvtypes.NewImageRef("docker.io/openclarity/vmclarity-ui", "docker.io", "openclarity/vmclarity-ui", "test", ""),
							UIBackend:         testenvtypes.NewImageRef("docker.io/openclarity/vmclarity-uibackend", "docker.io", "openclarity/vmclarity-uibackend", "test", ""),
							Scanner:           testenvtypes.NewImageRef("docker.io/openclarity/vmclarity-cli", "docker.io", "openclarity/vmclarity-cli", "test", ""),
							CRDiscoveryServer: testenvtypes.NewImageRef("docker.io/openclarity/vmclarity-cr-discovery-server", "docker.io", "openclarity/vmclarity-cr-discovery-server", "test", ""),
							PluginKics:        testenvtypes.NewImageRef("docker.io/openclarity/vmclarity-plugin-kics", "docker.io", "openclarity/vmclarity-plugin-kics", "test", ""),
						},
					},
					AWS: &awsenv.Config{
						EnvName:        "vmclarity-e2e-test",
						Region:         "us-west-2",
						PrivateKeyFile: "/home/vmclarity/.ssh/id_rsa",
						PublicKeyFile:  "/home/vmclarity/.ssh/id_rsa.pub",
					},
					GCP: &gcpenv.Config{
						EnvName:        "vmclarity-e2e-test",
						PrivateKeyFile: "/home/vmclarity/.ssh/id_rsa",
						PublicKeyFile:  "/home/vmclarity/.ssh/id_rsa.pub",
					},
					Azure: &azureenv.Config{
						EnvName:        "vmclarity-e2e-test",
						Region:         "polandcentral",
						PrivateKeyFile: "/home/vmclarity/.ssh/id_rsa",
						PublicKeyFile:  "/home/vmclarity/.ssh/id_rsa.pub",
					},
				},
				TestSuiteParams: &TestSuiteParams{
					ServicesReadyTimeout: 5 * time.Minute,
					ScanTimeout:          5 * time.Minute,
					Scope:                "assetInfo/labels/any(t: t/key eq 'scanconfig' and t/value eq 'test') and assetInfo/containerName eq 'alpine'",
					FamiliesConfig:       kubernetesFamiliesConfig,
				},
			},
		},
		{
			Name: "Default config",
			EnvVars: map[string]string{
				"HOME": "/home/vmclarity",
			},
			ExpectedNewErrorMatcher: Not(HaveOccurred()),
			ExpectedConfig: &Config{
				ReuseEnv:        false,
				EnvSetupTimeout: DefaultEnvSetupTimeout,
				TestEnvConfig: testenv.Config{
					Platform: testenv.DefaultPlatform,
					EnvName:  testenv.DefaultEnvName,
					Images: testenvtypes.ContainerImages[string]{
						APIServer:         "ghcr.io/openclarity/vmclarity-apiserver:latest",
						Orchestrator:      "ghcr.io/openclarity/vmclarity-orchestrator:latest",
						UI:                "ghcr.io/openclarity/vmclarity-ui:latest",
						UIBackend:         "ghcr.io/openclarity/vmclarity-ui-backend:latest",
						Scanner:           "ghcr.io/openclarity/vmclarity-cli:latest",
						CRDiscoveryServer: "ghcr.io/openclarity/vmclarity-cr-discovery-server:latest",
						PluginKics:        "ghcr.io/openclarity/vmclarity-plugin-kics:latest",
					},
					Docker: &dockerenv.Config{
						EnvName: testenv.DefaultEnvName,
						Images: testenvtypes.ContainerImages[string]{
							APIServer:         "ghcr.io/openclarity/vmclarity-apiserver:latest",
							Orchestrator:      "ghcr.io/openclarity/vmclarity-orchestrator:latest",
							UI:                "ghcr.io/openclarity/vmclarity-ui:latest",
							UIBackend:         "ghcr.io/openclarity/vmclarity-ui-backend:latest",
							Scanner:           "ghcr.io/openclarity/vmclarity-cli:latest",
							CRDiscoveryServer: "ghcr.io/openclarity/vmclarity-cr-discovery-server:latest",
							PluginKics:        "ghcr.io/openclarity/vmclarity-plugin-kics:latest",
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
							KubeConfigPath:         "/home/vmclarity/.kube/config",
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
							APIServer:         testenvtypes.NewImageRef("ghcr.io/openclarity/vmclarity-apiserver", "ghcr.io", "openclarity/vmclarity-apiserver", "latest", ""),
							Orchestrator:      testenvtypes.NewImageRef("ghcr.io/openclarity/vmclarity-orchestrator", "ghcr.io", "openclarity/vmclarity-orchestrator", "latest", ""),
							UI:                testenvtypes.NewImageRef("ghcr.io/openclarity/vmclarity-ui", "ghcr.io", "openclarity/vmclarity-ui", "latest", ""),
							UIBackend:         testenvtypes.NewImageRef("ghcr.io/openclarity/vmclarity-ui-backend", "ghcr.io", "openclarity/vmclarity-ui-backend", "latest", ""),
							Scanner:           testenvtypes.NewImageRef("ghcr.io/openclarity/vmclarity-cli", "ghcr.io", "openclarity/vmclarity-cli", "latest", ""),
							CRDiscoveryServer: testenvtypes.NewImageRef("ghcr.io/openclarity/vmclarity-cr-discovery-server", "ghcr.io", "openclarity/vmclarity-cr-discovery-server", "latest", ""),
							PluginKics:        testenvtypes.NewImageRef("ghcr.io/openclarity/vmclarity-plugin-kics", "ghcr.io", "openclarity/vmclarity-plugin-kics", "latest", ""),
						},
					},
					AWS: &awsenv.Config{
						EnvName:        "vmclarity-testenv",
						Region:         "eu-central-1",
						PrivateKeyFile: "",
						PublicKeyFile:  "",
					},
					GCP: &gcpenv.Config{
						EnvName:        "vmclarity-testenv",
						PrivateKeyFile: "",
						PublicKeyFile:  "",
					},
					Azure: &azureenv.Config{
						EnvName:        "vmclarity-testenv",
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
