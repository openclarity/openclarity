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

package common

// Results for now will be as the gitleaks results struct since it is our only secret scanner.
// once another secret scanner is integrated, we will need to think of a common scheme.
type Results struct {
	Findings    []Findings
	Source      string
	ScannerName string
	Error       error
}

type Findings struct {
	Description string `json:"Description"`
	StartLine   int    `json:"StartLine"`
	EndLine     int    `json:"EndLine"`
	StartColumn int    `json:"StartColumn"`
	EndColumn   int    `json:"EndColumn"`

	Line string `json:"Line"`

	Match string `json:"Match"`

	// Secret contains the full content of what is matched in
	// the tree-sitter query.
	Secret string `json:"Secret"`

	// File is the name of the file containing the finding
	File        string `json:"File"`
	SymlinkFile string `json:"SymlinkFile"`
	Commit      string `json:"Commit"`

	// Entropy is the shannon entropy of Value
	Entropy float32 `json:"Entropy"`

	Author  string   `json:"Author"`
	Email   string   `json:"Email"`
	Date    string   `json:"Date"`
	Message string   `json:"Message"`
	Tags    []string `json:"Tags"`

	// Rule is the name of the rule that was matched
	RuleID string `json:"RuleID"`

	// unique identifier
	Fingerprint string `json:"Fingerprint"`
}

func (r *Results) GetError() error {
	return r.Error
}
