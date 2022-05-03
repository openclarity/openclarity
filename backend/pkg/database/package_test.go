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

package database

import (
	"reflect"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/openclarity/kubeclarity/backend/pkg/types"
)

func TestCreatePackage(t *testing.T) {
	pkgInfo := &types.PackageInfo{
		Language: "Language",
		License:  "License",
		Name:     "Name",
		Version:  "Version",
	}
	vulnerability := Vulnerability{
		ID:                "ID",
		Name:              "Name",
		ScannedAt:         time.Now(),
		Severity:          0,
		Description:       "Description",
		Links:             "Links",
		CVSS:              "CVSS",
		ReportingScanners: "ReportingScanners",
		Source:            "Source",
	}
	type args struct {
		pkg  *types.PackageInfo
		vuls []Vulnerability
	}
	tests := []struct {
		name string
		args args
		want *Package
	}{
		{
			name: "sanity",
			args: args{
				pkg: pkgInfo,
				vuls: []Vulnerability{
					vulnerability,
				},
			},
			want: &Package{
				ID:       CreatePackageID(pkgInfo),
				Language: pkgInfo.Language,
				License:  pkgInfo.License,
				Name:     pkgInfo.Name,
				Version:  pkgInfo.Version,
				Vulnerabilities: []Vulnerability{
					vulnerability,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreatePackage(tt.args.pkg, tt.args.vuls); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreatePackage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreatePackageID(t *testing.T) {
	pkgInfo := &types.PackageInfo{
		Language: "Language",
		License:  "License",
		Name:     "Name",
		Version:  "Version",
	}
	type args struct {
		pkgInfo *types.PackageInfo
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "sanity",
			args: args{
				pkgInfo: pkgInfo,
			},
			want: uuid.NewV5(uuid.Nil, pkgInfo.Name+"."+pkgInfo.Version).String(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreatePackageID(tt.args.pkgInfo); got != tt.want {
				t.Errorf("CreatePackageID() = %v, want %v", got, tt.want)
			}
		})
	}
}
