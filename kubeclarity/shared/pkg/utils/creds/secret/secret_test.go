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

package secret

import (
	"context"
	"testing"

	"github.com/containers/image/v5/docker/reference"
)

func TestImagePullSecret_GetCredentials(t *testing.T) {
	gcrImage, _ := reference.ParseNormalizedNamed("gcr.io/library/image:123")
	gcrImageSpecific, _ := reference.ParseNormalizedNamed("gcr.io/more/specific:123")
	noMatch, _ := reference.ParseNormalizedNamed("no/match:123")
	imageNoScheme, _ := reference.ParseNormalizedNamed("foo.example.com/image:123")
	imageDockerHub, _ := reference.ParseNormalizedNamed("foo")
	type fields struct {
		body string
	}
	type args struct {
		in0   context.Context
		named reference.Named
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantUsername string
		wantPassword string
		wantErr      bool
	}{
		{
			name: "gcr no specific image",
			fields: fields{
				body: "{\"auths\":{\"gcr.io\":{\"username\":\"gcr\",\"password\":\"io\",\"auth\":\"Z2NyOmlv\"},\"gcr.io/more/specific\":{\"username\":\"gcr\",\"password\":\"io/more/specific\",\"auth\":\"Z2NyOmlvL21vcmUvc3BlY2lmaWM=\"},\"http://foo.example.com\":{\"username\":\"foo\",\"password\":\"bar\",\"auth\":\"Zm9vOmJhcg==\"}}}",
			},
			args: args{
				named: gcrImage,
			},
			wantUsername: "gcr",
			wantPassword: "io",
			wantErr:      false,
		},
		{
			name: "gcr specific image",
			fields: fields{
				body: "{\"auths\":{\"gcr.io\":{\"username\":\"gcr\",\"password\":\"io\",\"auth\":\"Z2NyOmlv\"},\"gcr.io/more/specific\":{\"username\":\"gcr\",\"password\":\"io/more/specific\",\"auth\":\"Z2NyOmlvL21vcmUvc3BlY2lmaWM=\"},\"http://foo.example.com\":{\"username\":\"foo\",\"password\":\"bar\",\"auth\":\"Zm9vOmJhcg==\"}}}",
			},
			args: args{
				named: gcrImageSpecific,
			},
			wantUsername: "gcr",
			wantPassword: "io/more/specific",
			wantErr:      false,
		},
		{
			name: "no match",
			fields: fields{
				body: "{\"auths\":{\"gcr.io\":{\"username\":\"gcr\",\"password\":\"io\",\"auth\":\"Z2NyOmlv\"},\"gcr.io/more/specific\":{\"username\":\"gcr\",\"password\":\"io/more/specific\",\"auth\":\"Z2NyOmlvL21vcmUvc3BlY2lmaWM=\"},\"http://foo.example.com\":{\"username\":\"foo\",\"password\":\"bar\",\"auth\":\"Zm9vOmJhcg==\"}}}",
			},
			args: args{
				named: noMatch,
			},
			wantUsername: "",
			wantPassword: "",
			wantErr:      false,
		},
		{
			name: "match registry with scheme",
			fields: fields{
				body: "{\"auths\":{\"gcr.io\":{\"username\":\"gcr\",\"password\":\"io\",\"auth\":\"Z2NyOmlv\"},\"gcr.io/more/specific\":{\"username\":\"gcr\",\"password\":\"io/more/specific\",\"auth\":\"Z2NyOmlvL21vcmUvc3BlY2lmaWM=\"},\"http://foo.example.com\":{\"username\":\"foo\",\"password\":\"bar\",\"auth\":\"Zm9vOmJhcg==\"}}}",
			},
			args: args{
				named: imageNoScheme,
			},
			wantUsername: "foo",
			wantPassword: "bar",
			wantErr:      false,
		},
		{
			name: "match docker hub registry",
			fields: fields{
				body: "{\"auths\":{\"https://index.docker.io/v1/\":{\"username\":\"docker\",\"password\":\"io\",\"auth\":\"ZG9ja2VyOmlv\"},\"gcr.io/more/specific\":{\"username\":\"gcr\",\"password\":\"io/more/specific\",\"auth\":\"Z2NyOmlvL21vcmUvc3BlY2lmaWM=\"},\"http://foo.example.com\":{\"username\":\"foo\",\"password\":\"bar\",\"auth\":\"Zm9vOmJhcg==\"}}}",
			},
			args: args{
				named: imageDockerHub,
			},
			wantUsername: "docker",
			wantPassword: "io",
			wantErr:      false,
		},
		{
			name: "malformed body",
			fields: fields{
				body: "malformed body",
			},
			args: args{
				named: imageDockerHub,
			},
			wantUsername: "",
			wantPassword: "",
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ImagePullSecret{
				body: tt.fields.body,
			}
			gotUsername, gotPassword, err := s.GetCredentials(tt.args.in0, tt.args.named)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCredentials() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUsername != tt.wantUsername {
				t.Errorf("GetCredentials() gotUsername = %v, want %v", gotUsername, tt.wantUsername)
			}
			if gotPassword != tt.wantPassword {
				t.Errorf("GetCredentials() gotPassword = %v, want %v", gotPassword, tt.wantPassword)
			}
		})
	}
}
