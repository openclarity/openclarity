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

	log "github.com/sirupsen/logrus"

	misconfigurationTypes "github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
	"github.com/openclarity/vmclarity/scanner/job_manager"
	"github.com/openclarity/vmclarity/scanner/utils"
)

const ScannerName = "fake"

type Scanner struct {
	name       string
	logger     *log.Entry
	resultChan chan job_manager.Result
}

func New(_ string, _ job_manager.IsConfig, logger *log.Entry, resultChan chan job_manager.Result) job_manager.Job {
	return &Scanner{
		name:       ScannerName,
		logger:     logger.Dup().WithField("scanner", ScannerName),
		resultChan: resultChan,
	}
}

func (a *Scanner) Run(ctx context.Context, sourceType utils.SourceType, userInput string) error {
	go func() {
		retResults := misconfigurationTypes.ScannerResult{
			ScannerName:       ScannerName,
			Misconfigurations: createFakeMisconfigurationReport(),
		}

		a.sendResults(retResults, nil)
	}()

	return nil
}

func createFakeMisconfigurationReport() []misconfigurationTypes.Misconfiguration {
	return []misconfigurationTypes.Misconfiguration{
		{
			Location: "/fake",

			Category:    "FAKE",
			ID:          "Test1",
			Description: "Fake test number 1",

			Message:     "Fake test number 1 failed",
			Severity:    misconfigurationTypes.HighSeverity,
			Remediation: "fix the thing number 1",
		},
		{
			Location: "/fake",

			Category:    "FAKE",
			ID:          "Test2",
			Description: "Fake test number 2",

			Message:     "Fake test number 2 failed",
			Severity:    misconfigurationTypes.LowSeverity,
			Remediation: "fix the thing number 2",
		},
		{
			Location: "/fake",

			Category:    "FAKE",
			ID:          "Test3",
			Description: "Fake test number 3",

			Message:     "Fake test number 3 failed",
			Severity:    misconfigurationTypes.MediumSeverity,
			Remediation: "fix the thing number 3",
		},
		{
			Location: "/fake",

			Category:    "FAKE",
			ID:          "Test4",
			Description: "Fake test number 4",

			Message:     "Fake test number 4 failed",
			Severity:    misconfigurationTypes.HighSeverity,
			Remediation: "fix the thing number 4",
		},
	}
}

func (a *Scanner) sendResults(results misconfigurationTypes.ScannerResult, err error) {
	if err != nil {
		a.logger.Error(err)
		results.Error = err
	}
	select {
	case a.resultChan <- results:
	default:
		a.logger.Error("Failed to send results on channel")
	}
}
