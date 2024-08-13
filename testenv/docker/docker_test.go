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

package docker

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestProjectFromConfig(t *testing.T) {
	tests := []struct {
		Name string

		DockerConfig *Config

		ExpectedServiceImages      map[string]string
		ExpectedServiceEnvironment map[string]map[string]string
	}{
		{
			Name: "Custom config",
			DockerConfig: &Config{
				EnvName: "testenv",
				Images: ContainerImages{
					APIServer:    "openclarity/openclarity-api-server:latest",
					Orchestrator: "openclarity/openclarity-orchestrator:latest",
					UI:           "openclarity/openclarity-ui:latest",
					UIBackend:    "openclarity/openclarity-ui-backend:latest",
					Scanner:      "openclarity/openclarity-cli:latest",
					PluginKics:   "openclarity/openclarity-plugin-kics:latest",
				},
				ComposeFiles: []string{
					"../../installation/docker/docker-compose.yml",
					"asset/docker-compose.override.yml",
				},
			},
			ExpectedServiceImages: map[string]string{
				"apiserver":    "openclarity/openclarity-api-server:latest",
				"orchestrator": "openclarity/openclarity-orchestrator:latest",
				"ui":           "openclarity/openclarity-ui:latest",
				"uibackend":    "openclarity/openclarity-ui-backend:latest",
			},
			ExpectedServiceEnvironment: map[string]map[string]string{
				"orchestrator": {
					"VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_CONTAINER_IMAGE": "openclarity/openclarity-cli:latest",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			test.DockerConfig.ctx = ctx

			project, err := ProjectFromConfig(test.DockerConfig)
			g.Expect(err).ShouldNot(HaveOccurred())

			g.Expect(project.Name).Should(BeEquivalentTo(test.DockerConfig.EnvName))

			for service, image := range test.ExpectedServiceImages {
				_, ok := project.Services[service]
				g.Expect(ok).Should(BeTrue())
				g.Expect(project.Services[service].Image).Should(BeEquivalentTo(image))
			}

			for service, envs := range test.ExpectedServiceEnvironment {
				_, ok := project.Services[service]
				g.Expect(ok).Should(BeTrue())

				for name, value := range envs {
					v, ok := project.Services[service].Environment[name]
					g.Expect(ok).Should(BeTrue())
					g.Expect(v).ShouldNot(BeNil())
					g.Expect(*v).Should(BeEquivalentTo(value))
				}
			}
		})
	}
}
