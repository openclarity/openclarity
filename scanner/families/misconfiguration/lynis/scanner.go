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
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
	familiesutils "github.com/openclarity/vmclarity/scanner/families/utils"
	"github.com/openclarity/vmclarity/scanner/job_manager"
	"github.com/openclarity/vmclarity/scanner/utils"
)

const (
	ScannerName = "lynis"
	LynisBinary = "lynis"
)

type Scanner struct {
	name       string
	logger     *log.Entry
	config     types.LynisConfig
	resultChan chan job_manager.Result
}

func New(_ string, c job_manager.IsConfig, logger *log.Entry, resultChan chan job_manager.Result) job_manager.Job {
	conf := c.(types.ScannersConfig) // nolint:forcetypeassert
	return &Scanner{
		name:       ScannerName,
		logger:     logger.Dup().WithField("scanner", ScannerName),
		config:     conf.Lynis,
		resultChan: resultChan,
	}
}

// nolint: cyclop
func (a *Scanner) Run(ctx context.Context, sourceType utils.SourceType, userInput string) error {
	go func(ctx context.Context) {
		retResults := types.ScannerResult{
			ScannerName: ScannerName,
		}

		// Validate this is an input type supported by the scanner,
		// otherwise return skipped.
		if !a.isValidInputType(sourceType) {
			a.sendResults(retResults, nil)
			return
		}

		// Locate lynis binary
		if a.config.BinaryPath == "" {
			a.config.BinaryPath = LynisBinary
		}

		lynisBinaryPath, err := exec.LookPath(a.config.BinaryPath)
		if err != nil {
			a.sendResults(retResults, fmt.Errorf("failed to lookup executable %s: %w", a.config.BinaryPath, err))
			return
		}
		a.logger.Debugf("found lynis binary at: %s", lynisBinaryPath)

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

		fsPath, cleanup, err := familiesutils.ConvertInputToFilesystem(ctx, sourceType, userInput)
		if err != nil {
			a.sendResults(retResults, fmt.Errorf("failed to convert input to filesystem: %w", err))
			return
		}
		defer cleanup()

		// Build command:
		// lynis audit system \
		//     --report-file <reportDir>/report.dat \
		//     --no-log \
		//     --forensics \
		//     --rootdir <source>
		args := []string{
			"audit",
			"system",
			"--report-file",
			reportPath,
			"--no-log",
			"--forensics",
			"--tests",
			strings.Join(testsToRun, ","),
			"--rootdir",
			fsPath,
		}
		cmd := exec.Command(lynisBinaryPath, args...) // nolint:gosec

		a.logger.Infof("Running command: %v", cmd.String())
		_, err = utils.RunCommand(cmd)
		if err != nil {
			a.sendResults(retResults, fmt.Errorf("failed to run command: %w", err))
			return
		}

		// Get Lynis DB directory
		cmd = exec.Command(lynisBinaryPath, []string{"show", "dbdir"}...) // nolint:gosec
		out, err := utils.RunCommand(cmd)
		if err != nil {
			a.sendResults(retResults, fmt.Errorf("failed to run command: %w", err))
			return
		}
		lynisDBDir := filepath.Clean(strings.TrimSpace(string(out)))

		testDB, err := NewTestDB(a.logger, lynisDBDir)
		if err != nil {
			a.sendResults(retResults, fmt.Errorf("failed to load lynis test DB: %w", err))
			return
		}

		reportParser := NewReportParser(testDB)
		retResults.Misconfigurations, err = reportParser.ParseLynisReport(userInput, reportPath)
		if err != nil {
			a.sendResults(retResults, fmt.Errorf("failed to parse report file %v: %w", reportPath, err))
			return
		}

		a.sendResults(retResults, nil)
	}(ctx)

	return nil
}

func (a *Scanner) isValidInputType(sourceType utils.SourceType) bool {
	switch sourceType {
	case utils.ROOTFS, utils.IMAGE, utils.DOCKERARCHIVE, utils.OCIARCHIVE, utils.OCIDIR:
		return true
	case utils.DIR, utils.FILE, utils.SBOM:
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
