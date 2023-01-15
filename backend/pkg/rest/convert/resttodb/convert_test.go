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

package resttodb

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

const id = "f12d1ca7-1048-4e31-899c-7a25b357bed1"

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

	UUID, err := uuid.FromString(id)
	assert.NilError(t, err)

	type args struct {
		config *models.ScanConfig
		id     string
	}
	tests := []struct {
		name    string
		args    args
		want    *database.ScanConfig
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				id: id,
				config: &models.ScanConfig{
					Id:                 utils.StringPtr("1"),
					Name:               utils.StringPtr("scanConfigName"),
					ScanFamiliesConfig: &scanFamiliesConfig,
					Scheduled:          &runtimeScheduleScanConfigType,
					Scope:              &scanScopeType,
				},
			},
			want: &database.ScanConfig{
				Base:               database.Base{ID: UUID},
				Name:               utils.StringPtr("scanConfigName"),
				ScanFamiliesConfig: scanFamiliesConfigB,
				Scheduled:          runtimeScheduleScanConfigTypeB,
				Scope:              scanScopeTypeB,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertScanConfig(tt.args.config, tt.args.id)
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

func TestConvertScanResult(t *testing.T) {
	vulnerabilities := []models.Vulnerability{
		{
			VulnerabilityInfo: &models.VulnerabilityInfo{
				Description:       utils.StringPtr("desc"),
				Id:                utils.StringPtr("1"),
				VulnerabilityName: utils.StringPtr("name"),
			},
			Id: utils.StringPtr("1"),
		},
	}
	vulScan := models.VulnerabilityScan{Vulnerabilities: &vulnerabilities}

	vulScanB, err := json.Marshal(vulScan)
	assert.NilError(t, err)

	UUID, err := uuid.FromString(id)
	assert.NilError(t, err)

	type args struct {
		result *models.TargetScanResult
		id     string
	}
	tests := []struct {
		name    string
		args    args
		want    *database.ScanResult
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				id: id,
				result: &models.TargetScanResult{
					Id:              nil,
					ScanId:          "3",
					TargetId:        "2",
					Vulnerabilities: &vulScan,
				},
			},
			want: &database.ScanResult{
				Base:            database.Base{ID: UUID},
				ScanID:          "3",
				TargetID:        "2",
				Vulnerabilities: vulScanB,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertScanResult(tt.args.result, tt.args.id)
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

func TestConvertTarget(t *testing.T) {
	cloudProvider := models.CloudProvider("aws")
	vmInfo := models.VMInfo{
		InstanceID:       utils.StringPtr("instanceID"),
		InstanceProvider: &cloudProvider,
		Location:         utils.StringPtr("location"),
	}

	var targetType models.TargetType

	err := targetType.FromVMInfo(vmInfo)
	assert.NilError(t, err)

	UUID, err := uuid.FromString(id)
	assert.NilError(t, err)

	type args struct {
		target *models.Target
		id     string
	}
	tests := []struct {
		name    string
		args    args
		want    *database.Target
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				target: &models.Target{
					TargetInfo: &targetType,
				},
				id: id,
			},
			want: &database.Target{
				Base:             database.Base{ID: UUID},
				Type:             "VMInfo",
				Location:         utils.StringPtr("location"),
				InstanceID:       utils.StringPtr("instanceID"),
				InstanceProvider: utils.StringPtr("aws"),
				PodName:          nil,
				DirName:          nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertTarget(tt.args.target, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertTarget() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertScan(t *testing.T) {
	scanFamiliesConfig := models.ScanFamiliesConfig{
		Exploits: &models.ExploitsConfig{
			Enabled: utils.BoolPtr(true),
		},
	}

	scanFamiliesConfigB, err := json.Marshal(scanFamiliesConfig)
	assert.NilError(t, err)

	targetIDs := []string{"s1"}
	targetIDsB, err := json.Marshal(&targetIDs)
	assert.NilError(t, err)

	UUID, err := uuid.FromString(id)
	assert.NilError(t, err)

	type args struct {
		scan *models.Scan
		id   string
	}
	tests := []struct {
		name    string
		args    args
		want    *database.Scan
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				id: id,
				scan: &models.Scan{
					ScanConfigId:       utils.StringPtr("1"),
					ScanFamiliesConfig: &scanFamiliesConfig,
					TargetIDs:          &targetIDs,
				},
			},
			want: &database.Scan{
				Base:               database.Base{ID: UUID},
				ScanConfigID:       utils.StringPtr("1"),
				ScanFamiliesConfig: scanFamiliesConfigB,
				TargetIDs:          targetIDsB,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertScan(tt.args.scan, tt.args.id)
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
