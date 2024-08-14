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

	"github.com/openclarity/openclarity/scanner/common"
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

// ResultStore manages family result data to enable cross-family runs.
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

// Scanner defines implementation of a family scanner.
// Should be safe for concurrent use.
type Scanner[T any] interface {
	Scan(ctx context.Context, sourceType common.InputType, userInput string) (T, error)
}

// FamilyResult defines an object that FamilyNotifier receives on finished Family.Run.
type FamilyResult struct {
	FamilyType FamilyType
	Result     any
	Err        error
}

// FamilyNotifier is used to subscribe to family scanning progress.
// Should be safe for concurrent use.
type FamilyNotifier interface {
	FamilyStarted(ctx context.Context, familyType FamilyType) error
	FamilyFinished(ctx context.Context, resp FamilyResult) error
}

// FamilyMetadataObject is an interface to interact with FamilyMetadata.
type FamilyMetadataObject interface {
	GetAnnotations() map[string]string
	GetScans() []ScannerMetadata
	SetAnnotations(annotations map[string]string)
	SetScans(scans []ScannerMetadata)
	Merge(scan ScannerMetadata)
}

// ScannerMetadataObject is an interface to interact with ScannerMetadata.
type ScannerMetadataObject interface {
	GetScanInfo() common.ScanInfo
	GetTotalFindings() int
	SetScanInfo(scanInfo common.ScanInfo)
	SetTotalFindings(findings int)
}

// FamilyMetadata defines common family-specific result metadata.
type FamilyMetadata struct {
	Annotations map[string]string `json:"annotations"`
	Scans       []ScannerMetadata `json:"scans"`
}

// ScannerMetadata defines common (family) scanner-specific result metadata.
type ScannerMetadata struct {
	common.ScanInfo `json:"info"` // embed scan info

	TotalFindings int `json:"total_findings"`
}
