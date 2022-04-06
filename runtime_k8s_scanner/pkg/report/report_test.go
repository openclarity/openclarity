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

package report

import (
	"sort"
	"testing"

	"gotest.tools/assert"

	"github.com/cisco-open/kubei/runtime_scan/api/client/models"
	"github.com/cisco-open/kubei/shared/pkg/scanner"
)

func Test_createPackagesVulnerabilitiesScan(t *testing.T) {
	cvssVector := "CVSS:3.1/AV:N/AC:H/PR:H/UI:N/S:U/C:H/I:H/A:H"
	score := 9.9
	type args struct {
		results *scanner.MergedResults
	}
	tests := []struct {
		name string
		args args
		want []*models.PackageVulnerabilityScan
	}{
		{
			name: "first vulnerability in `VulnerabilityKey with diff` should be ignored",
			args: args{
				results: &scanner.MergedResults{
					MergedVulnerabilitiesByKey: map[scanner.VulnerabilityKey][]scanner.MergedVulnerability{
						"VulnerabilityKey with diff": {
							{
								ID: "ID1",
								Vulnerability: scanner.Vulnerability{
									ID:          "Vulnerability.ID",
									Description: "Vulnerability.Description",
									Links:       []string{"link"},
									Distro: scanner.Distro{
										Name:    "Vulnerability.Distro.Name",
										Version: "Vulnerability.Distro.Version",
									},
									CVSS: []scanner.CVSS{
										{
											Version: "3.1",
											Vector:  cvssVector,
											Metrics: scanner.CvssMetrics{
												BaseScore:           score,
												ExploitabilityScore: &score,
												ImpactScore:         &score,
											},
										},
									},
									Fix: scanner.Fix{
										Versions: []string{"fix"},
										State:    "",
									},
									Severity: "LOW",
									Package: scanner.Package{
										Name:     "Vulnerability.Package.Name",
										Version:  "Vulnerability.Package.Version",
										Type:     "Vulnerability.Package.Type",
										Language: "Vulnerability.Package.Language",
										Licenses: []string{"Vulnerability.Package.License"},
										CPEs:     []string{"Vulnerability.Package.CPE"},
										PURL:     "Vulnerability.Package.PURL",
									},
									LayerID: "Vulnerability.LayerID",
									Path:    "Vulnerability.Path",
								},
								ScannersInfo: []scanner.Info{
									{
										Name: "scanner2",
									},
								},
							},
							{
								ID: "ID2",
								Vulnerability: scanner.Vulnerability{
									ID:          "Vulnerability.ID",
									Description: "Vulnerability.Description",
									Links:       []string{"link"},
									Distro: scanner.Distro{
										Name:    "Vulnerability.Distro.Name",
										Version: "Vulnerability.Distro.Version",
									},
									CVSS: []scanner.CVSS{
										{
											Version: "3.1",
											Vector:  cvssVector,
											Metrics: scanner.CvssMetrics{
												BaseScore:           score,
												ExploitabilityScore: &score,
												ImpactScore:         &score,
											},
										},
									},
									Fix: scanner.Fix{
										Versions: []string{"fix"},
										State:    "",
									},
									Severity: "HIGH",
									Package: scanner.Package{
										Name:     "Vulnerability.Package.Name",
										Version:  "Vulnerability.Package.Version",
										Type:     "Vulnerability.Package.Type",
										Language: "Vulnerability.Package.Language",
										Licenses: []string{"Vulnerability.Package.License"},
										CPEs:     []string{"Vulnerability.Package.CPE"},
										PURL:     "Vulnerability.Package.PURL",
									},
									LayerID: "Vulnerability.LayerID",
									Path:    "Vulnerability.Path",
								},
								ScannersInfo: []scanner.Info{
									{
										Name: "scanner1",
									},
								},
								Diffs: []scanner.DiffInfo{
									{
										CompareToID: "ID1",
										JSONDiff:    map[string]interface{}{"severity": []string{"HIGH", "LOW"}},
									},
								},
							},
						},
						"VulnerabilityKey with no diff": {
							{
								ID: "ID3",
								Vulnerability: scanner.Vulnerability{
									ID:          "Vulnerability.ID3",
									Description: "Vulnerability.Description3",
									Links:       []string{"link3"},
									Distro: scanner.Distro{
										Name:    "Vulnerability.Distro.Name3",
										Version: "Vulnerability.Distro.Version3",
									},
									CVSS: []scanner.CVSS{
										{
											Version: "3.1",
											Vector:  cvssVector,
											Metrics: scanner.CvssMetrics{
												BaseScore:           score,
												ExploitabilityScore: &score,
												ImpactScore:         &score,
											},
										},
									},
									Fix: scanner.Fix{
										Versions: []string{"fix3"},
										State:    "",
									},
									Severity: "HIGH",
									Package: scanner.Package{
										Name:     "Vulnerability.Package.Name2",
										Version:  "Vulnerability.Package.Version2",
										Type:     "Vulnerability.Package.Type2",
										Language: "Vulnerability.Package.Language2",
										Licenses: []string{"Vulnerability.Package.License2"},
										CPEs:     []string{"Vulnerability.Package.CPE2"},
										PURL:     "Vulnerability.Package.PURL2",
									},
									LayerID: "Vulnerability.LayerID3",
									Path:    "Vulnerability.Path3",
								},
								ScannersInfo: []scanner.Info{
									{
										Name: "scanner3",
									},
								},
							},
						},
					},
					Source: scanner.Source{
						Type: "Source.Type",
						Name: "Source.Name",
						Hash: "Source.Hash",
					},
				},
			},
			want: []*models.PackageVulnerabilityScan{
				{
					Cvss:        createTestCVSS(cvssVector, score),
					Description: "Vulnerability.Description",
					FixVersion:  "fix",
					LayerID:     "Vulnerability.LayerID",
					Links:       []string{"link"},
					Package: &models.PackageInfo{
						Language: "Vulnerability.Package.Language",
						License:  "Vulnerability.Package.License",
						Name:     "Vulnerability.Package.Name",
						Version:  "Vulnerability.Package.Version",
					},
					Scanners:          []string{"scanner1"},
					Severity:          "HIGH",
					VulnerabilityName: "Vulnerability.ID",
				},
				{
					Cvss:        createTestCVSS(cvssVector, score),
					Description: "Vulnerability.Description3",
					FixVersion:  "fix3",
					LayerID:     "Vulnerability.LayerID3",
					Links:       []string{"link3"},
					Package: &models.PackageInfo{
						Language: "Vulnerability.Package.Language2",
						License:  "Vulnerability.Package.License2",
						Name:     "Vulnerability.Package.Name2",
						Version:  "Vulnerability.Package.Version2",
					},
					Scanners:          []string{"scanner3"},
					Severity:          "HIGH",
					VulnerabilityName: "Vulnerability.ID3",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createPackagesVulnerabilitiesScan(tt.args.results)
			sort.Slice(got, func(i, j int) bool {
				return got[i].VulnerabilityName < got[j].VulnerabilityName
			})
			sort.Slice(tt.want, func(i, j int) bool {
				return tt.want[i].VulnerabilityName < tt.want[j].VulnerabilityName
			})
			assert.DeepEqual(t, got, tt.want)
		})
	}
}

func createTestCVSS(vector string, score float64) *models.CVSS {
	return &models.CVSS{
		CvssV3Metrics: &models.CVSSV3Metrics{
			BaseScore:           score,
			ExploitabilityScore: score,
			ImpactScore:         score,
		},
		CvssV3Vector: &models.CVSSV3Vector{
			AttackComplexity:   models.AttackComplexityHIGH,
			AttackVector:       models.AttackVectorNETWORK,
			Availability:       models.AvailabilityHIGH,
			Confidentiality:    models.ConfidentialityHIGH,
			Integrity:          models.IntegrityHIGH,
			PrivilegesRequired: models.PrivilegesRequiredHIGH,
			Scope:              models.ScopeUNCHANGED,
			UserInteraction:    models.UserInteractionNONE,
			Vector:             vector,
		},
	}
}
