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

package windows

import (
	"context"
	"fmt"

	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/scanner/common"
	"github.com/openclarity/vmclarity/scanner/families"
	"github.com/openclarity/vmclarity/scanner/families/sbom/types"
)

const AnalyzerName = "windows"

type Analyzer struct{}

func New(_ context.Context, _ string, _ types.Config) (families.Scanner[*types.ScannerResult], error) {
	return &Analyzer{}, nil
}

// nolint:cyclop
func (a *Analyzer) Scan(ctx context.Context, sourceType common.InputType, userInput string) (*types.ScannerResult, error) {
	logger := log.GetLoggerFromContextOrDefault(ctx)

	// Create Windows registry based on supported input types
	var err error
	var registry *Registry
	if sourceType.IsOneOf(common.FILE) { // Use file location to the registry
		registry, err = NewRegistry(userInput, logger)
	} else if sourceType.IsOneOf(common.ROOTFS, common.DIR) { // Use mount drive as input
		registry, err = NewRegistryForMount(userInput, logger)
	} else {
		return nil, fmt.Errorf("skipping analyzing unsupported source type: %s", sourceType)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open registry: %w", err)
	}
	defer registry.Close()

	// Fetch BOM from registry details
	bom, err := registry.GetBOM()
	if err != nil {
		return nil, fmt.Errorf("failed to get bom from registry: %w", err)
	}

	// Return sbom
	result := types.CreateScannerResult(bom, AnalyzerName, userInput, sourceType)

	return result, nil
}
