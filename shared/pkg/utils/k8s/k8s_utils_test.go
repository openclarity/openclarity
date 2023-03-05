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

package k8s

import (
	"testing"
)

func TestParseImageHash(t *testing.T) {
	type args struct {
		imageID string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "valid image id",
			args: args{
				imageID: "docker-pullable://gcr.io/development-infra-208909/kubeclarity@sha256:6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa",
			},
			want: "6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa",
		},
		{
			name: "no image hash",
			args: args{
				imageID: "docker-pullable://gcr.io/development-infra-208909/kubeclarity@sha256:",
			},
			want: "",
		},
		{
			name: "no image id",
			args: args{
				imageID: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseImageHash(tt.args.imageID); got != tt.want {
				t.Errorf("ParseImageHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeImageID(t *testing.T) {
	type args struct {
		imageID string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "image id with docker-pullable prefix",
			args: args{
				imageID: "docker-pullable://gcr.io/development-infra-208909/kubeclarity@sha256:6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa",
			},
			want: "gcr.io/development-infra-208909/kubeclarity@sha256:6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa",
		},
		{
			name: "image id with docker-pullable prefix - not normalized",
			args: args{
				imageID: "docker-pullable://mongo@sha256:4200c3073389d5b303070e53ff8f5e4472efb534340d28599458ccc24f378025",
			},
			want: "docker.io/library/mongo@sha256:4200c3073389d5b303070e53ff8f5e4472efb534340d28599458ccc24f378025",
		},
		{
			name: "image id without docker-pullable prefix",
			args: args{
				imageID: "gcr.io/development-infra-208909/kubeclarity@sha256:6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa",
			},
			want: "gcr.io/development-infra-208909/kubeclarity@sha256:6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa",
		},
		{
			name: "non pullable image id, do not normalize",
			args: args{
				imageID: "sha256:6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa",
			},
			want: "sha256:6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa",
		},
		{
			name: "no image id",
			args: args{
				imageID: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeImageID(tt.args.imageID); got != tt.want {
				t.Errorf("NormalizeImageID() = %v, want %v", got, tt.want)
			}
		})
	}
}
