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

func TestConvertToRestScan(t *testing.T) {
	scanSnap := models.ScanConfigData{
		ScanFamiliesConfig: &models.ScanFamiliesConfig{
			Vulnerabilities: &models.VulnerabilitiesConfig{
				Enabled: nil,
			},
		},
	}

	scanSnapB, err := json.Marshal(&scanSnap)
	assert.NilError(t, err)

	targetIDs := []string{"s"}
	targetIDsB, err := json.Marshal(&targetIDs)
	assert.NilError(t, err)

	id := uuid.NewV4()

	type args struct {
		scan Scan
	}
	tests := []struct {
		name    string
		args    args
		want    models.Scan
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				scan: Scan{
					Base: Base{
						ID: id,
					},
					ScanStartTime:      nil,
					ScanEndTime:        nil,
					ScanConfigID:       utils.StringPtr("1"),
					ScanConfigSnapshot: scanSnapB,
					State:              "test-state",
					StateMessage:       "test-message",
					StateReason:        "test-reason",
					Summary:            nil,
					TargetIDs:          targetIDsB,
				},
			},
			want: models.Scan{
				Id: utils.StringPtr(id.String()),
				ScanConfig: &models.ScanConfigRelationship{
					Id: "1",
				},
				ScanConfigSnapshot: &scanSnap,
				State:              utils.PointerTo[models.ScanState]("test-state"),
				StateMessage:       utils.PointerTo("test-message"),
				StateReason:        utils.PointerTo[models.ScanStateReason]("test-reason"),
				Summary:            nil,
				TargetIDs:          &targetIDs,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertToRestScan(tt.args.scan)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToRestScan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				gotB, _ := json.Marshal(got)
				wantB, _ := json.Marshal(tt.want)
				t.Errorf("ConvertToRestScan() got = %s, want %s", gotB, wantB)
			}
		})
	}
}
