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

package gorm

import (
	"encoding/json"
	"reflect"
	"testing"

	uuid "github.com/satori/go.uuid"
	"gotest.tools/v3/assert"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
)

const id = "f12d1ca7-1048-4e31-899c-7a25b357bed1"

func TestConvertToDBScanResult(t *testing.T) {
	vulnerabilities := []models.Vulnerability{
		{
			VulnerabilityInfo: &models.VulnerabilityInfo{
				Description:       utils.StringPtr("desc"),
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

	idPtr := id

	type args struct {
		result models.TargetScanResult
	}
	tests := []struct {
		name    string
		args    args
		want    ScanResult
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				result: models.TargetScanResult{
					Id:              &idPtr,
					ScanId:          "3",
					TargetId:        "2",
					Vulnerabilities: &vulScan,
				},
			},
			want: ScanResult{
				Base:            Base{ID: UUID},
				ScanID:          "3",
				TargetID:        "2",
				Vulnerabilities: vulScanB,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertToDBScanResult(tt.args.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToDBScanResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertToDBScanResult() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToDBTarget(t *testing.T) {
	cloudProvider := models.CloudProvider("aws")
	vmInfo := models.VMInfo{
		InstanceID:       "instanceID",
		InstanceProvider: &cloudProvider,
		Location:         "location",
	}

	var targetType models.TargetType

	err := targetType.FromVMInfo(vmInfo)
	assert.NilError(t, err)

	UUID, err := uuid.FromString(id)
	assert.NilError(t, err)

	idPtr := id

	type args struct {
		target models.Target
	}
	tests := []struct {
		name    string
		args    args
		want    Target
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				target: models.Target{
					Id:         &idPtr,
					TargetInfo: &targetType,
				},
			},
			want: Target{
				Base:             Base{ID: UUID},
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
			got, err := ConvertToDBTarget(tt.args.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToDBTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertToDBTarget() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToDBScan(t *testing.T) {
	scanFamiliesConfig := models.ScanConfigData{
		ScanFamiliesConfig: &models.ScanFamiliesConfig{
			Exploits: &models.ExploitsConfig{
				Enabled: utils.BoolPtr(true),
			},
		},
	}

	scanFamiliesConfigB, err := json.Marshal(scanFamiliesConfig)
	assert.NilError(t, err)

	targetIDs := []string{"s1"}
	targetIDsB, err := json.Marshal(&targetIDs)
	assert.NilError(t, err)

	UUID, err := uuid.FromString(id)
	assert.NilError(t, err)

	idPtr := id

	type args struct {
		scan models.Scan
	}
	tests := []struct {
		name    string
		args    args
		want    Scan
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				scan: models.Scan{
					Id: &idPtr,
					ScanConfig: &models.ScanConfigRelationship{
						Id: "1",
					},
					ScanConfigSnapshot: &scanFamiliesConfig,
					TargetIDs:          &targetIDs,
				},
			},
			want: Scan{
				Base:               Base{ID: UUID},
				ScanConfigID:       utils.StringPtr("1"),
				ScanConfigSnapshot: scanFamiliesConfigB,
				TargetIDs:          targetIDsB,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertToDBScan(tt.args.scan)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToDBScan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertToDBScan() got = %v, want %v", got, tt.want)
			}
		})
	}
}
