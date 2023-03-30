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
	"bytes"
	"fmt"
	"os/exec"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
)

const (
	DefaultResourceReadyWaitTimeoutMin   = 3
	DefaultResourceReadyCheckIntervalSec = 3
)

type CmdRunError struct {
	Cmd    *exec.Cmd
	Err    error
	Stdout []byte
	Stderr string
}

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

func StringPtr(val string) *string {
	ret := val
	return &ret
}

func BoolPtr(val bool) *bool {
	ret := val
	return &ret
}

func PointerTo[T any](value T) *T {
	return &value
}

func StringKeyMapToArray[T any](m map[string]T) []T {
	ret := make([]T, 0, len(m))
	for _, t := range m {
		ret = append(ret, t)
	}
	return ret
}

func GetVulnerabilityTotalsPerSeverity(vulnerabilities *[]models.Vulnerability) *models.VulnerabilityScanSummary {
	ret := &models.VulnerabilityScanSummary{
		TotalCriticalVulnerabilities:   utils.PointerTo(0),
		TotalHighVulnerabilities:       utils.PointerTo(0),
		TotalMediumVulnerabilities:     utils.PointerTo(0),
		TotalLowVulnerabilities:        utils.PointerTo(0),
		TotalNegligibleVulnerabilities: utils.PointerTo(0),
	}
	if vulnerabilities == nil {
		return ret
	}
	for _, vulnerability := range *vulnerabilities {
		switch *vulnerability.Severity {
		case models.CRITICAL:
			ret.TotalCriticalVulnerabilities = utils.PointerTo(*ret.TotalCriticalVulnerabilities + 1)
		case models.HIGH:
			ret.TotalHighVulnerabilities = utils.PointerTo(*ret.TotalHighVulnerabilities + 1)
		case models.MEDIUM:
			ret.TotalMediumVulnerabilities = utils.PointerTo(*ret.TotalMediumVulnerabilities + 1)
		case models.LOW:
			ret.TotalLowVulnerabilities = utils.PointerTo(*ret.TotalLowVulnerabilities + 1)
		case models.NEGLIGIBLE:
			ret.TotalNegligibleVulnerabilities = utils.PointerTo(*ret.TotalNegligibleVulnerabilities + 1)
		}
	}
	return ret
}
