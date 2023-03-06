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

package gorm

import (
	"encoding/json"
	"fmt"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
)

func ConvertToRestTarget(target Target) (models.Target, error) {
	ret := models.Target{
		TargetInfo: &models.TargetType{},
	}

	switch target.Type {
	case targetTypeVM:
		var cloudProvider *models.CloudProvider
		if target.InstanceProvider != nil {
			cp := models.CloudProvider(*target.InstanceProvider)
			cloudProvider = &cp
		}
		if err := ret.TargetInfo.FromVMInfo(models.VMInfo{
			InstanceID:       *target.InstanceID,
			InstanceProvider: cloudProvider,
			Location:         *target.Location,
		}); err != nil {
			return ret, fmt.Errorf("FromVMInfo failed: %w", err)
		}
	case targetTypeDir, targetTypePod:
		fallthrough
	default:
		return ret, fmt.Errorf("unknown target type: %v", target.Type)
	}
	ret.Id = utils.StringPtr(target.ID.String())

	return ret, nil
}

func ConvertToRestTargets(targets []Target) (models.Targets, error) {
	ret := models.Targets{
		Items: &[]models.Target{},
	}

	for _, target := range targets {
		tr, err := ConvertToRestTarget(target)
		if err != nil {
			return ret, fmt.Errorf("failed to convert target: %w", err)
		}
		*ret.Items = append(*ret.Items, tr)
	}

	ret.Total = utils.IntPtr(len(targets))

	return ret, nil
}

// nolint:cyclop
func ConvertToRestScanResult(scanResult ScanResult) (models.TargetScanResult, error) {
	var ret models.TargetScanResult

	if scanResult.Secrets != nil {
		ret.Secrets = &models.SecretScan{}
		if err := json.Unmarshal(scanResult.Secrets, ret.Secrets); err != nil {
			return ret, fmt.Errorf("failed to unmarshal json: %w", err)
		}
	}
	if scanResult.Vulnerabilities != nil {
		ret.Vulnerabilities = &models.VulnerabilityScan{}
		if err := json.Unmarshal(scanResult.Vulnerabilities, ret.Vulnerabilities); err != nil {
			return ret, fmt.Errorf("failed to unmarshal json: %w", err)
		}
	}

	if scanResult.Exploits != nil {
		ret.Exploits = &models.ExploitScan{}
		if err := json.Unmarshal(scanResult.Exploits, ret.Exploits); err != nil {
			return ret, fmt.Errorf("failed to unmarshal json: %w", err)
		}
	}
	if scanResult.Malware != nil {
		ret.Malware = &models.MalwareScan{}
		if err := json.Unmarshal(scanResult.Malware, ret.Malware); err != nil {
			return ret, fmt.Errorf("failed to unmarshal json: %w", err)
		}
	}
	if scanResult.Misconfigurations != nil {
		ret.Misconfigurations = &models.MisconfigurationScan{}
		if err := json.Unmarshal(scanResult.Misconfigurations, ret.Misconfigurations); err != nil {
			return ret, fmt.Errorf("failed to unmarshal json: %w", err)
		}
	}
	if scanResult.Rootkits != nil {
		ret.Rootkits = &models.RootkitScan{}
		if err := json.Unmarshal(scanResult.Rootkits, ret.Rootkits); err != nil {
			return ret, fmt.Errorf("failed to unmarshal json: %w", err)
		}
	}
	if scanResult.Sboms != nil {
		ret.Sboms = &models.SbomScan{}
		if err := json.Unmarshal(scanResult.Sboms, ret.Sboms); err != nil {
			return ret, fmt.Errorf("failed to unmarshal json: %w", err)
		}
	}
	if scanResult.Status != nil {
		ret.Status = &models.TargetScanStatus{}
		if err := json.Unmarshal(scanResult.Status, ret.Status); err != nil {
			return ret, fmt.Errorf("failed to unmarshal json: %w", err)
		}
	}
	ret.Id = utils.StringPtr(scanResult.ID.String())
	ret.ScanId = scanResult.ScanID
	ret.TargetId = scanResult.TargetID

	return ret, nil
}

func ConvertToRestScanResults(scanResults []ScanResult) (models.TargetScanResults, error) {
	ret := models.TargetScanResults{
		Items: &[]models.TargetScanResult{},
	}

	for _, scanResult := range scanResults {
		sr, err := ConvertToRestScanResult(scanResult)
		if err != nil {
			return ret, fmt.Errorf("failed to convert scan result: %w", err)
		}
		*ret.Items = append(*ret.Items, sr)
	}

	ret.Total = utils.IntPtr(len(scanResults))

	return ret, nil
}

func ConvertToRestScan(scan Scan) (models.Scan, error) {
	var ret models.Scan

	if scan.ScanConfigSnapshot != nil {
		ret.ScanConfigSnapshot = &models.ScanConfigData{}
		if err := json.Unmarshal(scan.ScanConfigSnapshot, ret.ScanConfigSnapshot); err != nil {
			return ret, fmt.Errorf("failed to unmarshal json: %w", err)
		}
	}

	if scan.TargetIDs != nil {
		ret.TargetIDs = &[]string{}
		if err := json.Unmarshal(scan.TargetIDs, ret.TargetIDs); err != nil {
			return ret, fmt.Errorf("failed to unmarshal json: %w", err)
		}
	}

	ret.Id = utils.StringPtr(scan.ID.String())
	ret.StartTime = scan.ScanStartTime
	ret.EndTime = scan.ScanEndTime
	ret.ScanConfig = &models.ScanConfigRelationship{Id: *scan.ScanConfigID}

	return ret, nil
}

func ConvertToRestScans(scans []Scan) (models.Scans, error) {
	ret := models.Scans{
		Items: &[]models.Scan{},
	}

	for _, scan := range scans {
		sc, err := ConvertToRestScan(scan)
		if err != nil {
			return models.Scans{}, fmt.Errorf("failed to convert scan: %w", err)
		}
		*ret.Items = append(*ret.Items, sc)
	}

	ret.Total = utils.IntPtr(len(scans))

	return ret, nil
}
