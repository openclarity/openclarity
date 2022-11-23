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

package dependency_track // nolint:revive,stylecheck

import (
	"reflect"
	"testing"
)

func Test_getLinks(t *testing.T) {
	type args struct {
		references string
	}
	tests := []struct {
		name      string
		args      args
		wantLinks []string
	}{
		{
			name: "sanity",
			args: args{
				references: "* [https://salsa.debian.org/apt-team/apt/-/commit/dceb1e49e4b8e4dadaf056be34088b415939cda6](https://salsa.debian.org/apt-team/apt/-/commit/dceb1e49e4b8e4dadaf056be34088b415939cda6)\n* [https://github.com/Debian/apt/issues/111](https://github.com/Debian/apt/issues/111)\n* [https://bugs.launchpad.net/bugs/1878177](https://bugs.launchpad.net/bugs/1878177)\n* [https://lists.debian.org/debian-security-announce/2020/msg00089.html](https://lists.debian.org/debian-security-announce/2020/msg00089.html)\n* [https://tracker.debian.org/news/1144109/accepted-apt-212-source-into-unstable/](https://tracker.debian.org/news/1144109/accepted-apt-212-source-into-unstable/)\n* [https://usn.ubuntu.com/4359-1/](https://usn.ubuntu.com/4359-1/)\n* [https://usn.ubuntu.com/4359-2/](https://usn.ubuntu.com/4359-2/)\n* [https://lists.fedoraproject.org/archives/list/package-announce@lists.fedoraproject.org/message/U4PEH357MZM2SUGKETMEHMSGQS652QHH/](https://lists.fedoraproject.org/archives/list/package-announce@lists.fedoraproject.org/message/U4PEH357MZM2SUGKETMEHMSGQS652QHH/)",
			},
			wantLinks: []string{
				"https://salsa.debian.org/apt-team/apt/-/commit/dceb1e49e4b8e4dadaf056be34088b415939cda6",
				"https://github.com/Debian/apt/issues/111",
				"https://bugs.launchpad.net/bugs/1878177",
				"https://lists.debian.org/debian-security-announce/2020/msg00089.html",
				"https://tracker.debian.org/news/1144109/accepted-apt-212-source-into-unstable/",
				"https://usn.ubuntu.com/4359-1/",
				"https://usn.ubuntu.com/4359-2/",
				"https://lists.fedoraproject.org/archives/list/package-announce@lists.fedoraproject.org/message/U4PEH357MZM2SUGKETMEHMSGQS652QHH/",
			},
		},
		{
			name: "empty",
			args: args{
				references: "",
			},
			wantLinks: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotLinks := getLinks(tt.args.references); !reflect.DeepEqual(gotLinks, tt.wantLinks) {
				t.Errorf("getLinks() = %v, want %v", gotLinks, tt.wantLinks)
			}
		})
	}
}

func Test_removeParentheses(t *testing.T) {
	type args struct {
		vector string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no parentheses",
			args: args{
				vector: "AV:L/AC:M/Au:N/C:N/I:N/A:P",
			},
			want: "AV:L/AC:M/Au:N/C:N/I:N/A:P",
		},
		{
			name: "remove parentheses",
			args: args{
				vector: "(AV:L/AC:M/Au:N/C:N/I:N/A:P)",
			},
			want: "AV:L/AC:M/Au:N/C:N/I:N/A:P",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeParentheses(tt.args.vector); got != tt.want {
				t.Errorf("removeParentheses() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_unescapePurl(t *testing.T) {
	type args struct {
		purl string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "unescape needed",
			args: args{
				purl: "pkg:deb/debian/libxml2@2.9.4%20dfsg1-7%20deb10u2?arch=amd64",
			},
			want: "pkg:deb/debian/libxml2@2.9.4+dfsg1-7+deb10u2?arch=amd64",
		},
		{
			name: "unescape not needed",
			args: args{
				purl: "pkg:deb/debian/apt@1.8.2.3?arch=amd64",
			},
			want: "pkg:deb/debian/apt@1.8.2.3?arch=amd64",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unescapePurl(tt.args.purl); got != tt.want {
				t.Errorf("unescapePurl() = %v, want %v", got, tt.want)
			}
		})
	}
}
