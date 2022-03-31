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
	}{
		{
			name: "hashes is nil",
			args: args{
				component: &cdx.Component{},
			},
			want: "",
		},
		{
			name: "hashes is empty",
			args: args{
				component: &cdx.Component{
					Hashes: &[]cdx.Hash{},
				},
			},
			want: "",
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
			if got := GetComponentHash(tt.args.component); got != tt.want {
				t.Errorf("GetComponentHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
