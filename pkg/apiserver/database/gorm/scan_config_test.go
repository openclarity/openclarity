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
	"testing"
	"time"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

func Test_isEmptyOperationTime(t *testing.T) {
	now := time.Now()
	type args struct {
		operationTime *time.Time
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil operationTime",
			args: args{
				operationTime: nil,
			},
			want: true,
		},
		{
			name: "zero operationTime",
			args: args{
				operationTime: &time.Time{},
			},
			want: true,
		},
		{
			name: "not empty operationTime",
			args: args{
				operationTime: &now,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEmptyOperationTime(tt.args.operationTime); got != tt.want {
				t.Errorf("isEmptyOperationTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateRuntimeScheduleScanConfig(t *testing.T) {
	type args struct {
		scheduled *models.RuntimeScheduleScanConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil scheduled",
			args: args{
				scheduled: nil,
			},
			wantErr: true,
		},
		{
			name: "both cron and operation time is missing",
			args: args{
				scheduled: &models.RuntimeScheduleScanConfig{
					CronLine:      nil,
					OperationTime: nil,
				},
			},
			wantErr: true,
		},
		{
			name: "both cron and operation time is missing, time is empty",
			args: args{
				scheduled: &models.RuntimeScheduleScanConfig{
					CronLine:      nil,
					OperationTime: &time.Time{},
				},
			},
			wantErr: true,
		},
		{
			name: "operation time is missing - not a valid cron expression",
			args: args{
				scheduled: &models.RuntimeScheduleScanConfig{
					CronLine: utils.PointerTo("not valid"),
				},
			},
			wantErr: true,
		},
		{
			name: "operation time is missing - operation time should be set",
			args: args{
				scheduled: &models.RuntimeScheduleScanConfig{
					CronLine: utils.PointerTo("0 */4 * * *"),
				},
			},
			wantErr: false,
		},
		{
			name: "cron line is missing - do nothing",
			args: args{
				scheduled: &models.RuntimeScheduleScanConfig{
					OperationTime: utils.PointerTo(time.Now()),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRuntimeScheduleScanConfig(tt.args.scheduled)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRuntimeScheduleScanConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && tt.args.scheduled.OperationTime == nil {
				t.Errorf("validateRuntimeScheduleScanConfig() operation time must be set after successful validation")
			}
		})
	}
}
