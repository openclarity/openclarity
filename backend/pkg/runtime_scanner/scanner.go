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
}

type RuntimeScanner struct {
	vulnerabilitiesScanner orchestrator.VulnerabilitiesScanner
	// Scan config used in the latest runtime scan.
	lastScanConfig    *ScanConfig
	lastScanStartTime time.Time
	lastScanEndTime   time.Time
	// Channel for cancelling the currently running scan
	stopCurrentScanChan chan struct{}
	// List of latest scanned namespaces.
	scannedNamespaces []string
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
	MaxScanParallelism            int64
}

func CreateRuntimeScanner(scanner orchestrator.VulnerabilitiesScanner, scanChan chan *ScanConfig, resultsChan chan *types.ScanResults) Scanner {
	return &RuntimeScanner{
		vulnerabilitiesScanner: scanner,
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

// Abort and clean up any running scan.
func (s *RuntimeScanner) StopCurrentScan() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.stopCurrentScanChan != nil {
		close(s.stopCurrentScanChan)
		s.stopCurrentScanChan = nil
	}
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

func (s *RuntimeScanner) newStopScanChannel() (chan struct{}, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.stopCurrentScanChan != nil {
		return nil, fmt.Errorf("found existing cancel channel which may mean a scan is already running, call StopCurrentScan() before calling startScan again to ensure any existing scan is aborted and the cancel channel is cleaned up")
	}
	s.stopCurrentScanChan = make(chan struct{})
	return s.stopCurrentScanChan, nil
}

// Note: this is blocking until scan is done, or stop signal received.
func (s *RuntimeScanner) startScan(scanConfig *ScanConfig) error {
	startTime := time.Now().UTC()

	// Setup and save the cancel channel for this scan
	stop, err := s.newStopScanChannel()
	if err != nil {
		return err
	}
	// Defer closing and clearing up the stopCurrentScanChan
	defer s.StopCurrentScan()

	s.lock.Lock()
	s.lastScanStartTime = startTime
	s.scannedNamespaces = scanConfig.Namespaces
	s.lock.Unlock()

	namespaces := scanConfig.Namespaces

	// Save on state the config in the current runtime scan.
	s.setScanConfig(scanConfig)

	if len(namespaces) == 0 {
		// Empty namespaces list should scan all namespaces.
		namespaces = []string{corev1.NamespaceAll}
	}

	done, err := s.vulnerabilitiesScanner.Scan(&runtime_scan_config.ScanConfig{
		MaxScanParallelism:           scanConfig.MaxScanParallelism,
		TargetNamespaces:             namespaces,
		IgnoredNamespaces:            nil,
		JobResultTimeout:             10 * time.Minute, // nolint:gomnd
		DeleteJobPolicy:              runtime_scan_config.DeleteJobPolicySuccessful,
		ShouldScanCISDockerBenchmark: scanConfig.CisDockerBenchmarkScanEnabled,
	})
	if err != nil {
		return fmt.Errorf("failed to start scan: %v", err)
	}

	// Wait for scan to be complete or for the scan to be cancelled.
	select {
	case <-done:
	case <-stop:
		// Abort and reset the vulnerability scanner
		s.vulnerabilitiesScanner.Clear()
		<-done // Wait for abort to complete
	}

	// Get the results (whether complete or aborted) and return them to the
	// results handler via the resultsChan.
	s.lastScanEndTime = time.Now().UTC()
	results := s.vulnerabilitiesScanner.Results()
	select {
	case s.resultsChan <- results:
	default:
		log.Error("Failed to send results to channel")
	}

	return nil
}
