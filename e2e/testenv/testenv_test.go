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

package testenv

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/e2e/testenv/docker"
	"github.com/openclarity/vmclarity/e2e/testenv/kubernetes"
	"github.com/openclarity/vmclarity/e2e/testenv/kubernetes/helm"
	"github.com/openclarity/vmclarity/e2e/testenv/kubernetes/types"
	envtypes "github.com/openclarity/vmclarity/e2e/testenv/types"
	"github.com/openclarity/vmclarity/e2e/testenv/utils"
)

var ImageTag = utils.GetEnvOrDefault("VERSION", "latest")

var DockerContainerImages = docker.ContainerImages{
	APIServer:    "ghcr.io/openclarity/vmclarity-apiserver:" + ImageTag,
	Orchestrator: "ghcr.io/openclarity/vmclarity-orchestrator:" + ImageTag,
	UI:           "ghcr.io/openclarity/vmclarity-ui:" + ImageTag,
	UIBackend:    "ghcr.io/openclarity/vmclarity-ui-backend:" + ImageTag,
	Scanner:      "ghcr.io/openclarity/vmclarity-cli:" + ImageTag,
}

var KubernetesContainerImages = kubernetes.ContainerImages{
	APIServer:         envtypes.NewImageRef("ghcr.io/openclarity/vmclarity-apiserver", "ghcr.io", "openclarity/vmclarity-apiserver", ImageTag, ""),
	Orchestrator:      envtypes.NewImageRef("ghcr.io/openclarity/vmclarity-orchestrator", "ghcr.io", "openclarity/vmclarity-orchestrator", ImageTag, ""),
	UI:                envtypes.NewImageRef("ghcr.io/openclarity/vmclarity-ui", "ghcr.io", "openclarity/vmclarity-ui", ImageTag, ""),
	UIBackend:         envtypes.NewImageRef("ghcr.io/openclarity/vmclarity-ui-backend", "ghcr.io", "openclarity/vmclarity-ui-backend", ImageTag, ""),
	Scanner:           envtypes.NewImageRef("ghcr.io/openclarity/vmclarity-cli", "ghcr.io", "openclarity/vmclarity-cli", ImageTag, ""),
	CRDiscoveryServer: envtypes.NewImageRef("ghcr.io/openclarity/vmclarity-cr-discovery-server", "ghcr.io", "openclarity/vmclarity-cr-discovery-server", ImageTag, ""),
}

func TestTestEnv(t *testing.T) {
	tests := []struct {
		Name    string
		Timeout time.Duration

		TestEnvConfig    *Config
		ExpectedServices envtypes.Services
	}{
		{
			Name:    "Kind cluster with Helm installer and embedded Chart",
			Timeout: 30 * time.Minute,
			TestEnvConfig: &Config{
				Platform: "kubernetes",
				EnvName:  "testenv-k8s-test",
				Kubernetes: &kubernetes.Config{
					EnvName:   "testenv-k8s-test",
					Namespace: "default",
					Provider:  "kind",
					Installer: "helm",
					HelmConfig: &helm.Config{
						Namespace:     "default",
						ReleaseName:   "testenv-k8s-test",
						StorageDriver: "secret",
					},
					ProviderConfig: types.ProviderConfig{
						ClusterName:            "testenv-k8s-test",
						ClusterCreationTimeout: 15 * time.Minute,
						KubernetesVersion:      "1.27",
					},
					Images: KubernetesContainerImages,
				},
			},
			ExpectedServices: NewKubernetesServices("default", "testenv-k8s-test"),
		},
		{
			Name:    "Docker Compose with embedded Manifests",
			Timeout: 10 * time.Minute,
			TestEnvConfig: &Config{
				Platform: "docker",
				EnvName:  "testenv-docker-test",
				Docker: &docker.Config{
					EnvName: "testenv-docker-test",
					Images:  DockerContainerImages,
				},
			},
			ExpectedServices: NewDockerServices("testenv-docker-test"),
		},
	}

	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			ctx, cancel := context.WithTimeout(context.Background(), test.Timeout)

			// Cancel test context at the end of test run
			t.Cleanup(func() {
				if cancel != nil {
					cancel()
				}
			})

			log := logrus.StandardLogger()
			log.SetLevel(logrus.DebugLevel)
			logger := logrus.NewEntry(log)
			ctx = utils.SetLoggerForContext(ctx, logger)

			workDir := t.TempDir()
			t.Logf("work dir: %s", workDir)

			t.Log("initializing test environment...")
			env, err := New(test.TestEnvConfig,
				WithLogger(logger),
				WithContext(ctx),
				WithWorkDir(workDir),
			)
			g.Expect(err).Should(Not(HaveOccurred()))

			t.Log("setting up test environment...")
			err = env.SetUp(ctx)
			g.Expect(err).Should(Not(HaveOccurred()))

			// Tear down the testenv at the end of test run
			t.Cleanup(func() {
				if env != nil {
					t.Log("tearing down test environment...")
					if err := env.TearDown(ctx); err != nil {
						t.Logf("failed to cleanup testenv: %v", err)
					}
				}
			})

			t.Log("getting list of services in environment...")
			services, err := env.Services(ctx)
			g.Expect(err).Should(Not(HaveOccurred()))

			ok := AssertServicesContains(services, test.ExpectedServices, ServiceHashEquals)
			g.Expect(ok).Should(BeTrue())

			if services != nil {
				t.Log("waiting for services become ready...")
				g.Eventually(func() bool {
					ready, err := env.ServicesReady(ctx)
					g.Expect(err).Should(Not(HaveOccurred()))
					return ready
				}).Should(BeTrue())
			}

			t.Log("verify API endpoint...")
			endpoints, err := env.Endpoints(ctx)
			g.Expect(err).Should(Not(HaveOccurred()))

			reqURL, err := url.JoinPath(endpoints.API.String(), "scanConfigs")
			g.Expect(err).Should(Not(HaveOccurred()))

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
			g.Expect(err).Should(Not(HaveOccurred()))

			resp, err := http.DefaultClient.Do(req)
			g.Expect(err).Should(Not(HaveOccurred()))
			defer resp.Body.Close()

			g.Expect(resp.StatusCode).Should(BeNumerically("==", http.StatusOK))

			t.Log("tearing down test environment...")
			err = env.TearDown(ctx)
			g.Expect(err).Should(Not(HaveOccurred()))
		})
	}
}

// ServiceComparisonFn function for comparing envtypes.Service implementations.
type ServiceComparisonFn = func(left, right envtypes.Service) bool

func ServiceHashEquals(left, right envtypes.Service) bool {
	hash := func(s envtypes.Service) string {
		return s.GetID() + s.GetNamespace() + s.GetApplicationName() + s.GetComponentName()
	}

	return hash(left) == hash(right)
}

// AssertServicesContains returns true if left contains Service objects from right determined by fn.
func AssertServicesContains(left, right envtypes.Services, fn ServiceComparisonFn) bool {
	var result bool

	leftServiceMap := make(map[string]envtypes.Service, len(left))
	for _, l := range left {
		leftServiceMap[l.GetID()] = l
	}

	for _, r := range right {
		l, ok := leftServiceMap[r.GetID()]
		if !ok {
			result = false
			break
		}

		if fn(l, r) {
			result = true
		}
	}

	return result
}

func NewKubernetesServices(namespace, release string) envtypes.Services {
	return envtypes.Services{
		&kubernetes.Service{
			ID:          release + "-vmclarity-apiserver",
			Namespace:   namespace,
			Application: "vmclarity",
			Component:   "apiserver",
			State:       envtypes.ServiceStateReady,
		},
		&kubernetes.Service{
			ID:          release + "-vmclarity-exploit-db-server",
			Namespace:   namespace,
			Application: "vmclarity",
			Component:   "exploit-db-server",
			State:       envtypes.ServiceStateReady,
		},
		&kubernetes.Service{
			ID:          release + "-vmclarity-freshclam-mirror",
			Namespace:   namespace,
			Application: "vmclarity",
			Component:   "freshclam-mirror",
			State:       envtypes.ServiceStateReady,
		},
		&kubernetes.Service{
			ID:          release + "-vmclarity-gateway",
			Namespace:   namespace,
			Application: "vmclarity",
			Component:   "gateway",
			State:       envtypes.ServiceStateReady,
		},
		&kubernetes.Service{
			ID:          release + "-vmclarity-grype-server",
			Namespace:   namespace,
			Application: "vmclarity",
			Component:   "grype-server",
			State:       envtypes.ServiceStateReady,
		},
		&kubernetes.Service{
			ID:          release + "-vmclarity-orchestrator",
			Namespace:   namespace,
			Application: "vmclarity",
			Component:   "orchestrator",
			State:       envtypes.ServiceStateReady,
		},
		&kubernetes.Service{
			ID:          release + "-vmclarity-swagger-ui",
			Namespace:   namespace,
			Application: "vmclarity",
			Component:   "swagger-ui",
			State:       envtypes.ServiceStateReady,
		},
		&kubernetes.Service{
			ID:          release + "-vmclarity-trivy-server",
			Namespace:   namespace,
			Application: "vmclarity",
			Component:   "trivy-server",
			State:       envtypes.ServiceStateReady,
		},
		&kubernetes.Service{
			ID:          release + "-vmclarity-ui",
			Namespace:   namespace,
			Application: "vmclarity",
			Component:   "ui",
			State:       envtypes.ServiceStateReady,
		},
		&kubernetes.Service{
			ID:          release + "-vmclarity-uibackend",
			Namespace:   namespace,
			Application: "vmclarity",
			Component:   "uibackend",
			State:       envtypes.ServiceStateReady,
		},
		&kubernetes.Service{
			ID:          release + "-vmclarity-yara-rule-server",
			Namespace:   namespace,
			Application: "vmclarity",
			Component:   "yara-rule-server",
			State:       envtypes.ServiceStateReady,
		},
		&kubernetes.Service{
			ID:          release + "-postgresql",
			Namespace:   namespace,
			Application: "postgresql",
			Component:   "primary",
			State:       envtypes.ServiceStateReady,
		},
		&kubernetes.Service{
			ID:          release + "-vmclarity-cr-discovery-server",
			Namespace:   namespace,
			Application: "vmclarity",
			Component:   "cr-discovery-server",
			State:       envtypes.ServiceStateReady,
		},
	}
}

func NewDockerServices(project string) envtypes.Services {
	return envtypes.Services{
		&docker.Service{
			ID:          "alpine",
			Namespace:   project,
			Application: "vmclarity",
			Component:   "alpine",
			State:       envtypes.ServiceStateReady,
		},
		&docker.Service{
			ID:          "apiserver",
			Namespace:   project,
			Application: "vmclarity",
			Component:   "apiserver",
			State:       envtypes.ServiceStateReady,
		},
		&docker.Service{
			ID:          "exploit-db-server",
			Namespace:   project,
			Application: "vmclarity",
			Component:   "exploit-db-server",
			State:       envtypes.ServiceStateReady,
		},
		&docker.Service{
			ID:          "freshclam-mirror",
			Namespace:   project,
			Application: "vmclarity",
			Component:   "freshclam-mirror",
			State:       envtypes.ServiceStateReady,
		},
		&docker.Service{
			ID:          "gateway",
			Namespace:   project,
			Application: "vmclarity",
			Component:   "gateway",
			State:       envtypes.ServiceStateReady,
		},
		&docker.Service{
			ID:          "grype-server",
			Namespace:   project,
			Application: "vmclarity",
			Component:   "grype-server",
			State:       envtypes.ServiceStateReady,
		},
		&docker.Service{
			ID:          "orchestrator",
			Namespace:   project,
			Application: "vmclarity",
			Component:   "orchestrator",
			State:       envtypes.ServiceStateReady,
		},
		&docker.Service{
			ID:          "swagger-ui",
			Namespace:   project,
			Application: "vmclarity",
			Component:   "swagger-ui",
			State:       envtypes.ServiceStateReady,
		},
		&docker.Service{
			ID:          "trivy-server",
			Namespace:   project,
			Application: "vmclarity",
			Component:   "trivy-server",
			State:       envtypes.ServiceStateReady,
		},
		&docker.Service{
			ID:          "ui",
			Namespace:   project,
			Application: "vmclarity",
			Component:   "ui",
			State:       envtypes.ServiceStateReady,
		},
		&docker.Service{
			ID:          "uibackend",
			Namespace:   project,
			Application: "vmclarity",
			Component:   "uibackend",
			State:       envtypes.ServiceStateReady,
		},
		&docker.Service{
			ID:          "yara-rule-server",
			Namespace:   project,
			Application: "vmclarity",
			Component:   "yara-rule-server",
			State:       envtypes.ServiceStateReady,
		},
	}
}
