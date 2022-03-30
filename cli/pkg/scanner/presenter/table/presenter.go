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

package table

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"

	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/scanner"
	utils "wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/utils/vulnerability"
)

type Presenter struct {
	mergedResults *scanner.MergedResults
}

// NewPresenter is a *Presenter constructor.
func NewPresenter(mergedResults *scanner.MergedResults) *Presenter {
	return &Presenter{
		mergedResults: mergedResults,
	}
}

// Present creates a table-based reporting.
func (pres *Presenter) Present(output io.Writer) error {
	rows := make([][]string, 0)

	columns := []string{"Name", "Installed", "Fixed-In", "Vulnerability", "Severity", "Scanners"}
	results := sortBySeverity(pres.mergedResults.ToSlice())
	for _, mergedVulnerabilities := range results {
		// Show vulnerability details for the highest vulnerability found in the mergedVulnerabilities
		vulnerability := mergedVulnerabilities[0].Vulnerability

		scanners := strings.Join(getScanners(mergedVulnerabilities), ", ")

		rows = append(rows, []string{
			vulnerability.Package.Name, vulnerability.Package.Version,
			strings.Join(vulnerability.Fix.Versions, ", "),
			vulnerability.ID, vulnerability.Severity, scanners,
		})
	}

	if len(rows) == 0 {
		_, err := io.WriteString(output, "No vulnerabilities found\n")
		if err != nil {
			return fmt.Errorf("failed to write string: %v", err)
		}
		return nil
	}

	rows = removeDuplicateRows(rows)

	table := tablewriter.NewWriter(output)

	table.SetHeader(columns)
	table.SetAutoWrapText(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetAutoFormatHeaders(true)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)

	table.AppendBulk(rows)
	table.Render()

	return nil
}

func getScanners(mergedVulnerabilities []scanner.MergedVulnerability) (ret []string) {
	scannerNameFormat := "%s"
	if len(mergedVulnerabilities) > 1 {
		// (*) will be added to all scanners names if diffs was found in the vulnerability results
		scannerNameFormat += "(*)"
	}
	for _, mergedVulnerability := range mergedVulnerabilities {
		for _, info := range mergedVulnerability.ScannersInfo {
			ret = append(ret, fmt.Sprintf(scannerNameFormat, info.Name))
		}
	}
	return ret
}

func sortBySeverity(results [][]scanner.MergedVulnerability) [][]scanner.MergedVulnerability {
	// sort mergedVulnerabilities
	for _, mergedVulnerabilities := range results {
		sort.Slice(mergedVulnerabilities, func(i, j int) bool {
			return utils.GetSeverityIntFromString(mergedVulnerabilities[i].Vulnerability.Severity) >
				utils.GetSeverityIntFromString(mergedVulnerabilities[j].Vulnerability.Severity)
		})
	}

	// sort results, picking the first element since mergedVulnerabilities already sorted
	sort.Slice(results, func(i, j int) bool {
		return utils.GetSeverityIntFromString(results[i][0].Vulnerability.Severity) >
			utils.GetSeverityIntFromString(results[j][0].Vulnerability.Severity)
	})

	return results
}

func removeDuplicateRows(items [][]string) [][]string {
	seen := map[string][]string{}
	// nolint:prealloc
	var result [][]string

	for _, v := range items {
		key := strings.Join(v, "|")
		if seen[key] != nil {
			// dup!
			continue
		}

		seen[key] = v
		result = append(result, v)
	}
	return result
}
