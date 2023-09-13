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
	kubeclarityUtils "github.com/openclarity/kubeclarity/shared/pkg/utils"
	"gotest.tools/v3/assert"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
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

	var assetType models.AssetType
	err := assetType.FromVMInfo(models.VMInfo{
		RootVolume: models.RootVolume{
			SizeGB: RootVolumeSize,
		},
	})
	assert.NilError(t, err)

	type args struct {
		stats models.AssetScanStats
		asset *models.Asset
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
				stats: models.AssetScanStats{
					Sbom: &[]models.AssetScanInputScanStats{
						{
							Path: utils.PointerTo("/"),
							Size: utils.PointerTo(int64(10)),
							Type: utils.PointerTo(string(kubeclarityUtils.ROOTFS)),
						},
					},
				},
				asset: &models.Asset{
					AssetInfo: &assetType,
				},
			},
			want:    10,
			wantErr: false,
		},
		{
			name: "size found not from first family (Malware) stats",
			args: args{
				stats: models.AssetScanStats{
					Sbom: &[]models.AssetScanInputScanStats{
						{
							Path: utils.PointerTo("/dir"),
							Size: utils.PointerTo(int64(3)),
							Type: utils.PointerTo(string(kubeclarityUtils.DIR)),
						},
					},
					Malware: &[]models.AssetScanInputScanStats{
						{
							Path: utils.PointerTo("/"),
							Size: utils.PointerTo(int64(10)),
							Type: utils.PointerTo(string(kubeclarityUtils.ROOTFS)),
						},
					},
				},
				asset: &models.Asset{
					AssetInfo: &assetType,
				},
			},
			want:    10,
			wantErr: false,
		},
		{
			name: "size not found from stats, get it from root volume size",
			args: args{
				stats: models.AssetScanStats{},
				asset: &models.Asset{
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
		stats          models.AssetScanStats
		familiesConfig *models.ScanFamiliesConfig
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
				stats: models.AssetScanStats{
					Sbom: &[]models.AssetScanInputScanStats{
						{
							Path: utils.PointerTo("/"),
							ScanTime: &models.AssetScanScanTime{
								EndTime:   &timeNow,
								StartTime: utils.PointerTo(timeNow.Add(-50 * time.Second)),
							},
							Type: utils.PointerTo(string(kubeclarityUtils.ROOTFS)),
						},
					},
					Secrets: &[]models.AssetScanInputScanStats{
						{
							Path: utils.PointerTo("/"),
							ScanTime: &models.AssetScanScanTime{
								EndTime:   &timeNow,
								StartTime: utils.PointerTo(timeNow.Add(-360 * time.Second)),
							},
							Type: utils.PointerTo(string(kubeclarityUtils.ROOTFS)),
						},
					},
				},
				familiesConfig: &models.ScanFamiliesConfig{
					Misconfigurations: &models.MisconfigurationsConfig{
						Enabled: utils.PointerTo(true),
					},
					Secrets: &models.SecretsConfig{
						Enabled: utils.PointerTo(true),
					},
					Sbom: &models.SBOMConfig{
						Enabled: utils.PointerTo(true),
					},
					Vulnerabilities: &models.VulnerabilitiesConfig{
						Enabled: utils.PointerTo(true),
					},
					Malware: &models.MalwareConfig{
						Enabled: utils.PointerTo(true),
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
	var assetType models.AssetType
	err := assetType.FromVMInfo(models.VMInfo{
		RootVolume: models.RootVolume{
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
		want    *models.Estimation
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
					Stats: models.AssetScanStats{
						Sbom: &[]models.AssetScanInputScanStats{
							{
								Path: utils.PointerTo("/"),
								ScanTime: &models.AssetScanScanTime{
									EndTime:   &timeNow,
									StartTime: utils.PointerTo(timeNow.Add(-50 * time.Second)),
								},
								Type: utils.PointerTo(string(kubeclarityUtils.ROOTFS)),
							},
						},
						Secrets: &[]models.AssetScanInputScanStats{
							{
								Path: utils.PointerTo("/"),
								ScanTime: &models.AssetScanScanTime{
									EndTime:   &timeNow,
									StartTime: utils.PointerTo(timeNow.Add(-360 * time.Second)),
								},
								Type: utils.PointerTo(string(kubeclarityUtils.ROOTFS)),
							},
						},
					},
					Asset: &models.Asset{
						AssetInfo: &assetType,
						Id:        utils.PointerTo("id"),
					},
					AssetScanTemplate: &models.AssetScanTemplate{
						ScanFamiliesConfig: &models.ScanFamiliesConfig{
							Misconfigurations: &models.MisconfigurationsConfig{
								Enabled: utils.PointerTo(true),
							},
							Secrets: &models.SecretsConfig{
								Enabled: utils.PointerTo(true),
							},
							Sbom: &models.SBOMConfig{
								Enabled: utils.PointerTo(true),
							},
							Vulnerabilities: &models.VulnerabilitiesConfig{
								Enabled: utils.PointerTo(true),
							},
							Malware: &models.MalwareConfig{
								Enabled: utils.PointerTo(true),
							},
						},
						ScannerInstanceCreationConfig: nil,
					},
				},
			},
			want: &models.Estimation{
				Cost: utils.PointerTo(float32(0.5883259)),
				CostBreakdown: &[]models.CostBreakdownComponent{
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
				Size:     utils.PointerTo(8),
				Duration: utils.PointerTo(2304),
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
