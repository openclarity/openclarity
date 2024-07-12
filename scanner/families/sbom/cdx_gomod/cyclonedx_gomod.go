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

package cdx_gomod // nolint:revive,stylecheck

import (
	"context"
	"fmt"
	"os"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/generate/mod"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/licensedetect/local"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	zero_log "github.com/rs/zerolog/log"
	"github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/scanner/common"
	"github.com/openclarity/vmclarity/scanner/families"
	"github.com/openclarity/vmclarity/scanner/families/sbom/cdx_gomod/config"
	"github.com/openclarity/vmclarity/scanner/families/sbom/types"
)

const AnalyzerName = "gomod"

type Analyzer struct {
	config config.Config
}

func New(_ context.Context, _ string, _ types.Config) (families.Scanner[*types.ScannerResult], error) {
	return &Analyzer{
		config: config.Config{},
	}, nil
}

func (a *Analyzer) Scan(ctx context.Context, sourceType common.InputType, userInput string) (*types.ScannerResult, error) {
	logger := log.GetLoggerFromContextOrDefault(ctx)

	// Skip this analyser for input types we don't support
	if !sourceType.IsOneOf(common.DIR) {
		return nil, fmt.Errorf("unsupported input type=%s", sourceType)
	}

	zeroLogger := newZeroLogger(logger)
	licenseDetector := local.NewDetector(zeroLogger)

	generator, err := mod.NewGenerator(userInput,
		mod.WithLogger(zeroLogger),
		mod.WithComponentType(cdx.ComponentTypeApplication),
		mod.WithIncludeStdlib(true),
		mod.WithIncludeTestModules(false),
		mod.WithLicenseDetector(licenseDetector))
	if err != nil {
		return nil, fmt.Errorf("failed to create new CycloneDX-gomod generator: %w", err)
	}

	bom, err := generator.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate sbom: %w", err)
	}

	bom.SerialNumber = uuid.New().URN()
	if bom.Metadata == nil {
		bom.Metadata = &cdx.Metadata{}
	}
	// create cycloneDX sbom metadata
	tool, err := buildToolMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to build tool metadata: %w", err)
	}

	bom.Metadata.Timestamp = time.Now().Format(time.RFC3339)
	bom.Metadata.Tools = &cdx.ToolsChoice{
		Components: &[]cdx.Component{*tool},
	}
	assertLicenses(bom)

	result := types.CreateScannerResult(bom, AnalyzerName, userInput, sourceType)

	return result, nil
}

// newZeroLogger returns a zerolog.Logger.
func newZeroLogger(logrusLogger *logrus.Entry) zerolog.Logger {
	logger := zero_log.Output(zerolog.ConsoleWriter{
		Out: os.Stderr,
	})
	zeroLogLevel, err := convertLoglevel(logrusLogger.Level)
	if err != nil {
		logrusLogger.Warnf("Failed to convert logrus log level to zerolog level: %v", err)
	}
	logger = logger.Level(zeroLogLevel)

	return logger
}

// convertLoglevel converts Logrus level to Zerolog level.
func convertLoglevel(logLevel logrus.Level) (zerolog.Level, error) {
	switch logLevel {
	case logrus.PanicLevel:
		return zerolog.PanicLevel, nil
	case logrus.FatalLevel:
		return zerolog.FatalLevel, nil
	case logrus.ErrorLevel:
		return zerolog.ErrorLevel, nil
	case logrus.WarnLevel:
		return zerolog.WarnLevel, nil
	case logrus.InfoLevel:
		return zerolog.InfoLevel, nil
	case logrus.DebugLevel:
		return zerolog.DebugLevel, nil
	case logrus.TraceLevel:
		return zerolog.TraceLevel, nil
	default:
		return zerolog.Disabled, fmt.Errorf("unknown logrus Level: %v", logLevel)
	}
}
