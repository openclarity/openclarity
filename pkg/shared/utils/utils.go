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

func PointerTo[T any](value T) *T {
	return &value
}

// ValueOrZero returns the value that the pointer ptr pointers to. It returns
// the zero value if ptr is nil.
func ValueOrZero[T any](ptr *T) T {
	var t T
	if ptr != nil {
		t = *ptr
	}
	return t
}

func StringPointerValOrEmpty(val *string) string {
	if val == nil {
		return ""
	}
	return *val
}

func Int32PointerValOrEmpty(val *int32) int32 {
	if val == nil {
		return 0
	}
	return *val
}

func IntPointerValOrEmpty(val *int) int {
	if val == nil {
		return 0
	}
	return *val
}

func BoolPointerValOrFalse(val *bool) bool {
	if val == nil {
		return false
	}
	return *val
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
		TotalCriticalVulnerabilities:   PointerTo(0),
		TotalHighVulnerabilities:       PointerTo(0),
		TotalMediumVulnerabilities:     PointerTo(0),
		TotalLowVulnerabilities:        PointerTo(0),
		TotalNegligibleVulnerabilities: PointerTo(0),
	}
	if vulnerabilities == nil {
		return ret
	}
	for _, vulnerability := range *vulnerabilities {
		switch *vulnerability.Severity {
		case models.CRITICAL:
			ret.TotalCriticalVulnerabilities = PointerTo(*ret.TotalCriticalVulnerabilities + 1)
		case models.HIGH:
			ret.TotalHighVulnerabilities = PointerTo(*ret.TotalHighVulnerabilities + 1)
		case models.MEDIUM:
			ret.TotalMediumVulnerabilities = PointerTo(*ret.TotalMediumVulnerabilities + 1)
		case models.LOW:
			ret.TotalLowVulnerabilities = PointerTo(*ret.TotalLowVulnerabilities + 1)
		case models.NEGLIGIBLE:
			ret.TotalNegligibleVulnerabilities = PointerTo(*ret.TotalNegligibleVulnerabilities + 1)
		}
	}
	return ret
}
