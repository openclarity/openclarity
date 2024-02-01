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

	"github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

func newVulnerabilityScanSummary() *types.VulnerabilityScanSummary {
	return &types.VulnerabilityScanSummary{
		TotalCriticalVulnerabilities:   utils.PointerTo[int](0),
		TotalHighVulnerabilities:       utils.PointerTo[int](0),
		TotalMediumVulnerabilities:     utils.PointerTo[int](0),
		TotalLowVulnerabilities:        utils.PointerTo[int](0),
		TotalNegligibleVulnerabilities: utils.PointerTo[int](0),
	}
}

func newAssetScanSummary() *types.ScanFindingsSummary {
	return &types.ScanFindingsSummary{
		TotalExploits:          utils.PointerTo[int](0),
		TotalMalware:           utils.PointerTo[int](0),
		TotalMisconfigurations: utils.PointerTo[int](0),
		TotalPackages:          utils.PointerTo[int](0),
		TotalRootkits:          utils.PointerTo[int](0),
		TotalSecrets:           utils.PointerTo[int](0),
		TotalInfoFinder:        utils.PointerTo[int](0),
		TotalVulnerabilities:   newVulnerabilityScanSummary(),
	}
}

func newAssetScanFromScan(scan *types.Scan, assetID string) (*types.AssetScan, error) {
	if scan == nil {
		return nil, errors.New("failed to create AssetScan: Scan is nil")
	}

	if scan.AssetScanTemplate == nil || scan.AssetScanTemplate.ScanFamiliesConfig == nil {
		return nil, errors.New("failed to create AssetScan: AssetScanTemplate and/or AssetScanTemplate.ScanFamiliesConfig is nil")
	}
	familiesConfig := scan.AssetScanTemplate.ScanFamiliesConfig

	return &types.AssetScan{
		Summary: newAssetScanSummary(),
		Scan: &types.ScanRelationship{
			Id: *scan.Id,
		},
		Asset: &types.AssetRelationship{
			Id: assetID,
		},
		ScanFamiliesConfig:            familiesConfig,
		ScannerInstanceCreationConfig: scan.AssetScanTemplate.ScannerInstanceCreationConfig,
		Status: types.NewAssetScanStatus(
			types.AssetScanStatusStatePending,
			types.AssetScanStatusReasonCreated,
			nil,
		),
		ResourceCleanupStatus: types.NewResourceCleanupStatus(
			types.ResourceCleanupStatusStatePending,
			types.ResourceCleanupStatusReasonAssetScanCreated,
			nil,
		),
		Sbom: &types.SbomScan{
			Packages: nil,
			Status:   mapFamilyConfigToScannerStatus(familiesConfig.Sbom),
		},
		Exploits: &types.ExploitScan{
			Exploits: nil,
			Status:   mapFamilyConfigToScannerStatus(familiesConfig.Exploits),
		},
		Vulnerabilities: &types.VulnerabilityScan{
			Vulnerabilities: nil,
			Status:          mapFamilyConfigToScannerStatus(familiesConfig.Vulnerabilities),
		},
		Malware: &types.MalwareScan{
			Malware:  nil,
			Metadata: nil,
			Status:   mapFamilyConfigToScannerStatus(familiesConfig.Malware),
		},
		Rootkits: &types.RootkitScan{
			Rootkits: nil,
			Status:   mapFamilyConfigToScannerStatus(familiesConfig.Rootkits),
		},
		Secrets: &types.SecretScan{
			Secrets: nil,
			Status:  mapFamilyConfigToScannerStatus(familiesConfig.Secrets),
		},
		Misconfigurations: &types.MisconfigurationScan{
			Misconfigurations: nil,
			Scanners:          nil,
			Status:            mapFamilyConfigToScannerStatus(familiesConfig.Misconfigurations),
		},
		InfoFinder: &types.InfoFinderScan{
			Infos:    nil,
			Scanners: nil,
			Status:   mapFamilyConfigToScannerStatus(familiesConfig.InfoFinder),
		},
	}, nil
}

func mapFamilyConfigToScannerStatus(config types.FamilyConfigEnabler) *types.ScannerStatus {
	if config == nil || !config.IsEnabled() {
		return types.NewScannerStatus(types.ScannerStatusStateSkipped, types.ScannerStatusReasonNotScheduled, nil)
	}

	return types.NewScannerStatus(types.ScannerStatusStatePending, types.ScannerStatusReasonScheduled, nil)
}

func newScanSummary() *types.ScanSummary {
	return &types.ScanSummary{
		JobsCompleted:          utils.PointerTo(0),
		JobsLeftToRun:          utils.PointerTo(0),
		TotalExploits:          utils.PointerTo(0),
		TotalMalware:           utils.PointerTo(0),
		TotalMisconfigurations: utils.PointerTo(0),
		TotalPackages:          utils.PointerTo(0),
		TotalRootkits:          utils.PointerTo(0),
		TotalSecrets:           utils.PointerTo(0),
		TotalInfoFinder:        utils.PointerTo(0),
		TotalVulnerabilities: &types.VulnerabilityScanSummary{
			TotalCriticalVulnerabilities:   utils.PointerTo(0),
			TotalHighVulnerabilities:       utils.PointerTo(0),
			TotalLowVulnerabilities:        utils.PointerTo(0),
			TotalMediumVulnerabilities:     utils.PointerTo(0),
			TotalNegligibleVulnerabilities: utils.PointerTo(0),
		},
	}
}

func updateScanSummaryFromAssetScan(scan *types.Scan, result types.AssetScan) error {
	if result.Summary == nil {
		return errors.New("invalid AssetScan: Summary field is nil")
	}

	if scan.Summary == nil {
		scan.Summary = newScanSummary()
	}

	status, ok := result.GetStatus()
	if !ok {
		return fmt.Errorf("status must not be nil for AssetScan. AssetScanID=%s", *result.Id)
	}

	s, r := scan.Summary, result.Summary

	switch status.State {
	case types.AssetScanStatusStatePending, types.AssetScanStatusStateScheduled, types.AssetScanStatusStateReadyToScan:
		fallthrough
	case types.AssetScanStatusStateInProgress, types.AssetScanStatusStateAborted:
		s.JobsLeftToRun = utils.PointerTo(*s.JobsLeftToRun + 1)
	case types.AssetScanStatusStateDone, types.AssetScanStatusStateFailed:
		s.JobsCompleted = utils.PointerTo(*s.JobsCompleted + 1)
		s.TotalExploits = utils.PointerTo(*s.TotalExploits + *r.TotalExploits)
		s.TotalInfoFinder = utils.PointerTo(*s.TotalInfoFinder + *r.TotalInfoFinder)
		s.TotalMalware = utils.PointerTo(*s.TotalMalware + *r.TotalMalware)
		s.TotalMisconfigurations = utils.PointerTo(*s.TotalMisconfigurations + *r.TotalMisconfigurations)
		s.TotalPackages = utils.PointerTo(*s.TotalPackages + *r.TotalPackages)
		s.TotalRootkits = utils.PointerTo(*s.TotalRootkits + *r.TotalRootkits)
		s.TotalSecrets = utils.PointerTo(*s.TotalSecrets + *r.TotalSecrets)
		s.TotalVulnerabilities = &types.VulnerabilityScanSummary{
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
