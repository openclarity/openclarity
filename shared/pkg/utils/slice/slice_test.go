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

package slice

import (
	"reflect"
	"sort"
	"testing"
)

func TestRemoveStringDuplicates(t *testing.T) {
	type args struct {
		slice []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "has duplicates",
			args: args{
				slice: []string{"a", "a", "a", "b"},
			},
			want: []string{"a", "b"},
		},
		{
			name: "no duplicates",
			args: args{
				slice: []string{"a1", "a2", "a3", "b"},
			},
			want: []string{"a1", "a2", "a3", "b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveStringDuplicates(tt.args.slice)
			sort.Strings(got)
			sort.Strings(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveStringDuplicates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveEmptyStrings(t *testing.T) {
	type args struct {
		slice []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty slice",
			args: args{
				slice: []string{},
			},
			want: nil,
		},
		{
			name: "no empty elements - nothing to remove",
			args: args{
				slice: []string{"test1", "test2"},
			},
			want: []string{"test1", "test2"},
		},
		{
			name: "empty elements - need to remove",
			args: args{
				slice: []string{"test1", "", "", "test2"},
			},
			want: []string{"test1", "test2"},
		},
		{
			name: "all empty elements - need to remove all",
			args: args{
				slice: []string{"", ""},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveEmptyStrings(tt.args.slice); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveEmptyStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}
