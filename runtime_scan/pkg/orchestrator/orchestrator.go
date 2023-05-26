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

	_config "github.com/openclarity/vmclarity/runtime_scan/pkg/config"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/configwatcher"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/discovery"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/scanresultprocessor"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/scanwatcher"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider"
	"github.com/openclarity/vmclarity/shared/pkg/backendclient"
	"github.com/openclarity/vmclarity/shared/pkg/log"
)

type Orchestrator interface {
	Start(ctx context.Context)
	Stop(ctx context.Context)
}

type orchestrator struct {
	config              *_config.OrchestratorConfig
	scanConfigWatcher   *configwatcher.ScanConfigWatcher
	scopeDiscoverer     *discovery.ScopeDiscoverer
	scanResultProcessor *scanresultprocessor.ScanResultProcessor
	scanWatcher         *scanwatcher.Watcher
	cancelFunc          context.CancelFunc
}

func New(config *_config.OrchestratorConfig, providerClient provider.Client, backendClient *backendclient.BackendClient) (Orchestrator, error) {
	orc := &orchestrator{
		config:              config,
		scanConfigWatcher:   configwatcher.CreateScanConfigWatcher(backendClient, providerClient, config.ScannerConfig),
		scopeDiscoverer:     discovery.CreateScopeDiscoverer(backendClient, providerClient),
		scanResultProcessor: scanresultprocessor.NewScanResultProcessor(backendClient),
		scanWatcher: scanwatcher.New(scanwatcher.Config{
			Backend:          backendClient,
			PollPeriod:       scanwatcher.DefaultPollInterval,
			ReconcileTimeout: scanwatcher.DefaultReconcileTimeout,
		}),
	}

	return orc, nil
}

func (o *orchestrator) Start(ctx context.Context) {
	logger := log.GetLoggerFromContextOrDefault(ctx)

	logger.Infof("Starting Orchestrator server")
	ctx, cancel := context.WithCancel(ctx)
	o.cancelFunc = cancel
	o.scanConfigWatcher.Start(ctx)
	o.scopeDiscoverer.Start(ctx)
	o.scanResultProcessor.Start(ctx)
	o.scanWatcher.Start(ctx)
}

func (o *orchestrator) Stop(ctx context.Context) {
	logger := log.GetLoggerFromContextOrDefault(ctx)

	logger.Infof("Stopping Orchestrator server")
	if o.cancelFunc != nil {
		o.cancelFunc()
	}
}
