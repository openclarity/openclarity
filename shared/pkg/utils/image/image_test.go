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

package image

import "testing"

func TestIsLocalImage(t *testing.T) {
	type args struct {
		imageID string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "docker://sha256 prefix",
			args: args{
				imageID: "docker://sha256:12bae74413f7240099ba68a4b44c55541fa94c51c676681c2988a7571e6891eb",
			},
			want: true,
		},
		{
			name: "sha256: prefix",
			args: args{
				imageID: "sha256:12bae74413f7240099ba68a4b44c55541fa94c51c676681c2988a7571e6891eb",
			},
			want: true,
		},
		{
			name: "good",
			args: args{
				imageID: "gke.gcr.io/proxy-agent@sha256:d5ae8affd1ca510a4bfd808e14a563c573510a70196ad5b04fdf0fb5425abf35",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsLocalImage(tt.args.imageID); got != tt.want {
				t.Errorf("IsLocalImage() = %v, want %v", got, tt.want)
			}
		})
	}
}
