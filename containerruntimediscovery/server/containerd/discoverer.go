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

package containerd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/defaults"
	"github.com/containerd/containerd/errdefs"
	containerdImages "github.com/containerd/containerd/images"
	"github.com/containerd/containerd/images/archive"
	"github.com/containerd/containerd/leases"
	criConstants "github.com/containerd/containerd/pkg/cri/constants"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/nerdctl/pkg/imgutil"
	"github.com/containerd/nerdctl/pkg/imgutil/commit"
	"github.com/containerd/nerdctl/pkg/labels/k8slabels"
	"github.com/containers/image/v5/docker/reference"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/containerruntimediscovery/types"
	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/core/to"
)

const (
	ContainerdSockAddress  = "/var/run/containerd/containerd.sock"
	DefaultSnapshotterName = "overlayfs"
	DefaultClientTimeout   = 30 * time.Second
	DefaultNamespace       = criConstants.K8sContainerdNamespace
	DefaultSkipUnpackImage = false
)

type discoverer struct {
	client          *containerd.Client
	snapshotterName string
	namespace       string
	skipUnpackImage bool
}

var _ types.Discoverer = &discoverer{}

func New() (types.Discoverer, error) {
	// Containerd supports multiple namespaces so that a single daemon can
	// be used by multiple clients like Docker and Kubernetes and the
	// resources will not conflict etc. In order to discover all the
	// containers for kubernetes we need to set the kubernetes namespace as
	// the default for our client.
	client, err := containerd.New(ContainerdSockAddress,
		containerd.WithDefaultNamespace(DefaultNamespace),
		containerd.WithTimeout(DefaultClientTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client: %w", err)
	}

	return &discoverer{
		client:          client,
		snapshotterName: DefaultSnapshotterName,
		namespace:       DefaultNamespace,
		skipUnpackImage: DefaultSkipUnpackImage,
	}, nil
}

// StopWalk is used as a return value from ImageIDWalkFunc to indicate that all remaining images are to be skipped.
//
//nolint:errname,stylecheck
var StopWalk = errors.New("stop the walk")

// ImageIDWalkFunc is the type of function which is invoked by the discoverer.imageIDWalk method to visit all container
// images with the provided image ID.
type ImageIDWalkFunc func(image containerd.Image) error

// imageIDWalk iterates over the list of container images returned by containerd and invokes the provided ImageIDWalkFunc
// for images which have the ID defined by imageID.
func (d *discoverer) imageIDWalk(ctx context.Context, imageID string, fn ImageIDWalkFunc) error {
	images, err := d.client.ListImages(ctx)
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	for _, image := range images {
		configDescriptor, err := image.Config(ctx)
		if err != nil {
			return fmt.Errorf("failed to load image config descriptor: %w", err)
		}

		if configDescriptor.Digest.String() != imageID {
			continue
		}

		if err = fn(image); err != nil {
			if errors.Is(err, StopWalk) {
				break
			}

			return err
		}
	}

	return nil
}

func (d *discoverer) Images(ctx context.Context) ([]apitypes.ContainerImageInfo, error) {
	logger := log.GetLoggerFromContextOrDefault(ctx)

	images, err := d.client.ListImages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	imageSet := map[string]apitypes.ContainerImageInfo{}
	for _, image := range images {
		// Ignore our transient images used for snapshoting the
		// containers they will be cleaned up as soon as the export is
		// done.
		if strings.HasPrefix(image.Name(), "vmclarity.io/container-snapshot:") {
			continue
		}

		imageInfo, err := d.getContainerImageInfo(ctx, image)
		if err != nil {
			return nil, fmt.Errorf("unable to convert image %s to container image info: %w", image.Name(), err)
		}

		if imageInfo.ImageID == "" {
			logger.Warnf("found image with empty ImageID: %s", imageInfo.String())
			continue
		}

		existing, ok := imageSet[imageInfo.ImageID]
		if ok {
			merged, err := existing.Merge(imageInfo)
			if err != nil {
				return nil, fmt.Errorf("unable to merge image %v with %v: %w", existing, imageInfo, err)
			}
			imageInfo = merged
		}
		imageSet[imageInfo.ImageID] = imageInfo
	}

	result := []apitypes.ContainerImageInfo{}
	for _, image := range imageSet {
		result = append(result, image)
	}

	return result, nil
}

func (d *discoverer) Image(ctx context.Context, imageID string) (apitypes.ContainerImageInfo, error) {
	var result apitypes.ContainerImageInfo
	var found bool

	// Containerd doesn't allow to filter images by config digest (aka image ID), so we have to walk all the images
	// to find all images with the same ID and then merge them together.
	walkFn := func(image containerd.Image) error {
		found = true

		containerImageInfo, err := d.getContainerImageInfo(ctx, image)
		if err != nil {
			return fmt.Errorf("unable to convert image %s to container image info: %w", image.Name(), err)
		}

		result, err = result.Merge(containerImageInfo)
		if err != nil {
			return fmt.Errorf("unable to merge image %v with %v: %w", result, containerImageInfo, err)
		}

		return nil
	}

	err := d.imageIDWalk(ctx, imageID, walkFn)
	if err != nil {
		return apitypes.ContainerImageInfo{}, fmt.Errorf("failed to walk all image: %w", err)
	}
	if !found {
		return apitypes.ContainerImageInfo{}, types.ErrNotFound
	}

	return result, nil
}

func (d *discoverer) getContainerImageInfo(ctx context.Context, image containerd.Image) (apitypes.ContainerImageInfo, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	configDescriptor, err := image.Config(ctx)
	if err != nil {
		return apitypes.ContainerImageInfo{}, fmt.Errorf("failed to load image config descriptor: %w", err)
	}
	id := configDescriptor.Digest.String()

	imageSpec, err := image.Spec(ctx)
	if err != nil {
		return apitypes.ContainerImageInfo{}, fmt.Errorf("failed to load image spec: %w", err)
	}

	// Try to calculate uncompressed size of the image
	snapshotterName := d.getSnapshotterName(ctx)
	unpacked, err := image.IsUnpacked(ctx, snapshotterName)
	if err != nil {
		return apitypes.ContainerImageInfo{}, fmt.Errorf("failed to determine whether the image is unpacked or not: %w", err)
	}

	if !unpacked && !d.skipUnpackImage {
		if err = image.Unpack(ctx, snapshotterName); err != nil {
			logger.Warnf("failed to unpack image %s: %v", id, err)
		} else {
			unpacked = true
		}
	}

	var size int64
	if unpacked {
		// Get uncompressed image size
		size, err = imgutil.UnpackedImageSize(ctx, d.client.SnapshotService(snapshotterName), image)
		if err != nil {
			logger.Debugf("failed to get unpacked size of image %s. Falling back to the compressed size: %v", id, err)
		}
	}

	if size == 0 {
		// Get compressed image size
		size, err = image.Size(ctx)
		if err != nil {
			return apitypes.ContainerImageInfo{}, fmt.Errorf("failed to get compressed size of image %s: %w", id, err)
		}
	}

	repoTags, repoDigests := ParseImageReferences([]string{image.Name()})

	return apitypes.ContainerImageInfo{
		ImageID:      id,
		Architecture: to.Ptr(imageSpec.Architecture),
		Labels:       apitypes.MapToTags(imageSpec.Config.Labels),
		RepoTags:     &repoTags,
		RepoDigests:  &repoDigests,
		ObjectType:   "ContainerImageInfo",
		Os:           to.Ptr(imageSpec.OS),
		Size:         to.Ptr(size),
	}, nil
}

func (d discoverer) getSnapshotterName(ctx context.Context) string {
	name := d.snapshotterName
	label, _ := d.client.GetLabel(ctx, defaults.DefaultSnapshotterNSLabel)
	if label != "" {
		name = label
	}

	return name
}

// TODO(sambetts) Support auth config for fetching private images if they are missing.
func (d *discoverer) ExportImage(ctx context.Context, imageID string) (io.ReadCloser, error) {
	var img containerd.Image
	var found bool

	// Find the first container `image` with ID defined by `imageID`.
	walkFn := func(image containerd.Image) error {
		found = true
		img = image

		return StopWalk
	}

	err := d.imageIDWalk(ctx, imageID, walkFn)
	if err != nil {
		return nil, fmt.Errorf("failed to walk all images: %w", err)
	}
	if !found {
		return nil, types.ErrNotFound
	}

	// NOTE(sambetts) When running in Kubernetes containerd can be
	// configured to garbage collect the un-expanded blobs from the content
	// store after they are converted to a rootfs snapshot that is used to
	// boot containers. For this reason we need to re-fetch the image to
	// ensure that all the required blobs for export are in the content
	// store.
	// nolint: dogsled
	_, _, _, missing, err := containerdImages.Check(ctx, d.client.ContentStore(), img.Target(), platforms.Default())
	if err != nil {
		return nil, fmt.Errorf("unable to check image in content store: %w", err)
	}
	if len(missing) > 0 {
		imageInfo, err := d.Image(ctx, imageID)
		if err != nil {
			return nil, fmt.Errorf("failed to get image info to export: %w", err)
		}
		if imageInfo.RepoDigests == nil || len(*imageInfo.RepoDigests) == 0 {
			return nil, errors.New("image has no known repo digests can not safely fetch it")
		}

		// TODO(sambetts) Maybe try all the digests in case one has gone missing?
		ref := (*imageInfo.RepoDigests)[0]
		img, err = d.client.Pull(ctx, ref)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch image %s: %w", ref, err)
		}
	}

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		err := d.client.Export(
			ctx,
			pw,
			archive.WithImage(d.client.ImageService(), img.Name()),
			archive.WithPlatform(platforms.DefaultStrict()),
		)
		if err != nil {
			log.GetLoggerFromContextOrDefault(ctx).Errorf("failed to export image: %v", err)
		}
	}()
	return pr, nil
}

// ParseImageReferences parses a list of arbitrary image references and returns
// the repotags and repodigests.
func ParseImageReferences(refs []string) ([]string, []string) {
	var tags, digests []string
	for _, ref := range refs {
		parsed, err := reference.ParseAnyReference(ref)
		if err != nil {
			continue
		}
		if _, ok := parsed.(reference.Canonical); ok {
			digests = append(digests, parsed.String())
		} else if _, ok := parsed.(reference.Tagged); ok {
			tags = append(tags, parsed.String())
		}
	}
	return tags, digests
}

func (d *discoverer) Containers(ctx context.Context) ([]apitypes.ContainerInfo, error) {
	containers, err := d.client.Containers(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to list containers: %w", err)
	}

	result := make([]apitypes.ContainerInfo, len(containers))
	for i, container := range containers {
		// Get container info
		info, err := d.getContainerInfo(ctx, container)
		if err != nil {
			return nil, fmt.Errorf("failed to convert container to ContainerInfo: %w", err)
		}
		result[i] = info
	}
	return result, nil
}

func (d *discoverer) Container(ctx context.Context, containerID string) (apitypes.ContainerInfo, error) {
	container, err := d.client.LoadContainer(ctx, containerID)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return apitypes.ContainerInfo{}, types.ErrNotFound
		}
		return apitypes.ContainerInfo{}, fmt.Errorf("failed to get container from store: %w", err)
	}

	return d.getContainerInfo(ctx, container)
}

// nolint: cyclop
func (d *discoverer) ExportContainer(ctx context.Context, containerID string) (io.ReadCloser, func(), error) {
	clean := &types.Cleanup{}
	defer clean.Clean()

	container, err := d.client.LoadContainer(ctx, containerID)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil, func() {}, types.ErrNotFound
		}
		return nil, func() {}, fmt.Errorf("failed to get container from store: %w", err)
	}

	img, err := container.Image(ctx)
	if err != nil {
		return nil, func() {}, fmt.Errorf("unable to get image from container %s: %w", containerID, err)
	}

	// NOTE(sambetts) When running in Kubernetes containerd can be
	// configured to garbage collect the un-expanded blobs from the content
	// store after they are converted to a rootfs snapshot that is used to
	// boot containers. For this reason we need to re-fetch the image to
	// ensure that all the required blobs for export are in the content
	// store.
	// nolint: dogsled
	_, _, _, missing, err := containerdImages.Check(ctx, d.client.ContentStore(), img.Target(), platforms.Default())
	if err != nil {
		return nil, func() {}, fmt.Errorf("unable to check image in content store: %w", err)
	}
	if len(missing) > 0 {
		configDescriptor, err := img.Config(ctx)
		if err != nil {
			return nil, func() {}, fmt.Errorf("failed to load image config descriptor: %w", err)
		}
		imageID := configDescriptor.Digest.String()

		imageInfo, err := d.Image(ctx, imageID)
		if err != nil {
			return nil, func() {}, fmt.Errorf("failed to get image info to export: %w", err)
		}
		if imageInfo.RepoDigests == nil || len(*imageInfo.RepoDigests) == 0 {
			return nil, func() {}, errors.New("image has no known repo digests can not safely fetch it")
		}

		// TODO(sambetts) Maybe try all the digests in case one has gone missing?
		ref := (*imageInfo.RepoDigests)[0]
		_, err = d.client.Pull(ctx, ref)
		if err != nil {
			return nil, func() {}, fmt.Errorf("failed to fetch image %s: %w", ref, err)
		}
	}

	ctx, done, err := d.client.WithLease(ctx, leases.WithRandomID(), leases.WithExpiration(1*time.Hour))
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to get lease from containerd: %w", err)
	}
	clean.Add(func() {
		err := done(ctx)
		if err != nil {
			log.GetLoggerFromContextOrDefault(ctx).Errorf("failed to release lease: %v", err)
		}
	})

	imageName := "vmclarity.io/container-snapshot:" + containerID
	_, err = commit.Commit(ctx, d.client, container, &commit.Opts{
		Author:  "VMClarity",
		Message: fmt.Sprintf("Snapshot of container %s for security scanning", containerID),
		Ref:     imageName,
		Pause:   false,
	})
	if err != nil {
		return nil, func() {}, fmt.Errorf("unable to commit container to image: %w", err)
	}
	clean.Add(func() {
		err := d.client.ImageService().Delete(ctx, imageName)
		if err != nil {
			log.GetLoggerFromContextOrDefault(ctx).Errorf("failed to clean up snapshot %s for container %s: %v", imageName, containerID, err)
		}
	})

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		err := d.client.Export(
			ctx,
			pw,
			archive.WithImage(d.client.ImageService(), imageName),
			archive.WithPlatform(platforms.DefaultStrict()),
		)
		if err != nil {
			log.GetLoggerFromContextOrDefault(ctx).Errorf("failed to export container snapshot: %v", err)
		}
	}()

	return pr, clean.Release(), nil
}

func (d *discoverer) getContainerInfo(ctx context.Context, container containerd.Container) (apitypes.ContainerInfo, error) {
	id := container.ID()

	labels, err := container.Labels(ctx)
	if err != nil {
		return apitypes.ContainerInfo{}, fmt.Errorf("unable to get labels for container %s: %w", id, err)
	}
	// If this doesn't exist then use empty string as the name. Containerd
	// doesn't have the concept of a Name natively.
	name := labels[k8slabels.ContainerName]

	info, err := container.Info(ctx)
	if err != nil {
		return apitypes.ContainerInfo{}, fmt.Errorf("unable to get info for container %s: %w", id, err)
	}
	createdAt := info.CreatedAt

	image, err := container.Image(ctx)
	if err != nil {
		return apitypes.ContainerInfo{}, fmt.Errorf("unable to get image from container %s: %w", id, err)
	}

	configDescriptor, err := image.Config(ctx)
	if err != nil {
		return apitypes.ContainerInfo{}, fmt.Errorf("failed to load image config descriptor: %w", err)
	}
	imageID := configDescriptor.Digest.String()

	imageInfo, err := d.Image(ctx, imageID)
	if err != nil {
		return apitypes.ContainerInfo{}, fmt.Errorf("unable to convert image %s to container image info: %w", image.Name(), err)
	}

	return apitypes.ContainerInfo{
		ContainerID:   container.ID(),
		ContainerName: to.Ptr(name),
		CreatedAt:     to.Ptr(createdAt),
		Image:         to.Ptr(imageInfo),
		Labels:        apitypes.MapToTags(labels),
		ObjectType:    "ContainerInfo",
	}, nil
}

func (d *discoverer) Ready(ctx context.Context) (bool, error) {
	ok, err := d.client.IsServing(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get connection state: %w", err)
	}

	return ok, nil
}
