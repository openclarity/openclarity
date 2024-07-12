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

package runner

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/plugins/runner"
	runnertypes "github.com/openclarity/vmclarity/plugins/runner/types"
	plugintypes "github.com/openclarity/vmclarity/plugins/sdk-go/types"
	"github.com/openclarity/vmclarity/scanner/common"
	"github.com/openclarity/vmclarity/scanner/families"
	"github.com/openclarity/vmclarity/scanner/families/plugins/runner/config"
	"github.com/openclarity/vmclarity/scanner/families/plugins/types"
)

type Scanner struct {
	name   string
	config config.Config
}

func New(_ context.Context, name string, config types.ScannersConfig) (families.Scanner[*types.ScannerResult], error) {
	return &Scanner{
		name:   name,
		config: config[name],
	}, nil
}

func (s *Scanner) Scan(ctx context.Context, inputType common.InputType, userInput string) (*types.ScannerResult, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if !inputType.IsOneOf(common.ROOTFS) {
		return nil, fmt.Errorf("unsupported input type=%v", inputType)
	}

	logger := log.GetLoggerFromContextOrDefault(ctx)

	rr, err := runner.New(ctx, runnertypes.PluginConfig{
		Name:          s.name,
		ImageName:     s.config.ImageName,
		InputDir:      userInput,
		ScannerConfig: s.config.ScannerConfig,
		BinaryMode:    s.config.BinaryMode,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin runner: %w", err)
	}

	finishRunner := func(ctx context.Context) {
		if err := rr.Stop(ctx); err != nil {
			logger.WithError(err).Errorf("failed to stop runner")
		}

		// TODO: add short wait before removing to respect container shutdown procedure

		if err := rr.Remove(ctx); err != nil {
			logger.WithError(err).Errorf("failed to remove runner")
		}
	} //nolint:errcheck

	if err := rr.Start(ctx); err != nil {
		finishRunner(ctx)
		return nil, fmt.Errorf("failed to start plugin runner: %w", err)
	}

	if err := rr.WaitReady(ctx); err != nil {
		finishRunner(ctx)
		return nil, fmt.Errorf("failed to wait for plugin scanner to be ready: %w", err)
	}

	// Get plugin metadata
	metadata, err := rr.Metadata(ctx)
	if err != nil {
		finishRunner(ctx)
		return nil, fmt.Errorf("failed to get plugin scanner metadata: %w", err)
	}

	// Stream logs
	go func() {
		logger := logger.WithField("metadata", map[string]interface{}{
			"name":       to.ValueOrZero(metadata.Name),
			"version":    to.ValueOrZero(metadata.Version),
			"apiVersion": to.ValueOrZero(metadata.ApiVersion),
		}).WithField("plugin", s.config.Name)

		logs, err := rr.Logs(ctx)
		if err != nil {
			logger.WithError(err).Warnf("could not listen for logs on plugin runner")
			return
		}
		defer logs.Close()

		for r := bufio.NewScanner(logs); r.Scan(); {
			logger.Info(r.Text())
		}
	}()

	if err := rr.Run(ctx); err != nil {
		finishRunner(ctx)
		return nil, fmt.Errorf("failed to run plugin scanner: %w", err)
	}

	if err := rr.WaitDone(ctx); err != nil {
		finishRunner(ctx)
		return nil, fmt.Errorf("failed to wait for plugin scanner to finish: %w", err)
	}

	findings, pluginResult, err := s.parseResults(ctx, rr)
	if err != nil {
		finishRunner(ctx)
		return nil, fmt.Errorf("failed to parse plugin scanner results: %w", err)
	}

	finishRunner(ctx)

	return &types.ScannerResult{
		Findings: findings,
		Output:   pluginResult,
	}, nil
}

func (s *Scanner) parseResults(ctx context.Context, runner runnertypes.PluginRunner) ([]apitypes.FindingInfo, plugintypes.Result, error) {
	result, err := runner.Result(ctx)
	if err != nil {
		return nil, plugintypes.Result{}, fmt.Errorf("failed to get plugin scanner result: %w", err)
	}
	defer result.Close()

	b, err := io.ReadAll(result)
	if err != nil {
		return nil, plugintypes.Result{}, fmt.Errorf("failed to read plugin scanner output: %w", err)
	}

	var pluginResult plugintypes.Result
	err = json.Unmarshal(b, &pluginResult)
	if err != nil {
		return nil, plugintypes.Result{}, fmt.Errorf("failed to unmarshal plugin scanner output: %w", err)
	}

	findings, err := apitypes.DefaultPluginAdapter.Result(pluginResult)
	if err != nil {
		return nil, plugintypes.Result{}, fmt.Errorf("failed to convert plugin scanner result to vmclarity findings: %w", err)
	}

	return findings, pluginResult, nil
}
