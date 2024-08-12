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

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/openclarity/kubeclarity/shared/pkg/scanner"
	"github.com/openclarity/kubeclarity/shared/pkg/utils/cyclonedx_helper"
	"github.com/openclarity/kubeclarity/shared/pkg/utils/vulnerability"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/shared/pkg/families/exploits"
	"github.com/openclarity/vmclarity/shared/pkg/families/malware"
	"github.com/openclarity/vmclarity/shared/pkg/families/misconfiguration"
	misconfigurationTypes "github.com/openclarity/vmclarity/shared/pkg/families/misconfiguration/types"
	"github.com/openclarity/vmclarity/shared/pkg/families/rootkits"
	rootkitsTypes "github.com/openclarity/vmclarity/shared/pkg/families/rootkits/types"
	"github.com/openclarity/vmclarity/shared/pkg/families/sbom"
	"github.com/openclarity/vmclarity/shared/pkg/families/secrets"
	"github.com/openclarity/vmclarity/shared/pkg/families/vulnerabilities"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

func ConvertSBOMResultToAPIModel(sbomResults *sbom.Results) *models.SbomScan {
	packages := []models.Package{}

	if sbomResults.SBOM.Components != nil {
		for _, component := range *sbomResults.SBOM.Components {
			packages = append(packages, *ConvertPackageInfoToAPIModel(component))
		}
	}

	return &models.SbomScan{
		Packages: &packages,
	}
}

func ConvertPackageInfoToAPIModel(component cdx.Component) *models.Package {
	return &models.Package{
		Cpes:     utils.PointerTo([]string{component.CPE}),
		Language: utils.PointerTo(cyclonedx_helper.GetComponentLanguage(component)),
		Licenses: utils.PointerTo(cyclonedx_helper.GetComponentLicenses(component)),
		Name:     utils.PointerTo(component.Name),
		Purl:     utils.PointerTo(component.PackageURL),
		Type:     utils.PointerTo(string(component.Type)),
		Version:  utils.PointerTo(component.Version),
	}
}

func ConvertVulnResultToAPIModel(vulnerabilitiesResults *vulnerabilities.Results) *models.VulnerabilityScan {
	// nolint:prealloc
	var vuls []models.Vulnerability
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

	return &models.VulnerabilityScan{
		Vulnerabilities: &vuls,
	}
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

func ConvertMalwareResultToAPIModel(malwareResults *malware.MergedResults) *models.MalwareScan {
	if malwareResults == nil {
		return &models.MalwareScan{}
	}

	malwareList := []models.Malware{}
	for _, m := range malwareResults.DetectedMalware {
		mal := m // Prevent loop variable pointer export
		malwareList = append(malwareList, models.Malware{
			MalwareName: &mal.MalwareName,
			MalwareType: &mal.MalwareType,
			Path:        &mal.Path,
		})
	}

	metadata := []models.ScannerMetadata{}
	for name, summary := range malwareResults.Metadata {
		nameVal := name // Prevent loop variable pointer export
		metadata = append(metadata, models.ScannerMetadata{
			ScannerName: &nameVal,
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

	return &models.MalwareScan{
		Malware:  &malwareList,
		Metadata: &metadata,
	}
}

func ConvertSecretsResultToAPIModel(secretsResults *secrets.Results) *models.SecretScan {
	if secretsResults == nil || secretsResults.MergedResults == nil {
		return &models.SecretScan{}
	}

	var secretsSlice []models.Secret
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

	if secretsSlice == nil {
		return &models.SecretScan{}
	}

	return &models.SecretScan{
		Secrets: &secretsSlice,
	}
}

func ConvertExploitsResultToAPIModel(exploitsResults *exploits.Results) *models.ExploitScan {
	if exploitsResults == nil || exploitsResults.Exploits == nil {
		return &models.ExploitScan{}
	}

	// nolint:prealloc
	var retExploits []models.Exploit

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

	if retExploits == nil {
		return &models.ExploitScan{}
	}

	return &models.ExploitScan{
		Exploits: &retExploits,
	}
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

func ConvertMisconfigurationResultToAPIModel(misconfigurationResults *misconfiguration.Results) (*models.MisconfigurationScan, error) {
	if misconfigurationResults == nil || misconfigurationResults.Misconfigurations == nil {
		return &models.MisconfigurationScan{}, nil
	}

	retMisconfigurations := make([]models.Misconfiguration, len(misconfigurationResults.Misconfigurations))

	for i := range misconfigurationResults.Misconfigurations {
		// create a separate variable for the loop because we need
		// pointers for the API model and we can't safely take pointers
		// to a loop variable.
		misconfig := misconfigurationResults.Misconfigurations[i]

		severity, err := MisconfigurationSeverityToAPIMisconfigurationSeverity(misconfig.Severity)
		if err != nil {
			return nil, fmt.Errorf("unable to convert scanner result severity to API severity: %w", err)
		}

		retMisconfigurations[i] = models.Misconfiguration{
			ScannerName:     &misconfig.ScannerName,
			ScannedPath:     &misconfig.ScannedPath,
			TestCategory:    &misconfig.TestCategory,
			TestID:          &misconfig.TestID,
			TestDescription: &misconfig.TestDescription,
			Severity:        &severity,
			Message:         &misconfig.Message,
			Remediation:     &misconfig.Remediation,
		}
	}

	return &models.MisconfigurationScan{
		Scanners:          utils.PointerTo(misconfigurationResults.Metadata.Scanners),
		Misconfigurations: &retMisconfigurations,
	}, nil
}

func ConvertRootkitsResultToAPIModel(rootkitsResults *rootkits.Results) *models.RootkitScan {
	if rootkitsResults == nil || rootkitsResults.MergedResults == nil {
		return &models.RootkitScan{}
	}

	rootkitsList := []models.Rootkit{}
	for _, r := range rootkitsResults.MergedResults.Rootkits {
		rootkit := r // Prevent loop variable pointer export
		rootkitsList = append(rootkitsList, models.Rootkit{
			Message:     &rootkit.Message,
			RootkitName: &rootkit.RootkitName,
			RootkitType: ConvertRootkitTypeToAPIModel(rootkit.RootkitType),
		})
	}

	return &models.RootkitScan{
		Rootkits: &rootkitsList,
	}
}

func ConvertRootkitTypeToAPIModel(rootkitType rootkitsTypes.RootkitType) *models.RootkitType {
	switch rootkitType {
	case rootkitsTypes.APPLICATION:
		return utils.PointerTo(models.APPLICATION)
	case rootkitsTypes.FIRMWARE:
		return utils.PointerTo(models.FIRMWARE)
	case rootkitsTypes.KERNEL:
		return utils.PointerTo(models.KERNEL)
	case rootkitsTypes.MEMORY:
		return utils.PointerTo(models.MEMORY)
	case rootkitsTypes.UNKNOWN:
		return utils.PointerTo(models.UNKNOWN)
	default:
		log.Errorf("Can't convert rootkit type %q, treating as %v", rootkitType, models.UNKNOWN)
		return utils.PointerTo(models.UNKNOWN)
	}
}
