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
	"gotest.tools/assert"

	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/formatter"
	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/utils"
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
			if got := handleComponentWithExistingKey(tt.args.mergedComponent, tt.args.otherComponent, tt.args.analyzerInfo); !reflect.DeepEqual(got, tt.want) {
				t.Logf("properties got %v, properties want %v", got.Component.Properties, tt.want.Component.Properties)
				t.Errorf("handleComponentWithExistingKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergedResults_Merge(t *testing.T) {
	otherFormatter := formatter.New(formatter.CycloneDXFormat, []byte{})
	err := otherFormatter.SetSBOM(&cdx.BOM{
		Components: &[]cdx.Component{otherComponent, additionalComponent},
	})
	assert.NilError(t, err)
	_ = otherFormatter.Encode(formatter.CycloneDXFormat)

	type fields struct {
		MergedComponentByKey map[componentKey]*MergedComponent
		Source               utils.SourceType
		SrcMetaData          *cdx.Metadata
		SourceHash           string
	}
	type args struct {
		other  *Results
		format string
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
					Sbom:         otherFormatter.GetSBOMBytes(),
					AnalyzerInfo: "gomod",
				},
				format: formatter.CycloneDXFormat,
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
			if got := m.Merge(tt.args.other, tt.args.format); !reflect.DeepEqual(got, tt.want) {
				t.Logf("encoded %v\n", otherFormatter.GetSBOM())
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
			name: "sorceHash and hashes are empty",
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
			name: "sourceHash not empty, hashes is empty",
			fields: fields{
				SrcMetaData: &cdx.Metadata{
					Component: &cdx.Component{},
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
			name: "sourceHash not empty, hashes has value",
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
