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

	"github.com/openclarity/kubeclarity/shared/pkg/scanner"
	"github.com/openclarity/kubeclarity/shared/pkg/utils/cyclonedx_helper"
	"github.com/openclarity/kubeclarity/shared/pkg/utils/vulnerability"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/api/models"
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
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

func ConvertSBOMResultToPackages(sbomResults *sbom.Results) []models.Package {
	packages := []models.Package{}

	if sbomResults == nil || sbomResults.SBOM == nil || sbomResults.SBOM.Components == nil {
		return packages
	}

	for _, component := range *sbomResults.SBOM.Components {
		packages = append(packages, models.Package{
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

func ConvertVulnResultToVulnerabilities(vulnerabilitiesResults *vulnerabilities.Results) []models.Vulnerability {
	vuls := []models.Vulnerability{}

	if vulnerabilitiesResults == nil || vulnerabilitiesResults.MergedResults == nil || vulnerabilitiesResults.MergedResults.MergedVulnerabilitiesByKey == nil {
		return vuls
	}

	for _, vulCandidates := range vulnerabilitiesResults.MergedResults.MergedVulnerabilitiesByKey {
		if len(vulCandidates) < 1 {
			continue
		}

		vulCandidate := vulCandidates[0]

		vul := models.Vulnerability{
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

func ConvertVulnSeverityToAPIModel(severity string) *models.VulnerabilitySeverity {
	switch strings.ToUpper(severity) {
	case vulnerability.DEFCON1, vulnerability.CRITICAL:
		return utils.PointerTo(models.CRITICAL)
	case vulnerability.HIGH:
		return utils.PointerTo(models.HIGH)
	case vulnerability.MEDIUM:
		return utils.PointerTo(models.MEDIUM)
	case vulnerability.LOW:
		return utils.PointerTo(models.LOW)
	case vulnerability.NEGLIGIBLE, vulnerability.UNKNOWN, vulnerability.NONE:
		return utils.PointerTo(models.NEGLIGIBLE)
	default:
		log.Errorf("Can't convert severity %q, treating as negligible", severity)
		return utils.PointerTo(models.NEGLIGIBLE)
	}
}

func ConvertVulnFixToAPIModel(fix scanner.Fix) *models.VulnerabilityFix {
	return &models.VulnerabilityFix{
		State:    utils.PointerTo(fix.State),
		Versions: utils.PointerTo(fix.Versions),
	}
}

func ConvertVulnDistroToAPIModel(distro scanner.Distro) *models.VulnerabilityDistro {
	return &models.VulnerabilityDistro{
		IDLike:  utils.PointerTo(distro.IDLike),
		Name:    utils.PointerTo(distro.Name),
		Version: utils.PointerTo(distro.Version),
	}
}

func ConvertVulnPackageToAPIModel(p scanner.Package) *models.Package {
	return &models.Package{
		Cpes:     utils.PointerTo(p.CPEs),
		Language: utils.PointerTo(p.Language),
		Licenses: utils.PointerTo(p.Licenses),
		Name:     utils.PointerTo(p.Name),
		Purl:     utils.PointerTo(p.PURL),
		Type:     utils.PointerTo(p.Type),
		Version:  utils.PointerTo(p.Version),
	}
}

func ConvertVulnCvssToAPIModel(cvss []scanner.CVSS) *[]models.VulnerabilityCvss {
	if cvss == nil {
		return nil
	}
	// nolint:prealloc
	var ret []models.VulnerabilityCvss
	for _, c := range cvss {
		var exploitabilityScore *float32
		if c.Metrics.ExploitabilityScore != nil {
			exploitabilityScore = utils.PointerTo[float32](float32(*c.Metrics.ExploitabilityScore))
		}
		var impactScore *float32
		if c.Metrics.ImpactScore != nil {
			impactScore = utils.PointerTo[float32](float32(*c.Metrics.ImpactScore))
		}
		ret = append(ret, models.VulnerabilityCvss{
			Metrics: &models.VulnerabilityCvssMetrics{
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

func ConvertMalwareResultToMalwareAndMetadata(malwareResults *malware.MergedResults) ([]models.Malware, []models.ScannerMetadata) {
	malwareList := []models.Malware{}
	metadata := []models.ScannerMetadata{}

	if malwareResults == nil || malwareResults.DetectedMalware == nil {
		return malwareList, metadata
	}

	for _, m := range malwareResults.DetectedMalware {
		mal := m // Prevent loop variable pointer export
		malwareList = append(malwareList, models.Malware{
			MalwareName: utils.OmitEmpty(mal.MalwareName),
			MalwareType: utils.OmitEmpty(mal.MalwareType),
			Path:        utils.OmitEmpty(mal.Path),
			RuleName:    utils.OmitEmpty(mal.RuleName),
		})
	}

	for n, s := range malwareResults.ScansSummary {
		name, summary := n, s // Prevent loop variable pointer export
		metadata = append(metadata, models.ScannerMetadata{
			ScannerName: &name,
			ScannerSummary: &models.ScannerSummary{
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

func ConvertSecretsResultToSecrets(secretsResults *secrets.Results) []models.Secret {
	secretsSlice := []models.Secret{}

	if secretsResults == nil || secretsResults.MergedResults == nil || secretsResults.MergedResults.Results == nil {
		return secretsSlice
	}

	for _, resultsCandidate := range secretsResults.MergedResults.Results {
		for i := range resultsCandidate.Findings {
			finding := resultsCandidate.Findings[i]
			secretsSlice = append(secretsSlice, models.Secret{
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

func ConvertExploitsResultToExploits(exploitsResults *exploits.Results) []models.Exploit {
	retExploits := []models.Exploit{}

	if exploitsResults == nil || exploitsResults.Exploits == nil {
		return retExploits
	}

	for i := range exploitsResults.Exploits {
		exploit := exploitsResults.Exploits[i]
		retExploits = append(retExploits, models.Exploit{
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

func MisconfigurationSeverityToAPIMisconfigurationSeverity(sev misconfigurationTypes.Severity) (models.MisconfigurationSeverity, error) {
	switch sev {
	case misconfigurationTypes.HighSeverity:
		return models.MisconfigurationHighSeverity, nil
	case misconfigurationTypes.MediumSeverity:
		return models.MisconfigurationMediumSeverity, nil
	case misconfigurationTypes.LowSeverity:
		return models.MisconfigurationLowSeverity, nil
	default:
		return models.MisconfigurationLowSeverity, fmt.Errorf("unknown severity level %v", sev)
	}
}

func ConvertMisconfigurationResultToMisconfigurationsAndScanners(misconfigurationResults *misconfiguration.Results) ([]models.Misconfiguration, []string, error) {
	misconfigurations := []models.Misconfiguration{}
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

		misconfigurations = append(misconfigurations, models.Misconfiguration{
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

func ConvertInfoFinderResultToInfosAndScanners(results *infofinder.Results) ([]models.InfoFinderInfo, []string, error) {
	infos := []models.InfoFinderInfo{}

	if results == nil || results.Infos == nil {
		return infos, []string{}, nil
	}

	for i := range results.Infos {
		info := results.Infos[i]

		infos = append(infos, models.InfoFinderInfo{
			Data:        &info.Data,
			Path:        &info.Path,
			ScannerName: &info.ScannerName,
			Type:        convertInfoTypeToAPIModel(info.Type),
		})
	}

	return infos, results.Metadata.Scanners, nil
}

func convertInfoTypeToAPIModel(infoType types.InfoType) *models.InfoType {
	switch infoType {
	case types.SSHKnownHostFingerprint:
		return utils.PointerTo(models.InfoTypeSSHKnownHostFingerprint)
	case types.SSHAuthorizedKeyFingerprint:
		return utils.PointerTo(models.InfoTypeSSHAuthorizedKeyFingerprint)
	case types.SSHPrivateKeyFingerprint:
		return utils.PointerTo(models.InfoTypeSSHPrivateKeyFingerprint)
	case types.SSHDaemonKeyFingerprint:
		return utils.PointerTo(models.InfoTypeSSHDaemonKeyFingerprint)
	default:
		log.Errorf("Can't convert info type %q, treating as %v", infoType, models.InfoTypeUNKNOWN)
		return utils.PointerTo(models.InfoTypeUNKNOWN)
	}
}

func ConvertRootkitsResultToRootkits(rootkitsResults *rootkits.Results) []models.Rootkit {
	rootkitsList := []models.Rootkit{}

	if rootkitsResults == nil || rootkitsResults.MergedResults == nil || rootkitsResults.MergedResults.Rootkits == nil {
		return rootkitsList
	}

	for _, r := range rootkitsResults.MergedResults.Rootkits {
		rootkit := r // Prevent loop variable pointer export
		rootkitsList = append(rootkitsList, models.Rootkit{
			Message:     &rootkit.Message,
			RootkitName: &rootkit.RootkitName,
			RootkitType: ConvertRootkitTypeToAPIModel(rootkit.RootkitType),
		})
	}

	return rootkitsList
}

func ConvertRootkitTypeToAPIModel(rootkitType rootkitsTypes.RootkitType) *models.RootkitType {
	switch rootkitType {
	case rootkitsTypes.APPLICATION:
		return utils.PointerTo(models.RootkitTypeAPPLICATION)
	case rootkitsTypes.FIRMWARE:
		return utils.PointerTo(models.RootkitTypeFIRMWARE)
	case rootkitsTypes.KERNEL:
		return utils.PointerTo(models.RootkitTypeKERNEL)
	case rootkitsTypes.MEMORY:
		return utils.PointerTo(models.RootkitTypeMEMORY)
	case rootkitsTypes.UNKNOWN:
		return utils.PointerTo(models.RootkitTypeUNKNOWN)
	default:
		log.Errorf("Can't convert rootkit type %q, treating as %v", rootkitType, models.RootkitTypeUNKNOWN)
		return utils.PointerTo(models.RootkitTypeUNKNOWN)
	}
}
