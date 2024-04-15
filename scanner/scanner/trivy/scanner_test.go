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
	"testing"

	trivyDBTypes "github.com/aquasecurity/trivy-db/pkg/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/openclarity/vmclarity/scanner/scanner"
)

func Test_getTypeFromPurl(t *testing.T) {
	tests := []struct {
		name string
		purl string
		want string
		err  bool
	}{
		{
			name: "good purl",
			purl: `pkg:deb/debian/libsemanage1@3.1-1+b2?arch=amd64\u0026upstream=libsemanage%403.1-1\u0026distro=debian-11`,
			want: "deb",
			err:  false,
		},
		{
			name: "bad purl",
			purl: `deb/debian/libsemanage1@3.1-1+b2?arch=amd64\u0026upstream=libsemanage%403.1-1\u0026distro=debian-11`,
			want: "",
			err:  true,
		},
		{
			name: "empty purl",
			purl: ``,
			want: "",
			err:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getTypeFromPurl(tt.purl)
			if tt.err {
				if err == nil {
					t.Errorf("getTypeFromPurl() expected error, got: %v", got)
				}
			} else {
				if err != nil {
					t.Errorf("getTypeFromPurl() unexpected error: %v", err)
				}

				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("getTypeFromPurl() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func Test_getCVSSesFromVul(t *testing.T) {
	v3vector1 := "CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:H/A:N"
	v3score1 := 7.5
	v3exploit1 := 3.9
	v3impact1 := 3.6
	v3version1 := "3.0"

	v3vector2 := "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:L/I:N/A:N"
	v3score2 := 5.3
	v3exploit2 := 3.9
	v3impact2 := 1.4
	v3version2 := "3.1"

	v2vector := "AV:L/AC:H/Au:N/C:C/I:N/A:N"
	v2score := 4.0
	v2exploit := 1.9
	v2impact := 6.9
	v2version := "2.0"

	tests := []struct {
		name   string
		cvsses trivyDBTypes.VendorCVSS
		want   []scanner.CVSS
	}{
		{
			name:   "no cvsses",
			cvsses: trivyDBTypes.VendorCVSS{},
			want:   []scanner.CVSS{},
		},
		{
			name: "one of each",
			cvsses: trivyDBTypes.VendorCVSS{
				"nvd": {
					V2Vector: v2vector,
					V2Score:  v2score,
				},
				"redhat": {
					V3Vector: v3vector1,
					V3Score:  v3score1,
				},
			},
			want: []scanner.CVSS{
				{
					Version: v3version1,
					Vector:  v3vector1,
					Metrics: scanner.CvssMetrics{
						BaseScore:           v3score1,
						ExploitabilityScore: &v3exploit1,
						ImpactScore:         &v3impact1,
					},
				},
				{
					Version: v2version,
					Vector:  v2vector,
					Metrics: scanner.CvssMetrics{
						BaseScore:           v2score,
						ExploitabilityScore: &v2exploit,
						ImpactScore:         &v2impact,
					},
				},
			},
		},
		{
			name: "both have both",
			cvsses: trivyDBTypes.VendorCVSS{
				"nvd": {
					V2Vector: v2vector,
					V2Score:  v2score,
					V3Vector: v3vector2,
					V3Score:  v3score2,
				},
				"redhat": {
					V2Vector: v2vector,
					V2Score:  v2score,
					V3Vector: v3vector1,
					V3Score:  v3score1,
				},
			},
			want: []scanner.CVSS{
				{
					Version: v3version2,
					Vector:  v3vector2,
					Metrics: scanner.CvssMetrics{
						BaseScore:           v3score2,
						ExploitabilityScore: &v3exploit2,
						ImpactScore:         &v3impact2,
					},
				},
				{
					Version: v2version,
					Vector:  v2vector,
					Metrics: scanner.CvssMetrics{
						BaseScore:           v2score,
						ExploitabilityScore: &v2exploit,
						ImpactScore:         &v2impact,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCVSSesFromVul(tt.cvsses)
			if diff := cmp.Diff(tt.want, got, cmpopts.SortSlices(func(a, b scanner.CVSS) bool { return a.Vector < b.Vector })); diff != "" {
				t.Errorf("getCVSSesFromVul() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
