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

package configwatcher

import (
	"testing"
	"time"

	"github.com/openclarity/vmclarity/api/models"
)

func Test_anyScansRunningOrCompleted(t *testing.T) {
	testScanConfigID := "testID"
	operationTime := time.Now()
	afterOperationTime := operationTime.Add(time.Minute * 5)
	beforeOperationTime := operationTime.Add(-time.Minute * 5)
	type args struct {
		scans         *models.Scans
		scanConfigID  string
		operationTime time.Time
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "scans.Items is null",
			args: args{
				scans:         &models.Scans{},
				scanConfigID:  testScanConfigID,
				operationTime: operationTime,
			},
			want: false,
		},
		{
			name: "scans.Items is empty list",
			args: args{
				scans: &models.Scans{
					Items: &[]models.Scan{},
				},
				scanConfigID:  testScanConfigID,
				operationTime: operationTime,
			},
			want: false,
		},
		{
			name: "there is a scans without end time",
			args: args{
				scans: &models.Scans{
					Items: &[]models.Scan{
						{
							ScanConfig: &models.ScanConfigRelationship{Id: testScanConfigID},
						},
					},
				},
				scanConfigID:  testScanConfigID,
				operationTime: operationTime,
			},
			want: true,
		},
		{
			name: "there is a scans with end time and start time after operation time",
			args: args{
				scans: &models.Scans{
					Items: &[]models.Scan{
						{
							ScanConfig: &models.ScanConfigRelationship{Id: testScanConfigID},
							StartTime:  &afterOperationTime,
							EndTime:    &operationTime,
						},
					},
				},
				scanConfigID:  testScanConfigID,
				operationTime: operationTime,
			},
			want: true,
		},
		{
			name: "there is a scans with end time and start time before operation time",
			args: args{
				scans: &models.Scans{
					Items: &[]models.Scan{
						{
							ScanConfig: &models.ScanConfigRelationship{Id: testScanConfigID},
							StartTime:  &beforeOperationTime,
							EndTime:    &operationTime,
						},
					},
				},
				scanConfigID:  testScanConfigID,
				operationTime: operationTime,
			},
			want: false,
		},
		{
			name: "there is a scans with end time and no start time",
			args: args{
				scans: &models.Scans{
					Items: &[]models.Scan{
						{
							ScanConfig: &models.ScanConfigRelationship{Id: testScanConfigID},
							EndTime:    &operationTime,
						},
					},
				},
				scanConfigID:  testScanConfigID,
				operationTime: operationTime,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := anyScansRunningOrCompleted(tt.args.scans, tt.args.operationTime); got != tt.want {
				t.Errorf("anyScansRunningOrCompleted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isWithinTheWindow(t *testing.T) {
	type args struct {
		operationTime time.Time
		now           time.Time
		window        time.Duration
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "in the window",
			args: args{
				operationTime: time.Now().Add(2 * time.Minute),
				now:           time.Now(),
				window:        3 * time.Minute,
			},
			want: true,
		},
		{
			name: "not in the window - before now",
			args: args{
				operationTime: time.Now().Add(-2 * time.Minute),
				now:           time.Now(),
				window:        3 * time.Minute,
			},
			want: false,
		},
		{
			name: "not in the window - after now",
			args: args{
				operationTime: time.Now().Add(2 * time.Minute),
				now:           time.Now(),
				window:        1 * time.Minute,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isWithinTheWindow(tt.args.operationTime, tt.args.now, tt.args.window); got != tt.want {
				t.Errorf("test() = %v, want %v", got, tt.want)
			}
		})
	}
}
