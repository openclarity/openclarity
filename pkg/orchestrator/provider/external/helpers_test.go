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

package external

import (
	"reflect"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gotest.tools/v3/assert"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/orchestrator/provider"
	provider_service "github.com/openclarity/vmclarity/pkg/orchestrator/provider/external/proto"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

func Test_convertAssetToModels(t *testing.T) {
	timeNow := time.Now()

	wantVMInfo := models.AssetType{}
	err := wantVMInfo.FromVMInfo(models.VMInfo{
		Image:            "image1",
		InstanceID:       "id1",
		InstanceProvider: utils.PointerTo(models.External),
		InstanceType:     "type1",
		LaunchTime:       timestamppb.New(timeNow).AsTime(),
		Location:         "location1",
		Platform:         "linux",
		SecurityGroups:   &[]models.SecurityGroup{},
		Tags: &[]models.Tag{
			{
				Key:   "key1",
				Value: "val1",
			},
		},
	})
	assert.NilError(t, err)

	wantDirInfo := models.AssetType{}
	err = wantDirInfo.FromDirInfo(models.DirInfo{
		DirName:  utils.PointerTo("dir1"),
		Location: utils.PointerTo("dirLocation1"),
	})
	assert.NilError(t, err)

	wantPodInfo := models.AssetType{}
	err = wantPodInfo.FromPodInfo(models.PodInfo{
		PodName:  utils.PointerTo("pod1"),
		Location: utils.PointerTo("podLocation1"),
	})
	assert.NilError(t, err)

	type args struct {
		asset *provider_service.Asset
	}
	tests := []struct {
		name    string
		args    args
		want    models.Asset
		wantErr bool
	}{
		{
			name: "nil asset",
			args: args{
				asset: nil,
			},
			want:    models.Asset{},
			wantErr: true,
		},
		{
			name: "unsupported asset type",
			args: args{
				asset: &provider_service.Asset{},
			},
			want:    models.Asset{},
			wantErr: true,
		},
		{
			name: "vm info",
			args: args{
				asset: &provider_service.Asset{
					AssetType: &provider_service.Asset_Vminfo{Vminfo: &provider_service.VMInfo{
						Id:           "id1",
						Location:     "location1",
						Image:        "image1",
						InstanceType: "type1",
						Platform:     "linux",
						Tags: []*provider_service.Tag{
							{
								Key: "key1",
								Val: "val1",
							},
						},
						LaunchTime: timestamppb.New(timeNow),
					}},
				},
			},
			want: models.Asset{
				AssetInfo: &wantVMInfo,
			},
			wantErr: false,
		},
		{
			name: "dir info",
			args: args{
				asset: &provider_service.Asset{
					AssetType: &provider_service.Asset_Dirinfo{Dirinfo: &provider_service.DirInfo{
						DirName:  "dir1",
						Location: "dirLocation1",
					}},
				},
			},
			want: models.Asset{
				AssetInfo: &wantDirInfo,
			},
			wantErr: false,
		},
		{
			name: "pod info",
			args: args{
				asset: &provider_service.Asset{
					AssetType: &provider_service.Asset_Podinfo{Podinfo: &provider_service.PodInfo{
						PodName:  "pod1",
						Location: "podLocation1",
					}},
				},
			},
			want: models.Asset{
				AssetInfo: &wantPodInfo,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertAssetToModels(tt.args.asset)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertAssetToModels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertAssetToModels() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertAssetFromModels(t *testing.T) {
	timeNow := time.Now()

	vminfo := models.AssetType{}
	err := vminfo.FromVMInfo(models.VMInfo{
		Image:            "image1",
		InstanceID:       "id1",
		InstanceProvider: utils.PointerTo(models.External),
		InstanceType:     "type1",
		LaunchTime:       timestamppb.New(timeNow).AsTime(),
		Location:         "location1",
		Platform:         "linux",
		SecurityGroups:   &[]models.SecurityGroup{},
		Tags: &[]models.Tag{
			{
				Key:   "key1",
				Value: "val1",
			},
		},
	})
	assert.NilError(t, err)

	dirinfo := models.AssetType{}
	err = dirinfo.FromDirInfo(models.DirInfo{
		DirName:  utils.PointerTo("dir1"),
		Location: utils.PointerTo("dirLocation1"),
	})
	assert.NilError(t, err)

	podinfo := models.AssetType{}
	err = podinfo.FromPodInfo(models.PodInfo{
		PodName:  utils.PointerTo("pod1"),
		Location: utils.PointerTo("podLocation1"),
	})
	assert.NilError(t, err)

	type args struct {
		asset models.Asset
	}
	tests := []struct {
		name    string
		args    args
		want    *provider_service.Asset
		wantErr bool
	}{
		{
			name: "vm info",
			args: args{
				asset: models.Asset{
					AssetInfo: &vminfo,
				},
			},
			want: &provider_service.Asset{
				AssetType: &provider_service.Asset_Vminfo{Vminfo: &provider_service.VMInfo{
					Id:           "id1",
					Location:     "location1",
					Image:        "image1",
					InstanceType: "type1",
					Platform:     "linux",
					Tags: []*provider_service.Tag{
						{
							Key: "key1",
							Val: "val1",
						},
					},
					LaunchTime: timestamppb.New(timeNow),
				}},
			},
			wantErr: false,
		},
		{
			name: "dir info",
			args: args{
				asset: models.Asset{
					AssetInfo: &dirinfo,
				},
			},
			want: &provider_service.Asset{
				AssetType: &provider_service.Asset_Dirinfo{Dirinfo: &provider_service.DirInfo{
					DirName:  "dir1",
					Location: "dirLocation1",
				}},
			},
			wantErr: false,
		},
		{
			name: "pod info",
			args: args{
				asset: models.Asset{
					AssetInfo: &podinfo,
				},
			},
			want: &provider_service.Asset{
				AssetType: &provider_service.Asset_Podinfo{Podinfo: &provider_service.PodInfo{
					PodName:  "pod1",
					Location: "podLocation1",
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertAssetFromModels(tt.args.asset)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertAssetFromModels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertAssetFromModels() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertTagsToModels(t *testing.T) {
	type args struct {
		tags []*provider_service.Tag
	}
	tests := []struct {
		name string
		args args
		want *[]models.Tag
	}{
		{
			name: "no tags",
			args: args{
				tags: []*provider_service.Tag{},
			},
			want: nil,
		},
		{
			name: "one tag",
			args: args{
				tags: []*provider_service.Tag{
					{
						Key: "key1",
						Val: "val1",
					},
				},
			},
			want: &[]models.Tag{
				{
					Key:   "key1",
					Value: "val1",
				},
			},
		},
		{
			name: "two tag",
			args: args{
				tags: []*provider_service.Tag{
					{
						Key: "key1",
						Val: "val1",
					},
					{
						Key: "key2",
						Val: "val2",
					},
				},
			},
			want: &[]models.Tag{
				{
					Key:   "key1",
					Value: "val1",
				},
				{
					Key:   "key2",
					Value: "val2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertTagsToModels(tt.args.tags); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertTagsToModels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertScanJobConfig(t *testing.T) {
	podinfo := models.AssetType{}
	err := podinfo.FromPodInfo(models.PodInfo{
		PodName:  utils.PointerTo("pod1"),
		Location: utils.PointerTo("podLocation1"),
	})
	assert.NilError(t, err)

	type args struct {
		config *provider.ScanJobConfig
	}
	tests := []struct {
		name    string
		args    args
		want    *provider_service.ScanJobConfig
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				config: &provider.ScanJobConfig{
					ScannerImage:     "image1",
					ScannerCLIConfig: "cliconfig",
					VMClarityAddress: "addr",
					ScanMetadata: provider.ScanMetadata{
						ScanID:      "scanid1",
						AssetScanID: "assetscanid1",
						AssetID:     "assetid1",
					},
					ScannerInstanceCreationConfig: models.ScannerInstanceCreationConfig{
						MaxPrice:         nil,
						RetryMaxAttempts: nil,
						UseSpotInstances: false,
					},
					Asset: models.Asset{
						AssetInfo: &podinfo,
					},
				},
			},
			want: &provider_service.ScanJobConfig{
				ScannerImage:     "image1",
				ScannerCLIConfig: "cliconfig",
				VmClarityAddress: "addr",
				ScanMetadata: &provider_service.ScanMetadata{
					ScanID:      "scanid1",
					AssetScanID: "assetscanid1",
					AssetID:     "assetid1",
				},
				ScannerInstanceCreationConfig: &provider_service.ScannerInstanceCreationConfig{
					MaxPrice:         "",
					RetryMaxAttempts: 0,
					UseSpotInstances: false,
				},
				Asset: &provider_service.Asset{AssetType: &provider_service.Asset_Podinfo{Podinfo: &provider_service.PodInfo{
					PodName:  "pod1",
					Location: "podLocation1",
				}}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertScanJobConfig(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertScanJobConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertScanJobConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
