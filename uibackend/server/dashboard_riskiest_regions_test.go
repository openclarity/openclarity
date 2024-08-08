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
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/uibackend/types"
)

func Test_getAssetLocation(t *testing.T) {
	assetInfo := apitypes.AssetType{}
	err := assetInfo.FromVMInfo(apitypes.VMInfo{
		InstanceProvider: to.Ptr(apitypes.AWS),
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
					Exploits:          to.Ptr(1),
					Malware:           to.Ptr(2),
					Misconfigurations: to.Ptr(3),
					Rootkits:          to.Ptr(4),
					Secrets:           to.Ptr(5),
					Vulnerabilities:   to.Ptr(6),
				},
				summary: nil,
			},
			want: &types.FindingsCount{
				Exploits:          to.Ptr(1),
				Malware:           to.Ptr(2),
				Misconfigurations: to.Ptr(3),
				Rootkits:          to.Ptr(4),
				Secrets:           to.Ptr(5),
				Vulnerabilities:   to.Ptr(6),
			},
		},
		{
			name: "sanity",
			args: args{
				findingsCount: &types.FindingsCount{
					Exploits:          to.Ptr(1),
					Malware:           to.Ptr(2),
					Misconfigurations: to.Ptr(3),
					Rootkits:          to.Ptr(4),
					Secrets:           to.Ptr(5),
					Vulnerabilities:   to.Ptr(6),
				},
				summary: &apitypes.ScanFindingsSummary{
					TotalExploits:          to.Ptr(2),
					TotalMalware:           to.Ptr(3),
					TotalMisconfigurations: to.Ptr(4),
					TotalPackages:          to.Ptr(5),
					TotalRootkits:          to.Ptr(6),
					TotalSecrets:           to.Ptr(7),
					TotalVulnerabilities: &apitypes.VulnerabilitySeveritySummary{
						TotalCriticalVulnerabilities:   to.Ptr(10),
						TotalHighVulnerabilities:       to.Ptr(11),
						TotalLowVulnerabilities:        to.Ptr(12),
						TotalMediumVulnerabilities:     to.Ptr(13),
						TotalNegligibleVulnerabilities: to.Ptr(14),
					},
				},
			},
			want: &types.FindingsCount{
				Exploits:          to.Ptr(3),
				Malware:           to.Ptr(5),
				Misconfigurations: to.Ptr(7),
				Rootkits:          to.Ptr(10),
				Secrets:           to.Ptr(12),
				Vulnerabilities:   to.Ptr(66),
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
		summary *apitypes.VulnerabilitySeveritySummary
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
				summary: &apitypes.VulnerabilitySeveritySummary{
					TotalCriticalVulnerabilities:   to.Ptr(1),
					TotalHighVulnerabilities:       to.Ptr(2),
					TotalLowVulnerabilities:        to.Ptr(3),
					TotalMediumVulnerabilities:     to.Ptr(4),
					TotalNegligibleVulnerabilities: to.Ptr(5),
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
		DirName:  to.Ptr("test-name"),
		Location: to.Ptr("location-test"),
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
					Count: to.Ptr(1),
					Items: to.Ptr([]apitypes.Asset{
						{
							Summary: &apitypes.ScanFindingsSummary{
								TotalExploits:          to.Ptr(1),
								TotalMalware:           to.Ptr(1),
								TotalMisconfigurations: to.Ptr(1),
								TotalPackages:          to.Ptr(1),
								TotalRootkits:          to.Ptr(1),
								TotalSecrets:           to.Ptr(1),
								TotalVulnerabilities: &apitypes.VulnerabilitySeveritySummary{
									TotalCriticalVulnerabilities:   to.Ptr(1),
									TotalHighVulnerabilities:       to.Ptr(1),
									TotalLowVulnerabilities:        to.Ptr(1),
									TotalMediumVulnerabilities:     to.Ptr(1),
									TotalNegligibleVulnerabilities: to.Ptr(1),
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
					Count: to.Ptr(3),
					Items: &[]apitypes.Asset{
						{
							Summary: &apitypes.ScanFindingsSummary{
								TotalExploits:          to.Ptr(1),
								TotalMalware:           to.Ptr(2),
								TotalMisconfigurations: to.Ptr(3),
								TotalPackages:          to.Ptr(4),
								TotalRootkits:          to.Ptr(5),
								TotalSecrets:           to.Ptr(6),
								TotalVulnerabilities: &apitypes.VulnerabilitySeveritySummary{
									TotalCriticalVulnerabilities:   to.Ptr(7),
									TotalHighVulnerabilities:       to.Ptr(8),
									TotalLowVulnerabilities:        to.Ptr(9),
									TotalMediumVulnerabilities:     to.Ptr(10),
									TotalNegligibleVulnerabilities: to.Ptr(11),
								},
							},
							AssetInfo: &vmFromRegion1,
						},
						{
							Summary: &apitypes.ScanFindingsSummary{
								TotalExploits:          to.Ptr(2),
								TotalMalware:           to.Ptr(3),
								TotalMisconfigurations: to.Ptr(4),
								TotalPackages:          to.Ptr(5),
								TotalRootkits:          to.Ptr(6),
								TotalSecrets:           to.Ptr(7),
								TotalVulnerabilities: &apitypes.VulnerabilitySeveritySummary{
									TotalCriticalVulnerabilities:   to.Ptr(8),
									TotalHighVulnerabilities:       to.Ptr(9),
									TotalLowVulnerabilities:        to.Ptr(10),
									TotalMediumVulnerabilities:     to.Ptr(11),
									TotalNegligibleVulnerabilities: to.Ptr(12),
								},
							},
							AssetInfo: &vmFromRegion1,
						},
						{
							Summary: &apitypes.ScanFindingsSummary{
								TotalExploits:          to.Ptr(3),
								TotalMalware:           to.Ptr(4),
								TotalMisconfigurations: to.Ptr(5),
								TotalPackages:          to.Ptr(6),
								TotalRootkits:          to.Ptr(7),
								TotalSecrets:           to.Ptr(8),
								TotalVulnerabilities: &apitypes.VulnerabilitySeveritySummary{
									TotalCriticalVulnerabilities:   to.Ptr(9),
									TotalHighVulnerabilities:       to.Ptr(10),
									TotalLowVulnerabilities:        to.Ptr(11),
									TotalMediumVulnerabilities:     to.Ptr(12),
									TotalNegligibleVulnerabilities: to.Ptr(13),
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
						Exploits:          to.Ptr(3),
						Malware:           to.Ptr(5),
						Misconfigurations: to.Ptr(7),
						Rootkits:          to.Ptr(11),
						Secrets:           to.Ptr(13),
						Vulnerabilities:   to.Ptr(95),
					},
					RegionName: to.Ptr("region1"),
				},
				{
					FindingsCount: &types.FindingsCount{
						Exploits:          to.Ptr(3),
						Malware:           to.Ptr(4),
						Misconfigurations: to.Ptr(5),
						Rootkits:          to.Ptr(7),
						Secrets:           to.Ptr(8),
						Vulnerabilities:   to.Ptr(55),
					},
					RegionName: to.Ptr("region2"),
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
					InstanceProvider: to.Ptr(apitypes.AWS),
					Location:         "eu-central-1/vpc-1",
				},
			},
			want: "eu-central-1",
		},
		{
			name: "non AWS cloud provider",
			args: args{
				info: apitypes.VMInfo{
					InstanceProvider: to.Ptr(apitypes.CloudProvider("GCP")),
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
