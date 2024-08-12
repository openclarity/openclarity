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

package gcp

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"

	"github.com/openclarity/vmclarity/pkg/orchestrator/provider"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

var (
	DiskEstimateProvisionTime = 2 * time.Minute
	DiskDeleteEstimateTime    = 2 * time.Minute
)

func diskNameFromJobConfig(config *provider.ScanJobConfig) string {
	return fmt.Sprintf("targetvolume-%s", config.AssetScanID)
}

func (p *Provider) ensureDiskFromSnapshot(ctx context.Context, config *provider.ScanJobConfig, snapshot *computepb.Snapshot) (*computepb.Disk, error) {
	diskName := diskNameFromJobConfig(config)

	diskRes, err := p.disksClient.Get(ctx, &computepb.GetDiskRequest{
		Disk:    diskName,
		Project: p.config.ProjectID,
		Zone:    p.config.ScannerZone,
	})
	if err == nil {
		if *diskRes.Status != ProvisioningStateReady {
			return diskRes, provider.RetryableErrorf(DiskEstimateProvisionTime, "disk is not ready yet. status: %v", *diskRes.Status)
		}

		// Everything is good, the disk exists and is provisioned successfully
		return diskRes, nil
	}

	notFound, err := handleGcpRequestError(err, "getting disk %s", diskName)
	if !notFound {
		return nil, err
	}

	// create the disk if not exists
	req := &computepb.InsertDiskRequest{
		Project: p.config.ProjectID,
		Zone:    p.config.ScannerZone,
		DiskResource: &computepb.Disk{
			Name: &diskName,
			// Use pd-balanced so that we have SSD not spinning HDD
			Type:           utils.PointerTo(fmt.Sprintf("zones/%v/diskTypes/pd-balanced", p.config.ScannerZone)),
			SourceSnapshot: utils.PointerTo(snapshot.GetSelfLink()),
			SizeGb:         snapshot.DiskSizeGb, // specify the size of the source disk (target scan)
		},
	}

	_, err = p.disksClient.Insert(ctx, req)
	if err != nil {
		_, err := handleGcpRequestError(err, "create disk")
		return nil, err
	}

	return nil, provider.RetryableErrorf(DiskEstimateProvisionTime, "disk creating")
}

func (p *Provider) ensureTargetDiskDeleted(ctx context.Context, config *provider.ScanJobConfig) error {
	diskName := diskNameFromJobConfig(config)

	return ensureDeleted(
		"disk",
		func() error {
			_, err := p.disksClient.Get(ctx, &computepb.GetDiskRequest{
				Disk:    diskName,
				Project: p.config.ProjectID,
				Zone:    p.config.ScannerZone,
			})
			return err // nolint: wrapcheck
		},
		func() error {
			_, err := p.disksClient.Delete(ctx, &computepb.DeleteDiskRequest{
				Disk:    diskName,
				Project: p.config.ProjectID,
				Zone:    p.config.ScannerZone,
			})
			return err // nolint: wrapcheck
		},
		DiskDeleteEstimateTime,
	)
}
