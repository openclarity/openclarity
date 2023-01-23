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
	"github.com/openclarity/vmclarity/shared/pkg/families/secrets"
	"github.com/openclarity/vmclarity/shared/pkg/families/secrets/common"
	"github.com/openclarity/vmclarity/shared/pkg/families/vulnerabilities"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

func Test_convertSBOMResultToAPIModel(t *testing.T) {
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
							{
								BOMRef:     "bomref1",
								Type:       cdx.ComponentTypeLibrary,
								Name:       "testcomponent1",
								Version:    "v10.0.0-foo",
								PackageURL: "pkg:pypi/testcomponent1@v10.0.0-foo",
							},
							{
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
						{
							Id: utils.StringPtr("bomref1"),
							PackageInfo: &models.PackageInfo{
								PackageName:    utils.StringPtr("testcomponent1"),
								PackageVersion: utils.StringPtr("v10.0.0-foo"),
							},
						},
						{
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
			got := convertSBOMResultToAPIModel(tt.args.result)
			if diff := cmp.Diff(tt.want.sbomScan, got, cmpopts.SortSlices(func(a, b models.Package) bool { return *a.Id < *b.Id })); diff != "" {
				t.Errorf("convertSBOMResultToAPIModel() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_convertVulnResultToAPIModel(t *testing.T) {
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
							"vulkey1": {
								{
									ID: "id1",
									Vulnerability: scanner.Vulnerability{
										ID:          "CVE-test-test-foo",
										Description: "testbleed",
									},
								},
							},
							"vulkey2": {
								{
									ID: "id2",
									Vulnerability: scanner.Vulnerability{
										ID:          "CVE-test-test-bar",
										Description: "solartest",
									},
								},
							},
							"vulkey3": {},
						},
					},
				},
			},
			want: returns{
				vulScan: &models.VulnerabilityScan{
					Vulnerabilities: &[]models.Vulnerability{
						{
							Id: utils.StringPtr("id1"),
							VulnerabilityInfo: &models.VulnerabilityInfo{
								Id:                utils.StringPtr("CVE-test-test-foo"),
								VulnerabilityName: utils.StringPtr("testbleed"),
							},
						},
						{
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
			got := convertVulnResultToAPIModel(tt.args.result)
			if diff := cmp.Diff(tt.want.vulScan, got, cmpopts.SortSlices(func(a, b models.Vulnerability) bool { return *a.Id < *b.Id })); diff != "" {
				t.Errorf("convertVulnResultToAPIModel() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_convertSecretsResultToAPIModel(t *testing.T) {
	finding1 := common.Findings{
		Description: "Description1",
		StartLine:   1,
		EndLine:     11,
		File:        "File1",
		Fingerprint: "Fingerprint1",
	}
	finding2 := common.Findings{
		Description: "Description2",
		StartLine:   2,
		EndLine:     22,
		File:        "File2",
		Fingerprint: "Fingerprint2",
	}
	finding3 := common.Findings{
		Description: "Description3",
		StartLine:   3,
		EndLine:     33,
		File:        "File3",
		Fingerprint: "Fingerprint3",
	}
	type args struct {
		secretsResults *secrets.Results
	}
	tests := []struct {
		name string
		args args
		want *models.SecretScan
	}{
		{
			name: "nil secretsResults",
			args: args{
				secretsResults: nil,
			},
			want: &models.SecretScan{},
		},
		{
			name: "nil secretsResults.MergedResults",
			args: args{
				secretsResults: &secrets.Results{
					MergedResults: nil,
				},
			},
			want: &models.SecretScan{},
		},
		{
			name: "empty secretsResults.MergedResults.Results",
			args: args{
				secretsResults: &secrets.Results{
					MergedResults: &secrets.MergedResults{
						Results: nil,
					},
				},
			},
			want: &models.SecretScan{},
		},
		{
			name: "sanity",
			args: args{
				secretsResults: &secrets.Results{
					MergedResults: &secrets.MergedResults{
						Results: []*common.Results{
							{
								Findings: nil,
							},
							{
								Findings: []common.Findings{
									finding1,
									finding2,
								},
							},
							{
								Findings: []common.Findings{
									finding3,
								},
							},
						},
					},
				},
			},
			want: &models.SecretScan{
				Secrets: &[]models.Secret{
					{
						SecretInfo: &models.SecretInfo{
							Description: &finding1.Description,
							EndLine:     &finding1.EndLine,
							FilePath:    &finding1.File,
							Fingerprint: &finding1.Fingerprint,
							StartLine:   &finding1.StartLine,
						},
						Id: &finding1.Fingerprint,
					},
					{
						SecretInfo: &models.SecretInfo{
							Description: &finding2.Description,
							EndLine:     &finding2.EndLine,
							FilePath:    &finding2.File,
							Fingerprint: &finding2.Fingerprint,
							StartLine:   &finding2.StartLine,
						},
						Id: &finding2.Fingerprint,
					},
					{
						SecretInfo: &models.SecretInfo{
							Description: &finding3.Description,
							EndLine:     &finding3.EndLine,
							FilePath:    &finding3.File,
							Fingerprint: &finding3.Fingerprint,
							StartLine:   &finding3.StartLine,
						},
						Id: &finding3.Fingerprint,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertSecretsResultToAPIModel(tt.args.secretsResults)
			if diff := cmp.Diff(tt.want, got, cmpopts.SortSlices(func(a, b models.Secret) bool { return *a.Id < *b.Id })); diff != "" {
				t.Errorf("convertSBOMResultToAPIModel() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
