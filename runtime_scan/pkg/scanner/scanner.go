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

package scanner

import (
	"fmt"
	"sync"
	"sync/atomic"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	_config "github.com/openclarity/vmclarity/runtime_scan/pkg/config"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/types"
)

type Scanner struct {
	instanceIDToScanData map[string]*scanData
	progress             types.ScanProgress
	scanConfig           *_config.ScanConfig
	killSignal           chan bool
	providerClient       provider.Client
	logFields            log.Fields

	region string

	sync.Mutex
}

type scanData struct {
	instance              types.Instance
	scanUUID              string
	vulnerabilitiesResult vulnerabilitiesScanResult
	resultChan            chan bool
	success               bool
	completed             bool
	timeout               bool
	scanErr               *types.ScanError
}

type vulnerabilitiesScanResult struct {
	result []string
	// success   bool
	// completed bool
	// error *scanner.Error
}

func CreateScanner(config *_config.Config, providerClient provider.Client) *Scanner {
	s := &Scanner{
		progress: types.ScanProgress{
			Status: types.Idle,
		},
		killSignal:     make(chan bool),
		providerClient: providerClient,
		logFields:      log.Fields{"scanner id": uuid.NewV4().String()},
		region:         config.Region,
		Mutex:          sync.Mutex{},
	}

	return s
}

// initScan Calculate properties of scan targets
// nolint:cyclop,unparam
func (s *Scanner) initScan() error {
	instanceIDToScanData := make(map[string]*scanData)

	// Populate the instance to scanData map
	for _, instance := range s.scanConfig.Instances {
		instanceIDToScanData[instance.GetID()] = &scanData{
			instance:              instance,
			scanUUID:              uuid.NewV4().String(),
			vulnerabilitiesResult: vulnerabilitiesScanResult{},
			resultChan:            make(chan bool),
			success:               false,
			completed:             false,
			timeout:               false,
			scanErr:               nil,
		}
	}

	s.instanceIDToScanData = instanceIDToScanData
	s.progress.InstancesToScan = uint32(len(instanceIDToScanData))

	log.WithFields(s.logFields).Infof("Total %d unique instances to scan", s.progress.InstancesToScan)

	return nil
}

func (s *Scanner) Scan(scanConfig *_config.ScanConfig, scanDone chan struct{}) error {
	s.Lock()
	defer s.Unlock()

	s.scanConfig = scanConfig

	log.WithFields(s.logFields).Infof("Start scanning...")

	s.progress.Status = types.ScanInit

	if err := s.initScan(); err != nil {
		s.progress.SetStatus(types.ScanInitFailure)
		return fmt.Errorf("failed to initiate scan: %v", err)
	}

	if s.progress.InstancesToScan == 0 {
		log.WithFields(s.logFields).Info("Nothing to scan")
		s.progress.SetStatus(types.NothingToScan)
		nonBlockingNotification(scanDone)
		return nil
	}

	s.progress.SetStatus(types.Scanning)
	go func() {
		s.jobBatchManagement(scanDone)

		s.Lock()
		s.progress.SetStatus(types.DoneScanning)
		s.Unlock()
	}()

	return nil
}

func (s *Scanner) ScanProgress() types.ScanProgress {
	return types.ScanProgress{
		InstancesToScan:          s.progress.InstancesToScan,
		InstancesStartedToScan:   atomic.LoadUint32(&s.progress.InstancesStartedToScan),
		InstancesCompletedToScan: atomic.LoadUint32(&s.progress.InstancesCompletedToScan),
		Status:                   s.progress.Status,
	}
}

func (s *Scanner) Results() *types.ScanResults {
	s.Lock()
	defer s.Unlock()

	instanceScanResults := make([]*types.InstanceScanResult, 0)

	for _, scanD := range s.instanceIDToScanData {
		if !scanD.completed {
			continue
		}
		instanceScanResults = append(instanceScanResults, &types.InstanceScanResult{
			Instance:        scanD.instance,
			Vulnerabilities: scanD.vulnerabilitiesResult.result,
			Success:         scanD.success,
		})
	}

	return &types.ScanResults{
		InstanceScanResults: instanceScanResults,
		Progress:            s.ScanProgress(),
	}
}

func (s *Scanner) Clear() {
	s.Lock()
	defer s.Unlock()

	log.WithFields(s.logFields).Infof("Clearing...")
	close(s.killSignal)
}
