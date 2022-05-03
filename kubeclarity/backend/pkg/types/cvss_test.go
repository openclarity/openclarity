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
	"reflect"
	"testing"

	"github.com/openclarity/kubeclarity/api/server/models"
)

func TestCVSS_ToCVSSBackendAPI(t *testing.T) {
	type fields struct {
		cvss *CVSS
	}
	tests := []struct {
		name   string
		fields fields
		want   *models.CVSS
	}{
		{
			name: "nil cvss",
			fields: fields{
				cvss: nil,
			},
			want: nil,
		},
		{
			name: "nil cvss fields",
			fields: fields{
				cvss: &CVSS{},
			},
			want: &models.CVSS{
				CvssV3Metrics: nil,
				CvssV3Vector:  nil,
			},
		},
		{
			name: "sanity",
			fields: fields{
				cvss: &CVSS{
					CvssV3Metrics: &CVSSV3Metrics{
						BaseScore:           1.1,
						ExploitabilityScore: 2.2,
						ImpactScore:         3.3,
					},
					CvssV3Vector: &CVSSV3Vector{
						AttackComplexity:   AttackComplexityLOW,
						AttackVector:       AttackVectorNETWORK,
						Availability:       AvailabilityLOW,
						Confidentiality:    ConfidentialityHIGH,
						Integrity:          IntegrityHIGH,
						PrivilegesRequired: PrivilegesRequiredHIGH,
						Scope:              ScopeUNCHANGED,
						UserInteraction:    UserInteractionREQUIRED,
						Vector:             "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H",
					},
				},
			},
			want: &models.CVSS{
				CvssV3Metrics: &models.CVSSV3Metrics{
					BaseScore:           1.1,
					ExploitabilityScore: 2.2,
					ImpactScore:         3.3,
					Severity:            models.VulnerabilitySeverityLOW,
				},
				CvssV3Vector: &models.CVSSV3Vector{
					AttackComplexity:   models.AttackComplexityLOW,
					AttackVector:       models.AttackVectorNETWORK,
					Availability:       models.AvailabilityLOW,
					Confidentiality:    models.ConfidentialityHIGH,
					Integrity:          models.IntegrityHIGH,
					PrivilegesRequired: models.PrivilegesRequiredHIGH,
					Scope:              models.ScopeUNCHANGED,
					UserInteraction:    models.UserInteractionREQUIRED,
					Vector:             "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fields.cvss.ToCVSSBackendAPI(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToCVSSBackendAPI() = %v, want %v", got, tt.want)
			}
		})
	}
}
