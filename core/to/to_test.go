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

package to

import (
	"log"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// nolint:forcetypeassert
func Test_ValuesSlice(t *testing.T) {
	type TestObject struct {
		TestInt     int
		TestStr     string
		TestPointer *bool
	}
	type args struct {
		m map[string]any
	}
	tests := []struct {
		name string
		args args
		want []any
	}{
		{
			name: "nil map",
			args: args{
				m: nil,
			},
			want: []any{},
		},
		{
			name: "empty map",
			args: args{
				m: map[string]any{},
			},
			want: []any{},
		},
		{
			name: "string to int map",
			args: args{
				m: map[string]any{
					"a": 1,
					"b": 2,
					"c": 3,
				},
			},
			want: []any{1, 2, 3},
		},
		{
			name: "string to object map",
			args: args{
				m: map[string]any{
					"a": TestObject{
						TestInt:     1,
						TestStr:     "1",
						TestPointer: Ptr(true),
					},
					"b": TestObject{
						TestInt:     2,
						TestStr:     "2",
						TestPointer: Ptr(true),
					},
					"c": TestObject{
						TestInt:     3,
						TestStr:     "3",
						TestPointer: Ptr(false),
					},
				},
			},
			want: []any{
				TestObject{
					TestInt:     1,
					TestStr:     "1",
					TestPointer: Ptr(true),
				},
				TestObject{
					TestInt:     2,
					TestStr:     "2",
					TestPointer: Ptr(true),
				},
				TestObject{
					TestInt:     3,
					TestStr:     "3",
					TestPointer: Ptr(false),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Values(tt.args.m)
			if got != nil {
				sort.Slice(got, func(i, j int) bool {
					switch got[0].(type) {
					case int:
						return got[i].(int) < got[j].(int)
					case TestObject:
						return got[i].(TestObject).TestInt < got[j].(TestObject).TestInt
					default:
						log.Fatalf("unknown type returned %T", got[0])
					}
					return false
				})
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ValuesSlice() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_ValueOrZero(t *testing.T) {
	type args struct {
		val *int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "nil should be 0",
			args: args{
				val: nil,
			},
			want: 0,
		},
		{
			name: "not nil",
			args: args{
				val: Ptr(5),
			},
			want: 5,
		},
		{
			name: "not nil 0",
			args: args{
				val: Ptr(0),
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValueOrZero(tt.args.val); got != tt.want {
				t.Errorf("ValueOrZero() = %v, want %v", got, tt.want)
			}
		})
	}
}
