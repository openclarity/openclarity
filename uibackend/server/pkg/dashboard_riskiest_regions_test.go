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
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gotest.tools/v3/assert"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/cli/pkg/utils"
	"github.com/openclarity/vmclarity/uibackend/types"
)

func Test_getAssetLocation(t *testing.T) {
	assetInfo := apitypes.AssetType{}
	err := assetInfo.FromVMInfo(apitypes.VMInfo{
		InstanceProvider: utils.PointerTo(apitypes.AWS),
		Location:         "us-east-1/vpcid-1/sg-1",
	})
	assert.NilError(t, err)
	nonSupportedAssetInfo := apitypes.AssetType{}
	err = nonSupportedAssetInfo.FromDirInfo(apitypes.DirInfo{})
	assert.NilError(t, err)

	type args struct {
		asset apitypes.Asset
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				asset: apitypes.Asset{
					AssetInfo: &assetInfo,
				},
			},
			want:    "us-east-1",
			wantErr: false,
		},
		{
			name: "non supported asset",
			args: args{
				asset: apitypes.Asset{
					AssetInfo: &nonSupportedAssetInfo,
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getAssetRegion(tt.args.asset)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAssetRegion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getAssetRegion() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_addAssetSummaryToFindingsCount(t *testing.T) {
	type args struct {
		findingsCount *types.FindingsCount
		summary       *apitypes.ScanFindingsSummary
	}
	tests := []struct {
		name string
		args args
		want *types.FindingsCount
	}{
		{
			name: "nil",
			args: args{
				findingsCount: &types.FindingsCount{
					Exploits:          utils.PointerTo(1),
					Malware:           utils.PointerTo(2),
					Misconfigurations: utils.PointerTo(3),
					Rootkits:          utils.PointerTo(4),
					Secrets:           utils.PointerTo(5),
					Vulnerabilities:   utils.PointerTo(6),
				},
				summary: nil,
			},
			want: &types.FindingsCount{
				Exploits:          utils.PointerTo(1),
				Malware:           utils.PointerTo(2),
				Misconfigurations: utils.PointerTo(3),
				Rootkits:          utils.PointerTo(4),
				Secrets:           utils.PointerTo(5),
				Vulnerabilities:   utils.PointerTo(6),
			},
		},
		{
			name: "sanity",
			args: args{
				findingsCount: &types.FindingsCount{
					Exploits:          utils.PointerTo(1),
					Malware:           utils.PointerTo(2),
					Misconfigurations: utils.PointerTo(3),
					Rootkits:          utils.PointerTo(4),
					Secrets:           utils.PointerTo(5),
					Vulnerabilities:   utils.PointerTo(6),
				},
				summary: &apitypes.ScanFindingsSummary{
					TotalExploits:          utils.PointerTo(2),
					TotalMalware:           utils.PointerTo(3),
					TotalMisconfigurations: utils.PointerTo(4),
					TotalPackages:          utils.PointerTo(5),
					TotalRootkits:          utils.PointerTo(6),
					TotalSecrets:           utils.PointerTo(7),
					TotalVulnerabilities: &apitypes.VulnerabilityScanSummary{
						TotalCriticalVulnerabilities:   utils.PointerTo(10),
						TotalHighVulnerabilities:       utils.PointerTo(11),
						TotalLowVulnerabilities:        utils.PointerTo(12),
						TotalMediumVulnerabilities:     utils.PointerTo(13),
						TotalNegligibleVulnerabilities: utils.PointerTo(14),
					},
				},
			},
			want: &types.FindingsCount{
				Exploits:          utils.PointerTo(3),
				Malware:           utils.PointerTo(5),
				Misconfigurations: utils.PointerTo(7),
				Rootkits:          utils.PointerTo(10),
				Secrets:           utils.PointerTo(12),
				Vulnerabilities:   utils.PointerTo(66),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addAssetSummaryToFindingsCount(tt.args.findingsCount, tt.args.summary); !reflect.DeepEqual(got, tt.want) {
				gotB, _ := json.Marshal(got)
				wantB, _ := json.Marshal(tt.want)
				t.Errorf("addAssetSummaryToFindingsCount() = %v, want %v", string(gotB), string(wantB))
			}
		})
	}
}

func Test_getTotalVulnerabilities(t *testing.T) {
	type args struct {
		summary *apitypes.VulnerabilityScanSummary
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "nil",
			args: args{
				summary: nil,
			},
			want: 0,
		},
		{
			name: "sanity",
			args: args{
				summary: &apitypes.VulnerabilityScanSummary{
					TotalCriticalVulnerabilities:   utils.PointerTo(1),
					TotalHighVulnerabilities:       utils.PointerTo(2),
					TotalLowVulnerabilities:        utils.PointerTo(3),
					TotalMediumVulnerabilities:     utils.PointerTo(4),
					TotalNegligibleVulnerabilities: utils.PointerTo(5),
				},
			},
			want: 15,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTotalVulnerabilities(tt.args.summary); got != tt.want {
				t.Errorf("getTotalVulnerabilities() = %v, want %v", got, tt.want)
			}
		})
	}
}

// nolint:errcheck
func Test_createRegionFindingsFromAssets(t *testing.T) {
	dirAsset := apitypes.AssetType{}
	dirAsset.FromDirInfo(apitypes.DirInfo{
		DirName:  utils.PointerTo("test-name"),
		Location: utils.PointerTo("location-test"),
	})

	vmFromRegion1 := apitypes.AssetType{}
	vmFromRegion1.FromVMInfo(apitypes.VMInfo{
		Location: "region1",
	})
	vm1FromRegion2 := apitypes.AssetType{}
	vm1FromRegion2.FromVMInfo(apitypes.VMInfo{
		Location: "region2",
	})
	type args struct {
		assets *apitypes.Assets
	}
	tests := []struct {
		name string
		args args
		want []types.RegionFindings
	}{
		{
			name: "Unsupported asset is skipped",
			args: args{
				assets: &apitypes.Assets{
					Count: utils.PointerTo(1),
					Items: utils.PointerTo([]apitypes.Asset{
						{
							Summary: &apitypes.ScanFindingsSummary{
								TotalExploits:          utils.PointerTo(1),
								TotalMalware:           utils.PointerTo(1),
								TotalMisconfigurations: utils.PointerTo(1),
								TotalPackages:          utils.PointerTo(1),
								TotalRootkits:          utils.PointerTo(1),
								TotalSecrets:           utils.PointerTo(1),
								TotalVulnerabilities: &apitypes.VulnerabilityScanSummary{
									TotalCriticalVulnerabilities:   utils.PointerTo(1),
									TotalHighVulnerabilities:       utils.PointerTo(1),
									TotalLowVulnerabilities:        utils.PointerTo(1),
									TotalMediumVulnerabilities:     utils.PointerTo(1),
									TotalNegligibleVulnerabilities: utils.PointerTo(1),
								},
							},
							AssetInfo: &dirAsset,
						},
					}),
				},
			},
			want: []types.RegionFindings{},
		},
		{
			name: "sanity",
			args: args{
				assets: &apitypes.Assets{
					Count: utils.PointerTo(3),
					Items: &[]apitypes.Asset{
						{
							Summary: &apitypes.ScanFindingsSummary{
								TotalExploits:          utils.PointerTo(1),
								TotalMalware:           utils.PointerTo(2),
								TotalMisconfigurations: utils.PointerTo(3),
								TotalPackages:          utils.PointerTo(4),
								TotalRootkits:          utils.PointerTo(5),
								TotalSecrets:           utils.PointerTo(6),
								TotalVulnerabilities: &apitypes.VulnerabilityScanSummary{
									TotalCriticalVulnerabilities:   utils.PointerTo(7),
									TotalHighVulnerabilities:       utils.PointerTo(8),
									TotalLowVulnerabilities:        utils.PointerTo(9),
									TotalMediumVulnerabilities:     utils.PointerTo(10),
									TotalNegligibleVulnerabilities: utils.PointerTo(11),
								},
							},
							AssetInfo: &vmFromRegion1,
						},
						{
							Summary: &apitypes.ScanFindingsSummary{
								TotalExploits:          utils.PointerTo(2),
								TotalMalware:           utils.PointerTo(3),
								TotalMisconfigurations: utils.PointerTo(4),
								TotalPackages:          utils.PointerTo(5),
								TotalRootkits:          utils.PointerTo(6),
								TotalSecrets:           utils.PointerTo(7),
								TotalVulnerabilities: &apitypes.VulnerabilityScanSummary{
									TotalCriticalVulnerabilities:   utils.PointerTo(8),
									TotalHighVulnerabilities:       utils.PointerTo(9),
									TotalLowVulnerabilities:        utils.PointerTo(10),
									TotalMediumVulnerabilities:     utils.PointerTo(11),
									TotalNegligibleVulnerabilities: utils.PointerTo(12),
								},
							},
							AssetInfo: &vmFromRegion1,
						},
						{
							Summary: &apitypes.ScanFindingsSummary{
								TotalExploits:          utils.PointerTo(3),
								TotalMalware:           utils.PointerTo(4),
								TotalMisconfigurations: utils.PointerTo(5),
								TotalPackages:          utils.PointerTo(6),
								TotalRootkits:          utils.PointerTo(7),
								TotalSecrets:           utils.PointerTo(8),
								TotalVulnerabilities: &apitypes.VulnerabilityScanSummary{
									TotalCriticalVulnerabilities:   utils.PointerTo(9),
									TotalHighVulnerabilities:       utils.PointerTo(10),
									TotalLowVulnerabilities:        utils.PointerTo(11),
									TotalMediumVulnerabilities:     utils.PointerTo(12),
									TotalNegligibleVulnerabilities: utils.PointerTo(13),
								},
							},
							AssetInfo: &vm1FromRegion2,
						},
					},
				},
			},
			want: []types.RegionFindings{
				{
					FindingsCount: &types.FindingsCount{
						Exploits:          utils.PointerTo(3),
						Malware:           utils.PointerTo(5),
						Misconfigurations: utils.PointerTo(7),
						Rootkits:          utils.PointerTo(11),
						Secrets:           utils.PointerTo(13),
						Vulnerabilities:   utils.PointerTo(95),
					},
					RegionName: utils.PointerTo("region1"),
				},
				{
					FindingsCount: &types.FindingsCount{
						Exploits:          utils.PointerTo(3),
						Malware:           utils.PointerTo(4),
						Misconfigurations: utils.PointerTo(5),
						Rootkits:          utils.PointerTo(7),
						Secrets:           utils.PointerTo(8),
						Vulnerabilities:   utils.PointerTo(55),
					},
					RegionName: utils.PointerTo("region2"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createRegionFindingsFromAssets(tt.args.assets)
			if diff := cmp.Diff(tt.want, got, cmpopts.SortSlices(func(a, b types.RegionFindings) bool { return *a.RegionName < *b.RegionName })); diff != "" {
				t.Errorf("createRegionFindingsFromAssets() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_getRegionByProvider(t *testing.T) {
	type args struct {
		info apitypes.VMInfo
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "cloud provider is nil",
			args: args{
				info: apitypes.VMInfo{
					InstanceProvider: nil,
					Location:         "eu-central-1/vpc-1",
				},
			},
			want: "eu-central-1/vpc-1",
		},
		{
			name: "AWS cloud provider",
			args: args{
				info: apitypes.VMInfo{
					InstanceProvider: utils.PointerTo(apitypes.AWS),
					Location:         "eu-central-1/vpc-1",
				},
			},
			want: "eu-central-1",
		},
		{
			name: "non AWS cloud provider",
			args: args{
				info: apitypes.VMInfo{
					InstanceProvider: utils.PointerTo(apitypes.CloudProvider("GCP")),
					Location:         "eu-central-1/vpc-1",
				},
			},
			want: "eu-central-1/vpc-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRegionByProvider(tt.args.info); got != tt.want {
				t.Errorf("getRegionByProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}
