// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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

package cisdocker

import (
	"testing"

	dockle_types "github.com/Portshift/dockle/pkg/types"
	"github.com/google/go-cmp/cmp"

	"github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
	"github.com/openclarity/vmclarity/scanner/utils"
)

func TestParseDockleReport(t *testing.T) {
	tests := []struct {
		name          string
		assessmentMap *dockle_types.AssessmentMap
		imageName     string
		want          []types.Misconfiguration
	}{
		{
			name: "Fatal level",
			assessmentMap: &dockle_types.AssessmentMap{
				"CIS-DI-0009": dockle_types.CodeInfo{
					Code:  "CIS-DI-0009",
					Level: 6,
					Assessments: dockle_types.AssessmentSlice{
						&dockle_types.Assessment{
							Code:     "CIS-DI-0009",
							Level:    5,
							Filename: "metadata",
							Desc:     "Use COPY : /bin/sh-c #(nop) ADD file:81c0a803075715d1a6b4f75a29f8a01b21cc170cfc1bff6702317d1be2fe71a3 in /app/credentials.json",
						},
					},
				},
			},
			imageName: "goodwithtech/test-image:v1",
			want: []types.Misconfiguration{
				{
					Location:    "goodwithtech/test-image:v1",
					Category:    "best-practice",
					ID:          "CIS-DI-0009",
					Description: "Use COPY : /bin/sh-c #(nop) ADD file:81c0a803075715d1a6b4f75a29f8a01b21cc170cfc1bff6702317d1be2fe71a3 in /app/credentials.json\n",
					Severity:    types.HighSeverity,
					Message:     "Use COPY instead of ADD in Dockerfile",
				},
			},
		},
		{
			name: "Warn level",
			assessmentMap: &dockle_types.AssessmentMap{
				"CIS-DI-0001": dockle_types.CodeInfo{
					Code:  "CIS-DI-0001",
					Level: 5,
					Assessments: dockle_types.AssessmentSlice{
						&dockle_types.Assessment{
							Code:     "CIS-DI-0001",
							Level:    5,
							Filename: "metadata",
							Desc:     "Last user should not be root",
						},
					},
				},
			},
			imageName: "goodwithtech/test-image:v1",
			want: []types.Misconfiguration{
				{
					Location:    "goodwithtech/test-image:v1",
					Category:    "best-practice",
					ID:          "CIS-DI-0001",
					Description: "Last user should not be root\n",
					Severity:    types.MediumSeverity,
					Message:     "Create a user for the container",
				},
			},
		},
		{
			name: "Info level",
			assessmentMap: &dockle_types.AssessmentMap{
				"CIS-DI-0008": dockle_types.CodeInfo{
					Code:  "CIS-DI-0008",
					Level: 4,
					Assessments: dockle_types.AssessmentSlice{
						&dockle_types.Assessment{
							Code:     "CIS-DI-0008",
							Level:    4,
							Filename: "/usr/lib/openssh/ssh-keysign",
							Desc:     "setuid file: urwxr-xr-x /usr/lib/openssh/ssh-keysign",
						},
						&dockle_types.Assessment{
							Code:     "CIS-DI-0008",
							Level:    4,
							Filename: "/usr/bin/chsh",
							Desc:     "setuid file: urwxr-xr-x /usr/bin/chsh",
						},
						&dockle_types.Assessment{
							Code:     "CIS-DI-0008",
							Level:    4,
							Filename: "/bin/su",
							Desc:     "setuid file: urwxr-xr-x /bin/su",
						},
					},
				},
			},
			imageName: "goodwithtech/test-image:v1",
			want: []types.Misconfiguration{
				{
					Location:    "goodwithtech/test-image:v1",
					Category:    "best-practice",
					ID:          "CIS-DI-0008",
					Description: "setuid file: urwxr-xr-x /usr/lib/openssh/ssh-keysign\nsetuid file: urwxr-xr-x /usr/bin/chsh\nsetuid file: urwxr-xr-x /bin/su\n",
					Severity:    types.LowSeverity,
					Message:     "Confirm safety of setuid/setgid files",
				},
			},
		},
		{
			name: "Pass level",
			assessmentMap: &dockle_types.AssessmentMap{
				"PASS": dockle_types.CodeInfo{
					Code:        "PASS",
					Level:       1,
					Assessments: dockle_types.AssessmentSlice{},
				},
			},
			imageName: "goodwithtech/test-image:v1",
			want:      []types.Misconfiguration{},
		},
		{
			name: "Ignore level",
			assessmentMap: &dockle_types.AssessmentMap{
				"IGNORE": dockle_types.CodeInfo{
					Code:        "IGNORE",
					Level:       2,
					Assessments: dockle_types.AssessmentSlice{},
				},
			},
			imageName: "goodwithtech/test-image:v1",
			want:      []types.Misconfiguration{},
		},
		{
			name: "Skip level",
			assessmentMap: &dockle_types.AssessmentMap{
				"SKIP": dockle_types.CodeInfo{
					Code:        "SKIP",
					Level:       3,
					Assessments: dockle_types.AssessmentSlice{},
				},
			},
			imageName: "goodwithtech/test-image:v1",
			want:      []types.Misconfiguration{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDockleReport(utils.IMAGE, tt.imageName, *tt.assessmentMap)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NewReportParser() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
