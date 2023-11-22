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
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

func TestNewAssetScanFromScan(t *testing.T) {
	scanID := string(uuid.NewUUID())
	assetID := string(uuid.NewUUID())

	transitionTime := time.Now()

	resourceCleanupStatus := models.NewResourceCleanupStatus(
		models.ResourceCleanupStatusStatePending,
		models.ResourceCleanupStatusReasonAssetScanCreated,
		nil,
	)
	resourceCleanupStatus.LastTransitionTime = transitionTime

	sbomScanStatus := models.NewScannerStatus(
		models.ScannerStatusStatePending,
		models.ScannerStatusReasonScheduled,
		nil,
	)
	sbomScanStatus.LastTransitionTime = transitionTime

	exploitScanStatus := models.NewScannerStatus(
		models.ScannerStatusStatePending,
		models.ScannerStatusReasonScheduled,
		nil,
	)
	exploitScanStatus.LastTransitionTime = transitionTime

	vulnerabilityScanStatus := models.NewScannerStatus(
		models.ScannerStatusStatePending,
		models.ScannerStatusReasonScheduled,
		nil,
	)
	vulnerabilityScanStatus.LastTransitionTime = transitionTime

	malwareScanStatus := models.NewScannerStatus(
		models.ScannerStatusStatePending,
		models.ScannerStatusReasonScheduled,
		nil,
	)
	malwareScanStatus.LastTransitionTime = transitionTime

	rootkitScanStatus := models.NewScannerStatus(
		models.ScannerStatusStateSkipped,
		models.ScannerStatusReasonNotScheduled,
		nil,
	)
	rootkitScanStatus.LastTransitionTime = transitionTime

	secretScanStatus := models.NewScannerStatus(
		models.ScannerStatusStateSkipped,
		models.ScannerStatusReasonNotScheduled,
		nil,
	)
	secretScanStatus.LastTransitionTime = transitionTime

	misconfigurationScanStatus := models.NewScannerStatus(
		models.ScannerStatusStateSkipped,
		models.ScannerStatusReasonNotScheduled,
		nil,
	)
	misconfigurationScanStatus.LastTransitionTime = transitionTime

	infoFinderScanStatus := models.NewScannerStatus(
		models.ScannerStatusStatePending,
		models.ScannerStatusReasonScheduled,
		nil,
	)
	infoFinderScanStatus.LastTransitionTime = transitionTime

	tests := []struct {
		Name    string
		Scan    *models.Scan
		AssetID string

		ExpectedErrorMatcher types.GomegaMatcher
		ExpectedAssetScan    *models.AssetScan
	}{
		{
			Name: "AssetResult from valid Scan",
			Scan: &models.Scan{
				Name:                utils.PointerTo("test-1234"),
				Id:                  utils.PointerTo(scanID),
				MaxParallelScanners: utils.PointerTo(2),
				AssetScanTemplate: &models.AssetScanTemplate{
					ScanFamiliesConfig: &models.ScanFamiliesConfig{
						Exploits: &models.ExploitsConfig{
							Enabled: utils.PointerTo(true),
						},
						Malware: &models.MalwareConfig{
							Enabled: utils.PointerTo(true),
						},
						Misconfigurations: nil,
						Rootkits:          nil,
						Sbom: &models.SBOMConfig{
							Enabled: utils.PointerTo(true),
						},
						Secrets: nil,
						Vulnerabilities: &models.VulnerabilitiesConfig{
							Enabled: utils.PointerTo(true),
						},
						InfoFinder: &models.InfoFinderConfig{
							Enabled:  utils.PointerTo(true),
							Scanners: utils.PointerTo([]string{"test"}),
						},
					},
				},
			},
			AssetID:              assetID,
			ExpectedErrorMatcher: Not(HaveOccurred()),
			ExpectedAssetScan: &models.AssetScan{
				ResourceCleanupStatus: resourceCleanupStatus,
				Scan: &models.ScanRelationship{
					Id: scanID,
				},
				Status: &models.AssetScanStatus{
					General: &models.AssetScanState{
						Errors: nil,
						State:  utils.PointerTo(models.AssetScanStateStatePending),
					},
				},
				Sbom: &models.SbomScan{
					Packages: nil,
					Status:   sbomScanStatus,
				},
				Exploits: &models.ExploitScan{
					Exploits: nil,
					Status:   exploitScanStatus,
				},
				Vulnerabilities: &models.VulnerabilityScan{
					Vulnerabilities: nil,
					Status:          vulnerabilityScanStatus,
				},
				Malware: &models.MalwareScan{
					Malware:  nil,
					Metadata: nil,
					Status:   malwareScanStatus,
				},
				Rootkits: &models.RootkitScan{
					Rootkits: nil,
					Status:   rootkitScanStatus,
				},
				Secrets: &models.SecretScan{
					Secrets: nil,
					Status:  secretScanStatus,
				},
				Misconfigurations: &models.MisconfigurationScan{
					Misconfigurations: nil,
					Scanners:          nil,
					Status:            misconfigurationScanStatus,
				},
				InfoFinder: &models.InfoFinderScan{
					Infos:    nil,
					Scanners: nil,
					Status:   infoFinderScanStatus,
				},
				Summary: newAssetScanSummary(),
				Asset: &models.AssetRelationship{
					Id: assetID,
				},
				ScanFamiliesConfig: &models.ScanFamiliesConfig{
					Exploits: &models.ExploitsConfig{
						Enabled: utils.PointerTo(true),
					},
					Malware: &models.MalwareConfig{
						Enabled: utils.PointerTo(true),
					},
					Misconfigurations: nil,
					Rootkits:          nil,
					Sbom: &models.SBOMConfig{
						Enabled: utils.PointerTo(true),
					},
					Secrets: nil,
					Vulnerabilities: &models.VulnerabilitiesConfig{
						Enabled: utils.PointerTo(true),
					},
					InfoFinder: &models.InfoFinderConfig{
						Enabled:  utils.PointerTo(true),
						Scanners: utils.PointerTo([]string{"test"}),
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			result, err := newAssetScanFromScan(test.Scan, test.AssetID)
			result.ResourceCleanupStatus.LastTransitionTime = transitionTime
			result.Sbom.Status.LastTransitionTime = transitionTime
			result.Exploits.Status.LastTransitionTime = transitionTime
			result.Vulnerabilities.Status.LastTransitionTime = transitionTime
			result.Malware.Status.LastTransitionTime = transitionTime
			result.Rootkits.Status.LastTransitionTime = transitionTime
			result.Secrets.Status.LastTransitionTime = transitionTime
			result.Misconfigurations.Status.LastTransitionTime = transitionTime
			result.InfoFinder.Status.LastTransitionTime = transitionTime

			g.Expect(err).Should(test.ExpectedErrorMatcher)
			g.Expect(result).Should(BeComparableTo(test.ExpectedAssetScan))
		})
	}
}
