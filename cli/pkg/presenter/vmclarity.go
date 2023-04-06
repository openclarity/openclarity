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
	"context"
	"errors"
	"fmt"

	"github.com/openclarity/vmclarity/api/models"
	cliutils "github.com/openclarity/vmclarity/cli/pkg/utils"
	"github.com/openclarity/vmclarity/shared/pkg/backendclient"
	"github.com/openclarity/vmclarity/shared/pkg/families"
	"github.com/openclarity/vmclarity/shared/pkg/families/exploits"
	"github.com/openclarity/vmclarity/shared/pkg/families/malware"
	"github.com/openclarity/vmclarity/shared/pkg/families/misconfiguration"
	"github.com/openclarity/vmclarity/shared/pkg/families/results"
	"github.com/openclarity/vmclarity/shared/pkg/families/sbom"
	"github.com/openclarity/vmclarity/shared/pkg/families/secrets"
	"github.com/openclarity/vmclarity/shared/pkg/families/types"
	"github.com/openclarity/vmclarity/shared/pkg/families/vulnerabilities"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

type ScanResultID = models.ScanResultID

type VMClarityPresenter struct {
	client *backendclient.BackendClient

	scanResultID models.ScanResultID
}

func (v *VMClarityPresenter) ExportSbomResult(ctx context.Context, res *results.Results, famerr families.RunErrors) error {
	scanResult, err := v.client.GetScanResult(ctx, v.scanResultID, models.GetScanResultsScanResultIDParams{})
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
		scanResult.Summary = &models.ScanFindingsSummary{}
	}

	errs := []string{}

	if err, ok := famerr[types.SBOM]; ok {
		errs = append(errs, err.Error())
	} else {
		sbomResults, err := results.GetResult[*sbom.Results](res)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get sbom from scan: %w", err).Error())
		} else {
			scanResult.Sboms = cliutils.ConvertSBOMResultToAPIModel(sbomResults)
			if scanResult.Sboms.Packages != nil {
				scanResult.Summary.TotalPackages = utils.PointerTo(len(*scanResult.Sboms.Packages))
			}
		}
	}

	state := models.DONE
	scanResult.Status.Sbom.State = &state
	scanResult.Status.Sbom.Errors = &errs

	err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (v *VMClarityPresenter) ExportVulResult(ctx context.Context, res *results.Results, famerr families.RunErrors) error {
	scanResult, err := v.client.GetScanResult(ctx, v.scanResultID, models.GetScanResultsScanResultIDParams{})
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
		scanResult.Summary = &models.ScanFindingsSummary{}
	}

	errs := []string{}

	if err, ok := famerr[types.Vulnerabilities]; ok {
		errs = append(errs, err.Error())
	} else {
		vulnerabilitiesResults, err := results.GetResult[*vulnerabilities.Results](res)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get vulnerabilities from scan: %w", err).Error())
		} else {
			scanResult.Vulnerabilities = cliutils.ConvertVulnResultToAPIModel(vulnerabilitiesResults)
		}
		scanResult.Summary.TotalVulnerabilities = utils.GetVulnerabilityTotalsPerSeverity(scanResult.Vulnerabilities.Vulnerabilities)
	}

	state := models.DONE
	scanResult.Status.Vulnerabilities.State = &state
	scanResult.Status.Vulnerabilities.Errors = &errs

	err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (v *VMClarityPresenter) ExportSecretsResult(ctx context.Context, res *results.Results, famerr families.RunErrors) error {
	scanResult, err := v.client.GetScanResult(ctx, v.scanResultID, models.GetScanResultsScanResultIDParams{})
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
		scanResult.Summary = &models.ScanFindingsSummary{}
	}

	errs := []string{}

	if err, ok := famerr[types.Secrets]; ok {
		errs = append(errs, err.Error())
	} else {
		secretsResults, err := results.GetResult[*secrets.Results](res)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get secrets results from scan: %w", err).Error())
		} else {
			scanResult.Secrets = cliutils.ConvertSecretsResultToAPIModel(secretsResults)
			if scanResult.Secrets.Secrets != nil {
				scanResult.Summary.TotalSecrets = utils.PointerTo(len(*scanResult.Secrets.Secrets))
			}
		}
	}

	state := models.DONE
	scanResult.Status.Secrets.State = &state
	scanResult.Status.Secrets.Errors = &errs

	err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (v *VMClarityPresenter) ExportMalwareResult(ctx context.Context, res *results.Results, famerr families.RunErrors) error {
	scanResult, err := v.client.GetScanResult(ctx, v.scanResultID, models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get scan result: %w", err)
	}

	if scanResult.Status == nil {
		scanResult.Status = &models.TargetScanStatus{}
	}
	if scanResult.Status.Malware == nil {
		scanResult.Status.Malware = &models.TargetScanState{}
	}

	errs := []string{}

	if err, ok := famerr[types.Malware]; ok {
		errs = append(errs, err.Error())
	} else {
		malwareResults, err := results.GetResult[*malware.MergedResults](res)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get malware results from scan: %w", err).Error())
		} else {
			scanResult.Malware = cliutils.ConvertMalwareResultToAPIModel(malwareResults)
			if scanResult.Malware.Malware != nil {
				scanResult.Summary.TotalMalware = utils.PointerTo[int](len(*scanResult.Malware.Malware))
			}
		}
	}

	state := models.DONE
	scanResult.Status.Malware.State = &state
	scanResult.Status.Malware.Errors = &errs

	if err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID); err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (v *VMClarityPresenter) ExportExploitsResult(ctx context.Context, res *results.Results, famerr families.RunErrors) error {
	scanResult, err := v.client.GetScanResult(ctx, v.scanResultID, models.GetScanResultsScanResultIDParams{})
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
		scanResult.Summary = &models.ScanFindingsSummary{}
	}

	errs := []string{}

	if err, ok := famerr[types.Exploits]; ok {
		errs = append(errs, err.Error())
	} else {
		exploitsResults, err := results.GetResult[*exploits.Results](res)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get exploits results from scan: %w", err).Error())
		} else {
			scanResult.Exploits = cliutils.ConvertExploitsResultToAPIModel(exploitsResults)
			if scanResult.Exploits.Exploits != nil {
				scanResult.Summary.TotalExploits = utils.PointerTo(len(*scanResult.Exploits.Exploits))
			}
		}
	}

	state := models.DONE
	scanResult.Status.Exploits.State = &state
	scanResult.Status.Exploits.Errors = &errs

	err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (v *VMClarityPresenter) ExportMisconfigurationResult(ctx context.Context, res *results.Results, famerr families.RunErrors) error {
	scanResult, err := v.client.GetScanResult(ctx, v.scanResultID, models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get scan result: %w", err)
	}

	if scanResult.Status == nil {
		scanResult.Status = &models.TargetScanStatus{}
	}
	if scanResult.Status.Exploits == nil {
		scanResult.Status.Misconfigurations = &models.TargetScanState{}
	}
	if scanResult.Summary == nil {
		scanResult.Summary = &models.ScanFindingsSummary{}
	}

	var errs []string

	if err, ok := famerr[types.Misconfiguration]; ok {
		errs = append(errs, err.Error())
	} else {
		misconfigurationResults, err := results.GetResult[*misconfiguration.Results](res)
		if err != nil {
			errs = append(errs, fmt.Sprintf("failed to get misconfiguration results from scan: %v", err))
		} else {
			apiMisconfigurations, err := cliutils.ConvertMisconfigurationResultToAPIModel(misconfigurationResults)
			if err != nil {
				errs = append(errs, fmt.Sprintf("failed to convert misconfiguration results from scan to API model: %v", err))
			} else {
				scanResult.Misconfigurations = apiMisconfigurations
				scanResult.Summary.TotalMisconfigurations = utils.PointerTo(len(misconfigurationResults.Misconfigurations))
			}
		}
	}

	state := models.DONE
	scanResult.Status.Misconfigurations.State = &state
	scanResult.Status.Misconfigurations.Errors = &errs

	err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func NewVMClarityPresenter(client *backendclient.BackendClient, id ScanResultID) (*VMClarityPresenter, error) {
	if client == nil {
		return nil, errors.New("backend client must not be nil")
	}
	return &VMClarityPresenter{
		client:       client,
		scanResultID: id,
	}, nil
}
