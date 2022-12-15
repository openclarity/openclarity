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
	Description string `json:"description"`
	StartLine   int    `json:"start_line"`
	EndLine     int    `json:"end_line"`
	StartColumn int    `json:"start_column"`
	EndColumn   int    `json:"end_column"`

	Line string `json:"-" json:"line"`

	Match string `json:"match"`

	// Secret contains the full content of what is matched in
	// the tree-sitter query.
	Secret string `json:"secret"`

	// File is the name of the file containing the finding
	File        string `json:"file"`
	SymlinkFile string `json:"symlink_file"`
	Commit      string `json:"commit"`

	// Entropy is the shannon entropy of Value
	Entropy float32 `json:"entropy"`

	Author  string   `json:"author"`
	Email   string   `json:"email"`
	Date    string   `json:"date"`
	Message string   `json:"message"`
	Tags    []string `json:"tags"`

	// Rule is the name of the rule that was matched
	RuleID string `json:"rule_id"`

	// unique identifer
	Fingerprint string `json:"fingerprint"`
}

func (r *Results) GetError() error {
	return r.Error
}
