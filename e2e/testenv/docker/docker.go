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
	"net/url"
	"strings"
	"time"

	"github.com/docker/compose/v2/cmd/formatter"

	"github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/pkg/errors"

	envtypes "github.com/openclarity/vmclarity/e2e/testenv/types"
)

const (
	dockerVMclarityAPIServerServiceName = "apiserver"
	dockerStateRunning                  = "running"
	dockerHealthStateHealthy            = "healthy"
)

type ContextKeyType string

const DockerComposeContextKey ContextKeyType = "DockerCompose"

var dockerComposeFiles = []string{
	"../installation/docker/docker-compose.yml",
	"testenv/docker/docker-compose.override.yml",
}

type DockerEnv struct {
	composer api.Service
	project  *types.Project
}

// nolint:wrapcheck
func New(_ *envtypes.Config) (*DockerEnv, error) {
	projOpts, err := cli.NewProjectOptions(
		dockerComposeFiles,
		cli.WithName("vmclarity-e2e"),
		cli.WithResolvedPaths(true),
	)
	if err != nil {
		return nil, err
	}

	project, err := cli.ProjectFromOptions(projOpts)
	if err != nil {
		return nil, err
	}

	err = cli.WithOsEnv(projOpts)
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

	cliOpts := cliflags.NewClientOptions()

	if err = cmd.Initialize(cliOpts); err != nil {
		return nil, err
	}

	return &DockerEnv{
		composer: compose.NewComposeService(cmd),
		project:  project,
	}, nil
}

// nolint:wrapcheck
func (e *DockerEnv) Start(ctx context.Context) error {
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
func (e *DockerEnv) Stop(ctx context.Context) error {
	timeout := 1 * time.Minute
	opts := api.DownOptions{
		RemoveOrphans: true,
		Project:       e.project,
		Volumes:       true,
		Timeout:       &timeout,
	}
	return e.composer.Down(ctx, e.project.Name, opts)
}

func (e *DockerEnv) SetUp(_ context.Context) error {
	// NOTE(chrisgacsal): nothing to do
	return nil
}

func (e *DockerEnv) TearDown(_ context.Context) error {
	// NOTE(chrisgacsal): nothing to do
	return nil
}

// nolint:wrapcheck
func (e *DockerEnv) ServicesReady(ctx context.Context) (bool, error) {
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
		if c.State != dockerStateRunning && c.Health != dockerHealthStateHealthy {
			return false, nil
		}
	}

	return true, nil
}

// nolint:wrapcheck
func (e *DockerEnv) ServiceLogs(ctx context.Context, services []string, startTime time.Time, stdout, stderr io.Writer) error {
	consumer := formatter.NewLogConsumer(ctx, stdout, stderr, true, true, false)
	return e.composer.Logs(ctx, e.project.Name, consumer, api.LogOptions{
		Project:  e.project,
		Services: services,
		Since:    startTime.Format(time.RFC3339Nano),
	})
}

func (e *DockerEnv) Services() []string {
	services := make([]string, len(e.project.Services))
	for i, srv := range e.project.Services {
		services[i] = srv.Name
	}
	return services
}

func (e *DockerEnv) VMClarityAPIURL() (*url.URL, error) {
	var vmClarityBackend types.ServiceConfig
	var ok bool

	for _, srv := range e.project.Services {
		if srv.Name == dockerVMclarityAPIServerServiceName {
			vmClarityBackend = srv
			ok = true
			break
		}
	}

	if !ok {
		return nil, errors.Errorf("container with name %s is not available", dockerVMclarityAPIServerServiceName)
	}

	if len(vmClarityBackend.Ports) < 1 {
		return nil, errors.Errorf("container with name %s has no ports published", dockerVMclarityAPIServerServiceName)
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

func (e *DockerEnv) Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, DockerComposeContextKey, e.composer)
}
