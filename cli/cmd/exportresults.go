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
	"net/http"

	"github.com/openclarity/vmclarity/api/client"
	"github.com/openclarity/vmclarity/api/models"
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
	apiClient client.ClientWithResponsesInterface
}

func CreateExporter() (*Exporter, error) {
	apiClient, err := client.NewClientWithResponses(server)
	if err != nil {
		return nil, fmt.Errorf("unable to create VMClarity API client. server=%v: %w", server, err)
	}

	return &Exporter{
		apiClient: apiClient,
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
			},
		}
		vulnerabilities = append(vulnerabilities, vul)
	}

	return &models.VulnerabilityScan{
		Vulnerabilities: &vulnerabilities,
	}
}

// nolint:cyclop
func (e *Exporter) getExistingScanResult() (models.TargetScanResult, error) {
	newGetExistingError := func(err error) error {
		return fmt.Errorf("failed to get existing scan result %v: %w", scanResultID, err)
	}

	var scanResults models.TargetScanResult
	resp, err := e.apiClient.GetScanResultsScanResultIDWithResponse(context.TODO(), scanResultID, &models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return scanResults, newGetExistingError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return scanResults, newGetExistingError(fmt.Errorf("empty body"))
		}
		return *resp.JSON200, nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return scanResults, newGetExistingError(fmt.Errorf("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return scanResults, newGetExistingError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return scanResults, newGetExistingError(fmt.Errorf("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return scanResults, newGetExistingError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return scanResults, newGetExistingError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

// nolint:cyclop
func (e *Exporter) patchExistingScanResult(scanResults models.TargetScanResult) error {
	newUpdateScanResultError := func(err error) error {
		return fmt.Errorf("failed to update scan result %v on server %v: %w", scanResultID, server, err)
	}

	resp, err := e.apiClient.PatchScanResultsScanResultIDWithResponse(context.TODO(), scanResultID, scanResults)
	if err != nil {
		return newUpdateScanResultError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return newUpdateScanResultError(fmt.Errorf("empty body"))
		}
		return nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return newUpdateScanResultError(fmt.Errorf("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return newUpdateScanResultError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return newUpdateScanResultError(fmt.Errorf("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return newUpdateScanResultError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return newUpdateScanResultError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func (e *Exporter) MarkScanResultInProgress() error {
	scanResults, err := e.getExistingScanResult()
	if err != nil {
		return err
	}

	if scanResults.Status == nil {
		scanResults.Status = &models.TargetScanStatus{}
	}
	if scanResults.Status.General == nil {
		scanResults.Status.General = &models.TargetScanState{}
	}

	state := models.INPROGRESS
	scanResults.Status.General.State = &state

	err = e.patchExistingScanResult(scanResults)
	if err != nil {
		return err
	}

	return nil
}

func (e *Exporter) MarkScanResultDone(errors []error) error {
	scanResults, err := e.getExistingScanResult()
	if err != nil {
		return err
	}

	if scanResults.Status == nil {
		scanResults.Status = &models.TargetScanStatus{}
	}
	if scanResults.Status.General == nil {
		scanResults.Status.General = &models.TargetScanState{}
	}

	state := models.DONE
	scanResults.Status.General.State = &state

	// If we had any errors running the family or exporting results add it
	// to the general errors
	if len(errors) > 0 {
		var errorStrs []string
		// Pull the errors list out so that we can append to it (if there are
		// any errors at this point I would have hoped the orcestrator wouldn't
		// have spawned the VM) but we never know.
		if scanResults.Status.General.Errors != nil {
			errorStrs = *scanResults.Status.General.Errors
		}
		for _, err := range errors {
			if err != nil {
				errorStrs = append(errorStrs, err.Error())
			}
		}
		if len(errorStrs) > 0 {
			scanResults.Status.General.Errors = &errorStrs
		}
	}

	err = e.patchExistingScanResult(scanResults)
	if err != nil {
		return err
	}

	return nil
}

func (e *Exporter) ExportSbomResult(res *results.Results, famerr families.RunErrors) error {
	scanResults, err := e.getExistingScanResult()
	if err != nil {
		return err
	}

	if scanResults.Status == nil {
		scanResults.Status = &models.TargetScanStatus{}
	}
	if scanResults.Status.Sbom == nil {
		scanResults.Status.Sbom = &models.TargetScanState{}
	}

	var errors []string

	if err, ok := famerr[types.SBOM]; ok {
		errors = append(errors, err.Error())
	} else {
		sbomResults, err := results.GetResult[*sbom.Results](res)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to get sbom from scan: %w", err).Error())
		} else {
			scanResults.Sboms = convertSBOMResultToAPIModel(sbomResults)
		}
	}

	state := models.DONE
	scanResults.Status.Sbom.State = &state
	scanResults.Status.Sbom.Errors = &errors

	err = e.patchExistingScanResult(scanResults)
	if err != nil {
		return err
	}

	return nil
}

func (e *Exporter) ExportVulResult(res *results.Results, famerr families.RunErrors) error {
	scanResults, err := e.getExistingScanResult()
	if err != nil {
		return err
	}

	if scanResults.Status == nil {
		scanResults.Status = &models.TargetScanStatus{}
	}
	if scanResults.Status.Vulnerabilities == nil {
		scanResults.Status.Vulnerabilities = &models.TargetScanState{}
	}

	var errors []string

	if err, ok := famerr[types.Vulnerabilities]; ok {
		errors = append(errors, err.Error())
	} else {
		vulnerabilitiesResults, err := results.GetResult[*vulnerabilities.Results](res)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to get vulnerabilities from scan: %w", err).Error())
		} else {
			scanResults.Vulnerabilities = convertVulnResultToAPIModel(vulnerabilitiesResults)
		}
	}

	state := models.DONE
	scanResults.Status.Vulnerabilities.State = &state
	scanResults.Status.Vulnerabilities.Errors = &errors

	err = e.patchExistingScanResult(scanResults)
	if err != nil {
		return err
	}

	return nil
}

func (e *Exporter) ExportSecretsResult(res *results.Results, famerr families.RunErrors) error {
	scanResults, err := e.getExistingScanResult()
	if err != nil {
		return err
	}

	if scanResults.Status == nil {
		scanResults.Status = &models.TargetScanStatus{}
	}
	if scanResults.Status.Secrets == nil {
		scanResults.Status.Secrets = &models.TargetScanState{}
	}

	var errors []string

	if err, ok := famerr[types.Secrets]; ok {
		errors = append(errors, err.Error())
	} else {
		secretsResults, err := results.GetResult[*secrets.Results](res)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to get secrets results from scan: %w", err).Error())
		} else {
			scanResults.Secrets = convertSecretsResultToAPIModel(secretsResults)
		}
	}

	state := models.DONE
	scanResults.Status.Secrets.State = &state
	scanResults.Status.Secrets.Errors = &errors

	err = e.patchExistingScanResult(scanResults)
	if err != nil {
		return err
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
	scanResults, err := e.getExistingScanResult()
	if err != nil {
		return err
	}

	if scanResults.Status == nil {
		scanResults.Status = &models.TargetScanStatus{}
	}
	if scanResults.Status.Exploits == nil {
		scanResults.Status.Exploits = &models.TargetScanState{}
	}

	var errors []string

	exploitsResults, err := results.GetResult[*exploits.Results](res)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to get exploits results from scan: %w", err).Error())
	} else {
		scanResults.Exploits = convertExploitsResultToAPIModel(exploitsResults)
	}

	state := models.DONE
	scanResults.Status.Exploits.State = &state
	scanResults.Status.Exploits.Errors = &errors

	err = e.patchExistingScanResult(scanResults)
	if err != nil {
		return err
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
