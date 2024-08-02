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

package fake

import (
	"context"

	"github.com/openclarity/vmclarity/scanner/common"
	"github.com/openclarity/vmclarity/scanner/families"
	"github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
)

const ScannerName = "fake"

type Scanner struct{}

func New(_ context.Context, _ string, _ types.ScannersConfig) (families.Scanner[*types.ScannerResult], error) {
	return &Scanner{}, nil
}

func (a *Scanner) Scan(_ context.Context, _ common.InputType, _ string) (*types.ScannerResult, error) {
	misconfigurations := createFakeMisconfigurationReport()

	return types.NewScannerResult(ScannerName, misconfigurations), nil
}

func createFakeMisconfigurationReport() []types.Misconfiguration {
	return []types.Misconfiguration{
		{
			Location: "/fake",

			Category:    "FAKE",
			ID:          "Test1",
			Description: "Fake test number 1",

			Message:     "Fake test number 1 failed",
			Severity:    types.HighSeverity,
			Remediation: "fix the thing number 1",
		},
		{
			Location: "/fake",

			Category:    "FAKE",
			ID:          "Test2",
			Description: "Fake test number 2",

			Message:     "Fake test number 2 failed",
			Severity:    types.LowSeverity,
			Remediation: "fix the thing number 2",
		},
		{
			Location: "/fake",

			Category:    "FAKE",
			ID:          "Test3",
			Description: "Fake test number 3",

			Message:     "Fake test number 3 failed",
			Severity:    types.MediumSeverity,
			Remediation: "fix the thing number 3",
		},
		{
			Location: "/fake",

			Category:    "FAKE",
			ID:          "Test4",
			Description: "Fake test number 4",

			Message:     "Fake test number 4 failed",
			Severity:    types.HighSeverity,
			Remediation: "fix the thing number 4",
		},
	}
}
