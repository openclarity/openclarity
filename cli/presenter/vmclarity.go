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

	apiclient "github.com/openclarity/vmclarity/api/client"
	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/scanner/families"
	"github.com/openclarity/vmclarity/scanner/families/exploits"
	"github.com/openclarity/vmclarity/scanner/families/infofinder"
	"github.com/openclarity/vmclarity/scanner/families/malware"
	"github.com/openclarity/vmclarity/scanner/families/misconfiguration"
	"github.com/openclarity/vmclarity/scanner/families/plugins"
	"github.com/openclarity/vmclarity/scanner/families/rootkits"
	"github.com/openclarity/vmclarity/scanner/families/sbom"
	"github.com/openclarity/vmclarity/scanner/families/secrets"
	"github.com/openclarity/vmclarity/scanner/families/types"
	"github.com/openclarity/vmclarity/scanner/families/vulnerabilities"
	"github.com/openclarity/vmclarity/scanner/utils"
)

type AssetScanID = apitypes.AssetScanID

type VMClarityPresenter struct {
	client *apiclient.Client

	assetScanID apitypes.AssetScanID
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
	case types.Plugins:
		err = v.ExportPluginsResult(ctx, res)
	}

	return err
}

func (v *VMClarityPresenter) ExportSbomResult(ctx context.Context, res families.FamilyResult) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Sbom == nil {
		assetScan.Sbom = &apitypes.SbomScan{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &apitypes.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &apitypes.AssetScanStats{}
	}

	if res.Err != nil {
		assetScan.Sbom.Status = apitypes.NewScannerStatus(
			apitypes.ScannerStatusStateFailed,
			apitypes.ScannerStatusReasonError,
			to.Ptr(res.Err.Error()),
		)
	} else {
		sbomResults, ok := res.Result.(*sbom.Results)
		if !ok {
			assetScan.Sbom.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateFailed,
				apitypes.ScannerStatusReasonError,
				to.Ptr("failed to convert to sbom results"),
			)
		} else {
			assetScan.Sbom.Packages = to.Ptr(ConvertSBOMResultToPackages(sbomResults))
			assetScan.Summary.TotalPackages = to.Ptr(len(*assetScan.Sbom.Packages))
			assetScan.Stats.Sbom = getInputScanStats(sbomResults.Metadata.InputScans)
			assetScan.Sbom.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateDone,
				apitypes.ScannerStatusReasonSuccess,
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
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Vulnerabilities == nil {
		assetScan.Vulnerabilities = &apitypes.VulnerabilityScan{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &apitypes.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &apitypes.AssetScanStats{}
	}

	if res.Err != nil {
		assetScan.Vulnerabilities.Status = apitypes.NewScannerStatus(
			apitypes.ScannerStatusStateFailed,
			apitypes.ScannerStatusReasonError,
			to.Ptr(res.Err.Error()),
		)
	} else {
		vulnerabilitiesResults, ok := res.Result.(*vulnerabilities.Results)
		if !ok {
			assetScan.Vulnerabilities.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateFailed,
				apitypes.ScannerStatusReasonError,
				to.Ptr("failed to convert to vulnerabilities results"),
			)
		} else {
			assetScan.Vulnerabilities.Vulnerabilities = to.Ptr(ConvertVulnResultToVulnerabilities(vulnerabilitiesResults))
			assetScan.Summary.TotalVulnerabilities = utils.GetVulnerabilityTotalsPerSeverity(assetScan.Vulnerabilities.Vulnerabilities)
			assetScan.Stats.Vulnerabilities = getInputScanStats(vulnerabilitiesResults.Metadata.InputScans)
			assetScan.Vulnerabilities.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateDone,
				apitypes.ScannerStatusReasonSuccess,
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
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Secrets == nil {
		assetScan.Secrets = &apitypes.SecretScan{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &apitypes.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &apitypes.AssetScanStats{}
	}

	if res.Err != nil {
		assetScan.Secrets.Status = apitypes.NewScannerStatus(
			apitypes.ScannerStatusStateFailed,
			apitypes.ScannerStatusReasonError,
			to.Ptr(res.Err.Error()),
		)
	} else {
		secretsResults, ok := res.Result.(*secrets.Results)
		if !ok {
			assetScan.Secrets.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateFailed,
				apitypes.ScannerStatusReasonError,
				to.Ptr("failed to convert to secrets results"),
			)
		} else {
			assetScan.Secrets.Secrets = to.Ptr(ConvertSecretsResultToSecrets(secretsResults))
			assetScan.Summary.TotalSecrets = to.Ptr(len(*assetScan.Secrets.Secrets))
			assetScan.Stats.Secrets = getInputScanStats(secretsResults.Metadata.InputScans)
			assetScan.Secrets.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateDone,
				apitypes.ScannerStatusReasonSuccess,
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

func getInputScanStats(inputScans []types.InputScanMetadata) *[]apitypes.AssetScanInputScanStats {
	if len(inputScans) == 0 {
		return nil
	}

	ret := make([]apitypes.AssetScanInputScanStats, 0, len(inputScans))
	for i := range inputScans {
		scan := inputScans[i]
		ret = append(ret, apitypes.AssetScanInputScanStats{
			Path: &scan.InputPath,
			ScanTime: &apitypes.AssetScanScanTime{
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
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Malware == nil {
		assetScan.Malware = &apitypes.MalwareScan{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &apitypes.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &apitypes.AssetScanStats{}
	}

	if res.Err != nil {
		assetScan.Malware.Status = apitypes.NewScannerStatus(
			apitypes.ScannerStatusStateFailed,
			apitypes.ScannerStatusReasonError,
			to.Ptr(res.Err.Error()),
		)
	} else {
		malwareResults, ok := res.Result.(*malware.MergedResults)
		if !ok {
			assetScan.Sbom.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateFailed,
				apitypes.ScannerStatusReasonError,
				to.Ptr("failed to convert to malware results"),
			)
		} else {
			mware, mdata := ConvertMalwareResultToMalwareAndMetadata(malwareResults)
			assetScan.Summary.TotalMalware = to.Ptr(len(mware))
			assetScan.Stats.Malware = getInputScanStats(malwareResults.Metadata.InputScans)
			assetScan.Malware.Malware = to.Ptr(mware)
			assetScan.Malware.Metadata = to.Ptr(mdata)
			assetScan.Malware.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateDone,
				apitypes.ScannerStatusReasonSuccess,
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
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Exploits == nil {
		assetScan.Exploits = &apitypes.ExploitScan{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &apitypes.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &apitypes.AssetScanStats{}
	}

	if res.Err != nil {
		assetScan.Exploits.Status = apitypes.NewScannerStatus(
			apitypes.ScannerStatusStateFailed,
			apitypes.ScannerStatusReasonError,
			to.Ptr(res.Err.Error()),
		)
	} else {
		exploitsResults, ok := res.Result.(*exploits.Results)
		if !ok {
			assetScan.Exploits.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateFailed,
				apitypes.ScannerStatusReasonError,
				to.Ptr("failed to convert to exploits results"),
			)
		} else {
			assetScan.Exploits.Exploits = to.Ptr(ConvertExploitsResultToExploits(exploitsResults))
			assetScan.Summary.TotalExploits = to.Ptr(len(*assetScan.Exploits.Exploits))
			assetScan.Stats.Exploits = getInputScanStats(exploitsResults.Metadata.InputScans)
			assetScan.Exploits.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateDone,
				apitypes.ScannerStatusReasonSuccess,
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
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Misconfigurations == nil {
		assetScan.Misconfigurations = &apitypes.MisconfigurationScan{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &apitypes.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &apitypes.AssetScanStats{}
	}

	if res.Err != nil {
		assetScan.Misconfigurations.Status = apitypes.NewScannerStatus(
			apitypes.ScannerStatusStateFailed,
			apitypes.ScannerStatusReasonError,
			to.Ptr(res.Err.Error()),
		)
	} else {
		misconfigurationResults, ok := res.Result.(*misconfiguration.Results)
		if !ok {
			assetScan.Misconfigurations.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateFailed,
				apitypes.ScannerStatusReasonError,
				to.Ptr("failed to convert to misconfiguration results"),
			)
		} else {
			misconfigurations, scanners, err := ConvertMisconfigurationResultToMisconfigurationsAndScanners(misconfigurationResults)
			if err != nil {
				assetScan.Misconfigurations.Status = apitypes.NewScannerStatus(
					apitypes.ScannerStatusStateFailed,
					apitypes.ScannerStatusReasonError,
					to.Ptr(fmt.Errorf("failed to convert misconfiguration results from scan to API model: %w", err).Error()),
				)
			} else {
				assetScan.Misconfigurations.Status = apitypes.NewScannerStatus(
					apitypes.ScannerStatusStateDone,
					apitypes.ScannerStatusReasonSuccess,
					nil,
				)
				assetScan.Misconfigurations.Misconfigurations = to.Ptr(misconfigurations)
				assetScan.Misconfigurations.Scanners = to.Ptr(scanners)
			}
			assetScan.Summary.TotalMisconfigurations = to.Ptr(len(misconfigurationResults.Misconfigurations))
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
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.InfoFinder == nil {
		assetScan.InfoFinder = &apitypes.InfoFinderScan{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &apitypes.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &apitypes.AssetScanStats{}
	}

	if res.Err != nil {
		assetScan.InfoFinder.Status = apitypes.NewScannerStatus(
			apitypes.ScannerStatusStateFailed,
			apitypes.ScannerStatusReasonError,
			to.Ptr(res.Err.Error()),
		)
	} else {
		results, ok := res.Result.(*infofinder.Results)
		if !ok {
			assetScan.InfoFinder.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateFailed,
				apitypes.ScannerStatusReasonError,
				to.Ptr("failed to convert to info finder results"),
			)
		} else {
			apiInfoFinder, scanners, err := ConvertInfoFinderResultToInfosAndScanners(results)
			if err != nil {
				assetScan.InfoFinder.Status = apitypes.NewScannerStatus(
					apitypes.ScannerStatusStateFailed,
					apitypes.ScannerStatusReasonError,
					to.Ptr(fmt.Errorf("failed to convert info finder results from scan to API model: %w", err).Error()),
				)
			} else {
				assetScan.InfoFinder.Status = apitypes.NewScannerStatus(
					apitypes.ScannerStatusStateDone,
					apitypes.ScannerStatusReasonSuccess,
					nil,
				)
				assetScan.InfoFinder.Infos = to.Ptr(apiInfoFinder)
				assetScan.InfoFinder.Scanners = to.Ptr(scanners)
			}
			assetScan.Summary.TotalInfoFinder = to.Ptr(len(results.Infos))
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
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Rootkits == nil {
		assetScan.Rootkits = &apitypes.RootkitScan{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &apitypes.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &apitypes.AssetScanStats{}
	}

	if res.Err != nil {
		assetScan.Rootkits.Status = apitypes.NewScannerStatus(
			apitypes.ScannerStatusStateFailed,
			apitypes.ScannerStatusReasonError,
			to.Ptr(res.Err.Error()),
		)
	} else {
		rootkitsResults, ok := res.Result.(*rootkits.Results)
		if !ok {
			assetScan.Rootkits.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateFailed,
				apitypes.ScannerStatusReasonError,
				to.Ptr("failed to convert to rootkits results"),
			)
		} else {
			assetScan.Rootkits.Rootkits = to.Ptr(ConvertRootkitsResultToRootkits(rootkitsResults))
			assetScan.Summary.TotalRootkits = to.Ptr(len(*assetScan.Rootkits.Rootkits))
			assetScan.Stats.Rootkits = getInputScanStats(rootkitsResults.Metadata.InputScans)
			assetScan.Rootkits.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateDone,
				apitypes.ScannerStatusReasonSuccess,
				nil,
			)
		}
	}

	if err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID); err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityPresenter) ExportPluginsResult(ctx context.Context, res families.FamilyResult) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Plugins == nil {
		assetScan.Plugins = &apitypes.PluginScan{}
	}
	if assetScan.Summary == nil {
		assetScan.Summary = &apitypes.ScanFindingsSummary{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &apitypes.AssetScanStats{}
	}

	if res.Err != nil {
		assetScan.Plugins.Status = apitypes.NewScannerStatus(
			apitypes.ScannerStatusStateFailed,
			apitypes.ScannerStatusReasonError,
			to.Ptr(res.Err.Error()),
		)
	} else {
		pluginResults, ok := res.Result.(*plugins.Results)
		if !ok {
			assetScan.Plugins.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateFailed,
				apitypes.ScannerStatusReasonError,
				to.Ptr("failed to convert to plugins results"),
			)
		} else {
			assetScan.Plugins.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateDone,
				apitypes.ScannerStatusReasonSuccess,
				nil,
			)
			assetScan.Plugins.FindingInfos = &pluginResults.Findings
			// TODO Total plugins should be split by type
			assetScan.Summary.TotalPlugins = to.Ptr(pluginResults.GetTotal())
			assetScan.Stats.Plugins = getInputScanStats(pluginResults.Metadata.InputScans)
		}
	}

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func NewVMClarityPresenter(client *apiclient.Client, id AssetScanID) (*VMClarityPresenter, error) {
	if client == nil {
		return nil, errors.New("API client must not be nil")
	}
	return &VMClarityPresenter{
		client:      client,
		assetScanID: id,
	}, nil
}
