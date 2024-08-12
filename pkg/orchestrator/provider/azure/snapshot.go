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

package azure

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"

	"github.com/openclarity/vmclarity/pkg/orchestrator/provider"
)

var (
	SnapshotCreateEstimateProvisionTime = 2 * time.Minute
	SnapshotDeleteEstimateTime          = 2 * time.Minute
)

func snapshotNameFromJobConfig(config *provider.ScanJobConfig) string {
	return fmt.Sprintf("snapshot-%s", config.AssetScanID)
}

func (p *Provider) ensureSnapshotForVMRootVolume(ctx context.Context, config *provider.ScanJobConfig, vm armcompute.VirtualMachine) (armcompute.Snapshot, error) {
	snapshotName := snapshotNameFromJobConfig(config)

	snapshotRes, err := p.snapshotsClient.Get(ctx, p.config.ScannerResourceGroup, snapshotName, nil)
	if err == nil {
		if *snapshotRes.Properties.ProvisioningState != ProvisioningStateSucceeded {
			return snapshotRes.Snapshot, provider.RetryableErrorf(SnapshotCreateEstimateProvisionTime, "snapshot is not ready yet")
		}

		// Everything is good, the snapshot exists and is provisioned successfully
		return snapshotRes.Snapshot, nil
	}

	notFound, err := handleAzureRequestError(err, "getting snapshot %s", snapshotName)
	if !notFound {
		return armcompute.Snapshot{}, err
	}

	_, err = p.snapshotsClient.BeginCreateOrUpdate(ctx, p.config.ScannerResourceGroup, snapshotName, armcompute.Snapshot{
		Location: vm.Location,
		Properties: &armcompute.SnapshotProperties{
			CreationData: &armcompute.CreationData{
				CreateOption:     to.Ptr(armcompute.DiskCreateOptionCopy),
				SourceResourceID: vm.Properties.StorageProfile.OSDisk.ManagedDisk.ID,
			},
		},
	}, nil)
	if err != nil {
		_, err := handleAzureRequestError(err, "creating snapshot %s", snapshotName)
		return armcompute.Snapshot{}, err
	}

	return armcompute.Snapshot{}, provider.RetryableErrorf(SnapshotCreateEstimateProvisionTime, "snapshot creating")
}

func (p *Provider) ensureSnapshotDeleted(ctx context.Context, config *provider.ScanJobConfig) error {
	snapshotName := snapshotNameFromJobConfig(config)

	return ensureDeleted(
		"snapshot",
		func() error {
			_, err := p.snapshotsClient.Get(ctx, p.config.ScannerResourceGroup, snapshotName, nil)
			return err // nolint: wrapcheck
		},
		func() error {
			_, err := p.snapshotsClient.BeginDelete(ctx, p.config.ScannerResourceGroup, snapshotName, nil)
			return err // nolint: wrapcheck
		},
		SnapshotDeleteEstimateTime,
	)
}
