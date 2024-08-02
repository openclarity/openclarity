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

// FamilyNotification defines an object that FamilyNotifier receives on finished Family.Run.
type FamilyNotification struct {
	FamilyType FamilyType
	Result     any
	Err        error
}

// FamilyNotifier is used to subscribe to family scanning progress.
// Implementation should be concurrently-safe.
type FamilyNotifier interface {
	FamilyStarted(context.Context, FamilyType) error
	FamilyFinished(context.Context, FamilyNotification) error
}

// Family defines interface required to fully run a family.
type Family[T any] interface {
	GetType() FamilyType
	Run(context.Context, *Results) (T, error)
}

// FamilySummary defines shared Family result summary data.
type FamilySummary struct {
	FindingsCount int `json:"findings_count"`

	// can be extended with additional general or family-specific properties
	// e.g. PluginScansCount *int
}

// FamilyMetadata defines shared Family result metadata.
// Internal business logic should not rely on metadata.
type FamilyMetadata struct {
	Annotations map[string]string `json:"annotations"`
	Scans       []ScannerMetadata `json:"scans"`
	Summary     *FamilySummary    `json:"summary"`

	// can be extended with additional general or family-specific properties
	// e.g. PluginScannerVersions *map[string]string
}

// Scanner defines implementation of a family scanner. It should be
// concurrently-safe as Scan can be called concurrently.
type Scanner[T ScannerResulter] interface {
	Scan(ctx context.Context, sourceType common.InputType, userInput string) (T, error)
}

// ScannerResulter defines scanner-specific result interface.
type ScannerResulter interface {
	PatchMetadata(scan common.ScanMetadata)
}

// ScannerSummary defines shared Scanner result summary data.
type ScannerSummary struct {
	FindingsCount int `json:"findings_count"`

	// can be extended with additional general or scanner-specific properties
	// e.g. MalwareScannerVersion *string
}

// ScannerMetadata defines shared Scanner result metadata.
// Internal business logic should not rely on metadata.
type ScannerMetadata struct {
	Annotations map[string]string    `json:"annotations"`
	Scan        *common.ScanMetadata `json:"scan"`
	Summary     *ScannerSummary      `json:"summary"`

	// can be extended with additional general or scanner-specific properties
	// e.g. MalwareCount *int
}
