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
	"embed"
	"fmt"
	"io/fs"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

//go:embed all:testdata
var TestBundleFS embed.FS

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

func TestOpen(t *testing.T) {
	tests := []struct {
		Name          string
		BundleFS      FS
		Prefix        string
		Path          string
		ExpectedName  string
		ExpectedIsDir types.GomegaMatcher
	}{
		{
			Name:          "File with no prefix",
			BundleFS:      TestBundleFS,
			Prefix:        "",
			Path:          "testdata/file02.txt",
			ExpectedName:  "file02.txt",
			ExpectedIsDir: BeFalse(),
		},
		{
			Name:          "Dir with no prefix",
			BundleFS:      TestBundleFS,
			Prefix:        "",
			Path:          "testdata/dir01",
			ExpectedName:  "dir01",
			ExpectedIsDir: BeTrue(),
		},
		{
			Name:          "File with prefix",
			BundleFS:      TestBundleFS,
			Prefix:        "testdata",
			Path:          "file02.txt",
			ExpectedName:  "file02.txt",
			ExpectedIsDir: BeFalse(),
		},
		{
			Name:          "Dir with prefix",
			BundleFS:      TestBundleFS,
			Prefix:        "testdata",
			Path:          "dir01",
			ExpectedName:  "dir01",
			ExpectedIsDir: BeTrue(),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			bundle, err := New(test.BundleFS, WithPrefix(test.Prefix))
			g.Expect(err).ToNot(HaveOccurred())
			info, err := bundle.Open(test.Path)
			g.Expect(err).ToNot(HaveOccurred())
			if err == nil {
				stat, err := info.Stat()
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stat.Name()).Should(BeEquivalentTo(test.ExpectedName))
				g.Expect(stat.IsDir()).Should(test.ExpectedIsDir)
			}
		})
	}
}

func TestReadDir(t *testing.T) {
	tests := []struct {
		Name                string
		BundleFS            FS
		Prefix              string
		Path                string
		ExpectedContentList map[string]struct{}
	}{
		{
			Name:     "With no prefix",
			BundleFS: TestBundleFS,
			Prefix:   "",
			Path:     "testdata",
			ExpectedContentList: map[string]struct{}{
				"dir01":       {},
				"file02.txt":  {},
				"bundle.json": {},
			},
		},
		{
			Name:     "With prefix",
			BundleFS: TestBundleFS,
			Prefix:   "testdata",
			Path:     "dir01",
			ExpectedContentList: map[string]struct{}{
				"file01.txt": {},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			bundle, err := New(test.BundleFS, WithPrefix(test.Prefix))
			g.Expect(err).ToNot(HaveOccurred())
			entries, err := bundle.ReadDir(test.Path)
			g.Expect(err).ToNot(HaveOccurred())
			for _, entry := range entries {
				_, ok := test.ExpectedContentList[entry.Name()]
				g.Expect(ok).Should(BeTrue())
			}
		})
	}
}

func TestWalkDir(t *testing.T) {
	tests := []struct {
		Name                string
		BundleFS            FS
		Prefix              string
		ExpectedContentList map[string]struct{}
	}{
		{
			Name:     "Bundle with embed.FS with no prefix",
			BundleFS: TestBundleFS,
			Prefix:   "",
			ExpectedContentList: map[string]struct{}{
				".":                         {},
				"testdata":                  {},
				"testdata/dir01":            {},
				"testdata/dir01/file01.txt": {},
				"testdata/file02.txt":       {},
				"testdata/bundle.json":      {},
			},
		},
		{
			Name:     "Bundle with embed.FS with prefix",
			BundleFS: TestBundleFS,
			Prefix:   "testdata",
			ExpectedContentList: map[string]struct{}{
				".":                {},
				"dir01":            {},
				"dir01/file01.txt": {},
				"file02.txt":       {},
				"bundle.json":      {},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			bundle, err := New(test.BundleFS, WithPrefix(test.Prefix))
			g.Expect(err).ToNot(HaveOccurred())
			err = bundle.WalkDir(".", verifyContent(test.ExpectedContentList))
			g.Expect(err).ToNot(HaveOccurred())
		})
	}
}

func TestGlob(t *testing.T) {
	tests := []struct {
		Name            string
		BundleFS        FS
		Prefix          string
		Pattern         string
		ExpectedMatches []string
	}{
		{
			Name:     "Default PatternMatcher with wildcard pattern and no prefix",
			BundleFS: TestBundleFS,
			Prefix:   "",
			Pattern:  "testdata/*.txt",
			ExpectedMatches: []string{
				"testdata/file02.txt",
			},
		},
		{
			Name:     "Default PatternMatcher with wildcard pattern and prefix",
			BundleFS: TestBundleFS,
			Prefix:   "testdata",
			Pattern:  "*.txt",
			ExpectedMatches: []string{
				"file02.txt",
			},
		},
		{
			Name:     "Default PatternMatcher with exact pattern and no prefix",
			BundleFS: TestBundleFS,
			Prefix:   "",
			Pattern:  "testdata/dir01/file01.txt",
			ExpectedMatches: []string{
				"testdata/dir01/file01.txt",
			},
		},
		{
			Name:     "Default PatternMatcher with exact pattern and prefix",
			BundleFS: TestBundleFS,
			Prefix:   "testdata",
			Pattern:  "dir01/file01.txt",
			ExpectedMatches: []string{
				"dir01/file01.txt",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			bundle, err := New(test.BundleFS, WithPrefix(test.Prefix))
			g.Expect(err).ToNot(HaveOccurred())
			matches, err := bundle.Glob(test.Pattern)
			g.Expect(err).ToNot(HaveOccurred())
			if err == nil {
				g.Expect(matches).Should(BeEquivalentTo(test.ExpectedMatches))
			}
		})
	}
}

func TestStat(t *testing.T) {
	tests := []struct {
		Name                 string
		BundleFS             FS
		Prefix               string
		Path                 string
		ExpectedErrorMatcher types.GomegaMatcher
		ExpectedName         string
		ExpectedIsDirMatcher types.GomegaMatcher
	}{
		{
			Name:                 "Valid dir path and no prefix",
			BundleFS:             TestBundleFS,
			Prefix:               "",
			Path:                 "testdata/dir01",
			ExpectedErrorMatcher: Not(HaveOccurred()),
			ExpectedName:         "dir01",
			ExpectedIsDirMatcher: BeTrue(),
		},
		{
			Name:                 "Valid file path and no prefix",
			BundleFS:             TestBundleFS,
			Prefix:               "",
			Path:                 "testdata/dir01/file01.txt",
			ExpectedErrorMatcher: Not(HaveOccurred()),
			ExpectedName:         "file01.txt",
			ExpectedIsDirMatcher: BeFalse(),
		},
		{
			Name:                 "Invalid path and no prefix",
			BundleFS:             TestBundleFS,
			Prefix:               "",
			Path:                 "testdata/invalid/path",
			ExpectedErrorMatcher: HaveOccurred(),
		},
		{
			Name:                 "Valid dir path with prefix",
			BundleFS:             TestBundleFS,
			Prefix:               "testdata",
			Path:                 "dir01",
			ExpectedErrorMatcher: Not(HaveOccurred()),
			ExpectedName:         "dir01",
			ExpectedIsDirMatcher: BeTrue(),
		},
		{
			Name:                 "Valid file path with prefix",
			BundleFS:             TestBundleFS,
			Prefix:               "testdata",
			Path:                 "dir01/file01.txt",
			ExpectedErrorMatcher: Not(HaveOccurred()),
			ExpectedName:         "file01.txt",
			ExpectedIsDirMatcher: BeFalse(),
		},
		{
			Name:                 "Invalid path with prefix",
			BundleFS:             TestBundleFS,
			Prefix:               "testdata",
			Path:                 "invalid/path",
			ExpectedErrorMatcher: HaveOccurred(),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			bundle, err := New(test.BundleFS, WithPrefix(test.Prefix))
			g.Expect(err).ToNot(HaveOccurred())
			info, err := bundle.Stat(test.Path)
			g.Expect(err).Should(test.ExpectedErrorMatcher)
			if err == nil {
				g.Expect(info.Name()).Should(BeEquivalentTo(test.ExpectedName))
				g.Expect(info.IsDir()).Should(test.ExpectedIsDirMatcher)
			}
		})
	}
}

func TestSubFS(t *testing.T) {
	tests := []struct {
		Name                 string
		BundleFS             FS
		Prefix               string
		SubDir               string
		ExpectedErrorMatcher types.GomegaMatcher
		ExpectedContentList  map[string]struct{}
	}{
		{
			Name:                 "Valid sub-dir and no prefix",
			BundleFS:             TestBundleFS,
			Prefix:               "",
			SubDir:               "testdata/dir01",
			ExpectedErrorMatcher: Not(HaveOccurred()),
			ExpectedContentList: map[string]struct{}{
				".":          {},
				"file01.txt": {},
			},
		},
		{
			Name:                 "Invalid sub-dir and no prefix",
			BundleFS:             TestBundleFS,
			Prefix:               "",
			SubDir:               "testdata/invalid/path",
			ExpectedErrorMatcher: HaveOccurred(),
		},
		{
			Name:                 "Valid sub-dir with prefix",
			BundleFS:             TestBundleFS,
			Prefix:               "testdata",
			SubDir:               "dir01",
			ExpectedErrorMatcher: Not(HaveOccurred()),
			ExpectedContentList: map[string]struct{}{
				".":          {},
				"file01.txt": {},
			},
		},
		{
			Name:                 "Invalid sub-dir with prefix",
			BundleFS:             TestBundleFS,
			Prefix:               "testdata",
			SubDir:               "invalid/path",
			ExpectedErrorMatcher: HaveOccurred(),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			bundle, err := New(test.BundleFS, WithPrefix(test.Prefix))
			g.Expect(err).ToNot(HaveOccurred())
			subBundle, err := bundle.Sub(test.SubDir)
			g.Expect(err).Should(test.ExpectedErrorMatcher)
			if err == nil {
				err = subBundle.WalkDir(".", verifyContent(test.ExpectedContentList))
				g.Expect(err).ToNot(HaveOccurred())
			}
		})
	}
}
