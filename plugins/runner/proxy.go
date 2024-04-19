// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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

package runner

import (
	"context"
	"fmt"

	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
)

const (
	proxyImage         = "traefik:v2.11"
	proxyContainerName = "plugin-scanner-proxy"
	proxyHostAddress   = "127.0.0.1:8080"
)

func (m *PluginManager) StartProxyContainer(ctx context.Context) error {
	// Pull proxy image
	err := PullImage(ctx, m.dockerClient, proxyImage)
	if err != nil {
		return fmt.Errorf("failed to pull proxy image: %w", err)
	}

	// Get proxy port (HOST:CONTAINER)
	ports, bindings, _ := nat.ParsePortSpecs([]string{proxyHostAddress + ":80"})

	// Create container
	container, err := m.dockerClient.ContainerCreate(
		ctx,
		&containertypes.Config{
			Image: proxyImage,
			Cmd: []string{
				"--api.insecure=true",
				"--providers.docker=true",
				"--entrypoints.web.address=:80",
			},
			ExposedPorts: ports,
		},
		&containertypes.HostConfig{
			PortBindings: bindings,
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: "/var/run/docker.sock",
					Target: "/var/run/docker.sock",
				},
			},
		},
		nil,
		nil,
		proxyContainerName,
	)
	if err != nil {
		return fmt.Errorf("failed to create proxy container: %w", err)
	}

	m.proxyID = container.ID

	// Start proxy container
	err = m.dockerClient.ContainerStart(ctx, container.ID, containertypes.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start proxy container: %w", err)
	}

	return nil
}

func (m *PluginManager) StopProxyContainer(ctx context.Context) error {
	// Stop proxy container
	err := m.dockerClient.ContainerStop(ctx, m.proxyID, containertypes.StopOptions{})
	if err != nil {
		return fmt.Errorf("failed to stop proxy container: %w", err)
	}

	// Remove proxy container
	err = m.dockerClient.ContainerRemove(ctx, m.proxyID, containertypes.RemoveOptions{})
	if err != nil {
		return fmt.Errorf("failed to remove proxy container: %w", err)
	}

	return nil
}
