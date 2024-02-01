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
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/pkg/shared/families/exploits"
	"github.com/openclarity/vmclarity/pkg/shared/families/infofinder"
	"github.com/openclarity/vmclarity/pkg/shared/families/infofinder/types"
	"github.com/openclarity/vmclarity/pkg/shared/families/malware"
	"github.com/openclarity/vmclarity/pkg/shared/families/misconfiguration"
	misconfigurationTypes "github.com/openclarity/vmclarity/pkg/shared/families/misconfiguration/types"
	"github.com/openclarity/vmclarity/pkg/shared/families/rootkits"
	rootkitsTypes "github.com/openclarity/vmclarity/pkg/shared/families/rootkits/types"
	"github.com/openclarity/vmclarity/pkg/shared/families/sbom"
	"github.com/openclarity/vmclarity/pkg/shared/families/secrets"
	"github.com/openclarity/vmclarity/pkg/shared/families/vulnerabilities"
	"github.com/openclarity/vmclarity/pkg/shared/scanner"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
	"github.com/openclarity/vmclarity/pkg/shared/utils/cyclonedx_helper"
	"github.com/openclarity/vmclarity/pkg/shared/utils/vulnerability"
)

func ConvertSBOMResultToPackages(sbomResults *sbom.Results) []apitypes.Package {
	packages := []apitypes.Package{}

	if sbomResults == nil || sbomResults.SBOM == nil || sbomResults.SBOM.Components == nil {
		return packages
	}

	for _, component := range *sbomResults.SBOM.Components {
		packages = append(packages, apitypes.Package{
			Cpes:     utils.PointerTo([]string{component.CPE}),
			Language: utils.PointerTo(cyclonedx_helper.GetComponentLanguage(component)),
			Licenses: utils.PointerTo(cyclonedx_helper.GetComponentLicenses(component)),
			Name:     utils.PointerTo(component.Name),
			Purl:     utils.PointerTo(component.PackageURL),
			Type:     utils.PointerTo(string(component.Type)),
			Version:  utils.PointerTo(component.Version),
		})
	}

	return packages
}

func ConvertVulnResultToVulnerabilities(vulnerabilitiesResults *vulnerabilities.Results) []apitypes.Vulnerability {
	vuls := []apitypes.Vulnerability{}

	if vulnerabilitiesResults == nil || vulnerabilitiesResults.MergedResults == nil || vulnerabilitiesResults.MergedResults.MergedVulnerabilitiesByKey == nil {
		return vuls
	}

	for _, vulCandidates := range vulnerabilitiesResults.MergedResults.MergedVulnerabilitiesByKey {
		if len(vulCandidates) < 1 {
			continue
		}

		vulCandidate := vulCandidates[0]

		vul := apitypes.Vulnerability{
			Cvss:              ConvertVulnCvssToAPIModel(vulCandidate.Vulnerability.CVSS),
			Description:       utils.PointerTo(vulCandidate.Vulnerability.Description),
			Distro:            ConvertVulnDistroToAPIModel(vulCandidate.Vulnerability.Distro),
			Fix:               ConvertVulnFixToAPIModel(vulCandidate.Vulnerability.Fix),
			LayerId:           utils.PointerTo(vulCandidate.Vulnerability.LayerID),
			Links:             utils.PointerTo(vulCandidate.Vulnerability.Links),
			Package:           ConvertVulnPackageToAPIModel(vulCandidate.Vulnerability.Package),
			Path:              utils.PointerTo(vulCandidate.Vulnerability.Path),
			Severity:          ConvertVulnSeverityToAPIModel(vulCandidate.Vulnerability.Severity),
			VulnerabilityName: utils.PointerTo(vulCandidate.Vulnerability.ID),
		}
		vuls = append(vuls, vul)
	}

	return vuls
}

func ConvertVulnSeverityToAPIModel(severity string) *apitypes.VulnerabilitySeverity {
	switch strings.ToUpper(severity) {
	case vulnerability.DEFCON1, vulnerability.CRITICAL:
		return utils.PointerTo(apitypes.CRITICAL)
	case vulnerability.HIGH:
		return utils.PointerTo(apitypes.HIGH)
	case vulnerability.MEDIUM:
		return utils.PointerTo(apitypes.MEDIUM)
	case vulnerability.LOW:
		return utils.PointerTo(apitypes.LOW)
	case vulnerability.NEGLIGIBLE, vulnerability.UNKNOWN, vulnerability.NONE:
		return utils.PointerTo(apitypes.NEGLIGIBLE)
	default:
		log.Errorf("Can't convert severity %q, treating as negligible", severity)
		return utils.PointerTo(apitypes.NEGLIGIBLE)
	}
}

func ConvertVulnFixToAPIModel(fix scanner.Fix) *apitypes.VulnerabilityFix {
	return &apitypes.VulnerabilityFix{
		State:    utils.PointerTo(fix.State),
		Versions: utils.PointerTo(fix.Versions),
	}
}

func ConvertVulnDistroToAPIModel(distro scanner.Distro) *apitypes.VulnerabilityDistro {
	return &apitypes.VulnerabilityDistro{
		IDLike:  utils.PointerTo(distro.IDLike),
		Name:    utils.PointerTo(distro.Name),
		Version: utils.PointerTo(distro.Version),
	}
}

func ConvertVulnPackageToAPIModel(p scanner.Package) *apitypes.Package {
	return &apitypes.Package{
		Cpes:     utils.PointerTo(p.CPEs),
		Language: utils.PointerTo(p.Language),
		Licenses: utils.PointerTo(p.Licenses),
		Name:     utils.PointerTo(p.Name),
		Purl:     utils.PointerTo(p.PURL),
		Type:     utils.PointerTo(p.Type),
		Version:  utils.PointerTo(p.Version),
	}
}

func ConvertVulnCvssToAPIModel(cvss []scanner.CVSS) *[]apitypes.VulnerabilityCvss {
	if cvss == nil {
		return nil
	}
	// nolint:prealloc
	var ret []apitypes.VulnerabilityCvss
	for _, c := range cvss {
		var exploitabilityScore *float32
		if c.Metrics.ExploitabilityScore != nil {
			exploitabilityScore = utils.PointerTo[float32](float32(*c.Metrics.ExploitabilityScore))
		}
		var impactScore *float32
		if c.Metrics.ImpactScore != nil {
			impactScore = utils.PointerTo[float32](float32(*c.Metrics.ImpactScore))
		}
		ret = append(ret, apitypes.VulnerabilityCvss{
			Metrics: &apitypes.VulnerabilityCvssMetrics{
				BaseScore:           utils.PointerTo(float32(c.Metrics.BaseScore)),
				ExploitabilityScore: exploitabilityScore,
				ImpactScore:         impactScore,
			},
			Vector:  utils.PointerTo(c.Vector),
			Version: utils.PointerTo(c.Version),
		})
	}

	return &ret
}

func ConvertMalwareResultToMalwareAndMetadata(malwareResults *malware.MergedResults) ([]apitypes.Malware, []apitypes.ScannerMetadata) {
	malwareList := []apitypes.Malware{}
	metadata := []apitypes.ScannerMetadata{}

	if malwareResults == nil || malwareResults.DetectedMalware == nil {
		return malwareList, metadata
	}

	for _, m := range malwareResults.DetectedMalware {
		mal := m // Prevent loop variable pointer export
		malwareList = append(malwareList, apitypes.Malware{
			MalwareName: utils.OmitEmpty(mal.MalwareName),
			MalwareType: utils.OmitEmpty(mal.MalwareType),
			Path:        utils.OmitEmpty(mal.Path),
			RuleName:    utils.OmitEmpty(mal.RuleName),
		})
	}

	for n, s := range malwareResults.ScansSummary {
		name, summary := n, s // Prevent loop variable pointer export
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

func ConvertSecretsResultToSecrets(secretsResults *secrets.Results) []apitypes.Secret {
	secretsSlice := []apitypes.Secret{}

	if secretsResults == nil || secretsResults.MergedResults == nil || secretsResults.MergedResults.Results == nil {
		return secretsSlice
	}

	for _, resultsCandidate := range secretsResults.MergedResults.Results {
		for i := range resultsCandidate.Findings {
			finding := resultsCandidate.Findings[i]
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
	}

	return secretsSlice
}

func ConvertExploitsResultToExploits(exploitsResults *exploits.Results) []apitypes.Exploit {
	retExploits := []apitypes.Exploit{}

	if exploitsResults == nil || exploitsResults.Exploits == nil {
		return retExploits
	}

	for i := range exploitsResults.Exploits {
		exploit := exploitsResults.Exploits[i]
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

func MisconfigurationSeverityToAPIMisconfigurationSeverity(sev misconfigurationTypes.Severity) (apitypes.MisconfigurationSeverity, error) {
	switch sev {
	case misconfigurationTypes.HighSeverity:
		return apitypes.MisconfigurationHighSeverity, nil
	case misconfigurationTypes.MediumSeverity:
		return apitypes.MisconfigurationMediumSeverity, nil
	case misconfigurationTypes.LowSeverity:
		return apitypes.MisconfigurationLowSeverity, nil
	default:
		return apitypes.MisconfigurationLowSeverity, fmt.Errorf("unknown severity level %v", sev)
	}
}

func ConvertMisconfigurationResultToMisconfigurationsAndScanners(misconfigurationResults *misconfiguration.Results) ([]apitypes.Misconfiguration, []string, error) {
	misconfigurations := []apitypes.Misconfiguration{}
	scanners := []string{}

	if misconfigurationResults == nil || misconfigurationResults.Misconfigurations == nil {
		return misconfigurations, scanners, nil
	}

	for i := range misconfigurationResults.Misconfigurations {
		// create a separate variable for the loop because we need
		// pointers for the API model and we can't safely take pointers
		// to a loop variable.
		misconfig := misconfigurationResults.Misconfigurations[i]

		severity, err := MisconfigurationSeverityToAPIMisconfigurationSeverity(misconfig.Severity)
		if err != nil {
			return misconfigurations, scanners, fmt.Errorf("unable to convert scanner result severity to API severity: %w", err)
		}

		misconfigurations = append(misconfigurations, apitypes.Misconfiguration{
			ScannerName:     &misconfig.ScannerName,
			ScannedPath:     &misconfig.ScannedPath,
			TestCategory:    &misconfig.TestCategory,
			TestID:          &misconfig.TestID,
			TestDescription: &misconfig.TestDescription,
			Severity:        &severity,
			Message:         &misconfig.Message,
			Remediation:     &misconfig.Remediation,
		})
	}

	return misconfigurations, misconfigurationResults.Metadata.Scanners, nil
}

func ConvertInfoFinderResultToInfosAndScanners(results *infofinder.Results) ([]apitypes.InfoFinderInfo, []string, error) {
	infos := []apitypes.InfoFinderInfo{}

	if results == nil || results.Infos == nil {
		return infos, []string{}, nil
	}

	for i := range results.Infos {
		info := results.Infos[i]

		infos = append(infos, apitypes.InfoFinderInfo{
			Data:        &info.Data,
			Path:        &info.Path,
			ScannerName: &info.ScannerName,
			Type:        convertInfoTypeToAPIModel(info.Type),
		})
	}

	return infos, results.Metadata.Scanners, nil
}

func convertInfoTypeToAPIModel(infoType types.InfoType) *apitypes.InfoType {
	switch infoType {
	case types.SSHKnownHostFingerprint:
		return utils.PointerTo(apitypes.InfoTypeSSHKnownHostFingerprint)
	case types.SSHAuthorizedKeyFingerprint:
		return utils.PointerTo(apitypes.InfoTypeSSHAuthorizedKeyFingerprint)
	case types.SSHPrivateKeyFingerprint:
		return utils.PointerTo(apitypes.InfoTypeSSHPrivateKeyFingerprint)
	case types.SSHDaemonKeyFingerprint:
		return utils.PointerTo(apitypes.InfoTypeSSHDaemonKeyFingerprint)
	default:
		log.Errorf("Can't convert info type %q, treating as %v", infoType, apitypes.InfoTypeUNKNOWN)
		return utils.PointerTo(apitypes.InfoTypeUNKNOWN)
	}
}

func ConvertRootkitsResultToRootkits(rootkitsResults *rootkits.Results) []apitypes.Rootkit {
	rootkitsList := []apitypes.Rootkit{}

	if rootkitsResults == nil || rootkitsResults.MergedResults == nil || rootkitsResults.MergedResults.Rootkits == nil {
		return rootkitsList
	}

	for _, r := range rootkitsResults.MergedResults.Rootkits {
		rootkit := r // Prevent loop variable pointer export
		rootkitsList = append(rootkitsList, apitypes.Rootkit{
			Message:     &rootkit.Message,
			RootkitName: &rootkit.RootkitName,
			RootkitType: ConvertRootkitTypeToAPIModel(rootkit.RootkitType),
		})
	}

	return rootkitsList
}

func ConvertRootkitTypeToAPIModel(rootkitType rootkitsTypes.RootkitType) *apitypes.RootkitType {
	switch rootkitType {
	case rootkitsTypes.APPLICATION:
		return utils.PointerTo(apitypes.RootkitTypeAPPLICATION)
	case rootkitsTypes.FIRMWARE:
		return utils.PointerTo(apitypes.RootkitTypeFIRMWARE)
	case rootkitsTypes.KERNEL:
		return utils.PointerTo(apitypes.RootkitTypeKERNEL)
	case rootkitsTypes.MEMORY:
		return utils.PointerTo(apitypes.RootkitTypeMEMORY)
	case rootkitsTypes.UNKNOWN:
		return utils.PointerTo(apitypes.RootkitTypeUNKNOWN)
	default:
		log.Errorf("Can't convert rootkit type %q, treating as %v", rootkitType, apitypes.RootkitTypeUNKNOWN)
		return utils.PointerTo(apitypes.RootkitTypeUNKNOWN)
	}
}
