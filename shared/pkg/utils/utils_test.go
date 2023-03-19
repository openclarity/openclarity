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

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
)

func TestGetVulnerabilityTotalsPerSeverity(t *testing.T) {
	type args struct {
		vulnerabilities *[]models.Vulnerability
	}
	tests := []struct {
		name string
		args args
		want *models.VulnerabilityScanSummary
	}{
		{
			name: "nil should result in empty",
			args: args{
				vulnerabilities: nil,
			},
			want: &models.VulnerabilityScanSummary{
				TotalCriticalVulnerabilities:   utils.PointerTo(0),
				TotalHighVulnerabilities:       utils.PointerTo(0),
				TotalMediumVulnerabilities:     utils.PointerTo(0),
				TotalLowVulnerabilities:        utils.PointerTo(0),
				TotalNegligibleVulnerabilities: utils.PointerTo(0),
			},
		},
		{
			name: "check one type",
			args: args{
				vulnerabilities: utils.PointerTo([]models.Vulnerability{
					{
						Description:       utils.PointerTo("desc1"),
						Severity:          utils.PointerTo(models.CRITICAL),
						VulnerabilityName: utils.PointerTo("CVE-1"),
					},
				}),
			},
			want: &models.VulnerabilityScanSummary{
				TotalCriticalVulnerabilities:   utils.PointerTo(1),
				TotalHighVulnerabilities:       utils.PointerTo(0),
				TotalMediumVulnerabilities:     utils.PointerTo(0),
				TotalLowVulnerabilities:        utils.PointerTo(0),
				TotalNegligibleVulnerabilities: utils.PointerTo(0),
			},
		},
		{
			name: "check all severity types",
			args: args{
				vulnerabilities: utils.PointerTo([]models.Vulnerability{
					{
						Description:       utils.PointerTo("desc1"),
						Severity:          utils.PointerTo(models.CRITICAL),
						VulnerabilityName: utils.PointerTo("CVE-1"),
					},
					{
						Description:       utils.PointerTo("desc2"),
						Severity:          utils.PointerTo(models.HIGH),
						VulnerabilityName: utils.PointerTo("CVE-2"),
					},
					{
						Description:       utils.PointerTo("desc3"),
						Severity:          utils.PointerTo(models.MEDIUM),
						VulnerabilityName: utils.PointerTo("CVE-3"),
					},
					{
						Description:       utils.PointerTo("desc4"),
						Severity:          utils.PointerTo(models.LOW),
						VulnerabilityName: utils.PointerTo("CVE-4"),
					},
					{
						Description:       utils.PointerTo("desc5"),
						Severity:          utils.PointerTo(models.NEGLIGIBLE),
						VulnerabilityName: utils.PointerTo("CVE-5"),
					},
				}),
			},
			want: &models.VulnerabilityScanSummary{
				TotalCriticalVulnerabilities:   utils.PointerTo(1),
				TotalHighVulnerabilities:       utils.PointerTo(1),
				TotalMediumVulnerabilities:     utils.PointerTo(1),
				TotalLowVulnerabilities:        utils.PointerTo(1),
				TotalNegligibleVulnerabilities: utils.PointerTo(1),
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
