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

package scanresultprocessor

import (
	"context"
	"fmt"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
)

type packageKey struct {
	packageName    string
	pacakgeVersion string
}

func (srp *ScanResultProcessor) getExistingPackageFindingsForScan(ctx context.Context, scanResult models.TargetScanResult) (map[packageKey]string, error) {
	existingMap := map[packageKey]string{}

	existingFilter := fmt.Sprintf("findingInfo/objectType eq 'Package' and asset/id eq '%s' and scan/id eq '%s'",
		scanResult.Target.Id, scanResult.Scan.Id)
	existingFindings, err := srp.client.GetFindings(ctx, models.GetFindingsParams{
		Filter: &existingFilter,
		Select: utils.PointerTo("id,findingInfo/name,findingInfo/version"),
	})
	if err != nil {
		return existingMap, fmt.Errorf("failed to query for findings: %w", err)
	}

	for _, finding := range *existingFindings.Items {
		info, err := (*finding.FindingInfo).AsPackageFindingInfo()
		if err != nil {
			return existingMap, fmt.Errorf("unable to get package finding info: %w", err)
		}

		key := packageKey{*info.Name, *info.Version}
		if _, ok := existingMap[key]; ok {
			return existingMap, fmt.Errorf("found multiple matching existing findings for package %s version %s", *info.Name, *info.Version)
		}
		existingMap[key] = *finding.Id
	}

	srp.logger.Infof("Found %d existing package findings for this scan", len(existingMap))
	srp.logger.Debugf("Existing package map: %v", existingMap)

	return existingMap, nil
}

// nolint:cyclop
func (srp *ScanResultProcessor) reconcileResultPackagesToFindings(ctx context.Context, scanResult models.TargetScanResult) error {
	completedTime := scanResult.Status.General.LastTransitionTime

	newerFound, newerTime, err := srp.newerExistingFindingTime(ctx, scanResult.Target.Id, "Package", *completedTime)
	if err != nil {
		return fmt.Errorf("failed to check for newer existing package findings: %v", err)
	}

	// Build a map of existing findings for this scan to prevent us
	// recreating existings ones as we might be re-reconciling the same
	// scan result because of downtime or a previous failure.
	existingMap, err := srp.getExistingPackageFindingsForScan(ctx, scanResult)
	if err != nil {
		return fmt.Errorf("failed to check existing package findings: %w", err)
	}

	if scanResult.Sboms != nil && scanResult.Sboms.Packages != nil {
		// Create new or update existing findings all the packages found by the
		// scan.
		for _, item := range *scanResult.Sboms.Packages {
			itemFindingInfo := models.PackageFindingInfo{
				Cpes:     item.Cpes,
				Language: item.Language,
				Licenses: item.Licenses,
				Name:     item.Name,
				Purl:     item.Purl,
				Type:     item.Type,
				Version:  item.Version,
			}

			findingInfo := models.Finding_FindingInfo{}
			err = findingInfo.FromPackageFindingInfo(itemFindingInfo)
			if err != nil {
				return fmt.Errorf("unable to convert PackageFindingInfo into FindingInfo: %w", err)
			}

			finding := models.Finding{
				Scan:        scanResult.Scan,
				Asset:       scanResult.Target,
				FoundOn:     scanResult.Status.General.LastTransitionTime,
				FindingInfo: &findingInfo,
			}

			// Set InvalidatedOn time to the FoundOn time of the oldest
			// finding, found after this scan result.
			if newerFound {
				finding.InvalidatedOn = &newerTime
			}

			key := packageKey{*item.Name, *item.Version}
			if id, ok := existingMap[key]; ok {
				err = srp.client.PatchFinding(ctx, id, finding)
				if err != nil {
					return fmt.Errorf("failed to create finding: %w", err)
				}
			} else {
				_, err = srp.client.PostFinding(ctx, finding)
				if err != nil {
					return fmt.Errorf("failed to create finding: %w", err)
				}
			}
		}
	}

	// Invalidate any findings of this type for this asset where foundOn is
	// older than this scan result, and has not already been invalidated by
	// a scan result older than this scan result.
	err = srp.invalidateOlderFindingsByType(ctx, "Package", scanResult.Target.Id, *completedTime)
	if err != nil {
		return fmt.Errorf("failed to invalidate older package finding: %v", err)
	}

	// Get all findings which aren't invalidated, and then update the asset's summary
	target, err := srp.client.GetTarget(ctx, scanResult.Target.Id, models.GetTargetsTargetIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get target %s: %w", scanResult.Target.Id, err)
	}
	if target.Summary == nil {
		target.Summary = &models.ScanFindingsSummary{}
	}

	totalPackages, err := srp.getActiveFindingsByType(ctx, "Package", scanResult.Target.Id)
	if err != nil {
		return fmt.Errorf("failed to list active critial vulnerabilities: %w", err)
	}
	target.Summary.TotalPackages = &totalPackages

	err = srp.client.PatchTarget(ctx, target, scanResult.Target.Id)
	if err != nil {
		return fmt.Errorf("failed to patch target %s: %w", scanResult.Target.Id, err)
	}

	return nil
}
