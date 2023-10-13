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
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v2"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/orchestrator/provider"
	"github.com/openclarity/vmclarity/pkg/shared/families"
	"github.com/openclarity/vmclarity/pkg/shared/log"
)

// mountPointPath defines the location in the container where assets will be mounted.
var mountPointPath = "/mnt/snapshot"

type Client struct {
	dockerClient *client.Client
	config       *Config
}

func New(_ context.Context) (*Client, error) {
	config, err := NewConfig()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration. Provider=%s: %w", models.Docker, err)
	}

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to load provider configuration. Provider=%s: %w", models.Docker, err)
	}

	return &Client{
		dockerClient: dockerClient,
		config:       config,
	}, nil
}

func (c *Client) Kind() models.CloudProvider {
	return models.Docker
}

func (c *Client) Estimate(ctx context.Context, stats models.AssetScanStats, asset *models.Asset, assetScanTemplate *models.AssetScanTemplate) (*models.Estimation, error) {
	return &models.Estimation{}, provider.FatalErrorf("Not Implemented")
}

func (c *Client) DiscoverAssets(ctx context.Context) provider.AssetDiscoverer {
	assetDiscoverer := provider.NewSimpleAssetDiscoverer()

	go func() {
		defer close(assetDiscoverer.OutputChan)

		// Get image assets
		imageAssets, err := c.getImageAssets(ctx)
		if err != nil {
			assetDiscoverer.Error = provider.FatalErrorf("failed to get images. Provider=%s: %w", models.Docker, err)
			return
		}

		// Get container assets
		containerAssets, err := c.getContainerAssets(ctx)
		if err != nil {
			assetDiscoverer.Error = provider.FatalErrorf("failed to get containers. Provider=%s: %w", models.Docker, err)
			return
		}

		// Combine assets
		assets := append(imageAssets, containerAssets...)

		for _, asset := range assets {
			select {
			case assetDiscoverer.OutputChan <- asset:
			case <-ctx.Done():
				assetDiscoverer.Error = ctx.Err()
				return
			}
		}
	}()

	return assetDiscoverer
}

func (c *Client) RunAssetScan(ctx context.Context, config *provider.ScanJobConfig) error {
	assetVolume, err := c.prepareScanAssetVolume(ctx, config)
	if err != nil {
		return provider.FatalErrorf("failed to prepare scan volume. Provider=%s: %w", models.Docker, err)
	}

	networkID, err := c.createScanNetwork(ctx)
	if err != nil {
		return provider.FatalErrorf("failed to prepare scan network. Provider=%s: %w", models.Docker, err)
	}

	containerID, err := c.createScanContainer(ctx, assetVolume, networkID, config)
	if err != nil {
		return provider.FatalErrorf("failed to create scan container. Provider=%s: %w", models.Docker, err)
	}

	err = c.dockerClient.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		return provider.FatalErrorf("failed to start scan container. Provider=%s: %w", models.Docker, err)
	}

	return nil
}

func (c *Client) RemoveAssetScan(ctx context.Context, config *provider.ScanJobConfig) error {
	containerID, err := c.getContainerIDFromName(ctx, config.AssetScanID)
	if err != nil {
		return provider.FatalErrorf("failed to get scan container id. Provider=%s: %w", models.Docker, err)
	}
	err = c.dockerClient.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		return provider.FatalErrorf("failed to remove scan container. Provider=%s: %w", models.Docker, err)
	}

	scanConfigFileName := getScanConfigFilePath(config)
	err = os.Remove(scanConfigFileName)
	if err != nil {
		return provider.FatalErrorf("failed to remove scan config file. Provider=%s: %w", models.Docker, err)
	}

	err = c.dockerClient.VolumeRemove(ctx, config.AssetScanID, true)
	if err != nil {
		return provider.FatalErrorf("failed to remove volume. Provider=%s: %w", models.Docker, err)
	}

	return nil
}

// prepareScanAssetVolume returns volume name or error.
func (c *Client) prepareScanAssetVolume(ctx context.Context, config *provider.ScanJobConfig) (string, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)
	volumeName := config.AssetScanID

	// Create volume if not found
	err := c.createScanAssetVolume(ctx, volumeName)
	if err != nil {
		return "", fmt.Errorf("failed to create scan volume : %w", err)
	}

	// Pull image for ephemeral container
	imagePullResp, err := c.dockerClient.ImagePull(ctx, c.config.HelperImage, types.ImagePullOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to pull helper image: %w", err)
	}

	// Drain response to avoid blocking
	_, _ = io.Copy(io.Discard, imagePullResp)
	_ = imagePullResp.Close()

	// Create an ephemeral container to populate volume with asset contents
	containerResp, err := c.dockerClient.ContainerCreate(
		ctx,
		&container.Config{
			Image: c.config.HelperImage,
		},
		&container.HostConfig{
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
		return "", fmt.Errorf("failed to create helper container: %w", err)
	}
	defer func() {
		err := c.dockerClient.ContainerRemove(ctx, containerResp.ID, types.ContainerRemoveOptions{Force: true})
		if err != nil {
			logger.Errorf("Failed to remove helper container=%s: %v", containerResp.ID, err)
		}
	}()

	// Export asset data to tar reader
	assetContents, exportCleanup, err := c.exportAsset(ctx, config)
	if err != nil {
		return "", fmt.Errorf("failed to export asset: %w", err)
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
	err = c.dockerClient.CopyToContainer(ctx, containerResp.ID, "/data", assetContents, types.CopyToContainerOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to copy asset to container: %w", err)
	}

	return volumeName, nil
}

func (c *Client) createScanAssetVolume(ctx context.Context, volumeName string) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	// Create volume if not found
	volumesResp, err := c.dockerClient.VolumeList(ctx, volume.ListOptions{
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
		_, err = c.dockerClient.VolumeCreate(ctx, volume.CreateOptions{
			Name: volumeName,
		})
		if err != nil {
			return fmt.Errorf("failed to create scan volume: %w", err)
		}
		return nil
	}
	return fmt.Errorf("invalid number of volumes found")
}

// createScanNetwork returns network id or error.
func (c *Client) createScanNetwork(ctx context.Context) (string, error) {
	// Do nothing if network already exists
	networkID, _ := c.getNetworkIDFromName(ctx, c.config.NetworkName)
	if networkID != "" {
		return networkID, nil
	}

	// Create network
	networkResp, err := c.dockerClient.NetworkCreate(
		ctx,
		c.config.NetworkName,
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

// createScanConfigFile returns scan config file path identical to getScanConfigFilePath or error.
func (c *Client) createScanConfigFile(config *provider.ScanJobConfig) (string, error) {
	scanConfigFilePath := getScanConfigFilePath(config)

	// Add volume mount point to family configuration
	familiesConfig := families.Config{}
	err := yaml.Unmarshal([]byte(config.ScannerCLIConfig), &familiesConfig)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal family scan configuration: %w", err)
	}

	families.SetMountPointsForFamiliesInput([]string{mountPointPath}, &familiesConfig)
	familiesConfigByte, err := yaml.Marshal(familiesConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal family scan configuration: %w", err)
	}

	// Create scan config file
	_, err = os.Stat(scanConfigFilePath)
	if errors.Is(err, os.ErrNotExist) {
		err = os.WriteFile(scanConfigFilePath, familiesConfigByte, 0o600) // nolint:gomnd,gofumpt
	}
	if err != nil {
		return "", fmt.Errorf("failed to create scan configuration file: %w", err)
	}

	return scanConfigFilePath, nil
}

// createScanContainer returns container id or error.
func (c *Client) createScanContainer(ctx context.Context, assetVolume, networkID string, config *provider.ScanJobConfig) (string, error) {
	containerName := config.AssetScanID

	// Do nothing if scan container already exists
	containerID, _ := c.getContainerIDFromName(ctx, containerName)
	if containerID != "" {
		return containerID, nil
	}

	// Create scan config file
	scanConfigFilePath, err := c.createScanConfigFile(config)
	if err != nil {
		return "", fmt.Errorf("failed to create scan config file: %w", err)
	}

	// Pull scanner image if required
	images, err := c.dockerClient.ImageList(ctx, types.ImageListOptions{
		Filters: filters.NewArgs(filters.Arg("reference", config.ScannerImage)),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get images: %w", err)
	}
	if len(images) == 0 {
		imagePullResp, err := c.dockerClient.ImagePull(ctx, config.ScannerImage, types.ImagePullOptions{})
		if err != nil {
			return "", fmt.Errorf("failed to pull scanner image: %w", err)
		}
		// Drain response to avoid blocking
		_, _ = io.Copy(io.Discard, imagePullResp)
		_ = imagePullResp.Close()
	}

	// Create scan container
	containerResp, err := c.dockerClient.ContainerCreate(
		ctx,
		&container.Config{
			Image: config.ScannerImage,
			Entrypoint: []string{
				"/app/vmclarity-cli",
				"scan",
				"--config",
				"/tmp/" + filepath.Base(scanConfigFilePath),
				"--server",
				config.VMClarityAddress,
				"--asset-scan-id",
				config.AssetScanID,
			},
		},
		&container.HostConfig{
			Binds: []string{fmt.Sprintf("%s:/tmp", path.Dir(scanConfigFilePath))},
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: assetVolume,
					Target: mountPointPath,
				},
			},
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				config.AssetScanID: {
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

	return containerResp.ID, nil
}

func (c *Client) getContainerIDFromName(ctx context.Context, containerName string) (string, error) {
	containers, err := c.dockerClient.ContainerList(ctx, types.ContainerListOptions{
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

func (c *Client) getNetworkIDFromName(ctx context.Context, networkName string) (string, error) {
	networks, err := c.dockerClient.NetworkList(ctx, types.NetworkListOptions{
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

func (c *Client) exportAsset(ctx context.Context, config *provider.ScanJobConfig) (io.ReadCloser, func(), error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	objectType, err := config.AssetInfo.ValueByDiscriminator()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get asset object type: %w", err)
	}

	switch value := objectType.(type) {
	case models.ContainerInfo:
		contents, err := c.dockerClient.ContainerExport(ctx, value.ContainerID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to export container: %w", err)
		}
		return contents, nil, nil

	case models.ContainerImageInfo:
		// Create an ephemeral container to export asset
		containerResp, err := c.dockerClient.ContainerCreate(
			ctx,
			&container.Config{Image: value.ImageID},
			nil,
			nil,
			nil,
			"",
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create helper container: %w", err)
		}

		cleanup := func() {
			err := c.dockerClient.ContainerRemove(ctx, containerResp.ID, types.ContainerRemoveOptions{Force: true})
			if err != nil {
				logger.Errorf("failed to remove helper container=%s: %v", containerResp.ID, err)
			}
		}

		contents, err := c.dockerClient.ContainerExport(ctx, containerResp.ID)
		if err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("failed to export container: %w", err)
		}
		return contents, cleanup, nil

	default:
		return nil, nil, fmt.Errorf("failed to export asset object type %T: Not implemented", value)
	}
}

func getScanConfigFilePath(config *provider.ScanJobConfig) string {
	return path.Join(os.TempDir(), config.AssetScanID+"_scanconfig.yaml")
}

func convertTags(tags map[string]string) *[]models.Tag {
	ret := make([]models.Tag, 0, len(tags))
	for key, val := range tags {
		ret = append(ret, models.Tag{
			Key:   key,
			Value: val,
		})
	}
	return &ret
}
