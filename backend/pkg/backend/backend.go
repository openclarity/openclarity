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

package backend

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	runtime_scan_config "github.com/openclarity/vmclarity/runtime_scan/pkg/config"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider/aws"

	"github.com/Portshift/go-utils/healthz"
	log "github.com/sirupsen/logrus"

	_config "github.com/openclarity/vmclarity/backend/pkg/config"
)

type Backend struct{}

func CreateBackend() *Backend {
	return &Backend{}
}

const defaultChanSize = 100

func Run() {
	config, err := _config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	errChan := make(chan struct{}, defaultChanSize)

	healthServer := healthz.NewHealthServer(config.HealthCheckAddress)
	healthServer.Start()

	healthServer.SetIsReady(false)

	globalCtx, globalCancel := context.WithCancel(context.Background())
	defer globalCancel()

	log.Info("VMClarity backend is running")

	runtimeScanConfig, err := runtime_scan_config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load runtime scan orchestrator config: %v", err)
	}
	providerClient, err := aws.Create(globalCtx, runtimeScanConfig.AWSConfig)
	if err != nil {
		log.Fatalf("Failed to create provider client: %v", err)
	}
	orc, err := createRuntimeScanOrchestrator(providerClient, runtimeScanConfig)
	if err != nil {
		log.Fatalf("Failed to create runtime scan orchestrator: %v", err)
	}
	orc.Start(errChan)

	healthServer.SetIsReady(true)
	log.Info("VMClarity backend is ready")

	// Wait for deactivation
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-errChan:
		globalCancel()
		log.Errorf("Received an error - shutting down")
	case s := <-sig:
		globalCancel()
		log.Warningf("Received a termination signal: %v", s)
	}
}

func createRuntimeScanOrchestrator(client provider.Client, config *runtime_scan_config.Config) (orchestrator.VulnerabilitiesScanner, error) {
	orc, err := orchestrator.Create(config, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create runtime scan orchestrator: %v", err)
	}

	return orc, nil
}
