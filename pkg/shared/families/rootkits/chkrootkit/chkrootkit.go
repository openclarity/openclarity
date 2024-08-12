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

package chkrootkit

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
	"github.com/openclarity/kubeclarity/shared/pkg/utils"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/pkg/shared/families/rootkits/chkrootkit/config"
	chkrootkitutils "github.com/openclarity/vmclarity/pkg/shared/families/rootkits/chkrootkit/utils"
	"github.com/openclarity/vmclarity/pkg/shared/families/rootkits/common"
	familiesutils "github.com/openclarity/vmclarity/pkg/shared/families/utils"
	sharedutils "github.com/openclarity/vmclarity/pkg/shared/utils"
)

const (
	ScannerName      = "chkrootkit"
	ChkrootkitBinary = "chkrootkit"
)

type Scanner struct {
	name       string
	logger     *log.Entry
	config     config.Config
	resultChan chan job_manager.Result
}

func (s *Scanner) Run(sourceType utils.SourceType, userInput string) error {
	go func() {
		retResults := common.Results{
			ScannedInput: userInput,
			ScannerName:  ScannerName,
		}

		if !s.isValidInputType(sourceType) {
			retResults.Error = fmt.Errorf("received invalid input type for chkrootkit scanner: %v", sourceType)
			s.sendResults(retResults, nil)
			return
		}

		// Locate chkrootkit binary
		if s.config.BinaryPath == "" {
			s.config.BinaryPath = ChkrootkitBinary
		}

		chkrootkitBinaryPath, err := exec.LookPath(s.config.BinaryPath)
		if err != nil {
			s.sendResults(retResults, fmt.Errorf("failed to lookup executable %s: %w", s.config.BinaryPath, err))
			return
		}
		s.logger.Debugf("found chkrootkit binary at: %s", chkrootkitBinaryPath)

		fsPath, cleanup, err := familiesutils.ConvertInputToFilesystem(context.TODO(), sourceType, userInput)
		if err != nil {
			s.sendResults(retResults, fmt.Errorf("failed to convert input to filesystem: %w", err))
			return
		}
		defer cleanup()

		args := []string{
			"-r", // Set userInput as the path to the root volume
			fsPath,
		}

		// nolint:gosec
		cmd := exec.Command(chkrootkitBinaryPath, args...)
		s.logger.Infof("running chkrootkit command: %v", cmd.String())
		out, err := sharedutils.RunCommand(cmd)
		if err != nil {
			s.sendResults(retResults, fmt.Errorf("failed to run chkrootkit command: %w", err))
			return
		}

		rootkits, err := chkrootkitutils.ParseChkrootkitOutput(out)
		if err != nil {
			s.sendResults(retResults, fmt.Errorf("failed to parse chkrootkit output: %w", err))
			return
		}
		rootkits = filterResults(rootkits)

		retResults.Rootkits = toResultsRootkits(rootkits)

		s.sendResults(retResults, nil)
	}()

	return nil
}

func filterResults(rootkits []chkrootkitutils.Rootkit) []chkrootkitutils.Rootkit {
	// nolint:prealloc
	var ret []chkrootkitutils.Rootkit
	for _, rootkit := range rootkits {
		if rootkit.RkName == "suspicious files and dirs" {
			// This causes many false positives on every VM, as it's just checks for:
			// files=`${find} ${DIR} -name ".[A-Za-z]*" -o -name "...*" -o -name ".. *"`
			// dirs=`${find} ${DIR} -type d -name ".*"`
			continue
		}
		ret = append(ret, rootkit)
	}
	return ret
}

func toResultsRootkits(rootkits []chkrootkitutils.Rootkit) []common.Rootkit {
	// nolint:prealloc
	var ret []common.Rootkit
	for _, rootkit := range rootkits {
		if !rootkit.Infected {
			continue
		}

		ret = append(ret, common.Rootkit{
			Message:     rootkit.Message,
			RootkitName: rootkit.RkName,
			RootkitType: rootkit.RkType,
		})
	}

	return ret
}

func New(c job_manager.IsConfig, logger *log.Entry, resultChan chan job_manager.Result) job_manager.Job {
	conf := c.(*common.ScannersConfig) // nolint:forcetypeassert
	return &Scanner{
		name:       ScannerName,
		logger:     logger.Dup().WithField("scanner", ScannerName),
		config:     config.Config{BinaryPath: conf.Chkrootkit.BinaryPath},
		resultChan: resultChan,
	}
}

func (s *Scanner) isValidInputType(sourceType utils.SourceType) bool {
	switch sourceType {
	case utils.DIR, utils.ROOTFS, utils.IMAGE, utils.DOCKERARCHIVE, utils.OCIARCHIVE, utils.OCIDIR:
		return true
	case utils.FILE, utils.SBOM:
		fallthrough
	default:
		s.logger.Infof("source type %v is not supported for chkrootkit, skipping.", sourceType)
	}
	return false
}

func (s *Scanner) sendResults(results common.Results, err error) {
	if err != nil {
		s.logger.Error(err)
		results.Error = err
	}
	select {
	case s.resultChan <- &results:
	default:
		s.logger.Error("Failed to send results on channel")
	}
}
