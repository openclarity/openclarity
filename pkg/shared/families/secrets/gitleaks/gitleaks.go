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

package gitleaks

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
	"github.com/openclarity/kubeclarity/shared/pkg/utils"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/pkg/shared/families/secrets/common"
	gitleaksconfig "github.com/openclarity/vmclarity/pkg/shared/families/secrets/gitleaks/config"
	sharedutils "github.com/openclarity/vmclarity/pkg/shared/utils"
)

const ScannerName = "gitleaks"

type Scanner struct {
	name       string
	logger     *log.Entry
	config     gitleaksconfig.Config
	resultChan chan job_manager.Result
}

func New(c job_manager.IsConfig, logger *log.Entry, resultChan chan job_manager.Result) job_manager.Job {
	conf := c.(*common.ScannersConfig) // nolint:forcetypeassert
	return &Scanner{
		name:       ScannerName,
		logger:     logger.Dup().WithField("scanner", ScannerName),
		config:     gitleaksconfig.Config{BinaryPath: conf.Gitleaks.BinaryPath},
		resultChan: resultChan,
	}
}

func (a *Scanner) Run(sourceType utils.SourceType, userInput string) error {
	go func() {
		retResults := common.Results{
			Source:      userInput,
			ScannerName: ScannerName,
		}
		if !a.isValidInputType(sourceType) {
			a.sendResults(retResults, nil)
			return
		}
		// validate that gitleaks binary exists
		if _, err := os.Stat(a.config.BinaryPath); err != nil {
			a.sendResults(retResults, fmt.Errorf("failed to find binary in %v: %v", a.config.BinaryPath, err))
			return
		}

		file, err := os.CreateTemp("", "gitleaks")
		if err != nil {
			a.sendResults(retResults, fmt.Errorf("failed to create temp file. %v", err))
			return
		}
		defer func() {
			_ = os.Remove(file.Name())
		}()
		reportPath := file.Name()

		// ./gitleaks detect --source=<source> --no-git -r <report-path> -f json --exit-code 0
		// nolint:gosec
		cmd := exec.Command(a.config.BinaryPath, "detect", fmt.Sprintf("--source=%v", userInput), "--no-git", "-r", reportPath, "-f", "json", "--exit-code", "0")
		a.logger.Infof("Running gitleaks command: %v", cmd.String())
		_, err = sharedutils.RunCommand(cmd)
		if err != nil {
			a.sendResults(retResults, fmt.Errorf("failed to run gitleaks command: %v", err))
			return
		}

		out, err := os.ReadFile(reportPath)
		if err != nil {
			a.sendResults(retResults, fmt.Errorf("failed to read report file from path %v: %v", reportPath, err))
			return
		}

		if err := json.Unmarshal(out, &retResults.Findings); err != nil {
			a.sendResults(retResults, fmt.Errorf("failed to unmarshal results. out: %s. err: %v", out, err))
			return
		}
		a.sendResults(retResults, nil)
	}()

	return nil
}

func (a *Scanner) isValidInputType(sourceType utils.SourceType) bool {
	switch sourceType {
	case utils.DIR, utils.ROOTFS:
		return true
	case utils.FILE, utils.IMAGE, utils.SBOM:
		fallthrough
	default:
		a.logger.Infof("source type %v is not supported for gitleaks, skipping.", sourceType)
	}
	return false
}

func (a *Scanner) sendResults(results common.Results, err error) {
	if err != nil {
		a.logger.Error(err)
		results.Error = err
	}
	select {
	case a.resultChan <- &results:
	default:
		a.logger.Error("Failed to send results on channel")
	}
}
