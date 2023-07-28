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

package chkrootkit

import (
	"os"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gotest.tools/v3/assert"

	chkrootkitutils "github.com/openclarity/vmclarity/pkg/shared/families/rootkits/chkrootkit/utils"
	"github.com/openclarity/vmclarity/pkg/shared/families/rootkits/common"
)

func Test_toResultsRootkits(t *testing.T) {
	chkrootkitOutput, err := os.ReadFile("utils/testdata/chkrootkit_output.txt")
	assert.NilError(t, err)
	rootkits, err := chkrootkitutils.ParseChkrootkitOutput(chkrootkitOutput)
	assert.NilError(t, err)

	type args struct {
		rootkits []chkrootkitutils.Rootkit
	}
	tests := []struct {
		name string
		args args
		want []common.Rootkit
	}{
		{
			name: "sanity",
			args: args{
				rootkits: rootkits,
			},
			want: []common.Rootkit{
				{
					Message:     "/usr/lib/debug/usr/.dwz /usr/lib/debug/.dwz /usr/lib/debug/.build-id /usr/lib/.build-id /usr/lib/modules/6.1.21-1.45.amzn2023.x86_64/.vmlinuz.hmac /usr/lib/modules/6.1.21-1.45.amzn2023.x86_64/vdso/.build-id /usr/lib/python3.9/site-packages/awscli/botocore/.changes\n/usr/lib/debug/.dwz /usr/lib/debug/.build-id /usr/lib/.build-id /usr/lib/modules/6.1.21-1.45.amzn2023.x86_64/vdso/.build-id /usr/lib/python3.9/site-packages/awscli/botocore/.changes",
					RootkitName: "suspicious files and dirs",
					RootkitType: "UNKNOWN",
				},
				{
					Message:     "Warning: Possible Showtee Rootkit installed",
					RootkitName: "Showtee",
					RootkitType: "UNKNOWN",
				},
				{
					Message:     "/usr/include/file.h /usr/include/proc.h /usr/include/addr.h /usr/include/syslogs.h",
					RootkitName: "Romanian rootkit",
					RootkitType: "UNKNOWN",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toResultsRootkits(tt.args.rootkits)
			if diff := cmp.Diff(tt.want, got, cmpopts.SortSlices(func(a, b common.Rootkit) bool { return a.RootkitName < b.RootkitName })); diff != "" {
				t.Errorf("toResultsRootkits() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_filterResults(t *testing.T) {
	type args struct {
		rootkits []chkrootkitutils.Rootkit
	}
	tests := []struct {
		name string
		args args
		want []chkrootkitutils.Rootkit
	}{
		{
			name: "shouldn't filter",
			args: args{
				rootkits: []chkrootkitutils.Rootkit{
					{
						RkType:   "test-type",
						RkName:   "test-name1",
						Message:  "test-message1",
						Infected: true,
					},
					{
						RkType:   "test-type",
						RkName:   "test-name2",
						Message:  "test-message2",
						Infected: false,
					},
				},
			},
			want: []chkrootkitutils.Rootkit{
				{
					RkType:   "test-type",
					RkName:   "test-name1",
					Message:  "test-message1",
					Infected: true,
				},
				{
					RkType:   "test-type",
					RkName:   "test-name2",
					Message:  "test-message2",
					Infected: false,
				},
			},
		},
		{
			name: "filter out suspicious files and dirs",
			args: args{
				rootkits: []chkrootkitutils.Rootkit{
					{
						RkType:   "test-type",
						RkName:   "test-name1",
						Message:  "test-message1",
						Infected: true,
					},
					{
						RkType:   "test-type",
						RkName:   "suspicious files and dirs",
						Message:  "test-message2",
						Infected: true,
					},
				},
			},
			want: []chkrootkitutils.Rootkit{
				{
					RkType:   "test-type",
					RkName:   "test-name1",
					Message:  "test-message1",
					Infected: true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterResults(tt.args.rootkits); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterResults() = %v, want %v", got, tt.want)
			}
		})
	}
}
