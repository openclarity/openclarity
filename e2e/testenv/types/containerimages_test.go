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
				"apiserver_image":    "openclarity.io/vmclarity-apiserver",
				"orchestrator_image": "openclarity.io/vmclarity-orchestrator:test@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				"ui_image":           "ghcr.io/openclarity/vmclarity-ui:test",
				"uibackend_image":    "openclarity.io/vmclarity-uibackend:test",
				"scanner_image":      "vmclarity-scanner@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
			},
			ExpectedContainerImages: &ContainerImages[ImageRef]{
				APIServer: ImageRef{
					name:   "openclarity.io/vmclarity-apiserver",
					domain: "openclarity.io",
					path:   "vmclarity-apiserver",
					tag:    "latest",
					digest: "",
				},
				Orchestrator: ImageRef{
					name:   "openclarity.io/vmclarity-orchestrator",
					domain: "openclarity.io",
					path:   "vmclarity-orchestrator",
					tag:    "test",
					digest: "sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				},
				UI: ImageRef{
					name:   "ghcr.io/openclarity/vmclarity-ui",
					domain: "ghcr.io",
					path:   "openclarity/vmclarity-ui",
					tag:    "test",
					digest: "",
				},
				UIBackend: ImageRef{
					name:   "openclarity.io/vmclarity-uibackend",
					domain: "openclarity.io",
					path:   "vmclarity-uibackend",
					tag:    "test",
					digest: "",
				},
				Scanner: ImageRef{
					name:   "docker.io/library/vmclarity-scanner",
					domain: "docker.io",
					path:   "library/vmclarity-scanner",
					tag:    "",
					digest: "sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				},
			},
			ExpectedSlice: []ImageRef{
				{
					name:   "openclarity.io/vmclarity-apiserver",
					domain: "openclarity.io",
					path:   "vmclarity-apiserver",
					tag:    "latest",
					digest: "",
				},
				{
					name:   "openclarity.io/vmclarity-orchestrator",
					domain: "openclarity.io",
					path:   "vmclarity-orchestrator",
					tag:    "test",
					digest: "sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				},
				{
					name:   "ghcr.io/openclarity/vmclarity-ui",
					domain: "ghcr.io",
					path:   "openclarity/vmclarity-ui",
					tag:    "test",
					digest: "",
				},
				{
					name:   "openclarity.io/vmclarity-uibackend",
					domain: "openclarity.io",
					path:   "vmclarity-uibackend",
					tag:    "test",
					digest: "",
				},
				{
					name:   "docker.io/library/vmclarity-scanner",
					domain: "docker.io",
					path:   "library/vmclarity-scanner",
					tag:    "",
					digest: "sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				},
			},
			ExpectedStringSlice: []string{
				"openclarity.io/vmclarity-apiserver:latest",
				"openclarity.io/vmclarity-orchestrator:test@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				"ghcr.io/openclarity/vmclarity-ui:test",
				"openclarity.io/vmclarity-uibackend:test",
				"docker.io/library/vmclarity-scanner@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
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
				"apiserver_image":    "openclarity.io/vmclarity-apiserver",
				"orchestrator_image": "openclarity.io/vmclarity-orchestrator:test@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				"ui_image":           "ghcr.io/openclarity/vmclarity-ui:test",
				"uibackend_image":    "openclarity.io/vmclarity-uibackend:test",
				"scanner_image":      "vmclarity-scanner@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
			},
			ExpectedContainerImages: &ContainerImages[string]{
				APIServer:    "openclarity.io/vmclarity-apiserver",
				Orchestrator: "openclarity.io/vmclarity-orchestrator:test@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				UI:           "ghcr.io/openclarity/vmclarity-ui:test",
				UIBackend:    "openclarity.io/vmclarity-uibackend:test",
				Scanner:      "vmclarity-scanner@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
			},
			ExpectedSlice: []string{
				"openclarity.io/vmclarity-apiserver",
				"openclarity.io/vmclarity-orchestrator:test@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				"ghcr.io/openclarity/vmclarity-ui:test",
				"openclarity.io/vmclarity-uibackend:test",
				"vmclarity-scanner@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
			},
			ExpectedStringSlice: []string{
				"openclarity.io/vmclarity-apiserver",
				"openclarity.io/vmclarity-orchestrator:test@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
				"ghcr.io/openclarity/vmclarity-ui:test",
				"openclarity.io/vmclarity-uibackend:test",
				"vmclarity-scanner@sha256:96374b22a50bcfc96b96b5153b185ce5bf16d7a454766747633a32d2f1fefead",
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
