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
	"github.com/openclarity/vmclarity/core/to"
)

// nolint:cyclop,gocognit
func (asp *AssetScanProcessor) reconcileResultVulnerabilitiesToFindings(ctx context.Context, assetScan apitypes.AssetScan) error {
	if assetScan.Vulnerabilities != nil && assetScan.Vulnerabilities.Vulnerabilities != nil {
		// Create new findings for all the found vulnerabilities
		for _, vuln := range *assetScan.Vulnerabilities.Vulnerabilities {
			vulFindingInfo := apitypes.VulnerabilityFindingInfo{
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

			findingInfo := apitypes.FindingInfo{}
			err := findingInfo.FromVulnerabilityFindingInfo(vulFindingInfo)
			if err != nil {
				return fmt.Errorf("unable to convert VulnerabilityFindingInfo into FindingInfo: %w", err)
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

	err := asp.invalidateOlderAssetFindingsByType(ctx, "Vulnerability", assetScan.Asset.Id, assetScan.Status.LastTransitionTime)
	if err != nil {
		return fmt.Errorf("failed to invalidate older vulnerability finding: %w", err)
	}

	// Get all findings which aren't invalidated, and then update the asset's summary
	asset, err := asp.client.GetAsset(ctx, assetScan.Asset.Id, apitypes.GetAssetsAssetIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset %s: %w", assetScan.Asset.Id, err)
	}

	if asset.Summary == nil {
		asset.Summary = &apitypes.ScanFindingsSummary{}
	}

	critialVuls, err := asp.getActiveVulnerabilityFindingsCount(ctx, assetScan.Asset.Id, apitypes.CRITICAL)
	if err != nil {
		return fmt.Errorf("failed to list active critial vulnerabilities: %w", err)
	}
	highVuls, err := asp.getActiveVulnerabilityFindingsCount(ctx, assetScan.Asset.Id, apitypes.HIGH)
	if err != nil {
		return fmt.Errorf("failed to list active high vulnerabilities: %w", err)
	}
	mediumVuls, err := asp.getActiveVulnerabilityFindingsCount(ctx, assetScan.Asset.Id, apitypes.MEDIUM)
	if err != nil {
		return fmt.Errorf("failed to list active medium vulnerabilities: %w", err)
	}
	lowVuls, err := asp.getActiveVulnerabilityFindingsCount(ctx, assetScan.Asset.Id, apitypes.LOW)
	if err != nil {
		return fmt.Errorf("failed to list active low vulnerabilities: %w", err)
	}
	negligibleVuls, err := asp.getActiveVulnerabilityFindingsCount(ctx, assetScan.Asset.Id, apitypes.NEGLIGIBLE)
	if err != nil {
		return fmt.Errorf("failed to list active negligible vulnerabilities: %w", err)
	}

	asset.Summary.TotalVulnerabilities = &apitypes.VulnerabilitySeveritySummary{
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

func (asp *AssetScanProcessor) getActiveVulnerabilityFindingsCount(ctx context.Context, assetID string, severity apitypes.VulnerabilitySeverity) (int, error) {
	activeFindings, err := asp.client.GetAssetFindings(ctx, apitypes.GetAssetFindingsParams{
		Count: to.Ptr(true),
		Filter: to.Ptr(fmt.Sprintf(
			"finding/findingInfo/objectType eq 'Vulnerability' and asset/id eq '%s' and invalidatedOn eq null and finding/findingInfo/severity eq '%s'",
			assetID, string(severity)),
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
