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
	"reflect"
	"testing"

	"github.com/anchore/grype/grype/presenter/models"
	syft_source "github.com/anchore/syft/syft/source"
	"github.com/jinzhu/copier"
	"gotest.tools/assert"

	"github.com/openclarity/kubeclarity/shared/pkg/scanner"
)

func TestCreateResults(t *testing.T) {
	// read input document
	var doc models.Document
	file, err := os.ReadFile("./test_data/nginx.json")
	assert.NilError(t, err)
	assert.NilError(t, json.Unmarshal(file, &doc))
	// define Target properly for image input
	doc.Source.Target = syft_source.StereoscopeImageSourceMetadata{
		UserInput:      "nginx",
		ManifestDigest: "sha256:43ef2d67f4f458c2ac373ce0abf34ff6ad61616dd7cfd2880c6381d7904b6a94",
		RepoDigests:    []string{"sha256:43ef2d67f4f458c2ac373ce0abf34ff6ad61616dd7cfd2880c6381d7904b6a94"},
	}
	// read expected results
	var results scanner.Results
	file, err = os.ReadFile("./test_data/nginx.results.json")
	assert.NilError(t, err)
	assert.NilError(t, json.Unmarshal(file, &results))

	type args struct {
		doc         models.Document
		userInput   string
		scannerName string
		hash        string
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
			got := CreateResults(tt.args.doc, tt.args.userInput, tt.args.scannerName, tt.args.hash)
			// gotB, _ := json.Marshal(got)
			// assert.NilError(t, os.WriteFile("./test_data/nginx.results.json", gotB, 0666))
			assert.DeepEqual(t, got, tt.want)
		})
	}
}

func Test_getSource(t *testing.T) {
	// read input document
	var doc models.Document
	file, err := os.ReadFile("./test_data/nginx.json")
	assert.NilError(t, err)
	assert.NilError(t, json.Unmarshal(file, &doc))

	// make a copies of document
	var sbomDoc models.Document
	if err := copier.Copy(&sbomDoc, &doc); err != nil {
		t.Errorf("failed to copy document struct: %v", err)
	}
	var otherDoc models.Document
	if err := copier.Copy(&otherDoc, &doc); err != nil {
		t.Errorf("failed to copy document struct: %v", err)
	}

	// define Target properly for image input
	doc.Source.Target = syft_source.StereoscopeImageSourceMetadata{
		UserInput:      "nginx",
		ManifestDigest: "sha256:43ef2d67f4f458c2ac373ce0abf34ff6ad61616dd7cfd2880c6381d7904b6a94",
		RepoDigests:    []string{"sha256:43ef2d67f4f458c2ac373ce0abf34ff6ad61616dd7cfd2880c6381d7904b6a94"},
	}
	// empty imageMetadata for SBOM input
	sbomDoc.Source.Target = syft_source.StereoscopeImageSourceMetadata{}
	// string for other input
	otherDoc.Source.Target = "test"

	type args struct {
		doc       models.Document
		userInput string
		hash      string
	}
	tests := []struct {
		name string
		args args
		want scanner.Source
	}{
		{
			name: "input is an image",
			args: args{
				doc:       doc,
				userInput: "nginx",
				hash:      "",
			},
			want: scanner.Source{
				Type: "image",
				Name: "nginx",
				Hash: "43ef2d67f4f458c2ac373ce0abf34ff6ad61616dd7cfd2880c6381d7904b6a94",
			},
		},
		{
			name: "input is a SBOM",
			args: args{
				doc:       sbomDoc,
				userInput: "nginx",
				hash:      "testhash",
			},
			want: scanner.Source{
				Type: "image",
				Name: "nginx",
				Hash: "testhash",
			},
		},
		{
			name: "input is not SBOM or image",
			args: args{
				doc:       otherDoc,
				userInput: "test",
				hash:      "testhash",
			},
			want: scanner.Source{
				Type: "image",
				Name: "test",
				Hash: "testhash",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSource(tt.args.doc, tt.args.userInput, tt.args.hash); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSource() = %v, want %v", got, tt.want)
			}
		})
	}
}
