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

package presenter

import (
	"reflect"
	"testing"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/scanner/families/exploits"
	common2 "github.com/openclarity/vmclarity/scanner/families/exploits/common"
	"github.com/openclarity/vmclarity/scanner/families/infofinder"
	infofinderTypes "github.com/openclarity/vmclarity/scanner/families/infofinder/types"
	"github.com/openclarity/vmclarity/scanner/families/malware"
	malwarecommon "github.com/openclarity/vmclarity/scanner/families/malware/common"
	"github.com/openclarity/vmclarity/scanner/families/misconfiguration"
	misconfigurationTypes "github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
	"github.com/openclarity/vmclarity/scanner/families/sbom"
	"github.com/openclarity/vmclarity/scanner/families/secrets"
	"github.com/openclarity/vmclarity/scanner/families/secrets/common"
	"github.com/openclarity/vmclarity/scanner/families/types"
	"github.com/openclarity/vmclarity/scanner/families/vulnerabilities"
	"github.com/openclarity/vmclarity/scanner/scanner"
	"github.com/openclarity/vmclarity/scanner/utils/vulnerability"
)

func Test_ConvertSBOMResultToPackages(t *testing.T) {
	type args struct {
		result *sbom.Results
	}
	type returns struct {
		packages []apitypes.Package
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
								Licenses: to.Ptr[cdx.Licenses]([]cdx.LicenseChoice{
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
								Licenses: to.Ptr[cdx.Licenses]([]cdx.LicenseChoice{
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
				packages: []apitypes.Package{
					{
						Cpes:     to.Ptr([]string{"cpe1"}),
						Language: to.Ptr("python"),
						Licenses: to.Ptr([]string{"lic1", "lic2"}),
						Name:     to.Ptr("testcomponent1"),
						Purl:     to.Ptr("pkg:pypi/testcomponent1@v10.0.0-foo"),
						Type:     to.Ptr(string(cdx.ComponentTypeLibrary)),
						Version:  to.Ptr("v10.0.0-foo1"),
					},
					{
						Cpes:     to.Ptr([]string{"cpe2"}),
						Language: to.Ptr("python"),
						Licenses: to.Ptr([]string{"lic3", "lic4"}),
						Name:     to.Ptr("testcomponent2"),
						Purl:     to.Ptr("pkg:pypi/testcomponent2@v10.0.0-foo"),
						Type:     to.Ptr(string(cdx.ComponentTypeLibrary)),
						Version:  to.Ptr("v10.0.0-foo2"),
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
				packages: []apitypes.Package{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertSBOMResultToPackages(tt.args.result)
			if diff := cmp.Diff(tt.want.packages, got, cmpopts.SortSlices(func(a, b apitypes.Package) bool { return *a.Purl < *b.Purl })); diff != "" {
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
		vulnerabilities []apitypes.Vulnerability
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
													ExploitabilityScore: to.Ptr(2.1),
													ImpactScore:         to.Ptr(2.2),
												},
											},
										},
										Fix: scanner.Fix{
											Versions: []string{"fv1", "fv2"},
											State:    "fixed",
										},
										Severity: string(apitypes.CRITICAL),
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
													ExploitabilityScore: to.Ptr(4.1),
													ImpactScore:         to.Ptr(4.2),
												},
											},
										},
										Fix: scanner.Fix{
											Versions: []string{"fv3", "fv4"},
											State:    "not-fixed",
										},
										Severity: string(apitypes.HIGH),
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
				vulnerabilities: []apitypes.Vulnerability{
					{
						Cvss: to.Ptr([]apitypes.VulnerabilityCvss{
							{
								Metrics: &apitypes.VulnerabilityCvssMetrics{
									BaseScore:           to.Ptr[float32](1),
									ExploitabilityScore: nil,
									ImpactScore:         nil,
								},
								Vector:  to.Ptr("vector1"),
								Version: to.Ptr("v1"),
							},
							{
								Metrics: &apitypes.VulnerabilityCvssMetrics{
									BaseScore:           to.Ptr[float32](2),
									ExploitabilityScore: to.Ptr[float32](2.1),
									ImpactScore:         to.Ptr[float32](2.2),
								},
								Vector:  to.Ptr("vector2"),
								Version: to.Ptr("v2"),
							},
						}),
						Description: to.Ptr("testbleed"),
						Distro: &apitypes.VulnerabilityDistro{
							IDLike:  to.Ptr([]string{"IDLike1", "IDLike2"}),
							Name:    to.Ptr("distro1"),
							Version: to.Ptr("distrov1"),
						},
						Fix: &apitypes.VulnerabilityFix{
							State:    to.Ptr("fixed"),
							Versions: to.Ptr([]string{"fv1", "fv2"}),
						},
						LayerId: to.Ptr("lid1"),
						Links:   to.Ptr([]string{"link1", "link2"}),
						Package: &apitypes.Package{
							Cpes:     to.Ptr([]string{"cpe1", "cpe2"}),
							Language: to.Ptr("pl1"),
							Licenses: to.Ptr([]string{"plic1", "plic2"}),
							Name:     to.Ptr("package1"),
							Purl:     to.Ptr("purl1"),
							Type:     to.Ptr("pt1"),
							Version:  to.Ptr("pv1"),
						},
						Path:              to.Ptr("path1"),
						Severity:          to.Ptr[apitypes.VulnerabilitySeverity](apitypes.CRITICAL),
						VulnerabilityName: to.Ptr("CVE-test-test-foo"),
					},
					{
						Cvss: to.Ptr([]apitypes.VulnerabilityCvss{
							{
								Metrics: &apitypes.VulnerabilityCvssMetrics{
									BaseScore:           to.Ptr[float32](3),
									ExploitabilityScore: nil,
									ImpactScore:         nil,
								},
								Vector:  to.Ptr("vector3"),
								Version: to.Ptr("v3"),
							},
							{
								Metrics: &apitypes.VulnerabilityCvssMetrics{
									BaseScore:           to.Ptr[float32](4),
									ExploitabilityScore: to.Ptr[float32](4.1),
									ImpactScore:         to.Ptr[float32](4.2),
								},
								Vector:  to.Ptr("vector4"),
								Version: to.Ptr("v4"),
							},
						}),
						Description: to.Ptr("solartest"),
						Distro: &apitypes.VulnerabilityDistro{
							IDLike:  to.Ptr([]string{"IDLike3", "IDLike4"}),
							Name:    to.Ptr("distro2"),
							Version: to.Ptr("distrov2"),
						},
						Fix: &apitypes.VulnerabilityFix{
							State:    to.Ptr("not-fixed"),
							Versions: to.Ptr([]string{"fv3", "fv4"}),
						},
						LayerId: to.Ptr("lid2"),
						Links:   to.Ptr([]string{"link3", "link4"}),
						Package: &apitypes.Package{
							Cpes:     to.Ptr([]string{"cpe3", "cpe4"}),
							Language: to.Ptr("pl2"),
							Licenses: to.Ptr([]string{"plic3", "plic4"}),
							Name:     to.Ptr("package2"),
							Purl:     to.Ptr("purl2"),
							Type:     to.Ptr("pt2"),
							Version:  to.Ptr("pv2"),
						},
						Path:              to.Ptr("path2"),
						Severity:          to.Ptr[apitypes.VulnerabilitySeverity](apitypes.HIGH),
						VulnerabilityName: to.Ptr("CVE-test-test-bar"),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertVulnResultToVulnerabilities(tt.args.result)
			if diff := cmp.Diff(tt.want.vulnerabilities, got, cmpopts.SortSlices(func(a, b apitypes.Vulnerability) bool { return *a.VulnerabilityName < *b.VulnerabilityName })); diff != "" {
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
		want []apitypes.Secret
	}{
		{
			name: "nil secretsResults",
			args: args{
				secretsResults: nil,
			},
			want: []apitypes.Secret{},
		},
		{
			name: "nil secretsResults.MergedResults",
			args: args{
				secretsResults: &secrets.Results{
					MergedResults: nil,
				},
			},
			want: []apitypes.Secret{},
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
			want: []apitypes.Secret{},
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
			want: []apitypes.Secret{
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
			if diff := cmp.Diff(tt.want, got, cmpopts.SortSlices(func(a, b apitypes.Secret) bool { return *a.Fingerprint < *b.Fingerprint })); diff != "" {
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
		Malware  []apitypes.Malware
		Metadata []apitypes.ScannerMetadata
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
				Malware:  []apitypes.Malware{},
				Metadata: []apitypes.ScannerMetadata{},
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
				Malware:  []apitypes.Malware{},
				Metadata: []apitypes.ScannerMetadata{},
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
				Malware: []apitypes.Malware{
					{
						MalwareName: to.Ptr("Ransom!"),
						MalwareType: to.Ptr[apitypes.MalwareType]("RANSOMWARE"),
						Path:        to.Ptr("/somepath/givememoney.exe"),
					},
					{
						MalwareName: to.Ptr("Trojan:)"),
						MalwareType: to.Ptr[apitypes.MalwareType]("TROJAN"),
						Path:        to.Ptr("/somepath/gift.jar"),
					},
					{
						MalwareName: to.Ptr("Worm<3"),
						MalwareType: to.Ptr[apitypes.MalwareType]("WORM"),
						Path:        to.Ptr("/somepath/innocent.exe"),
					},
				},
				Metadata: []apitypes.ScannerMetadata{
					{
						ScannerName: to.Ptr("clam"),
						ScannerSummary: &apitypes.ScannerSummary{
							KnownViruses:       to.Ptr(100),
							EngineVersion:      to.Ptr("1"),
							ScannedDirectories: to.Ptr(10),
							ScannedFiles:       to.Ptr(1000),
							InfectedFiles:      to.Ptr(3),
							SuspectedFiles:     to.Ptr(4),
							DataScanned:        to.Ptr("800 MB"),
							DataRead:           to.Ptr("1.6 GB"),
							TimeTaken:          to.Ptr("1000000 ms"),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mware, mdata := ConvertMalwareResultToMalwareAndMetadata(tt.args.mergedResults)
			if diff := cmp.Diff(tt.want, returns{Malware: mware, Metadata: mdata}, cmpopts.SortSlices(func(a, b apitypes.Malware) bool { return *a.MalwareType < *b.MalwareType })); diff != "" {
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
		want []apitypes.Exploit
	}{
		{
			name: "nil exploitsResults",
			args: args{
				exploitsResults: nil,
			},
			want: []apitypes.Exploit{},
		},
		{
			name: "nil exploitsResults.Exploits",
			args: args{
				exploitsResults: &exploits.Results{
					Exploits: nil,
				},
			},
			want: []apitypes.Exploit{},
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
			want: []apitypes.Exploit{
				{
					CveID:       to.Ptr(exploit1.CveID),
					Description: to.Ptr(exploit1.Description),
					Name:        to.Ptr(exploit1.Name),
					SourceDB:    to.Ptr(exploit1.SourceDB),
					Title:       to.Ptr(exploit1.Title),
					Urls:        &exploit1.URLs,
				},
				{
					CveID:       to.Ptr(exploit2.CveID),
					Description: to.Ptr(exploit2.Description),
					Name:        to.Ptr(exploit2.Name),
					SourceDB:    to.Ptr(exploit2.SourceDB),
					Title:       to.Ptr(exploit2.Title),
					Urls:        &exploit2.URLs,
				},
				{
					CveID:       to.Ptr(exploit3.CveID),
					Description: to.Ptr(exploit3.Description),
					Name:        to.Ptr(exploit3.Name),
					SourceDB:    to.Ptr(exploit3.SourceDB),
					Title:       to.Ptr(exploit3.Title),
					Urls:        &exploit3.URLs,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertExploitsResultToExploits(tt.args.exploitsResults)
			if diff := cmp.Diff(tt.want, got, cmpopts.SortSlices(func(a, b apitypes.Exploit) bool { return *a.CveID < *b.CveID })); diff != "" {
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
		want    apitypes.MisconfigurationSeverity
		wantErr bool
	}{
		{
			name: "high severity",
			args: args{
				sev: misconfigurationTypes.HighSeverity,
			},
			want: apitypes.MisconfigurationHighSeverity,
		},
		{
			name: "medium severity",
			args: args{
				sev: misconfigurationTypes.MediumSeverity,
			},
			want: apitypes.MisconfigurationMediumSeverity,
		},
		{
			name: "low severity",
			args: args{
				sev: misconfigurationTypes.LowSeverity,
			},
			want: apitypes.MisconfigurationLowSeverity,
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
			Location: "/scanned/path",

			Category:    "category1",
			ID:          "id1",
			Description: "Test description 1",

			Severity:    misconfigurationTypes.HighSeverity,
			Message:     "You got a problem with 1",
			Remediation: "Fix your stuff",
		},
	}

	misconfiguration2 := misconfiguration.FlattenedMisconfiguration{
		ScannerName: "foo",
		Misconfiguration: misconfigurationTypes.Misconfiguration{
			Location: "/scanned/path",

			Category:    "category2",
			ID:          "id2",
			Description: "Test description 2",

			Severity:    misconfigurationTypes.MediumSeverity,
			Message:     "You got a problem",
			Remediation: "Fix your stuff",
		},
	}

	misconfiguration3 := misconfiguration.FlattenedMisconfiguration{
		ScannerName: "bar",
		Misconfiguration: misconfigurationTypes.Misconfiguration{
			Location: "/scanned/path",

			Category:    "category1",
			ID:          "id3",
			Description: "Test description 1",

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
		Misconfigs []apitypes.Misconfiguration
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
				[]apitypes.Misconfiguration{},
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
				[]apitypes.Misconfiguration{},
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
				[]apitypes.Misconfiguration{
					{
						Message:     to.Ptr(misconfiguration1.Message),
						Remediation: to.Ptr(misconfiguration1.Remediation),
						Location:    to.Ptr(misconfiguration1.Location),
						ScannerName: to.Ptr(misconfiguration1.ScannerName),
						Severity:    to.Ptr(apitypes.MisconfigurationHighSeverity),
						Category:    to.Ptr(misconfiguration1.Category),
						Description: to.Ptr(misconfiguration1.Description),
						Id:          to.Ptr(misconfiguration1.ID),
					},
					{
						Message:     to.Ptr(misconfiguration2.Message),
						Remediation: to.Ptr(misconfiguration2.Remediation),
						Location:    to.Ptr(misconfiguration2.Location),
						ScannerName: to.Ptr(misconfiguration2.ScannerName),
						Severity:    to.Ptr(apitypes.MisconfigurationMediumSeverity),
						Category:    to.Ptr(misconfiguration2.Category),
						Description: to.Ptr(misconfiguration2.Description),
						Id:          to.Ptr(misconfiguration2.ID),
					},
					{
						Message:     to.Ptr(misconfiguration3.Message),
						Remediation: to.Ptr(misconfiguration3.Remediation),
						Location:    to.Ptr(misconfiguration3.Location),
						ScannerName: to.Ptr(misconfiguration3.ScannerName),
						Severity:    to.Ptr(apitypes.MisconfigurationHighSeverity),
						Category:    to.Ptr(misconfiguration3.Category),
						Description: to.Ptr(misconfiguration3.Description),
						Id:          to.Ptr(misconfiguration3.ID),
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

			if diff := cmp.Diff(tt.want, returns{Misconfigs: misconfigs, Scanners: scanners}, cmpopts.SortSlices(func(a, b apitypes.Misconfiguration) bool { return *a.Id < *b.Id })); diff != "" {
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
		want *apitypes.VulnerabilitySeverity
	}{
		{
			name: "DEFCON1 -> CRITICAL",
			args: args{
				severity: vulnerability.DEFCON1,
			},
			want: to.Ptr(apitypes.CRITICAL),
		},
		{
			name: "CRITICAL -> CRITICAL",
			args: args{
				severity: vulnerability.CRITICAL,
			},
			want: to.Ptr(apitypes.CRITICAL),
		},
		{
			name: "HIGH -> HIGH",
			args: args{
				severity: vulnerability.HIGH,
			},
			want: to.Ptr(apitypes.HIGH),
		},
		{
			name: "MEDIUM -> MEDIUM",
			args: args{
				severity: vulnerability.MEDIUM,
			},
			want: to.Ptr(apitypes.MEDIUM),
		},
		{
			name: "LOW -> LOW",
			args: args{
				severity: vulnerability.LOW,
			},
			want: to.Ptr(apitypes.LOW),
		},
		{
			name: "NEGLIGIBLE -> NEGLIGIBLE",
			args: args{
				severity: vulnerability.NEGLIGIBLE,
			},
			want: to.Ptr(apitypes.NEGLIGIBLE),
		},
		{
			name: "UNKNOWN -> NEGLIGIBLE",
			args: args{
				severity: vulnerability.UNKNOWN,
			},
			want: to.Ptr(apitypes.NEGLIGIBLE),
		},
		{
			name: "NONE -> NEGLIGIBLE",
			args: args{
				severity: vulnerability.NONE,
			},
			want: to.Ptr(apitypes.NEGLIGIBLE),
		},
		{
			name: "invalid -> NEGLIGIBLE",
			args: args{
				severity: "catastrophic",
			},
			want: to.Ptr(apitypes.NEGLIGIBLE),
		},
		{
			name: "high -> HIGH",
			args: args{
				severity: "high",
			},
			want: to.Ptr(apitypes.HIGH),
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
		Infos    []apitypes.InfoFinderInfo
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
				Infos:    []apitypes.InfoFinderInfo{},
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
				Infos:    []apitypes.InfoFinderInfo{},
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
				Infos: []apitypes.InfoFinderInfo{
					{
						Type:        to.Ptr(apitypes.InfoTypeSSHKnownHostFingerprint),
						Path:        to.Ptr("Path1"),
						Data:        to.Ptr("Data1"),
						ScannerName: to.Ptr("scanner1"),
					},
					{
						Type:        to.Ptr(apitypes.InfoTypeSSHDaemonKeyFingerprint),
						Path:        to.Ptr("Path2"),
						Data:        to.Ptr("Data2"),
						ScannerName: to.Ptr("scanner2"),
					},
					{
						Type:        to.Ptr(apitypes.InfoTypeSSHAuthorizedKeyFingerprint),
						Path:        to.Ptr("Path3"),
						Data:        to.Ptr("Data3"),
						ScannerName: to.Ptr("scanner2"),
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
		want *apitypes.InfoType
	}{
		{
			name: "SSHKnownHostFingerprint",
			args: args{
				infoType: infofinderTypes.SSHKnownHostFingerprint,
			},
			want: to.Ptr(apitypes.InfoTypeSSHKnownHostFingerprint),
		},
		{
			name: "SSHAuthorizedKeyFingerprint",
			args: args{
				infoType: infofinderTypes.SSHAuthorizedKeyFingerprint,
			},
			want: to.Ptr(apitypes.InfoTypeSSHAuthorizedKeyFingerprint),
		},
		{
			name: "SSHPrivateKeyFingerprint",
			args: args{
				infoType: infofinderTypes.SSHPrivateKeyFingerprint,
			},
			want: to.Ptr(apitypes.InfoTypeSSHPrivateKeyFingerprint),
		},
		{
			name: "SSHDaemonKeyFingerprint",
			args: args{
				infoType: infofinderTypes.SSHDaemonKeyFingerprint,
			},
			want: to.Ptr(apitypes.InfoTypeSSHDaemonKeyFingerprint),
		},
		{
			name: "unknown",
			args: args{
				infoType: "unknown",
			},
			want: to.Ptr(apitypes.InfoTypeUNKNOWN),
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
