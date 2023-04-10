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

package utils

import (
	"encoding/json"
	"os"
	"testing"

	"gotest.tools/v3/assert"
)

func TestParseChkrootkitOutput(t *testing.T) {
	chkrootkitOutput, err := os.ReadFile("testdata/chkrootkit_output.txt")
	assert.NilError(t, err)

	type args struct {
		chkrootkitOutput []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []Rootkit
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				chkrootkitOutput: chkrootkitOutput,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseChkrootkitOutput(tt.args.chkrootkitOutput)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseChkrootkitOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if diff := cmp.Diff(tt.want, got); diff != "" {
			//	t.Errorf("ParseChkrootkitOutput() mismatch (-want +got):\n%s", diff)
			//}
			t.Logf("ParseChkrootkitOutput() result: %v", prettyPrint(t, got))
		})
	}
}

func prettyPrint(t *testing.T, got any) string {
	t.Helper()
	jsonResults, err := json.MarshalIndent(got, "", "    ")
	assert.NilError(t, err)
	return string(jsonResults)
}
