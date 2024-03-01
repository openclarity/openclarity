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

package discoverer

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
)

func Test_isEncrypted(t *testing.T) {
	type args struct {
		disk armcompute.DisksClientGetResponse
	}
	tests := []struct {
		name string
		args args
		want apitypes.RootVolumeEncrypted
	}{
		{
			name: "encrypted",
			args: args{
				disk: armcompute.DisksClientGetResponse{
					Disk: armcompute.Disk{
						Properties: &armcompute.DiskProperties{
							EncryptionSettingsCollection: &armcompute.EncryptionSettingsCollection{
								Enabled: to.Ptr(true),
							},
						},
					},
				},
			},
			want: apitypes.RootVolumeEncryptedYes,
		},
		{
			name: "not encrypted",
			args: args{
				disk: armcompute.DisksClientGetResponse{
					Disk: armcompute.Disk{
						Properties: &armcompute.DiskProperties{
							EncryptionSettingsCollection: nil,
						},
					},
				},
			},
			want: apitypes.RootVolumeEncryptedNo,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEncrypted(tt.args.disk); got != tt.want {
				t.Errorf("isEncrypted() = %v, want %v", got, tt.want)
			}
		})
	}
}
