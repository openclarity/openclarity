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

package docker

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/openclarity/vmclarity/plugins/runner/internal/containermanager"
	"github.com/openclarity/vmclarity/plugins/runner/types"

	dockertypes "github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	imagetypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"github.com/openclarity/vmclarity/plugins/sdk-go/plugin"
)

const (
	defaultInternalServerPort = nat.Port("8080/tcp")
	defaultPollInterval       = 2 * time.Second
	defaultPluginNetwork      = "vmclarity-plugins-network"
)

type containerManager struct {
	client *dockerClient
	config types.PluginConfig

	// set on Start
	containerID string

	// set when the container is in running state
	runningContainer atomic.Pointer[dockertypes.ContainerJSON]
	runningErr       atomic.Pointer[error]
}

func New(ctx context.Context, config types.PluginConfig) (containermanager.PluginContainerManager, error) {
	// Load docker client
	client, err := newDockerClient()
	if err != nil {
		return nil, err
	}

	return &containerManager{
		client: client,
		config: config,
	}, nil
}

func (cm *containerManager) Start(ctx context.Context) error {
	// Pull scanner image if required
	err := cm.pullPluginImage(ctx)
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	// Get scanner container mounts
	scanDirMount, err := cm.getScanInputDirMount(ctx)
	if err != nil {
		return fmt.Errorf("failed to get mounts: %w", err)
	}

	// Create scanner container
	container, err := cm.client.ContainerCreate(
		ctx,
		&containertypes.Config{
			Image: cm.config.ImageName,
			Env: []string{
				fmt.Sprintf("%s=http://0.0.0.0:%s", plugin.EnvListenAddress, defaultInternalServerPort.Port()),
			},
			ExposedPorts: nat.PortSet{defaultInternalServerPort: struct{}{}},
		},
		&containertypes.HostConfig{
			PortBindings: map[nat.Port][]nat.PortBinding{
				defaultInternalServerPort: {
					{
						HostIP:   "127.0.0.1", // attach to local network driver
						HostPort: "",          // randomly assign port on host
					},
				},
			},
			Mounts: []mount.Mount{*scanDirMount},
		},
		nil,
		nil,
		"", // assign random name
	)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	cm.containerID = container.ID

	// Connect plugin container to plugin network
	networkID, err := cm.client.GetOrCreateBridgeNetwork(ctx, defaultPluginNetwork)
	if err != nil {
		return fmt.Errorf("failed to get network ID: %w", err)
	}

	if err := cm.client.NetworkConnect(ctx, networkID, cm.containerID, nil); err != nil {
		return fmt.Errorf("failed to connect plugin to network: %w", err)
	}

	// Connect host container to plugin network
	if err := cm.connectHostContainer(ctx, networkID); err != nil {
		return fmt.Errorf("failed to connect host to network: %w", err)
	}

	// Start container
	err = cm.client.ContainerStart(ctx, cm.containerID, containertypes.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Set container running data when ready
	// nolint:contextcheck
	go func() {
		rContainer, rErr := cm.waitContainerRunning(context.Background())
		if err != nil {
			cm.runningErr.Store(&rErr)
		}
		cm.runningContainer.Store(rContainer)
	}()

	return nil
}

func (cm *containerManager) Ready() (bool, error) {
	rErrPtr := cm.runningErr.Load()
	if rErrPtr != nil && *rErrPtr != nil {
		return false, fmt.Errorf("failed waiting for running state: %w", *rErrPtr)
	}

	// container is ready when the running data is set
	rContainerPtr := cm.runningContainer.Load()
	ready := rContainerPtr != nil

	return ready, nil
}

func (cm *containerManager) Logs(ctx context.Context) (io.ReadCloser, error) {
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, errors.New("failed to create log pipe")
	}

	go func() {
		// Get docker log stream
		out, err := cm.client.ContainerLogs(ctx, cm.containerID, containertypes.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
		})
		if err != nil {
			return
		}
		defer out.Close() //nolint:errcheck

		// Process stream by removing magic bytes and appending stream type (stdout: or stderr:)
		scanner := bufio.NewScanner(out)
		for scanner.Scan() {
			// get current log part from docker container
			rawBytes := scanner.Bytes()
			if len(rawBytes) <= 8 { //nolint:mnd
				continue
			}

			// write docker log without magic bytes with newline at the end
			_, _ = writer.Write(rawBytes[8:])
			_, _ = writer.WriteString("\n")
		}
	}()

	return reader, nil
}

func (cm *containerManager) GetPluginServerEndpoint(ctx context.Context) (string, error) {
	// Get running container data
	container := cm.runningContainer.Load()
	if container == nil {
		return "", errors.New("scanner container not in ready state")
	}

	// If the host is running in a container, use the plugin container hostname to
	// communicate from host since they are on the same docker network.
	if hostContainer, _ := cm.client.GetHostContainer(ctx); hostContainer != nil {
		return "http://" + net.JoinHostPort(container.Config.Hostname, defaultInternalServerPort.Port()), nil
	}

	// If CLI is running in host, use the randomly assigned port on host to
	// communicate with the container.
	hostPorts, ok := container.NetworkSettings.Ports[defaultInternalServerPort]
	if !ok {
		return "", errors.New("failed to get scanner container ports")
	}
	if len(hostPorts) == 0 {
		return "", errors.New("no network ports attached to scanner container")
	}

	return "http://" + net.JoinHostPort("127.0.0.1", hostPorts[0].HostPort), nil
}

func (cm *containerManager) Result(ctx context.Context) (io.ReadCloser, error) {
	// Copy result file from container
	reader, _, err := cm.client.CopyFromContainer(ctx, cm.containerID, containermanager.RemoteScanResultFileOverride)
	if err != nil {
		return nil, fmt.Errorf("failed to copy scanner result file: %w", err)
	}

	// Extract the tar file and read the content
	tr := tar.NewReader(reader)
	_, err = tr.Next()
	if err != nil {
		return nil, fmt.Errorf("failed to read tar file: %w", err)
	}

	// TODO: use stream rather than copying everything
	buf := new(bytes.Buffer)
	for {
		_, err := io.CopyN(buf, tr, 1024) //nolint:mnd
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("failed to copy file contents: %w", err)
		}
	}

	return io.NopCloser(buf), nil
}

func (cm *containerManager) Remove(ctx context.Context) error {
	if cm.containerID != "" {
		serr := cm.client.ContainerStop(ctx, cm.containerID, containertypes.StopOptions{}) // soft fail to allow removal
		rerr := cm.client.ContainerRemove(ctx, cm.containerID, containertypes.RemoveOptions{})
		if rerr != nil && !dockerclient.IsErrNotFound(rerr) {
			return fmt.Errorf("failed to remove scanner container: %w", errors.Join(serr, rerr))
		}
	}

	return nil
}

func (cm *containerManager) pullPluginImage(ctx context.Context) error {
	images, err := cm.client.ImageList(ctx, imagetypes.ListOptions{
		Filters: filters.NewArgs(filters.Arg("reference", cm.config.ImageName)),
	})
	if err != nil {
		return fmt.Errorf("failed to get images: %w", err)
	}

	if len(images) == 0 {
		resp, err := cm.client.ImagePull(ctx, cm.config.ImageName, imagetypes.PullOptions{})
		if err != nil {
			return fmt.Errorf("failed to pull image: %w", err)
		}

		// consume output
		_, _ = io.Copy(io.Discard, resp)
		_ = resp.Close()
	}

	return nil
}

func (cm *containerManager) waitContainerRunning(ctx context.Context) (*dockertypes.ContainerJSON, error) {
	ctx, cancel := context.WithTimeout(ctx, types.WaitReadyTimeout)
	defer cancel()

	ticker := time.NewTicker(defaultPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timed out waiting for container %s to become ready", cm.containerID)

		case <-ticker.C:
			// Get state data needed to check the container
			container, err := cm.client.ContainerInspect(ctx, cm.containerID)
			if err != nil {
				return nil, fmt.Errorf("failed to inspect scanner container: %w", err)
			}

			// Set running container
			if container.State.Running {
				return &container, nil
			}
		}
	}
}

func (cm *containerManager) getScanInputDirMount(ctx context.Context) (*mount.Mount, error) {
	// Create set with all parent directories of the input dir
	dir := cm.config.InputDir
	dirSet := make(map[string]struct{})
	for len(dir) > 1 {
		dirSet[dir] = struct{}{}
		dir = filepath.Dir(dir)
	}

	// If the host is running in a container, use the input dir mounted on the host container
	// to mount on the plugin container.
	// This is required to allow the plugin container to access the input dir from the host.
	// TODO: add docs about flow
	if hostContainer, _ := cm.client.GetHostContainer(ctx); hostContainer != nil {
		for _, p := range hostContainer.Mounts {
			if _, ok := dirSet[p.Destination]; !ok {
				continue
			}

			return &mount.Mount{
				Type:   p.Type,
				Source: p.Source,                                    // actual source on the host
				Target: containermanager.RemoteScanInputDirOverride, // override remote path
			}, nil
		}

		return nil, errors.New("input dir not mounted on host container or invalid path")
	}

	// Use default mount
	return &mount.Mount{
		Type:   mount.TypeBind,
		Source: cm.config.InputDir,
		Target: containermanager.RemoteScanInputDirOverride, // override remote path
	}, nil
}

// connectHostContainer connects host (container) to plugin network if in container mode
// to enable container name discovery.
func (cm *containerManager) connectHostContainer(ctx context.Context, pluginNetworkID string) error {
	hostContainer, _ := cm.client.GetHostContainer(ctx)
	if hostContainer == nil {
		return nil
	}

	var connected bool
	if hostContainer.NetworkSettings != nil {
		_, connected = hostContainer.NetworkSettings.Networks[defaultPluginNetwork]
	}
	if connected {
		return nil
	}

	err := cm.client.NetworkConnect(ctx, pluginNetworkID, hostContainer.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to connect host to plugin network: %w", err)
	}

	return nil
}
