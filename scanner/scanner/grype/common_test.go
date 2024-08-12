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

	"github.com/openclarity/vmclarity/scanner/scanner"
)

func TestCreateResults(t *testing.T) {
	// read input document
	var doc models.Document
	file, err := os.ReadFile("./test_data/nginx.json")
	assert.NilError(t, err)
	assert.NilError(t, json.Unmarshal(file, &doc))

	// set ImageMetadata type as Target
	marshalTarget, err := json.Marshal(doc.Source.Target)
	assert.NilError(t, err)
	var imageSourceMetadata syft_source.ImageMetadata
	assert.NilError(t, json.Unmarshal(marshalTarget, &imageSourceMetadata))
	doc.Source.Target = imageSourceMetadata

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
			got := CreateResults(tt.args.doc, tt.args.userInput, tt.args.scannerName, tt.args.hash, nil)
			got.Source.Metadata = nil // ignore metadata
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

	// set ImageMetadata type as Target
	marshalTarget, err := json.Marshal(doc.Source.Target)
	assert.NilError(t, err)
	var imageSourceMetadata syft_source.ImageMetadata
	assert.NilError(t, json.Unmarshal(marshalTarget, &imageSourceMetadata))
	doc.Source.Target = imageSourceMetadata

	// make a copies of document
	var sbomDoc models.Document
	if err := copier.Copy(&sbomDoc, &doc); err != nil {
		t.Errorf("failed to copy document struct: %v", err)
	}
	var otherDoc models.Document
	if err := copier.Copy(&otherDoc, &doc); err != nil {
		t.Errorf("failed to copy document struct: %v", err)
	}

	// empty imageMetadata for SBOM input
	sbomDoc.Source.Target = syft_source.ImageMetadata{}
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
				Hash: "644a70516a26004c97d0d85c7fe1d0c3a67ea8ab7ddf4aff193d9f301670cf36",
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
			got := getSource(tt.args.doc, tt.args.userInput, tt.args.hash, nil)
			got.Metadata = nil // ignore metadata

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSource() = %v, want %v", got, tt.want)
			}
		})
	}
}
