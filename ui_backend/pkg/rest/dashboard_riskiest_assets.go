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
	"github.com/openclarity/vmclarity/shared/pkg/utils"
	"github.com/openclarity/vmclarity/ui_backend/api/models"
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
	totalHighVulnerabilitiesSummaryFieldName     = "totalHighVulnerabilities"
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
		return nil, fmt.Errorf("failed to get riskiest assets: %v", err)
	}

	return toAPIRiskyAssets(*riskiestAssets.Items, findingType), nil
}

func (s *ServerImpl) getRiskiestAssetsForVulnerabilityType(ctx context.Context) ([]models.VulnerabilityRiskyAsset, error) {
	targets, err := s.getRiskiestAssetsPerFinding(ctx, backendmodels.VULNERABILITY)
	if err != nil {
		return nil, fmt.Errorf("failed to get riskiest assets: %v", err)
	}

	return toAPIVulnerabilityRiskyAssets(*targets.Items), nil
}

func (s *ServerImpl) getRiskiestAssetsPerFinding(ctx context.Context, findingType backendmodels.ScanType) (*backendmodels.Targets, error) {
	totalFindingField, err := getTotalFindingFieldName(findingType)
	if err != nil {
		return nil, fmt.Errorf("failed to get total findings field name: %v", err)
	}

	riskiestAssets, err := s.BackendClient.GetTargets(ctx, backendmodels.GetTargetsParams{
		Select:  utils.PointerTo(fmt.Sprintf("summary/%s,targetInfo", totalFindingField)),
		Top:     utils.PointerTo(topRiskiestAssetsCount),
		OrderBy: utils.PointerTo(getOrderByOData(totalFindingField)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get targets: %v", err)
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

func toAPIVulnerabilityRiskyAssets(targets []backendmodels.Target) []models.VulnerabilityRiskyAsset {
	ret := make([]models.VulnerabilityRiskyAsset, 0, len(targets))

	for _, target := range targets {
		assetInfo, err := getAssetInfo(target.TargetInfo)
		if err != nil {
			log.Warningf("Failed to get asset info, skipping target: %v", err)
			continue
		}

		summary := target.Summary.TotalVulnerabilities
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

func toAPIRiskyAssets(targets []backendmodels.Target, findingType backendmodels.ScanType) []models.RiskyAsset {
	ret := make([]models.RiskyAsset, 0, len(targets))

	for _, target := range targets {
		assetInfo, err := getAssetInfo(target.TargetInfo)
		if err != nil {
			log.Warningf("Failed to get asset info, skipping target: %v", err)
			continue
		}

		count, err := getCountForFindingType(target.Summary, findingType)
		if err != nil {
			log.Warningf("Failed to get count from summary, skipping target (%v/%v): %v", *assetInfo.Location, *assetInfo.Name, err)
			continue
		}

		ret = append(ret, models.RiskyAsset{
			AssetInfo: assetInfo,
			Count:     count,
		})
	}

	return ret
}

func getAssetInfo(target *backendmodels.TargetType) (*models.AssetInfo, error) {
	discriminator, err := target.ValueByDiscriminator()
	if err != nil {
		return nil, fmt.Errorf("failed to get value by discriminator: %w", err)
	}

	switch info := discriminator.(type) {
	case backendmodels.VMInfo:
		return vmInfoToAssetInfo(info)
	default:
		return nil, fmt.Errorf("target type is not supported (%T)", discriminator)
	}
}

func vmInfoToAssetInfo(info backendmodels.VMInfo) (*models.AssetInfo, error) {
	assetType, err := getAssetType(info.InstanceProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset type: %v", err)
	}
	return &models.AssetInfo{
		Location: &info.Location,
		Name:     &info.InstanceID,
		Type:     assetType,
	}, nil
}

func getAssetType(provider *backendmodels.CloudProvider) (*models.AssetType, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider is nil")
	}
	switch *provider {
	case backendmodels.AWS:
		return utils.PointerTo(models.AWSEC2Instance), nil
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
	case backendmodels.VULNERABILITY, backendmodels.SBOM:
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
	case backendmodels.SBOM:
		fallthrough
	default:
		return "", fmt.Errorf("unsupported finding type: %v", findingType)
	}
}
