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

package helm

import (
	"testing"

	. "github.com/onsi/gomega"

	envtypes "github.com/openclarity/openclarity/testenv/types"
)

func TestChartWithContainerImages(t *testing.T) {
	tests := []struct {
		Name    string
		EnvVars map[string]string

		HemConfig          *Config
		ContainerImagesMap map[string]string
		ExpectedValues     map[string]map[string]string
	}{
		{
			Name:    "Values with container images",
			EnvVars: map[string]string{},
			HemConfig: &Config{
				Namespace:      "default",
				ReleaseName:    "testenv-k8s-test",
				ChartPath:      "",
				StorageDriver:  "secret",
				KubeConfigPath: "kubeconfig",
			},
			ContainerImagesMap: map[string]string{
				"apiserver_image":           "openclarity.io/openclarity-api-server",
				"orchestrator_image":        "openclarity.io/openclarity-orchestrator:test@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				"ui_image":                  "ghcr.io/openclarity/openclarity-ui:test",
				"uibackend_image":           "openclarity.io/openclarity-ui-backend:test",
				"scanner_image":             "openclarity-scanner:test",
				"cr_discovery_server_image": "openclarity.io/openclarity-cr-discovery-server:test",
			},
			ExpectedValues: map[string]map[string]string{
				"apiserver.image": {
					"registry":   "openclarity.io",
					"repository": "openclarity-api-server",
					"tag":        "latest",
				},
				"orchestrator.image": {
					"registry":   "openclarity.io",
					"repository": "openclarity-orchestrator",
					"tag":        "test",
					"digest":     "sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				},
				"orchestrator.scannerImage": {
					"registry":   "docker.io",
					"repository": "library/openclarity-scanner",
					"tag":        "test",
				},
				"orchestrator.serviceAccount": {
					"automountServiceAccountToken": "true",
				},
				"orchestrator": {
					"provider": "kubernetes",
				},
				"ui.image": {
					"registry":   "ghcr.io",
					"repository": "openclarity/openclarity-ui",
					"tag":        "test",
				},
				"uibackend.image": {
					"registry":   "openclarity.io",
					"repository": "openclarity-ui-backend",
					"tag":        "test",
				},
				"crDiscoveryServer.image": {
					"registry":   "openclarity.io",
					"repository": "openclarity-cr-discovery-server",
					"tag":        "test",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			containerImages, err := envtypes.NewContainerImages[envtypes.ImageRef](test.ContainerImagesMap)
			g.Expect(err).Should(Not(HaveOccurred()))

			values := NewEmptyValues()
			opts := []ValuesOpts{
				WithContainerImages(containerImages),
				WithKubernetesProvider(),
				WithGatewayNodePort(30000),
			}

			err = applyValuesWithOpts(&values, opts...)
			g.Expect(err).Should(Not(HaveOccurred()))

			yaml, err := values.YAML()
			g.Expect(err).Should(Not(HaveOccurred()))
			t.Logf("values YAML: %v", yaml)

			for subSection, expectedValues := range test.ExpectedValues {
				t.Logf("with subsection: %s", subSection)

				subValues, err := values.Table(subSection)
				g.Expect(err).Should(Not(HaveOccurred()))

				for key, expectedValue := range expectedValues {
					actualValue, ok := subValues[key]
					g.Expect(ok).Should(BeTrue())

					g.Expect(actualValue).Should(BeEquivalentTo(expectedValue))
				}
			}
		})
	}
}
