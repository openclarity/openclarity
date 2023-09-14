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

package scanestimationwatcher

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

func Test_updateScanEstimationSummaryFromAssetScanEstimation(t *testing.T) {
	type args struct {
		scanEstimation *models.ScanEstimation
		result         models.AssetScanEstimation
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				scanEstimation: &models.ScanEstimation{
					Summary: &models.ScanEstimationSummary{
						JobsCompleted: utils.PointerTo(0),
						JobsLeftToRun: utils.PointerTo(0),
					},
				},
				result: models.AssetScanEstimation{
					State: &models.AssetScanEstimationState{
						State: utils.PointerTo(models.AssetScanEstimationStateStateFailed),
					},
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
		scanEstimation *models.ScanEstimation
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
				scanEstimation: &models.ScanEstimation{
					Summary: &models.ScanEstimationSummary{
						JobsCompleted: utils.PointerTo(5),
						TotalScanTime: utils.PointerTo(30),
					},
					ScanTemplate: &models.ScanTemplate{},
				},
			},
			wantTotalScanTime: 30,
			wantErr:           false,
		},
		{
			name: "max parallel scanners == 1",
			args: args{
				scanEstimation: &models.ScanEstimation{
					Summary: &models.ScanEstimationSummary{
						JobsCompleted: utils.PointerTo(5),
						TotalScanTime: utils.PointerTo(30),
					},
					ScanTemplate: &models.ScanTemplate{
						MaxParallelScanners: utils.PointerTo(1),
					},
				},
			},
			wantTotalScanTime: 30,
			wantErr:           false,
		},
		{
			name: "max parallel scanners == number of jobs",
			args: args{
				scanEstimation: &models.ScanEstimation{
					Summary: &models.ScanEstimationSummary{
						JobsCompleted: utils.PointerTo(5),
						TotalScanTime: utils.PointerTo(30),
					},
					ScanTemplate: &models.ScanTemplate{
						MaxParallelScanners: utils.PointerTo(5),
					},
				},
			},
			wantTotalScanTime: 6,
			wantErr:           false,
		},
		{
			name: "max parallel scanners < number of jobs",
			args: args{
				scanEstimation: &models.ScanEstimation{
					Summary: &models.ScanEstimationSummary{
						JobsCompleted: utils.PointerTo(3),
						TotalScanTime: utils.PointerTo(30),
					},
					ScanTemplate: &models.ScanTemplate{
						MaxParallelScanners: utils.PointerTo(2),
					},
				},
			},
			wantTotalScanTime: 15,
			wantErr:           false,
		},
		{
			name: "max parallel scanners > number of jobs",
			args: args{
				scanEstimation: &models.ScanEstimation{
					Summary: &models.ScanEstimationSummary{
						JobsCompleted: utils.PointerTo(2),
						TotalScanTime: utils.PointerTo(30),
					},
					ScanTemplate: &models.ScanTemplate{
						MaxParallelScanners: utils.PointerTo(3),
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
