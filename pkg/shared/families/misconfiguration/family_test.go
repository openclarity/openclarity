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

	"github.com/openclarity/vmclarity/pkg/shared/families/misconfiguration/types"
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
							ScannedPath:     "/mnt/foo",
							TestCategory:    "test1",
							TestID:          "id1",
							TestDescription: "desc1",
						},
						{
							ScannedPath:     "/mnt/foo2",
							TestCategory:    "test2",
							TestID:          "id2",
							TestDescription: "desc2",
						},
						{
							ScannedPath:     "/foo3",
							TestCategory:    "test3",
							TestID:          "id3",
							TestDescription: "desc3",
						},
					},
				},
				path: "/mnt",
			},
			want: types.ScannerResult{
				ScannerName: "scanner1",
				Misconfigurations: []types.Misconfiguration{
					{
						ScannedPath:     "/foo",
						TestCategory:    "test1",
						TestID:          "id1",
						TestDescription: "desc1",
					},
					{
						ScannedPath:     "/foo2",
						TestCategory:    "test2",
						TestID:          "id2",
						TestDescription: "desc2",
					},
					{
						ScannedPath:     "/foo3",
						TestCategory:    "test3",
						TestID:          "id3",
						TestDescription: "desc3",
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
