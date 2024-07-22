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
	"fmt"
	"time"

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

// Family defines interface required to fully run a family.
type Family[T any] interface {
	GetType() FamilyType
	Run(context.Context, *Results) (T, error)
}

// Scanner defines implementation of a family scanner. It should be
// concurrently-safe as Scan can be called concurrently.
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
type FamilyNotifier interface {
	FamilyStarted(context.Context, FamilyType) error
	FamilyFinished(context.Context, FamilyResult) error
}

// ScanMetadata is unified metadata Family returns for all family-processed
// inputs and operations.
type ScanMetadata struct {
	Inputs        []ScanInputMetadata `json:"inputs" yaml:"inputs" mapstructure:"inputs"`
	StartTime     time.Time           `json:"start_time" yaml:"start_time" mapstructure:"start_time"`
	EndTime       time.Time           `json:"end_time" yaml:"end_time" mapstructure:"end_time"`
	TotalFindings int                 `json:"total_findings" yaml:"total_findings" mapstructure:"total_findings"`
}

// ScanInputMetadata is metadata Scanner returns for a successfully processed input.
type ScanInputMetadata struct {
	ScannerName   string           `json:"scanner_name" yaml:"scanner_name" mapstructure:"scanner_name"`
	InputType     common.InputType `json:"input_type" yaml:"input_type" mapstructure:"input_type"`
	InputPath     string           `json:"input_path" yaml:"input_path" mapstructure:"input_path"`
	InputSize     int64            `json:"input_size" yaml:"input_size" mapstructure:"input_size"`
	StartTime     time.Time        `json:"start_time" yaml:"start_time" mapstructure:"start_time"`
	EndTime       time.Time        `json:"end_time" yaml:"end_time" mapstructure:"end_time"`
	TotalFindings int              `json:"total_findings" yaml:"total_findings" mapstructure:"total_findings"`
}

func (m ScanInputMetadata) String() string {
	return fmt.Sprintf("Scanner=%s Input=%s:%s InputSize=%d MB Findings=%d",
		m.ScannerName, m.InputType, m.InputPath, m.InputSize, m.TotalFindings,
	)
}
