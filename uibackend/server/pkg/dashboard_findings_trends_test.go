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

package server

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"gotest.tools/v3/assert"

	"github.com/openclarity/vmclarity/uibackend/types"
)

func Test_createTimes(t *testing.T) {
	type args struct {
		startTime, endTime string
	}
	tests := []struct {
		name string
		args args
		want []time.Time
	}{
		{
			name: "10 min duration",
			args: args{
				startTime: "2006-01-02T15:00:00Z",
				endTime:   "2006-01-02T15:10:00Z",
			},
			want: []time.Time{
				mustParse(t, "2006-01-02T15:01:00Z"),
				mustParse(t, "2006-01-02T15:02:00Z"),
				mustParse(t, "2006-01-02T15:03:00Z"),
				mustParse(t, "2006-01-02T15:04:00Z"),
				mustParse(t, "2006-01-02T15:05:00Z"),
				mustParse(t, "2006-01-02T15:06:00Z"),
				mustParse(t, "2006-01-02T15:07:00Z"),
				mustParse(t, "2006-01-02T15:08:00Z"),
				mustParse(t, "2006-01-02T15:09:00Z"),
				mustParse(t, "2006-01-02T15:10:00Z"),
			},
		},
		{
			name: "1 day duration",
			args: args{
				startTime: "2006-01-02T15:20:00Z",
				endTime:   "2006-01-03T15:20:00Z",
			},
			want: []time.Time{
				mustParse(t, "2006-01-02T17:44:00Z"),
				mustParse(t, "2006-01-02T20:08:00Z"),
				mustParse(t, "2006-01-02T22:32:00Z"),
				mustParse(t, "2006-01-03T00:56:00Z"),
				mustParse(t, "2006-01-03T03:20:00Z"),
				mustParse(t, "2006-01-03T05:44:00Z"),
				mustParse(t, "2006-01-03T08:08:00Z"),
				mustParse(t, "2006-01-03T10:32:00Z"),
				mustParse(t, "2006-01-03T12:56:00Z"),
				mustParse(t, "2006-01-03T15:20:00Z"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := types.GetDashboardFindingsTrendsParams{
				StartTime: mustParse(t, tt.args.startTime),
				EndTime:   mustParse(t, tt.args.endTime),
			}
			if diff := cmp.Diff(tt.want, createTimes(params)); diff != "" {
				t.Errorf("createTimes mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func mustParse(t *testing.T, timeStr string) time.Time {
	t.Helper()
	ret, err := time.Parse(time.RFC3339, timeStr)
	assert.NilError(t, err)
	return ret
}

func Test_validateParams(t *testing.T) {
	type args struct {
		params types.GetDashboardFindingsTrendsParams
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "start time before end time",
			args: args{
				params: types.GetDashboardFindingsTrendsParams{
					StartTime: mustParse(t, "2006-01-03T10:32:00Z"),
					EndTime:   mustParse(t, "2006-01-03T12:56:00Z"),
				},
			},
			wantErr: false,
		},
		{
			name: "start time and end time are the same",
			args: args{
				params: types.GetDashboardFindingsTrendsParams{
					StartTime: mustParse(t, "2006-01-03T10:32:00Z"),
					EndTime:   mustParse(t, "2006-01-03T10:32:00Z"),
				},
			},
			wantErr: true,
		},
		{
			name: "start time after end time",
			args: args{
				params: types.GetDashboardFindingsTrendsParams{
					StartTime: mustParse(t, "2006-01-03T12:56:00Z"),
					EndTime:   mustParse(t, "2006-01-03T10:32:00Z"),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateParams(tt.args.params); (err != nil) != tt.wantErr {
				t.Errorf("validateParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getObjectType(t *testing.T) {
	type args struct {
		findingType types.FindingType
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Exploit",
			args: args{
				findingType: types.EXPLOIT,
			},
			want: "Exploit",
		},
		{
			name: "Malware",
			args: args{
				findingType: types.MALWARE,
			},
			want: "Malware",
		},
		{
			name: "Misconfiguration",
			args: args{
				findingType: types.MISCONFIGURATION,
			},
			want: "Misconfiguration",
		},
		{
			name: "Package",
			args: args{
				findingType: types.PACKAGE,
			},
			want: "Package",
		},
		{
			name: "Rootkit",
			args: args{
				findingType: types.ROOTKIT,
			},
			want: "Rootkit",
		},
		{
			name: "Secret",
			args: args{
				findingType: types.SECRET,
			},
			want: "Secret",
		},
		{
			name: "Vulnerability",
			args: args{
				findingType: types.VULNERABILITY,
			},
			want: "Vulnerability",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getObjectType(tt.args.findingType); got != tt.want {
				t.Errorf("getObjectType() = %v, want %v", got, tt.want)
			}
		})
	}
}
