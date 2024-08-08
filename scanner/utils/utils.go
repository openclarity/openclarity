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

package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"

	"golang.org/x/sync/errgroup"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
)

type CmdRunError struct {
	Cmd    *exec.Cmd
	Err    error
	Stdout []byte
	Stderr string
}

type processFn func(string)

func (r CmdRunError) Error() string {
	return fmt.Sprintf(
		"failed to run command %v, error: %s, stdout: %s, stderr: %s",
		r.Cmd, r.Err, string(r.Stdout), r.Stderr,
	)
}

func RunCommand(cmd *exec.Cmd) ([]byte, error) {
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return nil, CmdRunError{
			Cmd:    cmd,
			Err:    err,
			Stdout: outb.Bytes(),
			Stderr: errb.String(),
		}
	}

	return outb.Bytes(), nil
}

func RunCommandAndParseOutputLineByLine(cmd *exec.Cmd, pfn, ecFn processFn) error {
	// Get a pipe to read from standard out
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start the command and check for errors
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	eg := errgroup.Group{}

	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)

	eg.Go(func() error {
		for stderrScanner.Scan() {
			line := stderrScanner.Text()
			ecFn(line)
		}
		if err := stderrScanner.Err(); err != nil {
			return fmt.Errorf("stderr scanner error: %w", err)
		}
		return nil
	})

	// Use the scanner to scan the output line by line and parse it
	eg.Go(func() error {
		for stdoutScanner.Scan() {
			line := stdoutScanner.Text()
			pfn(line)
		}
		if err := stdoutScanner.Err(); err != nil {
			return fmt.Errorf("stdout scanner error: %w", err)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command returns error: %w", err)
	}

	return nil
}

func GetVulnerabilityTotalsPerSeverity(vulnerabilities *[]apitypes.Vulnerability) *apitypes.VulnerabilitySeveritySummary {
	ret := &apitypes.VulnerabilitySeveritySummary{
		TotalCriticalVulnerabilities:   to.Ptr(0),
		TotalHighVulnerabilities:       to.Ptr(0),
		TotalMediumVulnerabilities:     to.Ptr(0),
		TotalLowVulnerabilities:        to.Ptr(0),
		TotalNegligibleVulnerabilities: to.Ptr(0),
	}
	if vulnerabilities == nil {
		return ret
	}
	for _, vulnerability := range *vulnerabilities {
		switch *vulnerability.Severity {
		case apitypes.CRITICAL:
			ret.TotalCriticalVulnerabilities = to.Ptr(*ret.TotalCriticalVulnerabilities + 1)
		case apitypes.HIGH:
			ret.TotalHighVulnerabilities = to.Ptr(*ret.TotalHighVulnerabilities + 1)
		case apitypes.MEDIUM:
			ret.TotalMediumVulnerabilities = to.Ptr(*ret.TotalMediumVulnerabilities + 1)
		case apitypes.LOW:
			ret.TotalLowVulnerabilities = to.Ptr(*ret.TotalLowVulnerabilities + 1)
		case apitypes.NEGLIGIBLE:
			ret.TotalNegligibleVulnerabilities = to.Ptr(*ret.TotalNegligibleVulnerabilities + 1)
		}
	}
	return ret
}
