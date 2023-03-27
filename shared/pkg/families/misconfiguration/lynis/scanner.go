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

package lynis

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
	"github.com/openclarity/kubeclarity/shared/pkg/utils"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/shared/pkg/families/misconfiguration/types"
	sharedUtils "github.com/openclarity/vmclarity/shared/pkg/utils"
)

const ScannerName = "lynis"

type Scanner struct {
	name       string
	logger     *log.Entry
	config     types.LynisConfig
	resultChan chan job_manager.Result
}

func New(c job_manager.IsConfig, logger *log.Entry, resultChan chan job_manager.Result) job_manager.Job {
	conf := c.(types.ScannersConfig) // nolint:forcetypeassert
	return &Scanner{
		name:       ScannerName,
		logger:     logger.Dup().WithField("scanner", ScannerName),
		config:     conf.Lynis,
		resultChan: resultChan,
	}
}

func (a *Scanner) Run(sourceType utils.SourceType, userInput string) error {
	go func() {
		retResults := types.ScannerResult{
			ScannerName: ScannerName,
		}

		// Validate this is an input type supported by the scanner,
		// otherwise return skipped.
		if !a.isValidInputType(sourceType) {
			a.sendResults(retResults, nil)
			return
		}

		// Validate that lynis exists
		lynisPath := path.Join(a.config.InstallPath, "lynis")
		if _, err := os.Stat(lynisPath); err != nil {
			a.sendResults(retResults, fmt.Errorf("failed to find lynis @ %v: %w", lynisPath, err))
			return
		}

		reportDir, err := os.MkdirTemp("", "")
		if err != nil {
			a.sendResults(retResults, fmt.Errorf("failed to create temp directory: %w", err))
			return
		}
		defer func() {
			err := os.RemoveAll(reportDir)
			if err != nil {
				a.logger.Warningf("failed to remove temp directory: %v", err)
			}
		}()

		reportPath := path.Join(reportDir, "lynis.dat")

		// Build command:
		// <installPath>/lynis audit system \
		//     --report-file <reportDir>/report.dat \
		//     --log-file /dev/null \
		//     --forensics \
		//     --rootdir <source>
		args := []string{
			"audit",
			"system",
			"--report-file",
			reportPath,
			"--log-file",
			"/dev/null",
			"--forensics",
			"--tests",
			strings.Join(testsToRun, ","),
			"--rootdir",
			userInput,
		}
		cmd := exec.Command(lynisPath, args...) // nolint:gosec

		// Lynis requires that it is executed from inside of the lynis
		// install directory. So change the working dir to the lynis
		// install path.
		cmd.Dir = a.config.InstallPath

		a.logger.Infof("Running command: %v", cmd.String())
		_, err = sharedUtils.RunCommand(cmd)
		if err != nil {
			a.sendResults(retResults, fmt.Errorf("failed to run command: %w", err))
			return
		}

		testdb, err := NewTestDB(a.logger, a.config.InstallPath)
		if err != nil {
			a.sendResults(retResults, fmt.Errorf("failed to load lynis test DB: %w", err))
			return
		}

		reportParser := NewReportParser(testdb)
		retResults.Misconfigurations, err = reportParser.ParseLynisReport(userInput, reportPath)
		if err != nil {
			a.sendResults(retResults, fmt.Errorf("failed to parse report file %v: %w", reportPath, err))
			return
		}

		a.sendResults(retResults, nil)
	}()

	return nil
}

func (a *Scanner) isValidInputType(sourceType utils.SourceType) bool {
	switch sourceType {
	case utils.ROOTFS:
		return true
	case utils.DIR, utils.FILE, utils.IMAGE, utils.SBOM:
		a.logger.Infof("source type %v is not supported for lynis, skipping.", sourceType)
	default:
		a.logger.Infof("unknown source type %v, skipping.", sourceType)
	}
	return false
}

func (a *Scanner) sendResults(results types.ScannerResult, err error) {
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
