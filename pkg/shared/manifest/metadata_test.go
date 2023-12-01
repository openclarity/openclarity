// Copyright © 2023 Cisco Systems, Inc. and its affiliates.
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

package manifest

import (
	_ "embed"
	"testing"

	. "github.com/onsi/gomega"
)

//go:embed testdata/bundle.json
var TestBundleJSON []byte

func TestMetadata(t *testing.T) {
	tests := []struct {
		Name             string
		Data             []byte
		ExpectedMetadata *Metadata
	}{
		{
			Name: "Metadata from valid bundle.json",
			Data: TestBundleJSON,
			ExpectedMetadata: &Metadata{
				Name:        "test-bundle",
				Version:     "1.2.3",
				Description: "Test bundle",
				License:     "Apache-2.0",
				Maintainers: []Maintainer{
					{
						Name:  "John Doe",
						Email: "john.doe@example.com",
						URL:   "www.example.com",
					},
					{
						Name:  "Jane Doe",
						Email: "jane.doe@example.com",
						URL:   "www.example.com",
					},
				},
				Parameters: map[string]interface{}{
					"key1": "value1",
					"key2": map[string]interface{}{
						"key3": float64(100),
						"key4": []interface{}{
							"value41",
							"value42",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			meta, err := NewMetadataFromRawBytes(test.Data)
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(meta).Should(BeEquivalentTo(test.ExpectedMetadata))
		})
	}
}
