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

package scan

import (
	"errors"
	"fmt"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
)

func newAssetScanSummary() *apitypes.ScanFindingsSummary {
	return &apitypes.ScanFindingsSummary{
		TotalExploits:          to.Ptr[int](0),
		TotalMalware:           to.Ptr[int](0),
		TotalMisconfigurations: to.Ptr[int](0),
		TotalPackages:          to.Ptr[int](0),
		TotalRootkits:          to.Ptr[int](0),
		TotalSecrets:           to.Ptr[int](0),
		TotalInfoFinder:        to.Ptr[int](0),
		TotalVulnerabilities: &apitypes.VulnerabilitySeveritySummary{
			TotalCriticalVulnerabilities:   to.Ptr[int](0),
			TotalHighVulnerabilities:       to.Ptr[int](0),
			TotalMediumVulnerabilities:     to.Ptr[int](0),
			TotalLowVulnerabilities:        to.Ptr[int](0),
			TotalNegligibleVulnerabilities: to.Ptr[int](0),
		},
		TotalPlugins: to.Ptr[int](0),
	}
}

func newAssetScanFromScan(scan *apitypes.Scan, assetID string) (*apitypes.AssetScan, error) {
	if scan == nil {
		return nil, errors.New("failed to create AssetScan: Scan is nil")
	}

	if scan.AssetScanTemplate == nil || scan.AssetScanTemplate.ScanFamiliesConfig == nil {
		return nil, errors.New("failed to create AssetScan: AssetScanTemplate and/or AssetScanTemplate.ScanFamiliesConfig is nil")
	}
	familiesConfig := scan.AssetScanTemplate.ScanFamiliesConfig

	return &apitypes.AssetScan{
		Summary: newAssetScanSummary(),
		Scan: &apitypes.ScanRelationship{
			Id: *scan.Id,
		},
		Asset: &apitypes.AssetRelationship{
			Id: assetID,
		},
		ScanFamiliesConfig:            familiesConfig,
		ScannerInstanceCreationConfig: scan.AssetScanTemplate.ScannerInstanceCreationConfig,
		Status: apitypes.NewAssetScanStatus(
			apitypes.AssetScanStatusStatePending,
			apitypes.AssetScanStatusReasonCreated,
			nil,
		),
		ResourceCleanupStatus: apitypes.NewResourceCleanupStatus(
			apitypes.ResourceCleanupStatusStatePending,
			apitypes.ResourceCleanupStatusReasonAssetScanCreated,
			nil,
		),
		Sbom: &apitypes.SbomScan{
			Packages: nil,
			Status:   mapFamilyConfigToScannerStatus(familiesConfig.Sbom),
		},
		Exploits: &apitypes.ExploitScan{
			Exploits: nil,
			Status:   mapFamilyConfigToScannerStatus(familiesConfig.Exploits),
		},
		Vulnerabilities: &apitypes.VulnerabilityScan{
			Vulnerabilities: nil,
			Status:          mapFamilyConfigToScannerStatus(familiesConfig.Vulnerabilities),
		},
		Malware: &apitypes.MalwareScan{
			Malware:  nil,
			Metadata: nil,
			Status:   mapFamilyConfigToScannerStatus(familiesConfig.Malware),
		},
		Rootkits: &apitypes.RootkitScan{
			Rootkits: nil,
			Status:   mapFamilyConfigToScannerStatus(familiesConfig.Rootkits),
		},
		Secrets: &apitypes.SecretScan{
			Secrets: nil,
			Status:  mapFamilyConfigToScannerStatus(familiesConfig.Secrets),
		},
		Misconfigurations: &apitypes.MisconfigurationScan{
			Misconfigurations: nil,
			Scanners:          nil,
			Status:            mapFamilyConfigToScannerStatus(familiesConfig.Misconfigurations),
		},
		InfoFinder: &apitypes.InfoFinderScan{
			Infos:    nil,
			Scanners: nil,
			Status:   mapFamilyConfigToScannerStatus(familiesConfig.InfoFinder),
		},
		Plugins: &apitypes.PluginScan{
			FindingInfos: nil,
			Status:       mapFamilyConfigToScannerStatus(familiesConfig.Plugins),
		},
	}, nil
}

func mapFamilyConfigToScannerStatus(config apitypes.FamilyConfigEnabler) *apitypes.ScannerStatus {
	if config == nil || !config.IsEnabled() {
		return apitypes.NewScannerStatus(apitypes.ScannerStatusStateSkipped, apitypes.ScannerStatusReasonNotScheduled, nil)
	}

	return apitypes.NewScannerStatus(apitypes.ScannerStatusStatePending, apitypes.ScannerStatusReasonScheduled, nil)
}

func newScanSummary() *apitypes.ScanSummary {
	return &apitypes.ScanSummary{
		JobsCompleted:          to.Ptr(0),
		JobsLeftToRun:          to.Ptr(0),
		TotalExploits:          to.Ptr(0),
		TotalMalware:           to.Ptr(0),
		TotalMisconfigurations: to.Ptr(0),
		TotalPackages:          to.Ptr(0),
		TotalRootkits:          to.Ptr(0),
		TotalSecrets:           to.Ptr(0),
		TotalInfoFinder:        to.Ptr(0),
		TotalVulnerabilities: &apitypes.VulnerabilitySeveritySummary{
			TotalCriticalVulnerabilities:   to.Ptr(0),
			TotalHighVulnerabilities:       to.Ptr(0),
			TotalLowVulnerabilities:        to.Ptr(0),
			TotalMediumVulnerabilities:     to.Ptr(0),
			TotalNegligibleVulnerabilities: to.Ptr(0),
		},
		TotalPlugins: to.Ptr(0),
	}
}

func updateScanSummaryFromAssetScan(scan *apitypes.Scan, result apitypes.AssetScan) error {
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
	case apitypes.AssetScanStatusStatePending, apitypes.AssetScanStatusStateScheduled, apitypes.AssetScanStatusStateReadyToScan:
		fallthrough
	case apitypes.AssetScanStatusStateInProgress, apitypes.AssetScanStatusStateAborted:
		s.JobsLeftToRun = to.Ptr(*s.JobsLeftToRun + 1)
	case apitypes.AssetScanStatusStateDone, apitypes.AssetScanStatusStateFailed:
		s.JobsCompleted = to.Ptr(*s.JobsCompleted + 1)
		s.TotalExploits = to.Ptr(*s.TotalExploits + *r.TotalExploits)
		s.TotalInfoFinder = to.Ptr(*s.TotalInfoFinder + *r.TotalInfoFinder)
		s.TotalMalware = to.Ptr(*s.TotalMalware + *r.TotalMalware)
		s.TotalMisconfigurations = to.Ptr(*s.TotalMisconfigurations + *r.TotalMisconfigurations)
		s.TotalPackages = to.Ptr(*s.TotalPackages + *r.TotalPackages)
		s.TotalRootkits = to.Ptr(*s.TotalRootkits + *r.TotalRootkits)
		s.TotalSecrets = to.Ptr(*s.TotalSecrets + *r.TotalSecrets)
		s.TotalVulnerabilities = &apitypes.VulnerabilitySeveritySummary{
			TotalCriticalVulnerabilities: to.Ptr(*s.TotalVulnerabilities.TotalCriticalVulnerabilities +
				*r.TotalVulnerabilities.TotalCriticalVulnerabilities),
			TotalHighVulnerabilities: to.Ptr(*s.TotalVulnerabilities.TotalHighVulnerabilities +
				*r.TotalVulnerabilities.TotalHighVulnerabilities),
			TotalLowVulnerabilities: to.Ptr(*s.TotalVulnerabilities.TotalLowVulnerabilities +
				*r.TotalVulnerabilities.TotalLowVulnerabilities),
			TotalMediumVulnerabilities: to.Ptr(*s.TotalVulnerabilities.TotalMediumVulnerabilities +
				*r.TotalVulnerabilities.TotalMediumVulnerabilities),
			TotalNegligibleVulnerabilities: to.Ptr(*s.TotalVulnerabilities.TotalCriticalVulnerabilities +
				*r.TotalVulnerabilities.TotalNegligibleVulnerabilities),
		}
		s.TotalPlugins = to.Ptr(*s.TotalPlugins + *r.TotalPlugins)
	}

	return nil
}
