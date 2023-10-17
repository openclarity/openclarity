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

package scanestimation

import (
	"testing"
)

func Test_findClosestSSDSizeSymbol(t *testing.T) {
	type args struct {
		diskSize int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "E1",
			args: args{
				diskSize: 3,
			},
			want: "E1",
		},
		{
			name: "E3",
			args: args{
				diskSize: 16,
			},
			want: "E3",
		},
		{
			name: "E60",
			args: args{
				diskSize: 8191,
			},
			want: "E60",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findClosestSSDSizeSymbol(tt.args.diskSize); got != tt.want {
				t.Errorf("findClosestSSDSizeSymbol() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_addQueryParamToURL(t *testing.T) {
	type args struct {
		baseURL string
		key     string
		value   string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "",
			args: args{
				baseURL: priceListBaseURL,
				key:     "$filter",
				value:   "'foo' eq 'bar'",
			},
			want:    priceListBaseURL + "&$filter=%27foo%27+eq+%27bar%27",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := addQueryParamToURL(tt.args.baseURL, tt.args.key, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("addQueryParamToURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("addQueryParamToURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}
