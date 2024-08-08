// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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

package meta_enricher // nolint:revive,stylecheck

import (
	"github.com/openclarity/vmclarity/scanner/families"
	"github.com/openclarity/vmclarity/scanner/internal/scan_manager"
)

type metaEnricher[RT any] struct {
	familyMetadata           families.FamilyMetadataObject
	familyEnricher           families.FamilyMetadataEnricherFunc
	scannerEnricherGenerator func(RT) families.ScannerMetadataEnricherFunc
}

func New[RT any](meta families.FamilyMetadataObject, familyEnricher families.FamilyMetadataEnricherFunc, scannerEnricherGenerator func(RT) families.ScannerMetadataEnricherFunc) *metaEnricher[RT] {
	return &metaEnricher[RT]{
		familyMetadata:           meta,
		familyEnricher:           familyEnricher,
		scannerEnricherGenerator: scannerEnricherGenerator,
	}
}

func (m *metaEnricher[RT]) PatchMetadata(scans []scan_manager.ScanResult[RT]) func() {
	// Enrich metadata
	familyMeta := &families.FamilyMetadata{}
	for _, scan := range scans {
		// Update scanner metadata with scan details
		scannerEnricher := m.scannerEnricherGenerator(scan.Result)
		scannerMeta := scannerEnricher(families.ScannerMetadata{
			ScanInfo: scan.Scan,
			Summary:  families.ScannerSummary{},
		})

		// Propagate scanner metadata to family metadata
		familyMeta.AddScan(scannerMeta)
	}

	// On done, make sure to update family result metadata
	doneFn := func() {
		enrichedMeta := m.familyEnricher(familyMeta.ToMetadata())
		m.familyMetadata.SetAnnotations(enrichedMeta.GetAnnotations())
		m.familyMetadata.SetScans(enrichedMeta.GetScans())
		m.familyMetadata.SetSummary(enrichedMeta.GetSummary())
	}

	return doneFn
}
