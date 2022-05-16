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

package scheduler

import (
	"testing"
	"time"

	"gotest.tools/assert"
)

func Test_getStartsAt(t *testing.T) {
	timeNow, err := time.Parse(time.RFC3339, "2022-05-08T18:23:21+00:00")
	assert.NilError(t, err)

	type args struct {
		timeNow   time.Time
		startTime time.Time
		interval  time.Duration
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{
			name: "start time is 10 seconds from now - next scan starts in 10 seconds from now",
			args: args{
				timeNow:   timeNow,
				startTime: timeNow.Add(10 * time.Second),
				interval:  5 * time.Second,
			},
			want: 10 * time.Second,
		},
		{
			name: "start time is 10 seconds before now, next scan should be now (5 + 5 = 10)",
			args: args{
				timeNow:   timeNow,
				startTime: timeNow.Add(-10 * time.Second),
				interval:  5 * time.Second,
			},
			want: 0,
		},
		{
			name: "start time is 13 seconds before now, next scan should be in 2 seconds (5 + 5 + 5 = 15 -> 15 -13 = 2)",
			args: args{
				timeNow:   timeNow,
				startTime: timeNow.Add(-13 * time.Second),
				interval:  5 * time.Second,
			},
			want: 2 * time.Second,
		},
		{
			name: "start time is 0.5 second after now, next scan should be in 0.5 second",
			args: args{
				timeNow:   timeNow,
				startTime: timeNow.Add(500 * time.Millisecond),
				interval:  5 * time.Second,
			},
			want: 500 * time.Millisecond,
		},
		{
			name: "start time is 0.5 second before now, next scan should be now",
			args: args{
				timeNow:   timeNow,
				startTime: timeNow.Add(-500 * time.Millisecond),
				interval:  5 * time.Second,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStartsAt(tt.args.timeNow, tt.args.startTime, tt.args.interval); got != tt.want {
				t.Errorf("getStartsAt() = %v, want %v", got, tt.want)
			}
		})
	}
}
