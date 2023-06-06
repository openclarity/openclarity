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

package scanwatcher

import (
	"errors"
	"fmt"
	"time"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
)

func isScanTimedOut(scan *models.Scan, timeout time.Duration) bool {
	if scan == nil || scan.StartTime == nil {
		return false
	}
	deadline := (*scan.StartTime).Add(timeout)

	return time.Now().UTC().After(deadline)
}

func newVulnerabilityScanSummary() *models.VulnerabilityScanSummary {
	return &models.VulnerabilityScanSummary{
		TotalCriticalVulnerabilities:   utils.PointerTo[int](0),
		TotalHighVulnerabilities:       utils.PointerTo[int](0),
		TotalMediumVulnerabilities:     utils.PointerTo[int](0),
		TotalLowVulnerabilities:        utils.PointerTo[int](0),
		TotalNegligibleVulnerabilities: utils.PointerTo[int](0),
	}
}

func newScanResultSummary() *models.ScanFindingsSummary {
	return &models.ScanFindingsSummary{
		TotalExploits:          utils.PointerTo[int](0),
		TotalMalware:           utils.PointerTo[int](0),
		TotalMisconfigurations: utils.PointerTo[int](0),
		TotalPackages:          utils.PointerTo[int](0),
		TotalRootkits:          utils.PointerTo[int](0),
		TotalSecrets:           utils.PointerTo[int](0),
		TotalVulnerabilities:   newVulnerabilityScanSummary(),
	}
}

func newScanResultFromScan(scan *models.Scan, targetID string) (*models.TargetScanResult, error) {
	if scan == nil {
		return nil, errors.New("failed to create ScanResult: Scan is nil")
	}

	if scan.ScanConfigSnapshot == nil || scan.ScanConfigSnapshot.ScanFamiliesConfig == nil {
		return nil, errors.New("failed to create ScanResult: ScanConfigSnapshot and/or ScanFamiliesConfig are nil")
	}
	familiesConfig := scan.ScanConfigSnapshot.ScanFamiliesConfig

	return &models.TargetScanResult{
		Summary: newScanResultSummary(),
		Scan: &models.ScanRelationship{
			Id: *scan.Id,
		},
		Target: &models.TargetRelationship{
			Id: targetID,
		},
		Status: &models.TargetScanStatus{
			Exploits: &models.TargetScanState{
				Errors: nil,
				State:  getInitStateFromFamilyConfig(familiesConfig.Exploits),
			},
			General: &models.TargetScanState{
				Errors: nil,
				State:  utils.PointerTo(models.TargetScanStateStateINIT),
			},
			Malware: &models.TargetScanState{
				Errors: nil,
				State:  getInitStateFromFamilyConfig(familiesConfig.Malware),
			},
			Misconfigurations: &models.TargetScanState{
				Errors: nil,
				State:  getInitStateFromFamilyConfig(familiesConfig.Misconfigurations),
			},
			Rootkits: &models.TargetScanState{
				Errors: nil,
				State:  getInitStateFromFamilyConfig(familiesConfig.Rootkits),
			},
			Sbom: &models.TargetScanState{
				Errors: nil,
				State:  getInitStateFromFamilyConfig(familiesConfig.Sbom),
			},
			Secrets: &models.TargetScanState{
				Errors: nil,
				State:  getInitStateFromFamilyConfig(familiesConfig.Secrets),
			},
			Vulnerabilities: &models.TargetScanState{
				Errors: nil,
				State:  getInitStateFromFamilyConfig(familiesConfig.Vulnerabilities),
			},
		},
		ResourceCleanup: utils.PointerTo(models.ResourceCleanupStatePENDING),
	}, nil
}

func getInitStateFromFamilyConfig(config models.FamilyConfigEnabler) *models.TargetScanStateState {
	if config == nil || !config.IsEnabled() {
		return utils.PointerTo(models.TargetScanStateStateNOTSCANNED)
	}

	return utils.PointerTo(models.TargetScanStateStateINIT)
}

func newScanSummary() *models.ScanSummary {
	return &models.ScanSummary{
		JobsCompleted:          utils.PointerTo(0),
		JobsLeftToRun:          utils.PointerTo(0),
		TotalExploits:          utils.PointerTo(0),
		TotalMalware:           utils.PointerTo(0),
		TotalMisconfigurations: utils.PointerTo(0),
		TotalPackages:          utils.PointerTo(0),
		TotalRootkits:          utils.PointerTo(0),
		TotalSecrets:           utils.PointerTo(0),
		TotalVulnerabilities: &models.VulnerabilityScanSummary{
			TotalCriticalVulnerabilities:   utils.PointerTo(0),
			TotalHighVulnerabilities:       utils.PointerTo(0),
			TotalLowVulnerabilities:        utils.PointerTo(0),
			TotalMediumVulnerabilities:     utils.PointerTo(0),
			TotalNegligibleVulnerabilities: utils.PointerTo(0),
		},
	}
}

func updateScanSummaryFromScanResult(scan *models.Scan, result models.TargetScanResult) error {
	if result.Summary == nil {
		return errors.New("invalid ScanResult: Summary field is nil")
	}

	if scan.Summary == nil {
		scan.Summary = newScanSummary()
	}

	state, ok := result.GetGeneralState()
	if !ok {
		return fmt.Errorf("general state must not be nil for ScanResult. ScanResultID=%s", *result.Id)
	}

	s, r := scan.Summary, result.Summary

	switch state {
	case models.TargetScanStateStateNOTSCANNED:
	case models.TargetScanStateStateINIT, models.TargetScanStateStateATTACHED:
		fallthrough
	case models.TargetScanStateStateINPROGRESS, models.TargetScanStateStateABORTED:
		s.JobsLeftToRun = utils.PointerTo(*s.JobsLeftToRun + 1)
	case models.TargetScanStateStateDONE:
		s.JobsCompleted = utils.PointerTo(*s.JobsCompleted + 1)
		s.TotalExploits = utils.PointerTo(*s.TotalExploits + *r.TotalExploits)
		s.TotalMalware = utils.PointerTo(*s.TotalMalware + *r.TotalMalware)
		s.TotalMisconfigurations = utils.PointerTo(*s.TotalMisconfigurations + *r.TotalMisconfigurations)
		s.TotalPackages = utils.PointerTo(*s.TotalPackages + *r.TotalPackages)
		s.TotalRootkits = utils.PointerTo(*s.TotalRootkits + *r.TotalRootkits)
		s.TotalSecrets = utils.PointerTo(*s.TotalSecrets + *r.TotalSecrets)
		s.TotalVulnerabilities = &models.VulnerabilityScanSummary{
			TotalCriticalVulnerabilities: utils.PointerTo(*s.TotalVulnerabilities.TotalCriticalVulnerabilities +
				*r.TotalVulnerabilities.TotalCriticalVulnerabilities),
			TotalHighVulnerabilities: utils.PointerTo(*s.TotalVulnerabilities.TotalHighVulnerabilities +
				*r.TotalVulnerabilities.TotalHighVulnerabilities),
			TotalLowVulnerabilities: utils.PointerTo(*s.TotalVulnerabilities.TotalLowVulnerabilities +
				*r.TotalVulnerabilities.TotalLowVulnerabilities),
			TotalMediumVulnerabilities: utils.PointerTo(*s.TotalVulnerabilities.TotalMediumVulnerabilities +
				*r.TotalVulnerabilities.TotalMediumVulnerabilities),
			TotalNegligibleVulnerabilities: utils.PointerTo(*s.TotalVulnerabilities.TotalCriticalVulnerabilities +
				*r.TotalVulnerabilities.TotalNegligibleVulnerabilities),
		}
	}

	return nil
}
