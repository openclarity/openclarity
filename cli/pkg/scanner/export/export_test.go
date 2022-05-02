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

package export

import (
	"reflect"
	"testing"

	dockle_types "github.com/Portshift/dockle/pkg/types"
	"github.com/spiegel-im-spiegel/go-cvss/v3/metric"

	"github.com/openclarity/kubeclarity/api/client/models"
	"github.com/openclarity/kubeclarity/shared/pkg/scanner"
)

func Test_getScannerInfo(t *testing.T) {
	type args struct {
		mergedVulnerability scanner.MergedVulnerability
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "get scaner info",
			args: args{
				mergedVulnerability: scanner.MergedVulnerability{
					ScannersInfo: []scanner.Info{
						{
							Name: "grype",
						},
						{
							Name: "dependency-track",
						},
					},
				},
			},
			want: []string{"grype", "dependency-track"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getScannerInfo(tt.args.mergedVulnerability); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getScannerInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getPackageInfo(t *testing.T) {
	type args struct {
		vulnerability scanner.Vulnerability
	}
	tests := []struct {
		name string
		args args
		want *models.PackageInfo
	}{
		{
			name: "licenses are not set",
			args: args{
				vulnerability: scanner.Vulnerability{
					Package: scanner.Package{
						Name:     "test",
						Version:  "1.0.0",
						Language: "golang",
					},
				},
			},
			want: &models.PackageInfo{
				Name:     "test",
				Version:  "1.0.0",
				Language: "golang",
			},
		},
		{
			name: "licenses are set",
			args: args{
				vulnerability: scanner.Vulnerability{
					Package: scanner.Package{
						Name:     "test",
						Version:  "1.0.0",
						Language: "golang",
						Licenses: []string{"MIT"},
					},
				},
			},
			want: &models.PackageInfo{
				Name:     "test",
				Version:  "1.0.0",
				Language: "golang",
				License:  "MIT",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPackageInfo(tt.args.vulnerability); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPackageInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getCVSS(t *testing.T) {
	score := float64(1111)
	type args struct {
		vulnerability scanner.Vulnerability
	}
	tests := []struct {
		name string
		args args
		want *models.CVSS
	}{
		{
			name: "CVSS not found for vulnerability",
			args: args{
				vulnerability: scanner.Vulnerability{
					ID: "fake",
				},
			},
			want: nil,
		},
		{
			name: "CVSS version 3 not found for vulnerability",
			args: args{
				vulnerability: scanner.Vulnerability{
					ID: "fake",
					CVSS: []scanner.CVSS{
						{
							Version: "2.0",
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "CVSS version 3 found for vulnerability",
			args: args{
				vulnerability: scanner.Vulnerability{
					ID: "fake",
					CVSS: []scanner.CVSS{
						{
							Version: "2.0",
							Vector:  "AV:N/AC:L/Au:S/C:P/I:P/A:P",
							Metrics: scanner.CvssMetrics{
								BaseScore:           score,
								ExploitabilityScore: &score,
								ImpactScore:         &score,
							},
						},
						{
							Version: "3.1",
							Vector:  "CVSS:3.1/AV:N/AC:L/PR:H/UI:N/S:U/C:H/I:H/A:H",
							Metrics: scanner.CvssMetrics{
								BaseScore:           score,
								ExploitabilityScore: &score,
								ImpactScore:         &score,
							},
						},
					},
				},
			},
			want: &models.CVSS{
				CvssV3Metrics: &models.CVSSV3Metrics{
					BaseScore:           score,
					ExploitabilityScore: score,
					ImpactScore:         score,
				},
				CvssV3Vector: &models.CVSSV3Vector{
					AttackComplexity:   getAttackComplexity(metric.AttackComplexityLow),
					AttackVector:       getAttackVector(metric.AttackVectorNetwork),
					Availability:       getAvailability(metric.AvailabilityImpactHigh),
					Confidentiality:    getConfidentiality(metric.ConfidentialityImpactHigh),
					Integrity:          getIntegrity(metric.IntegrityImpactHigh),
					PrivilegesRequired: getPrivilegesRequired(metric.PrivilegesRequiredHigh),
					Scope:              getScope(metric.ScopeUnchanged),
					UserInteraction:    getUserInteraction(metric.UserInteractionNone),
					Vector:             "CVSS:3.1/AV:N/AC:L/PR:H/UI:N/S:U/C:H/I:H/A:H",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getCVSS(tt.args.vulnerability); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCVSS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createCISDockerBenchmarkAssesment(t *testing.T) {
	type args struct {
		assesments dockle_types.AssessmentSlice
	}
	tests := []struct {
		name string
		args args
		want []*models.CISDockerBenchmarkAssessment
	}{
		{
			name: "assesment slice is empty",
			args: args{
				assesments: dockle_types.AssessmentSlice{},
			},
			want: nil,
		},
		{
			name: "assesment slice is not empty",
			args: args{
				assesments: dockle_types.AssessmentSlice{
					{
						Code:     "testCode",
						Level:    1,
						Filename: "testFilename",
						Desc:     "testDescription",
					},
				},
			},
			want: []*models.CISDockerBenchmarkAssessment{
				{
					Code:     "testCode",
					Level:    1,
					Filename: "testFilename",
					Desc:     "testDescription",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createCISDockerBenchmarkAssesment(tt.args.assesments); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createCISDockerBenchmarkAssesment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createCISDockerBenchmarkResults(t *testing.T) {
	type args struct {
		results dockle_types.AssessmentMap
	}
	tests := []struct {
		name string
		args args
		want []*models.CISDockerBenchmarkCodeInfo
	}{
		{
			name: "assesment map is nil",
			args: args{
				results: nil,
			},
			want: nil,
		},
		{
			name: "assesment map is not nil",
			args: args{
				results: dockle_types.AssessmentMap{
					"test": {
						Code:  "mapcode",
						Level: 2,
						Assessments: dockle_types.AssessmentSlice{
							{
								Code:     "testCode",
								Level:    1,
								Filename: "testFilename",
								Desc:     "testDescription",
							},
						},
					},
				},
			},
			want: []*models.CISDockerBenchmarkCodeInfo{
				{
					Code:  "mapcode",
					Level: 2,
					Assessments: []*models.CISDockerBenchmarkAssessment{
						{
							Code:     "testCode",
							Level:    1,
							Filename: "testFilename",
							Desc:     "testDescription",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createCISDockerBenchmarkResults(tt.args.results); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createCISDockerBenchmarkResults() = %v, want %v", got, tt.want)
			}
		})
	}
}
