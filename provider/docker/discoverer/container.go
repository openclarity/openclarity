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

package discoverer

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"golang.org/x/sync/errgroup"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/core/to"
)

func (d *Discoverer) getContainerAssets(ctx context.Context) ([]apitypes.AssetType, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	// List all docker containers
	containers, err := d.DockerClient.ContainerList(ctx, containertypes.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	// Results will be written to assets concurrently
	assetMu := sync.Mutex{}
	assets := make([]apitypes.AssetType, 0, len(containers))

	// Process each container in an independent processor goroutine
	processGroup, processCtx := errgroup.WithContext(ctx)
	for _, container := range containers {
		processGroup.Go(
			// processGroup expects a function with empty signature, so we use a function
			// generator to enable adding arguments. This avoids issues when using loop
			// variables in goroutines via shared memory space.
			//
			// If any processor returns an error, it will stop all processors.
			// IDEA: Decide what the acceptance criteria should be (e.g. >= 50% container processed)
			func(container types.Container) func() error {
				return func() error {
					// Get container info
					info, err := d.getContainerInfo(processCtx, container.ID)
					if err != nil {
						logger.Warnf("Failed to get container. id=%v: %v", container.ID, err)
						return nil // skip fail
					}

					// Convert to asset
					asset := apitypes.AssetType{}
					err = asset.FromContainerInfo(info)
					if err != nil {
						return fmt.Errorf("failed to create AssetType from ContainerInfo: %w", err)
					}

					// Write to assets
					assetMu.Lock()
					assets = append(assets, asset)
					assetMu.Unlock()

					return nil
				}
			}(container),
		)
	}

	// This will block until all the processors have executed successfully or until
	// the first error. If an error is returned by any processors, processGroup will
	// cancel execution via processCtx and return that error.
	err = processGroup.Wait()
	if err != nil {
		return nil, fmt.Errorf("failed to process containers: %w", err)
	}

	return assets, nil
}

func (d *Discoverer) getContainerInfo(ctx context.Context, containerID string) (apitypes.ContainerInfo, error) {
	// Inspect container
	info, err := d.DockerClient.ContainerInspect(ctx, containerID)
	if err != nil {
		return apitypes.ContainerInfo{}, fmt.Errorf("failed to inspect container: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, info.Created)
	if err != nil {
		return apitypes.ContainerInfo{}, fmt.Errorf("failed to parse time: %w", err)
	}

	// Get container image info
	imageInfo, err := d.getContainerImageInfo(ctx, info.Image)
	if err != nil {
		return apitypes.ContainerInfo{}, err
	}

	containerName := strings.Trim(info.Name, "/")

	return apitypes.ContainerInfo{
		ContainerName: &containerName,
		CreatedAt:     to.Ptr(createdAt),
		ContainerID:   containerID,
		Image:         to.Ptr(imageInfo),
		Labels:        convertTags(info.Config.Labels),
		ObjectType:    "ContainerInfo",
	}, nil
}
