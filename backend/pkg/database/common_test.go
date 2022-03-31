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

package database

import "testing"

func TestToDBArrayElement(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "string with no |",
			args: args{
				s: "foo",
			},
			want: "|foo|",
		},
		{
			name: "string with | at the middle - strip",
			args: args{
				s: "fo|o",
			},
			want: "|foo|",
		},
		{
			name: "string with | at the start and middle - strip",
			args: args{
				s: "fo|o|",
			},
			want: "|foo|",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToDBArrayElement(tt.args.s); got != tt.want {
				t.Errorf("ToDBArrayElement() = %v, want %v", got, tt.want)
			}
		})
	}
}
