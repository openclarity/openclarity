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

package families

import (
	"github.com/openclarity/vmclarity/scanner/common"
)

var _ FamilyMetadataObject = &FamilyMetadata{}

func (meta *FamilyMetadata) GetAnnotations() map[string]string {
	if meta.Annotations == nil {
		meta.Annotations = map[string]string{}
	}
	return meta.Annotations
}

func (meta *FamilyMetadata) SetAnnotations(annotations map[string]string) {
	meta.Annotations = annotations
}

func (meta *FamilyMetadata) GetScans() map[common.ScanID]ScannerMetadata {
	if meta.Scans == nil {
		meta.Scans = map[common.ScanID]ScannerMetadata{}
	}
	return meta.Scans
}

func (meta *FamilyMetadata) AddScan(scan ScannerMetadata) {
	scans := meta.GetScans()
	scans[scan.ScanID] = scan
	meta.SetScans(scans)
}

func (meta *FamilyMetadata) SetScans(scans map[common.ScanID]ScannerMetadata) {
	meta.Scans = scans
}

func (meta *FamilyMetadata) GetSummary() FamilySummary {
	return meta.Summary
}

func (meta *FamilyMetadata) SetSummary(summary FamilySummary) {
	meta.Summary = summary
}

func (meta *FamilyMetadata) ToMetadata() FamilyMetadata {
	return FamilyMetadata{
		Annotations: deepCopyMap(meta.Annotations),
		Scans:       deepCopyMap(meta.Scans),
		Summary:     meta.Summary,
	}
}

var _ ScannerMetadataObject = &ScannerMetadata{}

func (s *ScannerMetadata) SetScanInfo(scanInfo common.ScanInfo) {
	s.ScanInfo = scanInfo
}

func (s *ScannerMetadata) GetScanInfo() common.ScanInfo {
	return s.ScanInfo
}

func (s *ScannerMetadata) SetSummary(summary ScannerSummary) {
	s.Summary = summary
}

func (s *ScannerMetadata) GetSummary() ScannerSummary {
	return s.Summary
}

func (s *ScannerMetadata) ToMetadata() ScannerMetadata {
	return ScannerMetadata{
		ScanInfo: s.ScanInfo,
		Summary:  s.Summary,
	}
}

func deepCopyMap[K comparable, V any](src map[K]V) map[K]V {
	cpy := map[K]V{}
	for k, v := range src {
		cpy[k] = v
	}
	return cpy
}
