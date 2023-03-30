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

package scanner

import (
	"sort"

	"github.com/openclarity/kubeclarity/shared/pkg/scanner/types"
	vulutil "github.com/openclarity/kubeclarity/shared/pkg/utils/vulnerability"
)

// SortBySeverityAndCVSS sorts vulnerabilities by severity, CVSSv3.1, CVSSv3.0 and CVSSv2.0.
func SortBySeverityAndCVSS(vulnerabilities []MergedVulnerability) []MergedVulnerability {
	sort.Slice(vulnerabilities, func(i, j int) bool {
		if vulutil.GetSeverityIntFromString(vulnerabilities[i].Vulnerability.Severity) >
			vulutil.GetSeverityIntFromString(vulnerabilities[j].Vulnerability.Severity) {
			return true
		}

		if getCVSSBaseScore(vulnerabilities[i].Vulnerability.CVSS, "3.1") >
			getCVSSBaseScore(vulnerabilities[j].Vulnerability.CVSS, "3.1") {
			return true
		}

		if getCVSSBaseScore(vulnerabilities[i].Vulnerability.CVSS, "3.0") >
			getCVSSBaseScore(vulnerabilities[j].Vulnerability.CVSS, "3.0") {
			return true
		}

		if getCVSSBaseScore(vulnerabilities[i].Vulnerability.CVSS, "2.0") >
			getCVSSBaseScore(vulnerabilities[j].Vulnerability.CVSS, "2.0") {
			return true
		}

		return false
	})

	return vulnerabilities
}

func getCVSSBaseScore(cvss []types.CVSS, version string) float64 {
	for _, c := range cvss {
		if c.Version == version {
			return c.Metrics.BaseScore
		}
	}
	return 0
}
