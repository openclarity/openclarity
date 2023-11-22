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

	if assetScan.Sbom == nil {
		assetScan.Sbom = &models.SbomScan{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &models.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &models.AssetScanStats{}
	}

	if res.Err != nil {
		assetScan.Sbom.Status = models.NewScannerStatus(
			models.ScannerStatusStateFailed,
			models.ScannerStatusReasonError,
			utils.PointerTo(res.Err.Error()),
		)
	} else {
		sbomResults, ok := res.Result.(*sbom.Results)
		if !ok {
			assetScan.Sbom.Status = models.NewScannerStatus(
				models.ScannerStatusStateFailed,
				models.ScannerStatusReasonError,
				utils.PointerTo("failed to convert to sbom results"),
			)
		} else {
			assetScan.Sbom.Packages = utils.PointerTo(cliutils.ConvertSBOMResultToPackages(sbomResults))
			assetScan.Summary.TotalPackages = utils.PointerTo(len(*assetScan.Sbom.Packages))
			assetScan.Stats.Sbom = getInputScanStats(sbomResults.Metadata.InputScans)
			assetScan.Sbom.Status = models.NewScannerStatus(
				models.ScannerStatusStateDone,
				models.ScannerStatusReasonSuccess,
				nil,
			)
		}
	}

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

	if assetScan.Vulnerabilities == nil {
		assetScan.Vulnerabilities = &models.VulnerabilityScan{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &models.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &models.AssetScanStats{}
	}

	if res.Err != nil {
		assetScan.Vulnerabilities.Status = models.NewScannerStatus(
			models.ScannerStatusStateFailed,
			models.ScannerStatusReasonError,
			utils.PointerTo(res.Err.Error()),
		)
	} else {
		vulnerabilitiesResults, ok := res.Result.(*vulnerabilities.Results)
		if !ok {
			assetScan.Vulnerabilities.Status = models.NewScannerStatus(
				models.ScannerStatusStateFailed,
				models.ScannerStatusReasonError,
				utils.PointerTo("failed to convert to vulnerabilities results"),
			)
		} else {
			assetScan.Vulnerabilities.Vulnerabilities = utils.PointerTo(cliutils.ConvertVulnResultToVulnerabilities(vulnerabilitiesResults))
			assetScan.Summary.TotalVulnerabilities = utils.GetVulnerabilityTotalsPerSeverity(assetScan.Vulnerabilities.Vulnerabilities)
			assetScan.Stats.Vulnerabilities = getInputScanStats(vulnerabilitiesResults.Metadata.InputScans)
			assetScan.Vulnerabilities.Status = models.NewScannerStatus(
				models.ScannerStatusStateDone,
				models.ScannerStatusReasonSuccess,
				nil,
			)
		}
	}

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

	if assetScan.Secrets == nil {
		assetScan.Secrets = &models.SecretScan{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &models.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &models.AssetScanStats{}
	}

	if res.Err != nil {
		assetScan.Secrets.Status = models.NewScannerStatus(
			models.ScannerStatusStateFailed,
			models.ScannerStatusReasonError,
			utils.PointerTo(res.Err.Error()),
		)
	} else {
		secretsResults, ok := res.Result.(*secrets.Results)
		if !ok {
			assetScan.Secrets.Status = models.NewScannerStatus(
				models.ScannerStatusStateFailed,
				models.ScannerStatusReasonError,
				utils.PointerTo("failed to convert to secrets results"),
			)
		} else {
			assetScan.Secrets.Secrets = utils.PointerTo(cliutils.ConvertSecretsResultToSecrets(secretsResults))
			assetScan.Summary.TotalSecrets = utils.PointerTo(len(*assetScan.Secrets.Secrets))
			assetScan.Stats.Secrets = getInputScanStats(secretsResults.Metadata.InputScans)
			assetScan.Secrets.Status = models.NewScannerStatus(
				models.ScannerStatusStateDone,
				models.ScannerStatusReasonSuccess,
				nil,
			)
		}
	}

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

	if assetScan.Malware == nil {
		assetScan.Malware = &models.MalwareScan{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &models.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &models.AssetScanStats{}
	}

	if res.Err != nil {
		assetScan.Malware.Status = models.NewScannerStatus(
			models.ScannerStatusStateFailed,
			models.ScannerStatusReasonError,
			utils.PointerTo(res.Err.Error()),
		)
	} else {
		malwareResults, ok := res.Result.(*malware.MergedResults)
		if !ok {
			assetScan.Sbom.Status = models.NewScannerStatus(
				models.ScannerStatusStateFailed,
				models.ScannerStatusReasonError,
				utils.PointerTo("failed to convert to malware results"),
			)
		} else {
			mware, mdata := cliutils.ConvertMalwareResultToMalwareAndMetadata(malwareResults)
			assetScan.Summary.TotalMalware = utils.PointerTo(len(mware))
			assetScan.Stats.Malware = getInputScanStats(malwareResults.Metadata.InputScans)
			assetScan.Malware.Malware = utils.PointerTo(mware)
			assetScan.Malware.Metadata = utils.PointerTo(mdata)
			assetScan.Malware.Status = models.NewScannerStatus(
				models.ScannerStatusStateDone,
				models.ScannerStatusReasonSuccess,
				nil,
			)
		}
	}

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

	if assetScan.Exploits == nil {
		assetScan.Exploits = &models.ExploitScan{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &models.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &models.AssetScanStats{}
	}

	if res.Err != nil {
		assetScan.Exploits.Status = models.NewScannerStatus(
			models.ScannerStatusStateFailed,
			models.ScannerStatusReasonError,
			utils.PointerTo(res.Err.Error()),
		)
	} else {
		exploitsResults, ok := res.Result.(*exploits.Results)
		if !ok {
			assetScan.Exploits.Status = models.NewScannerStatus(
				models.ScannerStatusStateFailed,
				models.ScannerStatusReasonError,
				utils.PointerTo("failed to convert to exploits results"),
			)
		} else {
			assetScan.Exploits.Exploits = utils.PointerTo(cliutils.ConvertExploitsResultToExploits(exploitsResults))
			assetScan.Summary.TotalExploits = utils.PointerTo(len(*assetScan.Exploits.Exploits))
			assetScan.Stats.Exploits = getInputScanStats(exploitsResults.Metadata.InputScans)
			assetScan.Exploits.Status = models.NewScannerStatus(
				models.ScannerStatusStateDone,
				models.ScannerStatusReasonSuccess,
				nil,
			)
		}
	}

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

	if assetScan.Misconfigurations == nil {
		assetScan.Misconfigurations = &models.MisconfigurationScan{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &models.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &models.AssetScanStats{}
	}

	if res.Err != nil {
		assetScan.Misconfigurations.Status = models.NewScannerStatus(
			models.ScannerStatusStateFailed,
			models.ScannerStatusReasonError,
			utils.PointerTo(res.Err.Error()),
		)
	} else {
		misconfigurationResults, ok := res.Result.(*misconfiguration.Results)
		if !ok {
			assetScan.Misconfigurations.Status = models.NewScannerStatus(
				models.ScannerStatusStateFailed,
				models.ScannerStatusReasonError,
				utils.PointerTo("failed to convert to misconfiguration results"),
			)
		} else {
			misconfigurations, scanners, err := cliutils.ConvertMisconfigurationResultToMisconfigurationsAndScanners(misconfigurationResults)
			if err != nil {
				assetScan.Misconfigurations.Status = models.NewScannerStatus(
					models.ScannerStatusStateFailed,
					models.ScannerStatusReasonError,
					utils.PointerTo(fmt.Errorf("failed to convert misconfiguration results from scan to API model: %w", err).Error()),
				)
			} else {
				assetScan.Misconfigurations.Status = models.NewScannerStatus(
					models.ScannerStatusStateDone,
					models.ScannerStatusReasonSuccess,
					nil,
				)
				assetScan.Misconfigurations.Misconfigurations = utils.PointerTo(misconfigurations)
				assetScan.Misconfigurations.Scanners = utils.PointerTo(scanners)
			}
			assetScan.Summary.TotalMisconfigurations = utils.PointerTo(len(misconfigurationResults.Misconfigurations))
			assetScan.Stats.Misconfigurations = getInputScanStats(misconfigurationResults.Metadata.InputScans)
		}
	}

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

	if assetScan.InfoFinder == nil {
		assetScan.InfoFinder = &models.InfoFinderScan{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &models.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &models.AssetScanStats{}
	}

	if res.Err != nil {
		assetScan.InfoFinder.Status = models.NewScannerStatus(
			models.ScannerStatusStateFailed,
			models.ScannerStatusReasonError,
			utils.PointerTo(res.Err.Error()),
		)
	} else {
		results, ok := res.Result.(*infofinder.Results)
		if !ok {
			assetScan.InfoFinder.Status = models.NewScannerStatus(
				models.ScannerStatusStateFailed,
				models.ScannerStatusReasonError,
				utils.PointerTo("failed to convert to info finder results"),
			)
		} else {
			apiInfoFinder, scanners, err := cliutils.ConvertInfoFinderResultToInfosAndScanners(results)
			if err != nil {
				assetScan.InfoFinder.Status = models.NewScannerStatus(
					models.ScannerStatusStateFailed,
					models.ScannerStatusReasonError,
					utils.PointerTo(fmt.Errorf("failed to convert info finder results from scan to API model: %w", err).Error()),
				)
			} else {
				assetScan.InfoFinder.Status = models.NewScannerStatus(
					models.ScannerStatusStateDone,
					models.ScannerStatusReasonSuccess,
					nil,
				)
				assetScan.InfoFinder.Infos = utils.PointerTo(apiInfoFinder)
				assetScan.InfoFinder.Scanners = utils.PointerTo(scanners)
			}
			assetScan.Summary.TotalInfoFinder = utils.PointerTo(len(results.Infos))
			assetScan.Stats.InfoFinder = getInputScanStats(results.Metadata.InputScans)
		}
	}

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

	if assetScan.Rootkits == nil {
		assetScan.Rootkits = &models.RootkitScan{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &models.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &models.AssetScanStats{}
	}

	if res.Err != nil {
		assetScan.Rootkits.Status = models.NewScannerStatus(
			models.ScannerStatusStateFailed,
			models.ScannerStatusReasonError,
			utils.PointerTo(res.Err.Error()),
		)
	} else {
		rootkitsResults, ok := res.Result.(*rootkits.Results)
		if !ok {
			assetScan.Rootkits.Status = models.NewScannerStatus(
				models.ScannerStatusStateFailed,
				models.ScannerStatusReasonError,
				utils.PointerTo("failed to convert to rootkits results"),
			)
		} else {
			assetScan.Rootkits.Rootkits = utils.PointerTo(cliutils.ConvertRootkitsResultToRootkits(rootkitsResults))
			assetScan.Summary.TotalRootkits = utils.PointerTo(len(*assetScan.Rootkits.Rootkits))
			assetScan.Stats.Rootkits = getInputScanStats(rootkitsResults.Metadata.InputScans)
			assetScan.Rootkits.Status = models.NewScannerStatus(
				models.ScannerStatusStateDone,
				models.ScannerStatusReasonSuccess,
				nil,
			)
		}
	}

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
