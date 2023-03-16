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

package backendclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/openclarity/vmclarity/api/client"
	"github.com/openclarity/vmclarity/api/models"
	runtimeScanUtils "github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

type BackendClient struct {
	apiClient client.ClientWithResponsesInterface
}

func Create(serverAddress string) (*BackendClient, error) {
	apiClient, err := client.NewClientWithResponses(serverAddress)
	if err != nil {
		return nil, fmt.Errorf("unable to create VMClarity API client. serverAddress=%v: %w", serverAddress, err)
	}

	return &BackendClient{
		apiClient: apiClient,
	}, nil
}

// nolint:cyclop
func (b *BackendClient) GetScanResult(ctx context.Context, scanResultID string, params models.GetScanResultsScanResultIDParams) (models.TargetScanResult, error) {
	newGetExistingError := func(err error) error {
		return fmt.Errorf("failed to get existing scan result %v: %w", scanResultID, err)
	}

	var scanResults models.TargetScanResult
	resp, err := b.apiClient.GetScanResultsScanResultIDWithResponse(ctx, scanResultID, &params)
	if err != nil {
		return scanResults, newGetExistingError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return scanResults, newGetExistingError(fmt.Errorf("empty body"))
		}
		return *resp.JSON200, nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return scanResults, newGetExistingError(fmt.Errorf("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return scanResults, newGetExistingError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return scanResults, newGetExistingError(fmt.Errorf("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return scanResults, newGetExistingError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return scanResults, newGetExistingError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

// nolint:cyclop
func (b *BackendClient) PatchScanResult(ctx context.Context, scanResult models.TargetScanResult, scanResultID string) error {
	newUpdateScanResultError := func(err error) error {
		return fmt.Errorf("failed to update scan result %v: %w", scanResultID, err)
	}

	resp, err := b.apiClient.PatchScanResultsScanResultIDWithResponse(ctx, scanResultID, scanResult)
	if err != nil {
		return newUpdateScanResultError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return newUpdateScanResultError(fmt.Errorf("empty body"))
		}
		return nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return newUpdateScanResultError(fmt.Errorf("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return newUpdateScanResultError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return newUpdateScanResultError(fmt.Errorf("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return newUpdateScanResultError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return newUpdateScanResultError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func (b *BackendClient) PostScan(ctx context.Context, scan models.Scan) (*models.Scan, error) {
	resp, err := b.apiClient.PostScansWithResponse(ctx, scan)
	if err != nil {
		return nil, fmt.Errorf("failed to post a scan: %v", err)
	}
	switch resp.StatusCode() {
	case http.StatusCreated:
		if resp.JSON201 == nil {
			return nil, fmt.Errorf("failed to create a scan: empty body. status code=%v", http.StatusCreated)
		}
		return resp.JSON201, nil
	case http.StatusConflict:
		if resp.JSON409 == nil {
			return nil, fmt.Errorf("failed to create a scan: empty body. status code=%v", http.StatusConflict)
		}
		if resp.JSON409.Scan == nil {
			return nil, fmt.Errorf("failed to create a scan: no scan data. status code=%v", http.StatusConflict)
		}
		return nil, ScanConflictError{
			ConflictingScan: resp.JSON409.Scan,
			Message:         "conflict",
		}
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to post scan. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to post scan. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) PostScanResult(ctx context.Context, scanResult models.TargetScanResult) (*models.TargetScanResult, error) {
	resp, err := b.apiClient.PostScanResultsWithResponse(ctx, scanResult)
	if err != nil {
		return nil, fmt.Errorf("failed to post scan result: %v", err)
	}
	switch resp.StatusCode() {
	case http.StatusCreated:
		if resp.JSON201 == nil {
			return nil, fmt.Errorf("failed to create a scan result: empty body. status code=%v", http.StatusCreated)
		}
		return resp.JSON201, nil
	case http.StatusConflict:
		if resp.JSON409 == nil {
			return nil, fmt.Errorf("failed to create a scan result: empty body. status code=%v", http.StatusConflict)
		}
		if resp.JSON409.TargetScanResult == nil {
			return nil, fmt.Errorf("failed to create a scan result: no scan result data. status code=%v", http.StatusConflict)
		}
		return nil, ScanResultConflictError{
			ConflictingScanResult: resp.JSON409.TargetScanResult,
			Message:               "conflict",
		}
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to post scan result. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to post scan result. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) PatchScan(ctx context.Context, scanID models.ScanID, scan *models.Scan) error {
	resp, err := b.apiClient.PatchScansScanIDWithResponse(ctx, scanID, *scan)
	if err != nil {
		return fmt.Errorf("failed to patch a scan: %v", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return fmt.Errorf("failed to patch a scan: empty body")
		}
		return nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return fmt.Errorf("failed to patch scan. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return fmt.Errorf("failed to patch scan. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) GetScanResultSummary(ctx context.Context, scanResultID string) (*models.ScanFindingsSummary, error) {
	params := models.GetScanResultsScanResultIDParams{
		Select: runtimeScanUtils.StringPtr("summary"),
	}
	scanResult, err := b.GetScanResult(ctx, scanResultID, params)
	if err != nil {
		return nil, err
	}
	return scanResult.Summary, nil
}

func (b *BackendClient) GetScanResultStatus(ctx context.Context, scanResultID string) (*models.TargetScanStatus, error) {
	params := models.GetScanResultsScanResultIDParams{
		Select: utils.StringPtr("status"),
	}
	scanResult, err := b.GetScanResult(ctx, scanResultID, params)
	if err != nil {
		return nil, err
	}
	return scanResult.Status, nil
}

// nolint:cyclop
func (b *BackendClient) PatchTargetScanStatus(ctx context.Context, scanResultID string, status *models.TargetScanStatus) error {
	scanResult := models.TargetScanResult{
		Status: status,
	}
	resp, err := b.apiClient.PatchScanResultsScanResultIDWithResponse(ctx, scanResultID, scanResult)
	if err != nil {
		return fmt.Errorf("failed to patch a scan result status: %v", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return fmt.Errorf("failed to update a scan result status: empty body")
		}
		return nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return fmt.Errorf("failed to update a scan result status: empty body on not found")
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return fmt.Errorf("failed to update scan result status, not found: %v", resp.JSON404.Message)
		}
		return fmt.Errorf("failed to update scan result status, not found")
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return fmt.Errorf("failed to update scan result status. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return fmt.Errorf("failed to update scan result status. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) GetScan(ctx context.Context, scanID string, params models.GetScansScanIDParams) (*models.Scan, error) {
	resp, err := b.apiClient.GetScansScanIDWithResponse(ctx, scanID, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get a scan: %v", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("failed to get a scan: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get a scan status. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get a scan status. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) GetScanConfigs(ctx context.Context, params models.GetScanConfigsParams) (*models.ScanConfigs, error) {
	resp, err := b.apiClient.GetScanConfigsWithResponse(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get scan configs: %v", err)
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

// nolint:cyclop
func (b *BackendClient) PatchScanConfig(ctx context.Context, scanConfigID string, scanConfig *models.ScanConfig) error {
	newPatchScanConfigResultError := func(err error) error {
		return fmt.Errorf("failed to patch scan config %v: %w", scanConfigID, err)
	}

	resp, err := b.apiClient.PatchScanConfigsScanConfigIDWithResponse(ctx, scanConfigID, *scanConfig)
	if err != nil {
		return newPatchScanConfigResultError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return newPatchScanConfigResultError(fmt.Errorf("empty body"))
		}
		return nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return newPatchScanConfigResultError(fmt.Errorf("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return newPatchScanConfigResultError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return newPatchScanConfigResultError(fmt.Errorf("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return newPatchScanConfigResultError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return newPatchScanConfigResultError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func (b *BackendClient) GetScans(ctx context.Context, params models.GetScansParams) (*models.Scans, error) {
	resp, err := b.apiClient.GetScansWithResponse(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get scans: %v", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("no scans: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get scans. status code=%v: %s", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get scans. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) PostTarget(ctx context.Context, target models.Target) (*models.Target, error) {
	resp, err := b.apiClient.PostTargetsWithResponse(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("failed to post target: %v", err)
	}
	switch resp.StatusCode() {
	case http.StatusCreated:
		if resp.JSON201 == nil {
			return nil, fmt.Errorf("failed to create a target: empty body. status code=%v", http.StatusCreated)
		}
		return resp.JSON201, nil
	case http.StatusConflict:
		if resp.JSON409 == nil {
			return nil, fmt.Errorf("failed to create a target: empty body. status code=%v", http.StatusConflict)
		}
		if resp.JSON409.Target == nil {
			return nil, fmt.Errorf("failed to create a target: no target data. status code=%v", http.StatusConflict)
		}
		return nil, TargetConflictError{
			ConflictingTarget: resp.JSON409.Target,
			Message:           "conflict",
		}
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to post target. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to post target. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) PutDiscoveryScopes(ctx context.Context, scope *models.Scopes) (*models.Scopes, error) {
	resp, err := b.apiClient.PutDiscoveryScopesWithResponse(ctx, *scope)
	if err != nil {
		return nil, fmt.Errorf("failed to put discovery scope: %v", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("failed to put scopes: empty body. status code=%v", http.StatusOK)
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to put scopes. status code=%v: %s", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to put scopes. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) GetTargets(ctx context.Context, params models.GetTargetsParams) (*models.Targets, error) {
	resp, err := b.apiClient.GetTargetsWithResponse(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get targets: %v", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("no targets: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get targets. status code=%v: %s", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get targets. status code=%v", resp.StatusCode())
	}
}
