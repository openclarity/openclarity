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

package table

import (
	"bytes"
	"os"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
	"gotest.tools/assert"

	"github.com/openclarity/kubeclarity/shared/pkg/scanner"
)

func TestTablePresenter(t *testing.T) {
	mergedResults := scanner.NewMergedResults()
	mergedResults.MergedVulnerabilitiesByKey = map[scanner.VulnerabilityKey][]scanner.MergedVulnerability{
		"cve-1.pkg-name-1.pkg-ver-1": {
			{
				ID: "1",
				Vulnerability: scanner.Vulnerability{
					ID: "cve-1",
					Fix: scanner.Fix{
						Versions: []string{"fix1", "fix2"},
					},
					Severity: "CRITICAL",
					Package: scanner.Package{
						Name:    "pkg-name-1",
						Version: "pkg-ver-1",
					},
				},
				ScannersInfo: []scanner.Info{
					{
						Name: "scanner-1",
					},
					{
						Name: "scanner-2",
					},
				},
				Diffs: nil,
			},
			{
				ID: "2",
				Vulnerability: scanner.Vulnerability{
					ID: "cve-1",
					Fix: scanner.Fix{
						Versions: []string{"fix1", "fix2"},
					},
					Severity: "HIGH",
					Package: scanner.Package{
						Name:    "pkg-name-1",
						Version: "pkg-ver-1",
					},
				},
				ScannersInfo: []scanner.Info{
					{
						Name: "scanner-3",
					},
					{
						Name: "scanner-4",
					},
				},
				Diffs: []scanner.DiffInfo{
					{
						CompareToID: "1",
						JSONDiff: map[string]interface{}{
							"severity": []interface{}{"CRITICAL", "HIGH"},
						},
					},
				},
			},
		},
		"cve-2.pkg-name-2.pkg-ver-2": { // simulate second finding with higher severity
			{
				ID: "1",
				Vulnerability: scanner.Vulnerability{
					ID:       "cve-2",
					Severity: "LOW",
					Package: scanner.Package{
						Name:    "pkg-name-2",
						Version: "pkg-ver-2",
					},
				},
				ScannersInfo: []scanner.Info{
					{
						Name: "scanner-1",
					},
				},
				Diffs: nil,
			},
			{
				ID: "2",
				Vulnerability: scanner.Vulnerability{
					ID:       "cve-2",
					Severity: "HIGH",
					Package: scanner.Package{
						Name:    "pkg-name-2",
						Version: "pkg-ver-2",
					},
				},
				ScannersInfo: []scanner.Info{
					{
						Name: "scanner-2",
					},
				},
				Diffs: []scanner.DiffInfo{
					{
						CompareToID: "1",
						JSONDiff: map[string]interface{}{
							"severity": []interface{}{"LOW", "HIGH"},
						},
					},
				},
			},
		},
		"cve-3.pkg-name-3.pkg-ver-3": { // simulate finding with no diffs
			{
				ID: "1",
				Vulnerability: scanner.Vulnerability{
					ID:       "cve-3",
					Severity: "MEDIUM",
					Package: scanner.Package{
						Name:    "pkg-name-3",
						Version: "pkg-ver-3",
					},
				},
				ScannersInfo: []scanner.Info{
					{
						Name: "scanner-1",
					},
					{
						Name: "scanner-4",
					},
				},
			},
		},
	}

	pres := NewPresenter(mergedResults)

	// run presenter
	var buffer bytes.Buffer
	err := pres.Present(&buffer)
	assert.NilError(t, err)
	actual := buffer.Bytes()
	expected, err := os.ReadFile("test_data/results.txt")
	assert.NilError(t, err)
	// err = os.WriteFile("test_data/results.txt", actual, 0666)
	// assert.NilError(t, err)

	if !bytes.Equal(expected, actual) {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(expected), string(actual), true)
		t.Errorf("mismatched output:\n%s", dmp.DiffPrettyText(diffs))
	}
}

func TestEmptyTablePresenter(t *testing.T) {
	// Expected to have no output
	var buffer bytes.Buffer

	mergedResults := scanner.NewMergedResults()

	pres := NewPresenter(mergedResults)

	// run presenter
	err := pres.Present(&buffer)
	assert.NilError(t, err)
	actual := buffer.Bytes()
	expected, err := os.ReadFile("test_data/empty_results.txt")
	assert.NilError(t, err)

	if !bytes.Equal(expected, actual) {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(expected), string(actual), true)
		t.Errorf("mismatched output:\n%s", dmp.DiffPrettyText(diffs))
	}
}
