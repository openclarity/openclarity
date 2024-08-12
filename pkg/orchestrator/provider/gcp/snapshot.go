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
)

var (
	SnapshotCreateEstimateProvisionTime = 2 * time.Minute
	SnapshotDeleteEstimateTime          = 2 * time.Minute
)

func snapshotNameFromJobConfig(config *provider.ScanJobConfig) string {
	return fmt.Sprintf("snapshot-%s", config.AssetScanID)
}

func (c *Client) ensureSnapshotFromAttachedDisk(ctx context.Context, config *provider.ScanJobConfig, disk *computepb.AttachedDisk) (*computepb.Snapshot, error) {
	snapshotName := snapshotNameFromJobConfig(config)

	snapshotRes, err := c.snapshotsClient.Get(ctx, &computepb.GetSnapshotRequest{
		Project:  c.gcpConfig.ProjectID,
		Snapshot: snapshotName,
	})
	if err == nil {
		if *snapshotRes.Status != ProvisioningStateReady {
			return snapshotRes, provider.RetryableErrorf(SnapshotCreateEstimateProvisionTime, "snapshot is not ready yet. status: %v", *snapshotRes.Status)
		}

		// Everything is good, the snapshot exists and is provisioned successfully
		return snapshotRes, nil
	}

	notFound, err := handleGcpRequestError(err, "getting snapshot %s", snapshotName)
	if !notFound {
		return nil, err
	}

	// Snapshot not found, Create the snapshot
	req := &computepb.InsertSnapshotRequest{
		Project: c.gcpConfig.ProjectID,
		SnapshotResource: &computepb.Snapshot{
			Name:       &snapshotName,
			SourceDisk: disk.Source,
		},
	}

	_, err = c.snapshotsClient.Insert(ctx, req)
	if err != nil {
		_, err := handleGcpRequestError(err, "create snapshot %s", snapshotName)
		return nil, err
	}

	return &computepb.Snapshot{}, provider.RetryableErrorf(SnapshotCreateEstimateProvisionTime, "snapshot creating")
}

func (c *Client) ensureSnapshotDeleted(ctx context.Context, config *provider.ScanJobConfig) error {
	snapshotName := snapshotNameFromJobConfig(config)

	return ensureDeleted(
		"snapshot",
		func() error {
			_, err := c.snapshotsClient.Get(ctx, &computepb.GetSnapshotRequest{
				Project:  c.gcpConfig.ProjectID,
				Snapshot: snapshotName,
			})
			return err // nolint: wrapcheck
		},
		func() error {
			_, err := c.snapshotsClient.Delete(ctx, &computepb.DeleteSnapshotRequest{
				Project:  c.gcpConfig.ProjectID,
				Snapshot: snapshotName,
			})
			return err // nolint: wrapcheck
		},
		SnapshotDeleteEstimateTime,
	)
}
