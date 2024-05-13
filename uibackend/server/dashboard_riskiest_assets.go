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

package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/uibackend/types"
)

const (
	topRiskiestAssetsCount                       = 5
	totalExploitsSummaryFieldName                = "totalExploits"
	totalMalwareSummaryFieldName                 = "totalMalware"
	totalMisconfigurationsSummaryFieldName       = "totalMisconfigurations"
	totalRootkitsSummaryFieldName                = "totalRootkits"
	totalSecretsSummaryFieldName                 = "totalSecrets"
	totalVulnerabilitiesSummaryFieldName         = "totalVulnerabilities"
	totalCriticalVulnerabilitiesSummaryFieldName = "totalCriticalVulnerabilities"
	totalHighVulnerabilitiesSummaryFieldName     = "totalHighVulnerabilities" // nolint:gosec
	totalMediumVulnerabilitiesSummaryFieldName   = "totalMediumVulnerabilities"
	totalLowVulnerabilitiesFieldName             = "totalLowVulnerabilities"
	totalNegligibleVulnerabilitiesFieldName      = "totalNegligibleVulnerabilities"
)

var orderedSeveritiesFields = []string{
	totalCriticalVulnerabilitiesSummaryFieldName,
	totalHighVulnerabilitiesSummaryFieldName,
	totalMediumVulnerabilitiesSummaryFieldName,
	totalLowVulnerabilitiesFieldName,
	totalNegligibleVulnerabilitiesFieldName,
}

func (s *ServerImpl) GetDashboardRiskiestAssets(ctx echo.Context) error {
	reqCtx := ctx.Request().Context()
	exploits, err := s.getRiskiestAssetsForFindingType(reqCtx, apitypes.EXPLOIT)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("failed to get riskiest assets for exploits: %v", err))
	}

	malware, err := s.getRiskiestAssetsForFindingType(reqCtx, apitypes.MALWARE)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("failed to get riskiest assets for malware: %v", err))
	}

	misconfigurations, err := s.getRiskiestAssetsForFindingType(reqCtx, apitypes.MISCONFIGURATION)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("failed to get riskiest assets for misconfigurations: %v", err))
	}

	rootkits, err := s.getRiskiestAssetsForFindingType(reqCtx, apitypes.ROOTKIT)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("failed to get riskiest assets for rootkits: %v", err))
	}

	secrets, err := s.getRiskiestAssetsForFindingType(reqCtx, apitypes.SECRET)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("failed to get riskiest assets for secrets: %v", err))
	}

	vulnerabilities, err := s.getRiskiestAssetsForVulnerabilityType(reqCtx)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("failed to get riskiest assets for vulnerabilities: %v", err))
	}

	return sendResponse(ctx, http.StatusOK, types.RiskiestAssets{
		Exploits:          &exploits,
		Malware:           &malware,
		Misconfigurations: &misconfigurations,
		Rootkits:          &rootkits,
		Secrets:           &secrets,
		Vulnerabilities:   &vulnerabilities,
	})
}

func (s *ServerImpl) getRiskiestAssetsForFindingType(ctx context.Context, findingType apitypes.ScanType) ([]types.RiskyAsset, error) {
	riskiestAssets, err := s.getRiskiestAssetsPerFinding(ctx, findingType)
	if err != nil {
		return nil, fmt.Errorf("failed to get riskiest assets: %w", err)
	}

	return toAPIRiskyAssets(*riskiestAssets.Items, findingType), nil
}

func (s *ServerImpl) getRiskiestAssetsForVulnerabilityType(ctx context.Context) ([]types.VulnerabilityRiskyAsset, error) {
	totalFindingField, err := getTotalFindingFieldName(apitypes.VULNERABILITY)
	if err != nil {
		return nil, fmt.Errorf("failed to get total findings field name: %w", err)
	}

	totalVulnerabilitiesNotZeroFilter := ""
	for _, field := range orderedSeveritiesFields {
		totalVulnerabilitiesNotZeroFilter += fmt.Sprintf("summary/%s.%s gt 0 or ", totalFindingField, field)
	}
	totalVulnerabilitiesNotZeroFilter = strings.TrimSuffix(totalVulnerabilitiesNotZeroFilter, " or ")

	assets, err := s.Client.GetAssets(ctx, apitypes.GetAssetsParams{
		Select:  to.Ptr(fmt.Sprintf("summary/%s,assetInfo", totalFindingField)),
		Top:     to.Ptr(topRiskiestAssetsCount),
		OrderBy: to.Ptr(getOrderByOData(totalFindingField)),
		Filter:  to.Ptr(fmt.Sprintf("terminatedOn eq null and summary/%s ne null and (%s)", totalFindingField, totalVulnerabilitiesNotZeroFilter)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get assets: %w", err)
	}

	return toAPIVulnerabilityRiskyAssets(*assets.Items), nil
}

func (s *ServerImpl) getRiskiestAssetsPerFinding(ctx context.Context, findingType apitypes.ScanType) (*apitypes.Assets, error) {
	totalFindingField, err := getTotalFindingFieldName(findingType)
	if err != nil {
		return nil, fmt.Errorf("failed to get total findings field name: %w", err)
	}

	riskiestAssets, err := s.Client.GetAssets(ctx, apitypes.GetAssetsParams{
		Select:  to.Ptr(fmt.Sprintf("summary/%s,assetInfo", totalFindingField)),
		Top:     to.Ptr(topRiskiestAssetsCount),
		OrderBy: to.Ptr(getOrderByOData(totalFindingField)),
		Filter:  to.Ptr(fmt.Sprintf("terminatedOn eq null and summary/%s ne null and summary/%s gt 0", totalFindingField, totalFindingField)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get assets: %w", err)
	}

	return riskiestAssets, nil
}

func getOrderByOData(totalFindingField string) string {
	switch totalFindingField {
	case totalVulnerabilitiesSummaryFieldName:
		return getOrderByOdataForVulnerabilities()
	default:
		return fmt.Sprintf("summary/%s desc", totalFindingField)
	}
}

func getOrderByOdataForVulnerabilities() string {
	ret := make([]string, len(orderedSeveritiesFields))
	for i, field := range orderedSeveritiesFields {
		ret[i] = fmt.Sprintf("summary/%s/%s desc", totalVulnerabilitiesSummaryFieldName, field)
	}
	return strings.Join(ret, ",")
}

func toAPIVulnerabilityRiskyAssets(assets []apitypes.Asset) []types.VulnerabilityRiskyAsset {
	ret := make([]types.VulnerabilityRiskyAsset, 0, len(assets))

	for _, asset := range assets {
		assetInfo, err := getAssetInfo(asset.AssetInfo)
		if err != nil {
			log.Warningf("Failed to get asset info, skipping asset: %v", err)
			continue
		}

		summary := asset.Summary.TotalVulnerabilities
		ret = append(ret, types.VulnerabilityRiskyAsset{
			AssetInfo:                      assetInfo,
			CriticalVulnerabilitiesCount:   summary.TotalCriticalVulnerabilities,
			HighVulnerabilitiesCount:       summary.TotalHighVulnerabilities,
			LowVulnerabilitiesCount:        summary.TotalLowVulnerabilities,
			MediumVulnerabilitiesCount:     summary.TotalMediumVulnerabilities,
			NegligibleVulnerabilitiesCount: summary.TotalNegligibleVulnerabilities,
		})
	}

	return ret
}

func toAPIRiskyAssets(assets []apitypes.Asset, findingType apitypes.ScanType) []types.RiskyAsset {
	ret := make([]types.RiskyAsset, 0, len(assets))

	for _, asset := range assets {
		assetInfo, err := getAssetInfo(asset.AssetInfo)
		if err != nil {
			log.Warningf("Failed to get asset info, skipping asset: %v", err)
			continue
		}

		count, err := getCountForFindingType(asset.Summary, findingType)
		if err != nil {
			log.Warningf("Failed to get count from summary, skipping asset (%v/%v): %v", *assetInfo.Location, *assetInfo.Name, err)
			continue
		}

		ret = append(ret, types.RiskyAsset{
			AssetInfo: assetInfo,
			Count:     count,
		})
	}

	return ret
}

func getAssetInfo(asset *apitypes.AssetType) (*types.AssetInfo, error) {
	discriminator, err := asset.ValueByDiscriminator()
	if err != nil {
		return nil, fmt.Errorf("failed to get value by discriminator: %w", err)
	}

	switch info := discriminator.(type) {
	case apitypes.VMInfo:
		return vmInfoToAssetInfo(info)
	case apitypes.ContainerInfo:
		return containerInfoToAssetInfo(info)
	case apitypes.ContainerImageInfo:
		return containerImageInfoToAssetInfo(info)
	default:
		return nil, fmt.Errorf("asset type is not supported (%T)", discriminator)
	}
}

func containerInfoToAssetInfo(info apitypes.ContainerInfo) (*types.AssetInfo, error) {
	return &types.AssetInfo{
		Name:     info.ContainerName,
		Location: info.Location,
		Type:     to.Ptr(types.Container),
	}, nil
}

func containerImageInfoToAssetInfo(info apitypes.ContainerImageInfo) (*types.AssetInfo, error) {
	location, _ := info.GetFirstRepoDigest()

	return &types.AssetInfo{
		Name:     &info.ImageID,
		Location: &location,
		Type:     to.Ptr(types.ContainerImage),
	}, nil
}

func vmInfoToAssetInfo(info apitypes.VMInfo) (*types.AssetInfo, error) {
	assetType, err := getVMAssetType(info.InstanceProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset type: %w", err)
	}
	return &types.AssetInfo{
		Location: &info.Location,
		Name:     &info.InstanceID,
		Type:     assetType,
	}, nil
}

func getVMAssetType(provider *apitypes.CloudProvider) (*types.AssetType, error) {
	if provider == nil {
		return nil, errors.New("provider is nil")
	}
	switch *provider {
	case apitypes.AWS:
		return to.Ptr(types.AWSEC2Instance), nil
	case apitypes.Azure:
		return to.Ptr(types.AzureInstance), nil
	case apitypes.GCP:
		return to.Ptr(types.GCPInstance), nil
	case apitypes.External:
		return to.Ptr(types.ExternalInstance), nil
	case apitypes.Docker, apitypes.Kubernetes:
		fallthrough
	default:
		return nil, fmt.Errorf("unsupported provider: %v", *provider)
	}
}

func getCountForFindingType(summary *apitypes.ScanFindingsSummary, findingType apitypes.ScanType) (*int, error) {
	switch findingType {
	case apitypes.EXPLOIT:
		return summary.TotalExploits, nil
	case apitypes.MALWARE:
		return summary.TotalMalware, nil
	case apitypes.MISCONFIGURATION:
		return summary.TotalMisconfigurations, nil
	case apitypes.ROOTKIT:
		return summary.TotalRootkits, nil
	case apitypes.SECRET:
		return summary.TotalSecrets, nil
	case apitypes.INFOFINDER, apitypes.VULNERABILITY, apitypes.SBOM:
		fallthrough
	default:
		return nil, fmt.Errorf("unsupported finding type: %v", findingType)
	}
}

func getTotalFindingFieldName(findingType apitypes.ScanType) (string, error) {
	switch findingType {
	case apitypes.EXPLOIT:
		return totalExploitsSummaryFieldName, nil
	case apitypes.MALWARE:
		return totalMalwareSummaryFieldName, nil
	case apitypes.MISCONFIGURATION:
		return totalMisconfigurationsSummaryFieldName, nil
	case apitypes.ROOTKIT:
		return totalRootkitsSummaryFieldName, nil
	case apitypes.SECRET:
		return totalSecretsSummaryFieldName, nil
	case apitypes.VULNERABILITY:
		return totalVulnerabilitiesSummaryFieldName, nil
	case apitypes.INFOFINDER, apitypes.SBOM:
		fallthrough
	default:
		return "", fmt.Errorf("unsupported finding type: %v", findingType)
	}
}
