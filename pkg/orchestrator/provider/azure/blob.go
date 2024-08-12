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
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"

	"github.com/openclarity/vmclarity/pkg/orchestrator/provider"
)

var (
	estimatedBlobCopyTime    = 2 * time.Minute
	estimatedBlobAbortTime   = 2 * time.Minute
	estimatedBlobDeleteTime  = 2 * time.Minute
	snapshotSASAccessSeconds = 3600
)

func blobNameFromJobConfig(config *provider.ScanJobConfig) string {
	return fmt.Sprintf("%s.vhd", config.AssetScanID)
}

func (p *Provider) blobURLFromBlobName(blobName string) string {
	return fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", p.config.ScannerStorageAccountName, p.config.ScannerStorageContainerName, blobName)
}

func (p *Provider) ensureBlobFromSnapshot(ctx context.Context, config *provider.ScanJobConfig, snapshot armcompute.Snapshot) (string, error) {
	blobName := blobNameFromJobConfig(config)
	blobURL := p.blobURLFromBlobName(blobName)
	blobClient, err := blob.NewClient(blobURL, p.cred, nil)
	if err != nil {
		return blobURL, provider.FatalErrorf("failed to init blob client: %w", err)
	}

	getMetadata, err := blobClient.GetProperties(ctx, nil)
	if err == nil {
		copyStatus := *getMetadata.CopyStatus
		if copyStatus != blob.CopyStatusTypeSuccess {
			log.Print("blob is still copying, status is ", copyStatus)
			return blobURL, provider.RetryableErrorf(estimatedBlobCopyTime, "blob is still copying")
		}

		revokepoller, err := p.snapshotsClient.BeginRevokeAccess(ctx, p.config.ScannerResourceGroup, *snapshot.Name, nil)
		if err != nil {
			_, err := handleAzureRequestError(err, "revoking SAS access for snapshot %s", *snapshot.Name)
			return blobURL, err
		}
		_, err = revokepoller.PollUntilDone(ctx, nil)
		if err != nil {
			_, err := handleAzureRequestError(err, "waiting for SAS access to be revoked for snapshot %s", *snapshot.Name)
			return blobURL, err
		}

		return blobURL, nil
	}

	notFound, err := handleAzureRequestError(err, "getting blob %s", blobName)
	if !notFound {
		return blobURL, err
	}

	// NOTE(sambetts) Granting SAS access to a snapshot must be done
	// atomically with starting the CopyFromUrl Operation because
	// GrantAccess only provides the URL once, and we don't want to store
	// it.
	poller, err := p.snapshotsClient.BeginGrantAccess(ctx, p.config.ScannerResourceGroup, *snapshot.Name, armcompute.GrantAccessData{
		Access:            to.Ptr(armcompute.AccessLevelRead),
		DurationInSeconds: to.Ptr[int32](int32(snapshotSASAccessSeconds)),
	}, nil)
	if err != nil {
		_, err := handleAzureRequestError(err, "granting SAS access to snapshot %s", *snapshot.Name)
		return blobURL, err
	}

	res, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		_, err := handleAzureRequestError(err, "waiting for SAS access to snapshot %s be granted", *snapshot.Name)
		return blobURL, err
	}

	accessURL := *res.AccessURI.AccessSAS

	_, err = blobClient.StartCopyFromURL(ctx, accessURL, nil)
	if err != nil {
		_, err := handleAzureRequestError(err, "starting copy from URL operation for blob %s", blobName)
		return blobURL, err
	}

	return blobURL, provider.RetryableErrorf(estimatedBlobCopyTime, "blob copy from url started")
}

func (p *Provider) ensureBlobDeleted(ctx context.Context, config *provider.ScanJobConfig) error {
	blobName := blobNameFromJobConfig(config)
	blobURL := p.blobURLFromBlobName(blobName)
	blobClient, err := blob.NewClient(blobURL, p.cred, nil)
	if err != nil {
		return provider.FatalErrorf("failed to init blob client: %w", err)
	}

	getMetadata, err := blobClient.GetProperties(ctx, nil)
	if err != nil {
		notFound, err := handleAzureRequestError(err, "getting blob %s", blobName)
		if notFound {
			return nil
		}
		return err
	}

	copyStatus := *getMetadata.CopyStatus
	if copyStatus == blob.CopyStatusTypePending {
		_, err = blobClient.AbortCopyFromURL(ctx, *getMetadata.CopyID, nil)
		if err != nil {
			_, err := handleAzureRequestError(err, "aborting copy from url for blob %s", blobName)
			return err
		}
		return provider.RetryableErrorf(estimatedBlobAbortTime, "blob copy aborting")
	}

	_, err = blobClient.Delete(ctx, nil)
	if err != nil {
		_, err := handleAzureRequestError(err, "deleting blob %s", blobName)
		return err
	}

	return provider.RetryableErrorf(estimatedBlobDeleteTime, "blob %s delete started", blobName)
}
