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

package cmd

import (
	"context"
	"fmt"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/shared/pkg/backendclient"
	"github.com/openclarity/vmclarity/shared/pkg/families"
	"github.com/openclarity/vmclarity/shared/pkg/families/exploits"
	"github.com/openclarity/vmclarity/shared/pkg/families/results"
	"github.com/openclarity/vmclarity/shared/pkg/families/sbom"
	"github.com/openclarity/vmclarity/shared/pkg/families/secrets"
	"github.com/openclarity/vmclarity/shared/pkg/families/types"
	"github.com/openclarity/vmclarity/shared/pkg/families/vulnerabilities"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

type Exporter struct {
	client *backendclient.BackendClient
}

func CreateExporter() (*Exporter, error) {
	client, err := backendclient.Create(server)
	if err != nil {
		return nil, fmt.Errorf("unable to create VMClarity API client. server=%v: %w", server, err)
	}

	return &Exporter{
		client: client,
	}, nil
}

func convertSBOMResultToAPIModel(sbomResults *sbom.Results) *models.SbomScan {
	packages := []models.Package{}

	if sbomResults.SBOM.Components != nil {
		for _, component := range *sbomResults.SBOM.Components {
			pkg := models.Package{
				Id: utils.StringPtr(component.BOMRef),
				PackageInfo: &models.PackageInfo{
					PackageName:    utils.StringPtr(component.Name),
					PackageVersion: utils.StringPtr(component.Version),
				},
			}
			packages = append(packages, pkg)
		}
	}

	return &models.SbomScan{
		Packages: &packages,
	}
}

func convertVulnResultToAPIModel(vulnerabilitiesResults *vulnerabilities.Results) *models.VulnerabilityScan {
	vulnerabilities := []models.Vulnerability{}
	for _, vulCandidates := range vulnerabilitiesResults.MergedResults.MergedVulnerabilitiesByKey {
		if len(vulCandidates) < 1 {
			continue
		}

		vulCandidate := vulCandidates[0]

		vul := models.Vulnerability{
			Id: utils.StringPtr(vulCandidate.ID),
			VulnerabilityInfo: &models.VulnerabilityInfo{
				VulnerabilityName: utils.StringPtr(vulCandidate.Vulnerability.ID),
				Description:       utils.StringPtr(vulCandidate.Vulnerability.Description),
				Severity:          (*models.VulnerabilitySeverity)(utils.StringPtr(vulCandidate.Vulnerability.Severity)),
			},
		}
		vulnerabilities = append(vulnerabilities, vul)
	}

	return &models.VulnerabilityScan{
		Vulnerabilities: &vulnerabilities,
	}
}

func (e *Exporter) MarkScanResultInProgress() error {
	scanResult, err := e.client.GetScanResult(context.TODO(), scanResultID, models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get scan result: %w", err)
	}

	if scanResult.Status == nil {
		scanResult.Status = &models.TargetScanStatus{}
	}
	if scanResult.Status.General == nil {
		scanResult.Status.General = &models.TargetScanState{}
	}

	state := models.INPROGRESS
	scanResult.Status.General.State = &state

	err = e.client.PatchScanResult(context.TODO(), scanResult, scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (e *Exporter) MarkScanResultDone(errors []error) error {
	scanResult, err := e.client.GetScanResult(context.TODO(), scanResultID, models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get scan result: %w", err)
	}

	if scanResult.Status == nil {
		scanResult.Status = &models.TargetScanStatus{}
	}
	if scanResult.Status.General == nil {
		scanResult.Status.General = &models.TargetScanState{}
	}

	state := models.DONE
	scanResult.Status.General.State = &state

	// If we had any errors running the family or exporting results add it
	// to the general errors
	if len(errors) > 0 {
		var errorStrs []string
		// Pull the errors list out so that we can append to it (if there are
		// any errors at this point I would have hoped the orcestrator wouldn't
		// have spawned the VM) but we never know.
		if scanResult.Status.General.Errors != nil {
			errorStrs = *scanResult.Status.General.Errors
		}
		for _, err := range errors {
			if err != nil {
				errorStrs = append(errorStrs, err.Error())
			}
		}
		if len(errorStrs) > 0 {
			scanResult.Status.General.Errors = &errorStrs
		}
	}

	err = e.client.PatchScanResult(context.TODO(), scanResult, scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (e *Exporter) ExportSbomResult(res *results.Results, famerr families.RunErrors) error {
	scanResult, err := e.client.GetScanResult(context.TODO(), scanResultID, models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get scan result: %w", err)
	}

	if scanResult.Status == nil {
		scanResult.Status = &models.TargetScanStatus{}
	}
	if scanResult.Status.Sbom == nil {
		scanResult.Status.Sbom = &models.TargetScanState{}
	}
	if scanResult.Summary == nil {
		scanResult.Summary = &models.TargetScanResultSummary{}
	}

	var errors []string

	if err, ok := famerr[types.SBOM]; ok {
		errors = append(errors, err.Error())
	} else {
		sbomResults, err := results.GetResult[*sbom.Results](res)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to get sbom from scan: %w", err).Error())
		} else {
			scanResult.Sboms = convertSBOMResultToAPIModel(sbomResults)
		}
		scanResult.Summary.TotalPackages = utils.PointerTo[int](len(*scanResult.Sboms.Packages))
	}

	state := models.DONE
	scanResult.Status.Sbom.State = &state
	scanResult.Status.Sbom.Errors = &errors

	err = e.client.PatchScanResult(context.TODO(), scanResult, scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (e *Exporter) ExportVulResult(res *results.Results, famerr families.RunErrors) error {
	scanResult, err := e.client.GetScanResult(context.TODO(), scanResultID, models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get scan result: %w", err)
	}

	if scanResult.Status == nil {
		scanResult.Status = &models.TargetScanStatus{}
	}
	if scanResult.Status.Vulnerabilities == nil {
		scanResult.Status.Vulnerabilities = &models.TargetScanState{}
	}
	if scanResult.Summary == nil {
		scanResult.Summary = &models.TargetScanResultSummary{}
	}

	var errors []string

	if err, ok := famerr[types.Vulnerabilities]; ok {
		errors = append(errors, err.Error())
	} else {
		vulnerabilitiesResults, err := results.GetResult[*vulnerabilities.Results](res)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to get vulnerabilities from scan: %w", err).Error())
		} else {
			scanResult.Vulnerabilities = convertVulnResultToAPIModel(vulnerabilitiesResults)
		}
		scanResult.Summary.TotalVulnerabilities = getVulnerabilityTotalsPerSeverity(scanResult.Vulnerabilities.Vulnerabilities)
	}

	state := models.DONE
	scanResult.Status.Vulnerabilities.State = &state
	scanResult.Status.Vulnerabilities.Errors = &errors

	err = e.client.PatchScanResult(context.TODO(), scanResult, scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func getVulnerabilityTotalsPerSeverity(vulnerabilities *[]models.Vulnerability) *models.VulnerabilityScanSummary {
	ret := &models.VulnerabilityScanSummary{
		TotalCriticalVulnerabilities:   utils.PointerTo[int](0),
		TotalHighVulnerabilities:       utils.PointerTo[int](0),
		TotalMediumVulnerabilities:     utils.PointerTo[int](0),
		TotalLowVulnerabilities:        utils.PointerTo[int](0),
		TotalNegligibleVulnerabilities: utils.PointerTo[int](0),
	}
	if vulnerabilities == nil {
		return ret
	}
	for _, vulnerability := range *vulnerabilities {
		switch *vulnerability.VulnerabilityInfo.Severity {
		case models.CRITICAL:
			ret.TotalCriticalVulnerabilities = utils.PointerTo[int](*ret.TotalCriticalVulnerabilities + 1)
		case models.HIGH:
			ret.TotalHighVulnerabilities = utils.PointerTo[int](*ret.TotalHighVulnerabilities + 1)
		case models.MEDIUM:
			ret.TotalMediumVulnerabilities = utils.PointerTo[int](*ret.TotalMediumVulnerabilities + 1)
		case models.LOW:
			ret.TotalLowVulnerabilities = utils.PointerTo[int](*ret.TotalLowVulnerabilities + 1)
		case models.NEGLIGIBLE:
			ret.TotalNegligibleVulnerabilities = utils.PointerTo[int](*ret.TotalNegligibleVulnerabilities + 1)
		}
	}
	return ret
}

func (e *Exporter) ExportSecretsResult(res *results.Results, famerr families.RunErrors) error {
	scanResult, err := e.client.GetScanResult(context.TODO(), scanResultID, models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get scan result: %w", err)
	}

	if scanResult.Status == nil {
		scanResult.Status = &models.TargetScanStatus{}
	}
	if scanResult.Status.Secrets == nil {
		scanResult.Status.Secrets = &models.TargetScanState{}
	}
	if scanResult.Summary == nil {
		scanResult.Summary = &models.TargetScanResultSummary{}
	}

	var errors []string

	if err, ok := famerr[types.Secrets]; ok {
		errors = append(errors, err.Error())
	} else {
		secretsResults, err := results.GetResult[*secrets.Results](res)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to get secrets results from scan: %w", err).Error())
		} else {
			scanResult.Secrets = convertSecretsResultToAPIModel(secretsResults)
		}
		scanResult.Summary.TotalSecrets = utils.PointerTo[int](len(*scanResult.Secrets.Secrets))
	}

	state := models.DONE
	scanResult.Status.Secrets.State = &state
	scanResult.Status.Secrets.Errors = &errors

	err = e.client.PatchScanResult(context.TODO(), scanResult, scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func convertSecretsResultToAPIModel(secretsResults *secrets.Results) *models.SecretScan {
	if secretsResults == nil || secretsResults.MergedResults == nil {
		return &models.SecretScan{}
	}

	var secretsSlice []models.Secret
	for _, resultsCandidate := range secretsResults.MergedResults.Results {
		for i := range resultsCandidate.Findings {
			finding := resultsCandidate.Findings[i]
			secretsSlice = append(secretsSlice, models.Secret{
				SecretInfo: &models.SecretInfo{
					Description: &finding.Description,
					EndLine:     &finding.EndLine,
					FilePath:    &finding.File,
					Fingerprint: &finding.Fingerprint,
					StartLine:   &finding.StartLine,
				},
				Id: &finding.Fingerprint, // TODO: Do we need the ID in the secret?
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

func convertExploitsResultToAPIModel(exploitsResults *exploits.Results) *models.ExploitScan {
	if exploitsResults == nil || exploitsResults.Exploits == nil {
		return &models.ExploitScan{}
	}

	// nolint:prealloc
	var retExploits []models.Exploit

	for i := range exploitsResults.Exploits {
		exploit := exploitsResults.Exploits[i]
		retExploits = append(retExploits, models.Exploit{
			ExploitInfo: &models.ExploitInfo{
				CveID:       &exploit.CveID,
				Description: &exploit.Description,
				Name:        &exploit.Name,
				SourceDB:    &exploit.SourceDB,
				Title:       &exploit.Title,
				Urls:        &exploit.URLs,
			},
			Id: &exploit.ID,
		})
	}

	if retExploits == nil {
		return &models.ExploitScan{}
	}

	return &models.ExploitScan{
		Exploits: &retExploits,
	}
}

func (e *Exporter) ExportExploitsResult(res *results.Results) error {
	scanResult, err := e.client.GetScanResult(context.TODO(), scanResultID, models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get scan result: %w", err)
	}

	if scanResult.Status == nil {
		scanResult.Status = &models.TargetScanStatus{}
	}
	if scanResult.Status.Exploits == nil {
		scanResult.Status.Exploits = &models.TargetScanState{}
	}
	if scanResult.Summary == nil {
		scanResult.Summary = &models.TargetScanResultSummary{}
	}

	var errors []string

	exploitsResults, err := results.GetResult[*exploits.Results](res)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to get exploits results from scan: %w", err).Error())
	} else {
		scanResult.Exploits = convertExploitsResultToAPIModel(exploitsResults)
		scanResult.Summary.TotalExploits = utils.PointerTo[int](len(*scanResult.Exploits.Exploits))
	}

	state := models.DONE
	scanResult.Status.Exploits.State = &state
	scanResult.Status.Exploits.Errors = &errors

	err = e.client.PatchScanResult(context.TODO(), scanResult, scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (e *Exporter) ExportResults(res *results.Results, famerr families.RunErrors) []error {
	var errors []error
	if config.SBOM.Enabled {
		err := e.ExportSbomResult(res, famerr)
		if err != nil {
			err = fmt.Errorf("failed to export sbom to server: %w", err)
			logger.Error(err)
			errors = append(errors, err)
		}
	}

	if config.Vulnerabilities.Enabled {
		err := e.ExportVulResult(res, famerr)
		if err != nil {
			err = fmt.Errorf("failed to export vulnerabilties to server: %w", err)
			logger.Error(err)
			errors = append(errors, err)
		}
	}

	if config.Secrets.Enabled {
		err := e.ExportSecretsResult(res, famerr)
		if err != nil {
			err = fmt.Errorf("failed to export secrets findings to server: %w", err)
			logger.Error(err)
			errors = append(errors, err)
		}
	}

	if config.Exploits.Enabled {
		err := e.ExportExploitsResult(res)
		if err != nil {
			err = fmt.Errorf("failed to export exploits results to server: %w", err)
			logger.Error(err)
			errors = append(errors, err)
		}
	}

	return errors
}
