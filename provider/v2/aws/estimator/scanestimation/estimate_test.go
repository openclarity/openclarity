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

package scanestimation

import (
	"context"
	"reflect"
	"testing"
	"time"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"gotest.tools/v3/assert"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/scanner/utils"
)

type FakePriceFetcher struct{}

func (f *FakePriceFetcher) GetSnapshotMonthlyCostPerGB(ctx context.Context, regionCode string) (float64, error) {
	return 0.2, nil
}

func (f *FakePriceFetcher) GetVolumeMonthlyCostPerGB(ctx context.Context, regionCode string, volumeType ec2types.VolumeType) (float64, error) {
	return 0.1, nil
}

func (f *FakePriceFetcher) GetDataTransferCostPerGB(sourceRegion, destRegion string) (float64, error) {
	if sourceRegion == destRegion {
		return 0, nil
	}
	return 1.3, nil
}

func (f *FakePriceFetcher) GetInstancePerHourCost(ctx context.Context, regionCode string, instanceType ec2types.InstanceType, marketOption MarketOption) (float64, error) {
	return 0.6, nil
}

func Test_getScanSize(t *testing.T) {
	const (
		RootVolumeSize = 5
	)

	var assetType apitypes.AssetType
	err := assetType.FromVMInfo(apitypes.VMInfo{
		RootVolume: apitypes.RootVolume{
			SizeGB: RootVolumeSize,
		},
	})
	assert.NilError(t, err)

	type args struct {
		stats apitypes.AssetScanStats
		asset *apitypes.Asset
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "size found from first family (Sbom) stats",
			args: args{
				stats: apitypes.AssetScanStats{
					Sbom: &[]apitypes.AssetScanInputScanStats{
						{
							Path: to.Ptr("/"),
							Size: to.Ptr(int64(10)),
							Type: to.Ptr(string(utils.ROOTFS)),
						},
					},
				},
				asset: &apitypes.Asset{
					AssetInfo: &assetType,
				},
			},
			want:    10,
			wantErr: false,
		},
		{
			name: "size found not from first family (Malware) stats",
			args: args{
				stats: apitypes.AssetScanStats{
					Sbom: &[]apitypes.AssetScanInputScanStats{
						{
							Path: to.Ptr("/dir"),
							Size: to.Ptr(int64(3)),
							Type: to.Ptr(string(utils.DIR)),
						},
					},
					Malware: &[]apitypes.AssetScanInputScanStats{
						{
							Path: to.Ptr("/"),
							Size: to.Ptr(int64(10)),
							Type: to.Ptr(string(utils.ROOTFS)),
						},
					},
				},
				asset: &apitypes.Asset{
					AssetInfo: &assetType,
				},
			},
			want:    10,
			wantErr: false,
		},
		{
			name: "size not found from stats, get it from root volume size",
			args: args{
				stats: apitypes.AssetScanStats{},
				asset: &apitypes.Asset{
					AssetInfo: &assetType,
				},
			},
			want:    (RootVolumeSize * 1000) / 2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getScanSize(tt.args.stats, tt.args.asset)
			if (err != nil) != tt.wantErr {
				t.Errorf("getScanSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getScanSize() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getScanDuration(t *testing.T) {
	timeNow := time.Now()

	type args struct {
		stats          apitypes.AssetScanStats
		familiesConfig *apitypes.ScanFamiliesConfig
		scanSizeMB     int64
	}
	tests := []struct {
		name         string
		args         args
		wantDuration int64
	}{
		{
			name: "Sbom and Secrets has stats, the other scan durations will be taken from the static map",
			args: args{
				stats: apitypes.AssetScanStats{
					Sbom: &[]apitypes.AssetScanInputScanStats{
						{
							Path: to.Ptr("/"),
							ScanTime: &apitypes.AssetScanScanTime{
								EndTime:   &timeNow,
								StartTime: to.Ptr(timeNow.Add(-50 * time.Second)),
							},
							Type: to.Ptr(string(utils.ROOTFS)),
						},
					},
					Secrets: &[]apitypes.AssetScanInputScanStats{
						{
							Path: to.Ptr("/"),
							ScanTime: &apitypes.AssetScanScanTime{
								EndTime:   &timeNow,
								StartTime: to.Ptr(timeNow.Add(-360 * time.Second)),
							},
							Type: to.Ptr(string(utils.ROOTFS)),
						},
					},
				},
				familiesConfig: &apitypes.ScanFamiliesConfig{
					Misconfigurations: &apitypes.MisconfigurationsConfig{
						Enabled: to.Ptr(true),
					},
					Secrets: &apitypes.SecretsConfig{
						Enabled: to.Ptr(true),
					},
					Sbom: &apitypes.SBOMConfig{
						Enabled: to.Ptr(true),
					},
					Vulnerabilities: &apitypes.VulnerabilitiesConfig{
						Enabled: to.Ptr(true),
					},
					Malware: &apitypes.MalwareConfig{
						Enabled: to.Ptr(true),
					},
				},
				scanSizeMB: 2500,
			},
			// 360 seconds Secrets scan from stats
			// 50 seconds Sbom scan from stats
			// extrapolated value for  Misconfigurations, Malware and Vulnerabilities from static lab tests.
			wantDuration: 1953,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDuration := getScanDuration(tt.args.stats, tt.args.familiesConfig, tt.args.scanSizeMB)
			if gotDuration != tt.wantDuration {
				t.Errorf("getScanDuration() gotDuration = %v, want %v", gotDuration, tt.wantDuration)
			}
		})
	}
}

func TestScanEstimator_EstimateAssetScan(t *testing.T) {
	var assetType apitypes.AssetType
	err := assetType.FromVMInfo(apitypes.VMInfo{
		RootVolume: apitypes.RootVolume{
			SizeGB: 16,
		},
	})
	assert.NilError(t, err)
	timeNow := time.Now()
	fakePriceFetcher := FakePriceFetcher{}

	type fields struct {
		priceFetcher PriceFetcher
	}
	type args struct {
		params EstimateAssetScanParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *apitypes.Estimation
		wantErr bool
	}{
		{
			name: "Same region",
			fields: fields{
				priceFetcher: &fakePriceFetcher,
			},
			args: args{
				params: EstimateAssetScanParams{
					SourceRegion:            "us-east-1",
					DestRegion:              "us-east-1",
					ScannerVolumeType:       ec2types.VolumeTypeGp2,
					FromSnapshotVolumeType:  ec2types.VolumeTypeGp2,
					ScannerInstanceType:     ec2types.InstanceTypeT2Large,
					JobCreationTimeSec:      20 * 60,
					ScannerRootVolumeSizeGB: 8,
					Stats: apitypes.AssetScanStats{
						Sbom: &[]apitypes.AssetScanInputScanStats{
							{
								Path: to.Ptr("/"),
								ScanTime: &apitypes.AssetScanScanTime{
									EndTime:   &timeNow,
									StartTime: to.Ptr(timeNow.Add(-50 * time.Second)),
								},
								Type: to.Ptr(string(utils.ROOTFS)),
							},
						},
						Secrets: &[]apitypes.AssetScanInputScanStats{
							{
								Path: to.Ptr("/"),
								ScanTime: &apitypes.AssetScanScanTime{
									EndTime:   &timeNow,
									StartTime: to.Ptr(timeNow.Add(-360 * time.Second)),
								},
								Type: to.Ptr(string(utils.ROOTFS)),
							},
						},
					},
					Asset: &apitypes.Asset{
						AssetInfo: &assetType,
						Id:        to.Ptr("id"),
					},
					AssetScanTemplate: &apitypes.AssetScanTemplate{
						ScanFamiliesConfig: &apitypes.ScanFamiliesConfig{
							Misconfigurations: &apitypes.MisconfigurationsConfig{
								Enabled: to.Ptr(true),
							},
							Secrets: &apitypes.SecretsConfig{
								Enabled: to.Ptr(true),
							},
							Sbom: &apitypes.SBOMConfig{
								Enabled: to.Ptr(true),
							},
							Vulnerabilities: &apitypes.VulnerabilitiesConfig{
								Enabled: to.Ptr(true),
							},
							Malware: &apitypes.MalwareConfig{
								Enabled: to.Ptr(true),
							},
						},
						ScannerInstanceCreationConfig: nil,
					},
				},
			},
			want: &apitypes.Estimation{
				Cost: to.Ptr(float32(0.5883259)),
				CostBreakdown: &[]apitypes.CostBreakdownComponent{
					{
						Cost:      float32(0.002162963),
						Operation: string(Snapshot + "-us-east-1"),
					},
					{
						Cost:      float32(0.584),
						Operation: string(ScannerInstance),
					},
					{
						Cost:      float32(0.0010814815),
						Operation: string(VolumeFromSnapshot),
					},
					{
						Cost:      float32(0.0010814815),
						Operation: string(ScannerRootVolume),
					},
				},
				Size:     to.Ptr(8),
				Duration: to.Ptr(2304),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScanEstimator{
				priceFetcher: tt.fields.priceFetcher,
			}
			got, err := s.EstimateAssetScan(context.TODO(), tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("EstimateAssetScan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NilError(t, err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EstimateAssetScan() got = %v, want %v", got, tt.want)
			}
		})
	}
}
