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

package types

import (
	"testing"

	"gotest.tools/assert"

	"github.com/cisco-open/kubei/api/server/models"
	runtime_scan_models "github.com/cisco-open/kubei/runtime_scan/api/server/models"
)

func TestCISDockerBenchmarkResultsFromBackendAPI(t *testing.T) {
	type args struct {
		in []*models.CISDockerBenchmarkCodeInfo
	}
	tests := []struct {
		name string
		args args
		want []*CISDockerBenchmarkResult
	}{
		{
			name: "sanity",
			args: args{
				in: []*models.CISDockerBenchmarkCodeInfo{
					{
						Assessments: []*models.CISDockerBenchmarkAssessment{
							{
								Code:     "Code1",
								Desc:     "Desc1",
								Filename: "Filename1",
								Level:    1,
							},
							{
								Code:     "Code1",
								Desc:     "Desc2",
								Filename: "Filename2",
								Level:    1,
							},
						},
						Code:  "Code1",
						Level: 1,
					},
					{
						Assessments: []*models.CISDockerBenchmarkAssessment{
							{
								Code:     "Code2",
								Desc:     "Desc22",
								Filename: "Filename22",
								Level:    2,
							},
						},
						Code:  "Code2",
						Level: 2,
					},
					// Empty description
					{
						Assessments: []*models.CISDockerBenchmarkAssessment{
							{
								Code:     "Code3",
								Filename: "Filename33",
								Level:    3,
							},
						},
						Code:  "Code3",
						Level: 3,
					},
				},
			},
			want: []*CISDockerBenchmarkResult{
				{
					Code:         "Code1",
					Level:        1,
					Descriptions: "Desc1, Desc2",
				},
				{
					Code:         "Code2",
					Level:        2,
					Descriptions: "Desc22",
				},
				{
					Code:  "Code3",
					Level: 3,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CISDockerBenchmarkResultsFromBackendAPI(tt.args.in)
			assert.DeepEqual(t, tt.want, got)
		})
	}
}

func TestCISDockerBenchmarkResultsFromFromRuntimeScan(t *testing.T) {
	type args struct {
		in []*runtime_scan_models.CISDockerBenchmarkCodeInfo
	}
	tests := []struct {
		name string
		args args
		want []*CISDockerBenchmarkResult
	}{
		{
			name: "sanity",
			args: args{
				in: []*runtime_scan_models.CISDockerBenchmarkCodeInfo{
					{
						Assessments: []*runtime_scan_models.CISDockerBenchmarkAssessment{
							{
								Code:     "Code1",
								Desc:     "Desc1",
								Filename: "Filename1",
								Level:    1,
							},
							{
								Code:     "Code1",
								Desc:     "Desc2",
								Filename: "Filename2",
								Level:    1,
							},
						},
						Code:  "Code1",
						Level: 1,
					},
					{
						Assessments: []*runtime_scan_models.CISDockerBenchmarkAssessment{
							{
								Code:     "Code2",
								Desc:     "Desc22",
								Filename: "Filename22",
								Level:    2,
							},
						},
						Code:  "Code2",
						Level: 2,
					},
					// Empty description
					{
						Assessments: []*runtime_scan_models.CISDockerBenchmarkAssessment{
							{
								Code:     "Code3",
								Filename: "Filename33",
								Level:    3,
							},
						},
						Code:  "Code3",
						Level: 3,
					},
				},
			},
			want: []*CISDockerBenchmarkResult{
				{
					Code:         "Code1",
					Level:        1,
					Descriptions: "Desc1, Desc2",
				},
				{
					Code:         "Code2",
					Level:        2,
					Descriptions: "Desc22",
				},
				{
					Code:  "Code3",
					Level: 3,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CISDockerBenchmarkResultsFromFromRuntimeScan(tt.args.in)
			assert.DeepEqual(t, tt.want, got)
		})
	}
}
