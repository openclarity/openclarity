// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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

package scanner

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

type VulnerabilityKey string // Unique identification of a vulnerability ID per package (name and version)

type MergedResults struct {
	MergedVulnerabilitiesByKey map[VulnerabilityKey][]MergedVulnerability
	Source                     Source
}

type MergedVulnerability struct {
	ID            string        `json:"id"` // Merged vulnerability ID used in DiffInfo - THIS IS NOT THE CVE ID
	Vulnerability Vulnerability `json:"vulnerability"`
	ScannersInfo  []Info        `json:"scanners"`
	Diffs         []DiffInfo    `json:"diffs"`
}

type DiffInfo struct {
	CompareToID string                 `json:"compareToID"`
	JSONDiff    map[string]interface{} `json:"jsonDiff"`
	ASCIIDiff   string                 `json:"asciiDiff"`
}

func (mv *MergedVulnerability) AppendScannerInfo(info Info) *MergedVulnerability {
	mv.ScannersInfo = append(mv.ScannersInfo, info)
	return mv
}

func (mv *MergedVulnerability) AppendDiffInfo(diff DiffInfo) *MergedVulnerability {
	mv.Diffs = append(mv.Diffs, diff)
	return mv
}

func NewMergedResults() *MergedResults {
	return &MergedResults{
		MergedVulnerabilitiesByKey: make(map[VulnerabilityKey][]MergedVulnerability),
	}
}

func (m *MergedResults) SetHash(hash string) {
	m.Source.Hash = hash
}

func (m *MergedResults) SetName(name string) {
	m.Source.Name = name
}

func (m *MergedResults) SetType(srcType string) {
	m.Source.Type = srcType
}

func (m *MergedResults) SetSource(src Source) {
	m.Source = src
}

// ToSlice returns MergedResults in a slice format and not by key.
func (m *MergedResults) ToSlice() [][]MergedVulnerability {
	ret := make([][]MergedVulnerability, 0)
	for _, vulnerabilities := range m.MergedVulnerabilitiesByKey {
		ret = append(ret, vulnerabilities)
	}

	return ret
}

func (m *MergedResults) Merge(other *Results) *MergedResults {
	otherVulnerabilityByKey := toVulnerabilityByKey(other.Matches)

	// go over other vulnerabilities list
	// 1. merge mutual vulnerabilities
	// 2. add non mutual vulnerabilities
	for key, otherVulnerability := range otherVulnerabilityByKey {
		// look for other vulnerability key in the current merged vulnerabilities list
		if mergedVulnerabilities, ok := m.MergedVulnerabilitiesByKey[key]; !ok {
			// add non mutual vulnerability
			log.Debugf("Adding new vulnerability results from %v. key=%v", other.ScannerInfo, key)
			m.MergedVulnerabilitiesByKey[key] = []MergedVulnerability{*newMergedVulnerability(otherVulnerability, other.ScannerInfo)}
		} else {
			m.MergedVulnerabilitiesByKey[key] = handleVulnerabilityWithExistingKey(mergedVulnerabilities, otherVulnerability, other.ScannerInfo)
		}
	}

	// TODO: what should we do with other.Source
	// Set Source only once
	if m.Source.Type == "" {
		m.Source = other.Source
	}

	return m
}

// handleVulnerabilityWithExistingKey will look for an identical vulnerability for the given otherVulnerability in the mergedVulnerabilities list,
// if identical vulnerability was found, the new scanner info (otherScannerInfo) will be added
// if no identical vulnerability was found, a new MergedVulnerability (with all the differences that was found) will be added.
func handleVulnerabilityWithExistingKey(mergedVulnerabilities []MergedVulnerability, otherVulnerability Vulnerability, otherScannerInfo Info) []MergedVulnerability {
	shouldAppendMergedVulnerabilityCandidate := true
	mergedVulnerabilityCandidate := newMergedVulnerability(otherVulnerability, otherScannerInfo)

	for i := range mergedVulnerabilities {
		diff, err := getDiff(otherVulnerability, mergedVulnerabilities[i].Vulnerability, mergedVulnerabilities[i].ID)
		if err != nil {
			log.Warnf("Failed to calculate diff - keeping both vulnerabilities: %v", err)
		} else if diff != nil {
			// diff found - need to append diff info
			log.Debugf("Vulnerability results from %v is different from %v. diff=%+v", mergedVulnerabilities[i].ScannersInfo, otherScannerInfo, diff)
			mergedVulnerabilityCandidate = mergedVulnerabilityCandidate.AppendDiffInfo(*diff)
		} else {
			// no diff - need to append scanner info
			log.Debugf("Vulnerability results from %v is equal to %v", mergedVulnerabilities[i].ScannersInfo, otherScannerInfo)
			mergedVulnerabilities[i].AppendScannerInfo(otherScannerInfo)
			shouldAppendMergedVulnerabilityCandidate = false
			break
		}
	}

	if shouldAppendMergedVulnerabilityCandidate {
		mergedVulnerabilities = append(mergedVulnerabilities, *mergedVulnerabilityCandidate)
	}

	return mergedVulnerabilities
}

func getDiff(vulnerability, compareToVulnerability Vulnerability, compareToID string) (*DiffInfo, error) {
	compareToVulnerabilityB, err := json.Marshal(sortArrays(compareToVulnerability))
	if err != nil {
		return nil, fmt.Errorf("failed to Marshal. compareToVulnerability=%+v: %w", compareToVulnerability, err)
	}

	vulnerabilityB, err := json.Marshal(sortArrays(vulnerability))
	if err != nil {
		return nil, fmt.Errorf("failed to Marshal. vulnerability=%+v: %w", vulnerability, err)
	}

	differ := gojsondiff.New()
	diff, err := differ.Compare(compareToVulnerabilityB, vulnerabilityB)
	if err != nil {
		return nil, fmt.Errorf("failed to compare vulnerabilities: %w", err)
	}

	// nolint:nilnil
	if !diff.Modified() {
		return nil, nil
	}

	var templateJSON map[string]interface{}
	err = json.Unmarshal(compareToVulnerabilityB, &templateJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to Unmarshal. compareToVulnerabilityB=%s: %w", string(compareToVulnerabilityB), err)
	}

	asciiDiff, err := getASCIIFormatDiff(compareToVulnerabilityB, diff)
	if err != nil {
		return nil, fmt.Errorf("failed to get ascii format diff: %w", err)
	}

	jsonDiff, err := formatter.NewDeltaFormatter().FormatAsJson(diff)
	if err != nil {
		return nil, fmt.Errorf("failed to format delta diff: %w", err)
	}

	// TODO: do we want to ignore some fields in the diff calculation, links for example?

	return &DiffInfo{
		JSONDiff:    jsonDiff,
		ASCIIDiff:   asciiDiff,
		CompareToID: compareToID,
	}, nil
}

func getASCIIFormatDiff(compareToVulnerabilityB []byte, diff gojsondiff.Diff) (string, error) {
	config := formatter.AsciiFormatterConfig{
		ShowArrayIndex: true,
	}

	var compareToVulnerabilityJSON map[string]interface{}
	_ = json.Unmarshal(compareToVulnerabilityB, &compareToVulnerabilityJSON)
	asciiDiff, err := formatter.NewAsciiFormatter(compareToVulnerabilityJSON, config).Format(diff)
	if err != nil {
		return "", fmt.Errorf("failed to format ascii diff: %w", err)
	}

	return asciiDiff, nil
}

func sortArrays(vulnerability Vulnerability) Vulnerability {
	sort.Slice(vulnerability.CVSS, func(i, j int) bool {
		return vulnerability.CVSS[i].Version < vulnerability.CVSS[j].Version
	})
	sort.Strings(vulnerability.Links)
	sort.Strings(vulnerability.Fix.Versions)
	sort.Strings(vulnerability.Package.CPEs)
	sort.Strings(vulnerability.Package.Licenses)
	return vulnerability
}

func toVulnerabilityByKey(matches Matches) map[VulnerabilityKey]Vulnerability {
	ret := make(map[VulnerabilityKey]Vulnerability, len(matches))
	for _, match := range matches {
		key := createVulnerabilityKey(match.Vulnerability)
		if log.IsLevelEnabled(log.DebugLevel) {
			if vul, ok := ret[key]; ok {
				diff, err := getDiff(vul, match.Vulnerability, "")
				if err != nil {
					// nolint:errchkjson
					vulB, _ := json.Marshal(vul)
					// nolint:errchkjson
					newVulB, _ := json.Marshal(match.Vulnerability)
					log.Debugf("Existing vul with the same key %q. vul=%s, newVul=%s", key, vulB, newVulB)
				} else if diff != nil {
					log.Debugf("Existing vul with the same key %q. diff.JSONDiff=%+v", key, diff.JSONDiff)
				} else {
					log.Debugf("Existing vul with the same key %q - no diff", key)
				}
			}
		}
		ret[key] = match.Vulnerability
	}
	return ret
}

func createVulnerabilityKey(vulnerability Vulnerability) VulnerabilityKey {
	return VulnerabilityKey(fmt.Sprintf("%s.%s.%s", vulnerability.ID, vulnerability.Package.Name, vulnerability.Package.Version))
}

func newMergedVulnerability(vulnerability Vulnerability, scannerInfo Info) *MergedVulnerability {
	return &MergedVulnerability{
		ID:            uuid.New().String(),
		Vulnerability: vulnerability,
		ScannersInfo:  []Info{scannerInfo},
	}
}
