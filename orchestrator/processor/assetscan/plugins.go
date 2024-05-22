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
	"fmt"

	apitypes "github.com/openclarity/vmclarity/api/types"
)

// nolint:cyclop
func (asp *AssetScanProcessor) reconcileResultPluginsToFindings(ctx context.Context, assetScan apitypes.AssetScan) error {
	completedTime := assetScan.Status.LastTransitionTime

	newerFound, newerTime, err := asp.newerExistingFindingTime(ctx, assetScan.Asset.Id, "Plugin", completedTime)
	if err != nil {
		return fmt.Errorf("failed to check for newer existing plugin findings: %w", err)
	}

	if assetScan.Plugins != nil && assetScan.Plugins.FindingInfos != nil {
		// Create new or update existing findings all the plugin findingd found by the
		// scan.
		for _, findingInfo := range *assetScan.Plugins.FindingInfos {
			finding := apitypes.Finding{
				Asset: &apitypes.AssetRelationship{
					Id: assetScan.Asset.Id,
				},
				FoundBy: &apitypes.AssetScanRelationship{
					Id: *assetScan.Id,
				},
				FoundOn:     &assetScan.Status.LastTransitionTime,
				FindingInfo: &findingInfo,
			}

			// Set InvalidatedOn time to the FoundOn time of the oldest
			// finding, found after this asset scan.
			if newerFound {
				finding.InvalidatedOn = &newerTime
			}

			// TODO(paralta) Check if the finding already exists and patch it if it does.
			_, err = asp.client.PostFinding(ctx, finding)
			if err != nil {
				return fmt.Errorf("failed to create finding: %w", err)
			}
		}
	}

	// Invalidate any findings of this type for this asset where foundOn is
	// older than this asset scan, and has not already been invalidated by
	// an asset scan older than this asset scan.
	err = asp.invalidateOlderFindingsByType(ctx, "Plugin", assetScan.Asset.Id, completedTime)
	if err != nil {
		return fmt.Errorf("failed to invalidate older plugin finding: %w", err)
	}

	// Get all findings which aren't invalidated, and then update the asset's summary
	asset, err := asp.client.GetAsset(ctx, assetScan.Asset.Id, apitypes.GetAssetsAssetIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset %s: %w", assetScan.Asset.Id, err)
	}
	if asset.Summary == nil {
		asset.Summary = &apitypes.ScanFindingsSummary{}
	}

	totalPlugins, err := asp.getActiveFindingsByType(ctx, "Plugin", assetScan.Asset.Id)
	if err != nil {
		return fmt.Errorf("failed to get active plugin findings: %w", err)
	}
	asset.Summary.TotalPlugins = &totalPlugins

	err = asp.client.PatchAsset(ctx, asset, assetScan.Asset.Id)
	if err != nil {
		return fmt.Errorf("failed to patch asset %s: %w", assetScan.Asset.Id, err)
	}

	return nil
}
