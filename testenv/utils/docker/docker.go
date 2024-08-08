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
	"errors"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"

	"github.com/docker/compose/v2/pkg/api"

	envtypes "github.com/openclarity/vmclarity/testenv/types"
	"github.com/openclarity/vmclarity/testenv/utils"
)

const (
	DockerTimeout  = 5 * time.Minute
	TickerInterval = 5 * time.Second
)

type DockerHelper struct {
	client *client.Client
}

func (e *DockerHelper) ServicesReady(ctx context.Context) (bool, error) {
	logger := utils.GetLoggerFromContextOrDiscard(ctx)

	services, err := e.Services(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve list of services: %w", err)
	}

	result := true
	for _, service := range services {
		logger.Infof("checking service readiness. Service=%s State=%s", service.GetID(), service.GetState())
		switch service.GetState() {
		case envtypes.ServiceStateReady:
			// continue
		case envtypes.ServiceStateDegraded, envtypes.ServiceStateNotReady, envtypes.ServiceStateUnknown:
			fallthrough
		default:
			result = false
		}
	}

	return result, nil
}

func (e *DockerHelper) Services(ctx context.Context) (envtypes.Services, error) {
	containerFilters := filters.NewArgs([]filters.KeyValuePair{
		{
			Key:   "label",
			Value: api.ProjectLabel + "=vmclarity",
		},
	}...)
	containers, err := e.client.ContainerList(ctx, container.ListOptions{
		Filters: containerFilters,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	serviceCollection := make(ServiceCollection)
	for _, c := range containers {
		serviceCollection[c.ID] = &Service{
			ID:    c.ID,
			State: getContainerState(c.State),
		}
	}

	return serviceCollection.AsServices(), nil
}

func (e *DockerHelper) WaitForDockerReady(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, DockerTimeout)
	defer cancel()

	ticker := time.NewTicker(TickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.New("stopping periodic check due to timeout")
		case <-ticker.C:
			_, err := e.client.Ping(ctx)
			if err != nil {
				continue
			}
			return nil
		}
	}
}

func New(opts []client.Opt) (*DockerHelper, error) {
	apiClient, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &DockerHelper{
		client: apiClient,
	}, nil
}

func getContainerState(state string) envtypes.ServiceState {
	switch state {
	case ContainerStateRunning:
		return envtypes.ServiceStateReady
	case ContainerStateExited, ContainerStateDead:
		return envtypes.ServiceStateNotReady
	default:
		return envtypes.ServiceStateUnknown
	}
}
