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
func (asp *AssetScanProcessor) reconcileResultPackagesToFindings(ctx context.Context, assetScan apitypes.AssetScan) error {
	if assetScan.Sbom != nil && assetScan.Sbom.Packages != nil {
		// Create new or update existing findings all the packages found by the
		// scan.
		for _, item := range *assetScan.Sbom.Packages {
			itemFindingInfo := apitypes.PackageFindingInfo{
				Cpes:     item.Cpes,
				Language: item.Language,
				Licenses: item.Licenses,
				Name:     item.Name,
				Purl:     item.Purl,
				Type:     item.Type,
				Version:  item.Version,
			}

			findingInfo := apitypes.FindingInfo{}
			err := findingInfo.FromPackageFindingInfo(itemFindingInfo)
			if err != nil {
				return fmt.Errorf("unable to convert PackageFindingInfo into FindingInfo: %w", err)
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

	err := asp.invalidateOlderAssetFindingsByType(ctx, "Package", assetScan.Asset.Id, assetScan.Status.LastTransitionTime)
	if err != nil {
		return fmt.Errorf("failed to invalidate older package finding: %w", err)
	}

	// Get all findings which aren't invalidated, and then update the asset's summary
	asset, err := asp.client.GetAsset(ctx, assetScan.Asset.Id, apitypes.GetAssetsAssetIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset %s: %w", assetScan.Asset.Id, err)
	}
	if asset.Summary == nil {
		asset.Summary = &apitypes.ScanFindingsSummary{}
	}

	totalPackages, err := asp.getActiveFindingsByType(ctx, "Package", assetScan.Asset.Id)
	if err != nil {
		return fmt.Errorf("failed to get active package findings: %w", err)
	}
	asset.Summary.TotalPackages = &totalPackages

	err = asp.client.PatchAsset(ctx, asset, assetScan.Asset.Id)
	if err != nil {
		return fmt.Errorf("failed to patch asset %s: %w", assetScan.Asset.Id, err)
	}

	return nil
}
