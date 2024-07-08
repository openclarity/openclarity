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

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/openclarity/vmclarity/plugins/sdk-go/plugin"
	"github.com/openclarity/vmclarity/plugins/store/kics/formatter"

	"github.com/Checkmarx/kics/pkg/printer"
	"github.com/Checkmarx/kics/pkg/progress"
	"github.com/Checkmarx/kics/pkg/scan"

	"github.com/openclarity/vmclarity/plugins/sdk-go/types"
)

//nolint:containedctx
type Scanner struct {
	status *types.Status
	cancel context.CancelFunc
}

type ScannerConfig struct {
	PreviewLines     int      `json:"preview-lines" yaml:"preview-lines" toml:"preview-lines" hcl:"preview-lines"`
	ReportFormats    []string `json:"report-formats" yaml:"report-formats" toml:"report-formats" hcl:"report-formats"`
	Platform         []string `json:"platform" yaml:"platform" toml:"platform" hcl:"platform"`
	MaxFileSizeFlag  int      `json:"max-file-size-flag" yaml:"max-file-size-flag" toml:"max-file-size-flag" hcl:"max-file-size-flag"`
	DisableSecrets   bool     `json:"disable-secrets" yaml:"disable-secrets" toml:"disable-secrets" hcl:"disable-secrets"`
	QueryExecTimeout int      `json:"query-exec-timeout" yaml:"query-exec-timeout" toml:"query-exec-timeout" hcl:"query-exec-timeout"`
	Silent           bool     `json:"silent" yaml:"silent" toml:"silent" hcl:"silent"`
	Minimal          bool     `json:"minimal" yaml:"minimal" toml:"minimal" hcl:"minimal"`
}

func (s *Scanner) Metadata() *types.Metadata {
	return types.Ptr(types.Metadata{
		Name:    types.Ptr("KICS"),
		Version: types.Ptr("v1.7.13"),
	})
}

func (s *Scanner) GetStatus() *types.Status {
	return s.status
}

func (s *Scanner) SetStatus(newStatus *types.Status) {
	s.status = types.NewScannerStatus(newStatus.State, newStatus.Message)
}

func (s *Scanner) Start(config types.Config) {
	go func() {
		logger := plugin.GetLogger()

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.TimeoutSeconds)*time.Second)
		s.cancel = cancel
		defer cancel()

		logger.Info("Scanner is running...")
		s.SetStatus(types.NewScannerStatus(types.StateRunning, types.Ptr("Scanner is running...")))

		clientConfig, err := s.createConfig(config.ScannerConfig)
		if err != nil {
			logger.Error("Failed to parse config file", slog.Any("error", err))
			s.SetStatus(types.NewScannerStatus(types.StateFailed, types.Ptr(fmt.Errorf("failed to parse config file: %w", err).Error())))
			return
		}

		// Used to store the raw outputs of a KICS scan
		rawOutputDir := os.TempDir()

		c, err := scan.NewClient(
			&scan.Parameters{
				Path:             []string{config.InputDir},
				QueriesPath:      []string{"../../../queries"},
				PreviewLines:     clientConfig.PreviewLines,
				ReportFormats:    clientConfig.ReportFormats,
				Platform:         clientConfig.Platform,
				OutputPath:       rawOutputDir,
				MaxFileSizeFlag:  clientConfig.MaxFileSizeFlag,
				DisableSecrets:   clientConfig.DisableSecrets,
				QueryExecTimeout: clientConfig.QueryExecTimeout,
				OutputName:       "kics",
			},
			&progress.PbBuilder{Silent: clientConfig.Silent},
			printer.NewPrinter(clientConfig.Minimal), //nolint:forbidigo
		)
		if err != nil {
			logger.Error("Failed to create KICS client", slog.Any("error", err))
			s.SetStatus(types.NewScannerStatus(types.StateFailed, types.Ptr(fmt.Errorf("failed to create KICS client: %w", err).Error())))
			return
		}

		err = c.PerformScan(ctx)
		if err != nil {
			logger.Error("Failed to perform KICS scan", slog.Any("error", err))
			s.SetStatus(types.NewScannerStatus(types.StateFailed, types.Ptr(fmt.Errorf("failed to perform KICS scan: %w", err).Error())))
			return
		}

		if ctx.Err() != nil {
			logger.Error("The operation timed out", slog.Any("error", ctx.Err()))
			s.SetStatus(types.NewScannerStatus(types.StateFailed, types.Ptr(fmt.Errorf("failed due to timeout %w", ctx.Err()).Error())))
			return
		}

		err = s.formatOutput(rawOutputDir, config.OutputFile, clientConfig.ReportFormats)
		if err != nil {
			logger.Error("Failed to format KICS output", slog.Any("error", err))
			s.SetStatus(types.NewScannerStatus(types.StateFailed, types.Ptr(fmt.Errorf("failed to format KICS output: %w", err).Error())))
			return
		}

		logger.Info("Scanner finished running.")
		s.SetStatus(types.NewScannerStatus(types.StateDone, types.Ptr("Scanner finished running.")))
	}()
}

func (s *Scanner) Stop(_ types.Stop) {
	go func() {
		if s.cancel != nil {
			s.cancel()
		}
	}()
}

//nolint:mnd
func (s *Scanner) createConfig(scannerConfig *string) (*ScannerConfig, error) {
	config := types.Ptr(ScannerConfig{
		PreviewLines:     3,
		ReportFormats:    []string{"json"},
		Platform:         []string{"Ansible", "CloudFormation", "Common", "Crossplane", "Dockerfile", "DockerCompose", "Knative", "Kubernetes", "OpenAPI", "Terraform", "AzureResourceManager", "GRPC", "GoogleDeploymentManager", "Buildah", "Pulumi", "ServerlessFW", "CICD"},
		MaxFileSizeFlag:  100,
		DisableSecrets:   true,
		QueryExecTimeout: 60,
		Silent:           true,
		Minimal:          true,
	})

	if scannerConfig == nil || *scannerConfig == "" {
		return config, nil
	}

	if err := json.Unmarshal([]byte(*scannerConfig), config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON config: %w", err)
	}

	// Ensure JSON format is always included,
	// since it's the only format that can be consumed by VMClarity
	config.ReportFormats = ensureJSONFormat(config.ReportFormats)

	return config, nil
}

func (s *Scanner) formatOutput(rawOutputDir, outputFile string, reportFormats []string) error {
	var wg sync.WaitGroup
	var resultMutex sync.Mutex
	var result types.Result
	errCh := make(chan error, len(reportFormats))
	for _, format := range reportFormats {
		wg.Add(1)

		go func() {
			defer wg.Done()

			switch format {
			case "json":
				summaryJSON, err := formatter.FormatJSONOutput(rawOutputDir)
				if err != nil {
					errCh <- fmt.Errorf("failed to format JSON output: %w", err)
				}

				misconfigurations, err := formatter.FormatVMClarityOutput(summaryJSON)
				if err != nil {
					errCh <- fmt.Errorf("failed to format VMClarity output: %w", err)
				}

				resultMutex.Lock()
				result.RawJSON = summaryJSON
				result.Vmclarity.Misconfigurations = misconfigurations
				resultMutex.Unlock()

			case "sarif":
				summarySarif, err := formatter.FormatSarifOutput(rawOutputDir)
				if err != nil {
					errCh <- fmt.Errorf("failed to format Sarif output: %w", err)
				}

				resultMutex.Lock()
				result.RawSarif = summarySarif
				resultMutex.Unlock()

			default:
				errCh <- fmt.Errorf("unsupported report format: %s", format)
			}
		}()
	}
	wg.Wait()
	close(errCh)

	// Check for errors
	var errs error
	for e := range errCh {
		if e != nil {
			errs = errors.Join(errs, e)
		}
	}
	if errs != nil {
		return errs
	}

	if err := result.Export(outputFile); err != nil {
		return fmt.Errorf("failed to save KICS result: %w", err)
	}

	return nil
}

func ensureJSONFormat(reportFormats []string) []string {
	for _, format := range reportFormats {
		if format == "json" {
			return reportFormats
		}
	}

	return append(reportFormats, "json")
}

func main() {
	plugin.Run(&Scanner{
		status: types.NewScannerStatus(types.StateReady, types.Ptr("Starting scanner...")),
	})
}
