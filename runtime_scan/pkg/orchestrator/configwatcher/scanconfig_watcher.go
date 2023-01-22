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

package configwatcher

import (
	"context"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/api/client"
	"github.com/openclarity/vmclarity/api/models"
	_config "github.com/openclarity/vmclarity/runtime_scan/pkg/config"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider"
)

const (
	timeWindow = 5 * time.Minute
)

type ScanConfigWatcher struct {
	backendClient  *client.ClientWithResponses
	providerClient provider.Client
	scannerConfig  *_config.ScannerConfig
}

func CreateScanConfigWatcher(
	backendClient *client.ClientWithResponses,
	providerClient provider.Client,
	scannerConfig _config.ScannerConfig,
) *ScanConfigWatcher {
	return &ScanConfigWatcher{
		backendClient:  backendClient,
		providerClient: providerClient,
		scannerConfig:  &scannerConfig,
	}
}

func (scw *ScanConfigWatcher) getScanConfigs() (*models.ScanConfigs, error) {
	resp, err := scw.backendClient.GetScanConfigsWithResponse(context.TODO(), &models.GetScanConfigsParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to get a scan configs: %v", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("no scan configs: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get scan configs. status code=%v: %s", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get scan configs. status code=%v", resp.StatusCode())
	}
}

func (scw *ScanConfigWatcher) hasRunningScansByScanConfigIDAndOperationTime(scanConfigID string, operationTime time.Time) (bool, error) {
	//odataFilter := fmt.Sprintf("scanConfigId eq '%s' and (endTime eq null or startTime gte '%s')", scanConfigID, operationTime.String())
	//params := &models.GetScansParams{
	//	Filter: &odataFilter,
	//}
	resp, err := scw.backendClient.GetScansWithResponse(context.TODO(), &models.GetScansParams{})
	if err != nil {
		return false, fmt.Errorf("failed to get a scans with: %v", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return false, fmt.Errorf("no scans: empty body")
		}
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return false, fmt.Errorf("failed to get scans. status code=%v: %s", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return false, fmt.Errorf("failed to get scans. status code=%v", resp.StatusCode())
	}
	// After Odata filters will be implemented on the backend the filter function can be removed
	return hasRunningOrCompletedScan(resp.JSON200, scanConfigID, operationTime), nil
}

func (scw *ScanConfigWatcher) getScanConfigsToScan() ([]models.ScanConfig, error) {
	scanConfigsToScan := make([]models.ScanConfig, 0)
	scanConfigs, err := scw.getScanConfigs()
	if err != nil {
		return nil, fmt.Errorf("failed to check new scan configs: %v", err)
	}

	now := time.Now()
	for _, scanConfig := range *scanConfigs.Items {
		// Check only the SingleScheduledScanConfigs at the moment
		shouldScan := false
		scheduled, err := scanConfig.Scheduled.ValueByDiscriminator()
		if err != nil {
			log.Errorf("failed to determine scheduled scan config type: %v", err)
			continue
		}
		switch singleSchedule := scheduled.(type) {
		case models.SingleScheduleScanConfig:
			shouldScan, err = scw.shouldStartSingleScheduleScanConfig(*scanConfig.Id, singleSchedule, now)
			if err != nil {
				log.Errorf("Failed to get scans for scan config ID=%s: %v", *scanConfig.Id, err)
				continue
			}
		default:
			continue
		}

		if shouldScan {
			scanConfigsToScan = append(scanConfigsToScan, scanConfig)
		}
	}
	return scanConfigsToScan, nil
}

func hasRunningOrCompletedScan(scans *models.Scans, scanConfigID string, operationTime time.Time) bool {
	if scans.Items == nil {
		return false
	}
	for _, scan := range *scans.Items {
		if *scan.ScanConfigId != scanConfigID {
			continue
		}
		if scan.EndTime == nil {
			// There is a running scan for this scanConfig
			return true
		}
		if scan.StartTime.After(operationTime) {
			// There is a completed scan that started after the operation time
			return true
		}
	}
	return false
}

func (scw *ScanConfigWatcher) shouldStartSingleScheduleScanConfig(scanConfigID string, schedule models.SingleScheduleScanConfig, now time.Time) (bool, error) {
	// Skip processing ScanConfig because its operationTime is outside of the start window
	if schedule.OperationTime.Sub(now).Abs() >= timeWindow {
		return false, nil
	}
	// Check running or completed scan for specific scan config
	hasRunningOrCompletedScan, err := scw.hasRunningScansByScanConfigIDAndOperationTime(scanConfigID, schedule.OperationTime)
	if err != nil {
		return false, fmt.Errorf("failed to get scans: %v", err)
	}
	return !hasRunningOrCompletedScan, nil
}

func (scw *ScanConfigWatcher) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-time.After(scw.scannerConfig.ScanConfigWatchInterval):
				// nolint:contextcheck
				scanConfigsToScan, err := scw.getScanConfigsToScan()
				if err != nil {
					log.Warnf("Failed to get scan configs to scan: %v", err)
				}
				if len(scanConfigsToScan) > 0 {
					scw.runNewScans(ctx, scanConfigsToScan)
				}
			case <-ctx.Done():
				log.Infof("Stop watching scan configs.")
				return
			}
		}
	}()
}
