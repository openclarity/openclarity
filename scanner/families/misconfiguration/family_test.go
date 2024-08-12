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

package misconfiguration

import (
	"reflect"
	"testing"

	"github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
)

func TestStripPathFromResult(t *testing.T) {
	type args struct {
		result types.ScannerResult
		path   string
	}
	tests := []struct {
		name string
		args args
		want types.ScannerResult
	}{
		{
			name: "sanity",
			args: args{
				result: types.ScannerResult{
					ScannerName: "scanner1",
					Misconfigurations: []types.Misconfiguration{
						{
							Location:    "/mnt/foo",
							Category:    "test1",
							ID:          "id1",
							Description: "desc1",
						},
						{
							Location:    "/mnt/foo2",
							Category:    "test2",
							ID:          "id2",
							Description: "desc2",
						},
						{
							Location:    "/foo3",
							Category:    "test3",
							ID:          "id3",
							Description: "desc3",
						},
					},
				},
				path: "/mnt",
			},
			want: types.ScannerResult{
				ScannerName: "scanner1",
				Misconfigurations: []types.Misconfiguration{
					{
						Location:    "/foo",
						Category:    "test1",
						ID:          "id1",
						Description: "desc1",
					},
					{
						Location:    "/foo2",
						Category:    "test2",
						ID:          "id2",
						Description: "desc2",
					},
					{
						Location:    "/foo3",
						Category:    "test3",
						ID:          "id3",
						Description: "desc3",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripPathFromResult(tt.args.result, tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StripPathFromResult() = %v, want %v", got, tt.want)
			}
		})
	}
}
