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
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/compose/v2/pkg/api"

	envtypes "github.com/openclarity/vmclarity/e2e/testenv/types"
)

const (
	ApplicationName = "vmclarity"

	ContainerStateRunning = "running"
	ContainerStateExited  = "exited"
	ContainerStateDead    = "dead"
	ContainerStateHealthy = "healthy"
)

// Service is the types.Service interface implementation.
type Service struct {
	ID          string
	Namespace   string
	Application string
	Component   string
	State       envtypes.ServiceState
}

func (s Service) GetID() string {
	return s.ID
}

func (s Service) GetNamespace() string {
	return s.Namespace
}

func (s Service) GetApplicationName() string {
	return s.Application
}

func (s Service) GetComponentName() string {
	return s.Component
}

func (s Service) GetState() envtypes.ServiceState {
	return s.State
}

func (s Service) String() string {
	return s.ID
}

type ServiceCollection map[string]*Service

func (c ServiceCollection) ServiceNames() []string {
	names := make([]string, 0, len(c))
	for name := range c {
		names = append(names, name)
	}

	return names
}

func (c ServiceCollection) AsServices() envtypes.Services {
	services := make(envtypes.Services, 0, len(c))
	for _, service := range c {
		services = append(services, service)
	}

	return services
}

func NewServiceCollectionFromProject(p *types.Project) ServiceCollection {
	collection := make(ServiceCollection, len(p.Services))

	for _, config := range p.Services {
		collection[config.Name] = &Service{
			ID:          config.Name,
			Namespace:   p.Name,
			Application: ApplicationName,
			Component:   config.Name,
			State:       envtypes.ServiceStateUnknown,
		}
	}

	return collection
}

func getServiceState(summary api.ContainerSummary, healthCheck bool) envtypes.ServiceState {
	switch summary.State {
	case ContainerStateRunning:
		if healthCheck && summary.Health != ContainerStateHealthy {
			return envtypes.ServiceStateDegraded
		}
		return envtypes.ServiceStateReady
	case ContainerStateExited, ContainerStateDead:
		return envtypes.ServiceStateNotReady
	default:
		return envtypes.ServiceStateUnknown
	}
}
