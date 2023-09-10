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

package assetscanprocessor

import (
	"context"
	"fmt"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/findingkey"
	logutils "github.com/openclarity/vmclarity/pkg/shared/log"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

func (asp *AssetScanProcessor) getExistingInfoFinderFindingsForScan(ctx context.Context, assetScan models.AssetScan) (map[findingkey.InfoFinderKey]string, error) {
	logger := logutils.GetLoggerFromContextOrDiscard(ctx)

	existingMap := map[findingkey.InfoFinderKey]string{}

	existingFilter := fmt.Sprintf("findingInfo/objectType eq 'InfoFinder' and foundBy/id eq '%s'", *assetScan.Id)
	existingFindings, err := asp.client.GetFindings(ctx, models.GetFindingsParams{
		Filter: &existingFilter,
		Select: utils.PointerTo("id,findingInfo/scannerName,findingInfo/type,findingInfo/data,findingInfo/path"),
	})
	if err != nil {
		return existingMap, fmt.Errorf("failed to query for findings: %w", err)
	}

	for _, finding := range *existingFindings.Items {
		info, err := (*finding.FindingInfo).AsInfoFinderFindingInfo()
		if err != nil {
			return existingMap, fmt.Errorf("unable to get InfoFinder finding info: %w", err)
		}

		key := findingkey.GenerateInfoFinderKey(info)
		if _, ok := existingMap[key]; ok {
			return existingMap, fmt.Errorf("found multiple matching existing findings for InfoFinder %v", key)
		}
		existingMap[key] = *finding.Id
	}

	logger.Infof("Found %d existing InfoFinder findings for this scan", len(existingMap))
	logger.Debugf("Existing InfoFinder map: %v", existingMap)

	return existingMap, nil
}

// nolint:cyclop
func (asp *AssetScanProcessor) reconcileResultInfoFindersToFindings(ctx context.Context, assetScan models.AssetScan) error {
	completedTime := assetScan.Status.General.LastTransitionTime

	newerFound, newerTime, err := asp.newerExistingFindingTime(ctx, assetScan.Asset.Id, "InfoFinder", *completedTime)
	if err != nil {
		return fmt.Errorf("failed to check for newer existing InfoFinder findings: %w", err)
	}

	// Build a map of existing findings for this scan to prevent us
	// recreating existing ones as we might be re-reconciling the same
	// asset scan because of downtime or a previous failure.
	existingMap, err := asp.getExistingInfoFinderFindingsForScan(ctx, assetScan)
	if err != nil {
		return fmt.Errorf("failed to check existing InfoFinder findings: %w", err)
	}

	if assetScan.InfoFinder != nil && assetScan.InfoFinder.Infos != nil {
		// Create new or update existing findings all the infos found by the
		// scan.
		for _, item := range *assetScan.InfoFinder.Infos {
			itemFindingInfo := models.InfoFinderFindingInfo{
				Data:        item.Data,
				Path:        item.Path,
				ScannerName: item.ScannerName,
				Type:        item.Type,
			}

			findingInfo := models.Finding_FindingInfo{}
			err = findingInfo.FromInfoFinderFindingInfo(itemFindingInfo)
			if err != nil {
				return fmt.Errorf("unable to convert InfoFinderFindingInfo into FindingInfo: %w", err)
			}

			finding := models.Finding{
				Asset: &models.AssetRelationship{
					Id: assetScan.Asset.Id,
				},
				FoundBy: &models.AssetScanRelationship{
					Id: *assetScan.Id,
				},
				FoundOn:     assetScan.Status.General.LastTransitionTime,
				FindingInfo: &findingInfo,
			}

			// Set InvalidatedOn time to the FoundOn time of the oldest
			// finding, found after this asset scan.
			if newerFound {
				finding.InvalidatedOn = &newerTime
			}

			key := findingkey.GenerateInfoFinderKey(itemFindingInfo)
			if id, ok := existingMap[key]; ok {
				err = asp.client.PatchFinding(ctx, id, finding)
				if err != nil {
					return fmt.Errorf("failed to create finding: %w", err)
				}
			} else {
				_, err = asp.client.PostFinding(ctx, finding)
				if err != nil {
					return fmt.Errorf("failed to create finding: %w", err)
				}
			}
		}
	}

	// Invalidate any findings of this type for this asset where foundOn is
	// older than this asset scan, and has not already been invalidated by
	// an asset scan older than this asset scan.
	err = asp.invalidateOlderFindingsByType(ctx, "InfoFinder", assetScan.Asset.Id, *completedTime)
	if err != nil {
		return fmt.Errorf("failed to invalidate older InfoFinder finding: %w", err)
	}

	// Get all findings which aren't invalidated, and then update the asset's summary
	asset, err := asp.client.GetAsset(ctx, assetScan.Asset.Id, models.GetAssetsAssetIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset %s: %w", assetScan.Asset.Id, err)
	}
	if asset.Summary == nil {
		asset.Summary = &models.ScanFindingsSummary{}
	}

	totalInfoFinder, err := asp.getActiveFindingsByType(ctx, "InfoFinder", assetScan.Asset.Id)
	if err != nil {
		return fmt.Errorf("failed to get active info finder findings: %w", err)
	}
	asset.Summary.TotalInfoFinder = &totalInfoFinder

	err = asp.client.PatchAsset(ctx, asset, assetScan.Asset.Id)
	if err != nil {
		return fmt.Errorf("failed to patch asset %s: %w", assetScan.Asset.Id, err)
	}

	return nil
}
