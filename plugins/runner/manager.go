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
	"os"
	"path/filepath"

	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"

	runnerclient "github.com/openclarity/vmclarity/plugins/runner/internal/client"
)

type PluginManagerInterface interface {
	Init(ctx context.Context) (CleanupFunc, error)                               // prepares env (starts proxy container). should autoclean on error without returning CleanupFunc. called only once.
	Start(ctx context.Context, config PluginConfig) (Runner, CleanupFunc, error) // creates and starts container. should autoclean on error without returning CleanupFunc. used as a factory for PluginRunner
}

type PluginManager struct {
	dockerClient *client.Client
	proxyID      string
}

func NewPluginManager() (*PluginManager, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &PluginManager{
		dockerClient: dockerClient,
	}, nil
}

func (m *PluginManager) Init(ctx context.Context) (CleanupFunc, error) {
	err := m.StartProxyContainer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start proxy container: %w", err)
	}

	return m.StopProxyContainer, nil
}

func (m *PluginManager) Start(ctx context.Context, config PluginConfig) (Runner, CleanupFunc, error) {
	// Write scanner config file to temp dir
	err := os.WriteFile(getScannerConfigSourcePath(config.Name), []byte(config.ScannerConfig), 0o600) // nolint:gomnd
	if err != nil {
		return Runner{}, nil, fmt.Errorf("failed write scanner config file: %w", err)
	}

	// Pull scanner image if required
	err = PullImage(ctx, m.dockerClient, config.ImageName)
	if err != nil {
		return Runner{}, nil, fmt.Errorf("failed to pull scanner image: %w", err)
	}

	// Create scanner container
	//
	// Traefik redirects the requests to its API to our scanners so that requests to
	// localhost:TraefikContainerPort/{SCANNER_NAME} are redirected to the actual
	// {SCANNER_NAME} container. All traffic flow can be configured (e.g. auth,
	// encryption). This also enables having scanners on different hosts (although
	// not needed + requires additional configuration). The host network driver is
	// not limited by the ports anymore as we only use one (for the proxy), This way,
	// scanner containers wont overload the network driver.
	containerResp, err := m.dockerClient.ContainerCreate(
		ctx,
		&containertypes.Config{
			Image: config.ImageName,
			Env:   []string{"PLUGIN_SERVER_LISTEN_ADDRESS=0.0.0.0:" + DefaultScannerServerPort},
			Labels: map[string]string{
				"traefik.enable": "true",
				"traefik.http.routers." + config.Name + "-scanner.rule":                      "PathPrefix(`/" + config.Name + "/`)",
				"traefik.http.middlewares." + config.Name + "-scanner.stripprefix.prefixes":  "/" + config.Name,
				"traefik.http.routers." + config.Name + "-scanner.middlewares":               config.Name + "-scanner",
				"traefik.http.services." + config.Name + "-scanner.loadbalancer.server.port": DefaultScannerServerPort,
			},
		},
		&containertypes.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: getScannerConfigSourcePath(config.Name),
					Target: getScannerConfigDestinationPath(),
				},
				{
					Type:   mount.TypeBind,
					Source: config.InputDir,
					Target: DefaultScannerInputDir,
				},
				{
					Type:   mount.TypeBind,
					Source: filepath.Dir(config.OutputFile),
					Target: DefaultScannerOutputDir,
				},
			},
		},
		nil,
		nil,
		config.Name,
	)
	if err != nil {
		return Runner{}, nil, fmt.Errorf("failed to create scanner container: %w", err)
	}

	err = m.dockerClient.ContainerStart(ctx, containerResp.ID, containertypes.StartOptions{})
	if err != nil {
		return Runner{}, nil, fmt.Errorf("failed to start scanner container: %w", err)
	}

	client, err := runnerclient.NewClientWithResponses(
		fmt.Sprintf("http://%s/%s/", proxyHostAddress, config.Name),
	)
	if err != nil {
		return Runner{}, nil, fmt.Errorf("failed to create client: %w", err)
	}

	runner := Runner{
		client:       client,
		dockerClient: m.dockerClient,
		containerID:  containerResp.ID,
		PluginConfig: config,
	}

	return runner, runner.Stop, nil
}
