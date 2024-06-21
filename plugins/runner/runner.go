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
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/openclarity/vmclarity/plugins/runner/internal/containermanager"
	"github.com/openclarity/vmclarity/plugins/runner/internal/containermanager/docker"
	"github.com/openclarity/vmclarity/plugins/runner/types"

	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/core/to"
	runnerclient "github.com/openclarity/vmclarity/plugins/runner/internal/client"
	plugintypes "github.com/openclarity/vmclarity/plugins/sdk-go/types"
)

const defaultPollInterval = 2 * time.Second

type pluginRunner struct {
	config           types.PluginConfig
	containerManager containermanager.PluginContainerManager
	client           runnerclient.ClientWithResponsesInterface
}

func New(ctx context.Context, config types.PluginConfig) (types.PluginRunner, error) {
	// Create docker container
	// TODO: switch to factory once the support for more container engines is added
	manager, err := docker.New(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin manager: %w", err)
	}

	return &pluginRunner{
		config:           config,
		containerManager: manager,
	}, nil
}

func (r *pluginRunner) Start(ctx context.Context) error {
	if err := r.containerManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to create plugin container: %w", err)
	}

	return nil
}

func (r *pluginRunner) Logs(ctx context.Context) (io.ReadCloser, error) {
	logs, err := r.containerManager.Logs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load logs from container: %w", err)
	}

	return logs, nil
}

func (r *pluginRunner) WaitReady(ctx context.Context) error {
	logger := log.GetLoggerFromContextOrDefault(ctx)

	ctx, cancel := context.WithTimeout(ctx, types.WaitReadyTimeout)
	defer cancel()

	ticker := time.NewTicker(defaultPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting ready state for scanner %s timed out", r.config.Name)

		case <-ticker.C:
			// Check if plugin container is ready
			ready, err := r.containerManager.Ready()
			if err != nil {
				return fmt.Errorf("failed to check plugin container state: %w", err)
			}
			if !ready {
				continue
			}

			// Set plugin server endpoint and client if not already set
			if r.client == nil {
				serverEndpoint, err := r.containerManager.GetPluginServerEndpoint(ctx)
				if err != nil {
					return fmt.Errorf("failed to get plugin server endpoint: %w", err)
				}

				err = r.setPluginServerClientFor(serverEndpoint)
				if err != nil {
					return fmt.Errorf("failed to set plugin server client: %w", err)
				}
			}

			// Check for plugin server once container is ready
			resp, err := r.client.GetHealthzWithResponse(ctx)
			if err != nil {
				logger.WithError(err).Error("failed to get plugin server healthz, retrying...")
				continue
			}

			if resp.StatusCode() == 200 { //nolint:mnd
				return nil
			}
		}
	}
}

func (r *pluginRunner) Metadata(ctx context.Context) (*plugintypes.Metadata, error) {
	if r.client == nil {
		return nil, errors.New("client missing, did not wait for ready state")
	}

	metadata, err := r.client.GetMetadataWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to post scan config: %w", err)
	}

	return metadata.JSON200, nil
}

func (r *pluginRunner) Run(ctx context.Context) error {
	if r.client == nil {
		return errors.New("client missing, did not wait for ready state")
	}

	_, err := r.client.PostConfigWithResponse(
		ctx,
		containermanager.WithOverrides(plugintypes.Config{
			ScannerConfig:  to.Ptr(r.config.ScannerConfig),
			TimeoutSeconds: int(types.ScanTimeout.Seconds()),
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to post scan config: %w", err)
	}

	return nil
}

func (r *pluginRunner) Stop(ctx context.Context) error {
	if r.client == nil {
		return errors.New("client missing, did not wait for ready state")
	}

	_, err := r.client.PostStopWithResponse(
		ctx,
		plugintypes.PostStopJSONRequestBody{
			TimeoutSeconds: int(types.GracefulStopTimeout.Seconds()),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to post scan stop: %w", err)
	}

	return nil
}

func (r *pluginRunner) WaitDone(ctx context.Context) error {
	if r.client == nil {
		return errors.New("client missing, did not wait for ready state")
	}

	ctx, cancel := context.WithTimeout(ctx, types.ScanTimeout)
	defer cancel()

	ticker := time.NewTicker(defaultPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting done state for scanner %s timed out", r.config.Name)

		case <-ticker.C:
			resp, err := r.client.GetStatusWithResponse(ctx)
			if err != nil {
				return fmt.Errorf("failed to get scanner status: %w", err)
			}

			if resp.JSON200.State == plugintypes.StateDone {
				return nil
			}
			if resp.JSON200.State == plugintypes.StateFailed {
				var reason string
				if resp.JSON200.Message != nil {
					reason = *resp.JSON200.Message
				}
				return fmt.Errorf("scan failed, reason: %s", reason)
			}
		}
	}
}

func (r *pluginRunner) Result(ctx context.Context) (io.ReadCloser, error) {
	result, err := r.containerManager.Result(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get result from container: %w", err)
	}

	return result, nil
}

func (r *pluginRunner) Remove(ctx context.Context) error {
	err := r.containerManager.Remove(ctx)
	if err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	return nil
}

func (r *pluginRunner) setPluginServerClientFor(server string) error {
	client, err := runnerclient.NewClientWithResponses(server)
	if err != nil {
		return fmt.Errorf("could not create client for plugin server: %w", err)
	}

	r.client = client

	return nil
}
