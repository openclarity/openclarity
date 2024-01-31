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

package utils

import (
	"embed"
	"fmt"
	"io/fs"
	"testing"

	. "github.com/onsi/gomega"
)

//go:embed testdata/fs
var TestFSData embed.FS

func TestExportFS(t *testing.T) {
	tests := []struct {
		Name                string
		FS                  fs.ReadDirFS
		ExpectedContentList map[string]struct{}
	}{
		{
			Name: "With testdata FS",
			FS:   TestFSData,
			ExpectedContentList: map[string]struct{}{
				".":                            {},
				"testdata":                     {},
				"testdata/fs":                  {},
				"testdata/fs/dir01":            {},
				"testdata/fs/dir01/file01.txt": {},
				"testdata/fs/file02.txt":       {},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			dir := t.TempDir()
			t.Logf("directory: %s", dir)

			t.Log("export FS content to directory...")
			err := ExportFS(test.FS, dir)
			g.Expect(err).ToNot(HaveOccurred())

			t.Log("verify exported content...")
			err = fs.WalkDir(test.FS, ".", verifyContent(test.ExpectedContentList))
			g.Expect(err).ToNot(HaveOccurred())
		})
	}
}

func verifyContent(content map[string]struct{}) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		_, ok := content[path]
		if ok {
			return nil
		} else {
			return fmt.Errorf("unexpected content at %s: %v", path, d)
		}
	}
}
