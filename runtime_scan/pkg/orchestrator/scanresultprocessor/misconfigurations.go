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
	"github.com/openclarity/vmclarity/shared/pkg/findingkey"
	logutils "github.com/openclarity/vmclarity/shared/pkg/log"
)

func (srp *ScanResultProcessor) getExistingMisconfigurationFindingsForScan(ctx context.Context, scanResult models.TargetScanResult) (map[findingkey.MisconfigurationKey]string, error) {
	logger := logutils.GetLoggerFromContextOrDiscard(ctx)

	existingMap := map[findingkey.MisconfigurationKey]string{}

	existingFilter := fmt.Sprintf("findingInfo/objectType eq 'Misconfiguration' and asset/id eq '%s' and scan/id eq '%s'",
		scanResult.Target.Id, scanResult.Scan.Id)
	existingFindings, err := srp.client.GetFindings(ctx, models.GetFindingsParams{
		Filter: &existingFilter,
		Select: utils.PointerTo("id,findingInfo/scannerName,findingInfo/testId,findingInfo/message"),
	})
	if err != nil {
		return existingMap, fmt.Errorf("failed to query for findings: %w", err)
	}

	for _, finding := range *existingFindings.Items {
		info, err := (*finding.FindingInfo).AsMisconfigurationFindingInfo()
		if err != nil {
			return existingMap, fmt.Errorf("unable to get misconfiguration finding info: %w", err)
		}

		key := findingkey.GenerateMisconfigurationKey(info)
		if _, ok := existingMap[key]; ok {
			return existingMap, fmt.Errorf("found multiple matching existing findings for misconfiguration %v", key)
		}
		existingMap[key] = *finding.Id
	}

	logger.Infof("Found %d existing misconfiguration findings for this scan", len(existingMap))
	logger.Debugf("Existing misconfiguration map: %v", existingMap)

	return existingMap, nil
}

// nolint:cyclop
func (srp *ScanResultProcessor) reconcileResultMisconfigurationsToFindings(ctx context.Context, scanResult models.TargetScanResult) error {
	completedTime := scanResult.Status.General.LastTransitionTime

	newerFound, newerTime, err := srp.newerExistingFindingTime(ctx, scanResult.Target.Id, "Misconfiguration", *completedTime)
	if err != nil {
		return fmt.Errorf("failed to check for newer existing misconfiguration findings: %v", err)
	}

	// Build a map of existing findings for this scan to prevent us
	// recreating existings ones as we might be re-reconciling the same
	// scan result because of downtime or a previous failure.
	existingMap, err := srp.getExistingMisconfigurationFindingsForScan(ctx, scanResult)
	if err != nil {
		return fmt.Errorf("failed to check existing misconfiguration findings: %w", err)
	}

	if scanResult.Misconfigurations != nil && scanResult.Misconfigurations.Misconfigurations != nil {
		// Create new or update existing findings all the misconfigurations found by the
		// scan.
		for _, item := range *scanResult.Misconfigurations.Misconfigurations {
			itemFindingInfo := models.MisconfigurationFindingInfo{
				Message:         item.Message,
				Remediation:     item.Remediation,
				ScannedPath:     item.ScannedPath,
				ScannerName:     item.ScannerName,
				Severity:        item.Severity,
				TestCategory:    item.TestCategory,
				TestDescription: item.TestDescription,
				TestID:          item.TestID,
			}

			findingInfo := models.Finding_FindingInfo{}
			err = findingInfo.FromMisconfigurationFindingInfo(itemFindingInfo)
			if err != nil {
				return fmt.Errorf("unable to convert MisconfigurationFindingInfo into FindingInfo: %w", err)
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

			key := findingkey.GenerateMisconfigurationKey(itemFindingInfo)
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
	err = srp.invalidateOlderFindingsByType(ctx, "Misconfiguration", scanResult.Target.Id, *completedTime)
	if err != nil {
		return fmt.Errorf("failed to invalidate older misconfiguration finding: %v", err)
	}

	// Get all findings which aren't invalidated, and then update the asset's summary
	target, err := srp.client.GetTarget(ctx, scanResult.Target.Id, models.GetTargetsTargetIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get target %s: %w", scanResult.Target.Id, err)
	}
	if target.Summary == nil {
		target.Summary = &models.ScanFindingsSummary{}
	}

	totalMisconfigurations, err := srp.getActiveFindingsByType(ctx, "Misconfiguration", scanResult.Target.Id)
	if err != nil {
		return fmt.Errorf("failed to list active critial vulnerabilities: %w", err)
	}
	target.Summary.TotalMisconfigurations = &totalMisconfigurations

	err = srp.client.PatchTarget(ctx, target, scanResult.Target.Id)
	if err != nil {
		return fmt.Errorf("failed to patch target %s: %w", scanResult.Target.Id, err)
	}

	return nil
}
