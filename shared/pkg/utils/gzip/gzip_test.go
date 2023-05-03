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

package gzip

import (
	"os"
	"testing"

	"gotest.tools/assert"
)

func TestCompress(t *testing.T) {
	file, err := os.ReadFile("testdata/examples-bookinfo-reviews-v1-1.17.0.cdx.json")
	assert.NilError(t, err)
	t.Logf("file len %v", len(file))
	compress, err := CompressAndEncode(file)
	assert.NilError(t, err)
	t.Logf("compress file len %v", len(compress))
	t.Logf("compress by %v", float64(len(compress))/float64(len(file))*100)
	fileB, err := DecodeAndUncompress(compress)
	assert.NilError(t, err)
	t.Logf("fileB len %v", len(fileB))
	assert.DeepEqual(t, file, fileB)
}
