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
	"context"
	"fmt"
	"net/http"
	"sync"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/api/client"
	"github.com/openclarity/vmclarity/api/models"
	_config "github.com/openclarity/vmclarity/runtime_scan/pkg/config"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/types"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
)

type Scanner struct {
	targetIDToScanData map[string]*scanData
	scanConfig         *models.ScanConfig
	killSignal         chan bool
	providerClient     provider.Client
	logFields          log.Fields
	backendClient      *client.ClientWithResponses
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
	config *_config.OrchestratorConfig,
	providerClient provider.Client,
	backendClient *client.ClientWithResponses,
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
		config:             &config.ScannerConfig,
		Mutex:              sync.Mutex{},
	}
}

// initScan Calculate properties of scan targets
// nolint:cyclop,unparam
func (s *Scanner) initScan(ctx context.Context) error {
	targetIDToScanData := make(map[string]*scanData)

	// Populate the target to scanData map
	for _, targetInstance := range s.targetInstances {
		scanResultID, err := s.createInitTargetScanStatus(ctx, s.scanID, targetInstance.TargetID)
		if err != nil {
			log.Errorf("Failed to create an init scan result. instance id=%v, scan id=%v: %v", targetInstance.TargetID, s.scanConfig, err)
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

	log.WithFields(s.logFields).Infof("Total %d unique targets to scan", len(targetIDToScanData))

	return nil
}

func (s *Scanner) Scan(ctx context.Context, scanDone chan struct{}) error {
	s.Lock()
	defer s.Unlock()

	log.WithFields(s.logFields).Infof("Start scanning...")

	err := s.initScan(ctx)
	if err != nil {
		return fmt.Errorf("failed to init scan: %v", err)
	}

	if len(s.targetIDToScanData) == 0 {
		log.WithFields(s.logFields).Info("Nothing to scan")
		nonBlockingNotification(scanDone)
		return nil
	}

	go s.jobBatchManagement(ctx, scanDone)

	return nil
}

func (s *Scanner) GetTargetScanStatus(ctx context.Context, scanResultID string) (*models.TargetScanStatus, error) {
	params := &models.GetScanResultsScanResultIDParams{
		Select: utils.StringPtr("status"),
	}
	resp, err := s.backendClient.GetScanResultsScanResultIDWithResponse(ctx, scanResultID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get a target scan status: %v", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("failed to get a target scan status: empty body")
		}
		if resp.JSON200.Status == nil {
			return nil, fmt.Errorf("failed to get a target scan status: empty status in body")
		}
		return resp.JSON200.Status, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get a target scan status. status code=%v: %v", resp.StatusCode(), resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get a target scan status. status code=%v", resp.StatusCode())
	}
}

func (s *Scanner) SetTargetScanStatusCompletionError(ctx context.Context, scanResultID, errMsg string) error {
	// Get the status and set the completion error
	status, err := s.GetTargetScanStatus(ctx, scanResultID)
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

	err = s.patchTargetScanStatus(ctx, scanResultID, status)
	if err != nil {
		return fmt.Errorf("failed to put target scan status: %v", err)
	}

	return nil
}

// nolint:cyclop
func (s *Scanner) patchTargetScanStatus(ctx context.Context, scanResultID string, status *models.TargetScanStatus) error {
	scanResult := models.TargetScanResult{
		Status: status,
	}
	resp, err := s.backendClient.PatchScanResultsScanResultIDWithResponse(ctx, scanResultID, scanResult)
	if err != nil {
		return fmt.Errorf("failed to put a scan status: %v", err)
	}
	switch resp.StatusCode() {
	case http.StatusCreated:
		if resp.JSON200 == nil {
			return fmt.Errorf("failed to update a scan status: empty body")
		}
		return nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return fmt.Errorf("failed to update a scan status: empty body on not found")
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return fmt.Errorf("failed to update scan status, not found: %v", resp.JSON404.Message)
		}
		return fmt.Errorf("failed to update scan status, not found")
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return fmt.Errorf("failed to update scan status. status code=%v: %v", resp.StatusCode(), resp.JSONDefault.Message)
		}
		return fmt.Errorf("failed to update scan status. status code=%v", resp.StatusCode())
	}
}

func (s *Scanner) Clear() {
	s.Lock()
	defer s.Unlock()

	log.WithFields(s.logFields).Infof("Clearing...")
	close(s.killSignal)
}
