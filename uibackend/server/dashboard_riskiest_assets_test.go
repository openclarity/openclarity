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
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gotest.tools/v3/assert"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/uibackend/types"
)

func Test_getTotalFindingField(t *testing.T) {
	type args struct {
		findingType apitypes.ScanType
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "non supported finding type",
			args: args{
				findingType: "sboms",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "EXPLOIT",
			args: args{
				findingType: apitypes.EXPLOIT,
			},
			want:    totalExploitsSummaryFieldName,
			wantErr: false,
		},
		{
			name: "MALWARE",
			args: args{
				findingType: apitypes.MALWARE,
			},
			want:    totalMalwareSummaryFieldName,
			wantErr: false,
		},
		{
			name: "MISCONFIGURATION",
			args: args{
				findingType: apitypes.MISCONFIGURATION,
			},
			want:    totalMisconfigurationsSummaryFieldName,
			wantErr: false,
		},
		{
			name: "ROOTKIT",
			args: args{
				findingType: apitypes.ROOTKIT,
			},
			want:    totalRootkitsSummaryFieldName,
			wantErr: false,
		},
		{
			name: "SECRET",
			args: args{
				findingType: apitypes.SECRET,
			},
			want:    totalSecretsSummaryFieldName,
			wantErr: false,
		},
		{
			name: "VULNERABILITY",
			args: args{
				findingType: apitypes.VULNERABILITY,
			},
			want:    totalVulnerabilitiesSummaryFieldName,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getTotalFindingFieldName(tt.args.findingType)
			if (err != nil) != tt.wantErr {
				t.Errorf("getTotalFindingFieldName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getTotalFindingFieldName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getCount(t *testing.T) {
	type args struct {
		summary     *apitypes.ScanFindingsSummary
		findingType apitypes.ScanType
	}
	tests := []struct {
		name    string
		args    args
		want    *int
		wantErr bool
	}{
		{
			name: "unsupported finding type",
			args: args{
				summary: &apitypes.ScanFindingsSummary{
					TotalExploits:          nil,
					TotalMalware:           nil,
					TotalMisconfigurations: nil,
					TotalPackages:          nil,
					TotalRootkits:          nil,
					TotalSecrets:           nil,
					TotalVulnerabilities: &apitypes.VulnerabilitySeveritySummary{
						TotalCriticalVulnerabilities:   nil,
						TotalHighVulnerabilities:       nil,
						TotalLowVulnerabilities:        nil,
						TotalMediumVulnerabilities:     nil,
						TotalNegligibleVulnerabilities: nil,
					},
				},
				findingType: "unsupported finding type",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "TotalExploits",
			args: args{
				summary: &apitypes.ScanFindingsSummary{
					TotalExploits: to.Ptr(1),
				},
				findingType: apitypes.EXPLOIT,
			},
			want:    to.Ptr(1),
			wantErr: false,
		},
		{
			name: "TotalMalware",
			args: args{
				summary: &apitypes.ScanFindingsSummary{
					TotalMalware: to.Ptr(1),
				},
				findingType: apitypes.MALWARE,
			},
			want:    to.Ptr(1),
			wantErr: false,
		},
		{
			name: "TotalMisconfigurations",
			args: args{
				summary: &apitypes.ScanFindingsSummary{
					TotalMisconfigurations: to.Ptr(1),
				},
				findingType: apitypes.MISCONFIGURATION,
			},
			want:    to.Ptr(1),
			wantErr: false,
		},
		{
			name: "TotalRootkits",
			args: args{
				summary: &apitypes.ScanFindingsSummary{
					TotalRootkits: to.Ptr(1),
				},
				findingType: apitypes.ROOTKIT,
			},
			want:    to.Ptr(1),
			wantErr: false,
		},
		{
			name: "TotalSecrets",
			args: args{
				summary: &apitypes.ScanFindingsSummary{
					TotalSecrets: to.Ptr(1),
				},
				findingType: apitypes.SECRET,
			},
			want:    to.Ptr(1),
			wantErr: false,
		},
		{
			name: "TotalInfoFinder - unsupported",
			args: args{
				summary: &apitypes.ScanFindingsSummary{
					TotalInfoFinder: to.Ptr(1),
				},
				findingType: apitypes.INFOFINDER,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getCountForFindingType(tt.args.summary, tt.args.findingType)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCountForFindingType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCountForFindingType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getAssetInfo(t *testing.T) {
	type args struct {
		asset *apitypes.AssetType
	}
	tests := []struct {
		name    string
		args    args
		want    *types.AssetInfo
		wantErr bool
	}{
		{
			name: "unsupported asset type",
			args: args{
				asset: createPodInfo(t, "name", "location"),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "VMInfo",
			args: args{
				asset: createVMInfo(t, "name", "location"),
			},
			want: &types.AssetInfo{
				Location: to.Ptr("location"),
				Name:     to.Ptr("name"),
				Type:     to.Ptr(types.AWSEC2Instance),
			},
			wantErr: false,
		},
		{
			name: "ContainerInfo",
			args: args{
				asset: createContainerInfo(t),
			},
			want: &types.AssetInfo{
				Location: to.Ptr("gke-sambetts-dev-clu-sambetts-dev-nod-449204c7-gqx5"),
				Name:     to.Ptr("hungry_mcclintock"),
				Type:     to.Ptr(types.Container),
			},
			wantErr: false,
		},
		{
			name: "ContainerImageInfo",
			args: args{
				asset: createContainerImageInfo(t),
			},
			want: &types.AssetInfo{
				Location: to.Ptr("ghcr.io/openclarity/vmclarity-orchestrator@sha256:2ceda8090cfb24eb86c6b723eef4a562e90199f3c2b11120e60e5691f957b07b"),
				Name:     to.Ptr("sha256:b520c72cef1f30a38361cf9e3d686e2db0e718b69af8cb072e93ba9bcf5658ab"),
				Type:     to.Ptr(types.ContainerImage),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getAssetInfo(tt.args.asset)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAssetInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("getAssetInfo() %v", diff)
			}
		})
	}
}

func createVMInfo(t *testing.T, instanceID, location string) *apitypes.AssetType {
	t.Helper()
	info := apitypes.AssetType{}
	err := info.FromVMInfo(apitypes.VMInfo{
		InstanceID:       instanceID,
		InstanceProvider: to.Ptr(apitypes.AWS),
		Location:         location,
	})
	assert.NilError(t, err)
	return &info
}

func createContainerInfo(t *testing.T) *apitypes.AssetType {
	t.Helper()
	info := apitypes.AssetType{}
	err := info.FromContainerInfo(apitypes.ContainerInfo{
		ContainerID:   "d66da925d976b8caf60ea59c5ec75b1950f87d506144942cdbf10061052a8088",
		ContainerName: to.Ptr("hungry_mcclintock"),
		Location:      to.Ptr("gke-sambetts-dev-clu-sambetts-dev-nod-449204c7-gqx5"),
	})
	assert.NilError(t, err)
	return &info
}

func createContainerImageInfo(t *testing.T) *apitypes.AssetType {
	t.Helper()
	info := apitypes.AssetType{}
	err := info.FromContainerImageInfo(apitypes.ContainerImageInfo{
		ImageID:     "sha256:b520c72cef1f30a38361cf9e3d686e2db0e718b69af8cb072e93ba9bcf5658ab",
		RepoTags:    to.Ptr([]string{"ghcr.io/openclarity/vmclarity-orchestrator:latest"}),
		RepoDigests: to.Ptr([]string{"ghcr.io/openclarity/vmclarity-orchestrator@sha256:2ceda8090cfb24eb86c6b723eef4a562e90199f3c2b11120e60e5691f957b07b"}),
	})
	assert.NilError(t, err)
	return &info
}

func createPodInfo(t *testing.T, podName, location string) *apitypes.AssetType {
	t.Helper()
	info := apitypes.AssetType{}
	err := info.FromPodInfo(apitypes.PodInfo{
		Location: &location,
		PodName:  &podName,
	})
	assert.NilError(t, err)
	return &info
}

func Test_toAPIVulnerabilityRiskyAsset(t *testing.T) {
	type args struct {
		assets []apitypes.Asset
	}
	tests := []struct {
		name string
		args args
		want []types.VulnerabilityRiskyAsset
	}{
		{
			name: "nil assets",
			args: args{
				assets: nil,
			},
			want: []types.VulnerabilityRiskyAsset{},
		},
		{
			name: "supported and unsupported asset",
			args: args{
				assets: []apitypes.Asset{
					{
						Summary: &apitypes.ScanFindingsSummary{
							TotalVulnerabilities: &apitypes.VulnerabilitySeveritySummary{
								TotalCriticalVulnerabilities:   to.Ptr(1),
								TotalHighVulnerabilities:       to.Ptr(2),
								TotalLowVulnerabilities:        to.Ptr(3),
								TotalMediumVulnerabilities:     to.Ptr(4),
								TotalNegligibleVulnerabilities: to.Ptr(5),
							},
						},
						AssetInfo: createVMInfo(t, "vm name", "vm location"),
					},
					{
						Summary: &apitypes.ScanFindingsSummary{
							TotalVulnerabilities: &apitypes.VulnerabilitySeveritySummary{
								TotalHighVulnerabilities: to.Ptr(1),
							},
						},
						AssetInfo: createPodInfo(t, "pod name", "pod location"),
					},
				},
			},
			want: []types.VulnerabilityRiskyAsset{
				{
					CriticalVulnerabilitiesCount:   to.Ptr(1),
					HighVulnerabilitiesCount:       to.Ptr(2),
					LowVulnerabilitiesCount:        to.Ptr(3),
					MediumVulnerabilitiesCount:     to.Ptr(4),
					NegligibleVulnerabilitiesCount: to.Ptr(5),
					AssetInfo: &types.AssetInfo{
						Location: to.Ptr("vm location"),
						Name:     to.Ptr("vm name"),
						Type:     to.Ptr(types.AWSEC2Instance),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toAPIVulnerabilityRiskyAssets(tt.args.assets); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toAPIVulnerabilityRiskyAssets() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toAPIRiskyAssets(t *testing.T) {
	type args struct {
		assets      []apitypes.Asset
		findingType apitypes.ScanType
	}
	tests := []struct {
		name string
		args args
		want []types.RiskyAsset
	}{
		{
			name: "nil assets",
			args: args{
				assets:      nil,
				findingType: "",
			},
			want: []types.RiskyAsset{},
		},
		{
			name: "supported and unsupported asset",
			args: args{
				assets: []apitypes.Asset{
					{
						Summary: &apitypes.ScanFindingsSummary{
							TotalMalware: to.Ptr(1),
						},
						AssetInfo: createVMInfo(t, "vm name", "vm location"),
					},
					{
						Summary: &apitypes.ScanFindingsSummary{
							TotalMalware: to.Ptr(2),
						},
						AssetInfo: createPodInfo(t, "pod name", "pod location"),
					},
				},
				findingType: apitypes.MALWARE,
			},
			want: []types.RiskyAsset{
				{
					Count: to.Ptr(1),
					AssetInfo: &types.AssetInfo{
						Location: to.Ptr("vm location"),
						Name:     to.Ptr("vm name"),
						Type:     to.Ptr(types.AWSEC2Instance),
					},
				},
			},
		},
		{
			name: "unsupported finding type asset",
			args: args{
				assets: []apitypes.Asset{
					{
						Summary: &apitypes.ScanFindingsSummary{
							TotalVulnerabilities: &apitypes.VulnerabilitySeveritySummary{
								TotalHighVulnerabilities: to.Ptr(1),
							},
						},
						AssetInfo: createVMInfo(t, "name", "location"),
					},
					{
						Summary: &apitypes.ScanFindingsSummary{
							TotalVulnerabilities: &apitypes.VulnerabilitySeveritySummary{
								TotalHighVulnerabilities: to.Ptr(1),
							},
						},
						AssetInfo: createVMInfo(t, "name1", "location1"),
					},
				},
				findingType: "unsupported",
			},
			want: []types.RiskyAsset{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toAPIRiskyAssets(tt.args.assets, tt.args.findingType)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("toAPIRiskyAssets() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

const SummaryQueryTemplate = "summary/%s/%s desc"

func Test_getOrderByOdataForVulnerabilities(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "sanity",
			want: strings.Join(
				[]string{
					fmt.Sprintf(SummaryQueryTemplate, totalVulnerabilitiesSummaryFieldName, totalCriticalVulnerabilitiesSummaryFieldName),
					fmt.Sprintf(SummaryQueryTemplate, totalVulnerabilitiesSummaryFieldName, totalHighVulnerabilitiesSummaryFieldName),
					fmt.Sprintf(SummaryQueryTemplate, totalVulnerabilitiesSummaryFieldName, totalMediumVulnerabilitiesSummaryFieldName),
					fmt.Sprintf(SummaryQueryTemplate, totalVulnerabilitiesSummaryFieldName, totalLowVulnerabilitiesFieldName),
					fmt.Sprintf(SummaryQueryTemplate, totalVulnerabilitiesSummaryFieldName, totalNegligibleVulnerabilitiesFieldName),
				}, ","),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getOrderByOdataForVulnerabilities(); got != tt.want {
				t.Errorf("getOrderByOdataForVulnerabilities() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getOrderByOData(t *testing.T) {
	type args struct {
		totalFindingField string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "vulnerabilities",
			args: args{
				totalFindingField: totalVulnerabilitiesSummaryFieldName,
			},
			want: strings.Join(
				[]string{
					fmt.Sprintf(SummaryQueryTemplate, totalVulnerabilitiesSummaryFieldName, totalCriticalVulnerabilitiesSummaryFieldName),
					fmt.Sprintf(SummaryQueryTemplate, totalVulnerabilitiesSummaryFieldName, totalHighVulnerabilitiesSummaryFieldName),
					fmt.Sprintf(SummaryQueryTemplate, totalVulnerabilitiesSummaryFieldName, totalMediumVulnerabilitiesSummaryFieldName),
					fmt.Sprintf(SummaryQueryTemplate, totalVulnerabilitiesSummaryFieldName, totalLowVulnerabilitiesFieldName),
					fmt.Sprintf(SummaryQueryTemplate, totalVulnerabilitiesSummaryFieldName, totalNegligibleVulnerabilitiesFieldName),
				}, ","),
		},
		{
			name: "not vulnerabilities",
			args: args{
				totalFindingField: totalMisconfigurationsSummaryFieldName,
			},
			want: "summary/totalMisconfigurations desc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getOrderByOData(tt.args.totalFindingField); got != tt.want {
				t.Errorf("getOrderByOData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_vmInfoToAssetInfo(t *testing.T) {
	type args struct {
		info apitypes.VMInfo
	}
	tests := []struct {
		name    string
		args    args
		want    *types.AssetInfo
		wantErr bool
	}{
		{
			name: "unsupported provider",
			args: args{
				info: apitypes.VMInfo{
					InstanceProvider: nil,
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "AWS EC2 Instance",
			args: args{
				info: apitypes.VMInfo{
					InstanceID:       "name",
					InstanceProvider: to.Ptr(apitypes.AWS),
					Location:         "location",
				},
			},
			want: &types.AssetInfo{
				Location: to.Ptr("location"),
				Name:     to.Ptr("name"),
				Type:     to.Ptr(types.AWSEC2Instance),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vmInfoToAssetInfo(tt.args.info)
			if (err != nil) != tt.wantErr {
				t.Errorf("vmInfoToAssetInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("vmInfoToAssetInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getVMAssetType(t *testing.T) {
	type args struct {
		provider *apitypes.CloudProvider
	}
	tests := []struct {
		name    string
		args    args
		want    *types.AssetType
		wantErr bool
	}{
		{
			name: "nil provider",
			args: args{
				provider: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "unsupported provider",
			args: args{
				provider: to.Ptr(apitypes.CloudProvider("unsupported provider")),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "aws provider",
			args: args{
				provider: to.Ptr(apitypes.AWS),
			},
			want:    to.Ptr(types.AWSEC2Instance),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getVMAssetType(tt.args.provider)
			if (err != nil) != tt.wantErr {
				t.Errorf("getVMAssetType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getVMAssetType() got = %v, want %v", got, tt.want)
			}
		})
	}
}
