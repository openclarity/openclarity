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

package report

import (
	"context"
	"fmt"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	transport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	log "github.com/sirupsen/logrus"
	"github.com/spiegel-im-spiegel/go-cvss/v3/metric"

	"github.com/openclarity/kubeclarity/runtime_k8s_scanner/pkg/config"
	"github.com/openclarity/kubeclarity/runtime_scan/api/client/client"
	"github.com/openclarity/kubeclarity/runtime_scan/api/client/client/operations"
	"github.com/openclarity/kubeclarity/runtime_scan/api/client/models"
	"github.com/openclarity/kubeclarity/shared/pkg/analyzer"
	"github.com/openclarity/kubeclarity/shared/pkg/scanner"
	cdx_helper "github.com/openclarity/kubeclarity/shared/pkg/utils/cyclonedx_helper"
	"github.com/openclarity/kubeclarity/shared/pkg/utils/image_helper"
	sharedUtilsVulnerability "github.com/openclarity/kubeclarity/shared/pkg/utils/vulnerability"
)

type Reporter interface {
	ReportScanResults(mergedResults *scanner.MergedResults, layerCommands []*image_helper.FsLayerCommand) error
	ReportScanError(scanError *models.ScanError) error
	ReportScanContentAnalysis(mergedResults *analyzer.MergedResults) error
}

// nolint:containedctx
type ReporterImpl struct {
	client *client.KubeClarityRuntimeScanAPIs
	conf   *config.Config
	ctx    context.Context
}

func CreateReporter(ctx context.Context, conf *config.Config) Reporter {
	cfg := client.DefaultTransportConfig().WithHost(conf.ResultServiceAddress)

	return &ReporterImpl{
		client: client.New(transport.New(cfg.Host, cfg.BasePath, cfg.Schemes), strfmt.Default),
		conf:   conf,
		ctx:    ctx,
	}
}

/* VulnerabilityScan Report Start */

func (r *ReporterImpl) ReportScanError(scanError *models.ScanError) error {
	_, err := r.client.Operations.PostScanScanUUIDResults(operations.NewPostScanScanUUIDResultsParams().
		WithScanUUID(strfmt.UUID(r.conf.ScanUUID)).
		WithBody(r.createFailedImageVulnerabilityScan(scanError)).
		WithContext(r.ctx))
	if err != nil {
		return fmt.Errorf("failed to report scan error: %v", err)
	}

	return nil
}

func (r *ReporterImpl) ReportScanResults(
	mergedResults *scanner.MergedResults,
	layerCommands []*image_helper.FsLayerCommand,
) error {
	_, err := r.client.Operations.PostScanScanUUIDResults(operations.NewPostScanScanUUIDResultsParams().
		WithScanUUID(strfmt.UUID(r.conf.ScanUUID)).
		WithBody(r.createSuccessfulImageVulnerabilityScan(mergedResults, layerCommands)).
		WithContext(r.ctx))
	if err != nil {
		return fmt.Errorf("failed to report scan results: %v", err)
	}

	return nil
}

func (r *ReporterImpl) createFailedImageVulnerabilityScan(scanError *models.ScanError) *models.ImageVulnerabilityScan {
	return &models.ImageVulnerabilityScan{
		ImageID: r.conf.ImageIDToScan,
		ResourceVulnerabilityScan: &models.ResourceVulnerabilityScan{
			Error:    scanError,
			Resource: r.createResourceInfo(),
			Status:   models.ScanStatusFAILED,
		},
	}
}

func (r *ReporterImpl) createSuccessfulImageVulnerabilityScan(
	results *scanner.MergedResults,
	layerCommands []*image_helper.FsLayerCommand,
) *models.ImageVulnerabilityScan {
	return &models.ImageVulnerabilityScan{
		ImageID: r.conf.ImageIDToScan,
		ResourceVulnerabilityScan: &models.ResourceVulnerabilityScan{
			PackageVulnerabilities: createPackagesVulnerabilitiesScan(results),
			ResourceLayerCommands:  createResourceLayerCommands(layerCommands),
			Resource:               r.createResourceInfo(),
			Status:                 models.ScanStatusSUCCESS,
		},
	}
}

func (r *ReporterImpl) createResourceInfo() *models.ResourceInfo {
	return &models.ResourceInfo{
		ResourceHash: r.conf.ImageHashToScan,
		ResourceName: r.conf.ImageNameToScan,
		ResourceType: models.ResourceTypeIMAGE,
	}
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

var runtimeAPISeverity = sharedUtilsVulnerability.SeverityModel[models.VulnerabilitySeverity]{
	Critical:   models.VulnerabilitySeverityCRITICAL,
	High:       models.VulnerabilitySeverityHIGH,
	Medium:     models.VulnerabilitySeverityMEDIUM,
	Low:        models.VulnerabilitySeverityLOW,
	Negligible: models.VulnerabilitySeverityNEGLIGIBLE,
}

func createPackagesVulnerabilitiesScan(results *scanner.MergedResults) []*models.PackageVulnerabilityScan {
	packageVulnerabilityScan := make([]*models.PackageVulnerabilityScan, 0, len(results.MergedVulnerabilitiesByKey))
	for _, vulnerabilities := range results.MergedVulnerabilitiesByKey {
		if len(vulnerabilities) > 1 {
			vulnerabilities = scanner.SortBySeverityAndCVSS(vulnerabilities)
			scanner.PrintIgnoredVulnerabilities(vulnerabilities)
		}
		vulnerability := vulnerabilities[0]
		packageVulnerabilityScan = append(packageVulnerabilityScan, &models.PackageVulnerabilityScan{
			Cvss:              getCVSS(vulnerability.Vulnerability),
			Description:       vulnerability.Vulnerability.Description,
			FixVersion:        getFixVersion(vulnerability.Vulnerability),
			LayerID:           vulnerability.Vulnerability.LayerID,
			Links:             vulnerability.Vulnerability.Links,
			Package:           getPackageInfo(vulnerability.Vulnerability),
			Scanners:          getScannerInfo(vulnerability),
			Severity:          runtimeAPISeverity.GetVulnerabilitySeverityFromString(vulnerability.Vulnerability.Severity),
			VulnerabilityName: vulnerability.Vulnerability.ID,
		})
	}

	return packageVulnerabilityScan
}

// TODO: Do we want to unified the convert logic below with the logic in cli/pkg/scanner/export/export.go
// Same logic but returns different models
// We can use the generate the shared types and create a converter for each model, not sure it is better.
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

// TODO can be multiple fix version?
func getFixVersion(vulnerability scanner.Vulnerability) string {
	if len(vulnerability.Fix.Versions) > 0 {
		return vulnerability.Fix.Versions[0]
	}
	return ""
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

/* VulnerabilityScan Report End */

/* ContentAnalysis Report Start */

func (r *ReporterImpl) ReportScanContentAnalysis(mergedResults *analyzer.MergedResults) error {
	_, err := r.client.Operations.PostScanScanUUIDContentAnalysis(operations.NewPostScanScanUUIDContentAnalysisParams().
		WithScanUUID(strfmt.UUID(r.conf.ScanUUID)).
		WithBody(r.createImageContentAnalysis(mergedResults)).
		WithContext(r.ctx))
	if err != nil {
		return fmt.Errorf("failed to report scan content analysis: %v", err)
	}

	return nil
}

func (r *ReporterImpl) createImageContentAnalysis(results *analyzer.MergedResults) *models.ImageContentAnalysis {
	return &models.ImageContentAnalysis{
		ImageID: r.conf.ImageIDToScan,
		ResourceContentAnalysis: &models.ResourceContentAnalysis{
			Packages: r.createPackagesContentAnalysis(results),
			Resource: r.createResourceInfo(),
		},
	}
}

func (r *ReporterImpl) createPackagesContentAnalysis(m *analyzer.MergedResults) []*models.PackageContentAnalysis {
	packageContentAnalysis := make([]*models.PackageContentAnalysis, 0, len(m.MergedComponentByKey))
	for _, component := range m.MergedComponentByKey {
		packageContentAnalysis = append(packageContentAnalysis, &models.PackageContentAnalysis{
			Analyzers: component.AnalyzerInfo,
			Package:   getContentAnalysisPackageInfo(component.Component),
		})
	}
	return packageContentAnalysis
}

func getContentAnalysisPackageInfo(component cdx.Component) *models.PackageInfo {
	// In the case of packages there can be only one license in the licence list
	var license string
	licenses := cdx_helper.GetComponentLicenses(component)
	if len(licenses) > 0 {
		license = licenses[0]
	}
	return &models.PackageInfo{
		Name:     component.Name,
		Version:  component.Version,
		License:  license,
		Language: cdx_helper.GetComponentLanguage(component),
	}
}

/* ContentAnalysis Report End */
