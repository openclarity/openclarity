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

package export

import (
	"reflect"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	purl "github.com/package-url/packageurl-go"

	"github.com/cisco-open/kubei/api/client/models"
)

func Test_getPackageInfo(t *testing.T) {
	type args struct {
		pkgInfo cdx.Component
	}
	tests := []struct {
		name string
		args args
		want *models.PackageInfo
	}{
		{
			name: "empty license info",
			args: args{
				pkgInfo: cdx.Component{
					Name:    "test",
					Version: "1.0.0",
					PackageURL: purl.NewPackageURL(
						"golang",
						"github.com/test",
						"test",
						"1.0.0",
						purl.Qualifiers{}, "").ToString(),
				},
			},
			want: &models.PackageInfo{
				Name:     "test",
				Version:  "1.0.0",
				Language: "go",
				License:  "",
			},
		},
		{
			name: "with license info",
			args: args{
				pkgInfo: cdx.Component{
					Name:    "test",
					Version: "1.0.0",
					Licenses: &cdx.Licenses{
						{
							License: &cdx.License{
								ID: "MIT",
							},
						},
					},
					PackageURL: purl.NewPackageURL(
						"golang",
						"github.com/test",
						"test",
						"1.0.0",
						purl.Qualifiers{}, "").ToString(),
				},
			},
			want: &models.PackageInfo{
				Name:     "test",
				Version:  "1.0.0",
				Language: "go",
				License:  "MIT",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPackageInfo(tt.args.pkgInfo); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPackageInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
