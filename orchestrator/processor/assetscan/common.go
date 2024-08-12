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

package assetscan

import (
	"context"
	"errors"
	"fmt"
	"time"

	apiclient "github.com/openclarity/vmclarity/api/client"
	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
)

func (asp *AssetScanProcessor) createOrUpdateDBFinding(ctx context.Context, info *apitypes.FindingInfo, assetScanID string, completedTime time.Time) (string, error) {
	// Create new finding
	finding := apitypes.Finding{
		FirstSeen: &completedTime,
		LastSeen:  &completedTime,
		LastSeenBy: &apitypes.AssetScanRelationship{
			Id: assetScanID,
		},
		FindingInfo: info,
	}

	fd, err := asp.client.PostFinding(ctx, finding)
	if err == nil {
		return *fd.Id, nil
	}

	var conflictError apiclient.FindingConflictError
	if !errors.As(err, &conflictError) {
		return "", fmt.Errorf("failed to create finding: %w", err)
	}

	var id string
	// Update existing finding if newer
	if conflictError.ConflictingFinding.LastSeen.Before(completedTime) {
		id = *conflictError.ConflictingFinding.Id
		finding := apitypes.Finding{
			LastSeen: &completedTime,
			LastSeenBy: &apitypes.AssetScanRelationship{
				Id: assetScanID,
			},
			FindingInfo: info,
		}

		err = asp.client.PatchFinding(ctx, id, finding)
		if err != nil {
			return id, fmt.Errorf("failed to patch finding: %w", err)
		}
	}

	return id, nil
}

func (asp *AssetScanProcessor) createOrUpdateDBAssetFinding(ctx context.Context, assetID string, findingID string, completedTime time.Time) error {
	// Create new asset finding
	assetFinding := apitypes.AssetFinding{
		Asset: &apitypes.AssetRelationship{
			Id: assetID,
		},
		Finding: &apitypes.FindingRelationship{
			Id: findingID,
		},
		FirstSeen: &completedTime,
		LastSeen:  &completedTime,
	}

	_, err := asp.client.PostAssetFinding(ctx, assetFinding)
	if err == nil {
		return nil
	}

	var conflictError apiclient.AssetFindingConflictError
	if !errors.As(err, &conflictError) {
		return fmt.Errorf("failed to create asset finding: %w", err)
	}

	// Update existing asset finding if newer
	if conflictError.ConflictingAssetFinding.LastSeen.Before(completedTime) {
		assetFinding := apitypes.AssetFinding{
			LastSeen: &completedTime,
		}

		err = asp.client.PatchAssetFinding(ctx, *conflictError.ConflictingAssetFinding.Id, assetFinding)
		if err != nil {
			return fmt.Errorf("failed to patch asset finding: %w", err)
		}
	}

	return nil
}

// Invalidate any asset findings of this type where lastSeen is
// older than this asset scan, and has not already been invalidated by
// an asset scan older than this asset scan.
func (asp *AssetScanProcessor) invalidateOlderAssetFindingsByType(ctx context.Context, findingType string, assetID string, completedTime time.Time) error {
	assetFindingsToInvalidate, err := asp.client.GetAssetFindings(ctx, apitypes.GetAssetFindingsParams{
		Filter: to.Ptr(fmt.Sprintf(
			"finding/findingInfo/objectType eq '%s' and asset/id eq '%s' and lastSeen lt %s and (invalidatedOn gt %s or invalidatedOn eq null)",
			findingType, assetID, completedTime.Format(time.RFC3339), completedTime.Format(time.RFC3339))),
	})
	if err != nil {
		return fmt.Errorf("failed to query asset findings to invalidate: %w", err)
	}

	for _, assetFinding := range *assetFindingsToInvalidate.Items {
		assetFinding.InvalidatedOn = &completedTime

		err := asp.client.PatchAssetFinding(ctx, *assetFinding.Id, assetFinding)
		if err != nil {
			return fmt.Errorf("failed to update existing asset finding %s: %w", *assetFinding.Id, err)
		}
	}

	return nil
}

func (asp *AssetScanProcessor) getActiveFindingsByType(ctx context.Context, findingType string, assetID string) (int, error) {
	activeFindings, err := asp.client.GetAssetFindings(ctx, apitypes.GetAssetFindingsParams{
		Count: to.Ptr(true),
		Filter: to.Ptr(
			fmt.Sprintf("finding/findingInfo/objectType eq '%s' and asset/id eq '%s' and invalidatedOn eq null",
				findingType, assetID),
		),

		// select the smallest amount of data to return in items, we
		// only care about the count.
		Top:    to.Ptr(1),
		Select: to.Ptr("id"),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to list all active findings: %w", err)
	}
	return *activeFindings.Count, nil
}
