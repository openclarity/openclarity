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

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	backendmodels "github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
	"github.com/openclarity/vmclarity/ui_backend/api/models"
)

func (s *ServerImpl) GetDashboardRiskiestRegions(ctx echo.Context) error {
	targets, err := s.BackendClient.GetTargets(context.TODO(), backendmodels.GetTargetsParams{
		Filter: utils.StringPtr("targetInfo/objectType eq 'VMInfo'"),
	})
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get targets: %v", err))
	}

	regionFindings := createRegionFindingsFromTargets(targets)
	return sendResponse(ctx, http.StatusOK, &models.RiskiestRegions{
		Regions: &regionFindings,
	})
}

func createRegionFindingsFromTargets(targets *backendmodels.Targets) []models.RegionFindings {
	// Map regions to findings count per finding type
	findingsPerRegion := make(map[string]*models.FindingsCount)

	// Sum all asset findings counts (the latest findings per asset) to the total region findings count.
	// target/ScanFindingsSummary should contain the latest results per family.
	for _, target := range *targets.Items {
		location, err := getTargetLocation(target)
		if err != nil {
			log.Warnf("Couldn't get target location, skipping target: %v", err)
			continue
		}
		if _, ok := findingsPerRegion[location]; !ok {
			findingsPerRegion[location] = &models.FindingsCount{
				Exploits:          utils.PointerTo(0),
				Malware:           utils.PointerTo(0),
				Misconfigurations: utils.PointerTo(0),
				Rootkits:          utils.PointerTo(0),
				Secrets:           utils.PointerTo(0),
				Vulnerabilities:   utils.PointerTo(0),
			}
		}
		regionFindings := findingsPerRegion[location]
		findingsPerRegion[location] = addTargetSummaryToFindingsCount(regionFindings, target.Summary)
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

func getTargetLocation(target backendmodels.Target) (string, error) {
	discriminator, err := target.TargetInfo.ValueByDiscriminator()
	if err != nil {
		return "", fmt.Errorf("failed to get value by discriminator: %w", err)
	}

	switch info := discriminator.(type) {
	case backendmodels.VMInfo:
		return info.Location, nil
	default:
		return "", fmt.Errorf("target type is not supported (%T): %w", discriminator, err)
	}
}

func addTargetSummaryToFindingsCount(findingsCount *models.FindingsCount, summary *backendmodels.ScanFindingsSummary) *models.FindingsCount {
	if summary == nil {
		return findingsCount
	}

	secrets := *findingsCount.Secrets + getPointerValOrZero(summary.TotalSecrets)
	exploits := *findingsCount.Exploits + getPointerValOrZero(summary.TotalExploits)
	vulnerabilities := *findingsCount.Vulnerabilities + getTotalVulnerabilities(summary.TotalVulnerabilities)
	rootkits := *findingsCount.Rootkits + getPointerValOrZero(summary.TotalRootkits)
	malware := *findingsCount.Malware + getPointerValOrZero(summary.TotalMalware)
	misconfigurations := *findingsCount.Misconfigurations + getPointerValOrZero(summary.TotalMisconfigurations)
	return &models.FindingsCount{
		Exploits:          &exploits,
		Malware:           &malware,
		Misconfigurations: &misconfigurations,
		Rootkits:          &rootkits,
		Secrets:           &secrets,
		Vulnerabilities:   &vulnerabilities,
	}
}

func getPointerValOrZero(val *int) int {
	if val == nil {
		return 0
	}
	return *val
}

func getTotalVulnerabilities(summary *backendmodels.VulnerabilityScanSummary) int {
	total := 0
	if summary == nil {
		return total
	}
	total += getPointerValOrZero(summary.TotalCriticalVulnerabilities)
	total += getPointerValOrZero(summary.TotalHighVulnerabilities)
	total += getPointerValOrZero(summary.TotalMediumVulnerabilities)
	total += getPointerValOrZero(summary.TotalLowVulnerabilities)
	total += getPointerValOrZero(summary.TotalNegligibleVulnerabilities)

	return total
}
