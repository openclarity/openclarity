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

package utils

import (
	"reflect"
	"testing"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/openclarity/kubeclarity/shared/pkg/scanner"
	"github.com/openclarity/kubeclarity/shared/pkg/utils/vulnerability"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/families/exploits"
	common2 "github.com/openclarity/vmclarity/pkg/shared/families/exploits/common"
	"github.com/openclarity/vmclarity/pkg/shared/families/infofinder"
	infofinderTypes "github.com/openclarity/vmclarity/pkg/shared/families/infofinder/types"
	"github.com/openclarity/vmclarity/pkg/shared/families/malware"
	malwarecommon "github.com/openclarity/vmclarity/pkg/shared/families/malware/common"
	"github.com/openclarity/vmclarity/pkg/shared/families/misconfiguration"
	misconfigurationTypes "github.com/openclarity/vmclarity/pkg/shared/families/misconfiguration/types"
	"github.com/openclarity/vmclarity/pkg/shared/families/sbom"
	"github.com/openclarity/vmclarity/pkg/shared/families/secrets"
	"github.com/openclarity/vmclarity/pkg/shared/families/secrets/common"
	"github.com/openclarity/vmclarity/pkg/shared/families/types"
	"github.com/openclarity/vmclarity/pkg/shared/families/vulnerabilities"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

func Test_ConvertSBOMResultToPackages(t *testing.T) {
	type args struct {
		result *sbom.Results
	}
	type returns struct {
		packages []models.Package
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
											ID: "lic1",
										},
									},
									{
										License: &cdx.License{
											ID: "lic2",
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
											ID: "lic3",
										},
									},
									{
										License: &cdx.License{
											ID: "lic4",
										},
									},
								}),
							},
						},
					},
				},
			},
			want: returns{
				packages: []models.Package{
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
				packages: []models.Package{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertSBOMResultToPackages(tt.args.result)
			if diff := cmp.Diff(tt.want.packages, got, cmpopts.SortSlices(func(a, b models.Package) bool { return *a.Purl < *b.Purl })); diff != "" {
				t.Errorf("convertSBOMResultToAPIModel() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_ConvertVulnResultToVulnerabilities(t *testing.T) {
	type args struct {
		result *vulnerabilities.Results
	}
	type returns struct {
		vulnerabilities []models.Vulnerability
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
				vulnerabilities: []models.Vulnerability{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertVulnResultToVulnerabilities(tt.args.result)
			if diff := cmp.Diff(tt.want.vulnerabilities, got, cmpopts.SortSlices(func(a, b models.Vulnerability) bool { return *a.VulnerabilityName < *b.VulnerabilityName })); diff != "" {
				t.Errorf("convertVulnResultToAPIModel() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_ConvertSecretsResultToAPIModel(t *testing.T) {
	finding1 := common.Findings{
		Description: "Description1",
		StartLine:   1,
		EndLine:     11,
		StartColumn: 101,
		EndColumn:   111,
		File:        "File1",
		Fingerprint: "Fingerprint1",
	}
	finding2 := common.Findings{
		Description: "Description2",
		StartLine:   2,
		EndLine:     22,
		StartColumn: 102,
		EndColumn:   122,
		File:        "File2",
		Fingerprint: "Fingerprint2",
	}
	finding3 := common.Findings{
		Description: "Description3",
		StartLine:   3,
		EndLine:     33,
		StartColumn: 103,
		EndColumn:   133,
		File:        "File3",
		Fingerprint: "Fingerprint3",
	}
	type args struct {
		secretsResults *secrets.Results
	}
	tests := []struct {
		name string
		args args
		want []models.Secret
	}{
		{
			name: "nil secretsResults",
			args: args{
				secretsResults: nil,
			},
			want: []models.Secret{},
		},
		{
			name: "nil secretsResults.MergedResults",
			args: args{
				secretsResults: &secrets.Results{
					MergedResults: nil,
				},
			},
			want: []models.Secret{},
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
			want: []models.Secret{},
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
			want: []models.Secret{
				{
					Description: &finding1.Description,
					EndLine:     &finding1.EndLine,
					FilePath:    &finding1.File,
					Fingerprint: &finding1.Fingerprint,
					StartLine:   &finding1.StartLine,
					StartColumn: &finding1.StartColumn,
					EndColumn:   &finding1.EndColumn,
				},
				{
					Description: &finding2.Description,
					EndLine:     &finding2.EndLine,
					FilePath:    &finding2.File,
					Fingerprint: &finding2.Fingerprint,
					StartLine:   &finding2.StartLine,
					StartColumn: &finding2.StartColumn,
					EndColumn:   &finding2.EndColumn,
				},
				{
					Description: &finding3.Description,
					EndLine:     &finding3.EndLine,
					FilePath:    &finding3.File,
					Fingerprint: &finding3.Fingerprint,
					StartLine:   &finding3.StartLine,
					StartColumn: &finding3.StartColumn,
					EndColumn:   &finding3.EndColumn,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertSecretsResultToSecrets(tt.args.secretsResults)
			if diff := cmp.Diff(tt.want, got, cmpopts.SortSlices(func(a, b models.Secret) bool { return *a.Fingerprint < *b.Fingerprint })); diff != "" {
				t.Errorf("convertSBOMResultToAPIModel() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_ConvertMalwareResultToMalwareAndMetadata(t *testing.T) {
	type args struct {
		mergedResults *malware.MergedResults
	}
	type returns struct {
		Malware  []models.Malware
		Metadata []models.ScannerMetadata
	}
	tests := []struct {
		name string
		args args
		want returns
	}{
		{
			name: "nil mergedResults",
			args: args{
				mergedResults: nil,
			},
			want: returns{
				Malware:  []models.Malware{},
				Metadata: []models.ScannerMetadata{},
			},
		},
		{
			name: "nil malwareResults.Malware",
			args: args{
				mergedResults: &malware.MergedResults{
					DetectedMalware: nil,
				},
			},
			want: returns{
				Malware:  []models.Malware{},
				Metadata: []models.ScannerMetadata{},
			},
		},
		{
			name: "sanity",
			args: args{
				mergedResults: &malware.MergedResults{
					DetectedMalware: []malwarecommon.DetectedMalware{
						{
							MalwareName: "Worm<3",
							MalwareType: "WORM",
							Path:        "/somepath/innocent.exe",
						},
						{
							MalwareName: "Trojan:)",
							MalwareType: "TROJAN",
							Path:        "/somepath/gift.jar",
						},
						{
							MalwareName: "Ransom!",
							MalwareType: "RANSOMWARE",
							Path:        "/somepath/givememoney.exe",
						},
					},
					ScansSummary: map[string]*malwarecommon.ScanSummary{
						"clam": {
							KnownViruses:       100,
							EngineVersion:      "1",
							ScannedDirectories: 10,
							ScannedFiles:       1000,
							InfectedFiles:      3,
							SuspectedFiles:     4,
							DataScanned:        "800 MB",
							DataRead:           "1.6 GB",
							TimeTaken:          "1000000 ms",
						},
					},
				},
			},
			want: returns{
				Malware: []models.Malware{
					{
						MalwareName: utils.PointerTo("Ransom!"),
						MalwareType: utils.PointerTo[models.MalwareType]("RANSOMWARE"),
						Path:        utils.PointerTo("/somepath/givememoney.exe"),
					},
					{
						MalwareName: utils.PointerTo("Trojan:)"),
						MalwareType: utils.PointerTo[models.MalwareType]("TROJAN"),
						Path:        utils.PointerTo("/somepath/gift.jar"),
					},
					{
						MalwareName: utils.PointerTo("Worm<3"),
						MalwareType: utils.PointerTo[models.MalwareType]("WORM"),
						Path:        utils.PointerTo("/somepath/innocent.exe"),
					},
				},
				Metadata: []models.ScannerMetadata{
					{
						ScannerName: utils.PointerTo("clam"),
						ScannerSummary: &models.ScannerSummary{
							KnownViruses:       utils.PointerTo(100),
							EngineVersion:      utils.PointerTo("1"),
							ScannedDirectories: utils.PointerTo(10),
							ScannedFiles:       utils.PointerTo(1000),
							InfectedFiles:      utils.PointerTo(3),
							SuspectedFiles:     utils.PointerTo(4),
							DataScanned:        utils.PointerTo("800 MB"),
							DataRead:           utils.PointerTo("1.6 GB"),
							TimeTaken:          utils.PointerTo("1000000 ms"),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mware, mdata := ConvertMalwareResultToMalwareAndMetadata(tt.args.mergedResults)
			if diff := cmp.Diff(tt.want, returns{Malware: mware, Metadata: mdata}, cmpopts.SortSlices(func(a, b models.Malware) bool { return *a.MalwareType < *b.MalwareType })); diff != "" {
				t.Errorf("convertMalwareResultToAPIModel() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_ConvertExploitsResultToExploits(t *testing.T) {
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
		want []models.Exploit
	}{
		{
			name: "nil exploitsResults",
			args: args{
				exploitsResults: nil,
			},
			want: []models.Exploit{},
		},
		{
			name: "nil exploitsResults.Exploits",
			args: args{
				exploitsResults: &exploits.Results{
					Exploits: nil,
				},
			},
			want: []models.Exploit{},
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
			want: []models.Exploit{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertExploitsResultToExploits(tt.args.exploitsResults)
			if diff := cmp.Diff(tt.want, got, cmpopts.SortSlices(func(a, b models.Exploit) bool { return *a.CveID < *b.CveID })); diff != "" {
				t.Errorf("convertExploitsResultToAPIModel() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_MisconfigurationSeverityToAPIMisconfigurationSeverity(t *testing.T) {
	type args struct {
		sev misconfigurationTypes.Severity
	}
	tests := []struct {
		name    string
		args    args
		want    models.MisconfigurationSeverity
		wantErr bool
	}{
		{
			name: "high severity",
			args: args{
				sev: misconfigurationTypes.HighSeverity,
			},
			want: models.MisconfigurationHighSeverity,
		},
		{
			name: "medium severity",
			args: args{
				sev: misconfigurationTypes.MediumSeverity,
			},
			want: models.MisconfigurationMediumSeverity,
		},
		{
			name: "low severity",
			args: args{
				sev: misconfigurationTypes.LowSeverity,
			},
			want: models.MisconfigurationLowSeverity,
		},
		{
			name: "unknown severity",
			args: args{
				sev: misconfigurationTypes.Severity("doesn't exist"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MisconfigurationSeverityToAPIMisconfigurationSeverity(tt.args.sev)
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("Unexpected error: %v", err)
				}
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("misconfigurationSeverityToAPIMisconfigurationSeverity() mismatch (-want +got)\n%s", diff)
			}
		})
	}
}

func Test_ConvertMisconfigurationResultToMisconfigurations(t *testing.T) {
	misconfiguration1 := misconfiguration.FlattenedMisconfiguration{
		ScannerName: "foo",
		Misconfiguration: misconfigurationTypes.Misconfiguration{
			ScannedPath: "/scanned/path",

			TestCategory:    "category1",
			TestID:          "testid1",
			TestDescription: "Test description 1",

			Severity:    misconfigurationTypes.HighSeverity,
			Message:     "You got a problem with 1",
			Remediation: "Fix your stuff",
		},
	}

	misconfiguration2 := misconfiguration.FlattenedMisconfiguration{
		ScannerName: "foo",
		Misconfiguration: misconfigurationTypes.Misconfiguration{
			ScannedPath: "/scanned/path",

			TestCategory:    "category2",
			TestID:          "testid2",
			TestDescription: "Test description 2",

			Severity:    misconfigurationTypes.MediumSeverity,
			Message:     "You got a problem",
			Remediation: "Fix your stuff",
		},
	}

	misconfiguration3 := misconfiguration.FlattenedMisconfiguration{
		ScannerName: "bar",
		Misconfiguration: misconfigurationTypes.Misconfiguration{
			ScannedPath: "/scanned/path",

			TestCategory:    "category1",
			TestID:          "testid3",
			TestDescription: "Test description 1",

			Severity:    misconfigurationTypes.HighSeverity,
			Message:     "You got a problem with 1",
			Remediation: "Fix your stuff",
		},
	}

	timestamp := time.Now()

	type args struct {
		misconfigurationResults *misconfiguration.Results
	}
	type returns struct {
		Misconfigs []models.Misconfiguration
		Scanners   []string
	}
	tests := []struct {
		name string
		args args
		want returns
	}{
		{
			name: "nil misconfigurationResults",
			args: args{
				misconfigurationResults: nil,
			},
			want: returns{
				[]models.Misconfiguration{},
				[]string{},
			},
		},
		{
			name: "nil misconfigurationResults.Misconfigurations",
			args: args{
				misconfigurationResults: &misconfiguration.Results{
					Metadata: types.Metadata{
						Timestamp: timestamp,
						Scanners:  []string{"foo", "bar"},
					},
					Misconfigurations: nil,
				},
			},
			want: returns{
				[]models.Misconfiguration{},
				[]string{},
			},
		},
		{
			name: "sanity",
			args: args{
				misconfigurationResults: &misconfiguration.Results{
					Metadata: types.Metadata{
						Timestamp: timestamp,
						Scanners:  []string{"foo", "bar"},
					},
					Misconfigurations: []misconfiguration.FlattenedMisconfiguration{
						misconfiguration1,
						misconfiguration2,
						misconfiguration3,
					},
				},
			},
			want: returns{
				[]models.Misconfiguration{
					{
						Message:         utils.PointerTo(misconfiguration1.Message),
						Remediation:     utils.PointerTo(misconfiguration1.Remediation),
						ScannedPath:     utils.PointerTo(misconfiguration1.ScannedPath),
						ScannerName:     utils.PointerTo(misconfiguration1.ScannerName),
						Severity:        utils.PointerTo(models.MisconfigurationHighSeverity),
						TestCategory:    utils.PointerTo(misconfiguration1.TestCategory),
						TestDescription: utils.PointerTo(misconfiguration1.TestDescription),
						TestID:          utils.PointerTo(misconfiguration1.TestID),
					},
					{
						Message:         utils.PointerTo(misconfiguration2.Message),
						Remediation:     utils.PointerTo(misconfiguration2.Remediation),
						ScannedPath:     utils.PointerTo(misconfiguration2.ScannedPath),
						ScannerName:     utils.PointerTo(misconfiguration2.ScannerName),
						Severity:        utils.PointerTo(models.MisconfigurationMediumSeverity),
						TestCategory:    utils.PointerTo(misconfiguration2.TestCategory),
						TestDescription: utils.PointerTo(misconfiguration2.TestDescription),
						TestID:          utils.PointerTo(misconfiguration2.TestID),
					},
					{
						Message:         utils.PointerTo(misconfiguration3.Message),
						Remediation:     utils.PointerTo(misconfiguration3.Remediation),
						ScannedPath:     utils.PointerTo(misconfiguration3.ScannedPath),
						ScannerName:     utils.PointerTo(misconfiguration3.ScannerName),
						Severity:        utils.PointerTo(models.MisconfigurationHighSeverity),
						TestCategory:    utils.PointerTo(misconfiguration3.TestCategory),
						TestDescription: utils.PointerTo(misconfiguration3.TestDescription),
						TestID:          utils.PointerTo(misconfiguration3.TestID),
					},
				},
				[]string{"foo", "bar"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			misconfigs, scanners, err := ConvertMisconfigurationResultToMisconfigurationsAndScanners(tt.args.misconfigurationResults)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.want, returns{Misconfigs: misconfigs, Scanners: scanners}, cmpopts.SortSlices(func(a, b models.Misconfiguration) bool { return *a.TestID < *b.TestID })); diff != "" {
				t.Errorf("convertMisconfigurationResultToAPIModel() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_ConvertVulnSeverityToAPIModel(t *testing.T) {
	type args struct {
		severity string
	}
	tests := []struct {
		name string
		args args
		want *models.VulnerabilitySeverity
	}{
		{
			name: "DEFCON1 -> CRITICAL",
			args: args{
				severity: vulnerability.DEFCON1,
			},
			want: utils.PointerTo(models.CRITICAL),
		},
		{
			name: "CRITICAL -> CRITICAL",
			args: args{
				severity: vulnerability.CRITICAL,
			},
			want: utils.PointerTo(models.CRITICAL),
		},
		{
			name: "HIGH -> HIGH",
			args: args{
				severity: vulnerability.HIGH,
			},
			want: utils.PointerTo(models.HIGH),
		},
		{
			name: "MEDIUM -> MEDIUM",
			args: args{
				severity: vulnerability.MEDIUM,
			},
			want: utils.PointerTo(models.MEDIUM),
		},
		{
			name: "LOW -> LOW",
			args: args{
				severity: vulnerability.LOW,
			},
			want: utils.PointerTo(models.LOW),
		},
		{
			name: "NEGLIGIBLE -> NEGLIGIBLE",
			args: args{
				severity: vulnerability.NEGLIGIBLE,
			},
			want: utils.PointerTo(models.NEGLIGIBLE),
		},
		{
			name: "UNKNOWN -> NEGLIGIBLE",
			args: args{
				severity: vulnerability.UNKNOWN,
			},
			want: utils.PointerTo(models.NEGLIGIBLE),
		},
		{
			name: "NONE -> NEGLIGIBLE",
			args: args{
				severity: vulnerability.NONE,
			},
			want: utils.PointerTo(models.NEGLIGIBLE),
		},
		{
			name: "invalid -> NEGLIGIBLE",
			args: args{
				severity: "catastrophic",
			},
			want: utils.PointerTo(models.NEGLIGIBLE),
		},
		{
			name: "high -> HIGH",
			args: args{
				severity: "high",
			},
			want: utils.PointerTo(models.HIGH),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertVulnSeverityToAPIModel(tt.args.severity); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertVulnSeverityToAPIModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertInfoFinderResultToInfosAndScanners(t *testing.T) {
	type args struct {
		results *infofinder.Results
	}
	type returns struct {
		Infos    []models.InfoFinderInfo
		Scanners []string
	}
	tests := []struct {
		name    string
		args    args
		want    returns
		wantErr bool
	}{
		{
			name: "nil results",
			args: args{
				results: nil,
			},
			want: returns{
				Infos:    []models.InfoFinderInfo{},
				Scanners: []string{},
			},
			wantErr: false,
		},
		{
			name: "nil results.Info",
			args: args{
				results: &infofinder.Results{
					Infos: nil,
				},
			},
			want: returns{
				Infos:    []models.InfoFinderInfo{},
				Scanners: []string{},
			},
			wantErr: false,
		},
		{
			name: "sanity",
			args: args{
				results: &infofinder.Results{
					Metadata: types.Metadata{
						Scanners: []string{"scanner1", "scanner2"},
					},
					Infos: []infofinder.FlattenedInfos{
						{
							ScannerName: "scanner1",
							Info: infofinderTypes.Info{
								Type: infofinderTypes.SSHKnownHostFingerprint,
								Path: "Path1",
								Data: "Data1",
							},
						},
						{
							ScannerName: "scanner2",
							Info: infofinderTypes.Info{
								Type: infofinderTypes.SSHDaemonKeyFingerprint,
								Path: "Path2",
								Data: "Data2",
							},
						},
						{
							ScannerName: "scanner2",
							Info: infofinderTypes.Info{
								Type: infofinderTypes.SSHAuthorizedKeyFingerprint,
								Path: "Path3",
								Data: "Data3",
							},
						},
					},
				},
			},
			want: returns{
				Infos: []models.InfoFinderInfo{
					{
						Type:        utils.PointerTo(models.InfoTypeSSHKnownHostFingerprint),
						Path:        utils.PointerTo("Path1"),
						Data:        utils.PointerTo("Data1"),
						ScannerName: utils.PointerTo("scanner1"),
					},
					{
						Type:        utils.PointerTo(models.InfoTypeSSHDaemonKeyFingerprint),
						Path:        utils.PointerTo("Path2"),
						Data:        utils.PointerTo("Data2"),
						ScannerName: utils.PointerTo("scanner2"),
					},
					{
						Type:        utils.PointerTo(models.InfoTypeSSHAuthorizedKeyFingerprint),
						Path:        utils.PointerTo("Path3"),
						Data:        utils.PointerTo("Data3"),
						ScannerName: utils.PointerTo("scanner2"),
					},
				},
				Scanners: []string{"scanner1", "scanner2"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			infos, scanners, err := ConvertInfoFinderResultToInfosAndScanners(tt.args.results)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertInfoFinderResultToAPIModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(returns{
				Infos:    infos,
				Scanners: scanners,
			}, tt.want) {
				t.Errorf("ConvertInfoFinderResultToAPIModel() infos = %v, scanners = %v, want %v", infos, scanners, tt.want)
			}
		})
	}
}

func Test_convertInfoTypeToAPIModel(t *testing.T) {
	type args struct {
		infoType infofinderTypes.InfoType
	}
	tests := []struct {
		name string
		args args
		want *models.InfoType
	}{
		{
			name: "SSHKnownHostFingerprint",
			args: args{
				infoType: infofinderTypes.SSHKnownHostFingerprint,
			},
			want: utils.PointerTo(models.InfoTypeSSHKnownHostFingerprint),
		},
		{
			name: "SSHAuthorizedKeyFingerprint",
			args: args{
				infoType: infofinderTypes.SSHAuthorizedKeyFingerprint,
			},
			want: utils.PointerTo(models.InfoTypeSSHAuthorizedKeyFingerprint),
		},
		{
			name: "SSHPrivateKeyFingerprint",
			args: args{
				infoType: infofinderTypes.SSHPrivateKeyFingerprint,
			},
			want: utils.PointerTo(models.InfoTypeSSHPrivateKeyFingerprint),
		},
		{
			name: "SSHDaemonKeyFingerprint",
			args: args{
				infoType: infofinderTypes.SSHDaemonKeyFingerprint,
			},
			want: utils.PointerTo(models.InfoTypeSSHDaemonKeyFingerprint),
		},
		{
			name: "unknown",
			args: args{
				infoType: "unknown",
			},
			want: utils.PointerTo(models.InfoTypeUNKNOWN),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertInfoTypeToAPIModel(tt.args.infoType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertInfoTypeToAPIModel() = %v, want %v", got, tt.want)
			}
		})
	}
}
