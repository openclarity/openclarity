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

package utils

import (
	"testing"

	"github.com/openclarity/vmclarity/scanner/common"
)

func TestGenerateHash(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "non-empty dir",
			args: args{
				s: "testdata",
			},
			want: "93039ae6c8721d9acb744804c624edf91e5caf7912e0b709c81af0e3eb14bda6",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateHash(common.DIR, tt.args.s)
			if err != nil {
				t.Errorf("GenerateHash() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("GenerateHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
