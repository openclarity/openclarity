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
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types/filters"
	imagetypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"

	runnerclient "github.com/openclarity/vmclarity/plugins/runner/internal/client"
)

type CleanupFunc func(ctx context.Context) error

func PullImage(ctx context.Context, client *client.Client, imageName string) error {
	images, err := client.ImageList(ctx, imagetypes.ListOptions{
		Filters: filters.NewArgs(filters.Arg("reference", imageName)),
	})
	if err != nil {
		return fmt.Errorf("failed to get images: %w", err)
	}

	if len(images) == 0 {
		pullResp, err := client.ImagePull(ctx, imageName, imagetypes.PullOptions{})
		if err != nil {
			return fmt.Errorf("failed to pull image: %w", err)
		}

		// consume output
		_, _ = io.Copy(io.Discard, pullResp)
		_ = pullResp.Close()
	}

	return nil
}

func getScannerConfigSourcePath(name string) string {
	return filepath.Join(os.TempDir(), name+"-plugin.json")
}

func getScannerConfigDestinationPath() string {
	return filepath.Join("/plugin.json")
}

// Create http client for plugin scanner:
// * try connecting to the plugin scanner container with container IP address
// * if that fails, try connecting to the plugin scanner container with host IP address.
func (r *Runner) createHTTPClient(ctx context.Context, pollInterval, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("checking status of %s timed out", r.config.Name)

		case <-ticker.C:
			inspect, err := r.dockerClient.ContainerInspect(ctx, r.containerID)
			if err != nil {
				return fmt.Errorf("failed to inspect scanner container: %w", err)
			}

			// Create client
			scannerIP := inspect.NetworkSettings.IPAddress
			err = r.tryPluginScannerServer(ctx, "http://"+scannerIP+":8080")
			if err != nil {
				fmt.Printf("failed to use scanner IP address, trying scanner host IP address\n")
				hostPort := inspect.NetworkSettings.Ports["8080/tcp"][0].HostPort
				err = r.tryPluginScannerServer(ctx, "http://127.0.0.1:"+hostPort+"/")
				if err != nil {
					return errors.New("failed to create client")
				}
			}

			return nil
		}
	}
}

func (r *Runner) tryPluginScannerServer(ctx context.Context, server string) error {
	var err error

	r.client, err = runnerclient.NewClientWithResponses(server)
	if err != nil {
		return fmt.Errorf("failed to create plugin client: %w", err)
	}

	newCtx, cancel := context.WithTimeout(ctx, 2*time.Second) //nolint:gomnd
	defer cancel()
	_, err = r.client.GetStatusWithResponse(newCtx)
	if err != nil {
		return fmt.Errorf("failed to ping plugin scanner server: %w", err)
	}

	return nil
}
