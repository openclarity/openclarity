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

package sbom

import (
	"context"
	"errors"
	"fmt"
	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/core/version"
	"github.com/openclarity/vmclarity/scanner/common"
	"github.com/openclarity/vmclarity/scanner/families"
	"github.com/openclarity/vmclarity/scanner/families/sbom/types"
	"github.com/openclarity/vmclarity/scanner/internal/scan_manager"
	"github.com/openclarity/vmclarity/scanner/utils"
	"github.com/openclarity/vmclarity/scanner/utils/converter"
)

type SBOM struct {
	conf types.Config
}

func New(conf types.Config) families.Family[*types.Result] {
	return &SBOM{
		conf: conf,
	}
}

func (s SBOM) GetType() families.FamilyType {
	return families.SBOM
}

// nolint:cyclop
func (s SBOM) Run(ctx context.Context, _ *families.Results) (*types.Result, error) {
	if len(s.conf.Inputs) == 0 {
		return nil, errors.New("inputs list is empty")
	}

	logger := log.GetLoggerFromContextOrDiscard(ctx)

	// Calculate hash in a separate goroutine as it might take a while. In the
	// meantime, run actual SBOM scanning.
	hashCh := make(chan string, 1)
	hashErrCh := make(chan error, 1)
	go func() {
		// TODO: move the logic from cli utils to shared utils
		// TODO: now that we support multiple inputs,
		//  we need to change the fact the mergedResults assumes it is only for 1 input?
		logger.Infof("Generating hash for input %s", s.conf.Inputs[0].Input)

		hash, err := utils.GenerateHash(s.conf.Inputs[0].InputType, s.conf.Inputs[0].Input)
		hashCh <- hash
		hashErrCh <- err
	}()

	// Run all scanners using scan manager
	manager := scan_manager.New(s.conf.AnalyzersList, s.conf, Factory)
	scans, err := manager.Scan(ctx, s.conf.Inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to process inputs for sbom: %w", err)
	}

	// Get hash after processing all the scans
	hash := <-hashCh
	if err := <-hashErrCh; err != nil {
		return nil, fmt.Errorf("failed to generate hash for source %s: %w", s.conf.Inputs[0].Input, err)
	}

	mergedResults := newMergedResults(s.conf.Inputs[0].InputType, hash)

	// Merge scan results
	for _, scan := range scans {
		logger.Infof("Merging result from %q", scan)

		mergedResults.Merge(scan.Result)
	}

	// Merge data from config
	//
	// TODO(ramizpolic): This logic can be moved to a new standalone SBOM scanner
	// that only loads the data from the filesystem (only works with SBOM input type)
	// and returns loaded SBOM to simplify this.
	for i, with := range s.conf.MergeWith {
		logger.Infof("Merging result from %q", with.SbomPath)

		// Import SBOM from file
		cdxBOM, err := converter.GetCycloneDXSBOMFromFile(with.SbomPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get CDX SBOM from path=%s: %w", with.SbomPath, err)
		}

		// Merge result
		result := types.NewScannerResult(cdxBOM, fmt.Sprintf("merge_with_%d", i), with.SbomPath, common.SBOM)
		mergedResults.Merge(result)
	}

	logger.Info("Converting SBOM results...")

	// TODO(sambetts) Expose CreateMergedSBOM as well as
	// CreateMergedSBOMBytes so that we don't need to re-convert it
	mergedSBOMBytes, err := mergedResults.CreateMergedSBOMBytes("cyclonedx-json", version.CommitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create merged output: %w", err)
	}

	cdxBom, err := converter.GetCycloneDXSBOMFromBytes(mergedSBOMBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to load merged output to CDX bom: %w", err)
	}

	// Create result from merged data
	sbom := types.NewResult(mergedResults.Metadata, cdxBom)

	return sbom, nil
}
