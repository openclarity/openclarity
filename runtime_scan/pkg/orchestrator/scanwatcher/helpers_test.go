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

func TestNewScanResultFromScan(t *testing.T) {
	scanID := string(uuid.NewUUID())
	targetID := string(uuid.NewUUID())

	tests := []struct {
		Name     string
		Scan     *models.Scan
		TargetID string

		ExpectedErrorMatcher types.GomegaMatcher
		ExpectedScanResult   *models.TargetScanResult
	}{
		{
			Name: "TargetResult from valid Scan",
			Scan: &models.Scan{
				Id: utils.PointerTo(scanID),
				ScanConfigSnapshot: &models.ScanConfigSnapshot{
					Disabled:            utils.PointerTo(false),
					MaxParallelScanners: utils.PointerTo(2),
					Name:                utils.PointerTo("test"),
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
			TargetID:             targetID,
			ExpectedErrorMatcher: Not(HaveOccurred()),
			ExpectedScanResult: &models.TargetScanResult{
				ResourceCleanup: utils.PointerTo(models.ResourceCleanupStatePENDING),
				Scan: &models.ScanRelationship{
					Id: scanID,
				},
				Status: &models.TargetScanStatus{
					Exploits: &models.TargetScanState{
						Errors: nil,
						State:  utils.PointerTo(models.TargetScanStateStateINIT),
					},
					General: &models.TargetScanState{
						Errors: nil,
						State:  utils.PointerTo(models.TargetScanStateStateINIT),
					},
					Malware: &models.TargetScanState{
						Errors: nil,
						State:  utils.PointerTo(models.TargetScanStateStateINIT),
					},
					Misconfigurations: &models.TargetScanState{
						Errors: nil,
						State:  utils.PointerTo(models.TargetScanStateStateNOTSCANNED),
					},
					Rootkits: &models.TargetScanState{
						Errors: nil,
						State:  utils.PointerTo(models.TargetScanStateStateNOTSCANNED),
					},
					Sbom: &models.TargetScanState{
						Errors: nil,
						State:  utils.PointerTo(models.TargetScanStateStateINIT),
					},
					Secrets: &models.TargetScanState{
						Errors: nil,
						State:  utils.PointerTo(models.TargetScanStateStateNOTSCANNED),
					},
					Vulnerabilities: &models.TargetScanState{
						Errors: nil,
						State:  utils.PointerTo(models.TargetScanStateStateINIT),
					},
				},
				Summary: newScanResultSummary(),
				Target: &models.TargetRelationship{
					Id: targetID,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			result, err := newScanResultFromScan(test.Scan, test.TargetID)

			g.Expect(err).Should(test.ExpectedErrorMatcher)
			g.Expect(result).Should(Equal(test.ExpectedScanResult))
		})
	}
}
