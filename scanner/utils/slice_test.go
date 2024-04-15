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

package utils

import "testing"

func TestContains_strings(t *testing.T) {
	type args struct {
		s []string
		v string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil slice",
			args: args{
				s: nil,
				v: "str",
			},
			want: false,
		},
		{
			name: "empty slice",
			args: args{
				s: []string{},
				v: "str",
			},
			want: false,
		},
		{
			name: "empty value",
			args: args{
				s: []string{"str", "str1"},
				v: "",
			},
			want: false,
		},
		{
			name: "string slice - contains",
			args: args{
				s: []string{"str", "str1"},
				v: "str",
			},
			want: true,
		},
		{
			name: "string slice - contains twice",
			args: args{
				s: []string{"str", "str1", "str"},
				v: "str",
			},
			want: true,
		},
		{
			name: "string slice - does not contains",
			args: args{
				s: []string{"str", "str1"},
				v: "str2",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(tt.args.s, tt.args.v); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContains_int(t *testing.T) {
	type args struct {
		s []int
		v int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil slice",
			args: args{
				s: nil,
				v: 1,
			},
			want: false,
		},
		{
			name: "empty slice",
			args: args{
				s: []int{},
				v: 1,
			},
			want: false,
		},
		{
			name: "int slice - contains",
			args: args{
				s: []int{1, 2},
				v: 1,
			},
			want: true,
		},
		{
			name: "int slice - contains twice",
			args: args{
				s: []int{1, 2, 1},
				v: 1,
			},
			want: true,
		},
		{
			name: "int slice - does not contains",
			args: args{
				s: []int{1, 2},
				v: 3,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(tt.args.s, tt.args.v); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}
