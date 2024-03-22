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
	"sync"

	imagetypes "github.com/docker/docker/api/types/image"
	"golang.org/x/sync/errgroup"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/core/to"
)

func (d *Discoverer) getImageAssets(ctx context.Context) ([]apitypes.AssetType, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	// List all docker images
	images, err := d.DockerClient.ImageList(ctx, imagetypes.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	// Results will be written to assets concurrently
	assetMu := sync.Mutex{}
	assets := make([]apitypes.AssetType, 0, len(images))

	// Process each image in an independent processor goroutine
	processGroup, processCtx := errgroup.WithContext(ctx)
	for _, image := range images {
		processGroup.Go(
			// processGroup expects a function with empty signature, so we use a function
			// generator to enable adding arguments. This avoids issues when using loop
			// variables in goroutines via shared memory space.
			//
			// If any processor returns an error, it will stop all processors.
			// IDEA: Decide what the acceptance criteria should be (e.g. >= 50% images processed)
			func(image imagetypes.Summary) func() error {
				return func() error {
					// Get container image info
					info, err := d.getContainerImageInfo(processCtx, image.ID)
					if err != nil {
						logger.Warnf("Failed to get image. id=%v: %v", image.ID, err)
						return nil // skip fail
					}

					// Convert to asset
					asset := apitypes.AssetType{}
					err = asset.FromContainerImageInfo(info)
					if err != nil {
						return fmt.Errorf("failed to create AssetType from ContainerImageInfo: %w", err)
					}

					// Write to assets
					assetMu.Lock()
					assets = append(assets, asset)
					assetMu.Unlock()

					return nil
				}
			}(image),
		)
	}

	// This will block until all the processors have executed successfully or until
	// the first error. If an error is returned by any processors, processGroup will
	// cancel execution via processCtx and return that error.
	err = processGroup.Wait()
	if err != nil {
		return nil, fmt.Errorf("failed to process images: %w", err)
	}

	return assets, nil
}

func (d *Discoverer) getContainerImageInfo(ctx context.Context, imageID string) (apitypes.ContainerImageInfo, error) {
	image, _, err := d.DockerClient.ImageInspectWithRaw(ctx, imageID)
	if err != nil {
		return apitypes.ContainerImageInfo{}, fmt.Errorf("failed to inspect image: %w", err)
	}

	return apitypes.ContainerImageInfo{
		Architecture: to.Ptr(image.Architecture),
		ImageID:      image.ID,
		Labels:       convertTags(image.Config.Labels),
		RepoTags:     to.Ptr(image.RepoTags),
		RepoDigests:  to.Ptr(image.RepoDigests),
		ObjectType:   "ContainerImageInfo",
		Os:           to.Ptr(image.Os),
		Size:         to.Ptr(image.Size),
	}, nil
}
