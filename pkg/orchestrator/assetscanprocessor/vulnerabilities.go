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

// nolint:cyclop,gocognit
func (asp *AssetScanProcessor) reconcileResultVulnerabilitiesToFindings(ctx context.Context, assetScan models.AssetScan) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	completedTime := assetScan.Status.General.LastTransitionTime

	newerFound, newerTime, err := asp.newerExistingFindingTime(ctx, assetScan.Asset.Id, "Vulnerability", *completedTime)
	if err != nil {
		return fmt.Errorf("failed to check for newer existing vulnerability findings: %w", err)
	}

	existingFilter := fmt.Sprintf("findingInfo/objectType eq 'Vulnerability' and foundBy/id eq '%s'", *assetScan.Id)
	existingFindings, err := asp.client.GetFindings(ctx, models.GetFindingsParams{
		Filter: &existingFilter,
		Select: utils.PointerTo("id,findingInfo/vulnerabilityName,findingInfo/package/name,findingInfo/package/version"),
	})
	if err != nil {
		return fmt.Errorf("failed to check for existing finding: %w", err)
	}

	existingMap := map[findingkey.VulnerabilityKey]string{}
	for _, finding := range *existingFindings.Items {
		vuln, err := (*finding.FindingInfo).AsVulnerabilityFindingInfo()
		if err != nil {
			return fmt.Errorf("unable to get vulnerability finding info: %w", err)
		}

		key := findingkey.GenerateVulnerabilityKey(vuln)
		if _, ok := existingMap[key]; ok {
			return fmt.Errorf("found multiple matching existing findings for vulnerability %s for package %s version %s", *vuln.VulnerabilityName, *vuln.Package.Name, *vuln.Package.Version)
		}
		existingMap[key] = *finding.Id
	}

	logger.Infof("Found %d existing vulnerabilities findings for this scan", len(existingMap))
	logger.Debugf("Existing vulnerabilities map: %v", existingMap)

	if assetScan.Vulnerabilities != nil && assetScan.Vulnerabilities.Vulnerabilities != nil {
		// Create new findings for all the found vulnerabilities
		for _, vuln := range *assetScan.Vulnerabilities.Vulnerabilities {
			vulFindingInfo := models.VulnerabilityFindingInfo{
				VulnerabilityName: vuln.VulnerabilityName,
				Description:       vuln.Description,
				Severity:          vuln.Severity,
				Links:             vuln.Links,
				Distro:            vuln.Distro,
				Cvss:              vuln.Cvss,
				Package:           vuln.Package,
				Fix:               vuln.Fix,
				LayerId:           vuln.LayerId,
				Path:              vuln.Path,
			}

			findingInfo := models.Finding_FindingInfo{}
			err = findingInfo.FromVulnerabilityFindingInfo(vulFindingInfo)
			if err != nil {
				return fmt.Errorf("unable to convert VulnerabilityFindingInfo into FindingInfo: %w", err)
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

			key := findingkey.GenerateVulnerabilityKey(vulFindingInfo)
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
	err = asp.invalidateOlderFindingsByType(ctx, "Vulnerability", assetScan.Asset.Id, *completedTime)
	if err != nil {
		return fmt.Errorf("failed to invalidate older vulnerability finding: %w", err)
	}

	// Get all findings which aren't invalidated, and then update the asset's summary
	asset, err := asp.client.GetAsset(ctx, assetScan.Asset.Id, models.GetAssetsAssetIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset %s: %w", assetScan.Asset.Id, err)
	}

	if asset.Summary == nil {
		asset.Summary = &models.ScanFindingsSummary{}
	}

	critialVuls, err := asp.getActiveVulnerabilityFindingsCount(ctx, assetScan.Asset.Id, models.CRITICAL)
	if err != nil {
		return fmt.Errorf("failed to list active critial vulnerabilities: %w", err)
	}
	highVuls, err := asp.getActiveVulnerabilityFindingsCount(ctx, assetScan.Asset.Id, models.HIGH)
	if err != nil {
		return fmt.Errorf("failed to list active high vulnerabilities: %w", err)
	}
	mediumVuls, err := asp.getActiveVulnerabilityFindingsCount(ctx, assetScan.Asset.Id, models.MEDIUM)
	if err != nil {
		return fmt.Errorf("failed to list active medium vulnerabilities: %w", err)
	}
	lowVuls, err := asp.getActiveVulnerabilityFindingsCount(ctx, assetScan.Asset.Id, models.LOW)
	if err != nil {
		return fmt.Errorf("failed to list active low vulnerabilities: %w", err)
	}
	negligibleVuls, err := asp.getActiveVulnerabilityFindingsCount(ctx, assetScan.Asset.Id, models.NEGLIGIBLE)
	if err != nil {
		return fmt.Errorf("failed to list active negligible vulnerabilities: %w", err)
	}

	asset.Summary.TotalVulnerabilities = &models.VulnerabilityScanSummary{
		TotalCriticalVulnerabilities:   &critialVuls,
		TotalHighVulnerabilities:       &highVuls,
		TotalMediumVulnerabilities:     &mediumVuls,
		TotalLowVulnerabilities:        &lowVuls,
		TotalNegligibleVulnerabilities: &negligibleVuls,
	}

	err = asp.client.PatchAsset(ctx, asset, assetScan.Asset.Id)
	if err != nil {
		return fmt.Errorf("failed to patch asset %s: %w", assetScan.Asset.Id, err)
	}

	return nil
}

func (asp *AssetScanProcessor) getActiveVulnerabilityFindingsCount(ctx context.Context, assetID string, severity models.VulnerabilitySeverity) (int, error) {
	filter := fmt.Sprintf("findingInfo/objectType eq 'Vulnerability' and asset/id eq '%s' and invalidatedOn eq null and findingInfo/severity eq '%s'", assetID, string(severity))
	activeFindings, err := asp.client.GetFindings(ctx, models.GetFindingsParams{
		Count:  utils.PointerTo(true),
		Filter: &filter,

		// select the smallest amount of data to return in items, we
		// only care about the count.
		Top:    utils.PointerTo(1),
		Select: utils.PointerTo("id"),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to list all active findings: %w", err)
	}
	return *activeFindings.Count, nil
}
