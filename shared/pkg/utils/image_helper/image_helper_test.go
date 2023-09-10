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

package image_helper // nolint:revive,stylecheck

import (
	_ "crypto/sha256"
	"testing"
)

func TestGetRepoDigest(t *testing.T) {
	type args struct {
		repoDigests []string
		imageName   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "missing imageName",
			args: args{},
			want: "",
		},
		{
			name: "missing repoDigests",
			args: args{
				imageName: "poke/test",
			},
			want: "",
		},
		{
			name: "RepoDigests doesn't match the source",
			args: args{
				repoDigests: []string{
					"debian@sha256:2906804d2a64e8a13a434a1a127fe3f6a28bf7cf3696be4223b06276f32f1f2d",
					"poke/debian@sha256:a4c378901a2ba14fd331e96a49101556e91ed592d5fd68ba7405fdbf9b969e61",
				},
				imageName: "poke/test",
			},
			want: "",
		},
		{
			name: "RepoDigests match the source and source has tag",
			args: args{
				repoDigests: []string{
					"debian@sha256:2906804d2a64e8a13a434a1a127fe3f6a28bf7cf3696be4223b06276f32f1f2d",
					"poke/debian@sha256:a4c378901a2ba14fd331e96a49101556e91ed592d5fd68ba7405fdbf9b969e61",
				},
				imageName: "poke/debian:latest",
			},
			want: "a4c378901a2ba14fd331e96a49101556e91ed592d5fd68ba7405fdbf9b969e61",
		},
		{
			name: "RepoDigests match the source and source doesn't have tag",
			args: args{
				repoDigests: []string{
					"debian@sha256:2906804d2a64e8a13a434a1a127fe3f6a28bf7cf3696be4223b06276f32f1f2d",
					"poke/debian@sha256:a4c378901a2ba14fd331e96a49101556e91ed592d5fd68ba7405fdbf9b969e61",
				},
				imageName: "poke/debian",
			},
			want: "a4c378901a2ba14fd331e96a49101556e91ed592d5fd68ba7405fdbf9b969e61",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetHashFromRepoDigest(tt.args.repoDigests, tt.args.imageName); got != tt.want {
				t.Errorf("GetHashFromRepoDigest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetHashFromRepoOrManifestDigest(t *testing.T) {
	type args struct {
		repoDigests    []string
		manifestDigest string
		imageName      string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "RepoDigests is not missing",
			args: args{
				manifestDigest: "sha256:38f8c1d9613f3f42e7969c3b1dd5c3277e635d4576713e6453c6193e66270a6d",
				repoDigests: []string{
					"debian@sha256:2906804d2a64e8a13a434a1a127fe3f6a28bf7cf3696be4223b06276f32f1f2d",
					"poke/debian@sha256:a4c378901a2ba14fd331e96a49101556e91ed592d5fd68ba7405fdbf9b969e61",
				},
				imageName: "poke/debian:latest",
			},
			want: "a4c378901a2ba14fd331e96a49101556e91ed592d5fd68ba7405fdbf9b969e61",
		},
		{
			name: "RepoDigests is missing",
			args: args{
				manifestDigest: "sha256:38f8c1d9613f3f42e7969c3b1dd5c3277e635d4576713e6453c6193e66270a6d",
				repoDigests:    nil,
				imageName:      "poke/debian:latest",
			},
			want: "38f8c1d9613f3f42e7969c3b1dd5c3277e635d4576713e6453c6193e66270a6d",
		},
		{
			name: "RepoDigests is missing, manifestDigest is missing :",
			args: args{
				manifestDigest: "38f8c1d9613f3f42e7969c3b1dd5c3277e635d4576713e6453c6193e66270a6d",
				repoDigests:    nil,
				imageName:      "poke/debian:latest",
			},
			want: "38f8c1d9613f3f42e7969c3b1dd5c3277e635d4576713e6453c6193e66270a6d",
		},
		{
			name: "Both RepoDigests and ManifestDigest is missing",
			args: args{
				manifestDigest: "",
				repoDigests:    nil,
				imageName:      "poke/debian:latest",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetHashFromRepoOrManifestDigest(tt.args.repoDigests, tt.args.manifestDigest, tt.args.imageName); got != tt.want {
				t.Errorf("GetHashFromRepoOrManifestDigest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stripDockerMetaFromCommand(t *testing.T) {
	type args struct {
		command string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{
				command: "",
			},
			want: "",
		},
		{
			name: "space strip",
			args: args{
				command: "    space strip   ",
			},
			want: "space strip",
		},
		{
			name: "no strip",
			args: args{
				command: "bin/sh #(nop) CMD [/bin/bash]",
			},
			want: "bin/sh #(nop) CMD [/bin/bash]",
		},
		{
			name: "strip with #(nop)",
			args: args{
				command: "/bin/sh -c #(nop)           CMD [/bin/bash]      ",
			},
			want: "CMD [/bin/bash]",
		},
		{
			name: "strip without #(nop)",
			args: args{
				command: "/bin/sh -c         CMD [/bin/bash]      ",
			},
			want: "CMD [/bin/bash]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripDockerMetaFromCommand(tt.args.command); got != tt.want {
				t.Errorf("stripDockerMetaFromCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
