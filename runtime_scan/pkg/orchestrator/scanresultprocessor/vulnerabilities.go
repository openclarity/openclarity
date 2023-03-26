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

type vulKey struct {
	vulName        string
	packageName    string
	pacakgeVersion string
}

// nolint:cyclop
func (srp *ScanResultProcessor) reconcileResultVulnerabilitiesToFindings(ctx context.Context, scanResult models.TargetScanResult) error {
	completedTime := scanResult.Status.General.LastTransitionTime

	newerFound, newerTime, err := srp.newerExistingFindingTime(ctx, scanResult.Target.Id, "Vulnerability", *completedTime)
	if err != nil {
		return fmt.Errorf("failed to check for newer existing vulnerability findings: %v", err)
	}

	existingFilter := fmt.Sprintf("findingInfo/objectType eq 'Vulnerability' and asset/id eq '%s' and scan/id eq '%s'",
		scanResult.Target.Id, scanResult.Scan.Id)
	existingFindings, err := srp.client.GetFindings(ctx, models.GetFindingsParams{
		Filter: &existingFilter,
		Select: utils.PointerTo("id,findingInfo/vulnerabilityName,findingInfo/package/name,findingInfo/package/version"),
	})
	if err != nil {
		return fmt.Errorf("failed to check for existing finding: %w", err)
	}

	existingMap := map[vulKey]string{}
	for _, finding := range *existingFindings.Items {
		vuln, err := (*finding.FindingInfo).AsVulnerabilityFindingInfo()
		if err != nil {
			return fmt.Errorf("unable to get vulnerability finding info: %w", err)
		}

		key := vulKey{*vuln.VulnerabilityName, *vuln.Package.Name, *vuln.Package.Version}
		if _, ok := existingMap[key]; ok {
			return fmt.Errorf("found multiple matching existing findings for vulnerability %s for package %s version %s", *vuln.VulnerabilityName, *vuln.Package.Name, *vuln.Package.Version)
		}
		existingMap[key] = *finding.Id
	}

	srp.logger.Infof("Found %d existing vulnerability findings for this scan: %v", len(existingMap), existingMap)

	// Create new findings for all the found vulnerabilties
	for _, vuln := range *scanResult.Vulnerabilities.Vulnerabilities {
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

		key := vulKey{*vuln.VulnerabilityName, *vuln.Package.Name, *vuln.Package.Version}
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
	err = srp.invalidateOlderFindingsByType(ctx, "Vulnerability", scanResult.Target.Id, *completedTime)
	if err != nil {
		return fmt.Errorf("failed to invalidate older vulnerability finding: %v", err)
	}

	// Get all findings which aren't invalidated, and then update the asset's summary
	target, err := srp.client.GetTarget(ctx, scanResult.Target.Id, models.GetTargetsTargetIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get target %s: %w", scanResult.Target.Id, err)
	}

	if target.Summary == nil {
		target.Summary = &models.ScanFindingsSummary{}
	}

	critialVuls, err := srp.getActiveVulnerabilityFindingsCount(ctx, models.CRITICAL)
	if err != nil {
		return fmt.Errorf("failed to list active critial vulnerabilities: %w", err)
	}
	highVuls, err := srp.getActiveVulnerabilityFindingsCount(ctx, models.HIGH)
	if err != nil {
		return fmt.Errorf("failed to list active high vulnerabilities: %w", err)
	}
	mediumVuls, err := srp.getActiveVulnerabilityFindingsCount(ctx, models.MEDIUM)
	if err != nil {
		return fmt.Errorf("failed to list active medium vulnerabilities: %w", err)
	}
	lowVuls, err := srp.getActiveVulnerabilityFindingsCount(ctx, models.LOW)
	if err != nil {
		return fmt.Errorf("failed to list active low vulnerabilities: %w", err)
	}
	negligibleVuls, err := srp.getActiveVulnerabilityFindingsCount(ctx, models.NEGLIGIBLE)
	if err != nil {
		return fmt.Errorf("failed to list active negligible vulnerabilities: %w", err)
	}

	target.Summary.TotalVulnerabilities = &models.VulnerabilityScanSummary{
		TotalCriticalVulnerabilities:   &critialVuls,
		TotalHighVulnerabilities:       &highVuls,
		TotalMediumVulnerabilities:     &mediumVuls,
		TotalLowVulnerabilities:        &lowVuls,
		TotalNegligibleVulnerabilities: &negligibleVuls,
	}

	err = srp.client.PatchTarget(ctx, target, scanResult.Target.Id)
	if err != nil {
		return fmt.Errorf("failed to patch target %s: %w", scanResult.Target.Id, err)
	}

	return nil
}

func (srp *ScanResultProcessor) getActiveVulnerabilityFindingsCount(ctx context.Context, severity models.VulnerabilitySeverity) (int, error) {
	filter := fmt.Sprintf("findingInfo/objectType eq 'Vulnerability' and invalidatedOn eq null and findingInfo/severity eq '%s'", string(severity))
	activeFindings, err := srp.client.GetFindings(ctx, models.GetFindingsParams{
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
