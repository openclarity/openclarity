package common

import (
	"testing"
	"time"

	kubeclarityUtils "github.com/openclarity/kubeclarity/shared/pkg/utils"
	"gotest.tools/v3/assert"

	"github.com/openclarity/vmclarity/api/models"
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
			gotDuration := GetScanDuration(tt.args.stats, tt.args.familiesConfig, tt.args.scanSizeMB)
			if gotDuration != tt.wantDuration {
				t.Errorf("getScanDuration() gotDuration = %v, want %v", gotDuration, tt.wantDuration)
			}
		})
	}
}
