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

package utils

import (
	"testing"
)

func Test_setImageSource(t *testing.T) {
	type args struct {
		local  bool
		source string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "local image source",
			args: args{
				local:  true,
				source: "test:latest",
			},
			want: "docker:test:latest",
		},
		{
			name: "remote image source",
			args: args{
				source: "test:latest",
			},
			want: "registry:test:latest",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := setImageSource(tt.args.local, tt.args.source); got != tt.want {
				t.Errorf("SetSource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateSource(t *testing.T) {
	type args struct {
		sourceType SourceType
		src        string
		localImage bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "local image source",
			args: args{
				sourceType: IMAGE,
				localImage: true,
				src:        "test:latest",
			},
			want: "docker:test:latest",
		},
		{
			name: "remote image source",
			args: args{
				sourceType: IMAGE,
				src:        "test:latest",
			},
			want: "registry:test:latest",
		},
		{
			name: "local image source",
			args: args{
				sourceType: DIR,
				localImage: true,
				src:        "test/path",
			},
			want: "dir:test/path",
		},
		{
			name: "remote image source",
			args: args{
				sourceType: FILE,
				src:        "test/path",
			},
			want: "file:test/path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateSource(tt.args.sourceType, tt.args.src, tt.args.localImage); got != tt.want {
				t.Errorf("CreateSource() = %v, want %v", got, tt.want)
			}
		})
	}
}
