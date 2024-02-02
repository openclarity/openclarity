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

package types

type Severity string

const (
	HighSeverity   Severity = "HighSeverity"
	MediumSeverity Severity = "MediumSeverity"
	LowSeverity    Severity = "LowSeverity"
)

type Misconfiguration struct {
	// Path which was scanned to find this Misconfiguration, might be "/",
	// "partitionX/" or "/home/Dockerfile"
	//
	// This might just be the scanner input if the tool scans it as a whole
	// or it can be a specific file if the scanner performs some
	// sub-discovery like trivy.
	ScannedPath string `json:"ScannedPath"`

	// Information about the test that was run to detect this specific
	// misconfiguration, this is specific to each Scanner.
	TestCategory    string `json:"TestCategory"`
	TestID          string `json:"TestID"`
	TestDescription string `json:"TestDescription"`

	// Information about this specific misconfiguration hit
	Severity    Severity `json:"Severity"`
	Message     string   `json:"Message"`
	Remediation string   `json:"Remediation"`
}

type ScannerResult struct {
	ScannerName       string
	Misconfigurations []Misconfiguration
	Error             error
}

func (sr ScannerResult) GetError() error {
	return sr.Error
}
