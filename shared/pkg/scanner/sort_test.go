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

// nolint:maintidx
package scanner

import (
	"reflect"
	"testing"

	vulutil "github.com/openclarity/kubeclarity/shared/pkg/utils/vulnerability"
)

func Test_getCVSSBaseScore(t *testing.T) {
	type args struct {
		cvss    []CVSS
		version string
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "CVSSv3.1",
			args: args{
				cvss: []CVSS{
					{
						Version: "3.1",
						Metrics: CvssMetrics{
							BaseScore: 3.1,
						},
					},
					{
						Version: "3.0",
						Metrics: CvssMetrics{
							BaseScore: 3.0,
						},
					},
					{
						Version: "2.0",
						Metrics: CvssMetrics{
							BaseScore: 2.0,
						},
					},
				},
				version: "3.1",
			},
			want: 3.1,
		},
		{
			name: "CVSSv3.0",
			args: args{
				cvss: []CVSS{
					{
						Version: "3.1",
						Metrics: CvssMetrics{
							BaseScore: 3.1,
						},
					},
					{
						Version: "3.0",
						Metrics: CvssMetrics{
							BaseScore: 3.0,
						},
					},
					{
						Version: "2.0",
						Metrics: CvssMetrics{
							BaseScore: 2.0,
						},
					},
				},
				version: "3.0",
			},
			want: 3.0,
		},
		{
			name: "CVSSv2.0",
			args: args{
				cvss: []CVSS{
					{
						Version: "3.1",
						Metrics: CvssMetrics{
							BaseScore: 3.1,
						},
					},
					{
						Version: "3.0",
						Metrics: CvssMetrics{
							BaseScore: 3.0,
						},
					},
					{
						Version: "2.0",
						Metrics: CvssMetrics{
							BaseScore: 2.0,
						},
					},
				},
				version: "2.0",
			},
			want: 2.0,
		},
		{
			name: "missing CVSSv3.1",
			args: args{
				cvss: []CVSS{
					{
						Version: "3.0",
						Metrics: CvssMetrics{
							BaseScore: 3.0,
						},
					},
					{
						Version: "2.0",
						Metrics: CvssMetrics{
							BaseScore: 2.0,
						},
					},
				},
				version: "3.1",
			},
			want: 0,
		},
		{
			name: "missing CVSSv3.0",
			args: args{
				cvss: []CVSS{
					{
						Version: "3.1",
						Metrics: CvssMetrics{
							BaseScore: 3.1,
						},
					},
					{
						Version: "2.0",
						Metrics: CvssMetrics{
							BaseScore: 2.0,
						},
					},
				},
				version: "3.0",
			},
			want: 0,
		},
		{
			name: "missing CVSSv2.0",
			args: args{
				cvss: []CVSS{
					{
						Version: "3.1",
						Metrics: CvssMetrics{
							BaseScore: 3.1,
						},
					},
					{
						Version: "3.0",
						Metrics: CvssMetrics{
							BaseScore: 3.0,
						},
					},
				},
				version: "2.0",
			},
			want: 0,
		},
		{
			name: "empty CVSS slice",
			args: args{
				cvss:    nil,
				version: "2.0",
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getCVSSBaseScore(tt.args.cvss, tt.args.version); got != tt.want {
				t.Errorf("getCVSSBaseScore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortBySeverityAndCVSS(t *testing.T) {
	type args struct {
		vulnerabilities []MergedVulnerability
	}
	tests := []struct {
		name string
		args args
		want []MergedVulnerability
	}{
		{
			name: "sort by severity",
			args: args{
				vulnerabilities: []MergedVulnerability{
					{
						Vulnerability: Vulnerability{
							Severity: vulutil.HIGH,
						},
					},
					{
						Vulnerability: Vulnerability{
							Severity: vulutil.CRITICAL,
						},
					},
					{
						Vulnerability: Vulnerability{
							Severity: vulutil.LOW,
						},
					},
				},
			},
			want: []MergedVulnerability{
				{
					Vulnerability: Vulnerability{
						Severity: vulutil.CRITICAL,
					},
				},
				{
					Vulnerability: Vulnerability{
						Severity: vulutil.HIGH,
					},
				},
				{
					Vulnerability: Vulnerability{
						Severity: vulutil.LOW,
					},
				},
			},
		},
		{
			name: "sort by CVSSv3.1",
			args: args{
				vulnerabilities: []MergedVulnerability{
					{
						Vulnerability: Vulnerability{
							CVSS: []CVSS{
								{
									Version: "3.1",
									Metrics: CvssMetrics{
										BaseScore: 8,
									},
								},
							},
							Severity: vulutil.HIGH,
						},
					},
					{
						Vulnerability: Vulnerability{
							CVSS: []CVSS{
								{
									Version: "3.1",
									Metrics: CvssMetrics{
										BaseScore: 10,
									},
								},
							},
							Severity: vulutil.HIGH,
						},
					},
					{
						Vulnerability: Vulnerability{
							CVSS: []CVSS{
								{
									Version: "3.1",
									Metrics: CvssMetrics{
										BaseScore: 6,
									},
								},
							},
							Severity: vulutil.HIGH,
						},
					},
				},
			},
			want: []MergedVulnerability{
				{
					Vulnerability: Vulnerability{
						CVSS: []CVSS{
							{
								Version: "3.1",
								Metrics: CvssMetrics{
									BaseScore: 10,
								},
							},
						},
						Severity: vulutil.HIGH,
					},
				},
				{
					Vulnerability: Vulnerability{
						CVSS: []CVSS{
							{
								Version: "3.1",
								Metrics: CvssMetrics{
									BaseScore: 8,
								},
							},
						},
						Severity: vulutil.HIGH,
					},
				},
				{
					Vulnerability: Vulnerability{
						CVSS: []CVSS{
							{
								Version: "3.1",
								Metrics: CvssMetrics{
									BaseScore: 6,
								},
							},
						},
						Severity: vulutil.HIGH,
					},
				},
			},
		},
		{
			name: "sort by CVSSv3.0",
			args: args{
				vulnerabilities: []MergedVulnerability{
					{
						Vulnerability: Vulnerability{
							CVSS: []CVSS{
								{
									Version: "3.0",
									Metrics: CvssMetrics{
										BaseScore: 8,
									},
								},
							},
							Severity: vulutil.HIGH,
						},
					},
					{
						Vulnerability: Vulnerability{
							CVSS: []CVSS{
								{
									Version: "3.0",
									Metrics: CvssMetrics{
										BaseScore: 10,
									},
								},
							},
							Severity: vulutil.HIGH,
						},
					},
					{
						Vulnerability: Vulnerability{
							CVSS: []CVSS{
								{
									Version: "3.0",
									Metrics: CvssMetrics{
										BaseScore: 6,
									},
								},
							},
							Severity: vulutil.HIGH,
						},
					},
				},
			},
			want: []MergedVulnerability{
				{
					Vulnerability: Vulnerability{
						CVSS: []CVSS{
							{
								Version: "3.0",
								Metrics: CvssMetrics{
									BaseScore: 10,
								},
							},
						},
						Severity: vulutil.HIGH,
					},
				},
				{
					Vulnerability: Vulnerability{
						CVSS: []CVSS{
							{
								Version: "3.0",
								Metrics: CvssMetrics{
									BaseScore: 8,
								},
							},
						},
						Severity: vulutil.HIGH,
					},
				},
				{
					Vulnerability: Vulnerability{
						CVSS: []CVSS{
							{
								Version: "3.0",
								Metrics: CvssMetrics{
									BaseScore: 6,
								},
							},
						},
						Severity: vulutil.HIGH,
					},
				},
			},
		},
		{
			name: "sort by CVSSv2.0",
			args: args{
				vulnerabilities: []MergedVulnerability{
					{
						Vulnerability: Vulnerability{
							CVSS: []CVSS{
								{
									Version: "2.0",
									Metrics: CvssMetrics{
										BaseScore: 8,
									},
								},
							},
							Severity: vulutil.HIGH,
						},
					},
					{
						Vulnerability: Vulnerability{
							CVSS: []CVSS{
								{
									Version: "2.0",
									Metrics: CvssMetrics{
										BaseScore: 10,
									},
								},
							},
							Severity: vulutil.HIGH,
						},
					},
					{
						Vulnerability: Vulnerability{
							CVSS: []CVSS{
								{
									Version: "2.0",
									Metrics: CvssMetrics{
										BaseScore: 6,
									},
								},
							},
							Severity: vulutil.HIGH,
						},
					},
				},
			},
			want: []MergedVulnerability{
				{
					Vulnerability: Vulnerability{
						CVSS: []CVSS{
							{
								Version: "2.0",
								Metrics: CvssMetrics{
									BaseScore: 10,
								},
							},
						},
						Severity: vulutil.HIGH,
					},
				},
				{
					Vulnerability: Vulnerability{
						CVSS: []CVSS{
							{
								Version: "2.0",
								Metrics: CvssMetrics{
									BaseScore: 8,
								},
							},
						},
						Severity: vulutil.HIGH,
					},
				},
				{
					Vulnerability: Vulnerability{
						CVSS: []CVSS{
							{
								Version: "2.0",
								Metrics: CvssMetrics{
									BaseScore: 6,
								},
							},
						},
						Severity: vulutil.HIGH,
					},
				},
			},
		},
		{
			name: "mixed sort",
			args: args{
				vulnerabilities: []MergedVulnerability{
					{
						Vulnerability: Vulnerability{
							CVSS: []CVSS{
								{
									Version: "2.0",
									Metrics: CvssMetrics{
										BaseScore: 8,
									},
								},
							},
							Severity: vulutil.LOW,
						},
					},
					{
						Vulnerability: Vulnerability{
							CVSS: []CVSS{
								{
									Version: "3.0",
									Metrics: CvssMetrics{
										BaseScore: 10,
									},
								},
								{
									Version: "2.0",
									Metrics: CvssMetrics{
										BaseScore: 6,
									},
								},
							},
							Severity: vulutil.HIGH,
						},
					},
					{
						Vulnerability: Vulnerability{
							CVSS: []CVSS{
								{
									Version: "3.0",
									Metrics: CvssMetrics{
										BaseScore: 10,
									},
								},
								{
									Version: "2.0",
									Metrics: CvssMetrics{
										BaseScore: 7,
									},
								},
							},
							Severity: vulutil.HIGH,
						},
					},
					{
						Vulnerability: Vulnerability{
							CVSS: []CVSS{
								{
									Version: "3.1",
									Metrics: CvssMetrics{
										BaseScore: 5,
									},
								},
							},
							Severity: vulutil.CRITICAL,
						},
					},
					{
						Vulnerability: Vulnerability{
							CVSS: []CVSS{
								{
									Version: "3.1",
									Metrics: CvssMetrics{
										BaseScore: 6,
									},
								},
							},
							Severity: vulutil.CRITICAL,
						},
					},
				},
			},
			want: []MergedVulnerability{
				{
					Vulnerability: Vulnerability{
						CVSS: []CVSS{
							{
								Version: "3.1",
								Metrics: CvssMetrics{
									BaseScore: 6,
								},
							},
						},
						Severity: vulutil.CRITICAL,
					},
				},
				{
					Vulnerability: Vulnerability{
						CVSS: []CVSS{
							{
								Version: "3.1",
								Metrics: CvssMetrics{
									BaseScore: 5,
								},
							},
						},
						Severity: vulutil.CRITICAL,
					},
				},
				{
					Vulnerability: Vulnerability{
						CVSS: []CVSS{
							{
								Version: "3.0",
								Metrics: CvssMetrics{
									BaseScore: 10,
								},
							},
							{
								Version: "2.0",
								Metrics: CvssMetrics{
									BaseScore: 7,
								},
							},
						},
						Severity: vulutil.HIGH,
					},
				},
				{
					Vulnerability: Vulnerability{
						CVSS: []CVSS{
							{
								Version: "3.0",
								Metrics: CvssMetrics{
									BaseScore: 10,
								},
							},
							{
								Version: "2.0",
								Metrics: CvssMetrics{
									BaseScore: 6,
								},
							},
						},
						Severity: vulutil.HIGH,
					},
				},
				{
					Vulnerability: Vulnerability{
						CVSS: []CVSS{
							{
								Version: "2.0",
								Metrics: CvssMetrics{
									BaseScore: 8,
								},
							},
						},
						Severity: vulutil.LOW,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SortBySeverityAndCVSS(tt.args.vulnerabilities); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SortBySeverityAndCVSS() = %v, want %v", got, tt.want)
			}
		})
	}
}
