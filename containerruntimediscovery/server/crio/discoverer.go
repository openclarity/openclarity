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

package crio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	apitypes "github.com/openclarity/openclarity/api/types"
	"github.com/openclarity/openclarity/containerruntimediscovery/types"
	"github.com/openclarity/openclarity/core/to"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports/alltransports"
	imageTypes "github.com/containers/image/v5/types"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	cri "k8s.io/cri-api/pkg/apis"
	v1 "k8s.io/cri-api/pkg/apis/runtime/v1"
	remote "k8s.io/cri-client/pkg"
	"k8s.io/klog/v2"
)

const (
	CRIOSockAddress      = "unix:///var/run/crio/crio.sock"
	DefaultClientTimeout = 2 * time.Second
)

type discoverer struct {
	runtimeService cri.RuntimeService
	imageService   cri.ImageManagerService
}

var _ types.Discoverer = &discoverer{}

func New() (types.Discoverer, error) {
	logger := klog.Background()
	var tp trace.TracerProvider = noop.NewTracerProvider()

	r, err := remote.NewRemoteRuntimeService(CRIOSockAddress, DefaultClientTimeout, tp, &logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create CRIO runtime service client: %w", err)
	}

	i, err := remote.NewRemoteImageService(CRIOSockAddress, DefaultClientTimeout, tp, &logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create CRIO image service client: %w", err)
	}

	return &discoverer{
		runtimeService: r,
		imageService:   i,
	}, nil
}

func (d *discoverer) Images(ctx context.Context) ([]apitypes.ContainerImageInfo, error) {
	images, err := d.imageService.ListImages(ctx, &v1.ImageFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	result := make([]apitypes.ContainerImageInfo, len(images))
	for i, image := range images {
		imageInfo, err := d.getContainerImageInfo(ctx, image.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to convert image to ContainerImageInfo: %w", err)
		}

		result[i] = imageInfo
	}

	return result, nil
}

func (d *discoverer) Image(ctx context.Context, imageID string) (apitypes.ContainerImageInfo, error) {
	imageInfo, err := d.getContainerImageInfo(ctx, imageID)
	if err != nil {
		return apitypes.ContainerImageInfo{}, fmt.Errorf("failed to convert image to ContainerImageInfo: %w", err)
	}

	return imageInfo, nil
}

func (d *discoverer) getContainerImageInfo(ctx context.Context, imageID string) (apitypes.ContainerImageInfo, error) {
	images, err := d.imageService.ListImages(ctx, &v1.ImageFilter{
		Image: &v1.ImageSpec{Image: imageID},
	})
	if err != nil {
		return apitypes.ContainerImageInfo{}, fmt.Errorf("failed to list images: %w", err)
	}

	if len(images) == 0 {
		return apitypes.ContainerImageInfo{}, types.ErrNotFound
	}
	if len(images) > 1 {
		return apitypes.ContainerImageInfo{}, fmt.Errorf("found more than one container with id %s", imageID)
	}

	image := images[0]

	resp, err := d.imageService.ImageStatus(ctx, &v1.ImageSpec{Image: image.Id}, true)
	if err != nil {
		return apitypes.ContainerImageInfo{}, fmt.Errorf("failed to get image status: %w", err)
	}

	if _, ok := resp.Info["info"]; !ok {
		return apitypes.ContainerImageInfo{}, errors.New("failed to parse image status result: info field does not exist")
	}

	type info struct {
		ImageSpec struct {
			Architecture string `json:"architecture"`
			OS           string `json:"os"`
		} `json:"imageSpec"`
	}

	var i info
	err = json.Unmarshal([]byte(resp.Info["info"]), &i)
	if err != nil {
		return apitypes.ContainerImageInfo{}, fmt.Errorf("failed to parse image status result: %w", err)
	}

	return apitypes.ContainerImageInfo{
		Architecture: to.Ptr(i.ImageSpec.Architecture),
		ImageID:      image.Id,
		Labels:       apitypes.MapToTags(resp.Image.Spec.Annotations),
		RepoTags:     &image.RepoTags,
		RepoDigests:  &image.RepoDigests,
		ObjectType:   "ContainerImageInfo",
		Os:           to.Ptr(i.ImageSpec.OS),
		Size:         to.Ptr(int64(image.Size_)),
	}, nil
}

func (d *discoverer) ExportImage(ctx context.Context, imageID string) (io.ReadCloser, func(), error) {
	clean := &types.Cleanup{}
	defer clean.Clean()

	imageInfo, err := d.getContainerImageInfo(ctx, imageID)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to get image info by id: %w", err)
	}

	if len(*imageInfo.RepoDigests) == 0 {
		return nil, func() {}, fmt.Errorf("failed to determine image digest: %w", err)
	}

	digest := (*imageInfo.RepoDigests)[0]

	src, err := alltransports.ParseImageName("containers-storage:" + digest)
	if err != nil {
		return nil, func() {}, fmt.Errorf("error parsing image name: %w", err)
	}

	destFilePath := filepath.Join(os.TempDir(), uuid.New().String()+"-image.tar")

	dest, err := alltransports.ParseImageName("docker-archive:" + destFilePath)
	if err != nil {
		return nil, func() {}, fmt.Errorf("error creating destination file: %w", err)
	}

	systemContext := &imageTypes.SystemContext{}
	policyContext, err := signature.NewPolicyContext(&signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}})
	if err != nil {
		return nil, func() {}, fmt.Errorf("error creating policy context: %w", err)
	}
	//nolint:errcheck
	defer policyContext.Destroy()

	_, err = copy.Image(ctx, policyContext, dest, src, &copy.Options{
		SourceCtx:      systemContext,
		DestinationCtx: systemContext,
	})
	if err != nil {
		return nil, func() {}, fmt.Errorf("error copying image: %w", err)
	}

	destFile, err := os.Open(destFilePath)
	if err != nil {
		return nil, func() {}, fmt.Errorf("error opening image archive: %w", err)
	}

	clean.Add(func() {
		destFile.Close()
		os.Remove(destFilePath)
	})

	return destFile, clean.Release(), nil
}

func (d *discoverer) Containers(ctx context.Context) ([]apitypes.ContainerInfo, error) {
	containers, err := d.runtimeService.ListContainers(ctx, &v1.ContainerFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	result := make([]apitypes.ContainerInfo, len(containers))
	for i, container := range containers {
		containerInfo, err := d.getContainerInfo(ctx, container.Id)
		if err != nil {
			return nil, fmt.Errorf("unable to get container info: %w", err)
		}

		result[i] = containerInfo
	}

	return result, nil
}

func (d *discoverer) Container(ctx context.Context, containerID string) (apitypes.ContainerInfo, error) {
	containerInfo, err := d.getContainerInfo(ctx, containerID)
	if err != nil {
		return apitypes.ContainerInfo{}, fmt.Errorf("unable to get container info: %w", err)
	}

	return containerInfo, nil
}

func (d *discoverer) getContainerInfo(ctx context.Context, containerID string) (apitypes.ContainerInfo, error) {
	containers, err := d.runtimeService.ListContainers(ctx, &v1.ContainerFilter{
		Id: containerID,
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

	container := containers[0]

	imageInfo, err := d.getContainerImageInfo(ctx, container.ImageId)
	if err != nil {
		return apitypes.ContainerInfo{}, fmt.Errorf("failed to convert image to ContainerImageInfo: %w", err)
	}

	seconds := container.CreatedAt / int64(time.Second)
	nanos := container.CreatedAt % int64(time.Second)
	createdAt := time.Unix(seconds, nanos)

	return apitypes.ContainerInfo{
		ContainerName: to.Ptr(container.Metadata.Name),
		CreatedAt:     to.Ptr(createdAt),
		ContainerID:   container.Id,
		Image:         to.Ptr(imageInfo),
		Labels:        apitypes.MapToTags(container.Labels),
		ObjectType:    "ContainerInfo",
	}, nil
}

func (d *discoverer) ExportContainer(ctx context.Context, containerID string) (io.ReadCloser, func(), error) {
	err := d.runtimeService.CheckpointContainer(ctx, &v1.CheckpointContainerRequest{
		ContainerId: containerID,
		Location:    fmt.Sprintf("/tmp/%s.tar", containerID),
	})
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to checkpoint container: %w", err)
	}

	return nil, func() {}, nil
}

func (d *discoverer) Ready(ctx context.Context) (bool, error) {
	_, err := d.runtimeService.Status(ctx, false)
	if err != nil {
		return false, fmt.Errorf("failed to get connection state: %w", err)
	}

	return true, nil
}
