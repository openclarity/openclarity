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

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

func TestNewAssetScanFromScan(t *testing.T) {
	scanID := string(uuid.NewUUID())
	assetID := string(uuid.NewUUID())

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
					},
				},
			},
			AssetID:              assetID,
			ExpectedErrorMatcher: Not(HaveOccurred()),
			ExpectedAssetScan: &models.AssetScan{
				ResourceCleanup: utils.PointerTo(models.ResourceCleanupStatePending),
				Scan: &models.ScanRelationship{
					Id: scanID,
				},
				Status: &models.AssetScanStatus{
					Exploits: &models.AssetScanState{
						Errors: nil,
						State:  utils.PointerTo(models.AssetScanStateStatePending),
					},
					General: &models.AssetScanState{
						Errors: nil,
						State:  utils.PointerTo(models.AssetScanStateStatePending),
					},
					Malware: &models.AssetScanState{
						Errors: nil,
						State:  utils.PointerTo(models.AssetScanStateStatePending),
					},
					Misconfigurations: &models.AssetScanState{
						Errors: nil,
						State:  utils.PointerTo(models.AssetScanStateStateNotScanned),
					},
					Rootkits: &models.AssetScanState{
						Errors: nil,
						State:  utils.PointerTo(models.AssetScanStateStateNotScanned),
					},
					Sbom: &models.AssetScanState{
						Errors: nil,
						State:  utils.PointerTo(models.AssetScanStateStatePending),
					},
					Secrets: &models.AssetScanState{
						Errors: nil,
						State:  utils.PointerTo(models.AssetScanStateStateNotScanned),
					},
					Vulnerabilities: &models.AssetScanState{
						Errors: nil,
						State:  utils.PointerTo(models.AssetScanStateStatePending),
					},
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
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			result, err := newAssetScanFromScan(test.Scan, test.AssetID)

			g.Expect(err).Should(test.ExpectedErrorMatcher)
			g.Expect(result).Should(BeComparableTo(test.ExpectedAssetScan))
		})
	}
}
