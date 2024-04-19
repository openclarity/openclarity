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
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"github.com/openclarity/vmclarity/core/to"
	runnerclient "github.com/openclarity/vmclarity/plugins/runner/internal/client"
	"github.com/openclarity/vmclarity/plugins/sdk/types"
)

const (
	DefaultScannerInputDir   = "/asset"
	DefaultScannerOutputDir  = "/export"
	DefaultScannerServerPort = "8080"
	DefaultPollInterval      = 2 * time.Second
	DefaultTimeout           = 60 * time.Second
)

var ErrScanNotDone = errors.New("scan has not finished yet")

type PluginRunner interface {
	Start(ctx context.Context) (CleanupFunc, error)
	WaitReady(ctx context.Context) error
	Run(ctx context.Context) error
	WaitDone(ctx context.Context) error
	Result() (io.Reader, error)
	Stop(ctx context.Context) error
}

type PluginConfig struct {
	// Name is the name of the plugin scanner
	Name string `yaml:"name" mapstructure:"name"`
	// ImageName is the name of the docker image that will be used to run the plugin scanner
	ImageName string `yaml:"image_name" mapstructure:"image_name"`
	// InputDir is a directory where the plugin scanner will read the asset filesystem
	InputDir string `yaml:"input_dir" mapstructure:"input_dir"`
	// OutputFile is a file where the plugin scanner will write the result
	OutputFile string `yaml:"output_file" mapstructure:"output_file"`
	// ScannerConfig is a json string that will be passed to the scanner in the plugin
	ScannerConfig string `yaml:"scanner_config" mapstructure:"scanner_config"`
}

var _ PluginRunner = &Runner{}

type Runner struct {
	config       PluginConfig
	client       runnerclient.ClientWithResponsesInterface
	dockerClient *client.Client
	containerID  string
}

func New(config PluginConfig) (*Runner, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &Runner{
		dockerClient: dockerClient,
		config:       config,
	}, nil
}

func (r *Runner) Start(ctx context.Context) (CleanupFunc, error) {
	// Write scanner config file to temp dir
	err := os.WriteFile(getScannerConfigSourcePath(r.config.Name), []byte(r.config.ScannerConfig), 0o600) // nolint:gomnd
	if err != nil {
		return nil, fmt.Errorf("failed write scanner config file: %w", err)
	}

	// Pull scanner image if required
	err = PullImage(ctx, r.dockerClient, r.config.ImageName)
	if err != nil {
		return nil, fmt.Errorf("failed to pull scanner image: %w", err)
	}

	// Create scanner container
	containerResp, err := r.dockerClient.ContainerCreate(
		ctx,
		&containertypes.Config{
			Image:        r.config.ImageName,
			Env:          []string{"PLUGIN_SERVER_LISTEN_ADDRESS=0.0.0.0:" + DefaultScannerServerPort},
			ExposedPorts: nat.PortSet{"8080/tcp": struct{}{}},
		},
		&containertypes.HostConfig{
			PortBindings: map[nat.Port][]nat.PortBinding{
				"8080/tcp": {
					{
						HostIP:   "127.0.0.1",
						HostPort: "",
					},
				},
			},
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: getScannerConfigSourcePath(r.config.Name),
					Target: getScannerConfigDestinationPath(),
				},
				{
					Type:   mount.TypeBind,
					Source: r.config.InputDir,
					Target: DefaultScannerInputDir,
				},
				{
					Type:   mount.TypeBind,
					Source: filepath.Dir(r.config.OutputFile),
					Target: DefaultScannerOutputDir,
				},
			},
		},
		nil,
		nil,
		r.config.Name,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create scanner container: %w", err)
	}
	r.containerID = containerResp.ID

	err = r.dockerClient.ContainerStart(ctx, containerResp.ID, containertypes.StartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start scanner container: %w", err)
	}

	err = r.createHTTPClient(ctx, 1*time.Second, 20*time.Second) //nolint:gomnd
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	return r.Stop, nil
}

func (r *Runner) WaitReady(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	ticker := time.NewTicker(DefaultPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("checking health of %s timed out", r.config.Name)

		case <-ticker.C:
			resp, err := r.client.GetHealthzWithResponse(ctx)
			if err != nil {
				return fmt.Errorf("failed to get scanner health: %w", err)
			}

			if resp.StatusCode() == 200 { //nolint:gomnd
				return nil
			}
		}
	}
}

func (r *Runner) Run(ctx context.Context) error {
	_, err := r.client.PostConfigWithResponse(
		ctx,
		types.PostConfigJSONRequestBody{
			ScannerConfig:  to.Ptr(r.config.ScannerConfig),
			InputDir:       DefaultScannerInputDir,
			OutputFile:     filepath.Join(DefaultScannerOutputDir, filepath.Base(r.config.OutputFile)),
			TimeoutSeconds: int(DefaultTimeout),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to post scanner config: %w", err)
	}

	return nil
}

func (r *Runner) WaitDone(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	ticker := time.NewTicker(DefaultPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("checking status of %s timed out", r.config.Name)

		case <-ticker.C:
			resp, err := r.client.GetStatusWithResponse(ctx)
			if err != nil {
				return fmt.Errorf("failed to get scanner status: %w", err)
			}

			if resp.JSON200.State == types.Done {
				return nil
			}
		}
	}
}

func (r *Runner) Result() (io.Reader, error) {
	_, err := os.Stat(r.config.OutputFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrScanNotDone
		}
	}

	file, err := os.Open(r.config.OutputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open scanner result file: %w", err)
	}

	return bufio.NewReader(file), nil
}

func (r *Runner) Stop(ctx context.Context) error {
	err := r.dockerClient.ContainerStop(ctx, r.containerID, containertypes.StopOptions{})
	if err != nil {
		return fmt.Errorf("failed to stop scanner container: %w", err)
	}

	err = r.dockerClient.ContainerRemove(ctx, r.containerID, containertypes.RemoveOptions{})
	if err != nil {
		return fmt.Errorf("failed to remove scanner container: %w", err)
	}

	// Remove scanner config file
	err = os.RemoveAll(getScannerConfigSourcePath(r.config.Name))
	if err != nil {
		return fmt.Errorf("failed to remove scanner config file: %w", err)
	}

	return nil
}
