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

package rest

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gotest.tools/v3/assert"

	backendmodels "github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/findingkey"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
	"github.com/openclarity/vmclarity/pkg/uibackend/api/models"
)

func Test_processFindings(t *testing.T) {
	rootkitFindingInfo1 := createRootkitFindingInfo(t, "path1", "name1", "type1")
	rfKey1, err := findingkey.GenerateFindingKey(rootkitFindingInfo1)
	assert.NilError(t, err)
	rootkitFindingInfo2 := createRootkitFindingInfo(t, "path2", "name2", "type2")
	rfKey2, err := findingkey.GenerateFindingKey(rootkitFindingInfo2)
	assert.NilError(t, err)
	rootkitFindingInfo3 := createRootkitFindingInfo(t, "path3", "name3", "type3")
	rfKey3, err := findingkey.GenerateFindingKey(rootkitFindingInfo3)
	assert.NilError(t, err)
	type args struct {
		findings            []backendmodels.Finding
		findingAssetMap     map[findingAssetKey]struct{}
		findingToAssetCount map[string]findingInfoCount
	}
	tests := []struct {
		name                        string
		args                        args
		wantErr                     bool
		expectedFindingAssetMap     map[findingAssetKey]struct{}
		expectedFindingToAssetCount map[string]findingInfoCount
	}{
		{
			name: "findingAssetKey not in findingAssetMap - first time we see the finding",
			args: args{
				findings: []backendmodels.Finding{
					{
						Asset: &backendmodels.AssetRelationship{
							Id: "asset-1",
						},
						FoundBy: &backendmodels.AssetScanRelationship{
							Id: "assetscan-1",
							Asset: &backendmodels.AssetRelationship{
								Id: "asset-1",
							},
						},
						FindingInfo: rootkitFindingInfo1,
					},
				},
				findingAssetMap:     make(map[findingAssetKey]struct{}),
				findingToAssetCount: make(map[string]findingInfoCount),
			},
			wantErr: false,
			expectedFindingAssetMap: map[findingAssetKey]struct{}{
				{FindingKey: rfKey1, AssetID: "asset-1"}: {}, // nolint:gofmt,gofumpt
			},
			expectedFindingToAssetCount: map[string]findingInfoCount{
				rfKey1: {
					FindingInfo: rootkitFindingInfo1,
					AssetCount:  1,
				},
			},
		},
		{
			name: "findingAssetKey already in findingAssetMap - expected no change",
			args: args{
				findings: []backendmodels.Finding{
					{
						Asset: &backendmodels.AssetRelationship{
							Id: "asset-1",
						},
						FoundBy: &backendmodels.AssetScanRelationship{
							Id: "assetscan-1",
							Asset: &backendmodels.AssetRelationship{
								Id: "asset-1",
							},
						},
						FindingInfo: rootkitFindingInfo1,
					},
				},
				findingAssetMap: map[findingAssetKey]struct{}{
					{FindingKey: rfKey1, AssetID: "asset-1"}: {}, // nolint:gofmt,gofumpt
				},
				findingToAssetCount: map[string]findingInfoCount{
					rfKey1: {
						FindingInfo: rootkitFindingInfo1,
						AssetCount:  1,
					},
				},
			},
			wantErr: false,
			expectedFindingAssetMap: map[findingAssetKey]struct{}{
				{FindingKey: rfKey1, AssetID: "asset-1"}: {}, // nolint:gofmt,gofumpt
			},
			expectedFindingToAssetCount: map[string]findingInfoCount{
				rfKey1: {
					FindingInfo: rootkitFindingInfo1,
					AssetCount:  1,
				},
			},
		},
		{
			name: "findingAssetKey not in findingAssetMap - already saw the finding",
			args: args{
				findings: []backendmodels.Finding{
					{
						Asset: &backendmodels.AssetRelationship{
							Id: "asset-1",
						},
						FoundBy: &backendmodels.AssetScanRelationship{
							Id: "assetscan-1",
							Asset: &backendmodels.AssetRelationship{
								Id: "asset-1",
							},
						},
						FindingInfo: rootkitFindingInfo1,
					},
				},
				findingAssetMap: map[findingAssetKey]struct{}{
					{FindingKey: rfKey1, AssetID: "asset-2"}: {}, // nolint:gofmt,gofumpt
				},
				findingToAssetCount: map[string]findingInfoCount{
					rfKey1: {
						FindingInfo: rootkitFindingInfo1,
						AssetCount:  1,
					},
				},
			},
			wantErr: false,
			expectedFindingAssetMap: map[findingAssetKey]struct{}{
				{FindingKey: rfKey1, AssetID: "asset-2"}: {}, // nolint:gofmt,gofumpt
				{FindingKey: rfKey1, AssetID: "asset-1"}: {}, // nolint:gofmt,gofumpt
			},
			expectedFindingToAssetCount: map[string]findingInfoCount{
				rfKey1: {
					FindingInfo: rootkitFindingInfo1,
					AssetCount:  2,
				},
			},
		},
		{
			name: "multiple findings:" +
				"findingAssetKey already in findingAssetMap - expected no change," +
				"findingAssetKey not in findingAssetMap - already saw the finding," +
				"findingAssetKey not in findingAssetMap - first time we see the finding",
			args: args{
				findings: []backendmodels.Finding{
					{
						// findingAssetKey already in findingAssetMap
						Asset: &backendmodels.AssetRelationship{
							Id: "asset-1",
						},
						FoundBy: &backendmodels.AssetScanRelationship{
							Id: "assetscan-1",
							Asset: &backendmodels.AssetRelationship{
								Id: "asset-1",
							},
						},
						FindingInfo: rootkitFindingInfo1,
					},
					{
						// findingAssetKey not in findingAssetMap - already saw the finding
						Asset: &backendmodels.AssetRelationship{
							Id: "asset-1",
						},
						FoundBy: &backendmodels.AssetScanRelationship{
							Id: "assetscan-1",
							Asset: &backendmodels.AssetRelationship{
								Id: "asset-1",
							},
						},
						FindingInfo: rootkitFindingInfo2,
					},
					{
						// findingAssetKey not in findingAssetMap - first time we see the finding
						Asset: &backendmodels.AssetRelationship{
							Id: "asset-1",
						},
						FoundBy: &backendmodels.AssetScanRelationship{
							Id: "assetscan-1",
							Asset: &backendmodels.AssetRelationship{
								Id: "asset-1",
							},
						},
						FindingInfo: rootkitFindingInfo3,
					},
				},
				findingAssetMap: map[findingAssetKey]struct{}{
					{FindingKey: rfKey1, AssetID: "asset-1"}: {}, // nolint:gofmt,gofumpt
					{FindingKey: rfKey2, AssetID: "asset-2"}: {}, // nolint:gofmt,gofumpt
				},
				findingToAssetCount: map[string]findingInfoCount{
					rfKey1: {
						FindingInfo: rootkitFindingInfo1,
						AssetCount:  1,
					},
					rfKey2: {
						FindingInfo: rootkitFindingInfo2,
						AssetCount:  1,
					},
				},
			},
			wantErr: false,
			expectedFindingAssetMap: map[findingAssetKey]struct{}{
				{FindingKey: rfKey1, AssetID: "asset-1"}: {}, // nolint:gofmt,gofumpt
				{FindingKey: rfKey2, AssetID: "asset-1"}: {}, // nolint:gofmt,gofumpt
				{FindingKey: rfKey2, AssetID: "asset-2"}: {}, // nolint:gofmt,gofumpt
				{FindingKey: rfKey3, AssetID: "asset-1"}: {}, // nolint:gofmt,gofumpt
			},
			expectedFindingToAssetCount: map[string]findingInfoCount{
				rfKey1: {
					FindingInfo: rootkitFindingInfo1,
					AssetCount:  1,
				},
				rfKey2: {
					FindingInfo: rootkitFindingInfo2,
					AssetCount:  2,
				},
				rfKey3: {
					FindingInfo: rootkitFindingInfo3,
					AssetCount:  1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = processFindings(tt.args.findings, tt.args.findingAssetMap, tt.args.findingToAssetCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("processFindings() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				if diff := cmp.Diff(tt.expectedFindingAssetMap, tt.args.findingAssetMap, cmpopts.SortSlices(func(a, b findingAssetKey) bool { return a.FindingKey < b.FindingKey })); diff != "" {
					t.Errorf("compareFindingInfo mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expectedFindingToAssetCount, tt.args.findingToAssetCount, cmp.Comparer(compareFindingInfo)); diff != "" {
					t.Errorf("processFindings findingToAssetCount mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

// nolint:cyclop
func compareFindingInfo(a, b backendmodels.Finding_FindingInfo) bool {
	value, err := a.ValueByDiscriminator()
	if err != nil {
		return false
	}

	switch findingInfoA := value.(type) {
	case backendmodels.ExploitFindingInfo:
		findingInfoB, err := b.AsExploitFindingInfo()
		if err != nil {
			return false
		}
		if diff := cmp.Diff(findingInfoA, findingInfoB); diff != "" {
			fmt.Printf("compareFindingInfo mismatch (-want +got):\n%s", diff)
			return false
		}
		return true
	case backendmodels.VulnerabilityFindingInfo:
		findingInfoB, err := b.AsVulnerabilityFindingInfo()
		if err != nil {
			return false
		}
		if diff := cmp.Diff(findingInfoA, findingInfoB); diff != "" {
			fmt.Printf("compareFindingInfo mismatch (-want +got):\n%s", diff)
			return false
		}
		return true
	case backendmodels.MalwareFindingInfo:
		findingInfoB, err := b.AsMalwareFindingInfo()
		if err != nil {
			return false
		}
		if diff := cmp.Diff(findingInfoA, findingInfoB); diff != "" {
			fmt.Printf("compareFindingInfo mismatch (-want +got):\n%s", diff)
			return false
		}
		return true
	case backendmodels.MisconfigurationFindingInfo:
		findingInfoB, err := b.AsMisconfigurationFindingInfo()
		if err != nil {
			return false
		}
		if diff := cmp.Diff(findingInfoA, findingInfoB); diff != "" {
			fmt.Printf("compareFindingInfo mismatch (-want +got):\n%s", diff)
			return false
		}
		return true
	case backendmodels.RootkitFindingInfo:
		findingInfoB, err := b.AsRootkitFindingInfo()
		if err != nil {
			return false
		}
		if diff := cmp.Diff(findingInfoA, findingInfoB); diff != "" {
			fmt.Printf("compareFindingInfo mismatch (-want +got):\n%s", diff)
			return false
		}
		return true
	case backendmodels.SecretFindingInfo:
		findingInfoB, err := b.AsSecretFindingInfo()
		if err != nil {
			return false
		}
		if diff := cmp.Diff(findingInfoA, findingInfoB); diff != "" {
			fmt.Printf("compareFindingInfo mismatch (-want +got):\n%s", diff)
			return false
		}
		return true
	case backendmodels.PackageFindingInfo:
		findingInfoB, err := b.AsPackageFindingInfo()
		if err != nil {
			return false
		}
		if diff := cmp.Diff(findingInfoA, findingInfoB); diff != "" {
			fmt.Printf("compareFindingInfo mismatch (-want +got):\n%s", diff)
			return false
		}
		return true
	default:
		fmt.Printf("unsupported finding findingInfoA type %T", value)
		return false
	}
}

func createRootkitFindingInfo(t *testing.T, message, name, tpe string) *backendmodels.Finding_FindingInfo {
	t.Helper()
	findingInfoB := backendmodels.Finding_FindingInfo{}
	err := findingInfoB.FromRootkitFindingInfo(backendmodels.RootkitFindingInfo{
		Message:     utils.PointerTo(message),
		RootkitName: utils.PointerTo(name),
		RootkitType: utils.PointerTo(backendmodels.RootkitType(tpe)),
	})
	assert.NilError(t, err)
	return &findingInfoB
}

func Test_getSortedFindingInfoCountSlice(t *testing.T) {
	rootkitFindingInfo1 := createRootkitFindingInfo(t, "path1", "name1", "type1")
	rfKey1, err := findingkey.GenerateFindingKey(rootkitFindingInfo1)
	assert.NilError(t, err)
	rootkitFindingInfo2 := createRootkitFindingInfo(t, "path2", "name2", "type2")
	rfKey2, err := findingkey.GenerateFindingKey(rootkitFindingInfo2)
	assert.NilError(t, err)
	rootkitFindingInfo3 := createRootkitFindingInfo(t, "path3", "name3", "type3")
	rfKey3, err := findingkey.GenerateFindingKey(rootkitFindingInfo3)
	assert.NilError(t, err)
	rootkitFindingInfo4 := createRootkitFindingInfo(t, "path4", "name4", "type4")
	rfKey4, err := findingkey.GenerateFindingKey(rootkitFindingInfo4)
	assert.NilError(t, err)
	type args struct {
		findingAssetMapCount map[string]findingInfoCount
	}
	tests := []struct {
		name string
		args args
		want []findingInfoCount
	}{
		{
			name: "nil map",
			args: args{
				findingAssetMapCount: nil,
			},
			want: []findingInfoCount{},
		},
		{
			name: "sanity",
			args: args{
				findingAssetMapCount: map[string]findingInfoCount{
					rfKey1: {
						FindingInfo: rootkitFindingInfo1,
						AssetCount:  1,
					},
					rfKey2: {
						FindingInfo: rootkitFindingInfo2,
						AssetCount:  5,
					},
					rfKey3: {
						FindingInfo: rootkitFindingInfo3,
						AssetCount:  8,
					},
					rfKey4: {
						FindingInfo: rootkitFindingInfo4,
						AssetCount:  3,
					},
				},
			},
			want: []findingInfoCount{
				{
					FindingInfo: rootkitFindingInfo3,
					AssetCount:  8,
				},
				{
					FindingInfo: rootkitFindingInfo2,
					AssetCount:  5,
				},
				{
					FindingInfo: rootkitFindingInfo4,
					AssetCount:  3,
				},
				{
					FindingInfo: rootkitFindingInfo1,
					AssetCount:  1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSortedFindingInfoCountSlice(tt.args.findingAssetMapCount)
			if diff := cmp.Diff(tt.want, got, cmp.Comparer(compareFindingInfo)); diff != "" {
				t.Errorf("getSortedFindingInfoCountSlice mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_createFindingsImpact(t *testing.T) {
	type args struct {
		findingInfoCountSlice []findingInfoCount
		createFunc            func(findingInfo *backendmodels.Finding_FindingInfo, count int) (models.RootkitFindingImpact, error)
	}
	tests := []struct {
		name    string
		args    args
		want    []models.RootkitFindingImpact
		wantErr bool
	}{
		{
			name: "findingInfoCountSlice len is lt maxFindingsImpactCount",
			args: args{
				findingInfoCountSlice: []findingInfoCount{
					{
						FindingInfo: createRootkitFindingInfo(t, "path1", "name1", "type1"),
						AssetCount:  19,
					},
					{
						FindingInfo: createRootkitFindingInfo(t, "path2", "name2", "type2"),
						AssetCount:  7,
					},
					{
						FindingInfo: createRootkitFindingInfo(t, "path3", "name3", "type3"),
						AssetCount:  5,
					},
				},
				createFunc: createRootkitFindingImpact,
			},
			want: []models.RootkitFindingImpact{
				{
					AffectedAssetsCount: utils.PointerTo(19),
					Rootkit: &models.Rootkit{
						Message:     utils.PointerTo("path1"),
						RootkitName: utils.PointerTo("name1"),
						RootkitType: utils.PointerTo(models.RootkitType("type1")),
					},
				},
				{
					AffectedAssetsCount: utils.PointerTo(7),
					Rootkit: &models.Rootkit{
						Message:     utils.PointerTo("path2"),
						RootkitName: utils.PointerTo("name2"),
						RootkitType: utils.PointerTo(models.RootkitType("type2")),
					},
				},
				{
					AffectedAssetsCount: utils.PointerTo(5),
					Rootkit: &models.Rootkit{
						Message:     utils.PointerTo("path3"),
						RootkitName: utils.PointerTo("name3"),
						RootkitType: utils.PointerTo(models.RootkitType("type3")),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "findingInfoCountSlice len is gt maxFindingsImpactCount",
			args: args{
				findingInfoCountSlice: []findingInfoCount{
					{
						FindingInfo: createRootkitFindingInfo(t, "path1", "name1", "type1"),
						AssetCount:  19,
					},
					{
						FindingInfo: createRootkitFindingInfo(t, "path2", "name2", "type2"),
						AssetCount:  7,
					},
					{
						FindingInfo: createRootkitFindingInfo(t, "path3", "name3", "type3"),
						AssetCount:  5,
					},
					{
						FindingInfo: createRootkitFindingInfo(t, "path4", "name3", "type3"),
						AssetCount:  4,
					},
					{
						FindingInfo: createRootkitFindingInfo(t, "path5", "name3", "type3"),
						AssetCount:  3,
					},
					{
						FindingInfo: createRootkitFindingInfo(t, "path6", "name3", "type3"),
						AssetCount:  2,
					},
					{
						FindingInfo: createRootkitFindingInfo(t, "path7", "name3", "type3"),
						AssetCount:  1,
					},
				},
				createFunc: createRootkitFindingImpact,
			},
			want: []models.RootkitFindingImpact{
				{
					AffectedAssetsCount: utils.PointerTo(19),
					Rootkit: &models.Rootkit{
						Message:     utils.PointerTo("path1"),
						RootkitName: utils.PointerTo("name1"),
						RootkitType: utils.PointerTo(models.RootkitType("type1")),
					},
				},
				{
					AffectedAssetsCount: utils.PointerTo(7),
					Rootkit: &models.Rootkit{
						Message:     utils.PointerTo("path2"),
						RootkitName: utils.PointerTo("name2"),
						RootkitType: utils.PointerTo(models.RootkitType("type2")),
					},
				},
				{
					AffectedAssetsCount: utils.PointerTo(5),
					Rootkit: &models.Rootkit{
						Message:     utils.PointerTo("path3"),
						RootkitName: utils.PointerTo("name3"),
						RootkitType: utils.PointerTo(models.RootkitType("type3")),
					},
				},
				{
					AffectedAssetsCount: utils.PointerTo(4),
					Rootkit: &models.Rootkit{
						Message:     utils.PointerTo("path4"),
						RootkitName: utils.PointerTo("name3"),
						RootkitType: utils.PointerTo(models.RootkitType("type3")),
					},
				},
				{
					AffectedAssetsCount: utils.PointerTo(3),
					Rootkit: &models.Rootkit{
						Message:     utils.PointerTo("path5"),
						RootkitName: utils.PointerTo("name3"),
						RootkitType: utils.PointerTo(models.RootkitType("type3")),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createFindingsImpact(tt.args.findingInfoCountSlice, tt.args.createFunc)
			if (err != nil) != tt.wantErr {
				t.Errorf("createFindingsImpact() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("createFindingsImpact() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
