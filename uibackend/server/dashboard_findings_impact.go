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
	"fmt"
	"net/http"
	"sort"
	"sync"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/scanner/findingkey"
	"github.com/openclarity/vmclarity/uibackend/types"
)

const (
	maxFindingsImpactCount = 5
)

var orderedSeveritiesValues = []string{
	string(apitypes.CRITICAL),
	string(apitypes.HIGH),
	string(apitypes.MEDIUM),
	string(apitypes.LOW),
	string(apitypes.NEGLIGIBLE),
}

type findingInfoCount struct {
	FindingInfo *apitypes.FindingInfo
	AssetCount  int
}

type findingsImpactData struct {
	findingsImpact               types.FindingsImpact
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

func (s *ServerImpl) getFindingsImpact(ctx context.Context) (types.FindingsImpact, error) {
	exploits, err := s.getExploitsFindingImpact(ctx)
	if err != nil {
		return types.FindingsImpact{}, fmt.Errorf("failed to get exploits finding impact: %w", err)
	}
	malware, err := s.getMalwareFindingImpact(ctx)
	if err != nil {
		return types.FindingsImpact{}, fmt.Errorf("failed to get malware finding impact: %w", err)
	}
	misconfigurations, err := s.getMisconfigurationsFindingImpact(ctx)
	if err != nil {
		return types.FindingsImpact{}, fmt.Errorf("failed to get misconfigurations finding impact: %w", err)
	}
	rootkits, err := s.getRootkitsFindingImpact(ctx)
	if err != nil {
		return types.FindingsImpact{}, fmt.Errorf("failed to get rootkits finding impact: %w", err)
	}
	secrets, err := s.getSecretsFindingImpact(ctx)
	if err != nil {
		return types.FindingsImpact{}, fmt.Errorf("failed to get secrets finding impact: %w", err)
	}
	vulnerabilities, err := s.getVulnerabilitiesFindingImpact(ctx)
	if err != nil {
		return types.FindingsImpact{}, fmt.Errorf("failed to get vulnerabilities finding impact: %w", err)
	}
	packages, err := s.getPackagesFindingImpact(ctx)
	if err != nil {
		return types.FindingsImpact{}, fmt.Errorf("failed to get packages finding impact: %w", err)
	}

	return types.FindingsImpact{
		Exploits:          &exploits,
		Malware:           &malware,
		Misconfigurations: &misconfigurations,
		Packages:          &packages,
		Rootkits:          &rootkits,
		Secrets:           &secrets,
		Vulnerabilities:   &vulnerabilities,
	}, nil
}

func (s *ServerImpl) getExploitsFindingImpact(ctx context.Context) ([]types.ExploitFindingImpact, error) {
	var ret []types.ExploitFindingImpact

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

func createExploitFindingImpact(findingInfo *apitypes.FindingInfo, count int) (types.ExploitFindingImpact, error) {
	info, err := findingInfo.AsExploitFindingInfo()
	if err != nil {
		return types.ExploitFindingImpact{}, fmt.Errorf("failed to convert finding info to exploit info: %w", err)
	}

	return types.ExploitFindingImpact{
		AffectedAssetsCount: &count,
		Exploit: &types.Exploit{
			CveID:       info.CveID,
			Description: info.Description,
			Name:        info.Name,
			SourceDB:    info.SourceDB,
			Title:       info.Title,
			Urls:        info.Urls,
		},
	}, nil
}

func (s *ServerImpl) getMalwareFindingImpact(ctx context.Context) ([]types.MalwareFindingImpact, error) {
	var ret []types.MalwareFindingImpact

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

func createMalwareFindingImpact(findingInfo *apitypes.FindingInfo, count int) (types.MalwareFindingImpact, error) {
	info, err := findingInfo.AsMalwareFindingInfo()
	if err != nil {
		return types.MalwareFindingImpact{}, fmt.Errorf("failed to convert finding info to malware info: %w", err)
	}

	return types.MalwareFindingImpact{
		AffectedAssetsCount: &count,
		Malware: &types.Malware{
			MalwareName: info.MalwareName,
			MalwareType: info.MalwareType,
			Path:        info.Path,
			RuleName:    info.RuleName,
		},
	}, nil
}

func (s *ServerImpl) getMisconfigurationsFindingImpact(ctx context.Context) ([]types.MisconfigurationFindingImpact, error) {
	var ret []types.MisconfigurationFindingImpact

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

func createMisconfigurationFindingImpact(findingInfo *apitypes.FindingInfo, count int) (types.MisconfigurationFindingImpact, error) {
	info, err := findingInfo.AsMisconfigurationFindingInfo()
	if err != nil {
		return types.MisconfigurationFindingImpact{}, fmt.Errorf("failed to convert finding info to misconfiguration info: %w", err)
	}

	return types.MisconfigurationFindingImpact{
		AffectedAssetsCount: &count,
		Misconfiguration: &types.Misconfiguration{
			Message:     info.Message,
			Remediation: info.Remediation,
			Location:    info.Location,
			ScannerName: info.ScannerName,
			Severity:    toModelsMisconfigurationSeverity(info.Severity),
			Category:    info.Category,
			Description: info.Description,
			Id:          info.Id,
		},
	}, nil
}

func toModelsMisconfigurationSeverity(severity *apitypes.MisconfigurationSeverity) *types.MisconfigurationSeverity {
	return to.Ptr(types.MisconfigurationSeverity(*severity))
}

func (s *ServerImpl) getRootkitsFindingImpact(ctx context.Context) ([]types.RootkitFindingImpact, error) {
	var ret []types.RootkitFindingImpact

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

func createRootkitFindingImpact(findingInfo *apitypes.FindingInfo, count int) (types.RootkitFindingImpact, error) {
	info, err := findingInfo.AsRootkitFindingInfo()
	if err != nil {
		return types.RootkitFindingImpact{}, fmt.Errorf("failed to convert finding info to rootkit info: %w", err)
	}

	return types.RootkitFindingImpact{
		AffectedAssetsCount: &count,
		Rootkit: &types.Rootkit{
			Message:     info.Message,
			RootkitName: info.RootkitName,
			RootkitType: toModelsRootkitType(info.RootkitType),
		},
	}, nil
}

func toModelsRootkitType(rootkitType *apitypes.RootkitType) *types.RootkitType {
	return to.Ptr(types.RootkitType(*rootkitType))
}

func (s *ServerImpl) getSecretsFindingImpact(ctx context.Context) ([]types.SecretFindingImpact, error) {
	var ret []types.SecretFindingImpact

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

func createSecretFindingImpact(findingInfo *apitypes.FindingInfo, count int) (types.SecretFindingImpact, error) {
	info, err := findingInfo.AsSecretFindingInfo()
	if err != nil {
		return types.SecretFindingImpact{}, fmt.Errorf("failed to convert finding info to secret info: %w", err)
	}

	return types.SecretFindingImpact{
		AffectedAssetsCount: &count,
		Secret: &types.Secret{
			EndColumn:   info.EndColumn,
			EndLine:     info.EndLine,
			FilePath:    info.FilePath,
			Fingerprint: info.Fingerprint,
			StartColumn: info.StartColumn,
			StartLine:   info.StartLine,
		},
	}, nil
}

func (s *ServerImpl) getVulnerabilitiesFindingImpact(ctx context.Context) ([]types.VulnerabilityFindingImpact, error) {
	var ret []types.VulnerabilityFindingImpact
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

func createVulnerabilityFindingImpact(findingInfo *apitypes.FindingInfo, count int) (types.VulnerabilityFindingImpact, error) {
	info, err := findingInfo.AsVulnerabilityFindingInfo()
	if err != nil {
		return types.VulnerabilityFindingImpact{}, fmt.Errorf("failed to convert finding info to vulnerability info: %w", err)
	}

	return types.VulnerabilityFindingImpact{
		AffectedAssetsCount: &count,
		Vulnerability: &types.Vulnerability{
			Cvss:              toModelsVulnerabilityCVSSArray(info.Cvss),
			Severity:          toModelsVulnerabilitySeverity(info.Severity),
			VulnerabilityName: info.VulnerabilityName,
		},
	}, nil
}

func toModelsVulnerabilitySeverity(severity *apitypes.VulnerabilitySeverity) *types.VulnerabilitySeverity {
	switch *severity {
	case apitypes.CRITICAL:
		return to.Ptr(types.CRITICAL)
	case apitypes.HIGH:
		return to.Ptr(types.HIGH)
	case apitypes.MEDIUM:
		return to.Ptr(types.MEDIUM)
	case apitypes.LOW:
		return to.Ptr(types.LOW)
	case apitypes.NEGLIGIBLE:
		return to.Ptr(types.NEGLIGIBLE)
	default:
		// should not happen in runtime
		panic("unsupported severity")
	}
}

func toModelsVulnerabilityCVSSArray(cvss *[]apitypes.VulnerabilityCvss) *[]types.VulnerabilityCvss {
	if cvss == nil {
		return nil
	}
	ret := make([]types.VulnerabilityCvss, len(*cvss))
	for i, vulnerabilityCvss := range *cvss {
		ret[i] = toModelsVulnerabilityCVSS(vulnerabilityCvss)
	}
	return &ret
}

func toModelsVulnerabilityCVSS(cvss apitypes.VulnerabilityCvss) types.VulnerabilityCvss {
	return types.VulnerabilityCvss{
		Metrics: toModelsVulnerabilityCVSSMetrics(cvss.Metrics),
		Vector:  cvss.Vector,
		Version: cvss.Version,
	}
}

func toModelsVulnerabilityCVSSMetrics(metrics *apitypes.VulnerabilityCvssMetrics) *types.VulnerabilityCvssMetrics {
	if metrics == nil {
		return nil
	}
	return &types.VulnerabilityCvssMetrics{
		BaseScore:           metrics.BaseScore,
		ExploitabilityScore: metrics.ExploitabilityScore,
		ImpactScore:         metrics.ImpactScore,
	}
}

func (s *ServerImpl) getPackagesFindingImpact(ctx context.Context) ([]types.PackageFindingImpact, error) {
	var ret []types.PackageFindingImpact

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

func createPackageFindingImpact(findingInfo *apitypes.FindingInfo, count int) (types.PackageFindingImpact, error) {
	info, err := findingInfo.AsPackageFindingInfo()
	if err != nil {
		return types.PackageFindingImpact{}, fmt.Errorf("failed to convert finding info to package info: %w", err)
	}

	return types.PackageFindingImpact{
		AffectedAssetsCount: &count,
		Package: &types.Package{
			Name:    info.Name,
			Purl:    info.Purl,
			Version: info.Version,
		},
	}, nil
}

func (s *ServerImpl) getVulnerabilityFindingToAssetCountMap(ctx context.Context, severity string) (map[string]findingInfoCount, error) {
	filter := fmt.Sprintf("findingInfo/objectType eq 'Vulnerability' and findingInfo/severity eq '%s'", severity)
	return s.getFindingToAssetCountMapWithFilter(ctx, filter)
}

func (s *ServerImpl) getFindingToAssetCountMap(ctx context.Context, findingType string) (map[string]findingInfoCount, error) {
	filter := fmt.Sprintf("findingInfo/objectType eq '%s'", findingType)
	return s.getFindingToAssetCountMapWithFilter(ctx, filter)
}

func (s *ServerImpl) getFindingToAssetCountMapWithFilter(ctx context.Context, filter string) (map[string]findingInfoCount, error) {
	// Used to count unique assets count for each finding.
	findingToAssetCount := make(map[string]findingInfoCount)
	// We will fetch findings in batches of 100
	top := 100
	skip := 0
	for {
		f, err := s.Client.GetFindings(ctx, apitypes.GetFindingsParams{
			Filter: &filter,
			Top:    &top,
			Skip:   &skip,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get findings: %w", err)
		}

		findings := *f.Items
		log.Debugf("Got findings %+v", findings)

		for idx, finding := range findings {
			assetCount := 0
			assetFindings, err := s.Client.GetAssetFindings(ctx, apitypes.GetAssetFindingsParams{
				Filter: to.Ptr(fmt.Sprintf("finding/id eq '%s' and asset/terminatedOn eq null and invalidatedOn eq null", *finding.Id)),
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get asset findings: %w", err)
			}
			if assetFindings.Items != nil {
				assetCount = len(*assetFindings.Items)
			}

			fKey, err := findingkey.GenerateFindingKey(finding.FindingInfo)
			if err != nil {
				return nil, fmt.Errorf("failed to generate finding key: %w", err)
			}

			findingToAssetCount[fKey] = findingInfoCount{
				FindingInfo: findings[idx].FindingInfo,
				AssetCount:  assetCount,
			}
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

// getSortedFindingInfoCountSlice will return a slice of findingInfoCount desc sorted by AssetCount.
func getSortedFindingInfoCountSlice(findingAssetMapCount map[string]findingInfoCount) []findingInfoCount {
	findingInfoCountSlice := to.Values(findingAssetMapCount)
	sort.Slice(findingInfoCountSlice, func(i, j int) bool {
		return findingInfoCountSlice[i].AssetCount > findingInfoCountSlice[j].AssetCount
	})
	return findingInfoCountSlice
}

func createFindingsImpact[T any](findingInfoCountSlice []findingInfoCount, createFunc func(findingInfo *apitypes.FindingInfo, count int) (T, error)) ([]T, error) {
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
