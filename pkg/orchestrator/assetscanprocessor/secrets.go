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
	"github.com/openclarity/vmclarity/pkg/shared/log"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

func (asp *AssetScanProcessor) getExistingSecretFindingsForScan(ctx context.Context, assetScan models.AssetScan) (map[findingkey.SecretKey]string, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	existingMap := map[findingkey.SecretKey]string{}

	existingFilter := fmt.Sprintf("findingInfo/objectType eq 'Secret' and foundBy/id eq '%s'", *assetScan.Id)
	existingFindings, err := asp.client.GetFindings(ctx, models.GetFindingsParams{
		Filter: &existingFilter,
		Select: utils.PointerTo("id,findingInfo/fingerprint,findingInfo/startColumn,findingInfo/endColumn"),
	})
	if err != nil {
		return existingMap, fmt.Errorf("failed to query for findings: %w", err)
	}

	for _, finding := range *existingFindings.Items {
		info, err := (*finding.FindingInfo).AsSecretFindingInfo()
		if err != nil {
			return existingMap, fmt.Errorf("unable to get secret finding info: %w", err)
		}

		key := findingkey.GenerateSecretKey(info)
		if _, ok := existingMap[key]; ok {
			return existingMap, fmt.Errorf("found multiple matching existing findings for secret %v", key)
		}
		existingMap[key] = *finding.Id
	}

	logger.Infof("Found %d existing secret findings for this scan", len(existingMap))
	logger.Debugf("Existing secret map: %v", existingMap)

	return existingMap, nil
}

// nolint:cyclop
func (asp *AssetScanProcessor) reconcileResultSecretsToFindings(ctx context.Context, assetScan models.AssetScan) error {
	completedTime := assetScan.Status.General.LastTransitionTime

	newerFound, newerTime, err := asp.newerExistingFindingTime(ctx, assetScan.Asset.Id, "Secret", *completedTime)
	if err != nil {
		return fmt.Errorf("failed to check for newer existing secret findings: %w", err)
	}

	// Build a map of existing findings for this scan to prevent us
	// recreating existings ones as we might be re-reconciling the same
	// asset scan because of downtime or a previous failure.
	existingMap, err := asp.getExistingSecretFindingsForScan(ctx, assetScan)
	if err != nil {
		return fmt.Errorf("failed to check existing secret findings: %w", err)
	}

	if assetScan.Secrets != nil && assetScan.Secrets.Secrets != nil {
		// Create new or update existing findings all the secrets found by the
		// scan.
		for _, item := range *assetScan.Secrets.Secrets {
			itemFindingInfo := models.SecretFindingInfo{
				Description: item.Description,
				EndLine:     item.EndLine,
				FilePath:    item.FilePath,
				Fingerprint: item.Fingerprint,
				StartLine:   item.StartLine,
				StartColumn: item.StartColumn,
				EndColumn:   item.EndColumn,
			}

			findingInfo := models.Finding_FindingInfo{}
			err = findingInfo.FromSecretFindingInfo(itemFindingInfo)
			if err != nil {
				return fmt.Errorf("unable to convert SecretFindingInfo into FindingInfo: %w", err)
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

			key := findingkey.GenerateSecretKey(itemFindingInfo)
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
	err = asp.invalidateOlderFindingsByType(ctx, "Secret", assetScan.Asset.Id, *completedTime)
	if err != nil {
		return fmt.Errorf("failed to invalidate older secret finding: %w", err)
	}

	// Get all findings which aren't invalidated, and then update the asset's summary
	asset, err := asp.client.GetAsset(ctx, assetScan.Asset.Id, models.GetAssetsAssetIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset %s: %w", assetScan.Asset.Id, err)
	}
	if asset.Summary == nil {
		asset.Summary = &models.ScanFindingsSummary{}
	}

	totalSecrets, err := asp.getActiveFindingsByType(ctx, "Secret", assetScan.Asset.Id)
	if err != nil {
		return fmt.Errorf("failed to get active secret findings: %w", err)
	}
	asset.Summary.TotalSecrets = &totalSecrets

	err = asp.client.PatchAsset(ctx, asset, assetScan.Asset.Id)
	if err != nil {
		return fmt.Errorf("failed to patch asset %s: %w", assetScan.Asset.Id, err)
	}

	return nil
}
