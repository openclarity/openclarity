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

package common

import (
	"testing"
	"time"

	kubeclarityUtils "github.com/openclarity/kubeclarity/shared/pkg/utils"
	"gotest.tools/v3/assert"

	"github.com/openclarity/vmclarity/api/models"
	familiestypes "github.com/openclarity/vmclarity/pkg/shared/families/types"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

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
			got, err := GetScanSize(tt.args.stats, tt.args.asset)
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

var scanSizesGB = []float64{0.01, 5, 10}

var fakeFamilyScanDurationsMap = map[familiestypes.FamilyType]*LogarithmicFormula{
	familiestypes.SBOM:             MustLogarithmicFit(scanSizesGB, []float64{0.01, 5, 10}),
	familiestypes.Vulnerabilities:  MustLogarithmicFit(scanSizesGB, []float64{0.01, 10, 20}),
	familiestypes.Secrets:          MustLogarithmicFit(scanSizesGB, []float64{0.01, 3, 6}),
	familiestypes.Exploits:         MustLogarithmicFit(scanSizesGB, []float64{0, 0, 0}),
	familiestypes.Rootkits:         MustLogarithmicFit(scanSizesGB, []float64{0, 0, 0}),
	familiestypes.Misconfiguration: MustLogarithmicFit(scanSizesGB, []float64{0.01, 20, 40}),
	familiestypes.Malware:          MustLogarithmicFit(scanSizesGB, []float64{0.01, 500, 1000}),
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
			wantDuration: 1083,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDuration := GetScanDuration(tt.args.stats, tt.args.familiesConfig, tt.args.scanSizeMB, fakeFamilyScanDurationsMap)
			if gotDuration != tt.wantDuration {
				t.Errorf("getScanDuration() gotDuration = %v, want %v", gotDuration, tt.wantDuration)
			}
		})
	}
}
