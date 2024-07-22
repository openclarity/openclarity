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

package vulnerabilities

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/openclarity/vmclarity/scanner/common"
	"github.com/openclarity/vmclarity/scanner/families"
	"github.com/openclarity/vmclarity/scanner/families/vulnerabilities/types"
	"github.com/openclarity/vmclarity/scanner/internal/scan_manager"

	"github.com/openclarity/vmclarity/core/log"
	sbomtypes "github.com/openclarity/vmclarity/scanner/families/sbom/types"
)

const (
	sbomTempFilePath = "/tmp/sbom"
)

type Vulnerabilities struct {
	conf types.Config
}

func New(conf types.Config) families.Family[*types.Result] {
	return &Vulnerabilities{
		conf: conf,
	}
}

func (v Vulnerabilities) GetType() families.FamilyType {
	return families.Vulnerabilities
}

func (v Vulnerabilities) Run(ctx context.Context, res *families.Results) (*types.Result, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	if v.conf.InputFromSbom {
		logger.Infof("Using input from SBOM results")

		sbomResults, err := families.GetFamilyResult[*sbomtypes.Result](res)
		if err != nil {
			return nil, fmt.Errorf("failed to get sbom results: %w", err)
		}

		sbomBytes, err := sbomResults.EncodeToBytes("cyclonedx-json")
		if err != nil {
			return nil, fmt.Errorf("failed to encode sbom results to bytes: %w", err)
		}

		// TODO: need to avoid writing sbom to file
		if err := os.WriteFile(sbomTempFilePath, sbomBytes, 0o600 /* read & write */); err != nil { // nolint:mnd,gofumpt
			return nil, fmt.Errorf("failed to write sbom to file: %w", err)
		}

		v.conf.Inputs = append(v.conf.Inputs, common.ScanInput{
			Input:     sbomTempFilePath,
			InputType: common.SBOM,
		})
	}

	if len(v.conf.Inputs) == 0 {
		return nil, errors.New("inputs list is empty")
	}

	// Run all scanners using scan manager
	manager := scan_manager.New(v.conf.ScannersList, v.conf, Factory)
	scans, err := manager.Scan(ctx, v.conf.Inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to process inputs for vulnerabilities: %w", err)
	}

	vulnerabilities := types.NewResult()

	// Merge scan results
	for _, scan := range scans {
		logger.Infof("Merging result from %q", scan)

		vulnerabilities.Merge(scan.GetScanInputMetadata(len(scan.Result.Vulnerabilities)), scan.Result)
	}

	// TODO:
	// // Set source values.
	// mergedResults.SetSource(sharedscanner.Source{
	//	Type: "image",
	//	Name: config.ImageIDToScan,
	//	Hash: config.ImageHashToScan,
	// })

	return vulnerabilities, nil
}
