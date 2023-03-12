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

package scanner

import (
	"context"
	"fmt"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/api/models"
	_config "github.com/openclarity/vmclarity/runtime_scan/pkg/config"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/types"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
	"github.com/openclarity/vmclarity/shared/pkg/backendclient"
)

type Scanner struct {
	targetIDToScanData map[string]*scanData
	scanConfig         *models.ScanConfig
	killSignal         chan bool
	providerClient     provider.Client
	logFields          log.Fields
	backendClient      *backendclient.BackendClient
	scanID             string
	targetInstances    []*types.TargetInstance
	config             *_config.ScannerConfig

	sync.Mutex
}

type scanData struct {
	targetInstance *types.TargetInstance
	scanResultID   string
	success        bool // Needed for deletion policy in case we want to access the logs
	timeout        bool
	completed      bool
}

func CreateScanner(
	config *_config.ScannerConfig,
	providerClient provider.Client,
	backendClient *backendclient.BackendClient,
	scanConfig *models.ScanConfig,
	targetInstances []*types.TargetInstance,
	scanID string,
) *Scanner {
	return &Scanner{
		targetIDToScanData: nil,
		scanConfig:         scanConfig,
		killSignal:         make(chan bool),
		providerClient:     providerClient,
		logFields:          log.Fields{"scanner id": uuid.NewV4().String()},
		backendClient:      backendClient,
		scanID:             scanID,
		targetInstances:    targetInstances,
		config:             config,
		Mutex:              sync.Mutex{},
	}
}

// initScan Calculate properties of scan targets
// nolint:cyclop,unparam
func (s *Scanner) initScan(ctx context.Context) error {
	targetIDToScanData := make(map[string]*scanData)

	// Populate the target to scanData map and create ScanResult for each target.
	for _, targetInstance := range s.targetInstances {
		scanResultID, err := s.createInitTargetScanStatus(ctx, s.scanID, targetInstance.TargetID)
		if err != nil {
			log.Errorf("Failed to create an init scan result. instance id=%v, scan id=%v: %v", targetInstance.TargetID, s.scanID, err)
			continue
		}
		targetIDToScanData[targetInstance.TargetID] = &scanData{
			targetInstance: targetInstance,
			scanResultID:   scanResultID,
			success:        false,
			completed:      false,
			timeout:        false,
		}
	}

	s.targetIDToScanData = targetIDToScanData

	// Move scan to "In Progress" and update the summary.
	summary := createInitScanSummary()
	summary.JobsLeftToRun = utils.PointerTo[int](len(targetIDToScanData))
	scan := &models.Scan{
		State:   utils.PointerTo[models.ScanState](models.InProgress),
		Summary: summary,
	}
	err := s.backendClient.PatchScan(ctx, s.scanID, scan)
	if err != nil {
		return fmt.Errorf("failed to update scan: %v", err)
	}

	log.WithFields(s.logFields).Infof("Total %d unique targets to scan", len(targetIDToScanData))

	return nil
}

func createInitScanSummary() *models.ScanSummary {
	return &models.ScanSummary{
		JobsCompleted:          utils.PointerTo[int](0),
		JobsLeftToRun:          utils.PointerTo[int](0),
		TotalExploits:          utils.PointerTo[int](0),
		TotalMalware:           utils.PointerTo[int](0),
		TotalMisconfigurations: utils.PointerTo[int](0),
		TotalPackages:          utils.PointerTo[int](0),
		TotalRootkits:          utils.PointerTo[int](0),
		TotalSecrets:           utils.PointerTo[int](0),
		TotalVulnerabilities: &models.VulnerabilityScanSummary{
			TotalCriticalVulnerabilities:   utils.PointerTo[int](0),
			TotalHighVulnerabilities:       utils.PointerTo[int](0),
			TotalLowVulnerabilities:        utils.PointerTo[int](0),
			TotalMediumVulnerabilities:     utils.PointerTo[int](0),
			TotalNegligibleVulnerabilities: utils.PointerTo[int](0),
		},
	}
}

func (s *Scanner) Scan(ctx context.Context) error {
	s.Lock()
	defer s.Unlock()

	log.WithFields(s.logFields).Infof("Start scanning ID=%s", s.scanID)

	err := s.initScan(ctx)
	if err != nil {
		return fmt.Errorf("failed to init scan ID=%s: %v", s.scanID, err)
	}

	if len(s.targetIDToScanData) == 0 {
		log.WithFields(s.logFields).Info("Nothing to scan")
		t := time.Now()
		reason := models.ScanStateReasonNothingToScan
		scan := &models.Scan{
			EndTime:      &t,
			State:        utils.PointerTo[models.ScanState](models.Done),
			StateMessage: utils.StringPtr("Nothing to scan"),
			StateReason:  &reason,
		}
		err := s.backendClient.PatchScan(ctx, s.scanID, scan)
		if err != nil {
			return fmt.Errorf("failed to set end time of the scan ID=%s: %v", s.scanID, err)
		}
		return nil
	}

	go s.jobBatchManagement(ctx)

	return nil
}

func (s *Scanner) SetTargetScanStatusCompletionError(ctx context.Context, scanResultID, errMsg string) error {
	// Get the status and set the completion error
	status, err := s.backendClient.GetScanResultStatus(ctx, scanResultID)
	if err != nil {
		return fmt.Errorf("failed to get a target scan status: %v", err)
	}

	var errors []string
	if status.General.Errors != nil {
		errors = *status.General.Errors
	}
	errors = append(errors, errMsg)
	status.General.Errors = &errors
	done := models.DONE
	status.General.State = &done

	err = s.backendClient.PatchTargetScanStatus(ctx, scanResultID, status)
	if err != nil {
		return fmt.Errorf("failed to put target scan status: %v", err)
	}

	return nil
}

func (s *Scanner) Clear() {
	s.Lock()
	defer s.Unlock()

	log.WithFields(s.logFields).Infof("Clearing...")
	close(s.killSignal)
}
