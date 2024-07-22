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

	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/scanner/common"
	"github.com/openclarity/vmclarity/scanner/families"
	"github.com/openclarity/vmclarity/scanner/families/rootkits/chkrootkit/config"
	chkrootkitutils "github.com/openclarity/vmclarity/scanner/families/rootkits/chkrootkit/utils"
	"github.com/openclarity/vmclarity/scanner/families/rootkits/types"
	familiesutils "github.com/openclarity/vmclarity/scanner/families/utils"
	"github.com/openclarity/vmclarity/scanner/utils"
)

const ScannerName = "chkrootkit"

type Scanner struct {
	config config.Config
}

func New(_ context.Context, _ string, config types.ScannersConfig) (families.Scanner[*types.ScannerResult], error) {
	return &Scanner{
		config: config.Chkrootkit,
	}, nil
}

func (s *Scanner) Scan(ctx context.Context, inputType common.InputType, userInput string) (*types.ScannerResult, error) {
	if !inputType.IsOneOf(common.DIR, common.ROOTFS, common.IMAGE, common.DOCKERARCHIVE, common.OCIARCHIVE, common.OCIDIR) {
		return nil, fmt.Errorf("unsupported input type=%v", inputType)
	}

	logger := log.GetLoggerFromContextOrDefault(ctx)

	chkrootkitBinaryPath, err := exec.LookPath(s.config.GetBinaryPath())
	if err != nil {
		return nil, fmt.Errorf("failed to lookup executable %s: %w", s.config.BinaryPath, err)
	}
	logger.Debugf("found chkrootkit binary at: %s", chkrootkitBinaryPath)

	fsPath, cleanup, err := familiesutils.ConvertInputToFilesystem(ctx, inputType, userInput)
	if err != nil {
		return nil, fmt.Errorf("failed to convert input to filesystem: %w", err)
	}
	defer cleanup()

	args := []string{
		"-r", // Set userInput as the path to the root volume
		fsPath,
	}

	// nolint:gosec
	cmd := exec.Command(chkrootkitBinaryPath, args...)
	logger.Infof("running chkrootkit command: %v", cmd.String())
	out, err := utils.RunCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to run chkrootkit command: %w", err)
	}

	parsedRootkits, err := chkrootkitutils.ParseChkrootkitOutput(out)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chkrootkit output: %w", err)
	}
	parsedRootkits = filterResults(parsedRootkits)

	rootkits := toResultsRootkits(parsedRootkits)

	return &types.ScannerResult{
		Rootkits: rootkits,
	}, nil
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

func toResultsRootkits(rootkits []chkrootkitutils.Rootkit) []types.Rootkit {
	// nolint:prealloc
	var ret []types.Rootkit
	for _, rootkit := range rootkits {
		if !rootkit.Infected {
			continue
		}

		ret = append(ret, types.Rootkit{
			Message:     rootkit.Message,
			RootkitName: rootkit.RkName,
			RootkitType: rootkit.RkType,
		})
	}

	return ret
}
