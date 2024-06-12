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

package cli

import (
	"testing"

	"github.com/openclarity/vmclarity/utils/fsutils/filesystem"
)

func Test_isSupportedFS(t *testing.T) {
	type args struct {
		fs filesystem.FilesystemType
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ext2 is supported",
			args: args{
				fs: filesystem.Ext2,
			},
			want: true,
		},
		{
			name: "ext3 is supported",
			args: args{
				fs: filesystem.Ext3,
			},
			want: true,
		},
		{
			name: "ext4 is supported",
			args: args{
				fs: filesystem.Ext4,
			},
			want: true,
		},
		{
			name: "xfs is supported",
			args: args{
				fs: filesystem.Xfs,
			},
			want: true,
		},
		{
			name: "ntfs is supported",
			args: args{
				fs: filesystem.Ntfs,
			},
			want: true,
		},
		{
			name: "btrfs is not supported",
			args: args{
				fs: "btrfs",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSupportedFS(string(tt.args.fs)); got != tt.want {
				t.Errorf("isSupportedFS() = %v, want %v", got, tt.want)
			}
		})
	}
}
