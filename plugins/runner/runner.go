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
	"github.com/docker/docker/client"

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

// rename this to Runner when the old version is stripped down, this controls the flow for running a scan via plugin scanner
// flow WaitReady -> Start -> WaitDone -> Result. All methods only interacts with HTTP endpoint.
type PluginRunner interface {
	WaitReady(ctx context.Context) error // use fixed interval of 2s for polling, use ctx for timeout, retry logic for requests should be added as container might not start right away
	Start(ctx context.Context) error     // sends a post request to container to start the scanning
	WaitDone(ctx context.Context) error  // use fixed interval of 2s for polling, use ctx for timeout, retry logic for requests should be added as container might not start right away
	Result() (io.Reader, error)          // return ErrScanNotDone if the scan is not done yet, otherwise return file stream of the result so that we can read and parse it upstream
	Stop(ctx context.Context) error      // stop and remove the container
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
	client       runnerclient.ClientWithResponsesInterface
	dockerClient *client.Client
	containerID  string

	PluginConfig
}

func (r *Runner) WaitReady(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	ticker := time.NewTicker(DefaultPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("checking health of %s timed out", r.PluginConfig.Name)

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

func (r *Runner) Start(ctx context.Context) error {
	_, err := r.client.PostConfigWithResponse(
		ctx,
		types.PostConfigJSONRequestBody{
			File:           to.Ptr(getScannerConfigDestinationPath()),
			InputDir:       DefaultScannerInputDir,
			OutputFile:     filepath.Join(DefaultScannerOutputDir, filepath.Base(r.OutputFile)),
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
			return fmt.Errorf("checking status of %s timed out", r.PluginConfig.Name)

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
	_, err := os.Stat(r.OutputFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrScanNotDone
		}
	}

	file, err := os.Open(r.OutputFile)
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
	err = os.RemoveAll(getScannerConfigSourcePath(r.Name))
	if err != nil {
		return fmt.Errorf("failed to remove scanner config file: %w", err)
	}

	return nil
}
