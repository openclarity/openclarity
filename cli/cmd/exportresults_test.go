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

package cmd

import (
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/openclarity/kubeclarity/shared/pkg/scanner"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/shared/pkg/families/exploits"
	common2 "github.com/openclarity/vmclarity/shared/pkg/families/exploits/common"
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
								Version:    "v10.0.0-foo1",
								PackageURL: "pkg:pypi/testcomponent1@v10.0.0-foo",
								CPE:        "cpe1",
								Licenses: utils.PointerTo[cdx.Licenses]([]cdx.LicenseChoice{
									{
										License: &cdx.License{
											Name: "lic1",
										},
									},
									{
										License: &cdx.License{
											Name: "lic2",
										},
									},
								}),
							},
							{
								BOMRef:     "bomref2",
								Type:       cdx.ComponentTypeLibrary,
								Name:       "testcomponent2",
								Version:    "v10.0.0-foo2",
								PackageURL: "pkg:pypi/testcomponent2@v10.0.0-foo",
								CPE:        "cpe2",
								Licenses: utils.PointerTo[cdx.Licenses]([]cdx.LicenseChoice{
									{
										License: &cdx.License{
											Name: "lic3",
										},
									},
									{
										License: &cdx.License{
											Name: "lic4",
										},
									},
								}),
							},
						},
					},
				},
			},
			want: returns{
				sbomScan: &models.SbomScan{
					Packages: &[]models.Package{
						{
							Cpes:     utils.PointerTo([]string{"cpe1"}),
							Language: utils.PointerTo("python"),
							Licenses: utils.PointerTo([]string{"lic1", "lic2"}),
							Name:     utils.PointerTo("testcomponent1"),
							Purl:     utils.PointerTo("pkg:pypi/testcomponent1@v10.0.0-foo"),
							Type:     utils.PointerTo(string(cdx.ComponentTypeLibrary)),
							Version:  utils.PointerTo("v10.0.0-foo1"),
						},
						{
							Cpes:     utils.PointerTo([]string{"cpe2"}),
							Language: utils.PointerTo("python"),
							Licenses: utils.PointerTo([]string{"lic3", "lic4"}),
							Name:     utils.PointerTo("testcomponent2"),
							Purl:     utils.PointerTo("pkg:pypi/testcomponent2@v10.0.0-foo"),
							Type:     utils.PointerTo(string(cdx.ComponentTypeLibrary)),
							Version:  utils.PointerTo("v10.0.0-foo2"),
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
			if diff := cmp.Diff(tt.want.sbomScan, got, cmpopts.SortSlices(func(a, b models.Package) bool { return *a.Purl < *b.Purl })); diff != "" {
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
										Links:       []string{"link1", "link2"},
										Distro: scanner.Distro{
											Name:    "distro1",
											Version: "distrov1",
											IDLike:  []string{"IDLike1", "IDLike2"},
										},
										CVSS: []scanner.CVSS{
											{
												Version: "v1",
												Vector:  "vector1",
												Metrics: scanner.CvssMetrics{
													BaseScore:           1,
													ExploitabilityScore: nil,
													ImpactScore:         nil,
												},
											},
											{
												Version: "v2",
												Vector:  "vector2",
												Metrics: scanner.CvssMetrics{
													BaseScore:           2,
													ExploitabilityScore: utils.PointerTo(2.1),
													ImpactScore:         utils.PointerTo(2.2),
												},
											},
										},
										Fix: scanner.Fix{
											Versions: []string{"fv1", "fv2"},
											State:    "fixed",
										},
										Severity: string(models.CRITICAL),
										Package: scanner.Package{
											Name:     "package1",
											Version:  "pv1",
											Type:     "pt1",
											Language: "pl1",
											Licenses: []string{"plic1", "plic2"},
											CPEs:     []string{"cpe1", "cpe2"},
											PURL:     "purl1",
										},
										LayerID: "lid1",
										Path:    "path1",
									},
								},
							},
							"vulkey2": {
								{
									ID: "id2",
									Vulnerability: scanner.Vulnerability{
										ID:          "CVE-test-test-bar",
										Description: "solartest",
										Links:       []string{"link3", "link4"},
										Distro: scanner.Distro{
											Name:    "distro2",
											Version: "distrov2",
											IDLike:  []string{"IDLike3", "IDLike4"},
										},
										CVSS: []scanner.CVSS{
											{
												Version: "v3",
												Vector:  "vector3",
												Metrics: scanner.CvssMetrics{
													BaseScore:           3,
													ExploitabilityScore: nil,
													ImpactScore:         nil,
												},
											},
											{
												Version: "v4",
												Vector:  "vector4",
												Metrics: scanner.CvssMetrics{
													BaseScore:           4,
													ExploitabilityScore: utils.PointerTo(4.1),
													ImpactScore:         utils.PointerTo(4.2),
												},
											},
										},
										Fix: scanner.Fix{
											Versions: []string{"fv3", "fv4"},
											State:    "not-fixed",
										},
										Severity: string(models.HIGH),
										Package: scanner.Package{
											Name:     "package2",
											Version:  "pv2",
											Type:     "pt2",
											Language: "pl2",
											Licenses: []string{"plic3", "plic4"},
											CPEs:     []string{"cpe3", "cpe4"},
											PURL:     "purl2",
										},
										LayerID: "lid2",
										Path:    "path2",
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
							Cvss: utils.PointerTo([]models.VulnerabilityCvss{
								{
									Metrics: &models.VulnerabilityCvssMetrics{
										BaseScore:           utils.PointerTo[float32](1),
										ExploitabilityScore: nil,
										ImpactScore:         nil,
									},
									Vector:  utils.PointerTo("vector1"),
									Version: utils.PointerTo("v1"),
								},
								{
									Metrics: &models.VulnerabilityCvssMetrics{
										BaseScore:           utils.PointerTo[float32](2),
										ExploitabilityScore: utils.PointerTo[float32](2.1),
										ImpactScore:         utils.PointerTo[float32](2.2),
									},
									Vector:  utils.PointerTo("vector2"),
									Version: utils.PointerTo("v2"),
								},
							}),
							Description: utils.PointerTo("testbleed"),
							Distro: &models.VulnerabilityDistro{
								IDLike:  utils.PointerTo([]string{"IDLike1", "IDLike2"}),
								Name:    utils.PointerTo("distro1"),
								Version: utils.PointerTo("distrov1"),
							},
							Fix: &models.VulnerabilityFix{
								State:    utils.PointerTo("fixed"),
								Versions: utils.PointerTo([]string{"fv1", "fv2"}),
							},
							LayerId: utils.PointerTo("lid1"),
							Links:   utils.PointerTo([]string{"link1", "link2"}),
							Package: &models.Package{
								Cpes:     utils.PointerTo([]string{"cpe1", "cpe2"}),
								Language: utils.PointerTo("pl1"),
								Licenses: utils.PointerTo([]string{"plic1", "plic2"}),
								Name:     utils.PointerTo("package1"),
								Purl:     utils.PointerTo("purl1"),
								Type:     utils.PointerTo("pt1"),
								Version:  utils.PointerTo("pv1"),
							},
							Path:              utils.PointerTo("path1"),
							Severity:          utils.PointerTo[models.VulnerabilitySeverity](models.CRITICAL),
							VulnerabilityName: utils.PointerTo("CVE-test-test-foo"),
						},
						{
							Cvss: utils.PointerTo([]models.VulnerabilityCvss{
								{
									Metrics: &models.VulnerabilityCvssMetrics{
										BaseScore:           utils.PointerTo[float32](3),
										ExploitabilityScore: nil,
										ImpactScore:         nil,
									},
									Vector:  utils.PointerTo("vector3"),
									Version: utils.PointerTo("v3"),
								},
								{
									Metrics: &models.VulnerabilityCvssMetrics{
										BaseScore:           utils.PointerTo[float32](4),
										ExploitabilityScore: utils.PointerTo[float32](4.1),
										ImpactScore:         utils.PointerTo[float32](4.2),
									},
									Vector:  utils.PointerTo("vector4"),
									Version: utils.PointerTo("v4"),
								},
							}),
							Description: utils.PointerTo("solartest"),
							Distro: &models.VulnerabilityDistro{
								IDLike:  utils.PointerTo([]string{"IDLike3", "IDLike4"}),
								Name:    utils.PointerTo("distro2"),
								Version: utils.PointerTo("distrov2"),
							},
							Fix: &models.VulnerabilityFix{
								State:    utils.PointerTo("not-fixed"),
								Versions: utils.PointerTo([]string{"fv3", "fv4"}),
							},
							LayerId: utils.PointerTo("lid2"),
							Links:   utils.PointerTo([]string{"link3", "link4"}),
							Package: &models.Package{
								Cpes:     utils.PointerTo([]string{"cpe3", "cpe4"}),
								Language: utils.PointerTo("pl2"),
								Licenses: utils.PointerTo([]string{"plic3", "plic4"}),
								Name:     utils.PointerTo("package2"),
								Purl:     utils.PointerTo("purl2"),
								Type:     utils.PointerTo("pt2"),
								Version:  utils.PointerTo("pv2"),
							},
							Path:              utils.PointerTo("path2"),
							Severity:          utils.PointerTo[models.VulnerabilitySeverity](models.HIGH),
							VulnerabilityName: utils.PointerTo("CVE-test-test-bar"),
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertVulnResultToAPIModel(tt.args.result)
			if diff := cmp.Diff(tt.want.vulScan, got, cmpopts.SortSlices(func(a, b models.Vulnerability) bool { return *a.VulnerabilityName < *b.VulnerabilityName })); diff != "" {
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
						Description: &finding1.Description,
						EndLine:     &finding1.EndLine,
						FilePath:    &finding1.File,
						Fingerprint: &finding1.Fingerprint,
						StartLine:   &finding1.StartLine,
					},
					{
						Description: &finding2.Description,
						EndLine:     &finding2.EndLine,
						FilePath:    &finding2.File,
						Fingerprint: &finding2.Fingerprint,
						StartLine:   &finding2.StartLine,
					},
					{
						Description: &finding3.Description,
						EndLine:     &finding3.EndLine,
						FilePath:    &finding3.File,
						Fingerprint: &finding3.Fingerprint,
						StartLine:   &finding3.StartLine,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertSecretsResultToAPIModel(tt.args.secretsResults)
			if diff := cmp.Diff(tt.want, got, cmpopts.SortSlices(func(a, b models.Secret) bool { return *a.Fingerprint < *b.Fingerprint })); diff != "" {
				t.Errorf("convertSBOMResultToAPIModel() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_convertExploitsResultToAPIModel(t *testing.T) {
	exploit1 := common2.Exploit{
		ID:          "id1",
		Name:        "name1",
		Title:       "title1",
		Description: "desc 1",
		CveID:       "cve1",
		URLs:        []string{"url1"},
		SourceDB:    "db1",
	}
	exploit2 := common2.Exploit{
		ID:          "id2",
		Name:        "name2",
		Title:       "title2",
		Description: "desc 2",
		CveID:       "cve2",
		URLs:        []string{"url2"},
		SourceDB:    "db2",
	}
	exploit3 := common2.Exploit{
		ID:          "id3",
		Name:        "name3",
		Title:       "title3",
		Description: "desc 3",
		CveID:       "cve3",
		URLs:        []string{"url3"},
		SourceDB:    "db3",
	}
	type args struct {
		exploitsResults *exploits.Results
	}
	tests := []struct {
		name string
		args args
		want *models.ExploitScan
	}{
		{
			name: "nil exploitsResults",
			args: args{
				exploitsResults: nil,
			},
			want: &models.ExploitScan{},
		},
		{
			name: "nil exploitsResults.Exploits",
			args: args{
				exploitsResults: &exploits.Results{
					Exploits: nil,
				},
			},
			want: &models.ExploitScan{},
		},
		{
			name: "sanity",
			args: args{
				exploitsResults: &exploits.Results{
					Exploits: exploits.MergedExploits{
						{
							Exploit: exploit1,
						},
						{
							Exploit: exploit2,
						},
						{
							Exploit: exploit3,
						},
					},
				},
			},
			want: &models.ExploitScan{
				Exploits: &[]models.Exploit{
					{
						CveID:       utils.PointerTo(exploit1.CveID),
						Description: utils.PointerTo(exploit1.Description),
						Name:        utils.PointerTo(exploit1.Name),
						SourceDB:    utils.PointerTo(exploit1.SourceDB),
						Title:       utils.PointerTo(exploit1.Title),
						Urls:        &exploit1.URLs,
					},
					{
						CveID:       utils.PointerTo(exploit2.CveID),
						Description: utils.PointerTo(exploit2.Description),
						Name:        utils.PointerTo(exploit2.Name),
						SourceDB:    utils.PointerTo(exploit2.SourceDB),
						Title:       utils.PointerTo(exploit2.Title),
						Urls:        &exploit2.URLs,
					},
					{
						CveID:       utils.PointerTo(exploit3.CveID),
						Description: utils.PointerTo(exploit3.Description),
						Name:        utils.PointerTo(exploit3.Name),
						SourceDB:    utils.PointerTo(exploit3.SourceDB),
						Title:       utils.PointerTo(exploit3.Title),
						Urls:        &exploit3.URLs,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertExploitsResultToAPIModel(tt.args.exploitsResults)
			if diff := cmp.Diff(tt.want, got, cmpopts.SortSlices(func(a, b models.Exploit) bool { return *a.CveID < *b.CveID })); diff != "" {
				t.Errorf("convertExploitsResultToAPIModel() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
