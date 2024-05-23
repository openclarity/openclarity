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

package scanner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	imagetypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"gopkg.in/yaml.v3"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/scanner/families"
)

// mountPointPath defines the location in the container where assets will be mounted.
var mountPointPath = "/mnt/snapshot"

var _ provider.Scanner = &Scanner{}

type Scanner struct {
	DockerClient *client.Client
	HelperImage  string
	NetworkName  string
}

func (s *Scanner) RunAssetScan(ctx context.Context, t *provider.ScanJobConfig) error {
	assetScanMount, err := s.prepareAssetScanMount(ctx, t)
	if err != nil {
		return provider.FatalErrorf("failed to prepare scan volume. Provider=%s: %w", apitypes.Docker, err)
	}

	networkID, err := s.createScanNetwork(ctx)
	if err != nil {
		return provider.FatalErrorf("failed to prepare scan network. Provider=%s: %w", apitypes.Docker, err)
	}

	containerID, err := s.createScanContainer(ctx, assetScanMount, networkID, t)
	if err != nil {
		return provider.FatalErrorf("failed to create scan container. Provider=%s: %w", apitypes.Docker, err)
	}

	err = s.DockerClient.ContainerStart(ctx, containerID, containertypes.StartOptions{})
	if err != nil {
		return provider.FatalErrorf("failed to start scan container. Provider=%s: %w", apitypes.Docker, err)
	}

	return nil
}

func (s *Scanner) RemoveAssetScan(ctx context.Context, t *provider.ScanJobConfig) error {
	containerID, err := s.getContainerIDFromName(ctx, t.AssetScanID)
	if err != nil {
		return provider.FatalErrorf("failed to get scan container id. Provider=%s: %w", apitypes.Docker, err)
	}
	err = s.DockerClient.ContainerRemove(ctx, containerID, containertypes.RemoveOptions{Force: true})
	if err != nil {
		return provider.FatalErrorf("failed to remove scan container. Provider=%s: %w", apitypes.Docker, err)
	}

	err = s.DockerClient.VolumeRemove(ctx, t.AssetScanID, true)
	if err != nil {
		return provider.FatalErrorf("failed to remove volume. Provider=%s: %w", apitypes.Docker, err)
	}

	return nil
}

// prepareAssetScanMount returns the mount for the asset scan.
func (s *Scanner) prepareAssetScanMount(ctx context.Context, config *provider.ScanJobConfig) (*mount.Mount, error) {
	objectType, err := config.AssetInfo.ValueByDiscriminator()
	if err != nil {
		return nil, fmt.Errorf("failed to get asset object type: %w", err)
	}

	switch value := objectType.(type) {
	case apitypes.ContainerInfo, apitypes.ContainerImageInfo:
		return s.prepareAssetScanVolume(ctx, config)

	case apitypes.DirInfo:
		return &mount.Mount{
			Type:   mount.TypeBind,
			Source: *value.Location,
			Target: mountPointPath,
		}, nil

	default:
		return nil, fmt.Errorf("failed to prepare mount for asset object type %T: Not implemented", value)
	}
}

// prepareAssetScanVolume returns the mount for the asset scan.
func (s *Scanner) prepareAssetScanVolume(ctx context.Context, config *provider.ScanJobConfig) (*mount.Mount, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)
	volumeName := config.AssetScanID

	// Create volume if not found
	err := s.createScanAssetVolume(ctx, volumeName)
	if err != nil {
		return nil, fmt.Errorf("failed to create scan volume : %w", err)
	}

	// Pull image for ephemeral container
	imagePullResp, err := s.DockerClient.ImagePull(ctx, s.HelperImage, imagetypes.PullOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to pull helper image: %w", err)
	}

	// Drain response to avoid blocking
	_, _ = io.Copy(io.Discard, imagePullResp)
	_ = imagePullResp.Close()

	// Create an ephemeral container to populate volume with asset contents
	containerResp, err := s.DockerClient.ContainerCreate(
		ctx,
		&containertypes.Config{
			Image: s.HelperImage,
		},
		&containertypes.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: volumeName,
					Target: "/data",
				},
			},
		},
		nil,
		nil,
		"",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create helper container: %w", err)
	}
	defer func() {
		err := s.DockerClient.ContainerRemove(ctx, containerResp.ID, containertypes.RemoveOptions{Force: true})
		if err != nil {
			logger.Errorf("Failed to remove helper container=%s: %v", containerResp.ID, err)
		}
	}()

	// Export asset data to tar reader
	assetContents, exportCleanup, err := s.exportAsset(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to export asset: %w", err)
	}
	defer func() {
		err := assetContents.Close()
		if err != nil {
			logger.Errorf("failed to close asset contents stream: %s", err.Error())
		}
		if exportCleanup != nil {
			exportCleanup()
		}
	}()

	// Copy asset data to ephemeral container
	err = s.DockerClient.CopyToContainer(ctx, containerResp.ID, "/data", assetContents, types.CopyToContainerOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to copy asset to container: %w", err)
	}

	return &mount.Mount{
		Type:   mount.TypeVolume,
		Source: volumeName,
		Target: mountPointPath,
	}, nil
}

func (s *Scanner) createScanAssetVolume(ctx context.Context, volumeName string) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	// Create volume if not found
	volumesResp, err := s.DockerClient.VolumeList(ctx, volume.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", volumeName)),
	})
	if err != nil {
		return fmt.Errorf("failed to get volumes: %w", err)
	}

	if len(volumesResp.Volumes) == 1 {
		logger.Infof("Scan volume=%s already exists", volumeName)
		return nil
	}
	if len(volumesResp.Volumes) == 0 {
		_, err = s.DockerClient.VolumeCreate(ctx, volume.CreateOptions{
			Name: volumeName,
		})
		if err != nil {
			return fmt.Errorf("failed to create scan volume: %w", err)
		}
		return nil
	}
	return errors.New("invalid number of volumes found")
}

// createScanNetwork returns network id or error.
func (s *Scanner) createScanNetwork(ctx context.Context) (string, error) {
	// Do nothing if network already exists
	networkID, _ := s.getNetworkIDFromName(ctx, s.NetworkName)
	if networkID != "" {
		return networkID, nil
	}

	// Create network
	networkResp, err := s.DockerClient.NetworkCreate(
		ctx,
		s.NetworkName,
		types.NetworkCreate{
			CheckDuplicate: true,
			Driver:         "bridge",
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to create scan network: %w", err)
	}

	return networkResp.ID, nil
}

// copyScanConfigToContainer copies scan configuration as a file to the scan container.
func (s *Scanner) copyScanConfigToContainer(ctx context.Context, containerID string, t *provider.ScanJobConfig) error {
	// Add volume mount point to family configuration
	familiesConfig := families.Config{}
	err := yaml.Unmarshal([]byte(t.ScannerCLIConfig), &familiesConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal family scan configuration: %w", err)
	}
	families.SetMountPointsForFamiliesInput([]string{mountPointPath}, &familiesConfig)
	familiesConfigByte, err := yaml.Marshal(familiesConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal family scan configuration: %w", err)
	}

	// Write scan config file to temp dir
	src := filepath.Join(os.TempDir(), getScanConfigFileName(t))
	err = os.WriteFile(src, familiesConfigByte, 0o400) // nolint:mnd
	if err != nil {
		return fmt.Errorf("failed write scan config file: %w", err)
	}

	// Create tar archive from scan config file
	srcInfo, err := archive.CopyInfoSourcePath(src, false)
	if err != nil {
		return fmt.Errorf("failed to get copy info: %w", err)
	}
	srcArchive, err := archive.TarResource(srcInfo)
	if err != nil {
		return fmt.Errorf("failed to create tar archive: %w", err)
	}
	defer srcArchive.Close()

	// Prepare archive for copy
	dstInfo := archive.CopyInfo{Path: filepath.Join("/", getScanConfigFileName(t))}
	dst, preparedArchive, err := archive.PrepareArchiveCopy(srcArchive, srcInfo, dstInfo)
	if err != nil {
		return fmt.Errorf("failed to prepare archive: %w", err)
	}
	defer preparedArchive.Close()

	// Copy scan config file to container
	err = s.DockerClient.CopyToContainer(ctx, containerID, dst, preparedArchive, types.CopyToContainerOptions{})
	if err != nil {
		return fmt.Errorf("failed to copy config file to container: %w", err)
	}

	return nil
}

// createScanContainer returns container id or error.
func (s *Scanner) createScanContainer(ctx context.Context, assetScanMount *mount.Mount, networkID string, t *provider.ScanJobConfig) (string, error) {
	containerName := t.AssetScanID

	// Do nothing if scan container already exists
	containerID, _ := s.getContainerIDFromName(ctx, containerName)
	if containerID != "" {
		return containerID, nil
	}

	// Pull scanner image if required
	images, err := s.DockerClient.ImageList(ctx, imagetypes.ListOptions{
		Filters: filters.NewArgs(filters.Arg("reference", t.ScannerImage)),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get images: %w", err)
	}
	if len(images) == 0 {
		imagePullResp, err := s.DockerClient.ImagePull(ctx, t.ScannerImage, imagetypes.PullOptions{})
		if err != nil {
			return "", fmt.Errorf("failed to pull scanner image: %w", err)
		}
		// Drain response to avoid blocking
		_, _ = io.Copy(io.Discard, imagePullResp)
		_ = imagePullResp.Close()
	}

	// Create scan container
	containerResp, err := s.DockerClient.ContainerCreate(
		ctx,
		&containertypes.Config{
			Image: t.ScannerImage,
			Cmd: []string{
				"scan",
				"--config",
				filepath.Join("/", getScanConfigFileName(t)),
				"--server",
				t.VMClarityAddress,
				"--asset-scan-id",
				t.AssetScanID,
			},
		},
		&containertypes.HostConfig{
			Mounts: []mount.Mount{
				*assetScanMount,
				{
					Type:   mount.TypeBind,
					Source: "/var/run/docker.sock",
					Target: "/var/run/docker.sock",
				},
			},
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				t.AssetScanID: {
					NetworkID: networkID,
				},
			},
		},
		nil,
		containerName,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create scan container: %w", err)
	}

	err = s.copyScanConfigToContainer(ctx, containerResp.ID, t)
	if err != nil {
		return "", fmt.Errorf("failed to copy scan config to container: %w", err)
	}

	return containerResp.ID, nil
}

func (s *Scanner) getContainerIDFromName(ctx context.Context, containerName string) (string, error) {
	containers, err := s.DockerClient.ContainerList(ctx, containertypes.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", containerName)),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list containers: %w", err)
	}
	if len(containers) == 0 {
		return "", fmt.Errorf("scan container not found: %w", err)
	}
	if len(containers) > 1 {
		return "", fmt.Errorf("found more than one scan container: %w", err)
	}
	return containers[0].ID, nil
}

func (s *Scanner) getNetworkIDFromName(ctx context.Context, networkName string) (string, error) {
	networks, err := s.DockerClient.NetworkList(ctx, types.NetworkListOptions{
		Filters: filters.NewArgs(filters.Arg("name", networkName)),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list networks: %w", err)
	}
	if len(networks) == 0 {
		return "", fmt.Errorf("scan network not found: %w", err)
	}
	if len(networks) > 1 {
		for _, n := range networks {
			if n.Name == networkName {
				return n.ID, nil
			}
		}
		return "", fmt.Errorf("found more than one scan network: %w", err)
	}
	return networks[0].ID, nil
}

// nolint:cyclop
func (s *Scanner) exportAsset(ctx context.Context, t *provider.ScanJobConfig) (io.ReadCloser, func(), error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	objectType, err := t.AssetInfo.ValueByDiscriminator()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get asset object type: %w", err)
	}

	switch value := objectType.(type) {
	case apitypes.ContainerInfo:
		contents, err := s.DockerClient.ContainerExport(ctx, value.ContainerID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to export container: %w", err)
		}
		return contents, nil, nil

	case apitypes.ContainerImageInfo:
		// Create an ephemeral container to export asset
		containerResp, err := s.DockerClient.ContainerCreate(
			ctx,
			&containertypes.Config{Image: value.ImageID},
			nil,
			nil,
			nil,
			"",
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create helper container: %w", err)
		}

		cleanup := func() {
			err := s.DockerClient.ContainerRemove(ctx, containerResp.ID, containertypes.RemoveOptions{Force: true})
			if err != nil {
				logger.Errorf("failed to remove helper container=%s: %v", containerResp.ID, err)
			}
		}

		contents, err := s.DockerClient.ContainerExport(ctx, containerResp.ID)
		if err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("failed to export container: %w", err)
		}
		return contents, cleanup, nil

	default:
		return nil, nil, fmt.Errorf("failed to export asset object type %T: Not implemented", value)
	}
}

func getScanConfigFileName(t *provider.ScanJobConfig) string {
	return fmt.Sprintf("scanconfig_%s.yaml", t.AssetScanID)
}
