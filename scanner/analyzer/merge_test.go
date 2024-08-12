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

package analyzer

import (
	"reflect"
	"sort"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/openclarity/vmclarity/scanner/utils"
)

func TestMergedResults_createComponentListFromMap(t *testing.T) {
	type fields struct {
		MergedComponentByKey map[componentKey]*MergedComponent
		Source               utils.SourceType
		SrcMetaData          *cdx.Metadata
	}
	tests := []struct {
		name   string
		fields fields
		want   *[]cdx.Component
	}{
		{
			name: "create list from map",
			fields: fields{
				MergedComponentByKey: map[componentKey]*MergedComponent{
					"1": {
						Component: cdx.Component{
							Name: "1",
						},
					},
					"2": {
						Component: cdx.Component{
							Name: "2",
						},
					},
				},
			},
			want: &[]cdx.Component{
				{
					Name: "1",
				},
				{
					Name: "2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MergedResults{
				MergedComponentByKey: tt.fields.MergedComponentByKey,
				Source:               tt.fields.Source,
				SrcMetaData:          tt.fields.SrcMetaData,
			}
			got := m.createComponentListFromMap()
			sort.Slice(*got, func(i, j int) bool {
				return (*got)[i].Name < (*got)[j].Name
			})
			sort.Slice(*tt.want, func(i, j int) bool {
				return (*tt.want)[i].Name < (*tt.want)[j].Name
			})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createComponentListFromMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergedComponent_appendAnalyzerInfo(t *testing.T) {
	type fields struct {
		Component    cdx.Component
		AnalyzerInfo []string
	}
	type args struct {
		info string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *MergedComponent
	}{
		{
			name: "append empty analyzer info",
			args: args{
				info: "syft",
			},
			want: &MergedComponent{
				Component: cdx.Component{
					Properties: &[]cdx.Property{
						{
							Name:  "analyzers",
							Value: "syft",
						},
					},
				},
				AnalyzerInfo: []string{"syft"},
			},
		},
		{
			name: "append existing analyzer info",
			fields: fields{
				Component: cdx.Component{
					Properties: &[]cdx.Property{
						{
							Name:  "analyzers",
							Value: "syft",
						},
					},
				},
				AnalyzerInfo: []string{"syft"},
			},
			args: args{
				info: "gomod",
			},
			want: &MergedComponent{
				Component: cdx.Component{
					Properties: &[]cdx.Property{
						{
							Name:  "analyzers",
							Value: "syft",
						},
						{
							Name:  "analyzers",
							Value: "gomod",
						},
					},
				},
				AnalyzerInfo: []string{"syft", "gomod"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &MergedComponent{
				Component:    tt.fields.Component,
				AnalyzerInfo: tt.fields.AnalyzerInfo,
			}
			if got := mc.appendAnalyzerInfo(tt.args.info); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("appendAnalyzerInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

// syft cycloneDX output only contains these fields.
var existingComponent = cdx.Component{
	Name:       "test",
	Version:    "1.0.0",
	Type:       cdx.ComponentTypeLibrary,
	PackageURL: "pkg:golang/test.org/test@v1.0.0",
	Properties: &[]cdx.Property{
		{
			Name:  "test",
			Value: "test",
		},
	},
}

// gomod cycloneDX output contains some additional fileds.
var otherComponent = cdx.Component{
	Name:       "test",
	Version:    "1.0.0",
	Type:       cdx.ComponentTypeLibrary,
	BOMRef:     "pkg:golang/test.org/test@v1.0.0?type=module",
	PackageURL: "pkg:golang/test.org/test@v1.0.0?type=module",
	Scope:      "required",
	Hashes: &[]cdx.Hash{
		{
			Algorithm: cdx.HashAlgoSHA256,
			Value:     "1111",
		},
	},
	ExternalReferences: &[]cdx.ExternalReference{
		{
			Type: cdx.ERTypeVCS,
			URL:  "https://test.org/test-reference",
		},
	},
}

var additionalComponent = cdx.Component{
	Name:       "test-2",
	Version:    "1.1.0",
	Type:       cdx.ComponentTypeLibrary,
	PackageURL: "pkg:golang/test.org/test-2@v1.0.0",
	Properties: &[]cdx.Property{
		{
			Name:  "test",
			Value: "test",
		},
	},
}

func createExpectedMergedComponent() *MergedComponent {
	expectedComponent := &MergedComponent{
		Component: otherComponent,
	}
	expectedComponent.Component.Properties = &[]cdx.Property{
		{
			Name:  "test",
			Value: "test",
		},
	}
	expectedComponent.appendAnalyzerInfo("syft")
	expectedComponent.appendAnalyzerInfo("gomod")
	expectedComponent.BomRefs = []string{"pkg:golang/test.org/test@v1.0.0?type=module"}

	return expectedComponent
}

func createAdditionalMergedComponent() *MergedComponent {
	expectedComponent := &MergedComponent{
		Component: additionalComponent,
	}
	expectedComponent.appendAnalyzerInfo("gomod")

	return expectedComponent
}

func Test_handleComponentWithExistingKey(t *testing.T) {
	type args struct {
		mergedComponent *MergedComponent
		otherComponent  cdx.Component
		analyzerInfo    string
	}
	tests := []struct {
		name string
		args args
		want *MergedComponent
	}{
		{
			name: "update missing fileds",
			args: args{
				mergedComponent: newMergedComponent(existingComponent, "syft"),
				otherComponent:  otherComponent,
				analyzerInfo:    "gomod",
			},
			want: createExpectedMergedComponent(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handleComponentWithExistingKey(tt.args.mergedComponent, tt.args.otherComponent, tt.args.analyzerInfo)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("handleComponentWithExistingKey() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMergedResults_Merge(t *testing.T) {
	expectedSbom := &cdx.BOM{
		Components: &[]cdx.Component{otherComponent, additionalComponent},
	}

	type fields struct {
		MergedComponentByKey map[componentKey]*MergedComponent
		Source               utils.SourceType
		SrcMetaData          *cdx.Metadata
		SourceHash           string
	}
	type args struct {
		other *Results
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *MergedResults
	}{
		{
			name: "add and merge components",
			fields: fields{
				MergedComponentByKey: map[componentKey]*MergedComponent{
					createComponentKey(existingComponent): newMergedComponent(existingComponent, "syft"),
				},
			},
			args: args{
				other: &Results{
					Sbom:         expectedSbom,
					AnalyzerInfo: "gomod",
				},
			},
			want: &MergedResults{
				MergedComponentByKey: map[componentKey]*MergedComponent{
					createComponentKey(existingComponent):   createExpectedMergedComponent(),
					createComponentKey(additionalComponent): createAdditionalMergedComponent(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MergedResults{
				MergedComponentByKey: tt.fields.MergedComponentByKey,
				Source:               tt.fields.Source,
				SrcMetaData:          tt.fields.SrcMetaData,
			}
			if got := m.Merge(tt.args.other); !reflect.DeepEqual(got, tt.want) {
				t.Logf("encoded %v\n", expectedSbom)
				t.Errorf("Merge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkMainComponentName(t *testing.T) {
	type args struct {
		mergedName string
		otherName  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "main mergedName and oterName is directory path",
			args: args{
				mergedName: "./test",
				otherName:  "./test",
			},
			want: "./test",
		},
		{
			name: "main mergedName is directory path and oterName is module path",
			args: args{
				mergedName: "./test",
				otherName:  "github.com/test/test",
			},
			want: "github.com/test/test",
		},
		{
			name: "main mergedName is module path and oterName is directory path",
			args: args{
				mergedName: "github.com/test/test",
				otherName:  "./test",
			},
			want: "github.com/test/test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkMainComponentName(tt.args.mergedName, tt.args.otherName); got != tt.want {
				t.Errorf("checkMainComponentName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergedResults_addSourceHash(t *testing.T) {
	type fields struct {
		MergedComponentByKey map[componentKey]*MergedComponent
		Source               utils.SourceType
		SourceHash           string
		SrcMetaData          *cdx.Metadata
	}
	type args struct {
		sourceHash string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *MergedResults
	}{
		{
			name: "sourceHash and hashes are empty",
			fields: fields{
				SrcMetaData: &cdx.Metadata{
					Component: &cdx.Component{},
				},
			},
			args: args{
				sourceHash: "",
			},
			want: &MergedResults{
				SrcMetaData: &cdx.Metadata{
					Component: &cdx.Component{},
				},
			},
		},
		{
			name: "sourceHash empty, hashes has value",
			fields: fields{
				SrcMetaData: &cdx.Metadata{
					Component: &cdx.Component{
						Hashes: &[]cdx.Hash{
							{
								Algorithm: cdx.HashAlgoSHA1,
								Value:     "1111",
							},
						},
					},
				},
			},
			args: args{
				sourceHash: "",
			},
			want: &MergedResults{
				SrcMetaData: &cdx.Metadata{
					Component: &cdx.Component{
						Hashes: &[]cdx.Hash{
							{
								Algorithm: cdx.HashAlgoSHA1,
								Value:     "1111",
							},
						},
					},
				},
			},
		},
		{
			name: "sourceHash not empty, hashes is nil",
			fields: fields{
				SrcMetaData: &cdx.Metadata{
					Component: &cdx.Component{
						Hashes: nil,
					},
				},
			},
			args: args{
				sourceHash: "2222",
			},
			want: &MergedResults{
				SrcMetaData: &cdx.Metadata{
					Component: &cdx.Component{
						Hashes: &[]cdx.Hash{
							{
								Algorithm: cdx.HashAlgoSHA256,
								Value:     "2222",
							},
						},
					},
				},
			},
		},
		{
			name: "sourceHash not empty, hashes is empty",
			fields: fields{
				SrcMetaData: &cdx.Metadata{
					Component: &cdx.Component{
						Hashes: &[]cdx.Hash{},
					},
				},
			},
			args: args{
				sourceHash: "2222",
			},
			want: &MergedResults{
				SrcMetaData: &cdx.Metadata{
					Component: &cdx.Component{
						Hashes: &[]cdx.Hash{
							{
								Algorithm: cdx.HashAlgoSHA256,
								Value:     "2222",
							},
						},
					},
				},
			},
		},
		{
			name: "sourceHash not empty, hashes are conflicting - ignoring the new one",
			fields: fields{
				SrcMetaData: &cdx.Metadata{
					Component: &cdx.Component{
						Hashes: &[]cdx.Hash{
							{
								Algorithm: cdx.HashAlgoSHA256,
								Value:     "1111",
							},
						},
					},
				},
			},
			args: args{
				sourceHash: "2222",
			},
			want: &MergedResults{
				SrcMetaData: &cdx.Metadata{
					Component: &cdx.Component{
						Hashes: &[]cdx.Hash{
							{
								Algorithm: cdx.HashAlgoSHA256,
								Value:     "1111",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := &MergedResults{
				MergedComponentByKey: tt.fields.MergedComponentByKey,
				Source:               tt.fields.Source,
				SourceHash:           tt.fields.SourceHash,
				SrcMetaData:          tt.fields.SrcMetaData,
			}
			if got := mr.addSourceHash(tt.args.sourceHash); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("addSourceHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergedResults_getRealBomRefFromPreviousBomRef(t *testing.T) {
	mergedResults := &MergedResults{
		MergedComponentByKey: map[componentKey]*MergedComponent{
			"1": {
				Component: cdx.Component{
					Name:   "1",
					BOMRef: "bomref1",
				},
				BomRefs: []string{
					"bomref1",
					"oldbomref1",
				},
			},
			"2": {
				Component: cdx.Component{
					Name:   "2",
					BOMRef: "bomref2",
				},
				BomRefs: []string{
					"bomref2",
					"oldbomref2",
				},
			},
		},
		SrcMetaDataBomRefs: []string{
			"mainbomref",
			"oldmainbomref",
		},
		SrcMetaData: &cdx.Metadata{
			Component: &cdx.Component{
				BOMRef: "mainbomref",
			},
		},
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "find bomref from old bomref",
			input: "oldbomref1",
			want:  "bomref1",
		},
		{
			name:  "find bomref from current bomref",
			input: "bomref2",
			want:  "bomref2",
		},
		{
			name:  "find bomref from old main component bomref",
			input: "oldmainbomref",
			want:  "mainbomref",
		},
		{
			name:  "get unknown bomref returns input",
			input: "unknownbomref",
			want:  "unknownbomref",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergedResults.getRealBomRefFromPreviousBomRef(tt.input)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getRealBomRefFromPreviousBomRef() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMergedResults_normalizeDependencies(t *testing.T) {
	mergedResults := &MergedResults{
		MergedComponentByKey: map[componentKey]*MergedComponent{
			"1": {
				Component: cdx.Component{
					Name:   "1",
					BOMRef: "bomref1",
				},
				BomRefs: []string{
					"bomref1",
					"oldbomref1",
				},
			},
			"2": {
				Component: cdx.Component{
					Name:   "2",
					BOMRef: "bomref2",
				},
				BomRefs: []string{
					"bomref2",
					"oldbomref2",
				},
			},
		},
		SrcMetaDataBomRefs: []string{
			"mainbomref",
			"oldmainbomref",
		},
		SrcMetaData: &cdx.Metadata{
			Component: &cdx.Component{
				BOMRef: "mainbomref",
			},
		},
	}

	tests := []struct {
		name         string
		dependencies *[]cdx.Dependency
		want         *[]cdx.Dependency
	}{
		{
			name: "don't change already ok dependencies",
			dependencies: &[]cdx.Dependency{
				{
					Ref: "bomref1",
					Dependencies: &[]string{
						"bomref2",
					},
				},
				{
					Ref: "mainbomref",
					Dependencies: &[]string{
						"bomref1",
					},
				},
			},
			want: &[]cdx.Dependency{
				{
					Ref: "bomref1",
					Dependencies: &[]string{
						"bomref2",
					},
				},
				{
					Ref: "mainbomref",
					Dependencies: &[]string{
						"bomref1",
					},
				},
			},
		},
		{
			name: "all old bom refs",
			dependencies: &[]cdx.Dependency{
				{
					Ref: "oldbomref1",
					Dependencies: &[]string{
						"oldbomref2",
					},
				},
				{
					Ref: "oldmainbomref",
					Dependencies: &[]string{
						"oldbomref1",
					},
				},
			},
			want: &[]cdx.Dependency{
				{
					Ref: "bomref1",
					Dependencies: &[]string{
						"bomref2",
					},
				},
				{
					Ref: "mainbomref",
					Dependencies: &[]string{
						"bomref1",
					},
				},
			},
		},
		{
			name: "mix of old and new bom refs",
			dependencies: &[]cdx.Dependency{
				{
					Ref: "bomref1",
					Dependencies: &[]string{
						"oldbomref2",
					},
				},
				{
					Ref: "oldmainbomref",
					Dependencies: &[]string{
						"bomref1",
					},
				},
			},
			want: &[]cdx.Dependency{
				{
					Ref: "bomref1",
					Dependencies: &[]string{
						"bomref2",
					},
				},
				{
					Ref: "mainbomref",
					Dependencies: &[]string{
						"bomref1",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergedResults.normalizeDependencies(tt.dependencies)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getRealBomRefFromPreviousBomRef() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMergedResults_mergeDependencies(t *testing.T) {
	tests := []struct {
		name          string
		dependenciesA *[]cdx.Dependency
		dependenciesB *[]cdx.Dependency
		want          *[]cdx.Dependency
	}{
		{
			name: "merge",
			dependenciesA: &[]cdx.Dependency{
				{
					Ref: "bomref1",
					Dependencies: &[]string{
						"bomref2",
						"bomref4",
					},
				},
				{
					Ref: "mainbomref",
					Dependencies: &[]string{
						"bomref1",
					},
				},
			},
			dependenciesB: &[]cdx.Dependency{
				{
					Ref: "bomref1",
					Dependencies: &[]string{
						"bomref2",
						"bomref3",
					},
				},
			},
			want: &[]cdx.Dependency{
				{
					Ref: "bomref1",
					Dependencies: &[]string{
						"bomref2",
						"bomref3",
						"bomref4",
					},
				},
				{
					Ref: "mainbomref",
					Dependencies: &[]string{
						"bomref1",
					},
				},
			},
		},
		{
			name: "depA is nil",
			dependenciesA: &[]cdx.Dependency{
				{
					Ref: "bomref1",
					Dependencies: &[]string{
						"bomref2",
						"bomref4",
					},
				},
				{
					Ref: "mainbomref",
					Dependencies: &[]string{
						"bomref1",
					},
				},
			},
			dependenciesB: nil,
			want: &[]cdx.Dependency{
				{
					Ref: "bomref1",
					Dependencies: &[]string{
						"bomref2",
						"bomref4",
					},
				},
				{
					Ref: "mainbomref",
					Dependencies: &[]string{
						"bomref1",
					},
				},
			},
		},
		{
			name:          "depB is nil",
			dependenciesA: nil,
			dependenciesB: &[]cdx.Dependency{
				{
					Ref: "bomref1",
					Dependencies: &[]string{
						"bomref3",
					},
				},
				{
					Ref: "mainbomref",
					Dependencies: &[]string{
						"bomref1",
					},
				},
			},
			want: &[]cdx.Dependency{
				{
					Ref: "bomref1",
					Dependencies: &[]string{
						"bomref3",
					},
				},
				{
					Ref: "mainbomref",
					Dependencies: &[]string{
						"bomref1",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeDependencies(tt.dependenciesA, tt.dependenciesB)
			if diff := cmp.Diff(tt.want, got, cmpopts.SortSlices(func(a, b cdx.Dependency) bool { return a.Ref < b.Ref })); diff != "" {
				t.Errorf("mergeDependencies() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
