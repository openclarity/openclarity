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
	"fmt"
	"io"
	"time"

	"github.com/docker/cli/cli/command"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/cmd/formatter"
	"github.com/docker/compose/v2/pkg/compose"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/pkg/errors"

	envtypes "github.com/openclarity/vmclarity/testenv/types"
	"github.com/openclarity/vmclarity/testenv/utils"
)

type ContextKeyType string

const (
	GatewayServiceName                     = "gateway"
	DockerComposeContextKey ContextKeyType = "DockerCompose"
)

type DockerEnv struct {
	composer api.Service
	project  *types.Project

	meta map[string]interface{}
}

func (e *DockerEnv) SetUp(ctx context.Context) error {
	deadline, _ := ctx.Deadline()
	timeout := time.Until(deadline)

	services, err := e.Services(ctx)
	if err != nil {
		return err
	}

	opts := api.UpOptions{
		Create: api.CreateOptions{
			RemoveOrphans: true,
			QuietPull:     true,
			Timeout:       &timeout,
			Services:      services.IDs(),
			Inherit:       false,
		},
		Start: api.StartOptions{
			Project:     e.project,
			Wait:        true,
			WaitTimeout: timeout,
			Services:    services.IDs(),
		},
	}

	if err = e.composer.Up(ctx, e.project, opts); err != nil {
		return fmt.Errorf("failed to set up environment: %w", err)
	}

	return nil
}

func (e *DockerEnv) TearDown(ctx context.Context) error {
	timeout := 1 * time.Minute
	opts := api.DownOptions{
		RemoveOrphans: true,
		Project:       e.project,
		Volumes:       true,
		Timeout:       &timeout,
	}

	if err := e.composer.Down(ctx, e.project.Name, opts); err != nil {
		return fmt.Errorf("failed to tear down environment: %w", err)
	}

	return nil
}

func (e *DockerEnv) ServicesReady(ctx context.Context) (bool, error) {
	logger := utils.GetLoggerFromContextOrDiscard(ctx).WithFields(e.meta)

	services, err := e.Services(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve list of services: %w", err)
	}

	var result bool
	for _, service := range services {
		logger.Debugf("checking service readiness. Service=%s State=%s", service.GetID(), service.GetState())
		switch service.GetState() {
		case envtypes.ServiceStateReady:
			result = true
		case envtypes.ServiceStateDegraded, envtypes.ServiceStateNotReady, envtypes.ServiceStateUnknown:
			fallthrough
		default:
			result = false
		}
	}

	return result, nil
}

func (e *DockerEnv) ServiceLogs(ctx context.Context, services []string, startTime time.Time, stdout, stderr io.Writer) error {
	consumer := formatter.NewLogConsumer(ctx, stdout, stderr, true, true, false)

	err := e.composer.Logs(ctx, e.project.Name, consumer, api.LogOptions{
		Project:  e.project,
		Services: services,
		Since:    startTime.Format(time.RFC3339Nano),
	})
	if err != nil {
		return fmt.Errorf("failed to retrieve service logs: %w", err)
	}

	return nil
}

func (e *DockerEnv) Services(ctx context.Context) (envtypes.Services, error) {
	serviceCollection := NewServiceCollectionFromProject(e.project)

	ps, err := e.composer.Ps(ctx, e.project.Name, api.PsOptions{
		Services: serviceCollection.ServiceNames(),
		Project:  e.project,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get list of available services: %w", err)
	}

	serviceMap := make(map[string]types.ServiceConfig)
	for _, srv := range e.project.Services {
		serviceMap[srv.Name] = srv
	}

	for _, summary := range ps {
		var healthCheckEnabled bool
		if srv, ok := serviceMap[summary.Service]; ok {
			if srv.HealthCheck != nil && !srv.HealthCheck.Disable {
				healthCheckEnabled = true
			}
		}
		serviceCollection[summary.Service].State = getServiceState(summary, healthCheckEnabled)
	}

	return serviceCollection.AsServices(), nil
}

func (e *DockerEnv) Endpoints(_ context.Context) (*envtypes.Endpoints, error) {
	var gatewayService types.ServiceConfig
	var found bool

	for _, srv := range e.project.Services {
		if srv.Name == GatewayServiceName {
			gatewayService = srv
			found = true
			break
		}
	}

	if !found {
		return nil, errors.Errorf("service with name %s is not available", GatewayServiceName)
	}

	if len(gatewayService.Ports) < 1 {
		return nil, errors.Errorf("service with name %s has no published ports", GatewayServiceName)
	}

	port := gatewayService.Ports[0].Published
	host := gatewayService.Ports[0].HostIP
	if host == "" {
		host = "127.0.0.1"
	}

	endpoints := new(envtypes.Endpoints)
	endpoints.SetAPI("http", host, port, "/api")
	endpoints.SetUIBackend("http", host, port, "/ui/api")

	return endpoints, nil
}

func (e *DockerEnv) Context(ctx context.Context) (context.Context, error) {
	return context.WithValue(ctx, DockerComposeContextKey, e.composer), nil
}

func (e *DockerEnv) Deployments() []string {
	return []string{e.project.Name}
}

func New(config *Config, opts ...ConfigOptFn) (*DockerEnv, error) {
	if err := applyConfigWithOpts(config, opts...); err != nil {
		return nil, fmt.Errorf("failed to apply config options: %w", err)
	}

	project, err := ProjectFromConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create project from config: %w", err)
	}

	cmd, err := command.NewDockerCli()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	cliOpts := cliflags.NewClientOptions()

	if err = cmd.Initialize(cliOpts); err != nil {
		return nil, fmt.Errorf("failed to initialize docker client: %w", err)
	}

	return &DockerEnv{
		composer: compose.NewComposeService(cmd),
		project:  project,
		meta: map[string]interface{}{
			"environment": "docker",
			"name":        config.EnvName,
			"project":     project.Name,
		},
	}, nil
}
