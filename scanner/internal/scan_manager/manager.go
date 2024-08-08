// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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

package scan_manager // nolint:revive,stylecheck

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/sourcegraph/conc/pool"

	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/scanner/common"
	"github.com/openclarity/vmclarity/scanner/families"
	familiesutils "github.com/openclarity/vmclarity/scanner/families/utils"
)

type InputScanResult[T any] struct {
	Metadata   families.ScanInputMetadata
	ScanInput  common.ScanInput
	ScanResult T
}

// Manager allows parallelized scan of inputs for a single family scanner factory.
type Manager[CT, RT any] struct {
	config   CT
	scanners []string
	factory  *Factory[CT, RT]
}

func New[CT, RT any](scanners []string, config CT, factory *Factory[CT, RT]) *Manager[CT, RT] {
	return &Manager[CT, RT]{
		config:   config,
		scanners: scanners,
		factory:  factory,
	}
}

func (m *Manager[CT, RT]) Scan(ctx context.Context, inputs []common.ScanInput) ([]InputScanResult[RT], error) {
	logger := log.GetLoggerFromContextOrDefault(ctx)
	logger.WithField("inputs", inputs).Infof("Scanning inputs in progress...")

	// Validate request
	if len(inputs) == 0 {
		return nil, errors.New("no inputs to scan")
	}
	if len(m.scanners) == 0 {
		return nil, errors.New("no scanners available")
	}
	if m.factory == nil {
		return nil, errors.New("invalid scanner factory")
	}

	// Collect all scan-related errors into a single slice
	var scanErrs []error

	// Create worker pool and schedule all scan jobs. Do not cancel on error. Do not
	// limit the number of parallel workers.
	workerPool := pool.NewWithResults[InputScanResult[RT]]().WithContext(ctx)

	for _, scannerName := range m.scanners {
		// Override context with scanner params
		ctx, logger := log.NewContextLoggerOrDefault(ctx, map[string]interface{}{
			"scanner": scannerName,
		})

		// Create scanner, skip scheduling scan tasks if we cannot create the scanner itself.
		scanner, err := m.factory.createScanner(ctx, scannerName, m.config)
		if err != nil {
			logger.WithError(err).Errorf("Failed to create scanner %s", scannerName)
			scanErrs = append(scanErrs, fmt.Errorf("failed to create scanner %s: %w", scannerName, err))
			continue
		}

		// Submit scanner job for each input pair
		for _, input := range inputs {
			workerPool.Go(func(ctx context.Context) (InputScanResult[RT], error) {
				return m.scanInput(ctx, scannerName, scanner, input)
			})
		}
	}

	// Start workers and collect results and errs
	results, err := workerPool.Wait()
	if err != nil {
		scanErrs = append(scanErrs, err)
	}

	// Return error if all jobs failed to return results.
	// TODO: should it be configurable? allow the user to decide failure threshold?
	if len(results) == 0 {
		return nil, errors.Join(scanErrs...) // nolint:wrapcheck
	}

	logger.Infof("Scanning inputs finished with success")

	return results, nil
}

func (m *Manager[CT, RT]) scanInput(ctx context.Context, scannerName string, scanner families.Scanner[RT], input common.ScanInput) (InputScanResult[RT], error) {
	// Override context with scan request params
	ctx, logger := log.NewContextLoggerOrDefault(ctx, map[string]interface{}{
		"scanner": scannerName,
		"input":   input,
	})
	logger.Infof("Scan job %q for input %s in progress...", scannerName, input)

	// Fuzzy start processing to prevent spike requests for each input
	time.Sleep(time.Duration(rand.Int63n(int64(20 * time.Millisecond)))) // nolint:mnd,gosec,wrapcheck

	// Run scan
	startTime := time.Now()
	scanResult, scanErr := scanner.Scan(ctx, input.InputType, input.Input)

	// Create scan metadata
	endTime := time.Now()
	inputSize, _ := familiesutils.GetInputSize(input)
	metadata := families.NewScanInputMetadata(scannerName, startTime, endTime, inputSize, input)

	// Log and handle scan result
	if scanErr != nil {
		logger.WithError(scanErr).WithField("metadata", metadata).
			Warnf("Scan job %q for input %s failed", scannerName, input)

		return InputScanResult[RT]{}, fmt.Errorf("scan job %q for input %s failed, reason: %w", scannerName, input, scanErr)
	}

	logger.WithField("metadata", metadata).
		Infof("Scan job %q for input %s succeeded", scannerName, input)

	return InputScanResult[RT]{
		Metadata:   metadata,
		ScanInput:  input,
		ScanResult: scanResult,
	}, nil
}
