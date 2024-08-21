// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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

package benchmark

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/fbiville/markdown-table-formatter/pkg/markdown"

	apitypes "github.com/openclarity/openclarity/api/types"
)

const tableTitle = "# 🚀 Benchmark results"

type BenchmarkExtractor interface {
	ExportAsMarkdown(file string) error
	header() []string
	rows() [][]string
}

type familyStat struct {
	name  string
	stats *[]apitypes.AssetScanInputScanStats
}

type benchmark struct {
	familyStats        []familyStat
	generalStats       *apitypes.AssetScanGeneralStats
	totalFindingsCount int // TODO: This value does not represent the actual findings count, it should be calculated based on the findings of each family.
}

func NewBenchmarkExtractor(scanStats *apitypes.AssetScanStats) (BenchmarkExtractor, error) {
	b := &benchmark{
		generalStats: scanStats.General,
	}
	for _, family := range getFamiliesStats(scanStats) {
		if family.stats != nil {
			b.familyStats = append(b.familyStats, family)
		}
	}

	return b, nil
}

func (b *benchmark) header() []string {
	return []string{"Family/Scanner", "Start time", "End time", "Findings", "Total time"}
}

func (b *benchmark) rows() [][]string {
	var familyPerScannerRows [][]string //nolint:prealloc
	var summaryRows [][]string          //nolint:prealloc

	for _, family := range b.familyStats {
		earliestStartTime := time.Time{}
		latestEndTime := time.Time{}
		totalFindingsCount := 0
		for _, scanner := range *family.stats {
			if earliestStartTime.IsZero() || scanner.ScanTime.StartTime.Before(earliestStartTime) {
				earliestStartTime = *scanner.ScanTime.StartTime
			}
			if latestEndTime.IsZero() || scanner.ScanTime.EndTime.After(latestEndTime) {
				latestEndTime = *scanner.ScanTime.EndTime
			}
			totalFindingsCount += *scanner.FindingsCount

			familyPerScannerRows = append(familyPerScannerRows, []string{
				fmt.Sprintf("%s/%s", family.name, *scanner.Scanner),
				scanner.ScanTime.StartTime.Format(time.DateTime),
				scanner.ScanTime.EndTime.Format(time.DateTime),
				strconv.Itoa(*scanner.FindingsCount),
				scanner.ScanTime.EndTime.Sub(*scanner.ScanTime.StartTime).Round(time.Second).String(),
			})
		}

		summaryRows = append(summaryRows, []string{
			family.name + "/*",
			earliestStartTime.Format(time.DateTime),
			latestEndTime.Format(time.DateTime),
			strconv.Itoa(totalFindingsCount),
			latestEndTime.Sub(earliestStartTime).Round(time.Second).String(),
		})

		b.totalFindingsCount += totalFindingsCount
	}

	familyPerScannerRows = append(familyPerScannerRows, summaryRows...)

	b.orderByColumn(familyPerScannerRows, 0)

	result := append(familyPerScannerRows,
		[]string{"", "", "", "", ""},
		[]string{
			"_Scan summary_",
			fmt.Sprintf("_%s_", b.generalStats.ScanTime.StartTime.Format(time.DateTime)),
			fmt.Sprintf("_%s_", b.generalStats.ScanTime.EndTime.Format(time.DateTime)),
			fmt.Sprintf("_%d_", b.totalFindingsCount),
			fmt.Sprintf("_%s_", b.generalStats.ScanTime.EndTime.Sub(*b.generalStats.ScanTime.StartTime).Round(time.Second).String()),
		})

	return result
}

func (b *benchmark) orderByColumn(rows [][]string, columnNumber int) {
	sort.Slice(rows, func(i, j int) bool {
		return rows[i][columnNumber] < rows[j][columnNumber]
	})
}

func (b *benchmark) ExportAsMarkdown(file string) error {
	table, err := markdown.NewTableFormatterBuilder().
		WithPrettyPrint().
		Build(b.header()...).
		Format(b.rows())
	if err != nil {
		return fmt.Errorf("failed to format table: %w", err)
	}

	err = os.WriteFile(file, []byte(tableTitle+"\n\n"+table), 0o600) //nolint:mnd
	if err != nil {
		return fmt.Errorf("failed to write markdown table to file: %w", err)
	}

	return nil
}

func getFamiliesStats(scanStats *apitypes.AssetScanStats) []familyStat {
	return []familyStat{
		{"InfoFinders", scanStats.InfoFinder},
		{"Malware", scanStats.Malware},
		{"Misconfigurations", scanStats.Misconfigurations},
		{"Plugins", scanStats.Plugins},
		{"Rootkits", scanStats.Rootkits},
		{"SBOM", scanStats.Sbom},
		{"Secrets", scanStats.Secrets},
		{"Vulnerabilities", scanStats.Vulnerabilities},
	}
}
