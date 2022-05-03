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

package grype

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/anchore/grype/grype/presenter/models"
	"gotest.tools/assert"

	"github.com/openclarity/kubeclarity/shared/pkg/scanner"
)

func TestCreateResults(t *testing.T) {
	// read input document
	var doc models.Document
	file, err := os.ReadFile("./test_data/nginx.json")
	assert.NilError(t, err)
	assert.NilError(t, json.Unmarshal(file, &doc))

	// read expected results
	var results scanner.Results
	file, err = os.ReadFile("./test_data/nginx.results.json")
	assert.NilError(t, err)
	assert.NilError(t, json.Unmarshal(file, &results))

	type args struct {
		doc         models.Document
		userInput   string
		scannerName string
	}
	tests := []struct {
		name string
		args args
		want *scanner.Results
	}{
		{
			name: "sanity",
			args: args{
				doc:         doc,
				userInput:   "nginx",
				scannerName: "grype",
			},
			want: &results,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateResults(tt.args.doc, tt.args.userInput, tt.args.scannerName)
			// gotB, _ := json.Marshal(got)
			// assert.NilError(t, os.WriteFile("./test_data/nginx.results.json", gotB, 0666))
			assert.DeepEqual(t, got, tt.want)
		})
	}
}
