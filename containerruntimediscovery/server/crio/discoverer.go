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
	"archive/tar"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	apitypes "github.com/openclarity/openclarity/api/types"
	"github.com/openclarity/openclarity/containerruntimediscovery/types"
	"github.com/openclarity/openclarity/core/log"
	"github.com/openclarity/openclarity/core/to"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports/alltransports"
	imageTypes "github.com/containers/image/v5/types"
	"github.com/containers/storage"
	"github.com/google/uuid"
	imeta "github.com/opencontainers/image-spec/specs-go"
	ispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/opencontainers/umoci"
	"github.com/opencontainers/umoci/mutate"
	ocidir "github.com/opencontainers/umoci/oci/cas/dir"
	"github.com/opencontainers/umoci/oci/casext"
	igen "github.com/opencontainers/umoci/oci/config/generate"
	ocilayer "github.com/opencontainers/umoci/oci/layer"
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
	DefaultImageTag      = "latest"
	TmpLayerPostfix      = "-crd-server-tmp-layer"
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

	_, err := d.getContainerImageInfo(ctx, imageID)
	if err != nil {
		return nil, func() {}, fmt.Errorf("cannot find image: %w", err)
	}

	src, err := alltransports.ParseImageName("containers-storage:" + imageID)
	if err != nil {
		return nil, func() {}, fmt.Errorf("error parsing image name: %w", err)
	}

	destFilePath := filepath.Join(os.TempDir(), uuid.New().String()+"-image.tar")

	dest, err := alltransports.ParseImageName("oci-archive:" + destFilePath)
	if err != nil {
		return nil, func() {}, fmt.Errorf("error creating destination file: %w", err)
	}

	systemContext := &imageTypes.SystemContext{}
	policyContext, err := signature.NewPolicyContext(&signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}})
	if err != nil {
		return nil, func() {}, fmt.Errorf("error creating policy context: %w", err)
	}
	//nolint:errcheck
	defer func() {
		_ = policyContext.Destroy()
	}()

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
		_ = destFile.Close()
		_ = os.Remove(destFilePath)
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

// nolint:cyclop
func (d *discoverer) ExportContainer(ctx context.Context, containerID string) (io.ReadCloser, func(), error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	clean := &types.Cleanup{}
	defer clean.Clean()

	storeOptions, err := storage.DefaultStoreOptions()
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to get default store options: %w", err)
	}

	store, err := storage.GetStore(storeOptions)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to get container store: %w", err)
	}

	container, err := store.Container(containerID)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to retrieve container from store: %w", err)
	}

	// Mounting the container's RW layer.
	// We only need this layer since it will be merged with the parent layers on mount.
	layerID := container.LayerID
	layer, err := store.Layer(layerID)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to retrieve layer: %w", err)
	}

	layerPath, err := store.Mount(layer.ID, "")
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to mount layer: %w", err)
	}
	defer func() {
		_, err := store.Unmount(layer.ID, true)
		if err != nil {
			logger.Warnf("failed to unmount layer: %v", err)
		}
	}()

	// Create temp directory for layer.
	tmpLayerPath := filepath.Join(os.TempDir(), layer.ID+TmpLayerPostfix)
	err = os.MkdirAll(tmpLayerPath, 0o550) //nolint:mnd
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to create tmp dir for layer: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpLayerPath)
	}()

	// Copy layer into the tmp dir.
	// mutator.Add makes changes in the layer and does not support
	// archiving sockets that a running container may have.
	err = copyLayer(layerPath, tmpLayerPath)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to copy layer content into tmp dir: %w", err)
	}

	// Init OCI layout.
	ociDirPath := filepath.Join(os.TempDir(), uuid.New().String()+"-oci")

	err = ocidir.Create(ociDirPath)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to initialize OCI directory layout: %w", err)
	}

	engine, err := ocidir.Open(ociDirPath)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to open OCI directory: %w", err)
	}
	engineExt := casext.NewEngine(engine)
	defer func() {
		_ = engine.Close()
	}()

	// Create new image.
	g := igen.New()
	created := time.Now()

	g.SetCreated(created)
	g.SetOS(runtime.GOOS)
	g.SetArchitecture(runtime.GOARCH)
	g.ClearHistory()
	g.SetRootfsType("layers")
	g.ClearRootfsDiffIDs()

	config := g.Image()
	configDigest, configSize, err := engineExt.PutBlobJSON(ctx, config)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to add image config: %w", err)
	}

	manifest := ispec.Manifest{
		Versioned: imeta.Versioned{
			// nolint:mnd
			SchemaVersion: 2,
		},
		MediaType: ispec.MediaTypeImageManifest,
		Config: ispec.Descriptor{
			MediaType: ispec.MediaTypeImageConfig,
			Digest:    configDigest,
			Size:      configSize,
		},
		Layers: []ispec.Descriptor{},
	}

	manifestDigest, manifestSize, err := engineExt.PutBlobJSON(ctx, manifest)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to add manifest: %w", err)
	}

	descriptor := ispec.Descriptor{
		MediaType: ispec.MediaTypeImageManifest,
		Digest:    manifestDigest,
		Size:      manifestSize,
	}

	if err := engineExt.UpdateReference(ctx, DefaultImageTag, descriptor); err != nil {
		return nil, func() {}, fmt.Errorf("failed to update reference: %w", err)
	}

	descriptorPaths, err := engineExt.ResolveReference(ctx, DefaultImageTag)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to resolve reference: %w", err)
	}
	if len(descriptorPaths) == 0 {
		return nil, func() {}, errors.New("there is no image reference")
	}
	if len(descriptorPaths) != 1 {
		return nil, func() {}, errors.New("reference is ambiguous")
	}

	mutator, err := mutate.New(engine, descriptorPaths[0])
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to get mutator: %w", err)
	}

	var meta umoci.Meta
	meta.Version = umoci.MetaVersion
	meta.MapOptions.Rootless = false

	packOptions := ocilayer.RepackOptions{MapOptions: meta.MapOptions}

	// Adding the container's merged layer as a blob.
	reader := ocilayer.GenerateInsertLayer(tmpLayerPath, "/", false, &packOptions)
	defer func() {
		_ = reader.Close()
	}()

	history := &ispec.History{
		Author:     "VMClarity",
		Comment:    fmt.Sprintf("Snapshot of container %s for security scanning", container.ID),
		Created:    &created,
		CreatedBy:  "VMClarity",
		EmptyLayer: false,
	}

	_, err = mutator.Add(ctx, ispec.MediaTypeImageLayer, reader, history, mutate.GzipCompressor)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to insert layer: %w", err)
	}

	newDescriptorPath, err := mutator.Commit(ctx)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to commit: %w", err)
	}

	err = engineExt.UpdateReference(ctx, DefaultImageTag, newDescriptorPath.Root())
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to update reference: %w", err)
	}

	ociArchivePath := filepath.Join(os.TempDir(), "vmclarity-"+uuid.New().String()+".tar")
	err = tarDirectory(ociDirPath, ociArchivePath)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to create OCI archive: %w", err)
	}

	// After creating the archive, we don't need the OCI dir anymore.
	_ = os.RemoveAll(ociDirPath)

	ociArchive, err := os.Open(ociArchivePath)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to open OCI archive: %w", err)
	}

	clean.Add(func() {
		_ = ociArchive.Close()
		_ = os.Remove(ociArchivePath)
	})

	return ociArchive, clean.Release(), nil
}

// nolint:gocognit,cyclop
func copyLayer(src string, dst string) error {
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error during walking directory: %w", err)
		}

		relativePath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("cannot determine file's relative path: %w", err)
		}

		// Copied from libpod/diff.go.
		skipPaths := map[string]bool{
			"dev":  true,
			"proc": true,
			"run":  true,
			"sys":  true,
		}

		if skipPaths[relativePath] {
			if info.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		// Skip root dir
		if relativePath == "." {
			return nil
		}

		// Skip tmp layer paths to prevent recursive walk if we are scanning the cr-discovery-server's container
		if strings.HasSuffix(relativePath, TmpLayerPostfix) {
			if info.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		// Skip sockets
		if info.Mode()&os.ModeSocket != 0 {
			return nil
		}

		// Skip symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		destPath := filepath.Join(dst, relativePath)
		if info.IsDir() {
			err := os.MkdirAll(destPath, info.Mode())
			if err != nil {
				return fmt.Errorf("cannot create directory: %w", err)
			}
		} else {
			sourceFile, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("cannot open source file: %w", err)
			}
			defer func() {
				_ = sourceFile.Close()
			}()

			destFile, err := os.Create(destPath)
			if err != nil {
				return fmt.Errorf("cannot create file: %w", err)
			}
			defer func() {
				_ = destFile.Close()
			}()

			_, err = io.Copy(destFile, sourceFile)
			if err != nil {
				return fmt.Errorf("error during copying file: %w", err)
			}

			err = os.Chmod(dst, info.Mode())
			if err != nil {
				return fmt.Errorf("failed to chmod: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error during walking directory: %w", err)
	}

	return nil
}

// tarDirectory is a helper function to create the final OCI archive.
func tarDirectory(srcDir, tarFile string) error {
	file, err := os.Create(tarFile)
	if err != nil {
		return fmt.Errorf("cannot create file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	tw := tar.NewWriter(file)
	defer func() {
		_ = tw.Close()
	}()

	err = filepath.Walk(srcDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error during walking directory: %w", err)
		}

		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return fmt.Errorf("cannot retrieve file info header: %w", err)
		}

		// Mapping file's path to its relative path.
		header.Name, err = filepath.Rel(srcDir, file)
		if err != nil {
			return fmt.Errorf("cannot determine file's relative path: %w", err)
		}

		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("cannot write tar header: %w", err)
		}

		if !fi.IsDir() {
			fileContent, err := os.Open(file)
			if err != nil {
				return fmt.Errorf("cannot open file: %w", err)
			}
			defer func() {
				_ = fileContent.Close()
			}()

			if _, err := io.Copy(tw, fileContent); err != nil {
				return fmt.Errorf("error during copy: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error during walking directory: %w", err)
	}

	return nil
}

func (d *discoverer) Ready(ctx context.Context) (bool, error) {
	_, err := d.runtimeService.Status(ctx, false)
	if err != nil {
		return false, fmt.Errorf("failed to get connection state: %w", err)
	}

	return true, nil
}
