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

package server

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gotest.tools/v3/assert"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/scanner/findingkey"
	"github.com/openclarity/vmclarity/uibackend/types"
)

// nolint:cyclop
func compareFindingInfo(a, b apitypes.FindingInfo) bool {
	value, err := a.ValueByDiscriminator()
	if err != nil {
		return false
	}

	switch findingInfoA := value.(type) {
	case apitypes.ExploitFindingInfo:
		findingInfoB, err := b.AsExploitFindingInfo()
		if err != nil {
			return false
		}
		if diff := cmp.Diff(findingInfoA, findingInfoB); diff != "" {
			fmt.Printf("compareFindingInfo mismatch (-want +got):\n%s", diff)
			return false
		}
		return true
	case apitypes.VulnerabilityFindingInfo:
		findingInfoB, err := b.AsVulnerabilityFindingInfo()
		if err != nil {
			return false
		}
		if diff := cmp.Diff(findingInfoA, findingInfoB); diff != "" {
			fmt.Printf("compareFindingInfo mismatch (-want +got):\n%s", diff)
			return false
		}
		return true
	case apitypes.MalwareFindingInfo:
		findingInfoB, err := b.AsMalwareFindingInfo()
		if err != nil {
			return false
		}
		if diff := cmp.Diff(findingInfoA, findingInfoB); diff != "" {
			fmt.Printf("compareFindingInfo mismatch (-want +got):\n%s", diff)
			return false
		}
		return true
	case apitypes.MisconfigurationFindingInfo:
		findingInfoB, err := b.AsMisconfigurationFindingInfo()
		if err != nil {
			return false
		}
		if diff := cmp.Diff(findingInfoA, findingInfoB); diff != "" {
			fmt.Printf("compareFindingInfo mismatch (-want +got):\n%s", diff)
			return false
		}
		return true
	case apitypes.RootkitFindingInfo:
		findingInfoB, err := b.AsRootkitFindingInfo()
		if err != nil {
			return false
		}
		if diff := cmp.Diff(findingInfoA, findingInfoB); diff != "" {
			fmt.Printf("compareFindingInfo mismatch (-want +got):\n%s", diff)
			return false
		}
		return true
	case apitypes.SecretFindingInfo:
		findingInfoB, err := b.AsSecretFindingInfo()
		if err != nil {
			return false
		}
		if diff := cmp.Diff(findingInfoA, findingInfoB); diff != "" {
			fmt.Printf("compareFindingInfo mismatch (-want +got):\n%s", diff)
			return false
		}
		return true
	case apitypes.PackageFindingInfo:
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

func createRootkitFindingInfo(t *testing.T, message, name, tpe string) *apitypes.FindingInfo {
	t.Helper()
	findingInfoB := apitypes.FindingInfo{}
	err := findingInfoB.FromRootkitFindingInfo(apitypes.RootkitFindingInfo{
		Message:     to.Ptr(message),
		RootkitName: to.Ptr(name),
		RootkitType: to.Ptr(apitypes.RootkitType(tpe)),
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
		createFunc            func(findingInfo *apitypes.FindingInfo, count int) (types.RootkitFindingImpact, error)
	}
	tests := []struct {
		name    string
		args    args
		want    []types.RootkitFindingImpact
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
			want: []types.RootkitFindingImpact{
				{
					AffectedAssetsCount: to.Ptr(19),
					Rootkit: &types.Rootkit{
						Message:     to.Ptr("path1"),
						RootkitName: to.Ptr("name1"),
						RootkitType: to.Ptr(types.RootkitType("type1")),
					},
				},
				{
					AffectedAssetsCount: to.Ptr(7),
					Rootkit: &types.Rootkit{
						Message:     to.Ptr("path2"),
						RootkitName: to.Ptr("name2"),
						RootkitType: to.Ptr(types.RootkitType("type2")),
					},
				},
				{
					AffectedAssetsCount: to.Ptr(5),
					Rootkit: &types.Rootkit{
						Message:     to.Ptr("path3"),
						RootkitName: to.Ptr("name3"),
						RootkitType: to.Ptr(types.RootkitType("type3")),
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
			want: []types.RootkitFindingImpact{
				{
					AffectedAssetsCount: to.Ptr(19),
					Rootkit: &types.Rootkit{
						Message:     to.Ptr("path1"),
						RootkitName: to.Ptr("name1"),
						RootkitType: to.Ptr(types.RootkitType("type1")),
					},
				},
				{
					AffectedAssetsCount: to.Ptr(7),
					Rootkit: &types.Rootkit{
						Message:     to.Ptr("path2"),
						RootkitName: to.Ptr("name2"),
						RootkitType: to.Ptr(types.RootkitType("type2")),
					},
				},
				{
					AffectedAssetsCount: to.Ptr(5),
					Rootkit: &types.Rootkit{
						Message:     to.Ptr("path3"),
						RootkitName: to.Ptr("name3"),
						RootkitType: to.Ptr(types.RootkitType("type3")),
					},
				},
				{
					AffectedAssetsCount: to.Ptr(4),
					Rootkit: &types.Rootkit{
						Message:     to.Ptr("path4"),
						RootkitName: to.Ptr("name3"),
						RootkitType: to.Ptr(types.RootkitType("type3")),
					},
				},
				{
					AffectedAssetsCount: to.Ptr(3),
					Rootkit: &types.Rootkit{
						Message:     to.Ptr("path5"),
						RootkitName: to.Ptr("name3"),
						RootkitType: to.Ptr(types.RootkitType("type3")),
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
