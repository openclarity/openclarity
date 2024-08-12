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

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/assetscanprocessor"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/assetscanwatcher"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/discovery"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/scanconfigwatcher"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/scanwatcher"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider/aws"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider/azure"
	"github.com/openclarity/vmclarity/shared/pkg/backendclient"
	"github.com/openclarity/vmclarity/shared/pkg/log"
)

type Orchestrator struct {
	controllers []Controller
	cancelFunc  context.CancelFunc

	controllerStartupDelay time.Duration
}

// NewWithProvider returns an Orchestrator initialized using the p provider.Provider.
// Use this method when Orchestrator needs to rely on custom provider.Provider implementation.
// E.g. End-to-End testing.
func NewWithProvider(config *Config, p provider.Provider, b *backendclient.BackendClient) (*Orchestrator, error) {
	scanConfigWatcherConfig := config.ScanConfigWatcherConfig.WithBackendClient(b)
	discoveryConfig := config.DiscoveryConfig.WithBackendClient(b).WithProviderClient(p)
	scanWatcherConfig := config.ScanWatcherConfig.WithBackendClient(b).WithProviderClient(p)
	assetScanWatcherConfig := config.AssetScanWatcherConfig.WithBackendClient(b).WithProviderClient(p)
	assetScanProcessorConfig := config.AssetScanProcessorConfig.WithBackendClient(b)

	return &Orchestrator{
		controllers: []Controller{
			scanconfigwatcher.New(scanConfigWatcherConfig),
			discovery.New(discoveryConfig),
			assetscanprocessor.New(assetScanProcessorConfig),
			scanwatcher.New(scanWatcherConfig),
			assetscanwatcher.New(assetScanWatcherConfig),
		},
		controllerStartupDelay: config.ControllerStartupDelay,
	}, nil
}

// New returns a new Orchestrator initialized using the provided configuration.
func New(ctx context.Context, config *Config, b *backendclient.BackendClient) (*Orchestrator, error) {
	p, err := NewProvider(ctx, config.ProviderKind)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider. Provider=%s: %w", config.ProviderKind, err)
	}

	return NewWithProvider(config, p, b)
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
// NewProvider returns an initialized provider.Provider based on the kind models.CloudProvider.
func NewProvider(ctx context.Context, kind models.CloudProvider) (provider.Provider, error) {
	switch kind {
	case models.Azure:
		return azure.New(ctx)
	case models.AWS:
		return aws.New(ctx)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", kind)
	}
}
