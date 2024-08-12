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

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/scanner/analyzer"
	"github.com/openclarity/vmclarity/scanner/job_manager"
	"github.com/openclarity/vmclarity/scanner/utils"
)

const AnalyzerName = "windows"

type Analyzer struct {
	name       string
	logger     *log.Entry
	resultChan chan job_manager.Result
}

func New(_ string, _ job_manager.IsConfig, logger *log.Entry, resultChan chan job_manager.Result) job_manager.Job {
	return &Analyzer{
		name:       AnalyzerName,
		logger:     logger.Dup().WithField("analyzer", AnalyzerName),
		resultChan: resultChan,
	}
}

// nolint:cyclop
func (a *Analyzer) Run(ctx context.Context, sourceType utils.SourceType, userInput string) error {
	a.logger.Infof("Called %s analyzer on source %v %v", a.name, sourceType, userInput)

	go func() {
		res := &analyzer.Results{}

		// Create Windows registry based on supported input types
		var err error
		var registry *Registry
		switch sourceType {
		case utils.FILE: // Use file location to the registry
			registry, err = NewRegistry(userInput, a.logger)
		case utils.ROOTFS, utils.DIR: // Use mount drive as input
			registry, err = NewRegistryForMount(userInput, a.logger)
		case utils.SBOM, utils.IMAGE, utils.DOCKERARCHIVE, utils.OCIARCHIVE, utils.OCIDIR: // Unsupported
			fallthrough
		default:
			a.logger.Infof("Skipping analyzing unsupported source type: %s", sourceType)
			a.resultChan <- res
			return
		}
		if err != nil {
			a.setError(res, fmt.Errorf("failed to open registry: %w", err))
			return
		}
		defer registry.Close()

		// Fetch BOM from registry details
		bom, err := registry.GetBOM()
		if err != nil {
			a.setError(res, fmt.Errorf("failed to get bom from registry: %w", err))
			return
		}

		// Return sbom
		res = analyzer.CreateResults(bom, a.name, userInput, sourceType)
		a.logger.Infof("Sending successful results")
		a.resultChan <- res
	}()

	return nil
}

func (a *Analyzer) setError(res *analyzer.Results, err error) {
	res.Error = err
	a.logger.Error(res.Error)
	a.resultChan <- res
}
