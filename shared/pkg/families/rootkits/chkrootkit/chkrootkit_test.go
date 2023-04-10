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
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"gotest.tools/v3/assert"

	chkrootkitutils "github.com/openclarity/vmclarity/shared/pkg/families/rootkits/chkrootkit/utils"
	"github.com/openclarity/vmclarity/shared/pkg/families/rootkits/common"
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
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toResultsRootkits(tt.args.rootkits)
			//if diff := cmp.Diff(tt.want, got); diff != "" {
			//	t.Errorf("toResultsRootkits() mismatch (-want +got):\n%s", diff)
			//}
			t.Logf("toResultsRootkits() results: %+v", prettyPrint(t, got))
		})
	}
}

func prettyPrint(t *testing.T, got any) string {
	t.Helper()
	jsonResults, err := json.MarshalIndent(got, "", "    ")
	assert.NilError(t, err)
	return string(jsonResults)
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
