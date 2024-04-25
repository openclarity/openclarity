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
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"k8s.io/apimachinery/pkg/util/uuid"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
)

func TestNewAssetScanFromScan(t *testing.T) {
	scanID := string(uuid.NewUUID())
	assetID := string(uuid.NewUUID())

	transitionTime := time.Now()

	status := apitypes.NewAssetScanStatus(
		apitypes.AssetScanStatusStatePending,
		apitypes.AssetScanStatusReasonCreated,
		nil,
	)
	status.LastTransitionTime = transitionTime

	resourceCleanupStatus := apitypes.NewResourceCleanupStatus(
		apitypes.ResourceCleanupStatusStatePending,
		apitypes.ResourceCleanupStatusReasonAssetScanCreated,
		nil,
	)
	resourceCleanupStatus.LastTransitionTime = transitionTime

	sbomScanStatus := apitypes.NewScannerStatus(
		apitypes.ScannerStatusStatePending,
		apitypes.ScannerStatusReasonScheduled,
		nil,
	)
	sbomScanStatus.LastTransitionTime = transitionTime

	exploitScanStatus := apitypes.NewScannerStatus(
		apitypes.ScannerStatusStatePending,
		apitypes.ScannerStatusReasonScheduled,
		nil,
	)
	exploitScanStatus.LastTransitionTime = transitionTime

	vulnerabilityScanStatus := apitypes.NewScannerStatus(
		apitypes.ScannerStatusStatePending,
		apitypes.ScannerStatusReasonScheduled,
		nil,
	)
	vulnerabilityScanStatus.LastTransitionTime = transitionTime

	malwareScanStatus := apitypes.NewScannerStatus(
		apitypes.ScannerStatusStatePending,
		apitypes.ScannerStatusReasonScheduled,
		nil,
	)
	malwareScanStatus.LastTransitionTime = transitionTime

	rootkitScanStatus := apitypes.NewScannerStatus(
		apitypes.ScannerStatusStateSkipped,
		apitypes.ScannerStatusReasonNotScheduled,
		nil,
	)
	rootkitScanStatus.LastTransitionTime = transitionTime

	secretScanStatus := apitypes.NewScannerStatus(
		apitypes.ScannerStatusStateSkipped,
		apitypes.ScannerStatusReasonNotScheduled,
		nil,
	)
	secretScanStatus.LastTransitionTime = transitionTime

	misconfigurationScanStatus := apitypes.NewScannerStatus(
		apitypes.ScannerStatusStateSkipped,
		apitypes.ScannerStatusReasonNotScheduled,
		nil,
	)
	misconfigurationScanStatus.LastTransitionTime = transitionTime

	infoFinderScanStatus := apitypes.NewScannerStatus(
		apitypes.ScannerStatusStatePending,
		apitypes.ScannerStatusReasonScheduled,
		nil,
	)
	infoFinderScanStatus.LastTransitionTime = transitionTime

	pluginScanStatus := apitypes.NewScannerStatus(
		apitypes.ScannerStatusStateSkipped,
		apitypes.ScannerStatusReasonNotScheduled,
		nil,
	)
	pluginScanStatus.LastTransitionTime = transitionTime

	tests := []struct {
		Name    string
		Scan    *apitypes.Scan
		AssetID string

		ExpectedErrorMatcher types.GomegaMatcher
		ExpectedAssetScan    *apitypes.AssetScan
	}{
		{
			Name: "AssetResult from valid Scan",
			Scan: &apitypes.Scan{
				Name:                to.Ptr("test-1234"),
				Id:                  to.Ptr(scanID),
				MaxParallelScanners: to.Ptr(2),
				AssetScanTemplate: &apitypes.AssetScanTemplate{
					ScanFamiliesConfig: &apitypes.ScanFamiliesConfig{
						Exploits: &apitypes.ExploitsConfig{
							Enabled: to.Ptr(true),
						},
						Malware: &apitypes.MalwareConfig{
							Enabled: to.Ptr(true),
						},
						Misconfigurations: nil,
						Rootkits:          nil,
						Sbom: &apitypes.SBOMConfig{
							Enabled: to.Ptr(true),
						},
						Secrets: nil,
						Vulnerabilities: &apitypes.VulnerabilitiesConfig{
							Enabled: to.Ptr(true),
						},
						InfoFinder: &apitypes.InfoFinderConfig{
							Enabled:  to.Ptr(true),
							Scanners: to.Ptr([]string{"test"}),
						},
						Plugins: nil,
					},
				},
			},
			AssetID:              assetID,
			ExpectedErrorMatcher: Not(HaveOccurred()),
			ExpectedAssetScan: &apitypes.AssetScan{
				ResourceCleanupStatus: resourceCleanupStatus,
				Scan: &apitypes.ScanRelationship{
					Id: scanID,
				},
				Status: status,
				Sbom: &apitypes.SbomScan{
					Packages: nil,
					Status:   sbomScanStatus,
				},
				Exploits: &apitypes.ExploitScan{
					Exploits: nil,
					Status:   exploitScanStatus,
				},
				Vulnerabilities: &apitypes.VulnerabilityScan{
					Vulnerabilities: nil,
					Status:          vulnerabilityScanStatus,
				},
				Malware: &apitypes.MalwareScan{
					Malware:  nil,
					Metadata: nil,
					Status:   malwareScanStatus,
				},
				Rootkits: &apitypes.RootkitScan{
					Rootkits: nil,
					Status:   rootkitScanStatus,
				},
				Secrets: &apitypes.SecretScan{
					Secrets: nil,
					Status:  secretScanStatus,
				},
				Misconfigurations: &apitypes.MisconfigurationScan{
					Misconfigurations: nil,
					Scanners:          nil,
					Status:            misconfigurationScanStatus,
				},
				InfoFinder: &apitypes.InfoFinderScan{
					Infos:    nil,
					Scanners: nil,
					Status:   infoFinderScanStatus,
				},
				Plugins: &apitypes.PluginScan{
					FindingInfos: nil,
					Status:       pluginScanStatus,
				},
				Summary: newAssetScanSummary(),
				Asset: &apitypes.AssetRelationship{
					Id: assetID,
				},
				ScanFamiliesConfig: &apitypes.ScanFamiliesConfig{
					Exploits: &apitypes.ExploitsConfig{
						Enabled: to.Ptr(true),
					},
					Malware: &apitypes.MalwareConfig{
						Enabled: to.Ptr(true),
					},
					Misconfigurations: nil,
					Rootkits:          nil,
					Sbom: &apitypes.SBOMConfig{
						Enabled: to.Ptr(true),
					},
					Secrets: nil,
					Vulnerabilities: &apitypes.VulnerabilitiesConfig{
						Enabled: to.Ptr(true),
					},
					InfoFinder: &apitypes.InfoFinderConfig{
						Enabled:  to.Ptr(true),
						Scanners: to.Ptr([]string{"test"}),
					},
					Plugins: nil,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			result, err := newAssetScanFromScan(test.Scan, test.AssetID)
			result.Status.LastTransitionTime = transitionTime
			result.ResourceCleanupStatus.LastTransitionTime = transitionTime
			result.Sbom.Status.LastTransitionTime = transitionTime
			result.Exploits.Status.LastTransitionTime = transitionTime
			result.Vulnerabilities.Status.LastTransitionTime = transitionTime
			result.Malware.Status.LastTransitionTime = transitionTime
			result.Rootkits.Status.LastTransitionTime = transitionTime
			result.Secrets.Status.LastTransitionTime = transitionTime
			result.Misconfigurations.Status.LastTransitionTime = transitionTime
			result.InfoFinder.Status.LastTransitionTime = transitionTime
			result.Plugins.Status.LastTransitionTime = transitionTime

			g.Expect(err).Should(test.ExpectedErrorMatcher)
			g.Expect(result).Should(BeComparableTo(test.ExpectedAssetScan))
		})
	}
}
