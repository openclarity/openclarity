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
	"github.com/openclarity/openclarity/scanner/common"
)

var _ FamilyMetadataObject = &FamilyMetadata{}

func (meta *FamilyMetadata) GetAnnotations() map[string]string {
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	return meta.Annotations
}

func (meta *FamilyMetadata) SetAnnotations(annotations map[string]string) {
	meta.Annotations = annotations
}

func (meta *FamilyMetadata) GetScans() []ScannerMetadata {
	return meta.Scans
}

func (meta *FamilyMetadata) SetScans(scans []ScannerMetadata) {
	meta.Scans = scans
}

func (meta *FamilyMetadata) Merge(scan ScannerMetadata) {
	meta.Scans = append(meta.Scans, scan)
}

var _ ScannerMetadataObject = &ScannerMetadata{}

func (s *ScannerMetadata) SetScanInfo(scanInfo common.ScanInfo) {
	s.ScanInfo = scanInfo
}

func (s *ScannerMetadata) GetScanInfo() common.ScanInfo {
	return s.ScanInfo
}

func (s *ScannerMetadata) SetTotalFindings(findings int) {
	s.TotalFindings = findings
}

func (s *ScannerMetadata) GetTotalFindings() int {
	return s.TotalFindings
}
