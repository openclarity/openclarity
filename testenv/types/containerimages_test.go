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

package types

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestContainerImagesWithImageRef(t *testing.T) {
	tests := []struct {
		Name                    string
		ContainerImagesMap      map[string]string
		ExpectedContainerImages *ContainerImages[ImageRef]
		ExpectedSlice           []ImageRef
		ExpectedStringSlice     []string
	}{
		{
			Name: "ImageRef",
			ContainerImagesMap: map[string]string{
				"apiserver_image":           "openclarity.io/openclarity-api-server",
				"orchestrator_image":        "openclarity.io/openclarity-orchestrator:test@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				"ui_image":                  "ghcr.io/openclarity/openclarity-ui:test",
				"uibackend_image":           "openclarity.io/openclarity-ui-backend:test",
				"scanner_image":             "openclarity-scanner@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				"cr_discovery_server_image": "ghcr.io/openclarity/openclarity-cr-discovery-server:test",
				"plugin_kics_image":         "ghcr.io/openclarity/openclarity-plugin-kics:test",
			},
			ExpectedContainerImages: &ContainerImages[ImageRef]{
				APIServer: ImageRef{
					name:   "openclarity.io/openclarity-api-server",
					domain: "openclarity.io",
					path:   "openclarity-api-server",
					tag:    "latest",
					digest: "",
				},
				Orchestrator: ImageRef{
					name:   "openclarity.io/openclarity-orchestrator",
					domain: "openclarity.io",
					path:   "openclarity-orchestrator",
					tag:    "test",
					digest: "sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				},
				UI: ImageRef{
					name:   "ghcr.io/openclarity/openclarity-ui",
					domain: "ghcr.io",
					path:   "openclarity/openclarity-ui",
					tag:    "test",
					digest: "",
				},
				UIBackend: ImageRef{
					name:   "openclarity.io/openclarity-ui-backend",
					domain: "openclarity.io",
					path:   "openclarity-ui-backend",
					tag:    "test",
					digest: "",
				},
				Scanner: ImageRef{
					name:   "docker.io/library/openclarity-scanner",
					domain: "docker.io",
					path:   "library/openclarity-scanner",
					tag:    "",
					digest: "sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				},
				CRDiscoveryServer: ImageRef{
					name:   "ghcr.io/openclarity/openclarity-cr-discovery-server",
					domain: "ghcr.io",
					path:   "openclarity/openclarity-cr-discovery-server",
					tag:    "test",
					digest: "",
				},
				PluginKics: ImageRef{
					name:   "ghcr.io/openclarity/openclarity-plugin-kics",
					domain: "ghcr.io",
					path:   "openclarity/openclarity-plugin-kics",
					tag:    "test",
					digest: "",
				},
			},
			ExpectedSlice: []ImageRef{
				{
					name:   "openclarity.io/openclarity-api-server",
					domain: "openclarity.io",
					path:   "openclarity-api-server",
					tag:    "latest",
					digest: "",
				},
				{
					name:   "openclarity.io/openclarity-orchestrator",
					domain: "openclarity.io",
					path:   "openclarity-orchestrator",
					tag:    "test",
					digest: "sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				},
				{
					name:   "ghcr.io/openclarity/openclarity-ui",
					domain: "ghcr.io",
					path:   "openclarity/openclarity-ui",
					tag:    "test",
					digest: "",
				},
				{
					name:   "openclarity.io/openclarity-ui-backend",
					domain: "openclarity.io",
					path:   "openclarity-ui-backend",
					tag:    "test",
					digest: "",
				},
				{
					name:   "docker.io/library/openclarity-scanner",
					domain: "docker.io",
					path:   "library/openclarity-scanner",
					tag:    "",
					digest: "sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				},
				{
					name:   "ghcr.io/openclarity/openclarity-cr-discovery-server",
					domain: "ghcr.io",
					path:   "openclarity/openclarity-cr-discovery-server",
					tag:    "test",
					digest: "",
				},
				{
					name:   "ghcr.io/openclarity/openclarity-plugin-kics",
					domain: "ghcr.io",
					path:   "openclarity/openclarity-plugin-kics",
					tag:    "test",
					digest: "",
				},
			},
			ExpectedStringSlice: []string{
				"openclarity.io/openclarity-api-server:latest",
				"openclarity.io/openclarity-orchestrator:test@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				"ghcr.io/openclarity/openclarity-ui:test",
				"openclarity.io/openclarity-ui-backend:test",
				"docker.io/library/openclarity-scanner@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				"ghcr.io/openclarity/openclarity-cr-discovery-server:test",
				"ghcr.io/openclarity/openclarity-plugin-kics:test",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			containerImages, err := NewContainerImages[ImageRef](test.ContainerImagesMap)
			g.Expect(err).Should(Not(HaveOccurred()))
			g.Expect(containerImages).Should(BeEquivalentTo(test.ExpectedContainerImages))

			asSlice := containerImages.AsSlice()
			g.Expect(asSlice).Should(BeEquivalentTo(test.ExpectedSlice))

			asStringSlice, err := containerImages.AsStringSlice()
			g.Expect(err).Should(Not(HaveOccurred()))
			g.Expect(asStringSlice).Should(BeEquivalentTo(test.ExpectedStringSlice))
		})
	}
}

func TestContainerImagesWithString(t *testing.T) {
	tests := []struct {
		Name                    string
		ContainerImagesMap      map[string]string
		ExpectedContainerImages *ContainerImages[string]
		ExpectedSlice           []string
		ExpectedStringSlice     []string
	}{
		{
			Name: "Values with container images",
			ContainerImagesMap: map[string]string{
				"apiserver_image":           "openclarity.io/openclarity-api-server",
				"orchestrator_image":        "openclarity.io/openclarity-orchestrator:test@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				"ui_image":                  "ghcr.io/openclarity/openclarity-ui:test",
				"uibackend_image":           "openclarity.io/openclarity-ui-backend:test",
				"scanner_image":             "openclarity-scanner@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				"cr_discovery_server_image": "ghcr.io/openclarity/openclarity-cr-discovery-server:test",
				"plugin_kics_image":         "openclarity-plugin-kics:test",
			},
			ExpectedContainerImages: &ContainerImages[string]{
				APIServer:         "openclarity.io/openclarity-api-server",
				Orchestrator:      "openclarity.io/openclarity-orchestrator:test@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				UI:                "ghcr.io/openclarity/openclarity-ui:test",
				UIBackend:         "openclarity.io/openclarity-ui-backend:test",
				Scanner:           "openclarity-scanner@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				CRDiscoveryServer: "ghcr.io/openclarity/openclarity-cr-discovery-server:test",
				PluginKics:        "openclarity-plugin-kics:test",
			},
			ExpectedSlice: []string{
				"openclarity.io/openclarity-api-server",
				"openclarity.io/openclarity-orchestrator:test@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				"ghcr.io/openclarity/openclarity-ui:test",
				"openclarity.io/openclarity-ui-backend:test",
				"openclarity-scanner@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				"ghcr.io/openclarity/openclarity-cr-discovery-server:test",
				"openclarity-plugin-kics:test",
			},
			ExpectedStringSlice: []string{
				"openclarity.io/openclarity-api-server",
				"openclarity.io/openclarity-orchestrator:test@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				"ghcr.io/openclarity/openclarity-ui:test",
				"openclarity.io/openclarity-ui-backend:test",
				"openclarity-scanner@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				"ghcr.io/openclarity/openclarity-cr-discovery-server:test",
				"openclarity-plugin-kics:test",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			containerImages, err := NewContainerImages[string](test.ContainerImagesMap)
			g.Expect(err).Should(Not(HaveOccurred()))
			g.Expect(containerImages).Should(BeEquivalentTo(test.ExpectedContainerImages))

			asSlice := containerImages.AsSlice()
			g.Expect(asSlice).Should(BeEquivalentTo(test.ExpectedSlice))

			asStringSlice, err := containerImages.AsStringSlice()
			g.Expect(err).Should(Not(HaveOccurred()))
			g.Expect(asStringSlice).Should(BeEquivalentTo(test.ExpectedStringSlice))
		})
	}
}
