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

	"github.com/openclarity/vmclarity/e2e/testenv"
	dockerenv "github.com/openclarity/vmclarity/e2e/testenv/docker"
	k8senv "github.com/openclarity/vmclarity/e2e/testenv/kubernetes"
	"github.com/openclarity/vmclarity/e2e/testenv/kubernetes/helm"
	k8senvtypes "github.com/openclarity/vmclarity/e2e/testenv/kubernetes/types"
	testenvtypes "github.com/openclarity/vmclarity/e2e/testenv/types"
)

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
				// E2E configuration
				"VMCLARITY_E2E_USE_EXISTING":      "true",
				"VMCLARITY_E2E_ENV_SETUP_TIMEOUT": "45m",
				// testenv configuration
				"VMCLARITY_E2E_PLATFORM":           "kuBERnetES",
				"VMCLARITY_E2E_ENV_NAME":           "vmclarity-e2e-test",
				"VMCLARITY_E2E_APISERVER_IMAGE":    "openclarity/vmclarity-apiserver:latest",
				"VMCLARITY_E2E_ORCHESTRATOR_IMAGE": "openclarity/vmclarity-orchestrator:latest",
				"VMCLARITY_E2E_UI_IMAGE":           "openclarity/vmclarity-ui:latest",
				"VMCLARITY_E2E_UIBACKEND_IMAGE":    "openclarity/vmclarity-uibackend:latest",
				"VMCLARITY_E2E_SCANNER_IMAGE":      "openclarity/vmclarity-cli:latest",
				// testenv.docker
				"VMCLARITY_E2E_DOCKER_COMPOSE_FILES": "docker-compose.yml,docker-compose-override.yml",
				// testenv.kubernetes
				"VMCLARITY_E2E_KUBERNETES_NAMESPACE":                "vmclarity-e2e-namespace",
				"VMCLARITY_E2E_KUBERNETES_PROVIDER":                 "eXTERnal",
				"VMCLARITY_E2E_KUBERNETES_CLUSTER_NAME":             "vmclarity-e2e-cluster",
				"VMCLARITY_E2E_KUBERNETES_CLUSTER_CREATION_TIMEOUT": "1h",
				"VMCLARITY_E2E_KUBERNETES_VERSION":                  "1.28",
				"VMCLARITY_E2E_KUBERNETES_KUBECONFIG":               "kubeconfig/default.yaml",
			},
			ExpectedNewErrorMatcher: Not(HaveOccurred()),
			ExpectedConfig: &Config{
				ReuseEnv:        true,
				EnvSetupTimeout: 45 * time.Minute,
				TestEnvConfig: testenv.Config{
					Platform: testenvtypes.EnvironmentTypeKubernetes,
					EnvName:  "vmclarity-e2e-test",
					Images: testenvtypes.ContainerImages[string]{
						APIServer:    "openclarity/vmclarity-apiserver:latest",
						Orchestrator: "openclarity/vmclarity-orchestrator:latest",
						UI:           "openclarity/vmclarity-ui:latest",
						UIBackend:    "openclarity/vmclarity-uibackend:latest",
						Scanner:      "openclarity/vmclarity-cli:latest",
					},
					Docker: &dockerenv.Config{
						EnvName: "vmclarity-e2e-test",
						Images: testenvtypes.ContainerImages[string]{
							APIServer:    "openclarity/vmclarity-apiserver:latest",
							Orchestrator: "openclarity/vmclarity-orchestrator:latest",
							UI:           "openclarity/vmclarity-ui:latest",
							UIBackend:    "openclarity/vmclarity-uibackend:latest",
							Scanner:      "openclarity/vmclarity-cli:latest",
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
							APIServer:    testenvtypes.NewImageRef("docker.io/openclarity/vmclarity-apiserver", "docker.io", "openclarity/vmclarity-apiserver", "latest", ""),
							Orchestrator: testenvtypes.NewImageRef("docker.io/openclarity/vmclarity-orchestrator", "docker.io", "openclarity/vmclarity-orchestrator", "latest", ""),
							UI:           testenvtypes.NewImageRef("docker.io/openclarity/vmclarity-ui", "docker.io", "openclarity/vmclarity-ui", "latest", ""),
							UIBackend:    testenvtypes.NewImageRef("docker.io/openclarity/vmclarity-uibackend", "docker.io", "openclarity/vmclarity-uibackend", "latest", ""),
							Scanner:      testenvtypes.NewImageRef("docker.io/openclarity/vmclarity-cli", "docker.io", "openclarity/vmclarity-cli", "latest", ""),
						},
					},
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
						APIServer:    "ghcr.io/openclarity/vmclarity-apiserver:latest",
						Orchestrator: "ghcr.io/openclarity/vmclarity-orchestrator:latest",
						UI:           "ghcr.io/openclarity/vmclarity-ui:latest",
						UIBackend:    "ghcr.io/openclarity/vmclarity-ui-backend:latest",
						Scanner:      "ghcr.io/openclarity/vmclarity-cli:latest",
					},
					Docker: &dockerenv.Config{
						EnvName: testenv.DefaultEnvName,
						Images: testenvtypes.ContainerImages[string]{
							APIServer:    "ghcr.io/openclarity/vmclarity-apiserver:latest",
							Orchestrator: "ghcr.io/openclarity/vmclarity-orchestrator:latest",
							UI:           "ghcr.io/openclarity/vmclarity-ui:latest",
							UIBackend:    "ghcr.io/openclarity/vmclarity-ui-backend:latest",
							Scanner:      "ghcr.io/openclarity/vmclarity-cli:latest",
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
							APIServer:    testenvtypes.NewImageRef("ghcr.io/openclarity/vmclarity-apiserver", "ghcr.io", "openclarity/vmclarity-apiserver", "latest", ""),
							Orchestrator: testenvtypes.NewImageRef("ghcr.io/openclarity/vmclarity-orchestrator", "ghcr.io", "openclarity/vmclarity-orchestrator", "latest", ""),
							UI:           testenvtypes.NewImageRef("ghcr.io/openclarity/vmclarity-ui", "ghcr.io", "openclarity/vmclarity-ui", "latest", ""),
							UIBackend:    testenvtypes.NewImageRef("ghcr.io/openclarity/vmclarity-ui-backend", "ghcr.io", "openclarity/vmclarity-ui-backend", "latest", ""),
							Scanner:      testenvtypes.NewImageRef("ghcr.io/openclarity/vmclarity-cli", "ghcr.io", "openclarity/vmclarity-cli", "latest", ""),
						},
					},
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
