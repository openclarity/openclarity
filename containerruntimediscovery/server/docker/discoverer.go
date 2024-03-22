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

	dcontainer "github.com/docker/docker/api/types/container"
	dfilters "github.com/docker/docker/api/types/filters"
	imagetypes "github.com/docker/docker/api/types/image"
	dclient "github.com/docker/docker/client"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/containerruntimediscovery/types"
	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/core/to"
)

type discoverer struct {
	client *dclient.Client
}

var _ types.Discoverer = &discoverer{}

func New() (types.Discoverer, error) {
	client, err := dclient.NewClientWithOpts(dclient.FromEnv, dclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &discoverer{
		client: client,
	}, nil
}

func (d *discoverer) Images(ctx context.Context) ([]apitypes.ContainerImageInfo, error) {
	// List all docker images
	images, err := d.client.ImageList(ctx, imagetypes.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	// Convert images to container image info
	result := make([]apitypes.ContainerImageInfo, len(images))
	for i, image := range images {
		ii, err := d.getContainerImageInfo(ctx, image.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert image to ContainerImageInfo: %w", err)
		}
		result[i] = ii
	}
	return result, nil
}

func (d *discoverer) Image(ctx context.Context, imageID string) (apitypes.ContainerImageInfo, error) {
	info, err := d.getContainerImageInfo(ctx, imageID)
	if dclient.IsErrNotFound(err) {
		return info, types.ErrNotFound
	}
	return info, err
}

func (d *discoverer) getContainerImageInfo(ctx context.Context, imageID string) (apitypes.ContainerImageInfo, error) {
	image, _, err := d.client.ImageInspectWithRaw(ctx, imageID)
	if err != nil {
		return apitypes.ContainerImageInfo{}, fmt.Errorf("failed to inspect image: %w", err)
	}

	return apitypes.ContainerImageInfo{
		Architecture: to.Ptr(image.Architecture),
		ImageID:      image.ID,
		Labels:       apitypes.MapToTags(image.Config.Labels),
		RepoTags:     &image.RepoTags,
		RepoDigests:  &image.RepoDigests,
		ObjectType:   "ContainerImageInfo",
		Os:           to.Ptr(image.Os),
		Size:         to.Ptr(image.Size),
	}, nil
}

func (d *discoverer) ExportImage(ctx context.Context, imageID string) (io.ReadCloser, error) {
	reader, err := d.client.ImageSave(ctx, []string{imageID})
	if err != nil {
		return nil, fmt.Errorf("unable to save image from daemon: %w", err)
	}
	return reader, nil
}

func (d *discoverer) Containers(ctx context.Context) ([]apitypes.ContainerInfo, error) {
	// List all docker containers
	containers, err := d.client.ContainerList(ctx, dcontainer.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	result := make([]apitypes.ContainerInfo, len(containers))
	for i, container := range containers {
		// Get container info
		info, err := d.getContainerInfo(ctx, container.ID, container.ImageID)
		if err != nil {
			return nil, fmt.Errorf("failed to convert container to ContainerInfo: %w", err)
		}
		result[i] = info
	}
	return result, nil
}

func (d *discoverer) Container(ctx context.Context, containerID string) (apitypes.ContainerInfo, error) {
	// List all docker containers filtered by containerID
	containers, err := d.client.ContainerList(ctx, dcontainer.ListOptions{
		All:     true,
		Filters: dfilters.NewArgs(dfilters.KeyValuePair{Key: "id", Value: containerID}),
	})
	if err != nil {
		return apitypes.ContainerInfo{}, fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		return apitypes.ContainerInfo{}, types.ErrNotFound
	}
	if len(containers) > 1 {
		return apitypes.ContainerInfo{}, fmt.Errorf("found more than one container with id %s", containerID)
	}

	info, err := d.getContainerInfo(ctx, containers[0].ID, containers[0].ImageID)
	if err != nil {
		return apitypes.ContainerInfo{}, fmt.Errorf("failed to convert container to ContainerInfo: %w", err)
	}
	return info, nil
}

func (d *discoverer) ExportContainer(ctx context.Context, containerID string) (io.ReadCloser, func(), error) {
	clean := &types.Cleanup{}
	defer clean.Clean()

	imageName := "vmclarity.io/container-snapshot:" + containerID
	idresp, err := d.client.ContainerCommit(ctx, containerID, dcontainer.CommitOptions{
		Reference: imageName,
		Comment:   fmt.Sprintf("Snapshot of container %s for security scanning", containerID),
		Author:    "VMClarity",
		Pause:     false,
	})
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to commit container to image: %w", err)
	}
	clean.Add(func() {
		_, err := d.client.ImageRemove(ctx, idresp.ID, imagetypes.RemoveOptions{
			Force: true,
		})
		if err != nil {
			log.GetLoggerFromContextOrDefault(ctx).Errorf("failed to cleanup container snapshot: %v", err)
		}
	})

	reader, err := d.client.ImageSave(ctx, []string{idresp.ID})
	if err != nil {
		return nil, func() {}, fmt.Errorf("unable to save image from daemon: %w", err)
	}

	return reader, clean.Release(), nil
}

func (d *discoverer) getContainerInfo(ctx context.Context, containerID, imageID string) (apitypes.ContainerInfo, error) {
	// Inspect container
	info, err := d.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return apitypes.ContainerInfo{}, fmt.Errorf("failed to inspect container: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, info.Created)
	if err != nil {
		return apitypes.ContainerInfo{}, fmt.Errorf("failed to parse time: %w", err)
	}

	// Get container image info
	imageInfo, err := d.getContainerImageInfo(ctx, imageID)
	if err != nil {
		return apitypes.ContainerInfo{}, err
	}

	return apitypes.ContainerInfo{
		ContainerName: to.Ptr(info.Name),
		CreatedAt:     to.Ptr(createdAt),
		ContainerID:   containerID,
		Image:         to.Ptr(imageInfo),
		Labels:        apitypes.MapToTags(info.Config.Labels),
		ObjectType:    "ContainerInfo",
	}, nil
}

func (d *discoverer) Ready(ctx context.Context) (bool, error) {
	_, err := d.client.Ping(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get connection state: %w", err)
	}

	return true, nil
}
