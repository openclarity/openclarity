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

package configwatcher

import (
	"reflect"
	"testing"
	"time"
)

func Test_isWithinTheWindow(t *testing.T) {
	type args struct {
		operationTime time.Time
		now           time.Time
		window        time.Duration
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "in the window",
			args: args{
				operationTime: time.Now().Add(2 * time.Minute),
				now:           time.Now(),
				window:        3 * time.Minute,
			},
			want: true,
		},
		{
			name: "not in the window - before now",
			args: args{
				operationTime: time.Now().Add(-2 * time.Minute),
				now:           time.Now(),
				window:        3 * time.Minute,
			},
			want: false,
		},
		{
			name: "not in the window - after now",
			args: args{
				operationTime: time.Now().Add(2 * time.Minute),
				now:           time.Now(),
				window:        1 * time.Minute,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isWithinTheWindow(tt.args.operationTime, tt.args.now, tt.args.window); got != tt.want {
				t.Errorf("test() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findFirstOperationTimeInTheFuture(t *testing.T) {
	now, _ := time.Parse(time.RFC3339, "2006-01-02T20:00:00Z")
	oneHourFromNow := now.Add(1 * time.Hour)
	fiveHoursFromNow := now.Add(5 * time.Hour)
	fiveHoursBeforeNow := now.Add(-5 * time.Hour)
	type args struct {
		operationTime time.Time
		now           time.Time
		cronLine      string
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "operation time already in the future",
			args: args{
				operationTime: fiveHoursFromNow,
				now:           now,
				cronLine:      "0 */4 * * *",
			},
			want: fiveHoursFromNow,
		},
		{
			name: "operation time in the past",
			args: args{
				operationTime: fiveHoursBeforeNow,
				now:           now,
				cronLine:      "0 */3 * * *",
			},
			want: oneHourFromNow,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findFirstOperationTimeInTheFuture(tt.args.operationTime, tt.args.now, tt.args.cronLine); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findFirstOperationTimeInTheFuture() = %v, want %v", got, tt.want)
			}
		})
	}
}
