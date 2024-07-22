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

package secrets

import (
	"context"
	"fmt"

	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/scanner/families"
	"github.com/openclarity/vmclarity/scanner/families/secrets/types"
	familiesutils "github.com/openclarity/vmclarity/scanner/families/utils"
	"github.com/openclarity/vmclarity/scanner/internal/scan_manager"
)

type Secrets struct {
	conf types.Config
}

func New(conf types.Config) families.Family[*types.Result] {
	return &Secrets{
		conf: conf,
	}
}

func (s Secrets) GetType() families.FamilyType {
	return families.Secrets
}

func (s Secrets) Run(ctx context.Context, _ *families.Results) (*types.Result, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	// Run all scanners using scan manager
	manager := scan_manager.New(s.conf.ScannersList, s.conf.ScannersConfig, Factory)
	scans, err := manager.Scan(ctx, s.conf.Inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to process inputs for secrets: %w", err)
	}

	secrets := types.NewResult()

	// Merge scan results
	for _, scan := range scans {
		logger.Infof("Merging result from %q", scan)

		if familiesutils.ShouldStripInputPath(scan.StripInputPath, s.conf.StripInputPaths) {
			scan.Result = stripPathFromResult(scan.Result, scan.InputPath)
		}

		secrets.Merge(scan.GetScanInputMetadata(len(scan.Result)), scan.Result)
	}

	return secrets, nil
}

// StripPathFromResult strip input path from results wherever it is found.
func stripPathFromResult(findings []types.Finding, path string) []types.Finding {
	for i := range findings {
		findings[i].File = familiesutils.TrimMountPath(findings[i].File, path)
		findings[i].Fingerprint = familiesutils.RemoveMountPathSubStringIfNeeded(findings[i].Fingerprint, path)
	}

	return findings
}
