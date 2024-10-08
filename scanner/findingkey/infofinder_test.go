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

package findingkey

import (
	"reflect"
	"testing"

	apitypes "github.com/openclarity/openclarity/api/types"
	"github.com/openclarity/openclarity/core/to"
)

func TestGenerateInfoFinderKey(t *testing.T) {
	type args struct {
		info apitypes.InfoFinderFindingInfo
	}
	tests := []struct {
		name string
		args args
		want InfoFinderKey
	}{
		{
			name: "sanity",
			args: args{
				info: apitypes.InfoFinderFindingInfo{
					Data: to.Ptr("data"),
					Path: to.Ptr("path"),
					Type: to.Ptr(apitypes.InfoTypeSSHAuthorizedKeyFingerprint),
				},
			},
			want: InfoFinderKey{
				Type: string(apitypes.InfoTypeSSHAuthorizedKeyFingerprint),
				Data: "data",
				Path: "path",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateInfoFinderKey(tt.args.info); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateInfoFinderKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInfoFinderKey_String(t *testing.T) {
	type fields struct {
		Type string
		Data string
		Path string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "sanity",
			fields: fields{
				Type: string(apitypes.InfoTypeSSHAuthorizedKeyFingerprint),
				Data: "data",
				Path: "path",
			},
			want: "SSHAuthorizedKeyFingerprint.data.path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := InfoFinderKey{
				Type: tt.fields.Type,
				Data: tt.fields.Data,
				Path: tt.fields.Path,
			}
			if got := k.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
