package filter

import (
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/kubeclarity/cli/pkg/scanner/common"
	"github.com/openclarity/kubeclarity/shared/pkg/scanner"
)

type Ignores struct {
	NoFix           bool
	Vulnerabilities []string
}

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

func shouldIgnore(vulnerability scanner.Vulnerability, ignores Ignores) bool {
	if common.SliceContains(ignores.Vulnerabilities, vulnerability.ID) {
		log.Debugf("Ignoring vulnerability due to ignore list %q", vulnerability.ID)
		return true
	}
	if ignores.NoFix {
		if common.GetFixVersion(vulnerability) == "" {
			log.Debugf("Ignoring vulnerability due to no fix %q", vulnerability.ID)
			return true
		}
	}

	return false
}
