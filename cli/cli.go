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

//nolint:wrapcheck
package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jaypipes/ghw"
	k8smount "k8s.io/mount-utils"

	"github.com/openclarity/vmclarity/cli/presenter"
	"github.com/openclarity/vmclarity/cli/state"
	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/scanner/families"
	"github.com/openclarity/vmclarity/scanner/families/types"
	"github.com/openclarity/vmclarity/utils/fsutils/filesystem"
	"github.com/openclarity/vmclarity/utils/fsutils/mount"
)

const (
	MountPointTemplate = "/mnt/snapshots/%s"
	MountPointDirPerm  = 0o770
)

// DefaultMountOptions is a set of filesystem independent mount options.
var DefaultMountOptions = []string{
	"noatime",    // Do not update inode access times on this filesystem (e.g. for faster access on the news spool to speed up news servers).
	"noauto",     // Can only be mounted explicitly (i.e., the -a option will not cause the filesystem to be mounted).
	"noexec",     // Do not permit direct execution of any binaries on the mounted filesystem.
	"norelatime", // Do not use the relatime feature: Update inode access times relative to modify or change time. Access time is only updated if the previous access time was earlier than the current modify or change time.
	"nosuid",     // Do not honor set-user-ID and set-group-ID bits or file capabilities when executing programs from this filesystem.
	"ro",         // Mount the filesystem read-only.
}

type CLI struct {
	state.Manager
	presenter.Presenter

	FamiliesConfig *families.Config
}

func (c *CLI) FamilyStarted(ctx context.Context, famType types.FamilyType) error {
	return c.Manager.MarkFamilyScanInProgress(ctx, famType)
}

func (c *CLI) FamilyFinished(ctx context.Context, res families.FamilyResult) error {
	return c.Presenter.ExportFamilyResult(ctx, res)
}

func (c *CLI) MountVolumes(ctx context.Context) ([]string, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	blockInfo, err := ghw.Block()
	if err != nil {
		return nil, fmt.Errorf("failed to list block devices: %w", err)
	}
	logger.Infof("Found block devices: %+v", blockInfo)

	// Mount all block devices that are not already mounted and have a supported filesystem.
	var mountPoints []string
	for _, disk := range blockInfo.Disks {
		logger.Infof("Checking disk %+v", *disk)
		for _, part := range disk.Partitions {
			logger.Infof("Checking partition %+v", *part)
			if part.MountPoint != "" {
				logger.Infof("Partition %s is already mounted at %s", part.Name, part.MountPoint)
				continue
			}

			if !isSupportedFS(part.Type) {
				logger.Infof("Skipping partition %s with unsupported filesystem %s", part.Name, part.Type)
				continue
			}

			mountPoint := fmt.Sprintf(MountPointTemplate, uuid.New().String())

			if err := os.MkdirAll(mountPoint, MountPointDirPerm); err != nil {
				return nil, fmt.Errorf("failed to create mountpoint. Device=%s MountPoint=%s: %w",
					part.Name, mountPoint, err)
			}

			devicePath := fmt.Sprintf("/dev/%s", part.Name)
			if true {
				if err := mount.Mount(ctx, devicePath, mountPoint, part.Type, DefaultMountOptions); err != nil {
					return nil, fmt.Errorf("failed to mount device. Device=%s MountPoint=%s: %w",
						part.Name, mountPoint, err)
			} else {
				if err := k8smount.New("").Mount(devicePath, mountPoint, part.Type, DefaultMountOptions); err != nil {
					return nil, fmt.Errorf("failed to mount device. Device=%s MountPoint=%s: %w",
						part.Name, mountPoint, err)
									}
			}
			logger.Infof("Device is mounted. Device=%s MountPoint=%s", part.Name, mountPoint)

			mountPoints = append(mountPoints, mountPoint)
		}
	}

	return mountPoints, nil
}

// WatchForAbort is responsible for watching for abort events triggered and invoking the provided cancel function to mark
// the ctx context cancelled.
func (c *CLI) WatchForAbort(ctx context.Context, cancel context.CancelFunc, interval time.Duration) {
	go func() {
		timer := time.NewTicker(interval)
		defer timer.Stop()

		logger := log.GetLoggerFromContextOrDiscard(ctx)

		for {
			select {
			case <-timer.C:
				aborted, err := c.IsAborted(ctx)
				if err != nil {
					logger.Errorf("Failed to retrieve asset scan state: %v", err)
				}
				if aborted {
					cancel()
					return
				}
			case <-ctx.Done():
				logger.Debugf("Stop watching for abort event as context is cancelled")
				return
			}
		}
	}()
}

func isSupportedFS(fs string) bool {
	switch strings.ToLower(fs) {
	case string(filesystem.Ext2), string(filesystem.Ext3), string(filesystem.Ext4):
		return true
	case string(filesystem.Xfs):
		return true
	case string(filesystem.Ntfs):
		return true
	default:
		return false
	}
}
