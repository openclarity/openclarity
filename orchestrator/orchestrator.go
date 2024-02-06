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

	"github.com/openclarity/vmclarity/api/client"
	"github.com/openclarity/vmclarity/api/types"
	discovery "github.com/openclarity/vmclarity/orchestrator/discoverer"
	assetscanprocessor "github.com/openclarity/vmclarity/orchestrator/processor/assetscan"
	assetscanwatcher "github.com/openclarity/vmclarity/orchestrator/watcher/assetscan"
	assetscanestimationwatcher "github.com/openclarity/vmclarity/orchestrator/watcher/assetscanestimation"
	scanwatcher "github.com/openclarity/vmclarity/orchestrator/watcher/scan"
	scanconfigwatcher "github.com/openclarity/vmclarity/orchestrator/watcher/scanconfig"
	scanestimationwatcher "github.com/openclarity/vmclarity/orchestrator/watcher/scanestimation"
	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/provider/aws"
	"github.com/openclarity/vmclarity/provider/azure"
	"github.com/openclarity/vmclarity/provider/docker"
	"github.com/openclarity/vmclarity/provider/external"
	"github.com/openclarity/vmclarity/provider/gcp"
	"github.com/openclarity/vmclarity/provider/kubernetes"
	"github.com/openclarity/vmclarity/utils/log"
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
func NewWithProvider(config *Config, p provider.Provider, b *client.BackendClient) (*Orchestrator, error) {
	scanConfigWatcherConfig := config.ScanConfigWatcherConfig.WithBackendClient(b)
	discoveryConfig := config.DiscoveryConfig.WithBackendClient(b).WithProviderClient(p)
	scanWatcherConfig := config.ScanWatcherConfig.WithBackendClient(b).WithProviderClient(p)
	assetScanWatcherConfig := config.AssetScanWatcherConfig.WithBackendClient(b).WithProviderClient(p)
	assetScanProcessorConfig := config.AssetScanProcessorConfig.WithBackendClient(b)
	assetScanEstimationWatcherConfig := config.AssetScanEstimationWatcherConfig.WithBackendClient(b).WithProviderClient(p)
	scanEstimationWatcherConfig := config.ScanEstimationWatcherConfig.WithBackendClient(b).WithProviderClient(p)

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
	backendClient, err := client.Create(config.APIServerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create a backend client: %w", err)
	}

	p, err := NewProvider(ctx, config.ProviderKind)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider. Provider=%s: %w", config.ProviderKind, err)
	}

	return NewWithProvider(config, p, backendClient)
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

// nolint:wrapcheck
// NewProvider returns an initialized provider.Provider based on the kind types.CloudProvider.
func NewProvider(ctx context.Context, kind types.CloudProvider) (provider.Provider, error) {
	switch kind {
	case types.Azure:
		return azure.New(ctx)
	case types.Docker:
		return docker.New(ctx)
	case types.AWS:
		return aws.New(ctx)
	case types.GCP:
		return gcp.New(ctx)
	case types.External:
		return external.New(ctx)
	case types.Kubernetes:
		return kubernetes.New(ctx)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", kind)
	}
}
