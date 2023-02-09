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

	"github.com/openclarity/vmclarity/api/models"
	_scanner "github.com/openclarity/vmclarity/runtime_scan/pkg/scanner"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/types"
)

func (scw *ScanConfigWatcher) runNewScans(ctx context.Context, scanConfigs []models.ScanConfig) {
	for _, sc := range scanConfigs {
		scanConfig := sc
		if err := scw.scan(ctx, &scanConfig); err != nil {
			log.Errorf("Failed to schedule a scan with scan config ID=%s: %v", *scanConfig.Id, err)
		}
	}
}

func (scw *ScanConfigWatcher) scan(ctx context.Context, scanConfig *models.ScanConfig) error {
	// TODO: check if existing scan or a new scan
	targetInstances, scanID, err := scw.initNewScan(ctx, scanConfig)
	if err != nil {
		return fmt.Errorf("failed to init new scan: %v", err)
	}

	scanner := _scanner.CreateScanner(scw.scannerConfig, scw.providerClient, scw.backendClient, scanConfig, targetInstances, scanID)
	if err := scanner.Scan(ctx); err != nil {
		return fmt.Errorf("failed to scan: %v", err)
	}

	return nil
}

// initNewScan Initialized a new scan, returns target instances and scan ID.
func (scw *ScanConfigWatcher) initNewScan(ctx context.Context, scanConfig *models.ScanConfig) ([]*types.TargetInstance, string, error) {
	instances, err := scw.providerClient.Discover(ctx, scanConfig.Scope)
	if err != nil {
		return nil, "", fmt.Errorf("failed to discover instances to scan: %v", err)
	}

	targetInstances, err := scw.createTargetInstances(ctx, instances)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get or create targets: %v", err)
	}

	now := time.Now().UTC()
	scan := &models.Scan{
		ScanConfigId:       scanConfig.Id,
		ScanFamiliesConfig: scanConfig.ScanFamiliesConfig,
		StartTime:          &now,
		TargetIDs:          getTargetIDs(targetInstances),
	}
	scanID, err := scw.createScan(ctx, scan)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get or create a scan: %v", err)
	}

	return targetInstances, scanID, nil
}

func getTargetIDs(targetInstances []*types.TargetInstance) *[]string {
	ret := make([]string, len(targetInstances))
	for i, targetInstance := range targetInstances {
		ret[i] = targetInstance.TargetID
	}

	return &ret
}

func (scw *ScanConfigWatcher) createTargetInstances(ctx context.Context, instances []types.Instance) ([]*types.TargetInstance, error) {
	targetInstances := make([]*types.TargetInstance, 0, len(instances))
	for i, instance := range instances {
		target, err := scw.createTarget(ctx, instance)
		if err != nil {
			return nil, fmt.Errorf("failed to create target. instanceID=%v: %v", instance.GetID(), err)
		}
		targetInstances = append(targetInstances, &types.TargetInstance{
			TargetID: *target.Id,
			Instance: instances[i],
		})
	}

	return targetInstances, nil
}

func (scw *ScanConfigWatcher) createTarget(ctx context.Context, instance types.Instance) (*models.Target, error) {
	info := models.TargetType{}
	instanceProvider := models.AWS
	err := info.FromVMInfo(models.VMInfo{
		InstanceID:       instance.GetID(),
		InstanceProvider: &instanceProvider,
		Location:         instance.GetLocation(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create VMInfo: %v", err)
	}
	resp, err := scw.backendClient.PostTargetsWithResponse(ctx, models.Target{
		TargetInfo: &info,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to post target: %v", err)
	}
	switch resp.StatusCode() {
	case http.StatusCreated:
		if resp.JSON201 == nil {
			return nil, fmt.Errorf("failed to create a target: empty body")
		}
		return resp.JSON201, nil
	case http.StatusConflict:
		if resp.JSON409 == nil {
			return nil, fmt.Errorf("failed to create a target: empty body on conflict")
		}
		return resp.JSON409.Target, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to post target. status code=%v: %v", resp.StatusCode(), resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to post target. status code=%v", resp.StatusCode())
	}
}

// nolint:cyclop
func (scw *ScanConfigWatcher) createScan(ctx context.Context, scan *models.Scan) (string, error) {
	resp, err := scw.backendClient.PostScansWithResponse(ctx, *scan)
	if err != nil {
		return "", fmt.Errorf("failed to post a scan: %v", err)
	}
	switch resp.StatusCode() {
	case http.StatusCreated:
		if resp.JSON201 == nil {
			return "", fmt.Errorf("failed to create a scan: empty body")
		}
		if resp.JSON201 == nil {
			return "", fmt.Errorf("scan id is nil")
		}
		return *resp.JSON201.Id, nil
	case http.StatusConflict:
		if resp.JSON409 == nil {
			return "", fmt.Errorf("failed to create a scan: empty body on conflict")
		}
		if resp.JSON409.Scan.Id == nil {
			return "", fmt.Errorf("scan id on conflict is nil")
		}
		return *resp.JSON409.Scan.Id, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return "", fmt.Errorf("failed to post scan. status code=%v: %v", resp.StatusCode(), resp.JSONDefault.Message)
		}
		return "", fmt.Errorf("failed to post scan. status code=%v", resp.StatusCode())
	}
}
