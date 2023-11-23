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
	"sort"
	"sync"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	backendmodels "github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/findingkey"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
	"github.com/openclarity/vmclarity/pkg/uibackend/api/models"
)

const (
	maxFindingsImpactCount = 5
)

var orderedSeveritiesValues = []string{
	string(backendmodels.CRITICAL),
	string(backendmodels.HIGH),
	string(backendmodels.MEDIUM),
	string(backendmodels.LOW),
	string(backendmodels.NEGLIGIBLE),
}

type findingAssetKey struct {
	FindingKey string
	AssetID    string
}

type findingInfoCount struct {
	FindingInfo *backendmodels.Finding_FindingInfo
	AssetCount  int
}

type findingsImpactData struct {
	findingsImpact               models.FindingsImpact
	findingsImpactFetchedChannel chan struct{}
	findingsImpactMutex          sync.RWMutex
	once                         sync.Once
}

func (s *ServerImpl) GetDashboardFindingsImpact(ctx echo.Context) error {
	// Blocking call until data will be fetched at least once.
	select {
	case <-s.findingsImpactFetchedChannel:
	case <-ctx.Request().Context().Done():
		return sendError(ctx, http.StatusRequestTimeout, "request timeout")
	}
	s.findingsImpactMutex.RLock()
	findingsImpact := s.findingsImpact
	s.findingsImpactMutex.RUnlock()

	return sendResponse(ctx, http.StatusOK, findingsImpact)
}

func (s *ServerImpl) recalculateFindingsImpact(ctx context.Context) {
	log.Debugf("Recalculating findings impact...")
	findingsImpact, err := s.getFindingsImpact(ctx)
	if err != nil {
		log.Errorf("failed to get findings impact: %v", err)
	} else {
		s.findingsImpactMutex.Lock()
		s.findingsImpact = findingsImpact
		s.once.Do(func() {
			close(s.findingsImpactFetchedChannel)
		})
		s.findingsImpactMutex.Unlock()
	}
	log.Debugf("Done recalculating findings impact...")
}

func (s *ServerImpl) getFindingsImpact(ctx context.Context) (models.FindingsImpact, error) {
	exploits, err := s.getExploitsFindingImpact(ctx)
	if err != nil {
		return models.FindingsImpact{}, fmt.Errorf("failed to get exploits finding impact: %w", err)
	}
	malware, err := s.getMalwareFindingImpact(ctx)
	if err != nil {
		return models.FindingsImpact{}, fmt.Errorf("failed to get malware finding impact: %w", err)
	}
	misconfigurations, err := s.getMisconfigurationsFindingImpact(ctx)
	if err != nil {
		return models.FindingsImpact{}, fmt.Errorf("failed to get misconfigurations finding impact: %w", err)
	}
	rootkits, err := s.getRootkitsFindingImpact(ctx)
	if err != nil {
		return models.FindingsImpact{}, fmt.Errorf("failed to get rootkits finding impact: %w", err)
	}
	secrets, err := s.getSecretsFindingImpact(ctx)
	if err != nil {
		return models.FindingsImpact{}, fmt.Errorf("failed to get secrets finding impact: %w", err)
	}
	vulnerabilities, err := s.getVulnerabilitiesFindingImpact(ctx)
	if err != nil {
		return models.FindingsImpact{}, fmt.Errorf("failed to get vulnerabilities finding impact: %w", err)
	}
	packages, err := s.getPackagesFindingImpact(ctx)
	if err != nil {
		return models.FindingsImpact{}, fmt.Errorf("failed to get packages finding impact: %w", err)
	}

	return models.FindingsImpact{
		Exploits:          &exploits,
		Malware:           &malware,
		Misconfigurations: &misconfigurations,
		Packages:          &packages,
		Rootkits:          &rootkits,
		Secrets:           &secrets,
		Vulnerabilities:   &vulnerabilities,
	}, nil
}

func (s *ServerImpl) getExploitsFindingImpact(ctx context.Context) ([]models.ExploitFindingImpact, error) {
	var ret []models.ExploitFindingImpact

	findingAssetMapCount, err := s.getFindingToAssetCountMap(ctx, "Exploit")
	if err != nil {
		return nil, fmt.Errorf("failed to get finding to asset count map: %w", err)
	}

	findingInfoCountSlice := getSortedFindingInfoCountSlice(findingAssetMapCount)

	ret, err = createFindingsImpact(findingInfoCountSlice, createExploitFindingImpact)
	if err != nil {
		return nil, fmt.Errorf("failed to create exploit finding impact: %w", err)
	}

	return ret, nil
}

func createExploitFindingImpact(findingInfo *backendmodels.Finding_FindingInfo, count int) (models.ExploitFindingImpact, error) {
	info, err := findingInfo.AsExploitFindingInfo()
	if err != nil {
		return models.ExploitFindingImpact{}, fmt.Errorf("failed to convert finding info to exploit info: %w", err)
	}

	return models.ExploitFindingImpact{
		AffectedAssetsCount: &count,
		Exploit: &models.Exploit{
			CveID:       info.CveID,
			Description: info.Description,
			Name:        info.Name,
			SourceDB:    info.SourceDB,
			Title:       info.Title,
			Urls:        info.Urls,
		},
	}, nil
}

func (s *ServerImpl) getMalwareFindingImpact(ctx context.Context) ([]models.MalwareFindingImpact, error) {
	var ret []models.MalwareFindingImpact

	findingAssetMapCount, err := s.getFindingToAssetCountMap(ctx, "Malware")
	if err != nil {
		return nil, fmt.Errorf("failed to get finding asset map count: %w", err)
	}

	findingInfoCountSlice := getSortedFindingInfoCountSlice(findingAssetMapCount)

	ret, err = createFindingsImpact(findingInfoCountSlice, createMalwareFindingImpact)
	if err != nil {
		return nil, fmt.Errorf("failed to create malware finding impact: %w", err)
	}

	return ret, nil
}

func createMalwareFindingImpact(findingInfo *backendmodels.Finding_FindingInfo, count int) (models.MalwareFindingImpact, error) {
	info, err := findingInfo.AsMalwareFindingInfo()
	if err != nil {
		return models.MalwareFindingImpact{}, fmt.Errorf("failed to convert finding info to malware info: %w", err)
	}

	return models.MalwareFindingImpact{
		AffectedAssetsCount: &count,
		Malware: &models.Malware{
			MalwareName: info.MalwareName,
			MalwareType: info.MalwareType,
			Path:        info.Path,
			RuleName:    info.RuleName,
		},
	}, nil
}

func (s *ServerImpl) getMisconfigurationsFindingImpact(ctx context.Context) ([]models.MisconfigurationFindingImpact, error) {
	var ret []models.MisconfigurationFindingImpact

	findingAssetMapCount, err := s.getFindingToAssetCountMap(ctx, "Misconfiguration")
	if err != nil {
		return nil, fmt.Errorf("failed to get finding asset map count: %w", err)
	}

	findingInfoCountSlice := getSortedFindingInfoCountSlice(findingAssetMapCount)

	ret, err = createFindingsImpact(findingInfoCountSlice, createMisconfigurationFindingImpact)
	if err != nil {
		return nil, fmt.Errorf("failed to create misconfiguration finding impact: %w", err)
	}

	return ret, nil
}

func createMisconfigurationFindingImpact(findingInfo *backendmodels.Finding_FindingInfo, count int) (models.MisconfigurationFindingImpact, error) {
	info, err := findingInfo.AsMisconfigurationFindingInfo()
	if err != nil {
		return models.MisconfigurationFindingImpact{}, fmt.Errorf("failed to convert finding info to misconfiguration info: %w", err)
	}

	return models.MisconfigurationFindingImpact{
		AffectedAssetsCount: &count,
		Misconfiguration: &models.Misconfiguration{
			Message:         info.Message,
			Remediation:     info.Remediation,
			ScannedPath:     info.ScannedPath,
			ScannerName:     info.ScannerName,
			Severity:        toModelsMisconfigurationSeverity(info.Severity),
			TestCategory:    info.TestCategory,
			TestDescription: info.TestDescription,
			TestID:          info.TestID,
		},
	}, nil
}

func toModelsMisconfigurationSeverity(severity *backendmodels.MisconfigurationSeverity) *models.MisconfigurationSeverity {
	return utils.PointerTo(models.MisconfigurationSeverity(*severity))
}

func (s *ServerImpl) getRootkitsFindingImpact(ctx context.Context) ([]models.RootkitFindingImpact, error) {
	var ret []models.RootkitFindingImpact

	findingAssetMapCount, err := s.getFindingToAssetCountMap(ctx, "Rootkit")
	if err != nil {
		return nil, fmt.Errorf("failed to get finding asset map count: %w", err)
	}

	findingInfoCountSlice := getSortedFindingInfoCountSlice(findingAssetMapCount)

	ret, err = createFindingsImpact(findingInfoCountSlice, createRootkitFindingImpact)
	if err != nil {
		return nil, fmt.Errorf("failed to create rootkit finding impact: %w", err)
	}

	return ret, nil
}

func createRootkitFindingImpact(findingInfo *backendmodels.Finding_FindingInfo, count int) (models.RootkitFindingImpact, error) {
	info, err := findingInfo.AsRootkitFindingInfo()
	if err != nil {
		return models.RootkitFindingImpact{}, fmt.Errorf("failed to convert finding info to rootkit info: %w", err)
	}

	return models.RootkitFindingImpact{
		AffectedAssetsCount: &count,
		Rootkit: &models.Rootkit{
			Message:     info.Message,
			RootkitName: info.RootkitName,
			RootkitType: toModelsRootkitType(info.RootkitType),
		},
	}, nil
}

func toModelsRootkitType(rootkitType *backendmodels.RootkitType) *models.RootkitType {
	return utils.PointerTo(models.RootkitType(*rootkitType))
}

func (s *ServerImpl) getSecretsFindingImpact(ctx context.Context) ([]models.SecretFindingImpact, error) {
	var ret []models.SecretFindingImpact

	findingAssetMapCount, err := s.getFindingToAssetCountMap(ctx, "Secret")
	if err != nil {
		return nil, fmt.Errorf("failed to get finding asset map count: %w", err)
	}

	findingInfoCountSlice := getSortedFindingInfoCountSlice(findingAssetMapCount)

	ret, err = createFindingsImpact(findingInfoCountSlice, createSecretFindingImpact)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret finding impact: %w", err)
	}

	return ret, nil
}

func createSecretFindingImpact(findingInfo *backendmodels.Finding_FindingInfo, count int) (models.SecretFindingImpact, error) {
	info, err := findingInfo.AsSecretFindingInfo()
	if err != nil {
		return models.SecretFindingImpact{}, fmt.Errorf("failed to convert finding info to secret info: %w", err)
	}

	return models.SecretFindingImpact{
		AffectedAssetsCount: &count,
		Secret: &models.Secret{
			EndColumn:   info.EndColumn,
			EndLine:     info.EndLine,
			FilePath:    info.FilePath,
			Fingerprint: info.Fingerprint,
			StartColumn: info.StartColumn,
			StartLine:   info.StartLine,
		},
	}, nil
}

func (s *ServerImpl) getVulnerabilitiesFindingImpact(ctx context.Context) ([]models.VulnerabilityFindingImpact, error) {
	var ret []models.VulnerabilityFindingImpact
	var findingInfoCountSlice []findingInfoCount

	// We want to get the results ordered by severity first, so we will fetch findings per severity.
	// Once all severity findings info were collected, we will continue to the next severity only if we didn't reach maxFindingsImpactCount.
	for _, severity := range orderedSeveritiesValues {
		findingAssetMapCount, err := s.getVulnerabilityFindingToAssetCountMap(ctx, severity)
		if err != nil {
			return nil, fmt.Errorf("failed to get finding asset map count: %w", err)
		}

		findingInfoCountSliceForSeverity := getSortedFindingInfoCountSlice(findingAssetMapCount)
		findingInfoCountSlice = append(findingInfoCountSlice, findingInfoCountSliceForSeverity...)
		if len(findingInfoCountSlice) >= maxFindingsImpactCount {
			// We don't need to fetch findings.
			break
		}
	}

	ret, err := createFindingsImpact(findingInfoCountSlice, createVulnerabilityFindingImpact)
	if err != nil {
		return nil, fmt.Errorf("failed to create vulnerability finding impact: %w", err)
	}

	return ret, nil
}

func createVulnerabilityFindingImpact(findingInfo *backendmodels.Finding_FindingInfo, count int) (models.VulnerabilityFindingImpact, error) {
	info, err := findingInfo.AsVulnerabilityFindingInfo()
	if err != nil {
		return models.VulnerabilityFindingImpact{}, fmt.Errorf("failed to convert finding info to vulnerability info: %w", err)
	}

	return models.VulnerabilityFindingImpact{
		AffectedAssetsCount: &count,
		Vulnerability: &models.Vulnerability{
			Cvss:              toModelsVulnerabilityCVSSArray(info.Cvss),
			Severity:          toModelsVulnerabilitySeverity(info.Severity),
			VulnerabilityName: info.VulnerabilityName,
		},
	}, nil
}

func toModelsVulnerabilitySeverity(severity *backendmodels.VulnerabilitySeverity) *models.VulnerabilitySeverity {
	switch *severity {
	case backendmodels.CRITICAL:
		return utils.PointerTo(models.CRITICAL)
	case backendmodels.HIGH:
		return utils.PointerTo(models.HIGH)
	case backendmodels.MEDIUM:
		return utils.PointerTo(models.MEDIUM)
	case backendmodels.LOW:
		return utils.PointerTo(models.LOW)
	case backendmodels.NEGLIGIBLE:
		return utils.PointerTo(models.NEGLIGIBLE)
	default:
		// should not happen in runtime
		panic("unsupported severity")
	}
}

func toModelsVulnerabilityCVSSArray(cvss *[]backendmodels.VulnerabilityCvss) *[]models.VulnerabilityCvss {
	if cvss == nil {
		return nil
	}
	ret := make([]models.VulnerabilityCvss, len(*cvss))
	for i, vulnerabilityCvss := range *cvss {
		ret[i] = toModelsVulnerabilityCVSS(vulnerabilityCvss)
	}
	return &ret
}

func toModelsVulnerabilityCVSS(cvss backendmodels.VulnerabilityCvss) models.VulnerabilityCvss {
	return models.VulnerabilityCvss{
		Metrics: toModelsVulnerabilityCVSSMetrics(cvss.Metrics),
		Vector:  cvss.Vector,
		Version: cvss.Version,
	}
}

func toModelsVulnerabilityCVSSMetrics(metrics *backendmodels.VulnerabilityCvssMetrics) *models.VulnerabilityCvssMetrics {
	if metrics == nil {
		return nil
	}
	return &models.VulnerabilityCvssMetrics{
		BaseScore:           metrics.BaseScore,
		ExploitabilityScore: metrics.ExploitabilityScore,
		ImpactScore:         metrics.ImpactScore,
	}
}

func (s *ServerImpl) getPackagesFindingImpact(ctx context.Context) ([]models.PackageFindingImpact, error) {
	var ret []models.PackageFindingImpact

	findingAssetMapCount, err := s.getFindingToAssetCountMap(ctx, "Package")
	if err != nil {
		return nil, fmt.Errorf("failed to get finding asset map count: %w", err)
	}

	findingInfoCountSlice := getSortedFindingInfoCountSlice(findingAssetMapCount)

	ret, err = createFindingsImpact(findingInfoCountSlice, createPackageFindingImpact)
	if err != nil {
		return nil, fmt.Errorf("failed to create package finding impact: %w", err)
	}

	return ret, nil
}

func createPackageFindingImpact(findingInfo *backendmodels.Finding_FindingInfo, count int) (models.PackageFindingImpact, error) {
	info, err := findingInfo.AsPackageFindingInfo()
	if err != nil {
		return models.PackageFindingImpact{}, fmt.Errorf("failed to convert finding info to package info: %w", err)
	}

	return models.PackageFindingImpact{
		AffectedAssetsCount: &count,
		Package: &models.Package{
			Name:    info.Name,
			Purl:    info.Purl,
			Version: info.Version,
		},
	}, nil
}

func (s *ServerImpl) getVulnerabilityFindingToAssetCountMap(ctx context.Context, severity string) (map[string]findingInfoCount, error) {
	filter := fmt.Sprintf("asset/terminatedOn eq null and findingInfo/objectType eq 'Vulnerability' and findingInfo/severity eq '%s' and invalidatedOn eq null", severity)
	return s.getFindingToAssetCountMapWithFilter(ctx, filter)
}

func (s *ServerImpl) getFindingToAssetCountMap(ctx context.Context, findingType string) (map[string]findingInfoCount, error) {
	filter := fmt.Sprintf("asset/terminatedOn eq null and findingInfo/objectType eq '%s' and invalidatedOn eq null", findingType)
	return s.getFindingToAssetCountMapWithFilter(ctx, filter)
}

func (s *ServerImpl) getFindingToAssetCountMapWithFilter(ctx context.Context, filter string) (map[string]findingInfoCount, error) {
	// Used to make sure we are not counting the same asset more than once for a specific finding.
	findingAssetMap := make(map[findingAssetKey]struct{})
	// Used to count unique assets count for each finding.
	findingToAssetCount := make(map[string]findingInfoCount)
	// We will fetch findings in batches of 100
	top := 100
	skip := 0
	for {
		f, err := s.BackendClient.GetFindings(ctx, backendmodels.GetFindingsParams{
			Filter: &filter,
			Top:    &top,
			Skip:   &skip,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get findings: %w", err)
		}

		findings := *f.Items
		log.Debugf("Got findings %+v", findings)
		if err = processFindings(findings, findingAssetMap, findingToAssetCount); err != nil {
			return nil, fmt.Errorf("failed to process findings: %w", err)
		}

		if len(findings) < top {
			// No more findings to fetch.
			break
		}
		// Update 'skip' to fetch the next 'top' findings.
		skip += top
	}

	log.Debugf("Returning findingToAssetCount: +%v", findingToAssetCount)
	return findingToAssetCount, nil
}

// processFindings - updates the asset count (findingToAssetCount) for each finding.
// findings - list of findings to process.
// findingAssetMap - the current findingAssetMap to avoid counting the same asset.
// findingToAssetCount - the current findingToAssetCount to update the asset count for each finding.
func processFindings(findings []backendmodels.Finding, findingAssetMap map[findingAssetKey]struct{}, findingToAssetCount map[string]findingInfoCount) error {
	for idx, item := range findings {
		fKey, err := findingkey.GenerateFindingKey(item.FindingInfo)
		if err != nil {
			return fmt.Errorf("failed to generate finding key: %w", err)
		}
		fsKey := findingAssetKey{
			FindingKey: fKey,
			AssetID:    item.Asset.Id,
		}
		if _, ok := findingAssetMap[fsKey]; !ok {
			// Mark as seen to avoid counting the same asset more than once for a finding.
			findingAssetMap[fsKey] = struct{}{}
			log.Debugf("New finding info for an asset. fsKey=%+v, current finding to asset count %v", fsKey, findingToAssetCount[fKey])
			if curFindingInfoCount, found := findingToAssetCount[fKey]; !found {
				// First time we see the finding, save finding info and count the asset.
				findingToAssetCount[fKey] = findingInfoCount{
					FindingInfo: findings[idx].FindingInfo,
					AssetCount:  1,
				}
			} else {
				// We already saw the finding, just count the asset.
				findingToAssetCount[fKey] = findingInfoCount{
					FindingInfo: curFindingInfoCount.FindingInfo,
					AssetCount:  curFindingInfoCount.AssetCount + 1,
				}
			}
		} else {
			log.Debugf("Already count asset %q for finding (%+v).", item.Asset.Id, fKey)
		}
	}

	return nil
}

// getSortedFindingInfoCountSlice will return a slice of findingInfoCount desc sorted by AssetCount.
func getSortedFindingInfoCountSlice(findingAssetMapCount map[string]findingInfoCount) []findingInfoCount {
	findingInfoCountSlice := utils.StringKeyMapToArray(findingAssetMapCount)
	sort.Slice(findingInfoCountSlice, func(i, j int) bool {
		return findingInfoCountSlice[i].AssetCount > findingInfoCountSlice[j].AssetCount
	})
	return findingInfoCountSlice
}

func createFindingsImpact[T any](findingInfoCountSlice []findingInfoCount, createFunc func(findingInfo *backendmodels.Finding_FindingInfo, count int) (T, error)) ([]T, error) {
	var ret []T
	for i := 0; i < maxFindingsImpactCount && i < len(findingInfoCountSlice); i++ {
		infoCount := findingInfoCountSlice[i]
		findingImpact, err := createFunc(infoCount.FindingInfo, infoCount.AssetCount)
		if err != nil {
			return nil, fmt.Errorf("failed to create finding impact: %w", err)
		}
		ret = append(ret, findingImpact)
	}
	return ret, nil
}
