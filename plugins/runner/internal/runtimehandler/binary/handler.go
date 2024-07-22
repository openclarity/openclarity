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

//go:build linux

package binary

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/google/uuid"
	multierror "github.com/hashicorp/go-multierror"

	"github.com/openclarity/vmclarity/plugins/runner/internal/runtimehandler"
	"github.com/openclarity/vmclarity/plugins/runner/types"
	"github.com/openclarity/vmclarity/plugins/sdk-go/plugin"
	"github.com/openclarity/vmclarity/utils/fsutils/containerrootfs"
)

type binaryRuntimeHandler struct {
	config types.PluginConfig

	cmd        *exec.Cmd
	stdoutPipe io.ReadCloser
	stderrPipe io.ReadCloser

	pluginServerEndpoint string
	outputFilePath       string
	pluginDir            string
	inputDirMountPoint   string
	imageCleanup         func()
	ready                bool

	mu sync.Mutex
}

func New(ctx context.Context, config types.PluginConfig) (runtimehandler.PluginRuntimeHandler, error) {
	return &binaryRuntimeHandler{
		config:         config,
		outputFilePath: fmt.Sprintf("/tmp/%s.json", uuid.New().String()),
	}, nil
}

//nolint:cyclop
func (h *binaryRuntimeHandler) Start(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	image, cleanup, err := containerrootfs.GetImageWithCleanup(ctx, h.config.ImageName)
	if err != nil {
		return fmt.Errorf("unable to get image(%s): %w", h.config.ImageName, err)
	}
	h.imageCleanup = cleanup

	var binaryArtifactsPath string
	if h.config.BinaryArtifactsPath != "" {
		binaryArtifactsPath = h.config.BinaryArtifactsPath
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("unable to determine user's home directory: %w", err)
		}

		binaryArtifactsPath = filepath.Join(home, ".vmclarity/plugins")
	}

	h.pluginDir = filepath.Join(binaryArtifactsPath, h.config.Name, image.Metadata.ID)

	if _, err := os.Stat(h.pluginDir); os.IsNotExist(err) {
		err = containerrootfs.ToDirectory(ctx, image, h.pluginDir)
		if err != nil {
			return fmt.Errorf("unable to extract image(%s): %w", h.config.ImageName, err)
		}
	}

	// Mount input from host
	// /home/ubuntu/.vmclarity/plugins/kics/<id> + /input + /host-dir-to-scan
	h.inputDirMountPoint = filepath.Join(h.pluginDir, runtimehandler.RemoteScanInputDirOverride, h.config.InputDir)
	err = os.MkdirAll(h.inputDirMountPoint, 0o550) //nolint:mnd
	if err != nil {
		return fmt.Errorf("unable to create directory for mount point: %w", err)
	}

	err = syscall.Mount(h.config.InputDir, h.inputDirMountPoint, "", syscall.MS_BIND, "")
	if err != nil {
		return fmt.Errorf("unable to mount input directory (%s - %s): %w", h.config.InputDir, h.inputDirMountPoint, err)
	}

	defer func() {
		if r := recover(); r != nil {
			syscall.Unmount(h.inputDirMountPoint, 0) //nolint:errcheck
		}
	}()

	// https://lwn.net/Articles/281157/
	// "the read-only attribute can only be added with a remount operation afterward"
	err = syscall.Mount(h.config.InputDir, h.inputDirMountPoint, "", syscall.MS_BIND|syscall.MS_REMOUNT|syscall.MS_RDONLY, "")
	if err != nil {
		return fmt.Errorf("unable to remount input directory as read-only (%s - %s): %w", h.config.InputDir, h.inputDirMountPoint, err)
	}

	// Determine entrypoint or command to execute
	var args []string
	if len(image.Metadata.Config.Config.Entrypoint) > 0 {
		args = append(image.Metadata.Config.Config.Entrypoint[0:], image.Metadata.Config.Config.Cmd...)
	} else if len(image.Metadata.Config.Config.Cmd) > 0 {
		args = image.Metadata.Config.Config.Cmd[0:]
	} else {
		return errors.New("no entrypoint or command found in the config")
	}

	// Find a port
	openPortListener, err := net.Listen("tcp", ":0") //nolint:gosec
	if err != nil {
		return errors.New("unable to find port")
	}
	port := openPortListener.Addr().(*net.TCPAddr).Port //nolint:forcetypeassert

	h.pluginServerEndpoint = fmt.Sprintf("http://127.0.0.1:%d", port)

	// Set environment variables
	env := image.Metadata.Config.Config.Env
	env = append(env, fmt.Sprintf("%s=%s", plugin.EnvListenAddress, h.pluginServerEndpoint))

	// Set workdir
	workDir := image.Metadata.Config.Config.WorkingDir
	if workDir == "" {
		workDir = "/"
	}

	// Initialize command
	h.cmd = exec.CommandContext(ctx, args[0], args[1:]...) //nolint:gosec
	h.cmd.Env = env
	h.cmd.Dir = workDir

	h.cmd.SysProcAttr = &syscall.SysProcAttr{
		Chroot: h.pluginDir,
	}

	h.cmd.Stdin = &bytes.Buffer{}
	h.stdoutPipe, err = h.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("unable to open pipe for command stdout: %w", err)
	}

	h.stderrPipe, err = h.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("unable to open pipe for command stderr: %w", err)
	}

	// Start command
	openPortListener.Close()
	err = h.cmd.Start()
	if err != nil {
		return fmt.Errorf("unable to start process: %w", err)
	}

	// Waiting for command to be finished in the background
	go func() {
		h.cmd.Wait() //nolint:errcheck
	}()

	h.ready = true

	return nil
}

func (h *binaryRuntimeHandler) Ready() (bool, error) {
	if h.cmd == nil {
		return false, errors.New("plugin process is not running")
	}

	return h.ready, nil
}

func (h *binaryRuntimeHandler) GetPluginServerEndpoint(ctx context.Context) (string, error) {
	return h.pluginServerEndpoint, nil
}

func (h *binaryRuntimeHandler) GetOutputFilePath(ctx context.Context) (string, error) {
	return h.outputFilePath, nil
}

func (h *binaryRuntimeHandler) Logs(ctx context.Context) (io.ReadCloser, error) {
	if h.cmd == nil {
		return nil, errors.New("plugin process is not running")
	}

	stdoutPipeReader, stdoutPipeWriter := io.Pipe()
	stderrPipeReader, stderrPipeWriter := io.Pipe()

	go func() {
		io.Copy(stdoutPipeWriter, h.stdoutPipe) //nolint:errcheck
	}()

	go func() {
		io.Copy(stderrPipeWriter, h.stderrPipe) //nolint:errcheck
	}()

	return io.NopCloser(io.MultiReader(stdoutPipeReader, stderrPipeReader)), nil
}

func (h *binaryRuntimeHandler) Result(ctx context.Context) (io.ReadCloser, error) {
	f, err := os.Open(filepath.Join(h.pluginDir, h.outputFilePath))
	if err != nil {
		return nil, fmt.Errorf("unable to open result file: %w", err)
	}

	return f, nil
}

func (h *binaryRuntimeHandler) Remove(ctx context.Context) error {
	var removeErr error

	if h.cmd.ProcessState != nil {
		if !h.cmd.ProcessState.Exited() {
			if err := h.cmd.Process.Kill(); err != nil {
				removeErr = multierror.Append(removeErr, fmt.Errorf("failed to kill plugin process: %w", err))
			}
		}
	}

	// Unmount input directory
	if err := syscall.Unmount(h.inputDirMountPoint, 0); err != nil {
		removeErr = multierror.Append(removeErr, fmt.Errorf("failed to kill plugin process: %w", err))
	} else {
		if h.config.BinaryArtifactsClean {
			// Call the cleanup function for the image only after the input directory is unmounted, or else it will also remove
			// the root filesystem mounted under input
			h.imageCleanup()
		}
	}

	return removeErr //nolint:wrapcheck
}
