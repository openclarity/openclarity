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

package analyzer

import (
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_purlStringToStruct(t *testing.T) {
	type args struct {
		purlInput string
	}
	tests := []struct {
		name string
		args args
		want purl
	}{
		{
			name: "valid purl",
			args: args{
				purlInput: "pkg:maven/org.apache.xmlgraphics/batik-anim@1.9.1?packaging=sources",
			},
			want: purl{
				scheme:    "pkg",
				typ:       "maven",
				namespace: "org.apache.xmlgraphics",
				name:      "batik-anim",
				version:   "1.9.1",
				qualifers: url.Values{
					"packaging": []string{"sources"},
				},
				subpath: "",
			},
		},
		{
			name: "empty purl",
			args: args{
				purlInput: "",
			},
			want: newPurl(),
		},
		{
			name: "junk",
			args: args{
				purlInput: "asdkjnasdkhjbasdhb",
			},
			want: newPurl(),
		},
		{
			name: "no namespace purl",
			args: args{
				purlInput: "pkg:pypi/Babel@2.9.1",
			},
			want: purl{
				scheme:    "pkg",
				typ:       "pypi",
				namespace: "",
				name:      "Babel",
				version:   "2.9.1",
				qualifers: url.Values{},
				subpath:   "",
			},
		},
		{
			name: "no type or namespace",
			args: args{
				purlInput: "pkg:Babel@2.9.1",
			},
			want: newPurl(),
		},
		{
			name: "alpine package missing type",
			args: args{
				purlInput: "pkg:alpine/apk@2.12.9-r3?arch=x86",
			},
			want: purl{
				scheme:    "pkg",
				typ:       "apk",
				namespace: "alpine",
				name:      "apk",
				version:   "2.12.9-r3",
				qualifers: url.Values{
					"arch": []string{"x86"},
				},
				subpath: "",
			},
		},
		{
			name: "no namespace, unknown type",
			args: args{
				purlInput: "pkg:debi/curl@7.50.3-1?arch=i386&distro=jessie",
			},
			want: newPurl(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := purlStringToStruct(tt.args.purlInput)
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(purl{})); diff != "" {
				t.Errorf("purlStringToStruct() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
