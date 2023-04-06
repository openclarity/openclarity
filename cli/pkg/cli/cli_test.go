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
)

func Test_isSupportedFS(t *testing.T) {
	type args struct {
		fs string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "supported ext4",
			args: args{
				fs: fsTypeExt4,
			},
			want: true,
		},
		{
			name: "supported xfs",
			args: args{
				fs: fsTypeXFS,
			},
			want: true,
		},
		{
			name: "not supported btrfs",
			args: args{
				fs: "btrfs",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSupportedFS(tt.args.fs); got != tt.want {
				t.Errorf("isSupportedFS() = %v, want %v", got, tt.want)
			}
		})
	}
}
