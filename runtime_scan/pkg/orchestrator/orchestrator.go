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
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/api/client"
	"github.com/openclarity/vmclarity/api/models"
	_config "github.com/openclarity/vmclarity/runtime_scan/pkg/config"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider"
	_scanner "github.com/openclarity/vmclarity/runtime_scan/pkg/scanner"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/types"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
)

type Orchestrator struct {
	config         *_config.OrchestratorConfig
	providerClient provider.Client
	backendClient  *client.ClientWithResponses
	// server *rest.Server
	sync.Mutex
}

//go:generate $GOPATH/bin/mockgen -destination=./mock_orchestrator.go -package=orchestrator github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator VulnerabilitiesScanner
type VulnerabilitiesScanner interface {
	Start(errChan chan struct{})
	Scan(ctx context.Context, scanConfig *models.ScanConfig, scanDone chan struct{}) error
}

func Create(config *_config.OrchestratorConfig, providerClient provider.Client) (*Orchestrator, error) {
	backendClient, err := client.NewClientWithResponses(
		fmt.Sprintf("%s:%d/%s", config.BackendAddress, config.BackendRestPort, config.BackendBaseURL),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create a backend client: %v", err)
	}
	orc := &Orchestrator{
		config:         config,
		providerClient: providerClient,
		Mutex:          sync.Mutex{},
		backendClient:  backendClient,
	}

	return orc, nil
}

func (o *Orchestrator) Start(errChan chan struct{}) {
	// Start orchestrator server

	// TODO: watch scan configs and call Scan() when needed
	log.Infof("Starting Orchestrator server")
}

func (o *Orchestrator) Scan(ctx context.Context, scanConfig *models.ScanConfig, scanDone chan struct{}) error {
	// TODO: check if existing scan or a new scan
	targetInstances, scanID, err := o.initNewScan(ctx, scanConfig)
	if err != nil {
		return fmt.Errorf("failed to init new scan: %v", err)
	}

	scanner := _scanner.CreateScanner(o.config, o.providerClient, o.backendClient, scanConfig, targetInstances, scanID)

	if err := scanner.Scan(ctx, scanDone); err != nil {
		return fmt.Errorf("failed to scan: %v", err)
	}

	return nil
}

// initNewScan Initialized a new scan, returns target instances and scan ID.
func (o *Orchestrator) initNewScan(ctx context.Context, scanConfig *models.ScanConfig) ([]*types.TargetInstance, string, error) {
	instances, err := o.providerClient.Discover(ctx, scanConfig.Scope)
	if err != nil {
		return nil, "", fmt.Errorf("failed to discover instances to scan: %v", err)
	}

	targetInstances, err := o.createTargetInstances(ctx, instances)
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
	scanID, err := o.getOrCreateScan(ctx, scan)
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

func (o *Orchestrator) createTargetInstances(ctx context.Context, instances []types.Instance) ([]*types.TargetInstance, error) {
	targetInstances := make([]*types.TargetInstance, 0, len(instances))
	for i, instance := range instances {
		target, err := o.getOrCreateTarget(ctx, instance)
		if err != nil {
			return nil, fmt.Errorf("failed to get or create target. instanceID=%v: %v", instance.GetID(), err)
		}
		targetInstances = append(targetInstances, &types.TargetInstance{
			TargetID: *target.Id,
			Instance: instances[i],
		})
	}

	return targetInstances, nil
}

func (o *Orchestrator) getOrCreateTarget(ctx context.Context, instance types.Instance) (*models.Target, error) {
	info := models.TargetType{}
	instanceProvider := models.AWS
	err := info.FromVMInfo(models.VMInfo{
		InstanceID:       utils.StringPtr(instance.GetID()),
		InstanceProvider: &instanceProvider,
		Location:         utils.StringPtr(instance.GetLocation()),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create VMInfo: %v", err)
	}
	resp, err := o.backendClient.PostTargetsWithResponse(ctx, models.Target{
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
		return resp.JSON409, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to post target. status code=%v: %v", resp.StatusCode(), resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to post target. status code=%v", resp.StatusCode())
	}
}

// nolint:cyclop
func (o *Orchestrator) getOrCreateScan(ctx context.Context, scan *models.Scan) (string, error) {
	resp, err := o.backendClient.PostScansWithResponse(ctx, *scan)
	if err != nil {
		return "", fmt.Errorf("failed to post a scan: %v", err)
	}
	switch resp.StatusCode() {
	case http.StatusCreated:
		if resp.JSON201 == nil {
			return "", fmt.Errorf("failed to create a scan: empty body")
		}
		if resp.JSON201.Id == nil {
			return "", fmt.Errorf("scan id is nil")
		}
		return *resp.JSON201.Id, nil
	case http.StatusConflict:
		if resp.JSON409 == nil {
			return "", fmt.Errorf("failed to create a scan: empty body on conflict")
		}
		if resp.JSON409.Id == nil {
			return "", fmt.Errorf("scan id on conflict is nil")
		}
		return *resp.JSON409.Id, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return "", fmt.Errorf("failed to post scan. status code=%v: %v", resp.StatusCode(), resp.JSONDefault.Message)
		}
		return "", fmt.Errorf("failed to post scan. status code=%v", resp.StatusCode())
	}
}
