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

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/cli/pkg/mount"
	"github.com/openclarity/vmclarity/cli/pkg/presenter"
	"github.com/openclarity/vmclarity/cli/pkg/state"
	"github.com/openclarity/vmclarity/shared/pkg/families"
	"github.com/openclarity/vmclarity/shared/pkg/families/results"
)

const (
	fsTypeExt4 = "ext4"
	fsTypeXFS  = "xfs"
)

type CLI struct {
	state.Manager
	presenter.Presenter

	FamiliesConfig *families.Config
}

func (c *CLI) MountVolumes(ctx context.Context) ([]string, error) {
	var mountPoints []string

	devices, err := mount.ListBlockDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to list block devices: %v", err)
	}
	for _, device := range devices {
		// if the device is not mounted and of a supported filesystem type,
		// we assume it belongs to the attached volume, so we mount it.
		if device.MountPoint == "" && isSupportedFS(device.FilesystemType) {
			mountDir := "/mnt/snapshot" + uuid.NewV4().String()

			if err := device.Mount(mountDir); err != nil {
				return nil, fmt.Errorf("failed to mount device: %v", err)
			}
			log.Infof("device %v on %v is mounted", device.DeviceName, mountDir)
			mountPoints = append(mountPoints, mountDir)
		}
		if ctx.Err() != nil {
			return mountPoints, fmt.Errorf("failed to mount block devices: %w", ctx.Err())
		}
	}
	return mountPoints, nil
}

//nolint:cyclop
func (c *CLI) ExportResults(ctx context.Context, res *results.Results, errs families.RunErrors) []error {
	familiesSet := []struct {
		enabled  bool
		name     string
		exporter func(context.Context, *results.Results, families.RunErrors) error
	}{
		{
			c.FamiliesConfig.SBOM.Enabled,
			"sbom",
			c.ExportSbomResult,
		},
		{
			c.FamiliesConfig.Vulnerabilities.Enabled,
			"vulnerabilities",
			c.ExportVulResult,
		},
		{
			c.FamiliesConfig.Secrets.Enabled,
			"secrets",
			c.ExportSecretsResult,
		},
		{
			c.FamiliesConfig.Exploits.Enabled,
			"exploits",
			c.ExportExploitsResult,
		},
		{
			c.FamiliesConfig.Malware.Enabled,
			"malware",
			c.ExportMalwareResult,
		},
		{
			c.FamiliesConfig.Misconfiguration.Enabled,
			"misconfiguration",
			c.ExportMisconfigurationResult,
		},
		{
			c.FamiliesConfig.Rootkits.Enabled,
			"rootkits",
			c.ExportRootkitResult,
		},
	}

	result := make([]error, 0, len(familiesSet))
	for _, f := range familiesSet {
		if !f.enabled {
			continue
		}
		if err := f.exporter(ctx, res, errs); err != nil {
			err = fmt.Errorf("failed to export %s result to server: %w", f.name, err)
			log.Error(err)
			result = append(result, err)
		}
	}

	return result
}

func isSupportedFS(fs string) bool {
	switch fs {
	case fsTypeExt4, fsTypeXFS:
		return true
	}
	return false
}
