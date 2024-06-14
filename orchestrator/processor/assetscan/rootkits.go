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
func (asp *AssetScanProcessor) reconcileResultRootkitsToFindings(ctx context.Context, assetScan apitypes.AssetScan) error {
	if assetScan.Rootkits != nil && assetScan.Rootkits.Rootkits != nil {
		// Create new or update existing findings all the rootkits found by the
		// scan.
		for _, item := range *assetScan.Rootkits.Rootkits {
			itemFindingInfo := apitypes.RootkitFindingInfo{
				Message:     item.Message,
				RootkitName: item.RootkitName,
				RootkitType: item.RootkitType,
			}

			findingInfo := apitypes.FindingInfo{}
			err := findingInfo.FromRootkitFindingInfo(itemFindingInfo)
			if err != nil {
				return fmt.Errorf("unable to convert RootkitFindingInfo into FindingInfo: %w", err)
			}

			id, err := asp.createOrUpdateDBFinding(ctx, &findingInfo, *assetScan.Id, assetScan.Status.LastTransitionTime)
			if err != nil {
				return fmt.Errorf("failed to update finding: %w", err)
			}

			err = asp.createOrUpdateDBAssetFinding(ctx, assetScan.Asset.Id, id, assetScan.Status.LastTransitionTime)
			if err != nil {
				return fmt.Errorf("failed to update asset finding: %w", err)
			}
		}
	}

	err := asp.invalidateOlderAssetFindingsByType(ctx, "Rootkit", assetScan.Asset.Id, assetScan.Status.LastTransitionTime)
	if err != nil {
		return fmt.Errorf("failed to invalidate older rootkit finding: %w", err)
	}

	// Get all findings which aren't invalidated, and then update the asset's summary
	asset, err := asp.client.GetAsset(ctx, assetScan.Asset.Id, apitypes.GetAssetsAssetIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset %s: %w", assetScan.Asset.Id, err)
	}
	if asset.Summary == nil {
		asset.Summary = &apitypes.ScanFindingsSummary{}
	}

	totalRootkits, err := asp.getActiveFindingsByType(ctx, "Rootkit", assetScan.Asset.Id)
	if err != nil {
		return fmt.Errorf("failed to get active rootkit findings: %w", err)
	}
	asset.Summary.TotalRootkits = &totalRootkits

	err = asp.client.PatchAsset(ctx, asset, assetScan.Asset.Id)
	if err != nil {
		return fmt.Errorf("failed to patch asset %s: %w", assetScan.Asset.Id, err)
	}

	return nil
}
