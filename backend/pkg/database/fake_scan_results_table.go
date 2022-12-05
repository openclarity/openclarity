// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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

package database

import (
	"fmt"

	"github.com/openclarity/vmclarity/api/models"
)

func (fs *FakeScanResultsTable) ListScanResults(targetID models.TargetID, params models.GetTargetsTargetIDScanResultsParams,
) ([]ScanResults, error) {
	targets := *fs.targets
	if _, ok := targets[targetID]; !ok {
		return nil, fmt.Errorf("target not exists with ID: %s", targetID)
	}
	scanResults := make([]ScanResults, 0)
	results := *fs.scanResults
	for _, scanID := range targets[targetID].ScanResults {
		scanResults = append(scanResults, *results[scanID])
	}
	return scanResults, nil
}

func (fs *FakeScanResultsTable) CreateScanResults(targetID models.TargetID, scanResults *ScanResults,
) (*ScanResults, error) {
	sr := scanResults
	targets := *fs.targets
	if _, ok := targets[targetID]; !ok {
		return nil, fmt.Errorf("target not exists with ID: %s", targetID)
	}
	targets[targetID].ScanResults = append(targets[targetID].ScanResults, scanResults.ID)

	scanRes := *fs.scanResults
	scanRes[scanResults.ID] = scanResults
	fs.scanResults = &scanRes
	return sr, nil
}

func (fs *FakeScanResultsTable) GetScanResults(targetID models.TargetID, scanID models.ScanID) (*ScanResults, error) {
	targets := *fs.targets
	if _, ok := targets[targetID]; !ok {
		return nil, fmt.Errorf("target not exists with ID: %s", targetID)
	}
	if !contains(scanID, targets[targetID].ScanResults) {
		return nil, fmt.Errorf("scanID %s not exists for target with ID: %s", scanID, targetID)
	}
	results := *fs.scanResults
	return results[scanID], nil
}

func (fs *FakeScanResultsTable) GetSBOM(targetID models.TargetID, scanID models.ScanID) (*SbomScanResults, error) {
	targets := *fs.targets
	if _, ok := targets[targetID]; !ok {
		return nil, fmt.Errorf("target not exists with ID: %s", targetID)
	}
	if !contains(scanID, targets[targetID].ScanResults) {
		return nil, fmt.Errorf("scanID %s not exists for target with ID: %s", scanID, targetID)
	}
	results := *fs.scanResults
	return results[scanID].Sbom, nil
}

func (fs *FakeScanResultsTable) GetVulnerabilities(targetID models.TargetID, scanID models.ScanID) (*VulnerabilityScanResults, error) {
	targets := *fs.targets
	if _, ok := targets[targetID]; !ok {
		return nil, fmt.Errorf("target not exists with ID: %s", targetID)
	}
	if !contains(scanID, targets[targetID].ScanResults) {
		return nil, fmt.Errorf("scanID %s not exists for target with ID: %s", scanID, targetID)
	}
	results := *fs.scanResults
	return results[scanID].Vulnerability, nil
}

func (fs *FakeScanResultsTable) GetMalwares(targetID models.TargetID, scanID models.ScanID) (*MalwareScanResults, error) {
	targets := *fs.targets
	if _, ok := targets[targetID]; !ok {
		return nil, fmt.Errorf("target not exists with ID: %s", targetID)
	}
	if !contains(scanID, targets[targetID].ScanResults) {
		return nil, fmt.Errorf("scanID %s not exists for target with ID: %s", scanID, targetID)
	}
	results := *fs.scanResults
	return results[scanID].Malware, nil
}

func (fs *FakeScanResultsTable) GetRootkits(targetID models.TargetID, scanID models.ScanID) (*RootkitScanScanResults, error) {
	targets := *fs.targets
	if _, ok := targets[targetID]; !ok {
		return nil, fmt.Errorf("target not exists with ID: %s", targetID)
	}
	if !contains(scanID, targets[targetID].ScanResults) {
		return nil, fmt.Errorf("scanID %s not exists for target with ID: %s", scanID, targetID)
	}
	results := *fs.scanResults
	return results[scanID].Rootkit, nil
}

func (fs *FakeScanResultsTable) GetSecrets(targetID models.TargetID, scanID models.ScanID) (*SecretScanResults, error) {
	targets := *fs.targets
	if _, ok := targets[targetID]; !ok {
		return nil, fmt.Errorf("target not exists with ID: %s", targetID)
	}
	if !contains(scanID, targets[targetID].ScanResults) {
		return nil, fmt.Errorf("scanID %s not exists for target with ID: %s", scanID, targetID)
	}
	results := *fs.scanResults
	return results[scanID].Secret, nil
}

func (fs *FakeScanResultsTable) GetMisconfigurations(targetID models.TargetID, scanID models.ScanID) (*MisconfigurationScanResults, error) {
	targets := *fs.targets
	if _, ok := targets[targetID]; !ok {
		return nil, fmt.Errorf("target not exists with ID: %s", targetID)
	}
	if !contains(scanID, targets[targetID].ScanResults) {
		return nil, fmt.Errorf("scanID %s not exists for target with ID: %s", scanID, targetID)
	}
	results := *fs.scanResults
	return results[scanID].Misconfiguration, nil
}

func (fs *FakeScanResultsTable) GetExploits(targetID models.TargetID, scanID models.ScanID) (*ExploitScanResults, error) {
	targets := *fs.targets
	if _, ok := targets[targetID]; !ok {
		return nil, fmt.Errorf("target not exists with ID: %s", targetID)
	}
	if !contains(scanID, targets[targetID].ScanResults) {
		return nil, fmt.Errorf("scanID %s not exists for target with ID: %s", scanID, targetID)
	}
	results := *fs.scanResults
	return results[scanID].Exploit, nil
}

func (fs *FakeScanResultsTable) UpdateScanResults(
	targetID models.TargetID,
	scanID models.ScanID,
	scanResults *ScanResults,
) (*ScanResults, error) {
	targets := *fs.targets
	if _, ok := targets[targetID]; !ok {
		return nil, fmt.Errorf("target not exists with ID: %s", targetID)
	}
	if !contains(scanID, targets[targetID].ScanResults) {
		return nil, fmt.Errorf("scanID %s not exists for target with ID: %s", scanID, targetID)
	}
	results := *fs.scanResults
	results[scanID] = scanResults
	fs.scanResults = &results
	return results[scanID], nil
}

func contains(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}

	return false
}
