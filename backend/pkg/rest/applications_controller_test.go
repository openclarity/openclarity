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

package rest

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	dockle_types "github.com/Portshift/dockle/pkg/types"
	"github.com/golang/mock/gomock"
	"gotest.tools/assert"

	"github.com/cisco-open/kubei/backend/pkg/database"
	"github.com/cisco-open/kubei/backend/pkg/types"
)

func Test_mergeResourcePkgIDToAnalyzersMaps(t *testing.T) {
	type args struct {
		a map[database.ResourcePkgID][]string
		b map[database.ResourcePkgID][]string
	}
	tests := []struct {
		name string
		args args
		want map[database.ResourcePkgID][]string
	}{
		{
			name: "sanity merge",
			args: args{
				a: map[database.ResourcePkgID][]string{
					database.ResourcePkgID("a1"): {"1"},
					database.ResourcePkgID("a2"): {"2"},
				},
				b: map[database.ResourcePkgID][]string{
					database.ResourcePkgID("b1"): {"1"},
					database.ResourcePkgID("b2"): {"2"},
				},
			},
			want: map[database.ResourcePkgID][]string{
				database.ResourcePkgID("a1"): {"1"},
				database.ResourcePkgID("a2"): {"2"},
				database.ResourcePkgID("b1"): {"1"},
				database.ResourcePkgID("b2"): {"2"},
			},
		},
		{
			name: "merge with duplicate ids and duplicate values",
			args: args{
				a: map[database.ResourcePkgID][]string{
					database.ResourcePkgID("a1"): {"1"},
					database.ResourcePkgID("a2"): {"2"},
				},
				b: map[database.ResourcePkgID][]string{
					database.ResourcePkgID("a1"): {"1"},
					database.ResourcePkgID("a2"): {"3"},
				},
			},
			want: map[database.ResourcePkgID][]string{
				database.ResourcePkgID("a1"): {"1"},
				database.ResourcePkgID("a2"): {"2", "3"},
			},
		},
		{
			name: "empty a",
			args: args{
				b: map[database.ResourcePkgID][]string{
					database.ResourcePkgID("b1"): {"1"},
					database.ResourcePkgID("b2"): {"2"},
				},
			},
			want: map[database.ResourcePkgID][]string{
				database.ResourcePkgID("b1"): {"1"},
				database.ResourcePkgID("b2"): {"2"},
			},
		},
		{
			name: "empty b",
			args: args{
				a: map[database.ResourcePkgID][]string{
					database.ResourcePkgID("a1"): {"1"},
					database.ResourcePkgID("a2"): {"2"},
				},
			},
			want: map[database.ResourcePkgID][]string{
				database.ResourcePkgID("a1"): {"1"},
				database.ResourcePkgID("a2"): {"2"},
			},
		},
		{
			name: "empty both",
			args: args{},
			want: map[database.ResourcePkgID][]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeResourcePkgIDToAnalyzersMaps(tt.args.a, tt.args.b)
			for _, strings := range got {
				sort.Strings(strings)
			}
			for _, strings := range tt.want {
				sort.Strings(strings)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeResourcePkgIDToAnalyzersMaps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_updateCurrentResource(t *testing.T) {
	type args struct {
		currentResource   *database.Resource
		newResource       *database.Resource
		transactionParams *database.TransactionParams
	}
	tests := []struct {
		name                      string
		args                      args
		want                      *database.Resource
		expectedTransactionParams *database.TransactionParams
	}{
		{
			name: "1 new pkg",
			args: args{
				currentResource: &database.Resource{
					ID:                 "ResourceID",
					Hash:               "ResourceHash",
					Name:               "ResourceName",
					Type:               "ResourceType",
					SbomID:             "ResourceSbomId",
					ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer1"}),
					Packages: []database.Package{
						{
							ID:              "pkg-id-1",
							Name:            "pkg-name-1",
							Version:         "pkg-version-1",
							License:         "pkg-license-1",
							Language:        "pkg-language-1",
							Vulnerabilities: []database.Vulnerability{},
						},
					},
				},
				newResource: &database.Resource{
					ID:                 "ResourceID",
					Hash:               "ResourceHash",
					Name:               "ResourceName",
					Type:               "ResourceType",
					SbomID:             "ResourceSbomId",
					ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer1"}),
					Packages: []database.Package{
						{
							ID:       "pkg-id-2",
							Name:     "pkg-name-2",
							Version:  "pkg-version-2",
							License:  "pkg-license-2",
							Language: "pkg-language-2",
							Vulnerabilities: []database.Vulnerability{
								{
									ID:   "vul-id-1",
									Name: "vul-name-1",
								},
							},
						},
					},
				},
				transactionParams: &database.TransactionParams{
					Scanners: map[database.ResourcePkgID][]string{
						database.CreateResourcePkgID("ResourceID", "pkg-id-2"): {"scanner1"},
					},
					Analyzers: map[database.ResourcePkgID][]string{},
				},
			},
			expectedTransactionParams: &database.TransactionParams{
				Scanners: map[database.ResourcePkgID][]string{
					database.CreateResourcePkgID("ResourceID", "pkg-id-2"): {"scanner1"},
				},
				Analyzers: map[database.ResourcePkgID][]string{
					database.CreateResourcePkgID("ResourceID", "pkg-id-2"): {"scanner1"},
				},
			},
			want: &database.Resource{
				ID:                 "ResourceID",
				Hash:               "ResourceHash",
				Name:               "ResourceName",
				Type:               "ResourceType",
				SbomID:             "ResourceSbomId",
				ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer1"}),
				Packages: []database.Package{
					{
						ID:              "pkg-id-1",
						Name:            "pkg-name-1",
						Version:         "pkg-version-1",
						License:         "pkg-license-1",
						Language:        "pkg-language-1",
						Vulnerabilities: []database.Vulnerability{},
					},
					{
						ID:       "pkg-id-2",
						Name:     "pkg-name-2",
						Version:  "pkg-version-2",
						License:  "pkg-license-2",
						Language: "pkg-language-2",
						Vulnerabilities: []database.Vulnerability{
							{
								ID:   "vul-id-1",
								Name: "vul-name-1",
							},
						},
					},
				},
			},
		},
		{
			name: "1 existing pkg",
			args: args{
				currentResource: &database.Resource{
					ID:                 "ResourceID",
					Hash:               "ResourceHash",
					Name:               "ResourceName",
					Type:               "ResourceType",
					SbomID:             "ResourceSbomId",
					ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer1"}),
					Packages: []database.Package{
						{
							ID:              "pkg-id-1",
							Name:            "pkg-name-1",
							Version:         "pkg-version-1",
							License:         "pkg-license-1",
							Language:        "pkg-language-1",
							Vulnerabilities: []database.Vulnerability{},
						},
					},
				},
				newResource: &database.Resource{
					ID:                 "ResourceID",
					Hash:               "ResourceHash",
					Name:               "ResourceName",
					Type:               "ResourceType",
					SbomID:             "ResourceSbomId",
					ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer1"}),
					Packages: []database.Package{
						{
							ID:       "pkg-id-1",
							Name:     "pkg-name-1",
							Version:  "pkg-version-1",
							License:  "pkg-license-1",
							Language: "pkg-language-1",
							Vulnerabilities: []database.Vulnerability{
								{
									ID:   "vul-id-1",
									Name: "vul-name-1",
								},
							},
						},
					},
				},
			},
			want: &database.Resource{
				ID:                 "ResourceID",
				Hash:               "ResourceHash",
				Name:               "ResourceName",
				Type:               "ResourceType",
				SbomID:             "ResourceSbomId",
				ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer1"}),
				Packages: []database.Package{
					{
						ID:       "pkg-id-1",
						Name:     "pkg-name-1",
						Version:  "pkg-version-1",
						License:  "pkg-license-1",
						Language: "pkg-language-1",
						Vulnerabilities: []database.Vulnerability{
							{
								ID:   "vul-id-1",
								Name: "vul-name-1",
							},
						},
					},
				},
			},
		},
		{
			name: "1 new pkg + 1 existing pkg + 1 pkg only on current",
			args: args{
				currentResource: &database.Resource{
					ID:                 "ResourceID",
					Hash:               "ResourceHash",
					Name:               "ResourceName",
					Type:               "ResourceType",
					SbomID:             "ResourceSbomId",
					ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer1"}),
					Packages: []database.Package{
						{
							ID:              "pkg-id-0",
							Name:            "pkg-name-0",
							Version:         "pkg-version-0",
							License:         "pkg-license-0",
							Language:        "pkg-language-0",
							Vulnerabilities: []database.Vulnerability{},
						},
						{
							ID:              "pkg-id-1",
							Name:            "pkg-name-1",
							Version:         "pkg-version-1",
							License:         "pkg-license-1",
							Language:        "pkg-language-1",
							Vulnerabilities: []database.Vulnerability{},
						},
					},
				},
				newResource: &database.Resource{
					ID:                 "ResourceID",
					Hash:               "ResourceHash",
					Name:               "ResourceName",
					Type:               "ResourceType",
					SbomID:             "ResourceSbomId",
					ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer1"}),
					Packages: []database.Package{
						{
							ID:       "pkg-id-1",
							Name:     "pkg-name-1",
							Version:  "pkg-version-1",
							License:  "pkg-license-1",
							Language: "pkg-language-1",
							Vulnerabilities: []database.Vulnerability{
								{
									ID:   "vul-id-1",
									Name: "vul-name-1",
								},
							},
						},
						{
							ID:       "pkg-id-2",
							Name:     "pkg-name-2",
							Version:  "pkg-version-2",
							License:  "pkg-license-2",
							Language: "pkg-language-2",
							Vulnerabilities: []database.Vulnerability{
								{
									ID:   "vul-id-2",
									Name: "vul-name-2",
								},
							},
						},
					},
				},
				transactionParams: &database.TransactionParams{
					Scanners: map[database.ResourcePkgID][]string{
						database.CreateResourcePkgID("ResourceID", "pkg-id-2"): {"scanner1"},
					},
					Analyzers: map[database.ResourcePkgID][]string{},
				},
			},
			expectedTransactionParams: &database.TransactionParams{
				Scanners: map[database.ResourcePkgID][]string{
					database.CreateResourcePkgID("ResourceID", "pkg-id-2"): {"scanner1"},
				},
				Analyzers: map[database.ResourcePkgID][]string{
					database.CreateResourcePkgID("ResourceID", "pkg-id-2"): {"scanner1"},
				},
			},
			want: &database.Resource{
				ID:                 "ResourceID",
				Hash:               "ResourceHash",
				Name:               "ResourceName",
				Type:               "ResourceType",
				SbomID:             "ResourceSbomId",
				ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer1"}),
				Packages: []database.Package{
					{
						ID:              "pkg-id-0",
						Name:            "pkg-name-0",
						Version:         "pkg-version-0",
						License:         "pkg-license-0",
						Language:        "pkg-language-0",
						Vulnerabilities: []database.Vulnerability{},
					},
					{
						ID:       "pkg-id-1",
						Name:     "pkg-name-1",
						Version:  "pkg-version-1",
						License:  "pkg-license-1",
						Language: "pkg-language-1",
						Vulnerabilities: []database.Vulnerability{
							{
								ID:   "vul-id-1",
								Name: "vul-name-1",
							},
						},
					},
					{
						ID:       "pkg-id-2",
						Name:     "pkg-name-2",
						Version:  "pkg-version-2",
						License:  "pkg-license-2",
						Language: "pkg-language-2",
						Vulnerabilities: []database.Vulnerability{
							{
								ID:   "vul-id-2",
								Name: "vul-name-2",
							},
						},
					},
				},
			},
		},
		{
			name: "missing CISDockerBenchmarkChecks - keep old results",
			args: args{
				currentResource: &database.Resource{
					ID:                 "ResourceID",
					Hash:               "ResourceHash",
					Name:               "ResourceName",
					Type:               "ResourceType",
					SbomID:             "ResourceSbomId",
					ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer1"}),
					Packages: []database.Package{
						{
							ID:              "pkg-id-1",
							Name:            "pkg-name-1",
							Version:         "pkg-version-1",
							License:         "pkg-license-1",
							Language:        "pkg-language-1",
							Vulnerabilities: []database.Vulnerability{},
						},
					},
					CISDockerBenchmarkChecks: []database.CISDockerBenchmarkCheck{
						{
							ResourceID:   "ResourceID",
							Code:         "code1",
							Level:        int(database.CISDockerBenchmarkLevelINFO),
							Descriptions: "desc1",
						},
						{
							ResourceID:   "ResourceID",
							Code:         "code2",
							Level:        int(database.CISDockerBenchmarkLevelWARN),
							Descriptions: "desc2",
						},
					},
				},
				newResource: &database.Resource{
					ID:                 "ResourceID",
					Hash:               "ResourceHash",
					Name:               "ResourceName",
					Type:               "ResourceType",
					SbomID:             "ResourceSbomId",
					ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer1"}),
					Packages: []database.Package{
						{
							ID:       "pkg-id-2",
							Name:     "pkg-name-2",
							Version:  "pkg-version-2",
							License:  "pkg-license-2",
							Language: "pkg-language-2",
							Vulnerabilities: []database.Vulnerability{
								{
									ID:   "vul-id-1",
									Name: "vul-name-1",
								},
							},
						},
					},
				},
				transactionParams: &database.TransactionParams{
					Scanners: map[database.ResourcePkgID][]string{
						database.CreateResourcePkgID("ResourceID", "pkg-id-2"): {"scanner1"},
					},
					Analyzers: map[database.ResourcePkgID][]string{},
				},
			},
			expectedTransactionParams: &database.TransactionParams{
				Scanners: map[database.ResourcePkgID][]string{
					database.CreateResourcePkgID("ResourceID", "pkg-id-2"): {"scanner1"},
				},
				Analyzers: map[database.ResourcePkgID][]string{
					database.CreateResourcePkgID("ResourceID", "pkg-id-2"): {"scanner1"},
				},
			},
			want: &database.Resource{
				ID:                 "ResourceID",
				Hash:               "ResourceHash",
				Name:               "ResourceName",
				Type:               "ResourceType",
				SbomID:             "ResourceSbomId",
				ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer1"}),
				CISDockerBenchmarkChecks: []database.CISDockerBenchmarkCheck{
					{
						ResourceID:   "ResourceID",
						Code:         "code1",
						Level:        int(database.CISDockerBenchmarkLevelINFO),
						Descriptions: "desc1",
					},
					{
						ResourceID:   "ResourceID",
						Code:         "code2",
						Level:        int(database.CISDockerBenchmarkLevelWARN),
						Descriptions: "desc2",
					},
				},
				Packages: []database.Package{
					{
						ID:              "pkg-id-1",
						Name:            "pkg-name-1",
						Version:         "pkg-version-1",
						License:         "pkg-license-1",
						Language:        "pkg-language-1",
						Vulnerabilities: []database.Vulnerability{},
					},
					{
						ID:       "pkg-id-2",
						Name:     "pkg-name-2",
						Version:  "pkg-version-2",
						License:  "pkg-license-2",
						Language: "pkg-language-2",
						Vulnerabilities: []database.Vulnerability{
							{
								ID:   "vul-id-1",
								Name: "vul-name-1",
							},
						},
					},
				},
			},
		},
		{
			name: "new CISDockerBenchmarkChecks - replace",
			args: args{
				currentResource: &database.Resource{
					ID:                 "ResourceID",
					Hash:               "ResourceHash",
					Name:               "ResourceName",
					Type:               "ResourceType",
					SbomID:             "ResourceSbomId",
					ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer1"}),
					Packages: []database.Package{
						{
							ID:              "pkg-id-1",
							Name:            "pkg-name-1",
							Version:         "pkg-version-1",
							License:         "pkg-license-1",
							Language:        "pkg-language-1",
							Vulnerabilities: []database.Vulnerability{},
						},
					},
					CISDockerBenchmarkChecks: []database.CISDockerBenchmarkCheck{
						{
							ResourceID:   "ResourceID",
							Code:         "code1",
							Level:        int(database.CISDockerBenchmarkLevelINFO),
							Descriptions: "desc1",
						},
						{
							ResourceID:   "ResourceID",
							Code:         "code2",
							Level:        int(database.CISDockerBenchmarkLevelWARN),
							Descriptions: "desc2",
						},
					},
				},
				newResource: &database.Resource{
					ID:                 "ResourceID",
					Hash:               "ResourceHash",
					Name:               "ResourceName",
					Type:               "ResourceType",
					SbomID:             "ResourceSbomId",
					ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer1"}),
					CISDockerBenchmarkChecks: []database.CISDockerBenchmarkCheck{
						{
							ResourceID:   "ResourceID",
							Code:         "code1",
							Level:        int(database.CISDockerBenchmarkLevelWARN),
							Descriptions: "desc1",
						},
					},
					Packages: []database.Package{
						{
							ID:       "pkg-id-2",
							Name:     "pkg-name-2",
							Version:  "pkg-version-2",
							License:  "pkg-license-2",
							Language: "pkg-language-2",
							Vulnerabilities: []database.Vulnerability{
								{
									ID:   "vul-id-1",
									Name: "vul-name-1",
								},
							},
						},
					},
				},
				transactionParams: &database.TransactionParams{
					Scanners: map[database.ResourcePkgID][]string{
						database.CreateResourcePkgID("ResourceID", "pkg-id-2"): {"scanner1"},
					},
					Analyzers: map[database.ResourcePkgID][]string{},
				},
			},
			expectedTransactionParams: &database.TransactionParams{
				Scanners: map[database.ResourcePkgID][]string{
					database.CreateResourcePkgID("ResourceID", "pkg-id-2"): {"scanner1"},
				},
				Analyzers: map[database.ResourcePkgID][]string{
					database.CreateResourcePkgID("ResourceID", "pkg-id-2"): {"scanner1"},
				},
			},
			want: &database.Resource{
				ID:                 "ResourceID",
				Hash:               "ResourceHash",
				Name:               "ResourceName",
				Type:               "ResourceType",
				SbomID:             "ResourceSbomId",
				ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer1"}),
				CISDockerBenchmarkChecks: []database.CISDockerBenchmarkCheck{
					{
						ResourceID:   "ResourceID",
						Code:         "code1",
						Level:        int(database.CISDockerBenchmarkLevelWARN),
						Descriptions: "desc1",
					},
				},
				Packages: []database.Package{
					{
						ID:              "pkg-id-1",
						Name:            "pkg-name-1",
						Version:         "pkg-version-1",
						License:         "pkg-license-1",
						Language:        "pkg-language-1",
						Vulnerabilities: []database.Vulnerability{},
					},
					{
						ID:       "pkg-id-2",
						Name:     "pkg-name-2",
						Version:  "pkg-version-2",
						License:  "pkg-license-2",
						Language: "pkg-language-2",
						Vulnerabilities: []database.Vulnerability{
							{
								ID:   "vul-id-1",
								Name: "vul-name-1",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := updateResource(tt.args.currentResource, tt.args.newResource, tt.args.transactionParams)
			assert.DeepEqual(t, got, tt.want)
			assert.DeepEqual(t, tt.args.transactionParams, tt.expectedTransactionParams)
		})
	}
}

func Test_updateApplicationWithVulnerabilityScan(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	resourceInfo1 := &types.ResourceInfo{
		ResourceHash: "ResourceHash1",
		ResourceName: "ResourceName1",
		ResourceType: "ResourceType1",
	}
	resourceInfo2 := &types.ResourceInfo{
		ResourceHash: "ResourceHash2",
		ResourceName: "ResourceName2",
		ResourceType: "ResourceType2",
	}
	pkgInfo1 := &types.PackageInfo{
		Name:     "pkg-name-1",
		Version:  "pkg-version-1",
		License:  "pkg-license-1",
		Language: "pkg-language-1",
	}
	vulInfo1 := &types.PackageVulnerabilityScan{
		Cvss:              createTestCVSS(),
		Description:       "vul-desc",
		FixVersion:        "vul-fix-version",
		Links:             []string{"link1"},
		Package:           pkgInfo1,
		Scanners:          []string{"scanner1"},
		Severity:          types.VulnerabilitySeverityCRITICAL,
		VulnerabilityName: "vul-name",
	}
	pkgInfo2 := &types.PackageInfo{
		Name:     "pkg-name-2",
		Version:  "pkg-version-2",
		License:  "pkg-license-2",
		Language: "pkg-language-2",
	}
	vulInfo2 := &types.PackageVulnerabilityScan{
		Cvss:              createTestCVSS(),
		Description:       "vul-desc2",
		FixVersion:        "vul-fix-version2",
		Links:             []string{"link2"},
		Package:           pkgInfo2,
		Scanners:          []string{"scanner2"},
		Severity:          types.VulnerabilitySeverityCRITICAL,
		VulnerabilityName: "vul-name2",
	}
	type args struct {
		application                  *database.Application
		applicationVulnerabilityScan *types.ApplicationVulnerabilityScan
		shouldReplaceResources       bool
	}
	tests := []struct {
		name                      string
		args                      args
		want                      *database.Application
		expectedTransactionParams *database.TransactionParams
		mockResourceTableFunc     func(table *database.MockResourceTable)
	}{
		{
			name: "1 new resource + 1 only on current + should not replace resources list",
			args: args{
				application: &database.Application{
					ID:           "app-id",
					Name:         "app-name",
					Type:         "app-type",
					Labels:       "app-labels",
					Environments: "app-env",
					Resources: []database.Resource{
						{
							ID:                 database.CreateResourceID(resourceInfo1),
							Hash:               resourceInfo1.ResourceHash,
							Name:               resourceInfo1.ResourceName,
							Type:               resourceInfo1.ResourceType,
							ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer2"}),
							Packages: []database.Package{
								{
									ID:              database.CreatePackageID(pkgInfo1),
									Name:            pkgInfo1.Name,
									Version:         pkgInfo1.Version,
									License:         pkgInfo1.License,
									Language:        pkgInfo1.Language,
									Vulnerabilities: []database.Vulnerability{},
								},
							},
							CISDockerBenchmarkChecks: []database.CISDockerBenchmarkCheck{
								{
									ResourceID:   database.CreateResourceID(resourceInfo1),
									Code:         "code1",
									Level:        int(database.CISDockerBenchmarkLevelINFO),
									Descriptions: "desc1",
								},
								{
									ResourceID:   database.CreateResourceID(resourceInfo1),
									Code:         "code2",
									Level:        int(database.CISDockerBenchmarkLevelWARN),
									Descriptions: "desc2",
								},
							},
						},
					},
				},
				applicationVulnerabilityScan: &types.ApplicationVulnerabilityScan{
					Resources: []*types.ResourceVulnerabilityScan{
						{
							PackageVulnerabilities: []*types.PackageVulnerabilityScan{
								vulInfo1,
								vulInfo2,
							},
							Resource: resourceInfo2,
						},
					},
				},
				shouldReplaceResources: false,
			},
			want: &database.Application{
				ID:           "app-id",
				Name:         "app-name",
				Type:         "app-type",
				Labels:       "app-labels",
				Environments: "app-env",
				Resources: []database.Resource{
					{
						ID:                 database.CreateResourceID(resourceInfo1),
						Hash:               resourceInfo1.ResourceHash,
						Name:               resourceInfo1.ResourceName,
						Type:               resourceInfo1.ResourceType,
						ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer2"}),
						Packages: []database.Package{
							{
								ID:              database.CreatePackageID(pkgInfo1),
								Name:            pkgInfo1.Name,
								Version:         pkgInfo1.Version,
								License:         pkgInfo1.License,
								Language:        pkgInfo1.Language,
								Vulnerabilities: []database.Vulnerability{},
							},
						},
						CISDockerBenchmarkChecks: []database.CISDockerBenchmarkCheck{
							{
								ResourceID:   database.CreateResourceID(resourceInfo1),
								Code:         "code1",
								Level:        int(database.CISDockerBenchmarkLevelINFO),
								Descriptions: "desc1",
							},
							{
								ResourceID:   database.CreateResourceID(resourceInfo1),
								Code:         "code2",
								Level:        int(database.CISDockerBenchmarkLevelWARN),
								Descriptions: "desc2",
							},
						},
					},
					{
						ID:   database.CreateResourceID(resourceInfo2),
						Hash: resourceInfo2.ResourceHash,
						Name: resourceInfo2.ResourceName,
						Type: resourceInfo2.ResourceType,
						Packages: []database.Package{
							{
								ID:       database.CreatePackageID(pkgInfo1),
								Name:     pkgInfo1.Name,
								Version:  pkgInfo1.Version,
								License:  pkgInfo1.License,
								Language: pkgInfo1.Language,
								Vulnerabilities: []database.Vulnerability{
									database.CreateVulnerability(vulInfo1, &database.TransactionParams{}),
								},
							},
							{
								ID:       database.CreatePackageID(pkgInfo2),
								Name:     pkgInfo2.Name,
								Version:  pkgInfo2.Version,
								License:  pkgInfo2.License,
								Language: pkgInfo2.Language,
								Vulnerabilities: []database.Vulnerability{
									database.CreateVulnerability(vulInfo2, &database.TransactionParams{}),
								},
							},
						},
					},
				},
			},
			expectedTransactionParams: &database.TransactionParams{
				FixVersions: map[database.PkgVulID]string{
					database.CreatePkgVulID(database.CreatePackageID(pkgInfo1), database.CreateVulnerabilityID(vulInfo1)): "vul-fix-version",
					database.CreatePkgVulID(database.CreatePackageID(pkgInfo2), database.CreateVulnerabilityID(vulInfo2)): "vul-fix-version2",
				},
				Scanners: map[database.ResourcePkgID][]string{
					database.CreateResourcePkgID(database.CreateResourceID(resourceInfo2), database.CreatePackageID(pkgInfo1)): {"scanner1"},
					database.CreateResourcePkgID(database.CreateResourceID(resourceInfo2), database.CreatePackageID(pkgInfo2)): {"scanner2"},
				},
				Analyzers: map[database.ResourcePkgID][]string{
					database.CreateResourcePkgID(database.CreateResourceID(resourceInfo2), database.CreatePackageID(pkgInfo1)): {"scanner1"},
					database.CreateResourcePkgID(database.CreateResourceID(resourceInfo2), database.CreatePackageID(pkgInfo2)): {"scanner2"},
				},
			},
			mockResourceTableFunc: func(mock *database.MockResourceTable) {
				mock.EXPECT().
					GetDBResource(database.CreateResourceID(resourceInfo2), true).
					Return(nil, fmt.Errorf("resource not fount"))
			},
		},
		{
			name: "1 new resource + 1 only on current + should replace resources list",
			args: args{
				application: &database.Application{
					ID:           "app-id",
					Name:         "app-name",
					Type:         "app-type",
					Labels:       "app-labels",
					Environments: "app-env",
					Resources: []database.Resource{
						{
							ID:                 database.CreateResourceID(resourceInfo1),
							Hash:               resourceInfo1.ResourceHash,
							Name:               resourceInfo1.ResourceName,
							Type:               resourceInfo1.ResourceType,
							ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer2"}),
							Packages: []database.Package{
								{
									ID:              database.CreatePackageID(pkgInfo1),
									Name:            pkgInfo1.Name,
									Version:         pkgInfo1.Version,
									License:         pkgInfo1.License,
									Language:        pkgInfo1.Language,
									Vulnerabilities: []database.Vulnerability{},
								},
							},
							CISDockerBenchmarkChecks: []database.CISDockerBenchmarkCheck{
								{
									ResourceID:   database.CreateResourceID(resourceInfo1),
									Code:         "code1",
									Level:        int(database.CISDockerBenchmarkLevelINFO),
									Descriptions: "desc1",
								},
								{
									ResourceID:   database.CreateResourceID(resourceInfo1),
									Code:         "code2",
									Level:        int(database.CISDockerBenchmarkLevelWARN),
									Descriptions: "desc2",
								},
							},
						},
					},
				},
				applicationVulnerabilityScan: &types.ApplicationVulnerabilityScan{
					Resources: []*types.ResourceVulnerabilityScan{
						{
							PackageVulnerabilities: []*types.PackageVulnerabilityScan{
								vulInfo1,
								vulInfo2,
							},
							Resource: resourceInfo2,
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
						},
					},
				},
				shouldReplaceResources: true,
			},
			want: &database.Application{
				ID:           "app-id",
				Name:         "app-name",
				Type:         "app-type",
				Labels:       "app-labels",
				Environments: "app-env",
				Resources: []database.Resource{
					{
						ID:   database.CreateResourceID(resourceInfo2),
						Hash: resourceInfo2.ResourceHash,
						Name: resourceInfo2.ResourceName,
						Type: resourceInfo2.ResourceType,
						CISDockerBenchmarkChecks: []database.CISDockerBenchmarkCheck{
							{
								ResourceID:   database.CreateResourceID(resourceInfo2),
								Code:         "code1",
								Level:        int(database.CISDockerBenchmarkLevelINFO),
								Descriptions: "desc1",
							},
							{
								ResourceID:   database.CreateResourceID(resourceInfo2),
								Code:         "code2",
								Level:        int(database.CISDockerBenchmarkLevelWARN),
								Descriptions: "desc2",
							},
						},
						Packages: []database.Package{
							{
								ID:       database.CreatePackageID(pkgInfo1),
								Name:     pkgInfo1.Name,
								Version:  pkgInfo1.Version,
								License:  pkgInfo1.License,
								Language: pkgInfo1.Language,
								Vulnerabilities: []database.Vulnerability{
									database.CreateVulnerability(vulInfo1, &database.TransactionParams{}),
								},
							},
							{
								ID:       database.CreatePackageID(pkgInfo2),
								Name:     pkgInfo2.Name,
								Version:  pkgInfo2.Version,
								License:  pkgInfo2.License,
								Language: pkgInfo2.Language,
								Vulnerabilities: []database.Vulnerability{
									database.CreateVulnerability(vulInfo2, &database.TransactionParams{}),
								},
							},
						},
					},
				},
			},
			expectedTransactionParams: &database.TransactionParams{
				FixVersions: map[database.PkgVulID]string{
					database.CreatePkgVulID(database.CreatePackageID(pkgInfo1), database.CreateVulnerabilityID(vulInfo1)): "vul-fix-version",
					database.CreatePkgVulID(database.CreatePackageID(pkgInfo2), database.CreateVulnerabilityID(vulInfo2)): "vul-fix-version2",
				},
				Scanners: map[database.ResourcePkgID][]string{
					database.CreateResourcePkgID(database.CreateResourceID(resourceInfo2), database.CreatePackageID(pkgInfo1)): {"scanner1"},
					database.CreateResourcePkgID(database.CreateResourceID(resourceInfo2), database.CreatePackageID(pkgInfo2)): {"scanner2"},
				},
				Analyzers: map[database.ResourcePkgID][]string{
					// resourceInfo2 is new discover by scanners.
					database.CreateResourcePkgID(database.CreateResourceID(resourceInfo2), database.CreatePackageID(pkgInfo1)): {"scanner1"},
					database.CreateResourcePkgID(database.CreateResourceID(resourceInfo2), database.CreatePackageID(pkgInfo2)): {"scanner2"},
				},
			},
			mockResourceTableFunc: func(mock *database.MockResourceTable) {
				mock.EXPECT().
					GetDBResource(database.CreateResourceID(resourceInfo2), true).
					Return(nil, fmt.Errorf("resource not fount"))
			},
		},
		{
			name: "1 new resource on app not new in DB no new pkg + 1 only on current + should not replace resources list",
			args: args{
				application: &database.Application{
					ID:           "app-id",
					Name:         "app-name",
					Type:         "app-type",
					Labels:       "app-labels",
					Environments: "app-env",
					Resources: []database.Resource{
						{
							ID:                 database.CreateResourceID(resourceInfo1),
							Hash:               resourceInfo1.ResourceHash,
							Name:               resourceInfo1.ResourceName,
							Type:               resourceInfo1.ResourceType,
							ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer2"}),
							Packages: []database.Package{
								{
									ID:              database.CreatePackageID(pkgInfo1),
									Name:            pkgInfo1.Name,
									Version:         pkgInfo1.Version,
									License:         pkgInfo1.License,
									Language:        pkgInfo1.Language,
									Vulnerabilities: []database.Vulnerability{},
								},
							},
						},
					},
				},
				applicationVulnerabilityScan: &types.ApplicationVulnerabilityScan{
					Resources: []*types.ResourceVulnerabilityScan{
						{
							PackageVulnerabilities: []*types.PackageVulnerabilityScan{
								vulInfo1,
							},
							Resource: resourceInfo2,
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
						},
					},
				},
				shouldReplaceResources: false,
			},
			want: &database.Application{
				ID:           "app-id",
				Name:         "app-name",
				Type:         "app-type",
				Labels:       "app-labels",
				Environments: "app-env",
				Resources: []database.Resource{
					{
						ID:                 database.CreateResourceID(resourceInfo1),
						Hash:               resourceInfo1.ResourceHash,
						Name:               resourceInfo1.ResourceName,
						Type:               resourceInfo1.ResourceType,
						ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer2"}),
						Packages: []database.Package{
							{
								ID:              database.CreatePackageID(pkgInfo1),
								Name:            pkgInfo1.Name,
								Version:         pkgInfo1.Version,
								License:         pkgInfo1.License,
								Language:        pkgInfo1.Language,
								Vulnerabilities: []database.Vulnerability{},
							},
						},
					},
					{
						ID:   database.CreateResourceID(resourceInfo2),
						Hash: resourceInfo2.ResourceHash,
						Name: resourceInfo2.ResourceName,
						Type: resourceInfo2.ResourceType,
						CISDockerBenchmarkChecks: []database.CISDockerBenchmarkCheck{
							{
								ResourceID:   database.CreateResourceID(resourceInfo2),
								Code:         "code1",
								Level:        int(database.CISDockerBenchmarkLevelINFO),
								Descriptions: "desc1",
							},
							{
								ResourceID:   database.CreateResourceID(resourceInfo2),
								Code:         "code2",
								Level:        int(database.CISDockerBenchmarkLevelWARN),
								Descriptions: "desc2",
							},
						},
						Packages: []database.Package{
							{
								ID:       database.CreatePackageID(pkgInfo1),
								Name:     pkgInfo1.Name,
								Version:  pkgInfo1.Version,
								License:  pkgInfo1.License,
								Language: pkgInfo1.Language,
								Vulnerabilities: []database.Vulnerability{
									database.CreateVulnerability(vulInfo1, &database.TransactionParams{}),
								},
							},
							{
								ID:              database.CreatePackageID(pkgInfo2),
								Name:            pkgInfo2.Name,
								Version:         pkgInfo2.Version,
								License:         pkgInfo2.License,
								Language:        pkgInfo2.Language,
								Vulnerabilities: nil,
							},
						},
					},
				},
			},
			expectedTransactionParams: &database.TransactionParams{
				FixVersions: map[database.PkgVulID]string{
					database.CreatePkgVulID(database.CreatePackageID(pkgInfo1), database.CreateVulnerabilityID(vulInfo1)): "vul-fix-version",
				},
				Scanners: map[database.ResourcePkgID][]string{
					database.CreateResourcePkgID(database.CreateResourceID(resourceInfo2), database.CreatePackageID(pkgInfo1)): {"scanner1"},
				},
				Analyzers: map[database.ResourcePkgID][]string{},
			},
			mockResourceTableFunc: func(mock *database.MockResourceTable) {
				mock.EXPECT().
					GetDBResource(database.CreateResourceID(resourceInfo2), true).
					Return(&database.Resource{
						ID:   database.CreateResourceID(resourceInfo2),
						Hash: resourceInfo2.ResourceHash,
						Name: resourceInfo2.ResourceName,
						Type: resourceInfo2.ResourceType,
						Packages: []database.Package{
							// pkg to update on resource
							{
								ID:              database.CreatePackageID(pkgInfo1),
								Name:            pkgInfo1.Name,
								Version:         pkgInfo1.Version,
								License:         pkgInfo1.License,
								Language:        pkgInfo1.Language,
								Vulnerabilities: nil,
							},
							// pkg to keep with no changes
							{
								ID:              database.CreatePackageID(pkgInfo2),
								Name:            pkgInfo2.Name,
								Version:         pkgInfo2.Version,
								License:         pkgInfo2.License,
								Language:        pkgInfo2.Language,
								Vulnerabilities: nil,
							},
						},
					}, nil)
			},
		},
		{
			name: "1 new resource on app not new in DB no new pkg + 1 only on current + should replace resources list",
			args: args{
				application: &database.Application{
					ID:           "app-id",
					Name:         "app-name",
					Type:         "app-type",
					Labels:       "app-labels",
					Environments: "app-env",
					Resources: []database.Resource{
						{
							ID:                 database.CreateResourceID(resourceInfo1),
							Hash:               resourceInfo1.ResourceHash,
							Name:               resourceInfo1.ResourceName,
							Type:               resourceInfo1.ResourceType,
							ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer2"}),
							Packages: []database.Package{
								{
									ID:              database.CreatePackageID(pkgInfo1),
									Name:            pkgInfo1.Name,
									Version:         pkgInfo1.Version,
									License:         pkgInfo1.License,
									Language:        pkgInfo1.Language,
									Vulnerabilities: []database.Vulnerability{},
								},
							},
							CISDockerBenchmarkChecks: []database.CISDockerBenchmarkCheck{
								{
									ResourceID:   database.CreateResourceID(resourceInfo1),
									Code:         "code1",
									Level:        int(database.CISDockerBenchmarkLevelINFO),
									Descriptions: "desc1",
								},
								{
									ResourceID:   database.CreateResourceID(resourceInfo1),
									Code:         "code2",
									Level:        int(database.CISDockerBenchmarkLevelWARN),
									Descriptions: "desc2",
								},
							},
						},
					},
				},
				applicationVulnerabilityScan: &types.ApplicationVulnerabilityScan{
					Resources: []*types.ResourceVulnerabilityScan{
						{
							PackageVulnerabilities: []*types.PackageVulnerabilityScan{
								vulInfo1,
							},
							Resource: resourceInfo2,
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
						},
					},
				},
				shouldReplaceResources: true,
			},
			want: &database.Application{
				ID:           "app-id",
				Name:         "app-name",
				Type:         "app-type",
				Labels:       "app-labels",
				Environments: "app-env",
				Resources: []database.Resource{
					{
						ID:   database.CreateResourceID(resourceInfo2),
						Hash: resourceInfo2.ResourceHash,
						Name: resourceInfo2.ResourceName,
						Type: resourceInfo2.ResourceType,
						CISDockerBenchmarkChecks: []database.CISDockerBenchmarkCheck{
							{
								ResourceID:   database.CreateResourceID(resourceInfo2),
								Code:         "code1",
								Level:        int(database.CISDockerBenchmarkLevelINFO),
								Descriptions: "desc1",
							},
							{
								ResourceID:   database.CreateResourceID(resourceInfo2),
								Code:         "code2",
								Level:        int(database.CISDockerBenchmarkLevelWARN),
								Descriptions: "desc2",
							},
						},
						Packages: []database.Package{
							{
								ID:       database.CreatePackageID(pkgInfo1),
								Name:     pkgInfo1.Name,
								Version:  pkgInfo1.Version,
								License:  pkgInfo1.License,
								Language: pkgInfo1.Language,
								Vulnerabilities: []database.Vulnerability{
									database.CreateVulnerability(vulInfo1, &database.TransactionParams{}),
								},
							},
							{
								ID:              database.CreatePackageID(pkgInfo2),
								Name:            pkgInfo2.Name,
								Version:         pkgInfo2.Version,
								License:         pkgInfo2.License,
								Language:        pkgInfo2.Language,
								Vulnerabilities: nil,
							},
						},
					},
				},
			},
			expectedTransactionParams: &database.TransactionParams{
				FixVersions: map[database.PkgVulID]string{
					database.CreatePkgVulID(database.CreatePackageID(pkgInfo1), database.CreateVulnerabilityID(vulInfo1)): "vul-fix-version",
				},
				Scanners: map[database.ResourcePkgID][]string{
					database.CreateResourcePkgID(database.CreateResourceID(resourceInfo2), database.CreatePackageID(pkgInfo1)): {"scanner1"},
				},
				Analyzers: map[database.ResourcePkgID][]string{},
			},
			mockResourceTableFunc: func(mock *database.MockResourceTable) {
				mock.EXPECT().
					GetDBResource(database.CreateResourceID(resourceInfo2), true).
					Return(&database.Resource{
						ID:   database.CreateResourceID(resourceInfo2),
						Hash: resourceInfo2.ResourceHash,
						Name: resourceInfo2.ResourceName,
						Type: resourceInfo2.ResourceType,
						Packages: []database.Package{
							// pkg to update on resource
							{
								ID:              database.CreatePackageID(pkgInfo1),
								Name:            pkgInfo1.Name,
								Version:         pkgInfo1.Version,
								License:         pkgInfo1.License,
								Language:        pkgInfo1.Language,
								Vulnerabilities: nil,
							},
							// pkg to keep with no changes
							{
								ID:              database.CreatePackageID(pkgInfo2),
								Name:            pkgInfo2.Name,
								Version:         pkgInfo2.Version,
								License:         pkgInfo2.License,
								Language:        pkgInfo2.Language,
								Vulnerabilities: nil,
							},
						},
					}, nil)
			},
		},
		{
			name: "1 existing resource with new pkg + 1 only on current + should not replace resources list",
			args: args{
				application: &database.Application{
					ID:           "app-id",
					Name:         "app-name",
					Type:         "app-type",
					Labels:       "app-labels",
					Environments: "app-env",
					Resources: []database.Resource{
						// only on current
						{
							ID:                 database.CreateResourceID(resourceInfo1),
							Hash:               resourceInfo1.ResourceHash,
							Name:               resourceInfo1.ResourceName,
							Type:               resourceInfo1.ResourceType,
							ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer2"}),
							Packages: []database.Package{
								{
									ID:              database.CreatePackageID(pkgInfo1),
									Name:            pkgInfo1.Name,
									Version:         pkgInfo1.Version,
									License:         pkgInfo1.License,
									Language:        pkgInfo1.Language,
									Vulnerabilities: []database.Vulnerability{},
								},
							},
							CISDockerBenchmarkChecks: []database.CISDockerBenchmarkCheck{
								{
									ResourceID:   database.CreateResourceID(resourceInfo1),
									Code:         "code1",
									Level:        int(database.CISDockerBenchmarkLevelINFO),
									Descriptions: "desc1",
								},
								{
									ResourceID:   database.CreateResourceID(resourceInfo1),
									Code:         "code2",
									Level:        int(database.CISDockerBenchmarkLevelWARN),
									Descriptions: "desc2",
								},
							},
						},
						// need to update existing
						{
							ID:                 database.CreateResourceID(resourceInfo2),
							Hash:               resourceInfo2.ResourceHash,
							Name:               resourceInfo2.ResourceName,
							Type:               resourceInfo2.ResourceType,
							ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer2"}),
							Packages: []database.Package{
								{
									ID:              database.CreatePackageID(pkgInfo1),
									Name:            pkgInfo1.Name,
									Version:         pkgInfo1.Version,
									License:         pkgInfo1.License,
									Language:        pkgInfo1.Language,
									Vulnerabilities: nil,
								},
							},
							CISDockerBenchmarkChecks: []database.CISDockerBenchmarkCheck{
								{
									ResourceID:   database.CreateResourceID(resourceInfo2),
									Code:         "code1",
									Level:        int(database.CISDockerBenchmarkLevelINFO),
									Descriptions: "desc1",
								},
								{
									ResourceID:   database.CreateResourceID(resourceInfo2),
									Code:         "code2",
									Level:        int(database.CISDockerBenchmarkLevelWARN),
									Descriptions: "desc2",
								},
							},
						},
					},
				},
				applicationVulnerabilityScan: &types.ApplicationVulnerabilityScan{
					Resources: []*types.ResourceVulnerabilityScan{
						{
							PackageVulnerabilities: []*types.PackageVulnerabilityScan{
								vulInfo1,
								vulInfo2,
							},
							Resource: resourceInfo2,
							CisDockerBenchmarkResults: []*types.CISDockerBenchmarkResult{
								{
									Code:         "code1",
									Level:        int64(dockle_types.FatalLevel),
									Descriptions: "desc1",
								},
							},
						},
					},
				},
				shouldReplaceResources: false,
			},
			want: &database.Application{
				ID:           "app-id",
				Name:         "app-name",
				Type:         "app-type",
				Labels:       "app-labels",
				Environments: "app-env",
				Resources: []database.Resource{
					{
						ID:                 database.CreateResourceID(resourceInfo1),
						Hash:               resourceInfo1.ResourceHash,
						Name:               resourceInfo1.ResourceName,
						Type:               resourceInfo1.ResourceType,
						ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer2"}),
						Packages: []database.Package{
							{
								ID:              database.CreatePackageID(pkgInfo1),
								Name:            pkgInfo1.Name,
								Version:         pkgInfo1.Version,
								License:         pkgInfo1.License,
								Language:        pkgInfo1.Language,
								Vulnerabilities: []database.Vulnerability{},
							},
						},
						CISDockerBenchmarkChecks: []database.CISDockerBenchmarkCheck{
							{
								ResourceID:   database.CreateResourceID(resourceInfo1),
								Code:         "code1",
								Level:        int(database.CISDockerBenchmarkLevelINFO),
								Descriptions: "desc1",
							},
							{
								ResourceID:   database.CreateResourceID(resourceInfo1),
								Code:         "code2",
								Level:        int(database.CISDockerBenchmarkLevelWARN),
								Descriptions: "desc2",
							},
						},
					},
					{
						ID:                 database.CreateResourceID(resourceInfo2),
						Hash:               resourceInfo2.ResourceHash,
						Name:               resourceInfo2.ResourceName,
						Type:               resourceInfo2.ResourceType,
						ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer2"}),
						Packages: []database.Package{
							// existing pkg - vulnerabilities update
							{
								ID:       database.CreatePackageID(pkgInfo1),
								Name:     pkgInfo1.Name,
								Version:  pkgInfo1.Version,
								License:  pkgInfo1.License,
								Language: pkgInfo1.Language,
								Vulnerabilities: []database.Vulnerability{
									database.CreateVulnerability(vulInfo1, &database.TransactionParams{}),
								},
							},
							// new pkg
							{
								ID:       database.CreatePackageID(pkgInfo2),
								Name:     pkgInfo2.Name,
								Version:  pkgInfo2.Version,
								License:  pkgInfo2.License,
								Language: pkgInfo2.Language,
								Vulnerabilities: []database.Vulnerability{
									database.CreateVulnerability(vulInfo2, &database.TransactionParams{}),
								},
							},
						},
						// CISDockerBenchmarkChecks updated.
						CISDockerBenchmarkChecks: []database.CISDockerBenchmarkCheck{
							{
								ResourceID:   database.CreateResourceID(resourceInfo2),
								Code:         "code1",
								Level:        int(database.CISDockerBenchmarkLevelFATAL),
								Descriptions: "desc1",
							},
						},
					},
				},
			},
			expectedTransactionParams: &database.TransactionParams{
				FixVersions: map[database.PkgVulID]string{
					database.CreatePkgVulID(database.CreatePackageID(pkgInfo1), database.CreateVulnerabilityID(vulInfo1)): "vul-fix-version",
					database.CreatePkgVulID(database.CreatePackageID(pkgInfo2), database.CreateVulnerabilityID(vulInfo2)): "vul-fix-version2",
				},
				Scanners: map[database.ResourcePkgID][]string{
					database.CreateResourcePkgID(database.CreateResourceID(resourceInfo2), database.CreatePackageID(pkgInfo1)): {"scanner1"},
					database.CreateResourcePkgID(database.CreateResourceID(resourceInfo2), database.CreatePackageID(pkgInfo2)): {"scanner2"},
				},
				Analyzers: map[database.ResourcePkgID][]string{
					database.CreateResourcePkgID(database.CreateResourceID(resourceInfo2), database.CreatePackageID(pkgInfo2)): {"scanner2"},
				},
			},
		},
		{
			name: "1 existing resource with new pkg + 1 only on current + should replace resources list",
			args: args{
				application: &database.Application{
					ID:           "app-id",
					Name:         "app-name",
					Type:         "app-type",
					Labels:       "app-labels",
					Environments: "app-env",
					Resources: []database.Resource{
						// only on current
						{
							ID:                 database.CreateResourceID(resourceInfo1),
							Hash:               resourceInfo1.ResourceHash,
							Name:               resourceInfo1.ResourceName,
							Type:               resourceInfo1.ResourceType,
							ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer2"}),
							Packages: []database.Package{
								{
									ID:              database.CreatePackageID(pkgInfo1),
									Name:            pkgInfo1.Name,
									Version:         pkgInfo1.Version,
									License:         pkgInfo1.License,
									Language:        pkgInfo1.Language,
									Vulnerabilities: []database.Vulnerability{},
								},
							},
						},
						// need to update existing
						{
							ID:                 database.CreateResourceID(resourceInfo2),
							Hash:               resourceInfo2.ResourceHash,
							Name:               resourceInfo2.ResourceName,
							Type:               resourceInfo2.ResourceType,
							ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer2"}),
							Packages: []database.Package{
								{
									ID:              database.CreatePackageID(pkgInfo1),
									Name:            pkgInfo1.Name,
									Version:         pkgInfo1.Version,
									License:         pkgInfo1.License,
									Language:        pkgInfo1.Language,
									Vulnerabilities: nil,
								},
							},
							CISDockerBenchmarkChecks: []database.CISDockerBenchmarkCheck{
								{
									ResourceID:   database.CreateResourceID(resourceInfo2),
									Code:         "code1",
									Level:        int(database.CISDockerBenchmarkLevelINFO),
									Descriptions: "desc1",
								},
								{
									ResourceID:   database.CreateResourceID(resourceInfo2),
									Code:         "code2",
									Level:        int(database.CISDockerBenchmarkLevelWARN),
									Descriptions: "desc2",
								},
							},
						},
					},
				},
				applicationVulnerabilityScan: &types.ApplicationVulnerabilityScan{
					Resources: []*types.ResourceVulnerabilityScan{
						{
							PackageVulnerabilities: []*types.PackageVulnerabilityScan{
								vulInfo1,
								vulInfo2,
							},
							Resource: resourceInfo2,
							CisDockerBenchmarkResults: []*types.CISDockerBenchmarkResult{
								{
									Code:         "code1",
									Level:        int64(dockle_types.FatalLevel),
									Descriptions: "desc1",
								},
							},
						},
					},
				},
				shouldReplaceResources: true,
			},
			want: &database.Application{
				ID:           "app-id",
				Name:         "app-name",
				Type:         "app-type",
				Labels:       "app-labels",
				Environments: "app-env",
				Resources: []database.Resource{
					{
						ID:                 database.CreateResourceID(resourceInfo2),
						Hash:               resourceInfo2.ResourceHash,
						Name:               resourceInfo2.ResourceName,
						Type:               resourceInfo2.ResourceType,
						ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer2"}),
						Packages: []database.Package{
							// existing pkg - vulnerabilities update
							{
								ID:       database.CreatePackageID(pkgInfo1),
								Name:     pkgInfo1.Name,
								Version:  pkgInfo1.Version,
								License:  pkgInfo1.License,
								Language: pkgInfo1.Language,
								Vulnerabilities: []database.Vulnerability{
									database.CreateVulnerability(vulInfo1, &database.TransactionParams{}),
								},
							},
							// new pkg
							{
								ID:       database.CreatePackageID(pkgInfo2),
								Name:     pkgInfo2.Name,
								Version:  pkgInfo2.Version,
								License:  pkgInfo2.License,
								Language: pkgInfo2.Language,
								Vulnerabilities: []database.Vulnerability{
									database.CreateVulnerability(vulInfo2, &database.TransactionParams{}),
								},
							},
						},
						// CISDockerBenchmarkChecks updated.
						CISDockerBenchmarkChecks: []database.CISDockerBenchmarkCheck{
							{
								ResourceID:   database.CreateResourceID(resourceInfo2),
								Code:         "code1",
								Level:        int(database.CISDockerBenchmarkLevelFATAL),
								Descriptions: "desc1",
							},
						},
					},
				},
			},
			expectedTransactionParams: &database.TransactionParams{
				FixVersions: map[database.PkgVulID]string{
					database.CreatePkgVulID(database.CreatePackageID(pkgInfo1), database.CreateVulnerabilityID(vulInfo1)): "vul-fix-version",
					database.CreatePkgVulID(database.CreatePackageID(pkgInfo2), database.CreateVulnerabilityID(vulInfo2)): "vul-fix-version2",
				},
				Scanners: map[database.ResourcePkgID][]string{
					database.CreateResourcePkgID(database.CreateResourceID(resourceInfo2), database.CreatePackageID(pkgInfo1)): {"scanner1"},
					database.CreateResourcePkgID(database.CreateResourceID(resourceInfo2), database.CreatePackageID(pkgInfo2)): {"scanner2"},
				},
				Analyzers: map[database.ResourcePkgID][]string{
					// only resourceInfo2 + pkgInfo2 is new discover by scanners.
					database.CreateResourcePkgID(database.CreateResourceID(resourceInfo2), database.CreatePackageID(pkgInfo2)): {"scanner2"},
				},
			},
		},
		{
			name: "1 existing resource with new pkg + 1 only on current + should replace resources list + no cis update exist",
			args: args{
				application: &database.Application{
					ID:           "app-id",
					Name:         "app-name",
					Type:         "app-type",
					Labels:       "app-labels",
					Environments: "app-env",
					Resources: []database.Resource{
						// only on current
						{
							ID:                 database.CreateResourceID(resourceInfo1),
							Hash:               resourceInfo1.ResourceHash,
							Name:               resourceInfo1.ResourceName,
							Type:               resourceInfo1.ResourceType,
							ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer2"}),
							Packages: []database.Package{
								{
									ID:              database.CreatePackageID(pkgInfo1),
									Name:            pkgInfo1.Name,
									Version:         pkgInfo1.Version,
									License:         pkgInfo1.License,
									Language:        pkgInfo1.Language,
									Vulnerabilities: []database.Vulnerability{},
								},
							},
						},
						// need to update existing
						{
							ID:                 database.CreateResourceID(resourceInfo2),
							Hash:               resourceInfo2.ResourceHash,
							Name:               resourceInfo2.ResourceName,
							Type:               resourceInfo2.ResourceType,
							ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer2"}),
							Packages: []database.Package{
								{
									ID:              database.CreatePackageID(pkgInfo1),
									Name:            pkgInfo1.Name,
									Version:         pkgInfo1.Version,
									License:         pkgInfo1.License,
									Language:        pkgInfo1.Language,
									Vulnerabilities: nil,
								},
							},
							CISDockerBenchmarkChecks: []database.CISDockerBenchmarkCheck{
								{
									ResourceID:   database.CreateResourceID(resourceInfo2),
									Code:         "code1",
									Level:        int(database.CISDockerBenchmarkLevelINFO),
									Descriptions: "desc1",
								},
								{
									ResourceID:   database.CreateResourceID(resourceInfo2),
									Code:         "code2",
									Level:        int(database.CISDockerBenchmarkLevelWARN),
									Descriptions: "desc2",
								},
							},
						},
					},
				},
				applicationVulnerabilityScan: &types.ApplicationVulnerabilityScan{
					Resources: []*types.ResourceVulnerabilityScan{
						{
							PackageVulnerabilities: []*types.PackageVulnerabilityScan{
								vulInfo1,
								vulInfo2,
							},
							Resource: resourceInfo2,
							// CisDockerBenchmarkResults is missing - no need to update.
						},
					},
				},
				shouldReplaceResources: true,
			},
			want: &database.Application{
				ID:           "app-id",
				Name:         "app-name",
				Type:         "app-type",
				Labels:       "app-labels",
				Environments: "app-env",
				Resources: []database.Resource{
					{
						ID:                 database.CreateResourceID(resourceInfo2),
						Hash:               resourceInfo2.ResourceHash,
						Name:               resourceInfo2.ResourceName,
						Type:               resourceInfo2.ResourceType,
						ReportingAnalyzers: database.ArrayToDBArray([]string{"analyzer1", "analyzer2"}),
						Packages: []database.Package{
							// existing pkg - vulnerabilities update
							{
								ID:       database.CreatePackageID(pkgInfo1),
								Name:     pkgInfo1.Name,
								Version:  pkgInfo1.Version,
								License:  pkgInfo1.License,
								Language: pkgInfo1.Language,
								Vulnerabilities: []database.Vulnerability{
									database.CreateVulnerability(vulInfo1, &database.TransactionParams{}),
								},
							},
							// new pkg
							{
								ID:       database.CreatePackageID(pkgInfo2),
								Name:     pkgInfo2.Name,
								Version:  pkgInfo2.Version,
								License:  pkgInfo2.License,
								Language: pkgInfo2.Language,
								Vulnerabilities: []database.Vulnerability{
									database.CreateVulnerability(vulInfo2, &database.TransactionParams{}),
								},
							},
						},
						// CISDockerBenchmarkChecks is missing, no need to update.
						CISDockerBenchmarkChecks: []database.CISDockerBenchmarkCheck{
							{
								ResourceID:   database.CreateResourceID(resourceInfo2),
								Code:         "code1",
								Level:        int(database.CISDockerBenchmarkLevelINFO),
								Descriptions: "desc1",
							},
							{
								ResourceID:   database.CreateResourceID(resourceInfo2),
								Code:         "code2",
								Level:        int(database.CISDockerBenchmarkLevelWARN),
								Descriptions: "desc2",
							},
						},
					},
				},
			},
			expectedTransactionParams: &database.TransactionParams{
				FixVersions: map[database.PkgVulID]string{
					database.CreatePkgVulID(database.CreatePackageID(pkgInfo1), database.CreateVulnerabilityID(vulInfo1)): "vul-fix-version",
					database.CreatePkgVulID(database.CreatePackageID(pkgInfo2), database.CreateVulnerabilityID(vulInfo2)): "vul-fix-version2",
				},
				Scanners: map[database.ResourcePkgID][]string{
					database.CreateResourcePkgID(database.CreateResourceID(resourceInfo2), database.CreatePackageID(pkgInfo1)): {"scanner1"},
					database.CreateResourcePkgID(database.CreateResourceID(resourceInfo2), database.CreatePackageID(pkgInfo2)): {"scanner2"},
				},
				Analyzers: map[database.ResourcePkgID][]string{
					// only resourceInfo2 + pkgInfo2 is new discover by scanners.
					database.CreateResourcePkgID(database.CreateResourceID(resourceInfo2), database.CreatePackageID(pkgInfo2)): {"scanner2"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResourceTable := database.NewMockResourceTable(mockCtrl)
			if tt.mockResourceTableFunc != nil {
				tt.mockResourceTableFunc(mockResourceTable)
			}
			s := Server{
				dbHandler: &database.MockHandler{
					MockResourceTable: mockResourceTable,
				},
			}
			transactionParams := &database.TransactionParams{
				FixVersions: make(map[database.PkgVulID]string),        // will be populated during object creation
				Analyzers:   make(map[database.ResourcePkgID][]string), // will be populated during object creation
				Scanners:    make(map[database.ResourcePkgID][]string), // will be populated during object creation
			}
			got := s.updateApplicationWithVulnerabilityScan(tt.args.application, tt.args.applicationVulnerabilityScan,
				transactionParams, tt.args.shouldReplaceResources)
			got.Resources = sortResources(got.Resources)
			tt.want.Resources = sortResources(tt.want.Resources)
			assert.DeepEqual(t, got, tt.want)
			assert.DeepEqual(t, transactionParams, tt.expectedTransactionParams)
		})
	}
}

func sortResources(resources []database.Resource) []database.Resource {
	sort.Slice(resources, func(i, j int) bool {
		resources[i].Packages = sortPackages(resources[i].Packages)
		resources[j].Packages = sortPackages(resources[j].Packages)
		return resources[i].Name < resources[j].Name
	})

	return resources
}

func sortPackages(packages []database.Package) []database.Package {
	sort.Slice(packages, func(i, j int) bool {
		packages[i].Vulnerabilities = sortVulnerabilities(packages[i].Vulnerabilities)
		packages[j].Vulnerabilities = sortVulnerabilities(packages[j].Vulnerabilities)
		return packages[i].Name < packages[j].Name
	})
	return packages
}

func sortVulnerabilities(vulnerabilities []database.Vulnerability) []database.Vulnerability {
	sort.Slice(vulnerabilities, func(i, j int) bool {
		return vulnerabilities[i].Name < vulnerabilities[j].Name
	})
	return vulnerabilities
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
