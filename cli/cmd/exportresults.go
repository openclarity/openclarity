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
	"github.com/openclarity/vmclarity/shared/pkg/families/results"
	"github.com/openclarity/vmclarity/shared/pkg/families/sbom"
	"github.com/openclarity/vmclarity/shared/pkg/families/vulnerabilities"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

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
				Id:                utils.StringPtr(vulCandidate.Vulnerability.ID),
				VulnerabilityName: utils.StringPtr(vulCandidate.Vulnerability.Description),
			},
		}
		vulnerabilities = append(vulnerabilities, vul)
	}

	return &models.VulnerabilityScan{
		Vulnerabilities: &vulnerabilities,
	}
}

// nolint:cyclop
func getExistingScanResult(apiClient client.ClientWithResponsesInterface) (models.TargetScanResult, error) {
	newGetExistingError := func(err error) error {
		return fmt.Errorf("failed to get existing scan result %v: %w", scanResultID, err)
	}

	var scanResults models.TargetScanResult
	resp, err := apiClient.GetScanResultsScanResultIDWithResponse(context.TODO(), scanResultID, &models.GetScanResultsScanResultIDParams{})
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
			return scanResults, newGetExistingError(fmt.Errorf("not found: %v", resp.JSON404.Message))
		}
		return scanResults, newGetExistingError(fmt.Errorf("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return scanResults, newGetExistingError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), resp.JSONDefault.Message))
		}
		return scanResults, newGetExistingError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

// nolint:cyclop
func patchExistingScanResult(apiClient client.ClientWithResponsesInterface, scanResults models.TargetScanResult) error {
	newUpdateScanResultError := func(err error) error {
		return fmt.Errorf("failed to update scan result %v on server %v: %w", scanResultID, server, err)
	}

	resp, err := apiClient.PatchScanResultsScanResultIDWithResponse(context.TODO(), scanResultID, scanResults)
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
			return newUpdateScanResultError(fmt.Errorf("not found: %v", resp.JSON404.Message))
		}
		return newUpdateScanResultError(fmt.Errorf("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return newUpdateScanResultError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), resp.JSONDefault.Message))
		}
		return newUpdateScanResultError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func MarkScanResultInProgress() error {
	apiClient, err := client.NewClientWithResponses(server)
	if err != nil {
		return fmt.Errorf("unable to create VMClarity API client: %w", err)
	}

	scanResults, err := getExistingScanResult(apiClient)
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

	err = patchExistingScanResult(apiClient, scanResults)
	if err != nil {
		return err
	}

	return nil
}

func MarkScanResultDone(errors []error) error {
	apiClient, err := client.NewClientWithResponses(server)
	if err != nil {
		return fmt.Errorf("unable to create VMClarity API client: %w", err)
	}

	scanResults, err := getExistingScanResult(apiClient)
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
			errorStrs = append(errorStrs, err.Error())
		}
		if len(errorStrs) > 0 {
			scanResults.Status.General.Errors = &errorStrs
		}
	}

	err = patchExistingScanResult(apiClient, scanResults)
	if err != nil {
		return err
	}

	return nil
}

func ExportSbomResult(res *results.Results) error {
	apiClient, err := client.NewClientWithResponses(server)
	if err != nil {
		return fmt.Errorf("unable to create VMClarity API client: %w", err)
	}

	scanResults, err := getExistingScanResult(apiClient)
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

	sbomResults, err := results.GetResult[*sbom.Results](res)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to get sbom from scan: %w", err).Error())
	} else {
		scanResults.Sboms = convertSBOMResultToAPIModel(sbomResults)
	}

	state := models.DONE
	scanResults.Status.Sbom.State = &state
	scanResults.Status.Sbom.Errors = &errors

	err = patchExistingScanResult(apiClient, scanResults)
	if err != nil {
		return err
	}

	return nil
}

func ExportVulResult(res *results.Results) error {
	apiClient, err := client.NewClientWithResponses(server)
	if err != nil {
		return fmt.Errorf("unable to create VMClarity API client: %w", err)
	}

	scanResults, err := getExistingScanResult(apiClient)
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

	vulnerabilitiesResults, err := results.GetResult[*vulnerabilities.Results](res)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to get vulnerabilities from scan: %w", err).Error())
	} else {
		scanResults.Vulnerabilities = convertVulnResultToAPIModel(vulnerabilitiesResults)
	}

	state := models.DONE
	scanResults.Status.Vulnerabilities.State = &state
	scanResults.Status.Vulnerabilities.Errors = &errors

	err = patchExistingScanResult(apiClient, scanResults)
	if err != nil {
		return err
	}

	return nil
}

func ExportResults(res *results.Results) []error {
	var errors []error
	if config.SBOM.Enabled {
		err := ExportSbomResult(res)
		if err != nil {
			err = fmt.Errorf("failed to export sbom to server: %w", err)
			logger.Error(err)
			errors = append(errors, err)
		}
	}

	if config.Vulnerabilities.Enabled {
		err := ExportVulResult(res)
		if err != nil {
			err = fmt.Errorf("failed to export vulnerabilties to server: %w", err)
			logger.Error(err)
			errors = append(errors, err)
		}
	}
	return errors
}
