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

package orchestrator

import (
	"context"
	"fmt"
	"time"

	"github.com/Portshift/go-utils/healthz"

	apiclient "github.com/openclarity/vmclarity/api/client"
	"github.com/openclarity/vmclarity/core/log"
	discovery "github.com/openclarity/vmclarity/orchestrator/discoverer"
	assetscanprocessor "github.com/openclarity/vmclarity/orchestrator/processor/assetscan"
	assetscanwatcher "github.com/openclarity/vmclarity/orchestrator/watcher/assetscan"
	assetscanestimationwatcher "github.com/openclarity/vmclarity/orchestrator/watcher/assetscanestimation"
	scanwatcher "github.com/openclarity/vmclarity/orchestrator/watcher/scan"
	scanconfigwatcher "github.com/openclarity/vmclarity/orchestrator/watcher/scanconfig"
	scanestimationwatcher "github.com/openclarity/vmclarity/orchestrator/watcher/scanestimation"
	"github.com/openclarity/vmclarity/provider"
)

type Orchestrator struct {
	controllers []Controller
	cancelFunc  context.CancelFunc

	controllerStartupDelay time.Duration
}

func Run(ctx context.Context, config *Config) error {
	healthServer := healthz.NewHealthServer(config.HealthCheckAddress)
	healthServer.Start()
	healthServer.SetIsReady(false)

	orchestrator, err := New(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to initialize Orchestrator: %w", err)
	}

	orchestrator.Start(ctx)
	healthServer.SetIsReady(true)

	return nil
}

// NewWithProvider returns an Orchestrator initialized using the p provider.Provider.
// Use this method when Orchestrator needs to rely on custom provider.Provider implementation.
// E.g. End-to-End testing.
func NewWithProvider(config *Config, p provider.Provider, c *apiclient.Client) (*Orchestrator, error) {
	scanConfigWatcherConfig := config.ScanConfigWatcherConfig.WithBackendClient(c)
	discoveryConfig := config.DiscoveryConfig.WithBackendClient(c).WithProviderClient(p)
	scanWatcherConfig := config.ScanWatcherConfig.WithBackendClient(c).WithProviderClient(p)
	assetScanWatcherConfig := config.AssetScanWatcherConfig.WithBackendClient(c).WithProviderClient(p)
	assetScanProcessorConfig := config.AssetScanProcessorConfig.WithBackendClient(c)
	assetScanEstimationWatcherConfig := config.AssetScanEstimationWatcherConfig.WithBackendClient(c).WithProviderClient(p)
	scanEstimationWatcherConfig := config.ScanEstimationWatcherConfig.WithBackendClient(c).WithProviderClient(p)

	return &Orchestrator{
		controllers: []Controller{
			scanconfigwatcher.New(scanConfigWatcherConfig),
			discovery.New(discoveryConfig),
			assetscanprocessor.New(assetScanProcessorConfig),
			scanwatcher.New(scanWatcherConfig),
			assetscanwatcher.New(assetScanWatcherConfig),
			assetscanestimationwatcher.New(assetScanEstimationWatcherConfig),
			scanestimationwatcher.New(scanEstimationWatcherConfig),
		},
		controllerStartupDelay: config.ControllerStartupDelay,
	}, nil
}

// New returns a new Orchestrator initialized using the provided configuration.
func New(ctx context.Context, config *Config) (*Orchestrator, error) {
	client, err := apiclient.New(config.APIServerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create an API client: %w", err)
	}

	p, err := NewProvider(ctx, config.ProviderKind)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider. Provider=%s: %w", config.ProviderKind, err)
	}

	return NewWithProvider(config, p, client)
}

// Start makes the Orchestrator to start all Controller(s).
func (o *Orchestrator) Start(ctx context.Context) {
	log.GetLoggerFromContextOrDiscard(ctx).Info("Starting Orchestrator server")

	ctx, cancel := context.WithCancel(ctx)
	o.cancelFunc = cancel

	for _, controller := range o.controllers {
		controller.Start(ctx)
		time.Sleep(o.controllerStartupDelay)
	}
}

// Start makes the Orchestrator to stop all Controller(s).
func (o *Orchestrator) Stop(ctx context.Context) {
	log.GetLoggerFromContextOrDiscard(ctx).Info("Stopping Orchestrator server")

	if o.cancelFunc != nil {
		o.cancelFunc()
	}
}
