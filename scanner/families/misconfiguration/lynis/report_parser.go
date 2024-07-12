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
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
)

const (
	lynisSuggestionWarningParts = 5
)

type ReportParser struct {
	testdb *TestDB
}

func NewReportParser(testdb *TestDB) *ReportParser {
	return &ReportParser{
		testdb: testdb,
	}
}

func (a *ReportParser) ParseLynisReport(scanPath string, reportPath string) ([]types.Misconfiguration, error) {
	report, err := os.Open(reportPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open report file from path %v: %w", reportPath, err)
	}
	defer report.Close()

	scanner := bufio.NewScanner(report)
	return a.scanLynisReportFile(scanPath, scanner)
}

type FileScanner interface {
	Scan() bool
	Text() string
	Err() error
}

func (a *ReportParser) scanLynisReportFile(scanPath string, scanner FileScanner) ([]types.Misconfiguration, error) {
	output := []types.Misconfiguration{}
	for scanner.Scan() {
		line := scanner.Text()
		isMisconfiguration, misconfiguration, err := a.parseLynisReportLine(scanPath, line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse report line %q: %w", line, err)
		}
		if isMisconfiguration {
			output = append(output, misconfiguration)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read lines for report: %w", err)
	}
	return output, nil
}

func (a *ReportParser) parseLynisReportLine(scanPath string, line string) (bool, types.Misconfiguration, error) {
	switch line[0] {
	// Comment/Remark start with # or Sections which start with [
	case '#', '[':
		// skip these lines
		return false, types.Misconfiguration{}, nil
	}

	// Everything else should be in the "option" format, <option>=<value>
	option, value, ok := strings.Cut(line, "=")
	if !ok {
		return false, types.Misconfiguration{}, errors.New("line not in option=value format")
	}

	switch option {
	case "suggestion[]":
		mis, err := a.valueToMisconfiguration(scanPath, value, types.LowSeverity)
		if err != nil {
			return false, types.Misconfiguration{}, fmt.Errorf("could not convert suggestion value %s to low misconfiguration: %w", value, err)
		}

		// LYNIS suggestions are about the lynis install itself, we
		// should ignore these.
		if mis.ID == "LYNIS" {
			return false, types.Misconfiguration{}, nil
		}

		return true, mis, nil
	case "warning[]":
		mis, err := a.valueToMisconfiguration(scanPath, value, types.HighSeverity)
		if err != nil {
			return false, types.Misconfiguration{}, fmt.Errorf("could not convert warning value %s to high misconfiguration: %w", value, err)
		}
		return true, mis, nil
	default:
		// all other options aren't misconfigurations
		return false, types.Misconfiguration{}, nil
	}
}

// valueToMisconfiguration converts a lynis warning/suggestion into a
// Misconfiguration. They are made up of parts separated by pipes for example:
//
// BOOT-5122|Set a password on GRUB boot loader to prevent altering boot configuration (e.g. boot in single user mode without password)|-|-|
//
// These are translated as:
//
// <TestID>|<Message>|<Message Details/Part 2>|<Remediation>|
//
// The function will error if the value is malformed.
func (a *ReportParser) valueToMisconfiguration(scanPath string, value string, severity types.Severity) (types.Misconfiguration, error) {
	parts := strings.Split(value, "|")
	if len(parts) != lynisSuggestionWarningParts {
		return types.Misconfiguration{}, fmt.Errorf("got %d sections, expected %d", len(parts), lynisSuggestionWarningParts)
	}

	message := fmt.Sprintf("%s Details: %s", parts[1], parts[2])

	return types.Misconfiguration{
		Location:    scanPath,
		Category:    a.testdb.GetCategoryForTestID(parts[0]),
		ID:          parts[0],
		Description: a.testdb.GetDescriptionForTestID(parts[0]),
		Severity:    severity,
		Message:     message,
		Remediation: parts[3],
	}, nil
}
