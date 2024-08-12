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
	DiskEstimateProvisionTime = 2 * time.Minute
	DiskDeleteEstimateTime    = 2 * time.Minute
)

func volumeNameFromJobConfig(config *provider.ScanJobConfig) string {
	return fmt.Sprintf("targetvolume-%s", config.AssetScanID)
}

func (p *Provider) ensureManagedDiskFromSnapshot(ctx context.Context, config *provider.ScanJobConfig, snapshot armcompute.Snapshot) (armcompute.Disk, error) {
	volumeName := volumeNameFromJobConfig(config)

	volumeRes, err := p.disksClient.Get(ctx, p.config.ScannerResourceGroup, volumeName, nil)
	if err == nil {
		if *volumeRes.Disk.Properties.ProvisioningState != ProvisioningStateSucceeded {
			return volumeRes.Disk, provider.RetryableErrorf(DiskEstimateProvisionTime, "volume is not ready yet, provisioning state: %s", *volumeRes.Disk.Properties.ProvisioningState)
		}

		return volumeRes.Disk, nil
	}

	notFound, err := handleAzureRequestError(err, "getting volume %s", volumeName)
	if !notFound {
		return armcompute.Disk{}, err
	}

	_, err = p.disksClient.BeginCreateOrUpdate(ctx, p.config.ScannerResourceGroup, volumeName, armcompute.Disk{
		Location: to.Ptr(p.config.ScannerLocation),
		SKU: &armcompute.DiskSKU{
			Name: to.Ptr(armcompute.DiskStorageAccountTypesStandardSSDLRS),
		},
		Properties: &armcompute.DiskProperties{
			CreationData: &armcompute.CreationData{
				CreateOption:     to.Ptr(armcompute.DiskCreateOptionCopy),
				SourceResourceID: snapshot.ID,
			},
		},
	}, nil)
	if err != nil {
		_, err := handleAzureRequestError(err, "creating disk %s", volumeName)
		return armcompute.Disk{}, err
	}

	return armcompute.Disk{}, provider.RetryableErrorf(DiskEstimateProvisionTime, "disk creating")
}

func (p *Provider) ensureManagedDiskFromSnapshotInDifferentRegion(ctx context.Context, config *provider.ScanJobConfig, snapshot armcompute.Snapshot) (armcompute.Disk, error) {
	blobURL, err := p.ensureBlobFromSnapshot(ctx, config, snapshot)
	if err != nil {
		return armcompute.Disk{}, fmt.Errorf("failed to ensure blob from snapshot: %w", err)
	}

	volumeName := volumeNameFromJobConfig(config)

	volumeRes, err := p.disksClient.Get(ctx, p.config.ScannerResourceGroup, volumeName, nil)
	if err == nil {
		if *volumeRes.Disk.Properties.ProvisioningState != ProvisioningStateSucceeded {
			return volumeRes.Disk, provider.RetryableErrorf(DiskEstimateProvisionTime, "volume is not ready yet, provisioning state: %s", *volumeRes.Disk.Properties.ProvisioningState)
		}

		return volumeRes.Disk, nil
	}

	notFound, err := handleAzureRequestError(err, "getting volume %s", volumeName)
	if !notFound {
		return armcompute.Disk{}, err
	}

	_, err = p.disksClient.BeginCreateOrUpdate(ctx, p.config.ScannerResourceGroup, volumeName, armcompute.Disk{
		Location: to.Ptr(p.config.ScannerLocation),
		SKU: &armcompute.DiskSKU{
			Name: to.Ptr(armcompute.DiskStorageAccountTypesStandardSSDLRS),
		},
		Properties: &armcompute.DiskProperties{
			CreationData: &armcompute.CreationData{
				CreateOption:     to.Ptr(armcompute.DiskCreateOptionImport),
				SourceURI:        to.Ptr(blobURL),
				StorageAccountID: to.Ptr(fmt.Sprintf("subscriptions/%s/resourceGroups/%s/providers/Microsoft.Storage/storageAccounts/%s", p.config.SubscriptionID, p.config.ScannerResourceGroup, p.config.ScannerStorageAccountName)),
			},
		},
	}, nil)
	if err != nil {
		_, err := handleAzureRequestError(err, "creating disk %s", volumeName)
		return armcompute.Disk{}, err
	}
	return armcompute.Disk{}, provider.RetryableErrorf(DiskEstimateProvisionTime, "disk creating")
}

func (p *Provider) ensureTargetDiskDeleted(ctx context.Context, config *provider.ScanJobConfig) error {
	volumeName := volumeNameFromJobConfig(config)

	return ensureDeleted(
		"target disk",
		func() error {
			_, err := p.disksClient.Get(ctx, p.config.ScannerResourceGroup, volumeName, nil)
			return err // nolint: wrapcheck
		},
		func() error {
			_, err := p.disksClient.BeginDelete(ctx, p.config.ScannerResourceGroup, volumeName, nil)
			return err // nolint: wrapcheck
		},
		DiskDeleteEstimateTime,
	)
}
