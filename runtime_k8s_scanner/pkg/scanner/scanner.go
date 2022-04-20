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

package scanner

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/cisco-open/kubei/runtime_k8s_scanner/pkg/analyze"
	_config "github.com/cisco-open/kubei/runtime_k8s_scanner/pkg/config"
	"github.com/cisco-open/kubei/runtime_k8s_scanner/pkg/report"
	"github.com/cisco-open/kubei/runtime_k8s_scanner/pkg/sbomdb"
	"github.com/cisco-open/kubei/runtime_k8s_scanner/pkg/scan"
	"github.com/cisco-open/kubei/runtime_k8s_scanner/pkg/version"
	"github.com/cisco-open/kubei/runtime_scan/api/client/models"
	"github.com/cisco-open/kubei/shared/pkg/analyzer"
	"github.com/cisco-open/kubei/shared/pkg/utils/image_helper"
)

const (
	sbomTempFilePath = "/tmp/sbom"
)

func Run() {
	conf, err := _config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Infof("Runtime K8s Scanner job is running. imageID=%v, hash=%v", conf.ImageIDToScan, conf.ImageHashToScan)
	done := make(chan struct{})
	go func() {
		run(ctx, conf)
		done <- struct{}{}
	}()

	// Wait for deactivation
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-done:
		log.Infof("Runtime K8s Scanner job was done.")
	case s := <-sig:
		log.Warningf("Received a termination signal: %v", s)
	}
}

// nolint: cyclop
func run(ctx context.Context, conf *_config.Config) {
	logger := createLogger(conf)
	reporter := report.CreateReporter(ctx, conf)
	sbomDBClient := sbomdb.CreateClient(conf.SBOMDBAddress)

	// Get SBOM from db
	sbomBytes, err := sbomDBClient.Get(ctx, conf.ImageHashToScan)
	if err != nil {
		logger.Warningf("failed to get sbom from db: %v", err)
		sbomBytes = nil
	}

	// If SBOM does not exist
	if sbomBytes == nil {
		// Run analyzers
		var mergedResults *analyzer.MergedResults
		mergedResults, err = analyze.Create(logger).Analyze(conf)
		if err != nil {
			reportError(reporter, logger, fmt.Errorf("failed to analyze image: %v", err))
			return
		}

		// Report scan content analysis.
		if err = reporter.ReportScanContentAnalysis(mergedResults); err != nil {
			logger.Errorf("Failed to report content analysis for image: %v", err)
			// we should continue even if failed to report
		} else {
			logger.Infof("Content analysis was reported.")
		}

		// Create merged sbom.
		sbomBytes, err = mergedResults.CreateMergedSBOMBytes(conf.SharedConfig.Analyzer.OutputFormat, version.Version)
		if err != nil {
			reportError(reporter, logger, fmt.Errorf("failed to create merged sbom: %v", err))
			return
		}

		logger.Infof("Image was analyzed.")
		logger.Debugf("SBOM=\n %+v", string(sbomBytes))
		// Send SBOM to DB
		if err = sbomDBClient.Set(ctx, conf.ImageHashToScan, sbomBytes); err != nil {
			logger.Errorf("Failed to set SBOM for image: %v", err)
			// we should continue even if failed to set
		} else {
			logger.Infof("SBOM was set in SBOM DB.")
		}
	}

	// Save SBOM to temp file
	if err = os.WriteFile(sbomTempFilePath, sbomBytes, 0600 /* read & write */); err != nil { // nolint:gomnd,gofumpt
		reportError(reporter, logger, fmt.Errorf("failed to write sbom to file: %v", err))
		return
	}

	// Scan SBOM
	scanResults, err := scan.Create(logger).Scan(conf, sbomTempFilePath)
	if err != nil {
		reportError(reporter, logger, fmt.Errorf("failed to scan sbom: %v", err))
		return
	}
	logger.Infof("Image was scanned.")
	logger.Debugf("scanResults=%+v", scanResults)

	layerCommands, err := getLayerCommands(conf)
	if err != nil {
		logger.Errorf("Failed to get layer commands: %v", err)
		// we should continue even if failed to get layer commands
	}

	// Send results
	if err = reporter.ReportScanResults(scanResults, layerCommands); err != nil {
		reportError(reporter, logger, fmt.Errorf("failed to report on successful scan: %v", err))
		return
	}

	logger.Infof("Scan results was reported.")
}

func reportError(reporter report.Reporter, logger *log.Entry, errString error) {
	logger.Error(errString)
	if err := reporter.ReportScanError(&models.ScanError{
		Message: errString.Error(),
		Type:    models.ErrorTypeTBD, // TODO: XXX
	}); err != nil {
		logger.Errorf("failed to report on scan error: %v", err)
	}
}

func createLogger(conf *_config.Config) *log.Entry {
	logger := log.New()
	logger.SetLevel(log.GetLevel())
	return logger.WithFields(log.Fields{"scan-uuid": conf.ScanUUID, "image-id": conf.ImageIDToScan})
}

func getLayerCommands(conf *_config.Config) ([]*image_helper.FsLayerCommand, error) {
	layerCommands, err := image_helper.GetImageLayerCommands(conf.ImageIDToScan, conf.SharedConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get commands from image=%s: %v", conf.ImageIDToScan, err)
	}

	return layerCommands, nil
}
