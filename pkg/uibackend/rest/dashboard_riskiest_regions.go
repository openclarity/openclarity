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
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	backendmodels "github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
	"github.com/openclarity/vmclarity/pkg/uibackend/api/models"
)

func (s *ServerImpl) GetDashboardRiskiestRegions(ctx echo.Context) error {
	assets, err := s.BackendClient.GetAssets(ctx.Request().Context(), backendmodels.GetAssetsParams{
		Filter: utils.PointerTo("terminatedOn eq null and assetInfo/objectType eq 'VMInfo'"),
	})
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get assets: %v", err))
	}

	regionFindings := createRegionFindingsFromAssets(assets)
	return sendResponse(ctx, http.StatusOK, &models.RiskiestRegions{
		Regions: &regionFindings,
	})
}

func createRegionFindingsFromAssets(assets *backendmodels.Assets) []models.RegionFindings {
	// Map regions to findings count per finding type
	findingsPerRegion := make(map[string]*models.FindingsCount)

	// Sum all asset findings counts (the latest findings per asset) to the total region findings count.
	// asset/ScanFindingsSummary should contain the latest results per family.
	for _, asset := range *assets.Items {
		region, err := getAssetRegion(asset)
		if err != nil {
			log.Warnf("Couldn't get asset location, skipping asset: %v", err)
			continue
		}
		if _, ok := findingsPerRegion[region]; !ok {
			findingsPerRegion[region] = &models.FindingsCount{
				Exploits:          utils.PointerTo(0),
				Malware:           utils.PointerTo(0),
				Misconfigurations: utils.PointerTo(0),
				Rootkits:          utils.PointerTo(0),
				Secrets:           utils.PointerTo(0),
				Vulnerabilities:   utils.PointerTo(0),
			}
		}
		regionFindings := findingsPerRegion[region]
		findingsPerRegion[region] = addAssetSummaryToFindingsCount(regionFindings, asset.Summary)
	}

	items := []models.RegionFindings{}
	for region, findings := range findingsPerRegion {
		r := region
		items = append(items, models.RegionFindings{
			FindingsCount: findings,
			RegionName:    &r,
		})
	}

	return items
}

func getAssetRegion(asset backendmodels.Asset) (string, error) {
	discriminator, err := asset.AssetInfo.ValueByDiscriminator()
	if err != nil {
		return "", fmt.Errorf("failed to get value by discriminator: %w", err)
	}

	switch info := discriminator.(type) {
	case backendmodels.VMInfo:
		return getRegionByProvider(info), nil
	default:
		return "", fmt.Errorf("asset type is not supported (%T)", discriminator)
	}
}

func getRegionByProvider(info backendmodels.VMInfo) string {
	if info.InstanceProvider == nil {
		log.Warnf("Instace provider is nil. instance id: %v", info.InstanceID)
		return info.Location
	}
	if *info.InstanceProvider == backendmodels.AWS {
		// AWS location is represented as region/vpc, need to return only the region
		return strings.Split(info.Location, "/")[0]
	}
	// for other clouds, return the location
	return info.Location
}

func addAssetSummaryToFindingsCount(findingsCount *models.FindingsCount, summary *backendmodels.ScanFindingsSummary) *models.FindingsCount {
	if summary == nil {
		return findingsCount
	}

	secrets := *findingsCount.Secrets + utils.IntPointerValOrEmpty(summary.TotalSecrets)
	exploits := *findingsCount.Exploits + utils.IntPointerValOrEmpty(summary.TotalExploits)
	vulnerabilities := *findingsCount.Vulnerabilities + getTotalVulnerabilities(summary.TotalVulnerabilities)
	rootkits := *findingsCount.Rootkits + utils.IntPointerValOrEmpty(summary.TotalRootkits)
	malware := *findingsCount.Malware + utils.IntPointerValOrEmpty(summary.TotalMalware)
	misconfigurations := *findingsCount.Misconfigurations + utils.IntPointerValOrEmpty(summary.TotalMisconfigurations)
	return &models.FindingsCount{
		Exploits:          &exploits,
		Malware:           &malware,
		Misconfigurations: &misconfigurations,
		Rootkits:          &rootkits,
		Secrets:           &secrets,
		Vulnerabilities:   &vulnerabilities,
	}
}

func getTotalVulnerabilities(summary *backendmodels.VulnerabilityScanSummary) int {
	total := 0
	if summary == nil {
		return total
	}
	total += utils.IntPointerValOrEmpty(summary.TotalCriticalVulnerabilities)
	total += utils.IntPointerValOrEmpty(summary.TotalHighVulnerabilities)
	total += utils.IntPointerValOrEmpty(summary.TotalMediumVulnerabilities)
	total += utils.IntPointerValOrEmpty(summary.TotalLowVulnerabilities)
	total += utils.IntPointerValOrEmpty(summary.TotalNegligibleVulnerabilities)

	return total
}
