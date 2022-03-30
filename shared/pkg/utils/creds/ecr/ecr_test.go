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

package ecr

import (
	"testing"

	"github.com/containers/image/v5/docker/reference"
)

func TestECR_IsSupported(t *testing.T) {
	matchNamed, _ := reference.ParseNormalizedNamed("674200998650.dkr.ecr.eu-central-1.amazonaws.com/test/test:test")
	noMatchNamed, _ := reference.ParseNormalizedNamed("gcr.io/test/test:test")
	type args struct {
		named reference.Named
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "match",
			args: args{
				named: matchNamed,
			},
			want: true,
		},
		{
			name: "not match",
			args: args{
				named: noMatchNamed,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ECR{}
			if got := e.IsSupported(tt.args.named); got != tt.want {
				t.Errorf("IsSupported() = %v, want %v", got, tt.want)
			}
		})
	}
}
