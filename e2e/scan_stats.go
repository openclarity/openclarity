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

package e2e

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/fbiville/markdown-table-formatter/pkg/markdown"
	"github.com/onsi/gomega"

	apitypes "github.com/openclarity/openclarity/api/types"
	"github.com/openclarity/openclarity/testenv/types"
)

const (
	markDownFilePathDocker = "/tmp/scanner-benchmark-docker.md"
	markDownFilePathK8S    = "/tmp/scanner-benchmark-k8s.md"
	tableHeader            = "# 🚀 Benchmark results"
)

type familyStat struct {
	name  string
	stats *[]apitypes.AssetScanInputScanStats
}

type row struct {
	familyScanner string
	startTime     string
	endTime       string
	findings      string
	totalTime     string
}

type markdownTable struct {
	familyPerScannerRows []row
	summaryRows          []row
	totalFindingsCount   int
}

// generateMarkdownTable generates a markdown table using the provided scan stats.
func generateMarkdownTable(scanStats *apitypes.AssetScanStats) string {
	mdTable := markdownTable{
		familyPerScannerRows: []row{},
		summaryRows:          []row{{"", "", "", "", ""}},
		totalFindingsCount:   0,
	}

	// update rows with stats for each family
	for _, family := range getFamiliesStats(scanStats) {
		mdTable.totalFindingsCount += mdTable.updateRowsWithStats(family)
	}

	// append the summary
	mdTable.summaryRows = append(mdTable.summaryRows, row{
		"_Scan summary_",
		fmt.Sprintf("_%s_", scanStats.General.ScanTime.StartTime.Format(time.DateTime)),
		fmt.Sprintf("_%s_", scanStats.General.ScanTime.EndTime.Format(time.DateTime)),
		fmt.Sprintf("_%d_", mdTable.totalFindingsCount),
		fmt.Sprintf("_%s_", scanStats.General.ScanTime.EndTime.Sub(*scanStats.General.ScanTime.StartTime).Round(time.Second).String()),
	})

	// merge rows
	mdTable.familyPerScannerRows = append(mdTable.familyPerScannerRows, mdTable.summaryRows...)

	tableBody, err := markdown.NewTableFormatterBuilder().
		WithPrettyPrint().
		Build("Family/Scanner", "Start time", "End time", "Findings", "Total time").
		Format(mdTable.toSliceOfSlices())
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return tableHeader + "\n\n" + tableBody
}

func (md *markdownTable) updateRowsWithStats(famStat familyStat) int {
	if famStat.stats == nil {
		return 0
	}

	earliestStartTime := time.Time{}
	latestEndTime := time.Time{}
	totalFindingsCount := 0
	for _, scanner := range *famStat.stats {
		if earliestStartTime.IsZero() || scanner.ScanTime.StartTime.Before(earliestStartTime) {
			earliestStartTime = *scanner.ScanTime.StartTime
		}
		if latestEndTime.IsZero() || scanner.ScanTime.EndTime.After(latestEndTime) {
			latestEndTime = *scanner.ScanTime.EndTime
		}
		totalFindingsCount += *scanner.FindingsCount

		md.familyPerScannerRows = append(md.familyPerScannerRows, row{
			fmt.Sprintf("%s/%s", famStat.name, *scanner.Scanner),
			scanner.ScanTime.StartTime.Format(time.DateTime),
			scanner.ScanTime.EndTime.Format(time.DateTime),
			strconv.Itoa(*scanner.FindingsCount),
			scanner.ScanTime.EndTime.Sub(*scanner.ScanTime.StartTime).Round(time.Second).String(),
		})
	}

	md.summaryRows = append(md.summaryRows, row{
		famStat.name + "/*",
		earliestStartTime.Format(time.DateTime),
		latestEndTime.Format(time.DateTime),
		strconv.Itoa(totalFindingsCount),
		latestEndTime.Sub(earliestStartTime).Round(time.Second).String(),
	})

	return totalFindingsCount
}

func (md *markdownTable) toSliceOfSlices() [][]string {
	result := make([][]string, 0, len(md.familyPerScannerRows))
	for _, r := range md.familyPerScannerRows {
		result = append(result, []string{r.familyScanner, r.startTime, r.endTime, r.findings, r.totalTime})
	}

	return result
}

func getFamiliesStats(scanStats *apitypes.AssetScanStats) []familyStat {
	return []familyStat{
		{"Info finders", scanStats.InfoFinder},
		{"Malware", scanStats.Malware},
		{"Misconfigurations", scanStats.Misconfigurations},
		{"Plugins", scanStats.Plugins},
		{"Rootkits", scanStats.Rootkits},
		{"SBOM", scanStats.Sbom},
		{"Secrets", scanStats.Secrets},
		{"Vulnerabilities", scanStats.Vulnerabilities},
	}
}

// writes the constructed table to a file, that will be displayed in github summary during CI run.
func writeMarkdownTableToFile(mdTable string) {
	//nolint:exhaustive
	switch cfg.TestEnvConfig.Platform {
	case types.EnvironmentTypeDocker:
		err := os.WriteFile(markDownFilePathDocker, []byte(mdTable), 0o600) //nolint:mnd
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	case types.EnvironmentTypeKubernetes:
		err := os.WriteFile(markDownFilePathK8S, []byte(mdTable), 0o600) //nolint:mnd
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}
}
