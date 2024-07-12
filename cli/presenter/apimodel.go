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

package presenter

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
	exploits "github.com/openclarity/vmclarity/scanner/families/exploits/types"
	infofinder "github.com/openclarity/vmclarity/scanner/families/infofinder/types"
	malware "github.com/openclarity/vmclarity/scanner/families/malware/types"
	misconfiguration "github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
	rootkits "github.com/openclarity/vmclarity/scanner/families/rootkits/types"
	sbom "github.com/openclarity/vmclarity/scanner/families/sbom/types"
	secrets "github.com/openclarity/vmclarity/scanner/families/secrets/types"
	vulnerabilities "github.com/openclarity/vmclarity/scanner/families/vulnerabilities/types"
	"github.com/openclarity/vmclarity/scanner/utils/cyclonedx_helper"
	"github.com/openclarity/vmclarity/scanner/utils/vulnerability"
)

func ConvertSBOMResultToPackages(result *sbom.Result) []apitypes.Package {
	packages := []apitypes.Package{}

	if result == nil || result.SBOM == nil || result.SBOM.Components == nil {
		return packages
	}

	for _, component := range *result.SBOM.Components {
		packages = append(packages, apitypes.Package{
			Cpes:     to.Ptr([]string{component.CPE}),
			Language: to.Ptr(cyclonedx_helper.GetComponentLanguage(component)),
			Licenses: to.Ptr(cyclonedx_helper.GetComponentLicenses(component)),
			Name:     to.Ptr(component.Name),
			Purl:     to.Ptr(component.PackageURL),
			Type:     to.Ptr(string(component.Type)),
			Version:  to.Ptr(component.Version),
		})
	}

	return packages
}

func ConvertVulnResultToVulnerabilities(result *vulnerabilities.Result) []apitypes.Vulnerability {
	vuls := []apitypes.Vulnerability{}

	if result == nil || result.MergedVulnerabilitiesByKey == nil {
		return vuls
	}

	for _, vulCandidates := range result.MergedVulnerabilitiesByKey {
		if len(vulCandidates) < 1 {
			continue
		}

		vulCandidate := vulCandidates[0]

		vul := apitypes.Vulnerability{
			Cvss:              ConvertVulnCvssToAPIModel(vulCandidate.Vulnerability.CVSS),
			Description:       to.Ptr(vulCandidate.Vulnerability.Description),
			Distro:            ConvertVulnDistroToAPIModel(vulCandidate.Vulnerability.Distro),
			Fix:               ConvertVulnFixToAPIModel(vulCandidate.Vulnerability.Fix),
			LayerId:           to.Ptr(vulCandidate.Vulnerability.LayerID),
			Links:             to.Ptr(vulCandidate.Vulnerability.Links),
			Package:           ConvertVulnPackageToAPIModel(vulCandidate.Vulnerability.Package),
			Path:              to.Ptr(vulCandidate.Vulnerability.Path),
			Severity:          ConvertVulnSeverityToAPIModel(vulCandidate.Vulnerability.Severity),
			VulnerabilityName: to.Ptr(vulCandidate.Vulnerability.ID),
		}
		vuls = append(vuls, vul)
	}

	return vuls
}

func ConvertVulnSeverityToAPIModel(severity string) *apitypes.VulnerabilitySeverity {
	switch strings.ToUpper(severity) {
	case vulnerability.DEFCON1, vulnerability.CRITICAL:
		return to.Ptr(apitypes.CRITICAL)
	case vulnerability.HIGH:
		return to.Ptr(apitypes.HIGH)
	case vulnerability.MEDIUM:
		return to.Ptr(apitypes.MEDIUM)
	case vulnerability.LOW:
		return to.Ptr(apitypes.LOW)
	case vulnerability.NEGLIGIBLE, vulnerability.UNKNOWN, vulnerability.NONE:
		return to.Ptr(apitypes.NEGLIGIBLE)
	default:
		log.Errorf("Can't convert severity %q, treating as negligible", severity)
		return to.Ptr(apitypes.NEGLIGIBLE)
	}
}

func ConvertVulnFixToAPIModel(fix vulnerabilities.Fix) *apitypes.VulnerabilityFix {
	return &apitypes.VulnerabilityFix{
		State:    to.Ptr(fix.State),
		Versions: to.Ptr(fix.Versions),
	}
}

func ConvertVulnDistroToAPIModel(distro vulnerabilities.Distro) *apitypes.VulnerabilityDistro {
	return &apitypes.VulnerabilityDistro{
		IDLike:  to.Ptr(distro.IDLike),
		Name:    to.Ptr(distro.Name),
		Version: to.Ptr(distro.Version),
	}
}

func ConvertVulnPackageToAPIModel(pkg vulnerabilities.Package) *apitypes.Package {
	return &apitypes.Package{
		Cpes:     to.Ptr(pkg.CPEs),
		Language: to.Ptr(pkg.Language),
		Licenses: to.Ptr(pkg.Licenses),
		Name:     to.Ptr(pkg.Name),
		Purl:     to.Ptr(pkg.PURL),
		Type:     to.Ptr(pkg.Type),
		Version:  to.Ptr(pkg.Version),
	}
}

func ConvertVulnCvssToAPIModel(cvss []vulnerabilities.CVSS) *[]apitypes.VulnerabilityCvss {
	if cvss == nil {
		return nil
	}

	// nolint:prealloc
	var ret []apitypes.VulnerabilityCvss
	for _, c := range cvss {
		var exploitabilityScore *float32
		if c.Metrics.ExploitabilityScore != nil {
			exploitabilityScore = to.Ptr[float32](float32(*c.Metrics.ExploitabilityScore))
		}

		var impactScore *float32
		if c.Metrics.ImpactScore != nil {
			impactScore = to.Ptr[float32](float32(*c.Metrics.ImpactScore))
		}

		ret = append(ret, apitypes.VulnerabilityCvss{
			Metrics: &apitypes.VulnerabilityCvssMetrics{
				BaseScore:           to.Ptr(float32(c.Metrics.BaseScore)),
				ExploitabilityScore: exploitabilityScore,
				ImpactScore:         impactScore,
			},
			Vector:  to.Ptr(c.Vector),
			Version: to.Ptr(c.Version),
		})
	}

	return &ret
}

func ConvertMalwareResultToMalwareAndMetadata(result *malware.Result) ([]apitypes.Malware, []apitypes.ScannerMetadata) {
	malwareList := []apitypes.Malware{}
	metadata := []apitypes.ScannerMetadata{}

	if result == nil || result.Malwares == nil {
		return malwareList, metadata
	}

	for _, item := range result.Malwares {
		malwareList = append(malwareList, apitypes.Malware{
			MalwareName: to.PtrOrNil(item.MalwareName),
			MalwareType: to.PtrOrNil(item.MalwareType),
			Path:        to.PtrOrNil(item.Path),
			RuleName:    to.PtrOrNil(item.RuleName),
		})
	}

	for name, summary := range result.ScansSummary {
		if summary == nil {
			summary = &malware.ScanSummary{}
		}

		metadata = append(metadata, apitypes.ScannerMetadata{
			ScannerName: &name,
			ScannerSummary: &apitypes.ScannerSummary{
				DataRead:           &summary.DataRead,
				DataScanned:        &summary.DataScanned,
				EngineVersion:      &summary.EngineVersion,
				InfectedFiles:      &summary.InfectedFiles,
				KnownViruses:       &summary.KnownViruses,
				ScannedDirectories: &summary.ScannedDirectories,
				ScannedFiles:       &summary.ScannedFiles,
				SuspectedFiles:     &summary.SuspectedFiles,
				TimeTaken:          &summary.TimeTaken,
			},
		})
	}

	return malwareList, metadata
}

func ConvertSecretsResultToSecrets(result *secrets.Result) []apitypes.Secret {
	secretsSlice := []apitypes.Secret{}

	if result == nil || result.Findings == nil {
		return secretsSlice
	}

	for _, finding := range result.Findings {
		secretsSlice = append(secretsSlice, apitypes.Secret{
			Description: &finding.Description,
			EndLine:     &finding.EndLine,
			FilePath:    &finding.File,
			Fingerprint: &finding.Fingerprint,
			StartLine:   &finding.StartLine,
			StartColumn: &finding.StartColumn,
			EndColumn:   &finding.EndColumn,
		})
	}

	return secretsSlice
}

func ConvertExploitsResultToExploits(result *exploits.Result) []apitypes.Exploit {
	retExploits := []apitypes.Exploit{}

	if result == nil || result.Exploits == nil {
		return retExploits
	}

	for _, exploit := range result.Exploits {
		retExploits = append(retExploits, apitypes.Exploit{
			CveID:       &exploit.CveID,
			Description: &exploit.Description,
			Name:        &exploit.Name,
			SourceDB:    &exploit.SourceDB,
			Title:       &exploit.Title,
			Urls:        &exploit.URLs,
		})
	}

	return retExploits
}

func MisconfigurationSeverityToAPIMisconfigurationSeverity(sev misconfiguration.Severity) (apitypes.MisconfigurationSeverity, error) {
	switch sev {
	case misconfiguration.HighSeverity:
		return apitypes.MisconfigurationHighSeverity, nil
	case misconfiguration.MediumSeverity:
		return apitypes.MisconfigurationMediumSeverity, nil
	case misconfiguration.LowSeverity:
		return apitypes.MisconfigurationLowSeverity, nil
	case misconfiguration.InfoSeverity:
		return apitypes.MisconfigurationInfoSeverity, nil
	default:
		return apitypes.MisconfigurationLowSeverity, fmt.Errorf("unknown severity level %v", sev)
	}
}

func ConvertMisconfigurationResultToMisconfigurationsAndScanners(result *misconfiguration.Result) ([]apitypes.Misconfiguration, []string, error) {
	misconfigurations := []apitypes.Misconfiguration{}
	scanners := map[string]interface{}{}

	if result == nil || result.Misconfigurations == nil {
		return misconfigurations, to.SortedKeys(scanners), nil
	}

	for _, misconfig := range result.Misconfigurations {
		severity, err := MisconfigurationSeverityToAPIMisconfigurationSeverity(misconfig.Severity)
		if err != nil {
			return misconfigurations, to.SortedKeys(scanners), fmt.Errorf("unable to convert scanner result severity to API severity: %w", err)
		}

		misconfigurations = append(misconfigurations, apitypes.Misconfiguration{
			ScannerName: &misconfig.ScannerName,
			Location:    &misconfig.Location,
			Category:    &misconfig.Category,
			Id:          &misconfig.ID,
			Description: &misconfig.Description,
			Severity:    &severity,
			Message:     &misconfig.Message,
			Remediation: &misconfig.Remediation,
		})

		scanners[misconfig.ScannerName] = nil
	}

	return misconfigurations, to.SortedKeys(scanners), nil
}

func ConvertInfoFinderResultToInfosAndScanners(result *infofinder.Result) ([]apitypes.InfoFinderInfo, []string, error) {
	infos := []apitypes.InfoFinderInfo{}
	scanners := map[string]interface{}{}

	if result == nil || result.Infos == nil {
		return infos, to.SortedKeys(scanners), nil
	}

	for _, info := range result.Infos {
		infos = append(infos, apitypes.InfoFinderInfo{
			Data:        &info.Data,
			Path:        &info.Path,
			ScannerName: &info.ScannerName,
			Type:        convertInfoTypeToAPIModel(info.Type),
		})

		scanners[info.ScannerName] = nil
	}

	return infos, to.SortedKeys(scanners), nil
}

func convertInfoTypeToAPIModel(infoType infofinder.InfoType) *apitypes.InfoType {
	switch infoType {
	case infofinder.SSHKnownHostFingerprint:
		return to.Ptr(apitypes.InfoTypeSSHKnownHostFingerprint)
	case infofinder.SSHAuthorizedKeyFingerprint:
		return to.Ptr(apitypes.InfoTypeSSHAuthorizedKeyFingerprint)
	case infofinder.SSHPrivateKeyFingerprint:
		return to.Ptr(apitypes.InfoTypeSSHPrivateKeyFingerprint)
	case infofinder.SSHDaemonKeyFingerprint:
		return to.Ptr(apitypes.InfoTypeSSHDaemonKeyFingerprint)
	default:
		log.Errorf("Can't convert info type %q, treating as %v", infoType, apitypes.InfoTypeUNKNOWN)
		return to.Ptr(apitypes.InfoTypeUNKNOWN)
	}
}

func ConvertRootkitsResultToRootkits(result *rootkits.Result) []apitypes.Rootkit {
	rootkitsList := []apitypes.Rootkit{}

	if result == nil || result.Rootkits == nil {
		return rootkitsList
	}

	for _, rootkit := range result.Rootkits {
		rootkitsList = append(rootkitsList, apitypes.Rootkit{
			Message:     &rootkit.Message,
			RootkitName: &rootkit.RootkitName,
			RootkitType: ConvertRootkitTypeToAPIModel(rootkit.RootkitType),
		})
	}

	return rootkitsList
}

func ConvertRootkitTypeToAPIModel(rootkitType rootkits.RootkitType) *apitypes.RootkitType {
	switch rootkitType {
	case rootkits.APPLICATION:
		return to.Ptr(apitypes.RootkitTypeAPPLICATION)
	case rootkits.FIRMWARE:
		return to.Ptr(apitypes.RootkitTypeFIRMWARE)
	case rootkits.KERNEL:
		return to.Ptr(apitypes.RootkitTypeKERNEL)
	case rootkits.MEMORY:
		return to.Ptr(apitypes.RootkitTypeMEMORY)
	case rootkits.UNKNOWN:
		return to.Ptr(apitypes.RootkitTypeUNKNOWN)
	default:
		log.Errorf("Can't convert rootkit type %q, treating as %v", rootkitType, apitypes.RootkitTypeUNKNOWN)
		return to.Ptr(apitypes.RootkitTypeUNKNOWN)
	}
}
