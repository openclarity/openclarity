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

package cmd

import (
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/openclarity/kubeclarity/shared/pkg/scanner"
	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/shared/pkg/families/sbom"
	"github.com/openclarity/vmclarity/shared/pkg/families/vulnerabilities"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

func Test_convertSBOMResultToApiModel(t *testing.T) {
	type args struct {
		result *sbom.Results
	}
	type returns struct {
		sbomScan *models.SbomScan
	}
	tests := []struct {
		name string
		args args
		want returns
	}{
		{
			name: "Full SBOM",
			args: args{
				result: &sbom.Results{
					SBOM: &cdx.BOM{
						Components: &[]cdx.Component{
							cdx.Component{
								BOMRef:     "bomref1",
								Type:       cdx.ComponentTypeLibrary,
								Name:       "testcomponent1",
								Version:    "v10.0.0-foo",
								PackageURL: "pkg:pypi/testcomponent1@v10.0.0-foo",
							},
							cdx.Component{
								BOMRef:     "bomref2",
								Type:       cdx.ComponentTypeLibrary,
								Name:       "testcomponent2",
								Version:    "v10.0.0-foo",
								PackageURL: "pkg:pypi/testcomponent2@v10.0.0-foo",
							},
						},
					},
				},
			},
			want: returns{
				sbomScan: &models.SbomScan{
					Packages: &[]models.Package{
						models.Package{
							Id: utils.StringPtr("bomref1"),
							PackageInfo: &models.PackageInfo{
								PackageName:    utils.StringPtr("testcomponent1"),
								PackageVersion: utils.StringPtr("v10.0.0-foo"),
							},
						},
						models.Package{
							Id: utils.StringPtr("bomref2"),
							PackageInfo: &models.PackageInfo{
								PackageName:    utils.StringPtr("testcomponent2"),
								PackageVersion: utils.StringPtr("v10.0.0-foo"),
							},
						},
					},
				},
			},
		},
		{
			name: "Nil components",
			args: args{
				result: &sbom.Results{
					SBOM: &cdx.BOM{
						Components: nil,
					},
				},
			},
			want: returns{
				sbomScan: &models.SbomScan{
					Packages: &[]models.Package{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertSBOMResultToApiModel(tt.args.result)
			if diff := cmp.Diff(tt.want.sbomScan, got, cmpopts.SortSlices(func(a, b models.Package) bool { return *a.Id < *b.Id })); diff != "" {
				t.Errorf("convertSBOMResultToApiModel() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_convertVulnResultToApiModel(t *testing.T) {
	type args struct {
		result *vulnerabilities.Results
	}
	type returns struct {
		vulScan *models.VulnerabilityScan
	}
	tests := []struct {
		name string
		args args
		want returns
	}{
		{
			name: "Vuls",
			args: args{
				result: &vulnerabilities.Results{
					MergedResults: &scanner.MergedResults{
						MergedVulnerabilitiesByKey: map[scanner.VulnerabilityKey][]scanner.MergedVulnerability{
							"vulkey1": []scanner.MergedVulnerability{
								{
									ID: "id1",
									Vulnerability: scanner.Vulnerability{
										ID:          "CVE-test-test-foo",
										Description: "testbleed",
									},
								},
							},
							"vulkey2": []scanner.MergedVulnerability{
								{
									ID: "id2",
									Vulnerability: scanner.Vulnerability{
										ID:          "CVE-test-test-bar",
										Description: "solartest",
									},
								},
							},
							"vulkey3": []scanner.MergedVulnerability{},
						},
					},
				},
			},
			want: returns{
				vulScan: &models.VulnerabilityScan{
					Vulnerabilities: &[]models.Vulnerability{
						models.Vulnerability{
							Id: utils.StringPtr("id1"),
							VulnerabilityInfo: &models.VulnerabilityInfo{
								Id:                utils.StringPtr("CVE-test-test-foo"),
								VulnerabilityName: utils.StringPtr("testbleed"),
							},
						},
						models.Vulnerability{
							Id: utils.StringPtr("id2"),
							VulnerabilityInfo: &models.VulnerabilityInfo{
								Id:                utils.StringPtr("CVE-test-test-bar"),
								VulnerabilityName: utils.StringPtr("solartest"),
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertVulnResultToApiModel(tt.args.result)
			if diff := cmp.Diff(tt.want.vulScan, got, cmpopts.SortSlices(func(a, b models.Vulnerability) bool { return *a.Id < *b.Id })); diff != "" {
				t.Errorf("convertVulnResultToApiModel() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
