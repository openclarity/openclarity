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

	"github.com/cisco-open/kubei/api/server/models"
)

func TestCVSSV3Metrics_getCVSSSeverity(t *testing.T) {
	/*
		https://nvd.nist.gov/vuln-metrics/cvss
		CVSS v3.0 Ratings
			Severity	Base Score Range
			None		0.0
			Low			0.1-3.9
			Medium		4.0-6.9
			High		7.0-8.9
			Critical	9.0-10.0
	*/
	type fields struct {
		BaseScore float64
	}
	tests := []struct {
		name   string
		fields fields
		want   models.VulnerabilitySeverity
	}{
		{
			name: "None",
			fields: fields{
				BaseScore: 0,
			},
			want: "",
		},
		{
			name: "Low min",
			fields: fields{
				BaseScore: 0.1,
			},
			want: models.VulnerabilitySeverityLOW,
		},
		{
			name: "Low max",
			fields: fields{
				BaseScore: 3.9,
			},
			want: models.VulnerabilitySeverityLOW,
		},
		{
			name: "Medium min",
			fields: fields{
				BaseScore: 4,
			},
			want: models.VulnerabilitySeverityMEDIUM,
		},
		{
			name: "Medium max",
			fields: fields{
				BaseScore: 6.9,
			},
			want: models.VulnerabilitySeverityMEDIUM,
		},
		{
			name: "High min",
			fields: fields{
				BaseScore: 7,
			},
			want: models.VulnerabilitySeverityHIGH,
		},
		{
			name: "High max",
			fields: fields{
				BaseScore: 8.9,
			},
			want: models.VulnerabilitySeverityHIGH,
		},
		{
			name: "Critical min",
			fields: fields{
				BaseScore: 9,
			},
			want: models.VulnerabilitySeverityCRITICAL,
		},
		{
			name: "Critical max",
			fields: fields{
				BaseScore: 10,
			},
			want: models.VulnerabilitySeverityCRITICAL,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &CVSSV3Metrics{
				BaseScore: tt.fields.BaseScore,
			}
			if got := m.getCVSSSeverity(); got != tt.want {
				t.Errorf("getCVSSSeverity() = %v, want %v", got, tt.want)
			}
		})
	}
}
