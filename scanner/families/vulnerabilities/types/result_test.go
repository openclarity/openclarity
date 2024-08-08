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
)

func Test_sortArrays(t *testing.T) {
	type args struct {
		vulnerability Vulnerability
	}
	tests := []struct {
		name string
		args args
		want Vulnerability
	}{
		{
			name: "sort",
			args: args{
				vulnerability: Vulnerability{
					Links: []string{"link2", "link1"},
					CVSS: []CVSS{
						{
							Version: "3",
							Vector:  "456",
						},
						{
							Version: "2",
							Vector:  "123",
						},
					},
					Fix: Fix{
						Versions: []string{"ver2", "ver1"},
					},
					Package: Package{
						Licenses: []string{"lic2", "lic1"},
						CPEs:     []string{"cpes2", "cpes1"},
					},
				},
			},
			want: Vulnerability{
				Links: []string{"link1", "link2"},
				CVSS: []CVSS{
					{
						Version: "2",
						Vector:  "123",
					},
					{
						Version: "3",
						Vector:  "456",
					},
				},
				Fix: Fix{
					Versions: []string{"ver1", "ver2"},
				},
				Package: Package{
					Licenses: []string{"lic1", "lic2"},
					CPEs:     []string{"cpes1", "cpes2"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.vulnerability.sorted(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortArrays() = %v, want %v", got, tt.want)
			}
		})
	}
}
