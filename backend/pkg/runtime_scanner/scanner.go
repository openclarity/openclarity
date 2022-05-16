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

package runtimescanner

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"

	"github.com/openclarity/kubeclarity/api/server/models"
	runtime_scan_config "github.com/openclarity/kubeclarity/runtime_scan/pkg/config"
	"github.com/openclarity/kubeclarity/runtime_scan/pkg/orchestrator"
	"github.com/openclarity/kubeclarity/runtime_scan/pkg/types"
)

type Scanner interface {
	GetScanConfig() *ScanConfig
	ScanProgress() types.ScanProgress
	GetScannedNamespaces() []string
	GetLastScanStartTime() time.Time
	GetLastScanEndTime() time.Time
	Start(stopChan chan struct{})
	StopCurrentScan()
	Clear()
}

type RuntimeScanner struct {
	vulnerabilitiesScanner orchestrator.VulnerabilitiesScanner
	// Scan config used in the latest runtime scan.
	lastScanConfig    *ScanConfig
	lastScanStartTime time.Time
	lastScanEndTime   time.Time
	// List of latest scanned namespaces.
	scannedNamespaces   []string
	stopCurrentScanChan chan struct{}
	// Scan results will be sent through this channel.
	resultsChan chan *types.ScanResults
	// New scan requests are coming through this channel.
	scanChan chan *ScanConfig
	lock     sync.RWMutex
}

func (s *RuntimeScanner) ScanProgress() types.ScanProgress {
	return s.vulnerabilitiesScanner.ScanProgress()
}

type ScanConfig struct {
	ScanType                      models.ScanType
	CisDockerBenchmarkScanEnabled bool
	Namespaces                    []string
}

func CreateRuntimeScanner(scanner orchestrator.VulnerabilitiesScanner, scanChan chan *ScanConfig, resultsChan chan *types.ScanResults) Scanner {
	return &RuntimeScanner{
		vulnerabilitiesScanner: scanner,
		stopCurrentScanChan:    make(chan struct{}),
		scanChan:               scanChan,
		lock:                   sync.RWMutex{},
		resultsChan:            resultsChan,
		lastScanConfig:         &ScanConfig{},
	}
}

func (s *RuntimeScanner) GetLastScanStartTime() time.Time {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.lastScanStartTime
}

func (s *RuntimeScanner) GetLastScanEndTime() time.Time {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.lastScanEndTime
}

func (s *RuntimeScanner) GetScannedNamespaces() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.scannedNamespaces
}

func (s *RuntimeScanner) Clear() {
	s.vulnerabilitiesScanner.Clear()
}

func (s *RuntimeScanner) Start(stopChan chan struct{}) {
	s.vulnerabilitiesScanner.Start(stopChan)
	go func() {
		for {
			select {
			case <-stopChan:
				log.Info("Runtime scanner received stop event")
				return
			case scanConfig := <-s.scanChan:
				if err := s.startScan(scanConfig); err != nil {
					log.Errorf("Failed to start scan: %v", err)
					continue
				}
			}
		}
	}()
}

func (s *RuntimeScanner) GetScanConfig() *ScanConfig {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.lastScanConfig
}

func (s *RuntimeScanner) setScanConfig(scanConfig *ScanConfig) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.lastScanConfig = scanConfig
}

// Note: this is blocking until scan is done, or stop signal received.
func (s *RuntimeScanner) startScan(scanConfig *ScanConfig) error {
	startTime := time.Now().UTC()
	s.lock.Lock()
	s.lastScanStartTime = startTime
	stop := s.stopCurrentScanChan
	s.scannedNamespaces = scanConfig.Namespaces
	s.lock.Unlock()

	namespaces := scanConfig.Namespaces

	s.vulnerabilitiesScanner.Clear()

	// Save on state the config in the current runtime scan.
	s.setScanConfig(scanConfig)

	// need to create scan done channel for every new scan
	done := make(chan struct{})

	if len(namespaces) == 0 {
		// Empty namespaces list should scan all namespaces.
		namespaces = []string{corev1.NamespaceAll}
	}

	err := s.vulnerabilitiesScanner.Scan(&runtime_scan_config.ScanConfig{
		MaxScanParallelism:           10, // nolint:gomnd
		TargetNamespaces:             namespaces,
		IgnoredNamespaces:            nil,
		JobResultTimeout:             10 * time.Minute, // nolint:gomnd
		DeleteJobPolicy:              runtime_scan_config.DeleteJobPolicySuccessful,
		ShouldScanCISDockerBenchmark: scanConfig.CisDockerBenchmarkScanEnabled,
	}, done)
	if err != nil {
		return fmt.Errorf("failed to start scan: %v", err)
	}

	select {
	case <-done:
		s.lastScanEndTime = time.Now().UTC()
		results := s.vulnerabilitiesScanner.Results()
		select {
		case s.resultsChan <- results:
		default:
			log.Error("Failed to send results to channel")
		}
	case <-stop:
		log.Infof("Received a stop signal, not waiting for results")
	}

	return nil
}

func (s *RuntimeScanner) StopCurrentScan() {
	s.lock.Lock()
	close(s.stopCurrentScanChan)
	s.stopCurrentScanChan = make(chan struct{})
	s.scannedNamespaces = []string{}
	s.lock.Unlock()
}
