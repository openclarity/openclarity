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
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
			want: []Rootkit{
				{
					RkType:   "APPLICATION",
					RkName:   "UNKNOWN",
					Message:  "Application \"amd\" not found",
					Infected: false,
				},
				{
					RkType:   "APPLICATION",
					RkName:   "UNKNOWN",
					Message:  "Application \"basename\" not infected",
					Infected: false,
				},
				{
					RkType:   "UNKNOWN",
					RkName:   "sniffer",
					Message:  "nothing found",
					Infected: false,
				},
				{
					RkType:   "UNKNOWN",
					RkName:   "suspicious files and dirs",
					Message:  "/usr/lib/debug/usr/.dwz /usr/lib/debug/.dwz /usr/lib/debug/.build-id /usr/lib/.build-id /usr/lib/modules/6.1.21-1.45.amzn2023.x86_64/.vmlinuz.hmac /usr/lib/modules/6.1.21-1.45.amzn2023.x86_64/vdso/.build-id /usr/lib/python3.9/site-packages/awscli/botocore/.changes\n/usr/lib/debug/.dwz /usr/lib/debug/.build-id /usr/lib/.build-id /usr/lib/modules/6.1.21-1.45.amzn2023.x86_64/vdso/.build-id /usr/lib/python3.9/site-packages/awscli/botocore/.changes",
					Infected: true,
				},
				{
					RkType:   "UNKNOWN",
					RkName:   "Showtee",
					Message:  "Warning: Possible Showtee Rootkit installed",
					Infected: true,
				},
				{
					RkType:   "UNKNOWN",
					RkName:   "Romanian rootkit",
					Message:  "/usr/include/file.h /usr/include/proc.h /usr/include/addr.h /usr/include/syslogs.h",
					Infected: true,
				},
				{
					RkType:   "UNKNOWN",
					RkName:   "Linux/Ebury - Operation Windigo ssh",
					Message:  "not tested",
					Infected: false,
				},
			},
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
			if diff := cmp.Diff(tt.want, got, cmpopts.SortSlices(func(a, b Rootkit) bool { return a.RkName < b.RkName })); diff != "" {
				t.Errorf("ParseChkrootkitOutput() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
