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
	"testing"

	"github.com/openclarity/vmclarity/api/models"
)

func Test_validateScanConfigID(t *testing.T) {
	apiScan := models.Scan{
		ScanConfig: &models.ScanConfigRelationship{
			Id: "test",
		},
	}
	dbScan, err := json.Marshal(apiScan)
	if err != nil {
		t.Errorf("failed to marshal test scan data: %v", err)
	}

	type args struct {
		scan   models.Scan
		dbScan Scan
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "scan config ID not changed",
			args: args{
				scan: models.Scan{
					ScanConfig: &models.ScanConfigRelationship{
						Id: "test",
					},
				},
				dbScan: Scan{
					ODataObject{
						Data: dbScan,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "scan config ID is nil",
			args: args{
				scan: models.Scan{},
				dbScan: Scan{
					ODataObject{
						Data: dbScan,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "scan config ID changed",
			args: args{
				scan: models.Scan{
					ScanConfig: &models.ScanConfigRelationship{
						Id: "newID",
					},
				},
				dbScan: Scan{
					ODataObject{
						Data: dbScan,
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateScanConfigID(tt.args.scan, tt.args.dbScan); (err != nil) != tt.wantErr {
				t.Errorf("validateScanConfigID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
