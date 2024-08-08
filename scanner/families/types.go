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
	"context"
	"github.com/openclarity/vmclarity/scanner/common"
)

type FamilyType string

const (
	SBOM             FamilyType = "sbom"
	Vulnerabilities  FamilyType = "vulnerabilities"
	Secrets          FamilyType = "secrets"
	Rootkits         FamilyType = "rootkits"
	Malware          FamilyType = "malware"
	Misconfiguration FamilyType = "misconfiguration"
	InfoFinder       FamilyType = "infofinder"
	Exploits         FamilyType = "exploits"
	Plugins          FamilyType = "plugins"
)

type ResultStore interface {
	GetFamilyResult(family FamilyType) (any, bool)
	GetAllFamilyResults() []any
	SetFamilyResult(family FamilyType, result any)
}

// Family defines interface required to fully run a family.
type Family[T any] interface {
	GetType() FamilyType
	Run(ctx context.Context, store ResultStore) (T, error)
}

// Scanner defines implementation of a family scanner. It should be
// concurrently-safe as Scan can be called concurrently.
type Scanner[T any] interface {
	Scan(ctx context.Context, sourceType common.InputType, userInput string) (T, error)
}

// FamilyNotifier is used to subscribe to family scanning progress.
// Implementation should be concurrently-safe.
type FamilyNotifier interface {
	FamilyStarted(ctx context.Context, familyType FamilyType) error
	FamilyFinished(ctx context.Context, result FamilyNotifierResult) error
}

type FamilyMetadataObject interface {
	GetAnnotations() map[string]string
	GetScans() map[common.ScanID]ScannerMetadata
	GetSummary() FamilySummary
	SetAnnotations(annotations map[string]string)
	AddScan(scan ScannerMetadata)
	SetScans(scans map[common.ScanID]ScannerMetadata)
	SetSummary(summary FamilySummary)
	ToMetadata() FamilyMetadata
}

type ScannerMetadataObject interface {
	GetScanInfo() common.ScanInfo
	GetSummary() ScannerSummary
	SetScanInfo(scanInfo common.ScanInfo)
	SetSummary(summary ScannerSummary)
	ToMetadata() ScannerMetadata
}

type FamilyMetadataEnricherFunc func(meta FamilyMetadata) FamilyMetadata

type ScannerMetadataEnricherFunc func(meta ScannerMetadata) ScannerMetadata

// FamilyNotifierResult defines an object that FamilyNotifier receives on finished Family.Run.
type FamilyNotifierResult struct {
	FamilyType FamilyType
	Result     any
	Err        error
}

// FamilyMetadata defines common family-specific result metadata.
type FamilyMetadata struct {
	Annotations map[string]string                 `json:"annotations"`
	Scans       map[common.ScanID]ScannerMetadata `json:"scans"`
	Summary     FamilySummary                     `json:"summary"`
}

// FamilySummary defines common family-specific result summary data.
type FamilySummary struct {
	TotalFindings int `json:"total_findings"`
}

// ScannerMetadata defines common (family) scanner-specific result metadata.
type ScannerMetadata struct {
	common.ScanInfo `json:"info"` // embed scan info

	Summary ScannerSummary `json:"summary"`
}

// ScannerSummary defines common (family) scanner-specific result summary data.
type ScannerSummary struct {
	TotalFindings int `json:"total_findings"`
}
