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

package utils

import (
	"reflect"
	"testing"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
)

func TestGetVulnerabilityTotalsPerSeverity(t *testing.T) {
	type args struct {
		vulnerabilities *[]apitypes.Vulnerability
	}
	tests := []struct {
		name string
		args args
		want *apitypes.VulnerabilitySeveritySummary
	}{
		{
			name: "nil should result in empty",
			args: args{
				vulnerabilities: nil,
			},
			want: &apitypes.VulnerabilitySeveritySummary{
				TotalCriticalVulnerabilities:   to.Ptr(0),
				TotalHighVulnerabilities:       to.Ptr(0),
				TotalMediumVulnerabilities:     to.Ptr(0),
				TotalLowVulnerabilities:        to.Ptr(0),
				TotalNegligibleVulnerabilities: to.Ptr(0),
			},
		},
		{
			name: "check one type",
			args: args{
				vulnerabilities: to.Ptr([]apitypes.Vulnerability{
					{
						Description:       to.Ptr("desc1"),
						Severity:          to.Ptr(apitypes.CRITICAL),
						VulnerabilityName: to.Ptr("CVE-1"),
					},
				}),
			},
			want: &apitypes.VulnerabilitySeveritySummary{
				TotalCriticalVulnerabilities:   to.Ptr(1),
				TotalHighVulnerabilities:       to.Ptr(0),
				TotalMediumVulnerabilities:     to.Ptr(0),
				TotalLowVulnerabilities:        to.Ptr(0),
				TotalNegligibleVulnerabilities: to.Ptr(0),
			},
		},
		{
			name: "check all severity types",
			args: args{
				vulnerabilities: to.Ptr([]apitypes.Vulnerability{
					{
						Description:       to.Ptr("desc1"),
						Severity:          to.Ptr(apitypes.CRITICAL),
						VulnerabilityName: to.Ptr("CVE-1"),
					},
					{
						Description:       to.Ptr("desc2"),
						Severity:          to.Ptr(apitypes.HIGH),
						VulnerabilityName: to.Ptr("CVE-2"),
					},
					{
						Description:       to.Ptr("desc3"),
						Severity:          to.Ptr(apitypes.MEDIUM),
						VulnerabilityName: to.Ptr("CVE-3"),
					},
					{
						Description:       to.Ptr("desc4"),
						Severity:          to.Ptr(apitypes.LOW),
						VulnerabilityName: to.Ptr("CVE-4"),
					},
					{
						Description:       to.Ptr("desc5"),
						Severity:          to.Ptr(apitypes.NEGLIGIBLE),
						VulnerabilityName: to.Ptr("CVE-5"),
					},
				}),
			},
			want: &apitypes.VulnerabilitySeveritySummary{
				TotalCriticalVulnerabilities:   to.Ptr(1),
				TotalHighVulnerabilities:       to.Ptr(1),
				TotalMediumVulnerabilities:     to.Ptr(1),
				TotalLowVulnerabilities:        to.Ptr(1),
				TotalNegligibleVulnerabilities: to.Ptr(1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetVulnerabilityTotalsPerSeverity(tt.args.vulnerabilities); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetVulnerabilityTotalsPerSeverity() = %v, want %v", got, tt.want)
			}
		})
	}
}
