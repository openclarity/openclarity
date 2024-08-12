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

package scanestimation

import (
	"testing"

	"gotest.tools/v3/assert"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
)

func Test_updateScanEstimationSummaryFromAssetScanEstimation(t *testing.T) {
	type args struct {
		scanEstimation *apitypes.ScanEstimation
		result         apitypes.AssetScanEstimation
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				scanEstimation: &apitypes.ScanEstimation{
					Summary: &apitypes.ScanEstimationSummary{
						JobsCompleted: to.Ptr(0),
						JobsLeftToRun: to.Ptr(0),
					},
				},
				result: apitypes.AssetScanEstimation{
					Status: apitypes.NewAssetScanEstimationStatus(
						apitypes.AssetScanEstimationStatusStateFailed,
						apitypes.AssetScanEstimationStatusReasonError,
						nil,
					),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := updateScanEstimationSummaryFromAssetScanEstimation(tt.args.scanEstimation, tt.args.result); (err != nil) != tt.wantErr {
				t.Errorf("updateScanEstimationSummaryFromAssetScanEstimation() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Assert(t, *tt.args.scanEstimation.Summary.JobsLeftToRun == 0)
			assert.Assert(t, *tt.args.scanEstimation.Summary.JobsCompleted == 1)
		})
	}
}

func Test_updateTotalScanTimeWithParallelScans(t *testing.T) {
	type args struct {
		scanEstimation *apitypes.ScanEstimation
	}
	tests := []struct {
		name              string
		args              args
		wantErr           bool
		wantTotalScanTime int
	}{
		{
			name: "max parallel scanners == nil",
			args: args{
				scanEstimation: &apitypes.ScanEstimation{
					Summary: &apitypes.ScanEstimationSummary{
						JobsCompleted: to.Ptr(5),
						TotalScanTime: to.Ptr(30),
					},
					ScanTemplate: &apitypes.ScanTemplate{},
				},
			},
			wantTotalScanTime: 30,
			wantErr:           false,
		},
		{
			name: "max parallel scanners == 1",
			args: args{
				scanEstimation: &apitypes.ScanEstimation{
					Summary: &apitypes.ScanEstimationSummary{
						JobsCompleted: to.Ptr(5),
						TotalScanTime: to.Ptr(30),
					},
					ScanTemplate: &apitypes.ScanTemplate{
						MaxParallelScanners: to.Ptr(1),
					},
				},
			},
			wantTotalScanTime: 30,
			wantErr:           false,
		},
		{
			name: "max parallel scanners == number of jobs",
			args: args{
				scanEstimation: &apitypes.ScanEstimation{
					Summary: &apitypes.ScanEstimationSummary{
						JobsCompleted: to.Ptr(5),
						TotalScanTime: to.Ptr(30),
					},
					ScanTemplate: &apitypes.ScanTemplate{
						MaxParallelScanners: to.Ptr(5),
					},
				},
			},
			wantTotalScanTime: 6,
			wantErr:           false,
		},
		{
			name: "max parallel scanners < number of jobs",
			args: args{
				scanEstimation: &apitypes.ScanEstimation{
					Summary: &apitypes.ScanEstimationSummary{
						JobsCompleted: to.Ptr(3),
						TotalScanTime: to.Ptr(30),
					},
					ScanTemplate: &apitypes.ScanTemplate{
						MaxParallelScanners: to.Ptr(2),
					},
				},
			},
			wantTotalScanTime: 15,
			wantErr:           false,
		},
		{
			name: "max parallel scanners > number of jobs",
			args: args{
				scanEstimation: &apitypes.ScanEstimation{
					Summary: &apitypes.ScanEstimationSummary{
						JobsCompleted: to.Ptr(2),
						TotalScanTime: to.Ptr(30),
					},
					ScanTemplate: &apitypes.ScanTemplate{
						MaxParallelScanners: to.Ptr(3),
					},
				},
			},
			wantTotalScanTime: 15,
			wantErr:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := updateTotalScanTimeWithParallelScans(tt.args.scanEstimation); (err != nil) != tt.wantErr {
				t.Errorf("updateTotalScanTimeWithParallelScans() error = %v, wantErr %v", err, tt.wantErr)
			}
			if *tt.args.scanEstimation.Summary.TotalScanTime != tt.wantTotalScanTime {
				t.Errorf("updateTotalScanTimeWithParallelScans() failed. wantTotalScanTime = %v, gotTotalScanTime = %v", tt.wantTotalScanTime, *tt.args.scanEstimation.Summary.TotalScanTime)
			}
		})
	}
}
