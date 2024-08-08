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

package docker

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
)

var cachedClient atomic.Pointer[dockerclient.Client]

type dockerClient struct {
	*dockerclient.Client
}

// newDockerClient creates a wrapped docker client with helper methods.
func newDockerClient() (*dockerClient, error) {
	if cachedClient.Load() == nil {
		client, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
		if err != nil {
			return nil, fmt.Errorf("failed to create docker dockerClient: %w", err)
		}
		cachedClient.Store(client)
	}

	return &dockerClient{
		Client: cachedClient.Load(),
	}, nil
}

// GetHostContainer returns the container in which this code is being executed
// when running in container mode.
func (c *dockerClient) GetHostContainer(ctx context.Context) (*dockertypes.ContainerJSON, error) {
	// Check if the host is running inside a docker container.
	// The hostname is the container ID unless it was manually changed.
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get machine hostname: %w", err)
	}

	// Get host container details
	switch container, err := c.Client.ContainerInspect(ctx, hostname); {
	case err == nil:
		// host in container
		return &container, nil

	case dockerclient.IsErrNotFound(err):
		// host not in container
		return nil, nil // nolint:nilnil

	default:
		// docker error
		return nil, fmt.Errorf("failed to inspect host: %w", err)
	}
}

// GetOrCreateBridgeNetwork creates a bridge network if it does not exist or
// returns ID if it exists.
func (c *dockerClient) GetOrCreateBridgeNetwork(ctx context.Context, networkName string) (string, error) {
	// Do nothing if network already exists
	networkID, _ := c.getNetworkIDFromName(ctx, networkName)
	if networkID != "" {
		return networkID, nil
	}

	// Create network
	networkResp, err := c.Client.NetworkCreate(
		ctx,
		networkName,
		dockertypes.NetworkCreate{
			CheckDuplicate: true,
			Driver:         "bridge",
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to create plugin network: %w", err)
	}

	return networkResp.ID, nil
}

func (c *dockerClient) getNetworkIDFromName(ctx context.Context, networkName string) (string, error) {
	networks, err := c.Client.NetworkList(ctx, dockertypes.NetworkListOptions{
		Filters: filters.NewArgs(filters.Arg("name", networkName)),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list networks: %w", err)
	}
	if len(networks) == 0 {
		return "", fmt.Errorf("scan network not found: %w", err)
	}
	if len(networks) > 1 {
		for _, n := range networks {
			if n.Name == networkName {
				return n.ID, nil
			}
		}
		return "", fmt.Errorf("found more than one scan network: %w", err)
	}
	return networks[0].ID, nil
}
