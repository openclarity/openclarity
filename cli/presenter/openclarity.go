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

<<<<<<< HEAD:cli/presenter/openclarity.go
	apiclient "github.com/openclarity/openclarity/api/client"
	apitypes "github.com/openclarity/openclarity/api/types"
	"github.com/openclarity/openclarity/core/to"
	"github.com/openclarity/openclarity/scanner/families"
	exploits "github.com/openclarity/openclarity/scanner/families/exploits/types"
	infofinder "github.com/openclarity/openclarity/scanner/families/infofinder/types"
	malware "github.com/openclarity/openclarity/scanner/families/malware/types"
	misconfiguration "github.com/openclarity/openclarity/scanner/families/misconfiguration/types"
	plugins "github.com/openclarity/openclarity/scanner/families/plugins/types"
	rootkits "github.com/openclarity/openclarity/scanner/families/rootkits/types"
	sbom "github.com/openclarity/openclarity/scanner/families/sbom/types"
	secrets "github.com/openclarity/openclarity/scanner/families/secrets/types"
	vulnerabilities "github.com/openclarity/openclarity/scanner/families/vulnerabilities/types"
	"github.com/openclarity/openclarity/scanner/utils"
=======
	"github.com/oapi-codegen/nullable"
	apiclient "github.com/openclarity/vmclarity/api/client"
	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/scanner/families"
	exploits "github.com/openclarity/vmclarity/scanner/families/exploits/types"
	infofinder "github.com/openclarity/vmclarity/scanner/families/infofinder/types"
	malware "github.com/openclarity/vmclarity/scanner/families/malware/types"
	misconfiguration "github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
	plugins "github.com/openclarity/vmclarity/scanner/families/plugins/types"
	rootkits "github.com/openclarity/vmclarity/scanner/families/rootkits/types"
	sbom "github.com/openclarity/vmclarity/scanner/families/sbom/types"
	secrets "github.com/openclarity/vmclarity/scanner/families/secrets/types"
	vulnerabilities "github.com/openclarity/vmclarity/scanner/families/vulnerabilities/types"
	"github.com/openclarity/vmclarity/scanner/utils"
>>>>>>> b4ca021d (feat: change variables to nullable):cli/presenter/vmclarity.go
)

type AssetScanID = apitypes.AssetScanID

type OpenClarityPresenter struct {
	client *apiclient.Client

	assetScanID apitypes.AssetScanID
}

func NewOpenClarityPresenter(client *apiclient.Client, id AssetScanID) (*OpenClarityPresenter, error) {
	if client == nil {
		return nil, errors.New("API client must not be nil")
	}

	return &OpenClarityPresenter{
		client:      client,
		assetScanID: id,
	}, nil
}

func (o *OpenClarityPresenter) ExportFamilyResult(ctx context.Context, res families.FamilyResult) error {
	var err error

	switch res.FamilyType {
	case families.SBOM:
		err = o.ExportSbomResult(ctx, res)
	case families.Vulnerabilities:
		err = o.ExportVulResult(ctx, res)
	case families.Secrets:
		err = o.ExportSecretsResult(ctx, res)
	case families.Exploits:
		err = o.ExportExploitsResult(ctx, res)
	case families.Misconfiguration:
		err = o.ExportMisconfigurationResult(ctx, res)
	case families.Rootkits:
		err = o.ExportRootkitResult(ctx, res)
	case families.Malware:
		err = o.ExportMalwareResult(ctx, res)
	case families.InfoFinder:
		err = o.ExportInfoFinderResult(ctx, res)
	case families.Plugins:
		err = o.ExportPluginsResult(ctx, res)
	}

	return err
}

func (o *OpenClarityPresenter) ExportSbomResult(ctx context.Context, res families.FamilyResult) error {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
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
		sbomResults, ok := res.Result.(*sbom.Result)
		if !ok {
			assetScan.Sbom.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateFailed,
				apitypes.ScannerStatusReasonError,
				to.Ptr("failed to convert to sbom results"),
			)
		} else {
			packages := ConvertSBOMResultToPackages(sbomResults)
			assetScan.Sbom.Packages = nullable.NewNullableWithValue(packages)
			assetScan.Summary.TotalPackages = to.Ptr(len(packages))
			assetScan.Stats.Sbom = getInputScanStats(sbomResults.Metadata)
			assetScan.Sbom.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateDone,
				apitypes.ScannerStatusReasonSuccess,
				nil,
			)
		}
	}

	err = o.client.PatchAssetScan(ctx, assetScan, o.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (o *OpenClarityPresenter) ExportVulResult(ctx context.Context, res families.FamilyResult) error {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
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
		vulnerabilitiesResults, ok := res.Result.(*vulnerabilities.Result)
		if !ok {
			assetScan.Vulnerabilities.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateFailed,
				apitypes.ScannerStatusReasonError,
				to.Ptr("failed to convert to vulnerabilities results"),
			)
		} else {
			vulnerabilities := ConvertVulnResultToVulnerabilities(vulnerabilitiesResults)
			assetScan.Vulnerabilities.Vulnerabilities = nullable.NewNullableWithValue(vulnerabilities)
			assetScan.Summary.TotalVulnerabilities = utils.GetVulnerabilityTotalsPerSeverity(&vulnerabilities)
			assetScan.Stats.Vulnerabilities = getInputScanStats(vulnerabilitiesResults.Metadata)
			assetScan.Vulnerabilities.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateDone,
				apitypes.ScannerStatusReasonSuccess,
				nil,
			)
		}
	}

	err = o.client.PatchAssetScan(ctx, assetScan, o.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (o *OpenClarityPresenter) ExportSecretsResult(ctx context.Context, res families.FamilyResult) error {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
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
		secretsResults, ok := res.Result.(*secrets.Result)
		if !ok {
			assetScan.Secrets.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateFailed,
				apitypes.ScannerStatusReasonError,
				to.Ptr("failed to convert to secrets results"),
			)
		} else {
			secrets := ConvertSecretsResultToSecrets(secretsResults)
			assetScan.Secrets.Secrets = nullable.NewNullableWithValue(secrets)
			assetScan.Summary.TotalSecrets = to.Ptr(len(secrets))
			assetScan.Stats.Secrets = getInputScanStats(secretsResults.Metadata)
			assetScan.Secrets.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateDone,
				apitypes.ScannerStatusReasonSuccess,
				nil,
			)
		}
	}

	err = o.client.PatchAssetScan(ctx, assetScan, o.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func getInputScanStats(metadata families.FamilyMetadata) *[]apitypes.AssetScanInputScanStats {
	if len(metadata.Scans) == 0 {
		return nil
	}

	ret := make([]apitypes.AssetScanInputScanStats, 0, len(metadata.Scans))
	for _, scan := range metadata.Scans {
		ret = append(ret, apitypes.AssetScanInputScanStats{
			Scanner:       &scan.ScannerName,
			Path:          &scan.InputPath,
			Type:          to.Ptr[string](string(scan.InputType)),
			Size:          &scan.InputSize,
			FindingsCount: &scan.TotalFindings,
			ScanTime: &apitypes.AssetScanScanTime{
				EndTime:   &scan.EndTime,
				StartTime: &scan.StartTime,
			},
		})
	}

	return &ret
}

func (o *OpenClarityPresenter) ExportMalwareResult(ctx context.Context, res families.FamilyResult) error {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
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
		malwareResults, ok := res.Result.(*malware.Result)
		if !ok {
			assetScan.Sbom.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateFailed,
				apitypes.ScannerStatusReasonError,
				to.Ptr("failed to convert to malware results"),
			)
		} else {
			mware, mdata := ConvertMalwareResultToMalwareAndMetadata(malwareResults)
			assetScan.Summary.TotalMalware = to.Ptr(len(mware))
			assetScan.Stats.Malware = getInputScanStats(malwareResults.Metadata)
			assetScan.Malware.Malware = nullable.NewNullableWithValue(mware)
			assetScan.Malware.Metadata = nullable.NewNullableWithValue(mdata)
			assetScan.Malware.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateDone,
				apitypes.ScannerStatusReasonSuccess,
				nil,
			)
		}
	}

	if err = o.client.PatchAssetScan(ctx, assetScan, o.assetScanID); err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (o *OpenClarityPresenter) ExportExploitsResult(ctx context.Context, res families.FamilyResult) error {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
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
		exploitsResults, ok := res.Result.(*exploits.Result)
		if !ok {
			assetScan.Exploits.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateFailed,
				apitypes.ScannerStatusReasonError,
				to.Ptr("failed to convert to exploits results"),
			)
		} else {
			exploits := ConvertExploitsResultToExploits(exploitsResults)
			assetScan.Exploits.Exploits = nullable.NewNullableWithValue(exploits)
			assetScan.Summary.TotalExploits = to.Ptr(len(exploits))
			assetScan.Stats.Exploits = getInputScanStats(exploitsResults.Metadata)
			assetScan.Exploits.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateDone,
				apitypes.ScannerStatusReasonSuccess,
				nil,
			)
		}
	}

	err = o.client.PatchAssetScan(ctx, assetScan, o.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (o *OpenClarityPresenter) ExportMisconfigurationResult(ctx context.Context, res families.FamilyResult) error {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
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
		misconfigurationResults, ok := res.Result.(*misconfiguration.Result)
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
				assetScan.Misconfigurations.Misconfigurations = nullable.NewNullableWithValue(misconfigurations)
				assetScan.Misconfigurations.Scanners = nullable.NewNullableWithValue(scanners)
			}
			assetScan.Summary.TotalMisconfigurations = to.Ptr(len(misconfigurationResults.Misconfigurations))
			assetScan.Stats.Misconfigurations = getInputScanStats(misconfigurationResults.Metadata)
		}
	}

	err = o.client.PatchAssetScan(ctx, assetScan, o.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (o *OpenClarityPresenter) ExportInfoFinderResult(ctx context.Context, res families.FamilyResult) error {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
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
		results, ok := res.Result.(*infofinder.Result)
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
				assetScan.InfoFinder.Infos = nullable.NewNullableWithValue(apiInfoFinder)
				assetScan.InfoFinder.Scanners = nullable.NewNullableWithValue(scanners)
			}
			assetScan.Summary.TotalInfoFinder = to.Ptr(len(results.Infos))
			assetScan.Stats.InfoFinder = getInputScanStats(results.Metadata)
		}
	}

	err = o.client.PatchAssetScan(ctx, assetScan, o.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (o *OpenClarityPresenter) ExportRootkitResult(ctx context.Context, res families.FamilyResult) error {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
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
		rootkitsResults, ok := res.Result.(*rootkits.Result)
		if !ok {
			assetScan.Rootkits.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateFailed,
				apitypes.ScannerStatusReasonError,
				to.Ptr("failed to convert to rootkits results"),
			)
		} else {
			rootkits := ConvertRootkitsResultToRootkits(rootkitsResults)
			assetScan.Rootkits.Rootkits = nullable.NewNullableWithValue(rootkits)
			assetScan.Summary.TotalRootkits = to.Ptr(len(rootkits))
			assetScan.Stats.Rootkits = getInputScanStats(rootkitsResults.Metadata)
			assetScan.Rootkits.Status = apitypes.NewScannerStatus(
				apitypes.ScannerStatusStateDone,
				apitypes.ScannerStatusReasonSuccess,
				nil,
			)
		}
	}

	if err = o.client.PatchAssetScan(ctx, assetScan, o.assetScanID); err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (o *OpenClarityPresenter) ExportPluginsResult(ctx context.Context, res families.FamilyResult) error {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
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
		pluginResults, ok := res.Result.(*plugins.Result)
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
			assetScan.Plugins.FindingInfos = nullable.NewNullableWithValue(pluginResults.Findings)
			// TODO Total plugins should be split by type
			assetScan.Summary.TotalPlugins = to.Ptr(len(pluginResults.Findings))
			assetScan.Stats.Plugins = getInputScanStats(pluginResults.Metadata)
		}
	}

	err = o.client.PatchAssetScan(ctx, assetScan, o.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}
