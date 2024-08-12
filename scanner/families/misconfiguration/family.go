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

package misconfiguration

import (
	"context"
	"fmt"

	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/scanner/families"
	"github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
	familiesutils "github.com/openclarity/vmclarity/scanner/families/utils"
	"github.com/openclarity/vmclarity/scanner/internal/scan_manager"
)

type Misconfiguration struct {
	conf types.Config
}

func New(conf types.Config) families.Family[*types.Result] {
	return &Misconfiguration{
		conf: conf,
	}
}

func (m Misconfiguration) GetType() families.FamilyType {
	return families.Misconfiguration
}

func (m Misconfiguration) Run(ctx context.Context, _ *families.Results) (*types.Result, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	// Run all scanners using scan manager
	manager := scan_manager.New(m.conf.ScannersList, m.conf.ScannersConfig, Factory)
	results, err := manager.Scan(ctx, m.conf.Inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to process inputs for misconfigurations: %w", err)
	}

	misconfigurations := types.NewResult()

	// Merge results
	for _, result := range results {
		logger.Infof("Merging result from %q", result.Metadata)

		if familiesutils.ShouldStripInputPath(result.ScanInput.StripPathFromResult, m.conf.StripInputPaths) {
			result.ScanResult = stripPathFromResult(result.ScanResult, result.ScanInput.Input)
		}
		misconfigurations.Merge(result.Metadata, result.ScanResult)
	}

	return misconfigurations, nil
}

// stripPathFromResult strip input path from results wherever it is found.
func stripPathFromResult(items []types.Misconfiguration, path string) []types.Misconfiguration {
	for i := range items {
		items[i].Location = familiesutils.TrimMountPath(items[i].Location, path)
	}

	return items
}
