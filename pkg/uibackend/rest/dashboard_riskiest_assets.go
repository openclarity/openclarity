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

package rest

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	backendmodels "github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
	"github.com/openclarity/vmclarity/pkg/uibackend/api/models"
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
	exploits, err := s.getRiskiestAssetsForFindingType(reqCtx, backendmodels.EXPLOIT)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("failed to get riskiest assets for exploits: %v", err))
	}

	malware, err := s.getRiskiestAssetsForFindingType(reqCtx, backendmodels.MALWARE)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("failed to get riskiest assets for malware: %v", err))
	}

	misconfigurations, err := s.getRiskiestAssetsForFindingType(reqCtx, backendmodels.MISCONFIGURATION)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("failed to get riskiest assets for misconfigurations: %v", err))
	}

	rootkits, err := s.getRiskiestAssetsForFindingType(reqCtx, backendmodels.ROOTKIT)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("failed to get riskiest assets for rootkits: %v", err))
	}

	secrets, err := s.getRiskiestAssetsForFindingType(reqCtx, backendmodels.SECRET)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("failed to get riskiest assets for secrets: %v", err))
	}

	vulnerabilities, err := s.getRiskiestAssetsForVulnerabilityType(reqCtx)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError,
			fmt.Sprintf("failed to get riskiest assets for vulnerabilities: %v", err))
	}

	return sendResponse(ctx, http.StatusOK, models.RiskiestAssets{
		Exploits:          &exploits,
		Malware:           &malware,
		Misconfigurations: &misconfigurations,
		Rootkits:          &rootkits,
		Secrets:           &secrets,
		Vulnerabilities:   &vulnerabilities,
	})
}

func (s *ServerImpl) getRiskiestAssetsForFindingType(ctx context.Context, findingType backendmodels.ScanType) ([]models.RiskyAsset, error) {
	riskiestAssets, err := s.getRiskiestAssetsPerFinding(ctx, findingType)
	if err != nil {
		return nil, fmt.Errorf("failed to get riskiest assets: %w", err)
	}

	return toAPIRiskyAssets(*riskiestAssets.Items, findingType), nil
}

func (s *ServerImpl) getRiskiestAssetsForVulnerabilityType(ctx context.Context) ([]models.VulnerabilityRiskyAsset, error) {
	assets, err := s.getRiskiestAssetsPerFinding(ctx, backendmodels.VULNERABILITY)
	if err != nil {
		return nil, fmt.Errorf("failed to get riskiest assets: %w", err)
	}

	return toAPIVulnerabilityRiskyAssets(*assets.Items), nil
}

func (s *ServerImpl) getRiskiestAssetsPerFinding(ctx context.Context, findingType backendmodels.ScanType) (*backendmodels.Assets, error) {
	totalFindingField, err := getTotalFindingFieldName(findingType)
	if err != nil {
		return nil, fmt.Errorf("failed to get total findings field name: %w", err)
	}

	riskiestAssets, err := s.BackendClient.GetAssets(ctx, backendmodels.GetAssetsParams{
		Select:  utils.PointerTo(fmt.Sprintf("summary/%s,assetInfo", totalFindingField)),
		Top:     utils.PointerTo(topRiskiestAssetsCount),
		OrderBy: utils.PointerTo(getOrderByOData(totalFindingField)),
		Filter:  utils.PointerTo(fmt.Sprintf("terminatedOn eq null and summary/%s ne null and summary/%s gt 0", totalFindingField, totalFindingField)),
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

func toAPIVulnerabilityRiskyAssets(assets []backendmodels.Asset) []models.VulnerabilityRiskyAsset {
	ret := make([]models.VulnerabilityRiskyAsset, 0, len(assets))

	for _, asset := range assets {
		assetInfo, err := getAssetInfo(asset.AssetInfo)
		if err != nil {
			log.Warningf("Failed to get asset info, skipping asset: %v", err)
			continue
		}

		summary := asset.Summary.TotalVulnerabilities
		ret = append(ret, models.VulnerabilityRiskyAsset{
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

func toAPIRiskyAssets(assets []backendmodels.Asset, findingType backendmodels.ScanType) []models.RiskyAsset {
	ret := make([]models.RiskyAsset, 0, len(assets))

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

		ret = append(ret, models.RiskyAsset{
			AssetInfo: assetInfo,
			Count:     count,
		})
	}

	return ret
}

func getAssetInfo(asset *backendmodels.AssetType) (*models.AssetInfo, error) {
	discriminator, err := asset.ValueByDiscriminator()
	if err != nil {
		return nil, fmt.Errorf("failed to get value by discriminator: %w", err)
	}

	switch info := discriminator.(type) {
	case backendmodels.VMInfo:
		return vmInfoToAssetInfo(info)
	case backendmodels.ContainerInfo:
		return containerInfoToAssetInfo(info)
	case backendmodels.ContainerImageInfo:
		return containerImageInfoToAssetInfo(info)
	default:
		return nil, fmt.Errorf("asset type is not supported (%T)", discriminator)
	}
}

func containerInfoToAssetInfo(info backendmodels.ContainerInfo) (*models.AssetInfo, error) {
	return &models.AssetInfo{
		Name:     info.ContainerName,
		Location: info.Location,
		Type:     utils.PointerTo(models.Container),
	}, nil
}

func containerImageInfoToAssetInfo(info backendmodels.ContainerImageInfo) (*models.AssetInfo, error) {
	location, _ := info.GetFirstRepoDigest()

	return &models.AssetInfo{
		Name:     &info.ImageID,
		Location: &location,
		Type:     utils.PointerTo(models.ContainerImage),
	}, nil
}

func vmInfoToAssetInfo(info backendmodels.VMInfo) (*models.AssetInfo, error) {
	assetType, err := getVMAssetType(info.InstanceProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset type: %w", err)
	}
	return &models.AssetInfo{
		Location: &info.Location,
		Name:     &info.InstanceID,
		Type:     assetType,
	}, nil
}

func getVMAssetType(provider *backendmodels.CloudProvider) (*models.AssetType, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider is nil")
	}
	switch *provider {
	case backendmodels.AWS:
		return utils.PointerTo(models.AWSEC2Instance), nil
	case backendmodels.Azure:
		return utils.PointerTo(models.AzureInstance), nil
	case backendmodels.GCP:
		return utils.PointerTo(models.GCPInstance), nil
	case backendmodels.External:
		return utils.PointerTo(models.ExternalInstance), nil
	case backendmodels.Docker, backendmodels.Kubernetes:
		fallthrough
	default:
		return nil, fmt.Errorf("unsupported provider: %v", *provider)
	}
}

func getCountForFindingType(summary *backendmodels.ScanFindingsSummary, findingType backendmodels.ScanType) (*int, error) {
	switch findingType {
	case backendmodels.EXPLOIT:
		return summary.TotalExploits, nil
	case backendmodels.MALWARE:
		return summary.TotalMalware, nil
	case backendmodels.MISCONFIGURATION:
		return summary.TotalMisconfigurations, nil
	case backendmodels.ROOTKIT:
		return summary.TotalRootkits, nil
	case backendmodels.SECRET:
		return summary.TotalSecrets, nil
	case backendmodels.INFOFINDER, backendmodels.VULNERABILITY, backendmodels.SBOM:
		fallthrough
	default:
		return nil, fmt.Errorf("unsupported finding type: %v", findingType)
	}
}

func getTotalFindingFieldName(findingType backendmodels.ScanType) (string, error) {
	switch findingType {
	case backendmodels.EXPLOIT:
		return totalExploitsSummaryFieldName, nil
	case backendmodels.MALWARE:
		return totalMalwareSummaryFieldName, nil
	case backendmodels.MISCONFIGURATION:
		return totalMisconfigurationsSummaryFieldName, nil
	case backendmodels.ROOTKIT:
		return totalRootkitsSummaryFieldName, nil
	case backendmodels.SECRET:
		return totalSecretsSummaryFieldName, nil
	case backendmodels.VULNERABILITY:
		return totalVulnerabilitiesSummaryFieldName, nil
	case backendmodels.INFOFINDER, backendmodels.SBOM:
		fallthrough
	default:
		return "", fmt.Errorf("unsupported finding type: %v", findingType)
	}
}
