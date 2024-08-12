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

package cisdocker

import (
	"context"
	"fmt"

	dockle_run "github.com/Portshift/dockle/pkg"

	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/scanner/common"
	"github.com/openclarity/vmclarity/scanner/families"
	"github.com/openclarity/vmclarity/scanner/families/misconfiguration/cisdocker/config"
	"github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
)

const ScannerName = "cisdocker"

type Scanner struct {
	config config.Config
}

func New(_ context.Context, _ string, config types.ScannersConfig) (families.Scanner[[]types.Misconfiguration], error) {
	return &Scanner{
		config: config.CISDocker,
	}, nil
}

func (a *Scanner) Scan(ctx context.Context, inputType common.InputType, userInput string) ([]types.Misconfiguration, error) {
	// Validate this is an input type supported by the scanner,
	// otherwise return skipped.
	if !inputType.IsOneOf(common.IMAGE, common.DOCKERARCHIVE, common.ROOTFS, common.DIR) {
		return nil, fmt.Errorf("unsupported source type=%s", inputType)
	}

	logger := log.GetLoggerFromContextOrDefault(ctx)

	logger.Infof("Running %s scan on %s...", ScannerName, userInput)

	dockleCfg := createDockleConfig(logger, inputType, userInput, a.config)
	ctx, cancel := context.WithTimeout(ctx, dockleCfg.Timeout)
	defer cancel()

	assessmentMap, err := dockle_run.RunWithContext(ctx, dockleCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to run dockle: %w", err)
	}

	logger.Infof("Successfully scanned %s %s", inputType, userInput)

	misconfigurations := parseDockleReport(inputType, userInput, assessmentMap)

	return misconfigurations, nil
}
