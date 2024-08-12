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
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"

	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/provider/gcp/utils"
)

var (
	DiskEstimateProvisionTime = 2 * time.Minute
	DiskDeleteEstimateTime    = 2 * time.Minute
)

func diskNameFromJobConfig(config *provider.ScanJobConfig) string {
	return "targetvolume-" + config.AssetScanID
}

func (s *Scanner) ensureDiskFromSnapshot(ctx context.Context, config *provider.ScanJobConfig, snapshot *computepb.Snapshot) (*computepb.Disk, error) {
	diskName := diskNameFromJobConfig(config)

	diskRes, err := s.DisksClient.Get(ctx, &computepb.GetDiskRequest{
		Disk:    diskName,
		Project: s.ProjectID,
		Zone:    s.ScannerZone,
	})
	if err == nil {
		if *diskRes.Status != ProvisioningStateReady {
			return diskRes, provider.RetryableErrorf(DiskEstimateProvisionTime, "disk is not ready yet. status: %v", *diskRes.Status)
		}

		// Everything is good, the disk exists and is provisioned successfully
		return diskRes, nil
	}

	notFound, err := utils.HandleGcpRequestError(err, "getting disk %s", diskName)
	if !notFound {
		return nil, err // nolint: wrapcheck
	}

	// create the disk if not exists
	req := &computepb.InsertDiskRequest{
		Project: s.ProjectID,
		Zone:    s.ScannerZone,
		DiskResource: &computepb.Disk{
			Name: &diskName,
			// Use pd-balanced so that we have SSD not spinning HDD
			Type:           to.Ptr(fmt.Sprintf("zones/%v/diskTypes/pd-balanced", s.ScannerZone)),
			SourceSnapshot: to.Ptr(snapshot.GetSelfLink()),
			SizeGb:         snapshot.DiskSizeGb, // specify the size of the source disk (target scan)
		},
	}

	_, err = s.DisksClient.Insert(ctx, req)
	if err != nil {
		_, err := utils.HandleGcpRequestError(err, "create disk")
		return nil, err // nolint: wrapcheck
	}

	return nil, provider.RetryableErrorf(DiskEstimateProvisionTime, "disk creating")
}

func (s *Scanner) ensureTargetDiskDeleted(ctx context.Context, config *provider.ScanJobConfig) error {
	diskName := diskNameFromJobConfig(config)

	return utils.EnsureDeleted( // nolint: wrapcheck
		"disk",
		func() error {
			_, err := s.DisksClient.Get(ctx, &computepb.GetDiskRequest{
				Disk:    diskName,
				Project: s.ProjectID,
				Zone:    s.ScannerZone,
			})
			return err // nolint: wrapcheck
		},
		func() error {
			_, err := s.DisksClient.Delete(ctx, &computepb.DeleteDiskRequest{
				Disk:    diskName,
				Project: s.ProjectID,
				Zone:    s.ScannerZone,
			})
			return err // nolint: wrapcheck
		},
		DiskDeleteEstimateTime,
	)
}
