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

	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/discovery"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/scanconfigwatcher"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/scanresultprocessor"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/scanresultwatcher"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/scanwatcher"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider"
	"github.com/openclarity/vmclarity/shared/pkg/backendclient"
	"github.com/openclarity/vmclarity/shared/pkg/log"
)

type Orchestrator struct {
	controllers []Controller
	cancelFunc  context.CancelFunc
}

func New(config *Config, p provider.Provider, b *backendclient.BackendClient) *Orchestrator {
	scanConfigWatcherConfig := config.ScanConfigWatcherConfig.WithBackendClient(b)
	discoveryConfig := config.DiscoveryConfig.WithBackendClient(b).WithProviderClient(p)
	scanWatcherConfig := config.ScanWatcherConfig.WithBackendClient(b).WithProviderClient(p)
	scanResultWatcherConfig := config.ScanResultWatcherConfig.WithBackendClient(b).WithProviderClient(p)
	scanResultProcessorConfig := config.ScanResultProcessorConfig.WithBackendClient(b)

	return &Orchestrator{
		controllers: []Controller{
			scanconfigwatcher.New(scanConfigWatcherConfig),
			discovery.New(discoveryConfig),
			scanresultprocessor.New(scanResultProcessorConfig),
			scanwatcher.New(scanWatcherConfig),
			scanresultwatcher.New(scanResultWatcherConfig),
		},
	}
}

func (o *Orchestrator) Start(ctx context.Context) {
	log.GetLoggerFromContextOrDiscard(ctx).Infof("Starting Orchestrator server")

	ctx, cancel := context.WithCancel(ctx)
	o.cancelFunc = cancel

	for _, controller := range o.controllers {
		controller.Start(ctx)
	}
}

func (o *Orchestrator) Stop(ctx context.Context) {
	log.GetLoggerFromContextOrDiscard(ctx).Infof("Stopping Orchestrator server")

	if o.cancelFunc != nil {
		o.cancelFunc()
	}
}
