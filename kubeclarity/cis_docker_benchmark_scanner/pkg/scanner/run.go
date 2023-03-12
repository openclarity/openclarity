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

	dockle_config "github.com/Portshift/dockle/config"
	dockle_run "github.com/Portshift/dockle/pkg"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/kubeclarity/cis_docker_benchmark_scanner/pkg/config"
	"github.com/openclarity/kubeclarity/cis_docker_benchmark_scanner/pkg/report"
	"github.com/openclarity/kubeclarity/runtime_scan/api/client/models"
)

func createDockleConfig(scanConfig *config.Config) *dockle_config.Config {
	var username, password string
	if len(scanConfig.Registry.Auths) > 0 {
		username = scanConfig.Registry.Auths[0].Username
		password = scanConfig.Registry.Auths[0].Password
	}

	return &dockle_config.Config{
		Debug:     log.GetLevel() == log.DebugLevel,
		Timeout:   scanConfig.Timeout,
		Username:  username,
		Password:  password,
		Insecure:  scanConfig.Registry.SkipVerifyTLS,
		NonSSL:    scanConfig.Registry.UseHTTP,
		ImageName: scanConfig.ImageIDToScan,
	}
}

func Run() {
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Infof("Docker CIS Benchmark Scanner job is running. imageID=%v", conf.ImageIDToScan)
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
		log.Infof("Docker CIS Benchmark Scanner job was done.")
	case s := <-sig:
		log.Warningf("Received a termination signal: %v", s)
	}
}

func run(ctx context.Context, conf *config.Config) {
	logger := createLogger(conf)
	reporter := report.CreateReporter(ctx, conf)

	// nolint:contextcheck
	assessmentMap, err := dockle_run.RunFromConfig(createDockleConfig(conf))
	if err != nil {
		reportError(reporter, logger, fmt.Errorf("failed to run dockle: %w", err))
		return
	}

	logger.Infof("Image was scanned.")
	logger.Debugf("assessmentMap=%+v", assessmentMap)

	err = reporter.ReportScanResults(assessmentMap)
	if err != nil {
		reportError(reporter, logger, fmt.Errorf("failed to report on successful scan: %v", err))
		return
	}

	logger.Infof("Scan results was reported.")
}

func createLogger(conf *config.Config) *log.Entry {
	logger := log.New()
	logger.SetLevel(log.GetLevel())
	return logger.WithFields(log.Fields{"scan-uuid": conf.ScanUUID, "image-id": conf.ImageIDToScan})
}

func reportError(reporter report.Reporter, logger *log.Entry, errString error) {
	logger.Error(errString)
	if err := reporter.ReportScanError(&models.ScanError{
		Message: errString.Error(),
		Type:    models.ErrorTypeTBD, // TODO: XXX need to use dockle_types.ConvertError(errMsg)
	}); err != nil {
		logger.Errorf("failed to report on scan error: %v", err)
	}
}
