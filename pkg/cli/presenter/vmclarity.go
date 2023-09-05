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
	cliutils "github.com/openclarity/vmclarity/pkg/cli/utils"
	"github.com/openclarity/vmclarity/pkg/shared/backendclient"
	"github.com/openclarity/vmclarity/pkg/shared/families"
	"github.com/openclarity/vmclarity/pkg/shared/families/exploits"
	"github.com/openclarity/vmclarity/pkg/shared/families/infofinder"
	"github.com/openclarity/vmclarity/pkg/shared/families/malware"
	"github.com/openclarity/vmclarity/pkg/shared/families/misconfiguration"
	"github.com/openclarity/vmclarity/pkg/shared/families/rootkits"
	"github.com/openclarity/vmclarity/pkg/shared/families/sbom"
	"github.com/openclarity/vmclarity/pkg/shared/families/secrets"
	"github.com/openclarity/vmclarity/pkg/shared/families/types"
	"github.com/openclarity/vmclarity/pkg/shared/families/vulnerabilities"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

type AssetScanID = models.AssetScanID

type VMClarityPresenter struct {
	client *backendclient.BackendClient

	assetScanID models.AssetScanID
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
	case types.InfoFinder:
		err = v.ExportInfoFinderResult(ctx, res)
	}

	return err
}

func (v *VMClarityPresenter) ExportSbomResult(ctx context.Context, res families.FamilyResult) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, models.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Status == nil {
		assetScan.Status = &models.AssetScanStatus{}
	}
	if assetScan.Status.Sbom == nil {
		assetScan.Status.Sbom = &models.AssetScanState{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &models.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &models.AssetScanStats{}
	}

	errs := []string{}

	if res.Err != nil {
		errs = append(errs, res.Err.Error())
	} else {
		sbomResults, ok := res.Result.(*sbom.Results)
		if !ok {
			errs = append(errs, fmt.Errorf("failed to convert to sbom results").Error())
		} else {
			assetScan.Sboms = cliutils.ConvertSBOMResultToAPIModel(sbomResults)
			if assetScan.Sboms.Packages != nil {
				assetScan.Summary.TotalPackages = utils.PointerTo(len(*assetScan.Sboms.Packages))
			}
			assetScan.Stats.Sbom = getInputScanStats(sbomResults.Metadata.InputScans)
		}
	}

	state := models.AssetScanStateStateDone
	assetScan.Status.Sbom.State = &state
	assetScan.Status.Sbom.Errors = &errs

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityPresenter) ExportVulResult(ctx context.Context, res families.FamilyResult) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, models.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Status == nil {
		assetScan.Status = &models.AssetScanStatus{}
	}
	if assetScan.Status.Vulnerabilities == nil {
		assetScan.Status.Vulnerabilities = &models.AssetScanState{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &models.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &models.AssetScanStats{}
	}

	errs := []string{}

	if res.Err != nil {
		errs = append(errs, res.Err.Error())
	} else {
		vulnerabilitiesResults, ok := res.Result.(*vulnerabilities.Results)
		if !ok {
			errs = append(errs, fmt.Errorf("failed to convert to vulnerabilities results").Error())
		} else {
			assetScan.Vulnerabilities = cliutils.ConvertVulnResultToAPIModel(vulnerabilitiesResults)
		}
		assetScan.Summary.TotalVulnerabilities = utils.GetVulnerabilityTotalsPerSeverity(assetScan.Vulnerabilities.Vulnerabilities)
		assetScan.Stats.Vulnerabilities = getInputScanStats(vulnerabilitiesResults.Metadata.InputScans)
	}

	state := models.AssetScanStateStateDone
	assetScan.Status.Vulnerabilities.State = &state
	assetScan.Status.Vulnerabilities.Errors = &errs

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityPresenter) ExportSecretsResult(ctx context.Context, res families.FamilyResult) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, models.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Status == nil {
		assetScan.Status = &models.AssetScanStatus{}
	}
	if assetScan.Status.Secrets == nil {
		assetScan.Status.Secrets = &models.AssetScanState{}
	}

	if assetScan.Summary == nil {
		assetScan.Summary = &models.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &models.AssetScanStats{}
	}

	errs := []string{}

	if res.Err != nil {
		errs = append(errs, res.Err.Error())
	} else {
		secretsResults, ok := res.Result.(*secrets.Results)
		if !ok {
			errs = append(errs, fmt.Errorf("failed to convert to secrets results").Error())
		} else {
			assetScan.Secrets = cliutils.ConvertSecretsResultToAPIModel(secretsResults)
			if assetScan.Secrets.Secrets != nil {
				assetScan.Summary.TotalSecrets = utils.PointerTo(len(*assetScan.Secrets.Secrets))
			}
			assetScan.Stats.Secrets = getInputScanStats(secretsResults.Metadata.InputScans)
		}
	}

	state := models.AssetScanStateStateDone
	assetScan.Status.Secrets.State = &state
	assetScan.Status.Secrets.Errors = &errs

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func getInputScanStats(inputScans []types.InputScanMetadata) *[]models.AssetScanInputScanStats {
	if len(inputScans) == 0 {
		return nil
	}

	ret := make([]models.AssetScanInputScanStats, 0, len(inputScans))
	for i := range inputScans {
		scan := inputScans[i]
		ret = append(ret, models.AssetScanInputScanStats{
			Path: &scan.InputPath,
			ScanTime: &models.AssetScanScanTime{
				EndTime:   &scan.ScanEndTime,
				StartTime: &scan.ScanStartTime,
			},
			Size: &scan.InputSize,
			Type: &scan.InputType,
		})
	}

	return &ret
}

func (v *VMClarityPresenter) ExportMalwareResult(ctx context.Context, res families.FamilyResult) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, models.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Status == nil {
		assetScan.Status = &models.AssetScanStatus{}
	}
	if assetScan.Status.Malware == nil {
		assetScan.Status.Malware = &models.AssetScanState{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &models.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &models.AssetScanStats{}
	}

	errs := []string{}

	if res.Err != nil {
		errs = append(errs, res.Err.Error())
	} else {
		malwareResults, ok := res.Result.(*malware.MergedResults)
		if !ok {
			errs = append(errs, fmt.Errorf("failed to convert to malware results").Error())
		} else {
			assetScan.Malware = cliutils.ConvertMalwareResultToAPIModel(malwareResults)
			if assetScan.Malware.Malware != nil {
				assetScan.Summary.TotalMalware = utils.PointerTo[int](len(*assetScan.Malware.Malware))
			}
			assetScan.Stats.Malware = getInputScanStats(malwareResults.Metadata.InputScans)
		}
	}

	state := models.AssetScanStateStateDone
	assetScan.Status.Malware.State = &state
	assetScan.Status.Malware.Errors = &errs

	if err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID); err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityPresenter) ExportExploitsResult(ctx context.Context, res families.FamilyResult) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, models.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Status == nil {
		assetScan.Status = &models.AssetScanStatus{}
	}
	if assetScan.Status.Exploits == nil {
		assetScan.Status.Exploits = &models.AssetScanState{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &models.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &models.AssetScanStats{}
	}

	errs := []string{}

	if res.Err != nil {
		errs = append(errs, res.Err.Error())
	} else {
		exploitsResults, ok := res.Result.(*exploits.Results)
		if !ok {
			errs = append(errs, fmt.Errorf("failed to convert to exploits results").Error())
		} else {
			assetScan.Exploits = cliutils.ConvertExploitsResultToAPIModel(exploitsResults)
			if assetScan.Exploits.Exploits != nil {
				assetScan.Summary.TotalExploits = utils.PointerTo(len(*assetScan.Exploits.Exploits))
			}
			assetScan.Stats.Exploits = getInputScanStats(exploitsResults.Metadata.InputScans)
		}
	}

	state := models.AssetScanStateStateDone
	assetScan.Status.Exploits.State = &state
	assetScan.Status.Exploits.Errors = &errs

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityPresenter) ExportMisconfigurationResult(ctx context.Context, res families.FamilyResult) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, models.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Status == nil {
		assetScan.Status = &models.AssetScanStatus{}
	}
	if assetScan.Status.Misconfigurations == nil {
		assetScan.Status.Misconfigurations = &models.AssetScanState{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &models.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &models.AssetScanStats{}
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
				assetScan.Misconfigurations = apiMisconfigurations
				assetScan.Summary.TotalMisconfigurations = utils.PointerTo(len(misconfigurationResults.Misconfigurations))
			}
			assetScan.Stats.Misconfigurations = getInputScanStats(misconfigurationResults.Metadata.InputScans)
		}
	}

	state := models.AssetScanStateStateDone
	assetScan.Status.Misconfigurations.State = &state
	assetScan.Status.Misconfigurations.Errors = &errs

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityPresenter) ExportInfoFinderResult(ctx context.Context, res families.FamilyResult) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, models.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Status == nil {
		assetScan.Status = &models.AssetScanStatus{}
	}
	if assetScan.Status.InfoFinder == nil {
		assetScan.Status.InfoFinder = &models.AssetScanState{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &models.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &models.AssetScanStats{}
	}

	var errs []string

	if res.Err != nil {
		errs = append(errs, res.Err.Error())
	} else {
		results, ok := res.Result.(*infofinder.Results)
		if !ok {
			errs = append(errs, fmt.Errorf("failed to convert to info finder results").Error())
		} else {
			apiInfoFinder, err := cliutils.ConvertInfoFinderResultToAPIModel(results)
			if err != nil {
				errs = append(errs, fmt.Sprintf("failed to convert info finder results from scan to API model: %v", err))
			} else {
				assetScan.InfoFinder = apiInfoFinder
				assetScan.Summary.TotalInfoFinder = utils.PointerTo(len(results.Infos))
			}
			assetScan.Stats.InfoFinder = getInputScanStats(results.Metadata.InputScans)
		}
	}

	state := models.AssetScanStateStateDone
	assetScan.Status.InfoFinder.State = &state
	assetScan.Status.InfoFinder.Errors = &errs

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityPresenter) ExportRootkitResult(ctx context.Context, res families.FamilyResult) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, models.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Status == nil {
		assetScan.Status = &models.AssetScanStatus{}
	}
	if assetScan.Status.Rootkits == nil {
		assetScan.Status.Rootkits = &models.AssetScanState{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &models.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &models.AssetScanStats{}
	}

	var errs []string

	if res.Err != nil {
		errs = append(errs, res.Err.Error())
	} else {
		rootkitsResults, ok := res.Result.(*rootkits.Results)
		if !ok {
			errs = append(errs, fmt.Errorf("failed to convert to rootkits results").Error())
		} else {
			assetScan.Rootkits = cliutils.ConvertRootkitsResultToAPIModel(rootkitsResults)
			if assetScan.Rootkits.Rootkits != nil {
				assetScan.Summary.TotalRootkits = utils.PointerTo[int](len(*assetScan.Rootkits.Rootkits))
			}
			assetScan.Stats.Rootkits = getInputScanStats(rootkitsResults.Metadata.InputScans)
		}
	}

	state := models.AssetScanStateStateDone
	assetScan.Status.Rootkits.State = &state
	assetScan.Status.Rootkits.Errors = &errs

	if err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID); err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func NewVMClarityPresenter(client *backendclient.BackendClient, id AssetScanID) (*VMClarityPresenter, error) {
	if client == nil {
		return nil, errors.New("backend client must not be nil")
	}
	return &VMClarityPresenter{
		client:      client,
		assetScanID: id,
	}, nil
}
