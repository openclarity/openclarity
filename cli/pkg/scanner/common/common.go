package common

import "github.com/openclarity/kubeclarity/shared/pkg/scanner"

// TODO can be multiple fix version?
func GetFixVersion(vulnerability scanner.Vulnerability) string {
	if len(vulnerability.Fix.Versions) > 0 {
		return vulnerability.Fix.Versions[0]
	}
	return ""
}

func SliceContains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
