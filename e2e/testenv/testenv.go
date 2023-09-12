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

package testenv

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/pkg/errors"
)

const (
	vmclarityAPIServerServiceName = "apiserver"
	stateRunning                  = "running"
	healthStateHealthy            = "healthy"
)

type Environment struct {
	composer api.Service
	project  *types.Project
	reuse    bool
}

// nolint:wrapcheck
func New(o *cli.ProjectOptions, reuse bool) (*Environment, error) {
	project, err := cli.ProjectFromOptions(o)
	if err != nil {
		return nil, err
	}

	for i, service := range project.Services {
		service.CustomLabels = map[string]string{
			api.ProjectLabel:     project.Name,
			api.ServiceLabel:     service.Name,
			api.WorkingDirLabel:  project.WorkingDir,
			api.ConfigFilesLabel: strings.Join(project.ComposeFiles, ","),
			api.OneoffLabel:      "False",
		}
		project.Services[i] = service
	}

	cmd, err := command.NewDockerCli()
	if err != nil {
		return nil, err
	}

	opts := cliflags.NewClientOptions()

	if err = cmd.Initialize(opts); err != nil {
		return nil, err
	}

	return &Environment{
		composer: compose.NewComposeService(cmd),
		project:  project,
		reuse:    reuse,
	}, nil
}

// nolint:wrapcheck
func (e *Environment) Start(ctx context.Context) error {
	if e.reuse {
		return nil
	}

	timeout := 1 * time.Minute
	opts := api.UpOptions{
		Create: api.CreateOptions{
			RemoveOrphans: true,
			QuietPull:     true,
			Timeout:       &timeout,
			Services:      e.Services(),
			Inherit:       false,
		},
		Start: api.StartOptions{
			Project:     e.project,
			Wait:        true,
			WaitTimeout: 10 * time.Minute, // nolint:gomnd
			Services:    e.Services(),
		},
	}
	return e.composer.Up(ctx, e.project, opts)
}

// nolint:wrapcheck
func (e *Environment) Stop(ctx context.Context) error {
	if e.reuse {
		return nil
	}

	timeout := 1 * time.Minute
	opts := api.DownOptions{
		RemoveOrphans: true,
		Project:       e.project,
		Volumes:       true,
		Timeout:       &timeout,
	}
	return e.composer.Down(ctx, e.project.Name, opts)
}

// nolint:wrapcheck
func (e *Environment) ServicesReady(ctx context.Context) (bool, error) {
	services := e.Services()

	ps, err := e.composer.Ps(
		ctx,
		e.project.Name,
		api.PsOptions{
			Services: services,
			Project:  e.project,
		},
	)
	if err != nil {
		return false, err
	}

	if len(services) != len(ps) {
		return false, nil
	}

	for _, c := range ps {
		if c.State != stateRunning && c.Health != healthStateHealthy {
			return false, nil
		}
	}

	return true, nil
}

func (e *Environment) Services() []string {
	services := make([]string, len(e.project.Services))
	for i, srv := range e.project.Services {
		services[i] = srv.Name
	}
	return services
}

func (e *Environment) VMClarityURL() (*url.URL, error) {
	var vmClarityBackend types.ServiceConfig
	var ok bool

	for _, srv := range e.project.Services {
		if srv.Name == vmclarityAPIServerServiceName {
			vmClarityBackend = srv
			ok = true
			break
		}
	}

	if !ok {
		return nil, errors.Errorf("container with name %s is not available", vmclarityAPIServerServiceName)
	}

	if len(vmClarityBackend.Ports) < 1 {
		return nil, errors.Errorf("container with name %s has no ports published", vmclarityAPIServerServiceName)
	}

	port := vmClarityBackend.Ports[0].Published
	hostIP := vmClarityBackend.Ports[0].HostIP
	if hostIP == "" {
		hostIP = "127.0.0.1"
	}

	return &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%s", hostIP, port),
	}, nil
}
