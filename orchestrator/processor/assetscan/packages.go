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

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
)

type packageExtractorFn func(assetScan apitypes.AssetScan) []apitypes.Package

// nolint:cyclop
func (asp *AssetScanProcessor) reconcileResultPackagesToFindings(ctx context.Context, assetScan apitypes.AssetScan, packageExtractor packageExtractorFn) error {
	if packageExtractor == nil {
		return errors.New("unable to extract packages")
	}

	// Create new or update existing findings for all extracted packages
	for _, pkg := range packageExtractor(assetScan) {
		findingInfo := apitypes.FindingInfo{}
		err := findingInfo.FromPackageFindingInfo(pkg.ToPackageFindingInfo())
		if err != nil {
			return fmt.Errorf("unable to convert PackageFindingInfo into FindingInfo: %w", err)
		}

		id, err := asp.createOrUpdateDBFinding(ctx, &findingInfo, *assetScan.Id, assetScan.Status.LastTransitionTime)
		if err != nil {
			return fmt.Errorf("failed to update finding: %w", err)
		}

		err = asp.patchPackageFindingSummary(ctx, id, pkg)
		if err != nil {
			return fmt.Errorf("failed to update finding: %w", err)
		}

		err = asp.createOrUpdateDBAssetFinding(ctx, assetScan.Asset.Id, id, assetScan.Status.LastTransitionTime)
		if err != nil {
			return fmt.Errorf("failed to update asset finding: %w", err)
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

// nolint:cyclop
func (asp *AssetScanProcessor) patchPackageFindingSummary(ctx context.Context, findingID apitypes.FindingID, pkg apitypes.Package) error {
	// Get total vulnerabilities for package
	critialVuls, err := asp.getPackageVulnerabilitySeverityCount(ctx, pkg, apitypes.CRITICAL)
	if err != nil {
		return fmt.Errorf("failed to list critial vulnerabilities: %w", err)
	}
	highVuls, err := asp.getPackageVulnerabilitySeverityCount(ctx, pkg, apitypes.HIGH)
	if err != nil {
		return fmt.Errorf("failed to list high vulnerabilities: %w", err)
	}
	mediumVuls, err := asp.getPackageVulnerabilitySeverityCount(ctx, pkg, apitypes.MEDIUM)
	if err != nil {
		return fmt.Errorf("failed to list medium vulnerabilities: %w", err)
	}
	lowVuls, err := asp.getPackageVulnerabilitySeverityCount(ctx, pkg, apitypes.LOW)
	if err != nil {
		return fmt.Errorf("failed to list low vulnerabilities: %w", err)
	}
	negligibleVuls, err := asp.getPackageVulnerabilitySeverityCount(ctx, pkg, apitypes.NEGLIGIBLE)
	if err != nil {
		return fmt.Errorf("failed to list negligible vulnerabilities: %w", err)
	}

	// Patch finding with updated summary
	if err := asp.client.PatchFinding(ctx, findingID, apitypes.Finding{
		Id: &findingID,
		Summary: &apitypes.FindingSummary{
			UpdatedAt: to.Ptr(time.Now().Format(time.RFC3339)),
			TotalVulnerabilities: &apitypes.VulnerabilitySeveritySummary{
				TotalCriticalVulnerabilities:   to.Ptr(critialVuls),
				TotalHighVulnerabilities:       to.Ptr(highVuls),
				TotalMediumVulnerabilities:     to.Ptr(mediumVuls),
				TotalLowVulnerabilities:        to.Ptr(lowVuls),
				TotalNegligibleVulnerabilities: to.Ptr(negligibleVuls),
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to patch finding summary: %w", err)
	}

	return nil
}

func (asp *AssetScanProcessor) getPackageVulnerabilitySeverityCount(ctx context.Context, pkg apitypes.Package, severity apitypes.VulnerabilitySeverity) (int, error) {
	findings, err := asp.client.GetFindings(ctx, apitypes.GetFindingsParams{
		Count: to.Ptr(true),
		Filter: to.Ptr(fmt.Sprintf(
			"findingInfo/objectType eq 'Vulnerability' and findingInfo/severity eq '%s' and findingInfo/package/name eq '%s' and findingInfo/package/version eq '%s'",
			string(severity), to.ValueOrZero(pkg.Name), to.ValueOrZero(pkg.Version)),
		),
		// select the smallest amount of data to return in items, we
		// only care about the count.
		Top:    to.Ptr(1),
		Select: to.Ptr("id"),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to list package vulnerability findings: %w", err)
	}

	return *findings.Count, nil
}

// withVulnerabilityPackageExtractor returns all package findings from
// vulnerability scan.
func withVulnerabilityPackageExtractor(assetScan apitypes.AssetScan) []apitypes.Package {
	var packages []apitypes.Package

	if assetScan.Vulnerabilities != nil && assetScan.Vulnerabilities.Vulnerabilities != nil {
		for _, vuln := range *assetScan.Vulnerabilities.Vulnerabilities {
			if vuln.Package == nil {
				continue
			}

			packages = append(packages, *vuln.Package)
		}
	}

	return packages
}

// withSbomPackageExtractor returns all package findings from SBOM scan.
func withSbomPackageExtractor(assetScan apitypes.AssetScan) []apitypes.Package {
	if assetScan.Sbom != nil && assetScan.Sbom.Packages != nil {
		return *assetScan.Sbom.Packages
	}

	return nil
}
