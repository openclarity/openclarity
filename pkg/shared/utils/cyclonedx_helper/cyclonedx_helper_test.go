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

package cyclonedx_helper // nolint:revive,stylecheck

import (
	"reflect"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

func TestGetComponentHash(t *testing.T) {
	type args struct {
		component *cdx.Component
	}
	tests := []struct {
		name string
		args args
		want string
		err  string
	}{
		{
			name: "hashes is nil",
			args: args{
				component: &cdx.Component{},
			},
			want: "",
			err:  "no sha256 hash found in component",
		},
		{
			name: "hashes is empty",
			args: args{
				component: &cdx.Component{
					Hashes: &[]cdx.Hash{},
				},
			},
			want: "",
			err:  "no sha256 hash found in component",
		},
		{
			name: "sha256 hash exist",
			args: args{
				component: &cdx.Component{
					Hashes: &[]cdx.Hash{
						{
							Algorithm: cdx.HashAlgoMD5,
							Value:     "md5value",
						},
						{
							Algorithm: cdx.HashAlgoSHA1,
							Value:     "sha1value",
						},
						{
							Algorithm: cdx.HashAlgoSHA256,
							Value:     "sha256value",
						},
						{
							Algorithm: cdx.HashAlgoSHA512,
							Value:     "sha512value",
						},
					},
				},
			},
			want: "sha256value",
		},
		{
			name: "sha256 hash doesn't exist",
			args: args{
				component: &cdx.Component{
					Hashes: &[]cdx.Hash{
						{
							Algorithm: cdx.HashAlgoMD5,
							Value:     "md5value",
						},
						{
							Algorithm: cdx.HashAlgoSHA1,
							Value:     "sha1value",
						},
					},
				},
			},
			want: "sha1value",
		},
		{
			name: "hash doesn't exist and component type is container",
			args: args{
				component: &cdx.Component{
					Type:    cdx.ComponentTypeContainer,
					Version: "sha256:manifestDigest",
				},
			},
			want: "manifestDigest",
		},
		{
			name: "hash exist and component type is container",
			args: args{
				component: &cdx.Component{
					Type:    cdx.ComponentTypeContainer,
					Version: "sha256:manifestDigest",
					Hashes: &[]cdx.Hash{
						{
							Algorithm: cdx.HashAlgoSHA256,
							Value:     "sha256Value",
						},
					},
				},
			},
			want: "sha256Value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetComponentHash(tt.args.component)
			if got != tt.want || (err != nil && err.Error() != tt.err) {
				t.Errorf("GetComponentHash() = %v, want %v, err %v", got, tt.want, tt.err)
			}
		})
	}
}

func TestGetComponentLicenses(t *testing.T) {
	type args struct {
		component cdx.Component
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "licenses are nil in component",
			args: args{
				component: cdx.Component{
					Licenses: nil,
				},
			},
			want: nil,
		},
		{
			name: "licenses are empty in component",
			args: args{
				component: cdx.Component{
					Licenses: &cdx.Licenses{},
				},
			},
			want: nil,
		},
		{
			name: "licenses are empty in component",
			args: args{
				component: cdx.Component{
					Licenses: &cdx.Licenses{},
				},
			},
			want: nil,
		},
		{
			name: "License with no id an no name",
			args: args{
				component: cdx.Component{
					Licenses: &cdx.Licenses{
						{
							License: &cdx.License{
								ID:   "",
								Name: "",
								URL:  "test.com",
							},
							Expression: "",
						},
					},
				},
			},
			want: []string{},
		},
		{
			name: "License with id and no name",
			args: args{
				component: cdx.Component{
					Licenses: &cdx.Licenses{
						{
							License: &cdx.License{
								ID:   "test-id",
								Name: "",
								URL:  "test.com",
							},
							Expression: "",
						},
					},
				},
			},
			want: []string{"test-id"},
		},
		{
			name: "License with id and name - prefer id",
			args: args{
				component: cdx.Component{
					Licenses: &cdx.Licenses{
						{
							License: &cdx.License{
								ID:   "test-id",
								Name: "test-name",
								URL:  "test.com",
							},
							Expression: "",
						},
					},
				},
			},
			want: []string{"test-id"},
		},
		{
			name: "License with no id but with name",
			args: args{
				component: cdx.Component{
					Licenses: &cdx.Licenses{
						{
							License: &cdx.License{
								ID:   "",
								Name: "test-name",
								URL:  "test.com",
							},
							Expression: "",
						},
					},
				},
			},
			want: []string{"test-name"},
		},
		{
			name: "License with expression",
			args: args{
				component: cdx.Component{
					Licenses: &cdx.Licenses{
						{
							Expression: "test-expression",
						},
					},
				},
			},
			want: []string{"test-expression"},
		},
		{
			name: "License with id and expression",
			args: args{
				component: cdx.Component{
					Licenses: &cdx.Licenses{
						{
							License: &cdx.License{
								ID:   "test-id",
								Name: "test-name",
								URL:  "test.com",
							},
							Expression: "test-expression",
						},
					},
				},
			},
			want: []string{"test-id", "test-expression"},
		},
		{
			name: "One with license with id and expression",
			args: args{
				component: cdx.Component{
					Licenses: &cdx.Licenses{
						{
							License: &cdx.License{
								ID: "test-id",
							},
						},
						{
							License: &cdx.License{
								URL: "test.com",
							},
							Expression: "test-expression",
						},
					},
				},
			},
			want: []string{"test-id", "test-expression"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetComponentLicenses(tt.args.component); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetComponentLicenses() = %v, want %v", got, tt.want)
			}
		})
	}
}
