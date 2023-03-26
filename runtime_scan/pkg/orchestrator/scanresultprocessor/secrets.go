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

type secretKey struct {
	fingerprint string
	startColumn int
	endColumn   int
}

func (srp *ScanResultProcessor) getExistingSecretFindingsForScan(ctx context.Context, scanResult models.TargetScanResult) (map[secretKey]string, error) {
	existingMap := map[secretKey]string{}

	existingFilter := fmt.Sprintf("findingInfo/objectType eq 'Secret' and asset/id eq '%s' and scan/id eq '%s'",
		scanResult.Target.Id, scanResult.Scan.Id)
	existingFindings, err := srp.client.GetFindings(ctx, models.GetFindingsParams{
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

		key := secretKey{*info.Fingerprint, *info.StartColumn, *info.EndColumn}
		if _, ok := existingMap[key]; ok {
			return existingMap, fmt.Errorf("found multiple matching existing findings for secret %v", key)
		}
		existingMap[key] = *finding.Id
	}

	srp.logger.Infof("Found %d existing secret findings for this scan: %v", len(existingMap), existingMap)

	return existingMap, nil
}

// nolint:cyclop
func (srp *ScanResultProcessor) reconcileResultSecretsToFindings(ctx context.Context, scanResult models.TargetScanResult) error {
	completedTime := scanResult.Status.General.LastTransitionTime

	newerFound, newerTime, err := srp.newerExistingFindingTime(ctx, scanResult.Target.Id, "Secret", *completedTime)
	if err != nil {
		return fmt.Errorf("failed to check for newer existing secret findings: %v", err)
	}

	// Build a map of existing findings for this scan to prevent us
	// recreating existings ones as we might be re-reconciling the same
	// scan result because of downtime or a previous failure.
	existingMap, err := srp.getExistingSecretFindingsForScan(ctx, scanResult)
	if err != nil {
		return fmt.Errorf("failed to check existing secret findings: %w", err)
	}

	// Create new or update existing findings all the secrets found by the
	// scan.
	for _, item := range *scanResult.Secrets.Secrets {
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

		key := secretKey{*item.Fingerprint, *item.StartColumn, *item.EndColumn}
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

	// Invalidate any findings of this type for this asset where foundOn is
	// older than this scan result, and has not already been invalidated by
	// a scan result older than this scan result.
	err = srp.invalidateOlderFindingsByType(ctx, "Secret", scanResult.Target.Id, *completedTime)
	if err != nil {
		return fmt.Errorf("failed to invalidate older secret finding: %v", err)
	}

	// Get all findings which aren't invalidated, and then update the asset's summary
	target, err := srp.client.GetTarget(ctx, scanResult.Target.Id, models.GetTargetsTargetIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get target %s: %w", scanResult.Target.Id, err)
	}
	if target.Summary == nil {
		target.Summary = &models.ScanFindingsSummary{}
	}

	totalSecrets, err := srp.getActiveFindingsByType(ctx, "Secret", scanResult.Target.Id)
	if err != nil {
		return fmt.Errorf("failed to list active critial vulnerabilities: %w", err)
	}
	target.Summary.TotalSecrets = &totalSecrets

	err = srp.client.PatchTarget(ctx, target, scanResult.Target.Id)
	if err != nil {
		return fmt.Errorf("failed to patch target %s: %w", scanResult.Target.Id, err)
	}

	return nil
}
