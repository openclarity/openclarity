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

package export

import (
	"fmt"
	"strings"

	dockle_types "github.com/Portshift/dockle/pkg/types"
	log "github.com/sirupsen/logrus"
	"github.com/spiegel-im-spiegel/go-cvss/v3/metric"

	"github.com/openclarity/kubeclarity/api/client/client"
	"github.com/openclarity/kubeclarity/api/client/client/operations"
	"github.com/openclarity/kubeclarity/api/client/models"
	"github.com/openclarity/kubeclarity/shared/pkg/scanner"
	cdx_helper "github.com/openclarity/kubeclarity/shared/pkg/utils/cyclonedx_helper"
	"github.com/openclarity/kubeclarity/shared/pkg/utils/image_helper"
	sharedUtilsVulnerability "github.com/openclarity/kubeclarity/shared/pkg/utils/vulnerability"
)

func Export(apiClient *client.KubeClarityAPIs,
	mergedResults *scanner.MergedResults,
	layerCommands []*image_helper.FsLayerCommand,
	cisDockerBenchmarkResults dockle_types.AssessmentMap,
	id string,
) error {
	// create ApplicationVulnerabilityScan from mergedResults
	body := createApplicationVulnerabilityScan(mergedResults, layerCommands, cisDockerBenchmarkResults)
	// create post parameters
	postParams := operations.NewPostApplicationsVulnerabilityScanIDParams().WithID(id).WithBody(body)

	_, err := apiClient.Operations.PostApplicationsVulnerabilityScanID(postParams)
	if err != nil {
		return fmt.Errorf("failed to post vulnerabilify scan results: %v", err)
	}

	return nil
}

func createApplicationVulnerabilityScan(m *scanner.MergedResults,
	layerCommands []*image_helper.FsLayerCommand,
	cisDockerBenchmarkResults dockle_types.AssessmentMap,
) *models.ApplicationVulnerabilityScan {
	return &models.ApplicationVulnerabilityScan{
		Resources: []*models.ResourceVulnerabilityScan{
			{
				PackageVulnerabilities: createPackagesVulnerabilitiesScan(m),
				Resource: &models.ResourceInfo{
					ResourceName: m.Source.Name,
					ResourceType: getResourceType(m),
					ResourceHash: m.Source.Hash,
				},
				ResourceLayerCommands:     createResourceLayerCommands(layerCommands),
				CisDockerBenchmarkResults: createCISDockerBenchmarkResults(cisDockerBenchmarkResults),
			},
		},
	}
}

func createCISDockerBenchmarkResults(results dockle_types.AssessmentMap) []*models.CISDockerBenchmarkCodeInfo {
	// nolint:prealloc
	var ret []*models.CISDockerBenchmarkCodeInfo

	for _, info := range results {
		ret = append(ret, &models.CISDockerBenchmarkCodeInfo{
			Assessments: createCISDockerBenchmarkAssessment(info.Assessments),
			Code:        info.Code,
			Level:       int64(info.Level),
		})
	}

	return ret
}

func createCISDockerBenchmarkAssessment(assessments dockle_types.AssessmentSlice) []*models.CISDockerBenchmarkAssessment {
	// nolint:prealloc
	var ret []*models.CISDockerBenchmarkAssessment

	for _, assessment := range assessments {
		ret = append(ret, &models.CISDockerBenchmarkAssessment{
			Code:     assessment.Code,
			Desc:     assessment.Desc,
			Filename: assessment.Filename,
			Level:    int64(assessment.Level),
		})
	}
	return ret
}

func createResourceLayerCommands(layerCommands []*image_helper.FsLayerCommand) []*models.ResourceLayerCommand {
	if layerCommands == nil {
		return nil
	}
	resourceLayerCommands := make([]*models.ResourceLayerCommand, len(layerCommands))
	for i, layerCommand := range layerCommands {
		resourceLayerCommands[i] = &models.ResourceLayerCommand{
			Command: layerCommand.Command,
			Layer:   layerCommand.Layer,
		}
	}

	return resourceLayerCommands
}

var backendAPISeverity = sharedUtilsVulnerability.SeverityModel[models.VulnerabilitySeverity]{
	Critical:   models.VulnerabilitySeverityCRITICAL,
	High:       models.VulnerabilitySeverityHIGH,
	Medium:     models.VulnerabilitySeverityMEDIUM,
	Low:        models.VulnerabilitySeverityLOW,
	Negligible: models.VulnerabilitySeverityNEGLIGIBLE,
}

func createPackagesVulnerabilitiesScan(m *scanner.MergedResults) []*models.PackageVulnerabilityScan {
	packageVulnerabilityScan := make([]*models.PackageVulnerabilityScan, 0, len(m.MergedVulnerabilitiesByKey))
	for _, vulnerabilities := range m.MergedVulnerabilitiesByKey {
		if len(vulnerabilities) > 1 {
			vulnerabilities = scanner.SortBySeverityAndCVSS(vulnerabilities)
			scanner.PrintIgnoredVulnerabilities(vulnerabilities)
		}
		vulnerability := vulnerabilities[0]
		packageVulnerabilityScan = append(packageVulnerabilityScan, &models.PackageVulnerabilityScan{
			Cvss:              getCVSS(vulnerability.Vulnerability),
			Description:       vulnerability.Vulnerability.Description,
			FixVersion:        scanner.GetFixVersion(vulnerability.Vulnerability),
			LayerID:           vulnerability.Vulnerability.LayerID,
			Links:             vulnerability.Vulnerability.Links,
			Package:           getPackageInfo(vulnerability.Vulnerability),
			Scanners:          getScannerInfo(vulnerability),
			Severity:          backendAPISeverity.GetVulnerabilitySeverityFromString(vulnerability.Vulnerability.Severity),
			VulnerabilityName: vulnerability.Vulnerability.ID,
		})
	}

	return packageVulnerabilityScan
}

func getResourceType(m *scanner.MergedResults) models.ResourceType {
	switch m.Source.Type {
	case cdx_helper.InputTypeImage:
		return models.ResourceTypeIMAGE
	case cdx_helper.InputTypeDirectory:
		return models.ResourceTypeDIRECTORY
	case cdx_helper.InputTypeFile:
		return models.ResourceTypeFILE
	default:
		log.Errorf("Unknown resource type %s", m.Source.Type)
	}
	return ""
}

func getCVSS(vulnerability scanner.Vulnerability) *models.CVSS {
	for _, cvss := range vulnerability.CVSS {
		version := strings.Split(cvss.Version, ".")
		if version[0] == "3" {
			return &models.CVSS{
				CvssV3Metrics: &models.CVSSV3Metrics{
					BaseScore:           cvss.Metrics.BaseScore,
					ExploitabilityScore: *cvss.Metrics.ExploitabilityScore,
					ImpactScore:         *cvss.Metrics.ImpactScore,
				},
				CvssV3Vector: extractCVSSVectorString(cvss.Vector),
			}
		}
	}
	// TODO maybe need to convert cvssv2 vector to cvssv3 vector, and recalculate scores based on it
	if len(vulnerability.CVSS) > 0 {
		log.Infof("CVSS version 3.x is not found, but found cvss version: %s", vulnerability.CVSS[0].Version)
	} else {
		log.Infof("CVSS not found for vulnerability: %s", vulnerability.ID)
	}
	return nil
}

func getPackageInfo(vulnerability scanner.Vulnerability) *models.PackageInfo {
	var license string
	if len(vulnerability.Package.Licenses) > 0 {
		license = vulnerability.Package.Licenses[0]
	}
	return &models.PackageInfo{
		Name:     vulnerability.Package.Name,
		Version:  vulnerability.Package.Version,
		Language: vulnerability.Package.Language,
		License:  license,
	}
}

func getScannerInfo(mergedVulnerability scanner.MergedVulnerability) []string {
	scanners := make([]string, len(mergedVulnerability.ScannersInfo))
	for i, s := range mergedVulnerability.ScannersInfo {
		scanners[i] = s.Name
	}
	return scanners
}

func extractCVSSVectorString(vectorString string) *models.CVSSV3Vector {
	base, err := metric.NewBase().Decode(vectorString)
	if err != nil {
		log.Errorf("Failed to extract informations from CVSS vector string: %v", err)
		return &models.CVSSV3Vector{
			Vector: vectorString,
		}
	}
	return &models.CVSSV3Vector{
		AttackComplexity:   getAttackComplexity(base.AC),
		AttackVector:       getAttackVector(base.AV),
		Availability:       getAvailability(base.A),
		Confidentiality:    getConfidentiality(base.C),
		Integrity:          getIntegrity(base.I),
		PrivilegesRequired: getPrivilegesRequired(base.PR),
		Scope:              getScope(base.S),
		UserInteraction:    getUserInteraction(base.UI),
		Vector:             vectorString,
	}
}

// TODO Every field can be `X` that means NotDefined. Maybe we need to update the swagger.
func getAttackComplexity(attackComplexity metric.AttackComplexity) models.AttackComplexity {
	switch attackComplexity.String() {
	case "H":
		return models.AttackComplexityHIGH
	case "L":
		return models.AttackComplexityLOW
	default:
		return ""
	}
}

func getAttackVector(attackVector metric.AttackVector) models.AttackVector {
	switch attackVector.String() {
	case "L":
		return models.AttackVectorLOCAL
	case "N":
		return models.AttackVectorNETWORK
	case "A":
		return models.AttackVectorADJACENT
	case "P":
		return models.AttackVectorPHYSICAL
	default:
		return ""
	}
}

func getAvailability(availability metric.AvailabilityImpact) models.Availability {
	switch availability.String() {
	case "N":
		return models.AvailabilityNONE
	case "L":
		return models.AvailabilityLOW
	case "H":
		return models.AvailabilityHIGH
	default:
		return ""
	}
}

func getConfidentiality(confidentiality metric.ConfidentialityImpact) models.Confidentiality {
	switch confidentiality.String() {
	case "N":
		return models.ConfidentialityNONE
	case "L":
		return models.ConfidentialityLOW
	case "H":
		return models.ConfidentialityHIGH
	default:
		return ""
	}
}

func getIntegrity(integrity metric.IntegrityImpact) models.Integrity {
	switch integrity.String() {
	case "N":
		return models.IntegrityNONE
	case "L":
		return models.IntegrityLOW
	case "H":
		return models.IntegrityHIGH
	default:
		return ""
	}
}

func getPrivilegesRequired(privilegesRequired metric.PrivilegesRequired) models.PrivilegesRequired {
	switch privilegesRequired.String() {
	case "N":
		return models.PrivilegesRequiredNONE
	case "L":
		return models.PrivilegesRequiredLOW
	case "H":
		return models.PrivilegesRequiredHIGH
	default:
		return ""
	}
}

func getScope(scope metric.Scope) models.Scope {
	switch scope.String() {
	case "U":
		return models.ScopeUNCHANGED
	case "C":
		return models.ScopeCHANGED
	default:
		return ""
	}
}

func getUserInteraction(scope metric.UserInteraction) models.UserInteraction {
	switch scope.String() {
	case "N":
		return models.UserInteractionNONE
	case "R":
		return models.UserInteractionREQUIRED
	default:
		return ""
	}
}
