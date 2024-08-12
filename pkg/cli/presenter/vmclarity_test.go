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

package presenter

import (
	"reflect"
	"testing"
	"time"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/families/types"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

func Test_getInputScanStats(t *testing.T) {
	timeNow := time.Now()
	startTime := timeNow.Add(-5 * time.Second)
	startTime2 := timeNow.Add(-10 * time.Second)
	type args struct {
		inputScans []types.InputScanMetadata
	}
	tests := []struct {
		name string
		args args
		want *[]models.AssetScanInputScanStats
	}{
		{
			name: "no input scans",
			args: args{
				inputScans: []types.InputScanMetadata{},
			},
			want: nil,
		},
		{
			name: "one input scans",
			args: args{
				inputScans: []types.InputScanMetadata{
					{
						InputType:     "rootfs",
						InputPath:     "/mnt/snap",
						InputSize:     450,
						ScanStartTime: startTime,
						ScanEndTime:   timeNow,
					},
				},
			},
			want: &[]models.AssetScanInputScanStats{
				{
					Path: utils.PointerTo("/mnt/snap"),
					ScanTime: &models.AssetScanScanTime{
						EndTime:   &timeNow,
						StartTime: &startTime,
					},
					Size: utils.PointerTo(int64(450)),
					Type: utils.PointerTo("rootfs"),
				},
			},
		},
		{
			name: "two input scans",
			args: args{
				inputScans: []types.InputScanMetadata{
					{
						InputType:     "rootfs",
						InputPath:     "/mnt/snap",
						InputSize:     450,
						ScanStartTime: startTime,
						ScanEndTime:   timeNow,
					},
					{
						InputType:     "dir",
						InputPath:     "/mnt/snap2",
						InputSize:     30,
						ScanStartTime: startTime2,
						ScanEndTime:   timeNow,
					},
				},
			},
			want: &[]models.AssetScanInputScanStats{
				{
					Path: utils.PointerTo("/mnt/snap"),
					ScanTime: &models.AssetScanScanTime{
						EndTime:   &timeNow,
						StartTime: &startTime,
					},
					Size: utils.PointerTo(int64(450)),
					Type: utils.PointerTo("rootfs"),
				},
				{
					Path: utils.PointerTo("/mnt/snap2"),
					ScanTime: &models.AssetScanScanTime{
						EndTime:   &timeNow,
						StartTime: &startTime2,
					},
					Size: utils.PointerTo(int64(30)),
					Type: utils.PointerTo("dir"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getInputScanStats(tt.args.inputScans); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getInputScanStats() = %v, want %v", got, tt.want)
			}
		})
	}
}
