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

package database

import (
	"reflect"
	"sort"
	"testing"

	dockle_types "github.com/Portshift/dockle/pkg/types"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gotest.tools/assert"

	"github.com/cisco-open/kubei/api/server/models"
	"github.com/cisco-open/kubei/backend/pkg/types"
	runtime_scan_models "github.com/cisco-open/kubei/runtime_scan/api/server/models"
)

type vulnerabilityInfo struct {
	// cvss
	Cvss *types.CVSS `json:"cvss,omitempty"`

	// description
	Description string `json:"description,omitempty"`

	// links
	Links []string `json:"links"`

	// severity
	Severity types.VulnerabilitySeverity `json:"severity,omitempty"`

	// vulnerability name
	Name string `json:"vulnerabilityName,omitempty"`
}

func TestCreateResourceFromVulnerabilityScan(t *testing.T) {
	resourceInfo := &types.ResourceInfo{
		ResourceHash: "ResourceHash",
		ResourceName: "ResourceName",
		ResourceType: "ResourceType",
	}
	resourceID := CreateResourceID(resourceInfo)
	pkgInfo := &types.PackageInfo{
		Language: "pkg.language",
		License:  "pkg.license",
		Name:     "pkg.name",
		Version:  "pkg.version",
	}
	pkgID := CreatePackageID(pkgInfo)
	pkgInfo2 := &types.PackageInfo{
		Language: "pkg2.language",
		License:  "pkg2.license",
		Name:     "pkg2.name",
		Version:  "pkg2.version",
	}
	pkgID2 := CreatePackageID(pkgInfo2)
	vulInfo := vulnerabilityInfo{
		Cvss:        createTestCVSS(),
		Description: "Description",
		Links:       []string{"link1", "link2"},
		Severity:    types.VulnerabilitySeverityCRITICAL,
		Name:        "VulnerabilityName",
	}
	vulnerabilityID := CreateVulnerabilityID(&types.PackageVulnerabilityScan{VulnerabilityName: vulInfo.Name})
	vulInfo2 := vulnerabilityInfo{
		Cvss:        createTestCVSS(),
		Description: "Description2",
		Links:       []string{"link3", "link4"},
		Severity:    types.VulnerabilitySeverityCRITICAL,
		Name:        "VulnerabilityName2",
	}
	vulnerabilityID2 := CreateVulnerabilityID(&types.PackageVulnerabilityScan{VulnerabilityName: vulInfo2.Name})
	scannersList := []string{"scanner1", "scanner2"}
	type args struct {
		resource *types.ResourceVulnerabilityScan
		params   *TransactionParams
	}
	tests := []struct {
		name                      string
		args                      args
		want                      *Resource
		expectedTransactionParams *TransactionParams
	}{
		{
			name: "sanity",
			args: args{
				resource: &types.ResourceVulnerabilityScan{
					CisDockerBenchmarkResults: []*types.CISDockerBenchmarkResult{
						{
							Code:         "code1",
							Level:        int64(dockle_types.InfoLevel),
							Descriptions: "desc1",
						},
						{
							Code:         "code2",
							Level:        int64(dockle_types.WarnLevel),
							Descriptions: "desc2",
						},
					},
					PackageVulnerabilities: []*types.PackageVulnerabilityScan{
						{
							Cvss:              vulInfo.Cvss,
							Description:       vulInfo.Description,
							FixVersion:        "FixVersion",
							Links:             vulInfo.Links,
							Package:           pkgInfo,
							Scanners:          scannersList,
							Severity:          vulInfo.Severity,
							VulnerabilityName: vulInfo.Name,
						},
					},
					Resource: resourceInfo,
				},
				params: &TransactionParams{
					FixVersions:         map[PkgVulID]string{},
					Scanners:            map[ResourcePkgID][]string{},
					VulnerabilitySource: models.VulnerabilitySourceCICD,
				},
			},
			want: &Resource{
				ID:   resourceID,
				Hash: resourceInfo.ResourceHash,
				Name: resourceInfo.ResourceName,
				Type: resourceInfo.ResourceType,
				Packages: []Package{
					{
						ID:       pkgID,
						Name:     pkgInfo.Name,
						Version:  pkgInfo.Version,
						License:  pkgInfo.License,
						Language: pkgInfo.Language,
						Vulnerabilities: []Vulnerability{
							{
								ID:                vulnerabilityID,
								Name:              vulInfo.Name,
								Severity:          int(TypesVulnerabilitySeverityToInt[vulInfo.Severity]),
								Description:       vulInfo.Description,
								Links:             ArrayToDBArray(vulInfo.Links),
								CVSS:              CreateCVSSString(vulInfo.Cvss),
								CVSSBaseScore:     vulInfo.Cvss.GetBaseScore(),
								CVSSSeverity:      int(ModelsVulnerabilitySeverityToInt[vulInfo.Cvss.GetCVSSSeverity()]),
								ReportingScanners: ArrayToDBArray(scannersList),
								Source:            models.VulnerabilitySourceCICD,
							},
						},
					},
				},
				CISDockerBenchmarkResults: []CISDockerBenchmarkResult{
					{
						ResourceID:   resourceID,
						Code:         "code1",
						Level:        int(CISDockerBenchmarkLevelINFO),
						Descriptions: "desc1",
					},
					{
						ResourceID:   resourceID,
						Code:         "code2",
						Level:        int(CISDockerBenchmarkLevelWARN),
						Descriptions: "desc2",
					},
				},
			},
			expectedTransactionParams: &TransactionParams{
				FixVersions: map[PkgVulID]string{
					CreatePkgVulID(pkgID, vulnerabilityID): "FixVersion",
				},
				Scanners: map[ResourcePkgID][]string{
					CreateResourcePkgID(resourceID, pkgID): scannersList,
				},
				VulnerabilitySource: models.VulnerabilitySourceCICD,
			},
		},
		{
			name: "no cis docker benchmark results",
			args: args{
				resource: &types.ResourceVulnerabilityScan{
					PackageVulnerabilities: []*types.PackageVulnerabilityScan{
						{
							Cvss:              vulInfo.Cvss,
							Description:       vulInfo.Description,
							FixVersion:        "FixVersion",
							Links:             vulInfo.Links,
							Package:           pkgInfo,
							Scanners:          scannersList,
							Severity:          vulInfo.Severity,
							VulnerabilityName: vulInfo.Name,
						},
					},
					Resource: resourceInfo,
				},
				params: &TransactionParams{
					FixVersions:         map[PkgVulID]string{},
					Scanners:            map[ResourcePkgID][]string{},
					VulnerabilitySource: models.VulnerabilitySourceCICD,
				},
			},
			want: &Resource{
				ID:   resourceID,
				Hash: resourceInfo.ResourceHash,
				Name: resourceInfo.ResourceName,
				Type: resourceInfo.ResourceType,
				Packages: []Package{
					{
						ID:       pkgID,
						Name:     pkgInfo.Name,
						Version:  pkgInfo.Version,
						License:  pkgInfo.License,
						Language: pkgInfo.Language,
						Vulnerabilities: []Vulnerability{
							{
								ID:                vulnerabilityID,
								Name:              vulInfo.Name,
								Severity:          int(TypesVulnerabilitySeverityToInt[vulInfo.Severity]),
								Description:       vulInfo.Description,
								Links:             ArrayToDBArray(vulInfo.Links),
								CVSS:              CreateCVSSString(vulInfo.Cvss),
								CVSSBaseScore:     vulInfo.Cvss.GetBaseScore(),
								CVSSSeverity:      int(ModelsVulnerabilitySeverityToInt[vulInfo.Cvss.GetCVSSSeverity()]),
								ReportingScanners: ArrayToDBArray(scannersList),
								Source:            models.VulnerabilitySourceCICD,
							},
						},
					},
				},
			},
			expectedTransactionParams: &TransactionParams{
				FixVersions: map[PkgVulID]string{
					CreatePkgVulID(pkgID, vulnerabilityID): "FixVersion",
				},
				Scanners: map[ResourcePkgID][]string{
					CreateResourcePkgID(resourceID, pkgID): scannersList,
				},
				VulnerabilitySource: models.VulnerabilitySourceCICD,
			},
		},
		{
			name: "no fix version",
			args: args{
				resource: &types.ResourceVulnerabilityScan{
					PackageVulnerabilities: []*types.PackageVulnerabilityScan{
						{
							Cvss:              vulInfo.Cvss,
							Description:       vulInfo.Description,
							FixVersion:        "",
							Links:             vulInfo.Links,
							Package:           pkgInfo,
							Scanners:          scannersList,
							Severity:          vulInfo.Severity,
							VulnerabilityName: vulInfo.Name,
						},
					},
					Resource: resourceInfo,
				},
				params: &TransactionParams{
					FixVersions:         map[PkgVulID]string{},
					Scanners:            map[ResourcePkgID][]string{},
					VulnerabilitySource: models.VulnerabilitySourceCICD,
				},
			},
			want: &Resource{
				ID:   resourceID,
				Hash: resourceInfo.ResourceHash,
				Name: resourceInfo.ResourceName,
				Type: resourceInfo.ResourceType,
				Packages: []Package{
					{
						ID:       pkgID,
						Name:     pkgInfo.Name,
						Version:  pkgInfo.Version,
						License:  pkgInfo.License,
						Language: pkgInfo.Language,
						Vulnerabilities: []Vulnerability{
							{
								ID:                vulnerabilityID,
								Name:              vulInfo.Name,
								Severity:          int(TypesVulnerabilitySeverityToInt[vulInfo.Severity]),
								Description:       vulInfo.Description,
								Links:             ArrayToDBArray(vulInfo.Links),
								CVSS:              CreateCVSSString(vulInfo.Cvss),
								CVSSBaseScore:     vulInfo.Cvss.GetBaseScore(),
								CVSSSeverity:      int(ModelsVulnerabilitySeverityToInt[vulInfo.Cvss.GetCVSSSeverity()]),
								ReportingScanners: ArrayToDBArray(scannersList),
								Source:            models.VulnerabilitySourceCICD,
							},
						},
					},
				},
			},
			expectedTransactionParams: &TransactionParams{
				FixVersions: map[PkgVulID]string{},
				Scanners: map[ResourcePkgID][]string{
					CreateResourcePkgID(resourceID, pkgID): scannersList,
				},
				VulnerabilitySource: models.VulnerabilitySourceCICD,
			},
		},
		{
			name: "same package different vul",
			args: args{
				resource: &types.ResourceVulnerabilityScan{
					PackageVulnerabilities: []*types.PackageVulnerabilityScan{
						{
							Cvss:              vulInfo.Cvss,
							Description:       vulInfo.Description,
							FixVersion:        "FixVersion",
							Links:             vulInfo.Links,
							Package:           pkgInfo,
							Scanners:          scannersList,
							Severity:          vulInfo.Severity,
							VulnerabilityName: vulInfo.Name,
						},
						{
							Cvss:              vulInfo2.Cvss,
							Description:       vulInfo2.Description,
							FixVersion:        "FixVersion",
							Links:             vulInfo2.Links,
							Package:           pkgInfo,
							Scanners:          scannersList,
							Severity:          vulInfo2.Severity,
							VulnerabilityName: vulInfo2.Name,
						},
					},
					Resource: resourceInfo,
				},
				params: &TransactionParams{
					FixVersions:         map[PkgVulID]string{},
					Scanners:            map[ResourcePkgID][]string{},
					VulnerabilitySource: models.VulnerabilitySourceCICD,
				},
			},
			want: &Resource{
				ID:   resourceID,
				Hash: resourceInfo.ResourceHash,
				Name: resourceInfo.ResourceName,
				Type: resourceInfo.ResourceType,
				Packages: []Package{
					{
						ID:       pkgID,
						Name:     pkgInfo.Name,
						Version:  pkgInfo.Version,
						License:  pkgInfo.License,
						Language: pkgInfo.Language,
						Vulnerabilities: []Vulnerability{
							{
								ID:                vulnerabilityID,
								Name:              vulInfo.Name,
								Severity:          int(TypesVulnerabilitySeverityToInt[vulInfo.Severity]),
								Description:       vulInfo.Description,
								Links:             ArrayToDBArray(vulInfo.Links),
								CVSS:              CreateCVSSString(vulInfo.Cvss),
								CVSSBaseScore:     vulInfo.Cvss.GetBaseScore(),
								CVSSSeverity:      int(ModelsVulnerabilitySeverityToInt[vulInfo.Cvss.GetCVSSSeverity()]),
								ReportingScanners: ArrayToDBArray(scannersList),
								Source:            models.VulnerabilitySourceCICD,
							},
							{
								ID:                vulnerabilityID2,
								Name:              vulInfo2.Name,
								Severity:          int(TypesVulnerabilitySeverityToInt[vulInfo2.Severity]),
								Description:       vulInfo2.Description,
								Links:             ArrayToDBArray(vulInfo2.Links),
								CVSS:              CreateCVSSString(vulInfo2.Cvss),
								CVSSBaseScore:     vulInfo2.Cvss.GetBaseScore(),
								CVSSSeverity:      int(ModelsVulnerabilitySeverityToInt[vulInfo2.Cvss.GetCVSSSeverity()]),
								ReportingScanners: ArrayToDBArray(scannersList),
								Source:            models.VulnerabilitySourceCICD,
							},
						},
					},
				},
			},
			expectedTransactionParams: &TransactionParams{
				FixVersions: map[PkgVulID]string{
					CreatePkgVulID(pkgID, vulnerabilityID):  "FixVersion",
					CreatePkgVulID(pkgID, vulnerabilityID2): "FixVersion",
				},
				Scanners: map[ResourcePkgID][]string{
					CreateResourcePkgID(resourceID, pkgID): scannersList,
				},
				VulnerabilitySource: models.VulnerabilitySourceCICD,
			},
		},
		{
			name: "different package different vul",
			args: args{
				resource: &types.ResourceVulnerabilityScan{
					PackageVulnerabilities: []*types.PackageVulnerabilityScan{
						{
							Cvss:              vulInfo.Cvss,
							Description:       vulInfo.Description,
							FixVersion:        "FixVersion",
							Links:             vulInfo.Links,
							Package:           pkgInfo,
							Scanners:          scannersList,
							Severity:          vulInfo.Severity,
							VulnerabilityName: vulInfo.Name,
						},
						{
							Cvss:              vulInfo2.Cvss,
							Description:       vulInfo2.Description,
							FixVersion:        "FixVersion",
							Links:             vulInfo2.Links,
							Package:           pkgInfo2,
							Scanners:          scannersList,
							Severity:          vulInfo2.Severity,
							VulnerabilityName: vulInfo2.Name,
						},
					},
					Resource: resourceInfo,
				},
				params: &TransactionParams{
					FixVersions:         map[PkgVulID]string{},
					Scanners:            map[ResourcePkgID][]string{},
					VulnerabilitySource: models.VulnerabilitySourceCICD,
				},
			},
			want: &Resource{
				ID:   resourceID,
				Hash: resourceInfo.ResourceHash,
				Name: resourceInfo.ResourceName,
				Type: resourceInfo.ResourceType,
				Packages: []Package{
					{
						ID:       pkgID,
						Name:     pkgInfo.Name,
						Version:  pkgInfo.Version,
						License:  pkgInfo.License,
						Language: pkgInfo.Language,
						Vulnerabilities: []Vulnerability{
							{
								ID:                vulnerabilityID,
								Name:              vulInfo.Name,
								Severity:          int(TypesVulnerabilitySeverityToInt[vulInfo.Severity]),
								Description:       vulInfo.Description,
								Links:             ArrayToDBArray(vulInfo.Links),
								CVSS:              CreateCVSSString(vulInfo.Cvss),
								CVSSBaseScore:     vulInfo.Cvss.GetBaseScore(),
								CVSSSeverity:      int(ModelsVulnerabilitySeverityToInt[vulInfo.Cvss.GetCVSSSeverity()]),
								ReportingScanners: ArrayToDBArray(scannersList),
								Source:            models.VulnerabilitySourceCICD,
							},
						},
					},
					{
						ID:       pkgID2,
						Name:     pkgInfo2.Name,
						Version:  pkgInfo2.Version,
						License:  pkgInfo2.License,
						Language: pkgInfo2.Language,
						Vulnerabilities: []Vulnerability{
							{
								ID:                vulnerabilityID2,
								Name:              vulInfo2.Name,
								Severity:          int(TypesVulnerabilitySeverityToInt[vulInfo2.Severity]),
								Description:       vulInfo2.Description,
								Links:             ArrayToDBArray(vulInfo2.Links),
								CVSS:              CreateCVSSString(vulInfo2.Cvss),
								CVSSBaseScore:     vulInfo2.Cvss.GetBaseScore(),
								CVSSSeverity:      int(ModelsVulnerabilitySeverityToInt[vulInfo2.Cvss.GetCVSSSeverity()]),
								ReportingScanners: ArrayToDBArray(scannersList),
								Source:            models.VulnerabilitySourceCICD,
							},
						},
					},
				},
			},
			expectedTransactionParams: &TransactionParams{
				FixVersions: map[PkgVulID]string{
					CreatePkgVulID(pkgID, vulnerabilityID):   "FixVersion",
					CreatePkgVulID(pkgID2, vulnerabilityID2): "FixVersion",
				},
				Scanners: map[ResourcePkgID][]string{
					CreateResourcePkgID(resourceID, pkgID):  scannersList,
					CreateResourcePkgID(resourceID, pkgID2): scannersList,
				},
				VulnerabilitySource: models.VulnerabilitySourceCICD,
			},
		},
		{
			name: "different package same vul",
			args: args{
				resource: &types.ResourceVulnerabilityScan{
					PackageVulnerabilities: []*types.PackageVulnerabilityScan{
						{
							Cvss:              vulInfo.Cvss,
							Description:       vulInfo.Description,
							FixVersion:        "FixVersion",
							Links:             vulInfo.Links,
							Package:           pkgInfo,
							Scanners:          scannersList,
							Severity:          vulInfo.Severity,
							VulnerabilityName: vulInfo.Name,
						},
						{
							Cvss:              vulInfo.Cvss,
							Description:       vulInfo.Description,
							FixVersion:        "FixVersion",
							Links:             vulInfo.Links,
							Package:           pkgInfo2,
							Scanners:          scannersList,
							Severity:          vulInfo.Severity,
							VulnerabilityName: vulInfo.Name,
						},
					},
					Resource: resourceInfo,
				},
				params: &TransactionParams{
					FixVersions:         map[PkgVulID]string{},
					Scanners:            map[ResourcePkgID][]string{},
					VulnerabilitySource: models.VulnerabilitySourceCICD,
				},
			},
			want: &Resource{
				ID:   resourceID,
				Hash: resourceInfo.ResourceHash,
				Name: resourceInfo.ResourceName,
				Type: resourceInfo.ResourceType,
				Packages: []Package{
					{
						ID:       pkgID,
						Name:     pkgInfo.Name,
						Version:  pkgInfo.Version,
						License:  pkgInfo.License,
						Language: pkgInfo.Language,
						Vulnerabilities: []Vulnerability{
							{
								ID:                vulnerabilityID,
								Name:              vulInfo.Name,
								Severity:          int(TypesVulnerabilitySeverityToInt[vulInfo.Severity]),
								Description:       vulInfo.Description,
								Links:             ArrayToDBArray(vulInfo.Links),
								CVSS:              CreateCVSSString(vulInfo.Cvss),
								CVSSBaseScore:     vulInfo.Cvss.GetBaseScore(),
								CVSSSeverity:      int(ModelsVulnerabilitySeverityToInt[vulInfo.Cvss.GetCVSSSeverity()]),
								ReportingScanners: ArrayToDBArray(scannersList),
								Source:            models.VulnerabilitySourceCICD,
							},
						},
					},
					{
						ID:       pkgID2,
						Name:     pkgInfo2.Name,
						Version:  pkgInfo2.Version,
						License:  pkgInfo2.License,
						Language: pkgInfo2.Language,
						Vulnerabilities: []Vulnerability{
							{
								ID:                vulnerabilityID,
								Name:              vulInfo.Name,
								Severity:          int(TypesVulnerabilitySeverityToInt[vulInfo.Severity]),
								Description:       vulInfo.Description,
								Links:             ArrayToDBArray(vulInfo.Links),
								CVSS:              CreateCVSSString(vulInfo.Cvss),
								CVSSBaseScore:     vulInfo.Cvss.GetBaseScore(),
								CVSSSeverity:      int(ModelsVulnerabilitySeverityToInt[vulInfo.Cvss.GetCVSSSeverity()]),
								ReportingScanners: ArrayToDBArray(scannersList),
								Source:            models.VulnerabilitySourceCICD,
							},
						},
					},
				},
			},
			expectedTransactionParams: &TransactionParams{
				FixVersions: map[PkgVulID]string{
					CreatePkgVulID(pkgID, vulnerabilityID):  "FixVersion",
					CreatePkgVulID(pkgID2, vulnerabilityID): "FixVersion",
				},
				Scanners: map[ResourcePkgID][]string{
					CreateResourcePkgID(resourceID, pkgID):  scannersList,
					CreateResourcePkgID(resourceID, pkgID2): scannersList,
				},
				VulnerabilitySource: models.VulnerabilitySourceCICD,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateResourceFromVulnerabilityScan(tt.args.resource, tt.args.params)
			sort.Slice(got.Packages, func(i, j int) bool {
				return got.Packages[i].Name < got.Packages[j].Name
			})
			assert.DeepEqual(t, got, tt.want, cmpopts.IgnoreTypes(Vulnerability{}.ScannedAt))
			for id := range tt.args.params.Scanners {
				sort.Strings(tt.args.params.Scanners[id])
			}
			for id := range tt.expectedTransactionParams.Scanners {
				sort.Strings(tt.expectedTransactionParams.Scanners[id])
			}
			assert.DeepEqual(t, tt.args.params, tt.expectedTransactionParams)
		})
	}
}

func createTestCVSS() *types.CVSS {
	return &types.CVSS{
		CvssV3Metrics: &types.CVSSV3Metrics{
			BaseScore:           1.1,
			ExploitabilityScore: 2.2,
			ImpactScore:         3.3,
		},
		CvssV3Vector: &types.CVSSV3Vector{
			AttackComplexity:   types.AttackComplexityHIGH,
			AttackVector:       types.AttackVectorNETWORK,
			Availability:       types.AvailabilityLOW,
			Confidentiality:    types.ConfidentialityHIGH,
			Integrity:          types.IntegrityHIGH,
			PrivilegesRequired: types.PrivilegesRequiredHIGH,
			Scope:              types.ScopeCHANGED,
			UserInteraction:    types.UserInteractionNONE,
			Vector:             "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H",
		},
	}
}

func TestUpdateResourceAnalyzers(t *testing.T) {
	resourceInfo := &types.ResourceInfo{
		ResourceHash: "ResourceHash",
		ResourceName: "ResourceName",
		ResourceType: "ResourceType",
	}
	resourceID := CreateResourceID(resourceInfo)
	pkgInfo := &types.PackageInfo{
		Language: "pkg.language",
		License:  "pkg.license",
		Name:     "pkg.name",
		Version:  "pkg.version",
	}
	pkgID := CreatePackageID(pkgInfo)
	pkgInfo2 := &types.PackageInfo{
		Language: "pkg2.language",
		License:  "pkg2.license",
		Name:     "pkg2.name",
		Version:  "pkg2.version",
	}
	pkgID2 := CreatePackageID(pkgInfo2)
	type args struct {
		resources                []Resource
		resourcePkgIDToAnalyzers map[ResourcePkgID][]string
	}
	tests := []struct {
		name string
		args args
		want []Resource
	}{
		{
			name: "sanity",
			args: args{
				resources: []Resource{
					*(CreateResource(resourceInfo).WithPackages([]Package{
						*CreatePackage(pkgInfo, nil),
						*CreatePackage(pkgInfo2, nil),
					})),
				},
				resourcePkgIDToAnalyzers: map[ResourcePkgID][]string{
					CreateResourcePkgID(resourceID, pkgID):  {"analyzer1"},
					CreateResourcePkgID(resourceID, pkgID2): {"analyzer2"},
				},
			},
			want: []Resource{
				*(CreateResource(resourceInfo).
					WithPackages([]Package{
						*CreatePackage(pkgInfo, nil),
						*CreatePackage(pkgInfo2, nil),
					}).
					WithAnalyzers([]string{"analyzer1", "analyzer2"})),
			},
		},
		{
			name: "multiple resources",
			args: args{
				resources: []Resource{
					*(CreateResource(resourceInfo).WithPackages([]Package{
						*CreatePackage(pkgInfo, nil),
					})),
					*(CreateResource(resourceInfo).WithPackages([]Package{
						*CreatePackage(pkgInfo2, nil),
					})),
					*(CreateResource(resourceInfo).WithPackages([]Package{
						*CreatePackage(pkgInfo, nil),
						*CreatePackage(pkgInfo2, nil),
					})),
				},
				resourcePkgIDToAnalyzers: map[ResourcePkgID][]string{
					CreateResourcePkgID(resourceID, pkgID):  {"analyzer1"},
					CreateResourcePkgID(resourceID, pkgID2): {"analyzer2"},
				},
			},
			want: []Resource{
				*(CreateResource(resourceInfo).
					WithPackages([]Package{
						*CreatePackage(pkgInfo, nil),
					}).
					WithAnalyzers([]string{"analyzer1"})),
				*(CreateResource(resourceInfo).
					WithPackages([]Package{
						*CreatePackage(pkgInfo2, nil),
					}).
					WithAnalyzers([]string{"analyzer2"})),
				*(CreateResource(resourceInfo).
					WithPackages([]Package{
						*CreatePackage(pkgInfo, nil),
						*CreatePackage(pkgInfo2, nil),
					}).
					WithAnalyzers([]string{"analyzer1", "analyzer2"})),
			},
		},
		{
			name: "empty resourcePkgIDToAnalyzers",
			args: args{
				resources: []Resource{
					*(CreateResource(resourceInfo).WithPackages([]Package{
						*CreatePackage(pkgInfo, nil),
						*CreatePackage(pkgInfo2, nil),
					})),
				},
				resourcePkgIDToAnalyzers: nil,
			},
			want: []Resource{
				*(CreateResource(resourceInfo).
					WithPackages([]Package{
						*CreatePackage(pkgInfo, nil),
						*CreatePackage(pkgInfo2, nil),
					})),
			},
		},
		{
			name: "only one resource+pkg match",
			args: args{
				resources: []Resource{
					*(CreateResource(resourceInfo).WithPackages([]Package{
						*CreatePackage(pkgInfo, nil),
					})),
				},
				resourcePkgIDToAnalyzers: map[ResourcePkgID][]string{
					CreateResourcePkgID(resourceID, pkgID):  {"analyzer1"},
					CreateResourcePkgID(resourceID, pkgID2): {"analyzer2"},
				},
			},
			want: []Resource{
				*(CreateResource(resourceInfo).
					WithPackages([]Package{
						*CreatePackage(pkgInfo, nil),
					}).
					WithAnalyzers([]string{"analyzer1"})),
			},
		},
		{
			name: "remove duplicates",
			args: args{
				resources: []Resource{
					*(CreateResource(resourceInfo).WithPackages([]Package{
						*CreatePackage(pkgInfo, nil),
						*CreatePackage(pkgInfo2, nil),
					})),
				},
				resourcePkgIDToAnalyzers: map[ResourcePkgID][]string{
					CreateResourcePkgID(resourceID, pkgID):  {"analyzer1", "analyzer2"},
					CreateResourcePkgID(resourceID, pkgID2): {"analyzer2", "analyzer3"},
				},
			},
			want: []Resource{
				*(CreateResource(resourceInfo).
					WithPackages([]Package{
						*CreatePackage(pkgInfo, nil),
						*CreatePackage(pkgInfo2, nil),
					}).
					WithAnalyzers([]string{"analyzer1", "analyzer2", "analyzer3"})),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UpdateResourceAnalyzers(tt.args.resources, tt.args.resourcePkgIDToAnalyzers)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateResourceAnalyzers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateResourceFromRuntimeContentAnalysis(t *testing.T) {
	resourceInfo := &types.ResourceInfo{
		ResourceHash: "resource-1-hash",
		ResourceName: "resource-1-name",
		ResourceType: "resource-1-type",
	}
	resourceID := CreateResourceID(resourceInfo)
	pkgInfo1 := &types.PackageInfo{
		Language: "pkg-1-language",
		License:  "pkg-1-license",
		Name:     "pkg-1-name",
		Version:  "pkg-1-version",
	}
	pkgInfo1ID := CreatePackageID(pkgInfo1)
	pkgInfo2 := &types.PackageInfo{
		Language: "pkg-2-language",
		License:  "pkg-2-license",
		Name:     "pkg-2-name",
		Version:  "pkg-2-version",
	}
	pkgInfo2ID := CreatePackageID(pkgInfo2)
	type args struct {
		resourceContentAnalysis *runtime_scan_models.ResourceContentAnalysis
		params                  *TransactionParams
	}
	tests := []struct {
		name                      string
		args                      args
		want                      *Resource
		expectedTransactionParams *TransactionParams
	}{
		{
			name: "sanity",
			args: args{
				resourceContentAnalysis: &runtime_scan_models.ResourceContentAnalysis{
					Packages: []*runtime_scan_models.PackageContentAnalysis{
						{
							Analyzers: []string{"analyzer1", "analyzer11"},
							Package: &runtime_scan_models.PackageInfo{
								Language: pkgInfo1.Language,
								License:  pkgInfo1.License,
								Name:     pkgInfo1.Name,
								Version:  pkgInfo1.Version,
							},
						},
						{
							Analyzers: []string{"analyzer2", "analyzer11"},
							Package: &runtime_scan_models.PackageInfo{
								Language: pkgInfo2.Language,
								License:  pkgInfo2.License,
								Name:     pkgInfo2.Name,
								Version:  pkgInfo2.Version,
							},
						},
					},
					Resource: &runtime_scan_models.ResourceInfo{
						ResourceHash: resourceInfo.ResourceHash,
						ResourceName: resourceInfo.ResourceName,
						ResourceType: runtime_scan_models.ResourceType(resourceInfo.ResourceType),
					},
				},
				params: &TransactionParams{
					Analyzers: map[ResourcePkgID][]string{},
				},
			},
			want: &Resource{
				ID:                 resourceID,
				Hash:               resourceInfo.ResourceHash,
				Name:               resourceInfo.ResourceName,
				Type:               resourceInfo.ResourceType,
				ReportingAnalyzers: ArrayToDBArray([]string{"analyzer1", "analyzer11", "analyzer2"}),
				Packages: []Package{
					{
						ID:       pkgInfo1ID,
						Name:     pkgInfo1.Name,
						Version:  pkgInfo1.Version,
						License:  pkgInfo1.License,
						Language: pkgInfo1.Language,
					},
					{
						ID:       pkgInfo2ID,
						Name:     pkgInfo2.Name,
						Version:  pkgInfo2.Version,
						License:  pkgInfo2.License,
						Language: pkgInfo2.Language,
					},
				},
			},
			expectedTransactionParams: &TransactionParams{
				Analyzers: map[ResourcePkgID][]string{
					CreateResourcePkgID(resourceID, pkgInfo1ID): {"analyzer1", "analyzer11"},
					CreateResourcePkgID(resourceID, pkgInfo2ID): {"analyzer2", "analyzer11"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateResourceFromRuntimeContentAnalysis(tt.args.resourceContentAnalysis, tt.args.params); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateResourceFromRuntimeContentAnalysis() = %v, want %v", got, tt.want)
			}
			assert.DeepEqual(t, tt.args.params, tt.expectedTransactionParams)
		})
	}
}

func TestCreateResourceFromContentAnalysis(t *testing.T) {
	resourceInfo := &types.ResourceInfo{
		ResourceHash: "resource-1-hash",
		ResourceName: "resource-1-name",
		ResourceType: "resource-1-type",
	}
	resourceID := CreateResourceID(resourceInfo)
	pkgInfo1 := &types.PackageInfo{
		Language: "pkg-1-language",
		License:  "pkg-1-license",
		Name:     "pkg-1-name",
		Version:  "pkg-1-version",
	}
	pkgInfo1ID := CreatePackageID(pkgInfo1)
	pkgInfo2 := &types.PackageInfo{
		Language: "pkg-2-language",
		License:  "pkg-2-license",
		Name:     "pkg-2-name",
		Version:  "pkg-2-version",
	}
	pkgInfo2ID := CreatePackageID(pkgInfo2)
	type args struct {
		resourceContentAnalysis *models.ResourceContentAnalysis
		params                  *TransactionParams
	}
	tests := []struct {
		name                      string
		args                      args
		want                      *Resource
		expectedTransactionParams *TransactionParams
	}{
		{
			name: "sanity",
			args: args{
				resourceContentAnalysis: &models.ResourceContentAnalysis{
					Packages: []*models.PackageContentAnalysis{
						{
							Analyzers: []string{"analyzer1", "analyzer11"},
							Package: &models.PackageInfo{
								Language: pkgInfo1.Language,
								License:  pkgInfo1.License,
								Name:     pkgInfo1.Name,
								Version:  pkgInfo1.Version,
							},
						},
						{
							Analyzers: []string{"analyzer2", "analyzer11"},
							Package: &models.PackageInfo{
								Language: pkgInfo2.Language,
								License:  pkgInfo2.License,
								Name:     pkgInfo2.Name,
								Version:  pkgInfo2.Version,
							},
						},
					},
					Resource: &models.ResourceInfo{
						ResourceHash: resourceInfo.ResourceHash,
						ResourceName: resourceInfo.ResourceName,
						ResourceType: models.ResourceType(resourceInfo.ResourceType),
					},
				},
				params: &TransactionParams{
					Analyzers: map[ResourcePkgID][]string{},
				},
			},
			want: &Resource{
				ID:                 resourceID,
				Hash:               resourceInfo.ResourceHash,
				Name:               resourceInfo.ResourceName,
				Type:               resourceInfo.ResourceType,
				ReportingAnalyzers: ArrayToDBArray([]string{"analyzer1", "analyzer11", "analyzer2"}),
				Packages: []Package{
					{
						ID:       pkgInfo1ID,
						Name:     pkgInfo1.Name,
						Version:  pkgInfo1.Version,
						License:  pkgInfo1.License,
						Language: pkgInfo1.Language,
					},
					{
						ID:       pkgInfo2ID,
						Name:     pkgInfo2.Name,
						Version:  pkgInfo2.Version,
						License:  pkgInfo2.License,
						Language: pkgInfo2.Language,
					},
				},
			},
			expectedTransactionParams: &TransactionParams{
				Analyzers: map[ResourcePkgID][]string{
					CreateResourcePkgID(resourceID, pkgInfo1ID): {"analyzer1", "analyzer11"},
					CreateResourcePkgID(resourceID, pkgInfo2ID): {"analyzer2", "analyzer11"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateResourceFromContentAnalysis(tt.args.resourceContentAnalysis, tt.args.params); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateResourceFromContentAnalysis() = %v, want %v", got, tt.want)
			}
			assert.DeepEqual(t, tt.args.params, tt.expectedTransactionParams)
		})
	}
}
