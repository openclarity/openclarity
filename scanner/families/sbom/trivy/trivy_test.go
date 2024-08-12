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

package trivy

import (
	"reflect"
	"testing"

	"github.com/openclarity/vmclarity/scanner/utils/image_helper"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

func Test_getImageHashAndProperties(t *testing.T) {
	type args struct {
		properties *[]cdx.Property
		src        string
	}
	tests := []struct {
		name     string
		args     args
		wantHash string
		wantInfo *image_helper.ImageInfo
		wantErr  bool
	}{
		{
			name: "nil properties",
			args: args{
				properties: nil,
				src:        "",
			},
			wantHash: "",
			wantInfo: nil,
			wantErr:  true,
		},
		{
			name: "empty properties",
			args: args{
				properties: &[]cdx.Property{},
				src:        "",
			},
			wantHash: "",
			wantInfo: nil,
			wantErr:  true,
		},
		{
			name: "both RepoDigest and ImageID properties are missing",
			args: args{
				properties: &[]cdx.Property{
					{
						Name:  "name",
						Value: "value",
					},
				},
				src: "",
			},
			wantHash: "",
			wantInfo: nil,
			wantErr:  true,
		},
		{
			name: "RepoDigest is missing and ImageID is not",
			args: args{
				properties: &[]cdx.Property{
					{
						Name:  "aquasecurity:trivy:ImageID",
						Value: "sha256:62ed8ed20fdbb57a19639fc3a2dc8710dd66cb2364d61ec02e11cf9b35bc31dc",
					},
				},
				src: "",
			},
			wantHash: "62ed8ed20fdbb57a19639fc3a2dc8710dd66cb2364d61ec02e11cf9b35bc31dc",
			wantInfo: &image_helper.ImageInfo{
				ID: "sha256:62ed8ed20fdbb57a19639fc3a2dc8710dd66cb2364d61ec02e11cf9b35bc31dc",
			},
			wantErr: false,
		},
		{
			name: "RepoDigest is not missing and ImageID is missing",
			args: args{
				properties: &[]cdx.Property{
					{
						Name:  "aquasecurity:trivy:RepoDigest",
						Value: "poke/debian@sha256:a4c378901a2ba14fd331e96a49101556e91ed592d5fd68ba7405fdbf9b969e61",
					},
				},
				src: "poke/debian:latest",
			},
			wantHash: "a4c378901a2ba14fd331e96a49101556e91ed592d5fd68ba7405fdbf9b969e61",
			wantInfo: &image_helper.ImageInfo{
				Name:    "poke/debian:latest",
				Digests: []string{"poke/debian@sha256:a4c378901a2ba14fd331e96a49101556e91ed592d5fd68ba7405fdbf9b969e61"},
			},
			wantErr: false,
		},
		{
			name: "RepoDigest is not missing and ImageID is not missing - prefer RepoDigest",
			args: args{
				properties: &[]cdx.Property{
					{
						Name:  "aquasecurity:trivy:ImageID",
						Value: "sha256:62ed8ed20fdbb57a19639fc3a2dc8710dd66cb2364d61ec02e11cf9b35bc31dc",
					},
					{
						Name:  "aquasecurity:trivy:RepoDigest",
						Value: "poke/debian@sha256:a4c378901a2ba14fd331e96a49101556e91ed592d5fd68ba7405fdbf9b969e61",
					},
				},
				src: "poke/debian:latest",
			},
			wantHash: "a4c378901a2ba14fd331e96a49101556e91ed592d5fd68ba7405fdbf9b969e61",

			wantInfo: &image_helper.ImageInfo{
				Name:    "poke/debian:latest",
				ID:      "sha256:62ed8ed20fdbb57a19639fc3a2dc8710dd66cb2364d61ec02e11cf9b35bc31dc",
				Digests: []string{"poke/debian@sha256:a4c378901a2ba14fd331e96a49101556e91ed592d5fd68ba7405fdbf9b969e61"},
			},
			wantErr: false,
		},
		{
			name: "RepoDigest is not missing and ImageID is not missing - prefer RepoDigest and match the correct RepoDigest matching src",
			args: args{
				properties: &[]cdx.Property{
					{
						Name:  "aquasecurity:trivy:ImageID",
						Value: "sha256:62ed8ed20fdbb57a19639fc3a2dc8710dd66cb2364d61ec02e11cf9b35bc31dc",
					},
					{
						Name:  "aquasecurity:trivy:RepoDigest",
						Value: "debian@sha256:2906804d2a64e8a13a434a1a127fe3f6a28bf7cf3696be4223b06276f32f1f2d",
					},
					{
						Name:  "aquasecurity:trivy:RepoDigest",
						Value: "poke/debian@sha256:a4c378901a2ba14fd331e96a49101556e91ed592d5fd68ba7405fdbf9b969e61",
					},
				},
				src: "poke/debian:latest",
			},
			wantHash: "a4c378901a2ba14fd331e96a49101556e91ed592d5fd68ba7405fdbf9b969e61",
			wantInfo: &image_helper.ImageInfo{
				Name: "poke/debian:latest",
				ID:   "sha256:62ed8ed20fdbb57a19639fc3a2dc8710dd66cb2364d61ec02e11cf9b35bc31dc",
				Digests: []string{
					"debian@sha256:2906804d2a64e8a13a434a1a127fe3f6a28bf7cf3696be4223b06276f32f1f2d",
					"poke/debian@sha256:a4c378901a2ba14fd331e96a49101556e91ed592d5fd68ba7405fdbf9b969e61",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHash, imageInfo, err := getImageInfo(tt.args.properties, tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("getImageInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotHash != tt.wantHash {
				t.Errorf("getImageInfo() got hash = %v, want hash %v", gotHash, tt.wantHash)
			}
			if !reflect.DeepEqual(imageInfo, tt.wantInfo) {
				t.Errorf("getImageInfo() got properties = %v, want properties %v", imageInfo, tt.wantInfo)
			}
		})
	}
}
