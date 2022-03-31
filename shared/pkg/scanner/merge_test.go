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

package scanner

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/yudai/gojsondiff/formatter"
	"gotest.tools/assert"
)

func Test_handleVulnerabilityWithExistingKey(t *testing.T) {
	highVul := Vulnerability{
		ID:       "highVul",
		Severity: "HIGH",
		Package: Package{
			Name:    "pkg-name",
			Version: "pkg-version",
		},
	}
	lowVul := Vulnerability{
		ID:       "highVul",
		Severity: "LOW",
		Package: Package{
			Name:    "pkg-name",
			Version: "pkg-version",
		},
	}
	medVul := Vulnerability{
		ID:       "highVul",
		Severity: "MEDIUM",
		Package: Package{
			Name:    "pkg-name",
			Version: "pkg-version",
		},
	}
	type args struct {
		mergedVulnerabilities []MergedVulnerability
		otherVulnerability    Vulnerability
		otherScannerInfo      Info
	}
	tests := []struct {
		name                       string
		args                       args
		want                       []MergedVulnerability
		patchMergedVulnerabilityID func([]MergedVulnerability) []MergedVulnerability
	}{
		{
			name: "identical vulnerability",
			args: args{
				mergedVulnerabilities: []MergedVulnerability{
					{
						ID:            "1",
						Vulnerability: highVul,
						ScannersInfo: []Info{
							{
								Name: "scanner1",
							},
						},
					},
				},
				otherVulnerability: highVul,
				otherScannerInfo: Info{
					Name: "scanner2",
				},
			},
			want: []MergedVulnerability{
				{
					ID:            "1",
					Vulnerability: highVul,
					ScannersInfo: []Info{
						{
							Name: "scanner1",
						},
						{
							Name: "scanner2",
						},
					},
				},
			},
		},
		{
			name: "different vulnerability",
			args: args{
				mergedVulnerabilities: []MergedVulnerability{
					{
						ID:            "1",
						Vulnerability: highVul,
						ScannersInfo: []Info{
							{
								Name: "scanner1",
							},
						},
					},
				},
				otherVulnerability: lowVul,
				otherScannerInfo: Info{
					Name: "scanner2",
				},
			},
			want: []MergedVulnerability{
				{
					ID:            "1",
					Vulnerability: highVul,
					ScannersInfo: []Info{
						{
							Name: "scanner1",
						},
					},
				},
				{
					ID:            "2",
					Vulnerability: lowVul,
					ScannersInfo: []Info{
						{
							Name: "scanner2",
						},
					},
					Diffs: []DiffInfo{
						{
							JSONDiff: map[string]interface{}{
								"severity": []interface{}{"HIGH", "LOW"},
							},
							CompareToID: "1",
						},
					},
				},
			},
			patchMergedVulnerabilityID: func(vulnerability []MergedVulnerability) []MergedVulnerability {
				vulnerability[1].ID = "2"
				return vulnerability
			},
		},
		{
			name: "different vulnerability from first identical to second",
			args: args{
				mergedVulnerabilities: []MergedVulnerability{
					{
						ID:            "1",
						Vulnerability: highVul,
						ScannersInfo: []Info{
							{
								Name: "scanner1",
							},
						},
					},
					{
						ID:            "2",
						Vulnerability: lowVul,
						ScannersInfo: []Info{
							{
								Name: "scanner2",
							},
						},
						Diffs: []DiffInfo{
							{
								JSONDiff: map[string]interface{}{
									"severity": []interface{}{"HIGH", "LOW"},
								},
								CompareToID: "1",
							},
						},
					},
				},
				otherVulnerability: lowVul,
				otherScannerInfo: Info{
					Name: "scanner3",
				},
			},
			want: []MergedVulnerability{
				{
					ID:            "1",
					Vulnerability: highVul,
					ScannersInfo: []Info{
						{
							Name: "scanner1",
						},
					},
				},
				{
					ID:            "2",
					Vulnerability: lowVul,
					ScannersInfo: []Info{
						{
							Name: "scanner2",
						},
						{
							Name: "scanner3",
						},
					},
					Diffs: []DiffInfo{
						{
							JSONDiff: map[string]interface{}{
								"severity": []interface{}{"HIGH", "LOW"},
							},
							CompareToID: "1",
						},
					},
				},
			},
		},
		{
			name: "different vulnerability from all",
			args: args{
				mergedVulnerabilities: []MergedVulnerability{
					{
						ID:            "1",
						Vulnerability: highVul,
						ScannersInfo: []Info{
							{
								Name: "scanner1",
							},
						},
					},
					{
						ID:            "2",
						Vulnerability: lowVul,
						ScannersInfo: []Info{
							{
								Name: "scanner2",
							},
						},
						Diffs: []DiffInfo{
							{
								JSONDiff: map[string]interface{}{
									"severity": []interface{}{"HIGH", "LOW"},
								},
								CompareToID: "1",
							},
						},
					},
				},
				otherVulnerability: medVul,
				otherScannerInfo: Info{
					Name: "scanner3",
				},
			},
			want: []MergedVulnerability{
				{
					ID:            "1",
					Vulnerability: highVul,
					ScannersInfo: []Info{
						{
							Name: "scanner1",
						},
					},
				},
				{
					ID:            "2",
					Vulnerability: lowVul,
					ScannersInfo: []Info{
						{
							Name: "scanner2",
						},
					},
					Diffs: []DiffInfo{
						{
							JSONDiff: map[string]interface{}{
								"severity": []interface{}{"HIGH", "LOW"},
							},
							CompareToID: "1",
						},
					},
				},
				{
					ID:            "3",
					Vulnerability: medVul,
					ScannersInfo: []Info{
						{
							Name: "scanner3",
						},
					},
					Diffs: []DiffInfo{
						{
							JSONDiff: map[string]interface{}{
								"severity": []interface{}{"HIGH", "MEDIUM"},
							},
							CompareToID: "1",
						},
						{
							JSONDiff: map[string]interface{}{
								"severity": []interface{}{"LOW", "MEDIUM"},
							},
							CompareToID: "2",
						},
					},
				},
			},
			patchMergedVulnerabilityID: func(vulnerability []MergedVulnerability) []MergedVulnerability {
				vulnerability[2].ID = "3"
				return vulnerability
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handleVulnerabilityWithExistingKey(tt.args.mergedVulnerabilities, tt.args.otherVulnerability, tt.args.otherScannerInfo)
			if tt.patchMergedVulnerabilityID != nil {
				got = tt.patchMergedVulnerabilityID(got)
			}
			assert.DeepEqual(t, got, tt.want, cmpopts.IgnoreTypes(DiffInfo{}.ASCIIDiff))
		})
	}
}

func Test_getDiff(t *testing.T) {
	type args struct {
		vulnerability          Vulnerability
		compareToVulnerability Vulnerability
		compareToID            string
	}
	tests := []struct {
		name    string
		args    args
		want    *DiffInfo
		wantErr bool
	}{
		{
			name: "diff in fix",
			args: args{
				vulnerability: Vulnerability{
					ID: "id",
					Fix: Fix{
						Versions: []string{"1", "3"},
						State:    "not fixed",
					},
				},
				compareToVulnerability: Vulnerability{
					ID: "id",
					Fix: Fix{
						Versions: []string{"1", "2"},
						State:    "fixed",
					},
				},
				compareToID: "compareToID",
			},
			want: &DiffInfo{
				CompareToID: "compareToID",
				JSONDiff: map[string]interface{}{
					"fix": map[string]interface{}{
						"state": []interface{}{"fixed", "not fixed"},
						"versions": map[string]interface{}{
							"1":  []interface{}{"2", "3"}, // diff in array index 1
							"_t": "a",                     // sign for delta json
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "diff in links",
			args: args{
				vulnerability: Vulnerability{
					ID:    "id",
					Links: []string{"link1", "link2"},
				},
				compareToVulnerability: Vulnerability{
					ID:    "id",
					Links: []string{"link1", "link3", "link4"},
				},
				compareToID: "compareToID",
			},
			want: &DiffInfo{
				CompareToID: "compareToID",
				JSONDiff: map[string]interface{}{
					"links": map[string]interface{}{
						"1": []interface{}{"link4", "link2"},
						// "_1" means object was deleted from index 1
						// more info in github.com/yudai/gojsondiff@v1.0.0/formatter/delta.go
						"_1": []interface{}{"link3", 0, formatter.DeltaDelete},
						"_t": "a", // sign for delta json
					},
				},
				// ASCIIDiff: "{\n   \"cvss\": null,\n   \"distro\": {\n     \"idLike\": null,\n     \"name\": \"\",\n     \"version\": \"\"\n   },\n   \"fix\": {\n     \"state\": \"\",\n     \"versions\": null\n   },\n   \"id\": \"id\",\n   \"layerID\": \"\",\n   \"links\": [\n     0: \"link1\",\n-    1: \"link4\",\n+    1: \"link2\",\n-    1: \"link3\"\n     2: \"link4\"\n   ],\n   \"package\": {\n     \"cpes\": null,\n     \"language\": \"\",\n     \"licenses\": null,\n     \"name\": \"\",\n     \"purl\": \"\",\n     \"type\": \"\",\n     \"version\": \"\"\n   },\n   \"path\": \"\"\n }\n        ",
			},
			wantErr: false,
		},
		{
			name: "no diff - CVSS sort is needed",
			args: args{
				vulnerability: Vulnerability{
					CVSS: []CVSS{
						{
							Version: "3",
							Vector:  "456",
						},
						{
							Version: "2",
							Vector:  "123",
						},
					},
				},
				compareToVulnerability: Vulnerability{
					CVSS: []CVSS{
						{
							Version:        "2",
							Vector:         "123",
							Metrics:        CvssMetrics{},
							VendorMetadata: nil,
						},
						{
							Version:        "3",
							Vector:         "456",
							Metrics:        CvssMetrics{},
							VendorMetadata: nil,
						},
					},
				},
				compareToID: "compareToID",
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getDiff(tt.args.vulnerability, tt.args.compareToVulnerability, tt.args.compareToID)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDiff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.DeepEqual(t, got, tt.want, cmpopts.IgnoreTypes(DiffInfo{}.ASCIIDiff))
		})
	}
}

func Test_sortArrays(t *testing.T) {
	type args struct {
		vulnerability Vulnerability
	}
	tests := []struct {
		name string
		args args
		want Vulnerability
	}{
		{
			name: "sort",
			args: args{
				vulnerability: Vulnerability{
					Links: []string{"link2", "link1"},
					CVSS: []CVSS{
						{
							Version: "3",
							Vector:  "456",
						},
						{
							Version: "2",
							Vector:  "123",
						},
					},
					Fix: Fix{
						Versions: []string{"ver2", "ver1"},
					},
					Package: Package{
						Licenses: []string{"lic2", "lic1"},
						CPEs:     []string{"cpes2", "cpes1"},
					},
				},
			},
			want: Vulnerability{
				Links: []string{"link1", "link2"},
				CVSS: []CVSS{
					{
						Version: "2",
						Vector:  "123",
					},
					{
						Version: "3",
						Vector:  "456",
					},
				},
				Fix: Fix{
					Versions: []string{"ver1", "ver2"},
				},
				Package: Package{
					Licenses: []string{"lic1", "lic2"},
					CPEs:     []string{"cpes1", "cpes2"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sortArrays(tt.args.vulnerability); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortArrays() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergedResults_Merge(t *testing.T) {
	vul := Vulnerability{
		ID:       "id1",
		Severity: "HIGH",
		Package: Package{
			Name:    "pkg-name",
			Version: "pkg-version",
		},
	}
	sameVulDifferentSeverity := Vulnerability{
		ID:       "id1",
		Severity: "LOW",
		Package: Package{
			Name:    "pkg-name",
			Version: "pkg-version",
		},
	}
	differentVulID := Vulnerability{
		ID:       "id2",
		Severity: "HIGH",
		Package: Package{
			Name:    "pkg-name",
			Version: "pkg-version",
		},
	}
	type fields struct {
		MergedVulnerabilities map[VulnerabilityKey][]MergedVulnerability
	}
	type args struct {
		other *Results
	}
	tests := []struct {
		name                       string
		fields                     fields
		args                       args
		want                       *MergedResults
		patchMergedVulnerabilityID func(map[VulnerabilityKey][]MergedVulnerability) map[VulnerabilityKey][]MergedVulnerability
	}{
		{
			name: "all are non mutual vulnerabilities",
			fields: fields{
				MergedVulnerabilities: NewMergedResults().MergedVulnerabilitiesByKey,
			},
			args: args{
				other: &Results{
					Matches: Matches{
						{
							Vulnerability: vul,
						},
						{
							Vulnerability: differentVulID,
						},
					},
					ScannerInfo: Info{
						Name: "scanner1",
					},
				},
			},
			want: &MergedResults{
				MergedVulnerabilitiesByKey: map[VulnerabilityKey][]MergedVulnerability{
					createVulnerabilityKey(vul): {
						{
							ID:            "0",
							Vulnerability: vul,
							ScannersInfo: []Info{
								{
									Name: "scanner1",
								},
							},
						},
					},
					createVulnerabilityKey(differentVulID): {
						{
							ID:            "1",
							Vulnerability: differentVulID,
							ScannersInfo: []Info{
								{
									Name: "scanner1",
								},
							},
						},
					},
				},
			},
			patchMergedVulnerabilityID: func(v map[VulnerabilityKey][]MergedVulnerability) map[VulnerabilityKey][]MergedVulnerability {
				v[createVulnerabilityKey(vul)][0].ID = "0"
				v[createVulnerabilityKey(differentVulID)][0].ID = "1"
				return v
			},
		},
		{
			name: "1 non mutual vulnerability and 1 mutual vulnerability with no diff",
			fields: fields{
				MergedVulnerabilities: map[VulnerabilityKey][]MergedVulnerability{
					createVulnerabilityKey(vul): {
						{
							ID:            "0",
							Vulnerability: vul,
							ScannersInfo: []Info{
								{
									Name: "scanner1",
								},
							},
						},
					},
				},
			},
			args: args{
				other: &Results{
					Matches: Matches{
						{
							Vulnerability: vul,
						},
						{
							Vulnerability: differentVulID,
						},
					},
					ScannerInfo: Info{
						Name: "scanner2",
					},
				},
			},
			want: &MergedResults{
				MergedVulnerabilitiesByKey: map[VulnerabilityKey][]MergedVulnerability{
					createVulnerabilityKey(vul): {
						{
							ID:            "0",
							Vulnerability: vul,
							ScannersInfo: []Info{
								{
									Name: "scanner1",
								},
								{
									Name: "scanner2",
								},
							},
						},
					},
					createVulnerabilityKey(differentVulID): {
						{
							ID:            "1",
							Vulnerability: differentVulID,
							ScannersInfo: []Info{
								{
									Name: "scanner2",
								},
							},
						},
					},
				},
			},
			patchMergedVulnerabilityID: func(v map[VulnerabilityKey][]MergedVulnerability) map[VulnerabilityKey][]MergedVulnerability {
				v[createVulnerabilityKey(vul)][0].ID = "0"
				v[createVulnerabilityKey(differentVulID)][0].ID = "1"
				return v
			},
		},
		{
			name: "1 non mutual vulnerability and 1 mutual vulnerability with diff",
			fields: fields{
				MergedVulnerabilities: map[VulnerabilityKey][]MergedVulnerability{
					createVulnerabilityKey(vul): {
						{
							ID:            "0",
							Vulnerability: vul,
							ScannersInfo: []Info{
								{
									Name: "scanner1",
								},
							},
						},
					},
				},
			},
			args: args{
				other: &Results{
					Matches: Matches{
						{
							Vulnerability: sameVulDifferentSeverity, // mutual vulnerability with diff
						},
						{
							Vulnerability: differentVulID, // non mutual
						},
					},
					ScannerInfo: Info{
						Name: "scanner2",
					},
				},
			},
			want: &MergedResults{
				MergedVulnerabilitiesByKey: map[VulnerabilityKey][]MergedVulnerability{
					createVulnerabilityKey(vul): {
						{
							ID:            "0",
							Vulnerability: vul,
							ScannersInfo: []Info{
								{
									Name: "scanner1",
								},
							},
						},
						{
							ID:            "1",
							Vulnerability: sameVulDifferentSeverity,
							ScannersInfo: []Info{
								{
									Name: "scanner2",
								},
							},
							Diffs: []DiffInfo{
								{
									JSONDiff: map[string]interface{}{
										"severity": []interface{}{"HIGH", "LOW"},
									},
									CompareToID: "0",
								},
							},
						},
					},
					createVulnerabilityKey(differentVulID): {
						{
							ID:            "0",
							Vulnerability: differentVulID,
							ScannersInfo: []Info{
								{
									Name: "scanner2",
								},
							},
						},
					},
				},
			},
			patchMergedVulnerabilityID: func(v map[VulnerabilityKey][]MergedVulnerability) map[VulnerabilityKey][]MergedVulnerability {
				v[createVulnerabilityKey(vul)][0].ID = "0"
				v[createVulnerabilityKey(vul)][1].ID = "1"
				v[createVulnerabilityKey(differentVulID)][0].ID = "0"
				return v
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MergedResults{
				MergedVulnerabilitiesByKey: tt.fields.MergedVulnerabilities,
			}
			got := m.Merge(tt.args.other)
			if tt.patchMergedVulnerabilityID != nil {
				got.MergedVulnerabilitiesByKey = tt.patchMergedVulnerabilityID(got.MergedVulnerabilitiesByKey)
			}
			assert.DeepEqual(t, got, tt.want, cmpopts.IgnoreTypes(DiffInfo{}.ASCIIDiff))
		})
	}
}
