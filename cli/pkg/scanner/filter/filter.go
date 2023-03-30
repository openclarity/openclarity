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

package filter

import (
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/kubeclarity/shared/pkg/scanner"
	"github.com/openclarity/kubeclarity/shared/pkg/scanner/types"
	"github.com/openclarity/kubeclarity/shared/pkg/utils/slice"
)

type Ignores struct {
	NoFix           bool
	Vulnerabilities []string
}

// nolint:revive
func FilterIgnoredVulnerabilities(m *scanner.MergedResults, ignores Ignores) *scanner.MergedResults {
	filteredMergedResults := scanner.NewMergedResults()
	for key, vulnerabilities := range m.MergedVulnerabilitiesByKey {
		if len(vulnerabilities) > 1 {
			vulnerabilities = scanner.SortBySeverityAndCVSS(vulnerabilities)
		}
		vulnerability := vulnerabilities[0]
		if shouldIgnore(vulnerability.Vulnerability, ignores) {
			continue
		}
		filteredMergedResults.MergedVulnerabilitiesByKey[key] = vulnerabilities
	}
	filteredMergedResults.Source = m.Source

	return filteredMergedResults
}

func shouldIgnore(vulnerability types.Vulnerability, ignores Ignores) bool {
	if slice.Contains(ignores.Vulnerabilities, vulnerability.ID) {
		log.Debugf("Ignoring vulnerability due to ignore list %q", vulnerability.ID)
		return true
	}
	if ignores.NoFix {
		if types.GetFixVersion(vulnerability) == "" {
			log.Debugf("Ignoring vulnerability due to no fix %q", vulnerability.ID)
			return true
		}
	}

	return false
}
