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

package filter

import (
	"testing"

	"github.com/openclarity/kubeclarity/shared/pkg/scanner/types"
)

func Test_shouldIgnore(t *testing.T) {
	type args struct {
		vulnerability types.Vulnerability
		ignores       Ignores
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "no ignores",
			args: args{
				vulnerability: types.Vulnerability{
					ID: "CVE-123",
				},
			},
			want: false,
		},
		{
			name: "ignore no fix, the vulnerability doesn't have any fix",
			args: args{
				vulnerability: types.Vulnerability{
					ID: "CVE-123",
				},
				ignores: Ignores{
					NoFix: true,
				},
			},
			want: true,
		},
		{
			name: "ignore no fix, the vulnerability has fix",
			args: args{
				vulnerability: types.Vulnerability{
					ID: "CVE-123",
					Fix: types.Fix{
						Versions: []string{
							"1.1.1",
						},
					},
				},
				ignores: Ignores{
					NoFix: true,
				},
			},
			want: false,
		},
		{
			name: "the vulnerability is in the ignore list",
			args: args{
				vulnerability: types.Vulnerability{
					ID: "CVE-123",
				},
				ignores: Ignores{
					Vulnerabilities: []string{
						"CVE-123",
						"CVE-234",
					},
				},
			},
			want: true,
		},
		{
			name: "ignore no fix, the vulnerability has fix but it's in the ignore list",
			args: args{
				vulnerability: types.Vulnerability{
					ID: "CVE-123",
					Fix: types.Fix{
						Versions: []string{
							"1.1.1",
						},
					},
				},
				ignores: Ignores{
					NoFix: true,
					Vulnerabilities: []string{
						"CVE-123",
						"CVE-234",
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldIgnore(tt.args.vulnerability, tt.args.ignores); got != tt.want {
				t.Errorf("shouldIgnore() = %v, want %v", got, tt.want)
			}
		})
	}
}
