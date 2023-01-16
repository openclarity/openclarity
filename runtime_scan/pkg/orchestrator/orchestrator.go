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

package orchestrator

import (
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	runtime_scan_models "github.com/openclarity/kubeclarity/runtime_scan/api/server/models"
	"github.com/openclarity/kubeclarity/runtime_scan/api/server/restapi/operations"
	_config "github.com/openclarity/kubeclarity/runtime_scan/pkg/config"
	"github.com/openclarity/kubeclarity/runtime_scan/pkg/rest"
	_scanner "github.com/openclarity/kubeclarity/runtime_scan/pkg/scanner"
	"github.com/openclarity/kubeclarity/runtime_scan/pkg/types"
)

type ImageContentAnalysisHandlerCallback func(*runtime_scan_models.ImageContentAnalysis) error

type Orchestrator struct {
	scanner                             *_scanner.Scanner
	config                              *_config.Config
	clientset                           kubernetes.Interface
	server                              *rest.Server
	imageContentAnalysisHandlerCallback ImageContentAnalysisHandlerCallback
	sync.Mutex
}

//go:generate $GOPATH/bin/mockgen -destination=./mock_orchestrator.go -package=orchestrator github.com/openclarity/kubeclarity/runtime_scan/pkg/orchestrator VulnerabilitiesScanner
type VulnerabilitiesScanner interface {
	Start(errChan chan struct{})
	Scan(scanConfig *_config.ScanConfig) (chan struct{}, error)
	ScanProgress() types.ScanProgress
	Results() *types.ScanResults
	Clear()
	Stop()
}

func Create(config *_config.Config, clientset kubernetes.Interface) (*Orchestrator, error) {
	orc := &Orchestrator{
		scanner:   _scanner.CreateScanner(config, clientset),
		config:    config,
		clientset: clientset,
		Mutex:     sync.Mutex{},
	}

	server, err := rest.CreateRESTServer(config.ScannerJobResultListenPort, orc)
	if err != nil {
		return nil, fmt.Errorf("failed to create runtime scan rest server: %v", err)
	}
	orc.server = server

	return orc, nil
}

func (o *Orchestrator) SetImageContentAnalysisHandlerCallback(callback ImageContentAnalysisHandlerCallback) {
	o.imageContentAnalysisHandlerCallback = callback
}

func (o *Orchestrator) Start(errChan chan struct{}) {
	// Start result server
	log.Infof("Starting Orchestrator server")
	o.server.Start(errChan)
}

func (o *Orchestrator) Stop() {
	o.Clear()

	log.Infof("Stopping Orchestrator server")
	o.server.Stop()
}

func (o *Orchestrator) Scan(scanConfig *_config.ScanConfig) (chan struct{}, error) {
	o.scanner = _scanner.CreateScanner(o.config, o.clientset)
	done, err := o.getScanner().Scan(scanConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to scan: %v", err)
	}
	return done, nil
}

func (o *Orchestrator) ScanProgress() types.ScanProgress {
	return o.getScanner().ScanProgress()
}

func (o *Orchestrator) Results() *types.ScanResults {
	return o.getScanner().Results()
}

func (o *Orchestrator) Clear() {
	o.Lock()
	defer o.Unlock()

	log.Infof("Clearing Orchestrator")
	o.scanner.Clear()
}

func (o *Orchestrator) getScanner() *_scanner.Scanner {
	o.Lock()
	defer o.Unlock()

	return o.scanner
}

func (o *Orchestrator) HandleScanResults(params operations.PostScanScanUUIDResultsParams) error {
	// nolint:wrapcheck
	return o.getScanner().HandleScanResults(params)
}

func (o *Orchestrator) HandleCISDockerBenchmarkScanResults(params operations.PostScanScanUUIDCisDockerBenchmarkResultsParams) error {
	// nolint:wrapcheck
	return o.getScanner().HandleCISDockerBenchmarkScanResult(params)
}

func (o *Orchestrator) HandleScanContentAnalysis(params operations.PostScanScanUUIDContentAnalysisParams) error {
	if o.imageContentAnalysisHandlerCallback == nil {
		log.Infof("Ignoring scan contents analysis report - no callback")
		return nil
	}

	if !o.getScanner().ShouldHandleScanContentAnalysis(params) {
		return nil
	}

	return o.imageContentAnalysisHandlerCallback(params.Body)
}
