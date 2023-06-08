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
	"github.com/openclarity/vmclarity/shared/pkg/families/rootkits"
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

func (v *VMClarityPresenter) ExportFamilyResult(ctx context.Context, res families.FamilyResult) error {
	var err error

	switch res.FamilyType {
	case types.SBOM:
		err = v.ExportSbomResult(ctx, res)
	case types.Vulnerabilities:
		err = v.ExportVulResult(ctx, res)
	case types.Secrets:
		err = v.ExportSecretsResult(ctx, res)
	case types.Exploits:
		err = v.ExportExploitsResult(ctx, res)
	case types.Misconfiguration:
		err = v.ExportMisconfigurationResult(ctx, res)
	case types.Rootkits:
		err = v.ExportRootkitResult(ctx, res)
	case types.Malware:
		err = v.ExportMalwareResult(ctx, res)
	}

	return err
}

func (v *VMClarityPresenter) ExportSbomResult(ctx context.Context, res families.FamilyResult) error {
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

	if res.Err != nil {
		errs = append(errs, res.Err.Error())
	} else {
		sbomResults, ok := res.Result.(*sbom.Results)
		if !ok {
			errs = append(errs, fmt.Errorf("failed to convert to sbom results").Error())
		} else {
			scanResult.Sboms = cliutils.ConvertSBOMResultToAPIModel(sbomResults)
			if scanResult.Sboms.Packages != nil {
				scanResult.Summary.TotalPackages = utils.PointerTo(len(*scanResult.Sboms.Packages))
			}
		}
	}

	state := models.TargetScanStateStateDone
	scanResult.Status.Sbom.State = &state
	scanResult.Status.Sbom.Errors = &errs

	err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (v *VMClarityPresenter) ExportVulResult(ctx context.Context, res families.FamilyResult) error {
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

	if res.Err != nil {
		errs = append(errs, res.Err.Error())
	} else {
		vulnerabilitiesResults, ok := res.Result.(*vulnerabilities.Results)
		if !ok {
			errs = append(errs, fmt.Errorf("failed to convert to vulnerabilities results").Error())
		} else {
			scanResult.Vulnerabilities = cliutils.ConvertVulnResultToAPIModel(vulnerabilitiesResults)
		}
		scanResult.Summary.TotalVulnerabilities = utils.GetVulnerabilityTotalsPerSeverity(scanResult.Vulnerabilities.Vulnerabilities)
	}

	state := models.TargetScanStateStateDone
	scanResult.Status.Vulnerabilities.State = &state
	scanResult.Status.Vulnerabilities.Errors = &errs

	err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (v *VMClarityPresenter) ExportSecretsResult(ctx context.Context, res families.FamilyResult) error {
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

	if res.Err != nil {
		errs = append(errs, res.Err.Error())
	} else {
		secretsResults, ok := res.Result.(*secrets.Results)
		if !ok {
			errs = append(errs, fmt.Errorf("failed to convert to secrets results").Error())
		} else {
			scanResult.Secrets = cliutils.ConvertSecretsResultToAPIModel(secretsResults)
			if scanResult.Secrets.Secrets != nil {
				scanResult.Summary.TotalSecrets = utils.PointerTo(len(*scanResult.Secrets.Secrets))
			}
		}
	}

	state := models.TargetScanStateStateDone
	scanResult.Status.Secrets.State = &state
	scanResult.Status.Secrets.Errors = &errs

	err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (v *VMClarityPresenter) ExportMalwareResult(ctx context.Context, res families.FamilyResult) error {
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

	if res.Err != nil {
		errs = append(errs, res.Err.Error())
	} else {
		malwareResults, ok := res.Result.(*malware.MergedResults)
		if !ok {
			errs = append(errs, fmt.Errorf("failed to convert to malware results").Error())
		} else {
			scanResult.Malware = cliutils.ConvertMalwareResultToAPIModel(malwareResults)
			if scanResult.Malware.Malware != nil {
				scanResult.Summary.TotalMalware = utils.PointerTo[int](len(*scanResult.Malware.Malware))
			}
		}
	}

	state := models.TargetScanStateStateDone
	scanResult.Status.Malware.State = &state
	scanResult.Status.Malware.Errors = &errs

	if err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID); err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (v *VMClarityPresenter) ExportExploitsResult(ctx context.Context, res families.FamilyResult) error {
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

	if res.Err != nil {
		errs = append(errs, res.Err.Error())
	} else {
		exploitsResults, ok := res.Result.(*exploits.Results)
		if !ok {
			errs = append(errs, fmt.Errorf("failed to convert to exploits results").Error())
		} else {
			scanResult.Exploits = cliutils.ConvertExploitsResultToAPIModel(exploitsResults)
			if scanResult.Exploits.Exploits != nil {
				scanResult.Summary.TotalExploits = utils.PointerTo(len(*scanResult.Exploits.Exploits))
			}
		}
	}

	state := models.TargetScanStateStateDone
	scanResult.Status.Exploits.State = &state
	scanResult.Status.Exploits.Errors = &errs

	err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (v *VMClarityPresenter) ExportMisconfigurationResult(ctx context.Context, res families.FamilyResult) error {
	scanResult, err := v.client.GetScanResult(ctx, v.scanResultID, models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get scan result: %w", err)
	}

	if scanResult.Status == nil {
		scanResult.Status = &models.TargetScanStatus{}
	}
	if scanResult.Status.Misconfigurations == nil {
		scanResult.Status.Misconfigurations = &models.TargetScanState{}
	}
	if scanResult.Summary == nil {
		scanResult.Summary = &models.ScanFindingsSummary{}
	}

	var errs []string

	if res.Err != nil {
		errs = append(errs, res.Err.Error())
	} else {
		misconfigurationResults, ok := res.Result.(*misconfiguration.Results)
		if !ok {
			errs = append(errs, fmt.Errorf("failed to convert to misconfiguration results").Error())
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

	state := models.TargetScanStateStateDone
	scanResult.Status.Misconfigurations.State = &state
	scanResult.Status.Misconfigurations.Errors = &errs

	err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (v *VMClarityPresenter) ExportRootkitResult(ctx context.Context, res families.FamilyResult) error {
	scanResult, err := v.client.GetScanResult(ctx, v.scanResultID, models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get scan result: %w", err)
	}

	if scanResult.Status == nil {
		scanResult.Status = &models.TargetScanStatus{}
	}
	if scanResult.Status.Rootkits == nil {
		scanResult.Status.Rootkits = &models.TargetScanState{}
	}

	var errs []string

	if res.Err != nil {
		errs = append(errs, res.Err.Error())
	} else {
		rootkitsResults, ok := res.Result.(*rootkits.Results)
		if !ok {
			errs = append(errs, fmt.Errorf("failed to convert to rootkits results").Error())
		} else {
			scanResult.Rootkits = cliutils.ConvertRootkitsResultToAPIModel(rootkitsResults)
			if scanResult.Rootkits.Rootkits != nil {
				scanResult.Summary.TotalRootkits = utils.PointerTo[int](len(*scanResult.Rootkits.Rootkits))
			}
		}
	}

	state := models.TargetScanStateStateDone
	scanResult.Status.Rootkits.State = &state
	scanResult.Status.Rootkits.Errors = &errs

	if err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID); err != nil {
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
