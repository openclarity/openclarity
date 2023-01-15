// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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

package dbtorest

import (
	"encoding/json"
	"reflect"
	"testing"

	uuid "github.com/satori/go.uuid"
	"gotest.tools/v3/assert"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/backend/pkg/database"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
)

func TestConvertScanConfig(t *testing.T) {
	scanFamiliesConfig := models.ScanFamiliesConfig{
		Vulnerabilities: &models.VulnerabilitiesConfig{Enabled: utils.BoolPtr(true)},
	}

	scanFamiliesConfigB, err := json.Marshal(&scanFamiliesConfig)
	assert.NilError(t, err)

	awsScanScope := models.AwsScanScope{
		All:                        utils.BoolPtr(true),
		InstanceTagExclusion:       nil,
		InstanceTagSelector:        nil,
		ObjectType:                 "AwsScanScope",
		Regions:                    nil,
		ShouldScanStoppedInstances: utils.BoolPtr(false),
	}

	var scanScopeType models.ScanScopeType

	err = scanScopeType.FromAwsScanScope(awsScanScope)
	assert.NilError(t, err)

	scanScopeTypeB, err := scanScopeType.MarshalJSON()
	assert.NilError(t, err)

	byHoursScheduleScanConfig := models.ByHoursScheduleScanConfig{
		HoursInterval: utils.IntPtr(2),
		ObjectType:    "ByHoursScheduleScanConfig",
	}

	var runtimeScheduleScanConfigType models.RuntimeScheduleScanConfigType
	err = runtimeScheduleScanConfigType.FromByHoursScheduleScanConfig(byHoursScheduleScanConfig)
	assert.NilError(t, err)

	runtimeScheduleScanConfigTypeB, err := runtimeScheduleScanConfigType.MarshalJSON()
	assert.NilError(t, err)

	uid := uuid.NewV4()

	type args struct {
		config *database.ScanConfig
	}
	tests := []struct {
		name    string
		args    args
		want    *models.ScanConfig
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				config: &database.ScanConfig{
					Base: database.Base{
						ID: uid,
					},
					Name:               utils.StringPtr("test"),
					ScanFamiliesConfig: scanFamiliesConfigB,
					Scheduled:          runtimeScheduleScanConfigTypeB,
					Scope:              scanScopeTypeB,
				},
			},
			want: &models.ScanConfig{
				Id:                 utils.StringPtr(uid.String()),
				Name:               utils.StringPtr("test"),
				ScanFamiliesConfig: &scanFamiliesConfig,
				Scheduled:          &runtimeScheduleScanConfigType,
				Scope:              &scanScopeType,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertScanConfig(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertScanConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertScanConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertScanConfigs(t *testing.T) {
	scanFamiliesConfig := models.ScanFamiliesConfig{
		Vulnerabilities: &models.VulnerabilitiesConfig{Enabled: utils.BoolPtr(true)},
	}

	scanFamiliesConfigB, err := json.Marshal(&scanFamiliesConfig)
	assert.NilError(t, err)

	awsScanScope := models.AwsScanScope{
		All:                        utils.BoolPtr(true),
		InstanceTagExclusion:       nil,
		InstanceTagSelector:        nil,
		ObjectType:                 "AwsScanScope",
		Regions:                    nil,
		ShouldScanStoppedInstances: utils.BoolPtr(false),
	}

	var scanScopeType models.ScanScopeType

	err = scanScopeType.FromAwsScanScope(awsScanScope)
	assert.NilError(t, err)

	scanScopeTypeB, err := scanScopeType.MarshalJSON()
	assert.NilError(t, err)

	byHoursScheduleScanConfig := models.ByHoursScheduleScanConfig{
		HoursInterval: utils.IntPtr(2),
		ObjectType:    "ByHoursScheduleScanConfig",
	}

	var runtimeScheduleScanConfigType models.RuntimeScheduleScanConfigType
	err = runtimeScheduleScanConfigType.FromByHoursScheduleScanConfig(byHoursScheduleScanConfig)
	assert.NilError(t, err)

	runtimeScheduleScanConfigTypeB, err := runtimeScheduleScanConfigType.MarshalJSON()
	assert.NilError(t, err)

	uid := uuid.NewV4()

	type args struct {
		configs []*database.ScanConfig
		total   int64
	}
	tests := []struct {
		name    string
		args    args
		want    *models.ScanConfigs
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				configs: []*database.ScanConfig{
					{
						Base: database.Base{
							ID: uid,
						},
						Name:               utils.StringPtr("test"),
						ScanFamiliesConfig: scanFamiliesConfigB,
						Scheduled:          runtimeScheduleScanConfigTypeB,
						Scope:              scanScopeTypeB,
					},
				},
				total: 1,
			},
			want: &models.ScanConfigs{
				Items: &[]models.ScanConfig{
					{
						Id:                 utils.StringPtr(uid.String()),
						Name:               utils.StringPtr("test"),
						ScanFamiliesConfig: &scanFamiliesConfig,
						Scheduled:          &runtimeScheduleScanConfigType,
						Scope:              &scanScopeType,
					},
				},
				Total: 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertScanConfigs(tt.args.configs, tt.args.total)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertScanConfigs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertScanConfigs() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertScanResult(t *testing.T) {
	state := models.DONE
	status := models.TargetScanStatus{
		Vulnerabilities: &models.TargetScanState{
			Errors: nil,
			State:  &state,
		},
		Exploits: &models.TargetScanState{
			Errors: &[]string{"err"},
			State:  &state,
		},
	}

	vulsScan := models.VulnerabilityScan{
		Vulnerabilities: &[]models.Vulnerability{
			{
				VulnerabilityInfo: &models.VulnerabilityInfo{
					VulnerabilityName: utils.StringPtr("name"),
				},
			},
		},
	}

	vulScanB, err := json.Marshal(&vulsScan)
	assert.NilError(t, err)

	statusB, err := json.Marshal(&status)
	assert.NilError(t, err)

	uid := uuid.NewV4()

	type args struct {
		scanResult *database.ScanResult
	}
	tests := []struct {
		name    string
		args    args
		want    *models.TargetScanResult
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				scanResult: &database.ScanResult{
					Base: database.Base{
						ID: uid,
					},
					ScanID:          "1",
					TargetID:        "2",
					Status:          statusB,
					Vulnerabilities: vulScanB,
				},
			},
			want: &models.TargetScanResult{
				Id:              utils.StringPtr(uid.String()),
				ScanId:          "1",
				Status:          &status,
				TargetId:        "2",
				Vulnerabilities: &vulsScan,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertScanResult(tt.args.scanResult)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertScanResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertScanResult() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertScan(t *testing.T) {
	scanFam := models.ScanFamiliesConfig{
		Vulnerabilities: &models.VulnerabilitiesConfig{
			Enabled: nil,
		},
	}

	scanFamB, err := json.Marshal(&scanFam)
	assert.NilError(t, err)

	targetIDs := []string{"s"}
	targetIDsB, err := json.Marshal(&targetIDs)
	assert.NilError(t, err)

	id := uuid.NewV4()

	type args struct {
		scan *database.Scan
	}
	tests := []struct {
		name    string
		args    args
		want    *models.Scan
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				scan: &database.Scan{
					Base: database.Base{
						ID: id,
					},
					ScanStartTime:      nil,
					ScanEndTime:        nil,
					ScanConfigID:       utils.StringPtr("1"),
					ScanFamiliesConfig: scanFamB,
					TargetIDs:          targetIDsB,
				},
			},
			want: &models.Scan{
				Id:                 utils.StringPtr(id.String()),
				ScanConfigId:       utils.StringPtr("1"),
				ScanFamiliesConfig: &scanFam,
				TargetIDs:          &targetIDs,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertScan(tt.args.scan)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertScan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertScan() got = %v, want %v", got, tt.want)
			}
		})
	}
}
