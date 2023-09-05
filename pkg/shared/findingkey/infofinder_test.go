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

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

func TestGenerateInfoFinderKey(t *testing.T) {
	type args struct {
		info models.InfoFinderFindingInfo
	}
	tests := []struct {
		name string
		args args
		want InfoFinderKey
	}{
		{
			name: "sanity",
			args: args{
				info: models.InfoFinderFindingInfo{
					Data:        utils.PointerTo("data"),
					Path:        utils.PointerTo("path"),
					ScannerName: utils.PointerTo("scanner"),
					Type:        utils.PointerTo(models.InfoTypeSSHAuthorizedKeyFingerprint),
				},
			},
			want: InfoFinderKey{
				ScannerName: "scanner",
				Type:        string(models.InfoTypeSSHAuthorizedKeyFingerprint),
				Data:        "data",
				Path:        "path",
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
		ScannerName string
		Type        string
		Data        string
		Path        string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "sanity",
			fields: fields{
				ScannerName: "scanner",
				Type:        string(models.InfoTypeSSHAuthorizedKeyFingerprint),
				Data:        "data",
				Path:        "path",
			},
			want: "scanner.SSHAuthorizedKeyFingerprint.data.path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := InfoFinderKey{
				ScannerName: tt.fields.ScannerName,
				Type:        tt.fields.Type,
				Data:        tt.fields.Data,
				Path:        tt.fields.Path,
			}
			if got := k.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
