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
	"fmt"
	"os"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/generate/mod"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/licensedetect/local"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	zero_log "github.com/rs/zerolog/log"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/kubeclarity/shared/pkg/analyzer/types"
	"github.com/openclarity/kubeclarity/shared/pkg/config"
	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
	"github.com/openclarity/kubeclarity/shared/pkg/utils"
)

const AnalyzerName = "gomod"

type Analyzer struct {
	name   string
	logger *log.Entry
	config config.GomodConfig
}

func New(conf *config.Config, logger *log.Entry) (job_manager.Job[utils.SourceInput, types.Results], error) {
	return &Analyzer{
		name:   AnalyzerName,
		logger: logger.Dup().WithField("analyzer", AnalyzerName),
		config: config.ConvertToGomodConfig(conf.Analyzer),
	}, nil
}

func (a *Analyzer) Run(sourceInput utils.SourceInput) (types.Results, error) {
	res := types.Results{}
	if sourceInput.Type != utils.DIR {
		a.logger.Infof("Skipping analyze unsupported source type: %s", sourceInput.Type)
		return res, nil
	}

	zeroLogger := newZeroLogger(a.logger)
	licenseDetector := local.NewDetector(zeroLogger)

	generator, err := mod.NewGenerator(sourceInput.Source,
		mod.WithLogger(zeroLogger),
		mod.WithComponentType(cdx.ComponentTypeApplication),
		mod.WithIncludeStdlib(true),
		mod.WithIncludeTestModules(false),
		mod.WithLicenseDetector(licenseDetector))
	if err != nil {
		return res, fmt.Errorf("failed to create new CycloneDX-gomod generator: %v", err)
	}

	bom, err := generator.Generate()
	if err != nil {
		return res, fmt.Errorf("failed to generate sbom: %v", err)
	}

	bom.SerialNumber = uuid.New().URN()
	if bom.Metadata == nil {
		bom.Metadata = &cdx.Metadata{}
	}
	// create cycloneDX sbom metadata
	tool, err := buildToolMetadata()
	if err != nil {
		return res, fmt.Errorf("failed to build tool metadata: %v", err)
	}

	bom.Metadata.Timestamp = time.Now().Format(time.RFC3339)
	bom.Metadata.Tools = &[]cdx.Tool{*tool}
	assertLicenses(bom)

	res = types.CreateResults(bom, a.name, sourceInput.Source, sourceInput.Type)
	a.logger.Infof("Sending successful results")
	return res, nil
}

// newZeroLogger returns a zerolog.Logger.
func newZeroLogger(logrusLogger *log.Entry) zerolog.Logger {
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
func convertLoglevel(logLevel log.Level) (zerolog.Level, error) {
	switch logLevel {
	case log.PanicLevel:
		return zerolog.PanicLevel, nil
	case log.FatalLevel:
		return zerolog.FatalLevel, nil
	case log.ErrorLevel:
		return zerolog.ErrorLevel, nil
	case log.WarnLevel:
		return zerolog.WarnLevel, nil
	case log.InfoLevel:
		return zerolog.InfoLevel, nil
	case log.DebugLevel:
		return zerolog.DebugLevel, nil
	case log.TraceLevel:
		return zerolog.TraceLevel, nil
	default:
		return zerolog.Disabled, fmt.Errorf("unknown logrus Level: %v", logLevel)
	}
}
