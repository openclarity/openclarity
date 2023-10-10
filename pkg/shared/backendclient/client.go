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

// nolint:cyclop
package backendclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/openclarity/vmclarity/api/client"
	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
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

func (b *BackendClient) GetAssetScan(ctx context.Context, assetScanID string, params models.GetAssetScansAssetScanIDParams) (models.AssetScan, error) {
	newGetExistingError := func(err error) error {
		return fmt.Errorf("failed to get existing asset scan %v: %w", assetScanID, err)
	}

	var assetScans models.AssetScan
	resp, err := b.apiClient.GetAssetScansAssetScanIDWithResponse(ctx, assetScanID, &params)
	if err != nil {
		return assetScans, newGetExistingError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return assetScans, newGetExistingError(fmt.Errorf("empty body"))
		}
		return *resp.JSON200, nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return assetScans, newGetExistingError(fmt.Errorf("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return assetScans, newGetExistingError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return assetScans, newGetExistingError(fmt.Errorf("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return assetScans, newGetExistingError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return assetScans, newGetExistingError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func (b *BackendClient) GetAssetScans(ctx context.Context, params models.GetAssetScansParams) (models.AssetScans, error) {
	newGetAssetScansError := func(err error) error {
		return fmt.Errorf("failed to get asset scans: %w", err)
	}

	var assetScans models.AssetScans
	resp, err := b.apiClient.GetAssetScansWithResponse(ctx, &params)
	if err != nil {
		return assetScans, newGetAssetScansError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return assetScans, newGetAssetScansError(fmt.Errorf("empty body"))
		}
		return *resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return assetScans, newGetAssetScansError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return assetScans, newGetAssetScansError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func (b *BackendClient) PatchAssetScan(ctx context.Context, assetScan models.AssetScan, assetScanID string) error {
	newUpdateAssetScanError := func(err error) error {
		return fmt.Errorf("failed to update asset scan %v: %w", assetScanID, err)
	}

	params := models.PatchAssetScansAssetScanIDParams{}
	resp, err := b.apiClient.PatchAssetScansAssetScanIDWithResponse(ctx, assetScanID, &params, assetScan)
	if err != nil {
		return newUpdateAssetScanError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return newUpdateAssetScanError(fmt.Errorf("empty body"))
		}
		return nil
	case http.StatusBadRequest:
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			return newUpdateAssetScanError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSON400.Message))
		}
		return newUpdateAssetScanError(fmt.Errorf("status code=%v", resp.StatusCode()))
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return newUpdateAssetScanError(fmt.Errorf("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return newUpdateAssetScanError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return newUpdateAssetScanError(fmt.Errorf("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return newUpdateAssetScanError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return newUpdateAssetScanError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func (b *BackendClient) PostScan(ctx context.Context, scan models.Scan) (*models.Scan, error) {
	resp, err := b.apiClient.PostScansWithResponse(ctx, scan)
	if err != nil {
		return nil, fmt.Errorf("failed to create a scan: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusCreated:
		if resp.JSON201 == nil {
			return nil, fmt.Errorf("failed to create a scan: empty body. status code=%v", http.StatusCreated)
		}
		return resp.JSON201, nil
	case http.StatusBadRequest:
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			return nil, fmt.Errorf("failed to create a scan. status code=%v: %v", resp.StatusCode(), *resp.JSON400.Message)
		}
		return nil, fmt.Errorf("failed to create a scan. status code=%v", resp.StatusCode())
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
			return nil, fmt.Errorf("failed to create a scan. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to create a scan. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) PostAssetScan(ctx context.Context, assetScan models.AssetScan) (*models.AssetScan, error) {
	resp, err := b.apiClient.PostAssetScansWithResponse(ctx, assetScan)
	if err != nil {
		return nil, fmt.Errorf("failed to create an asset scan: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusCreated:
		if resp.JSON201 == nil {
			return nil, fmt.Errorf("failed to create an asset scan: empty body. status code=%v", http.StatusCreated)
		}
		return resp.JSON201, nil
	case http.StatusBadRequest:
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			return nil, fmt.Errorf("failed to create an asset scan. status code=%v: %v", resp.StatusCode(), *resp.JSON400.Message)
		}
		return nil, fmt.Errorf("failed to create an asset scan. status code=%v", resp.StatusCode())
	case http.StatusConflict:
		if resp.JSON409 == nil {
			return nil, fmt.Errorf("failed to create an asset scan: empty body. status code=%v", http.StatusConflict)
		}
		if resp.JSON409.AssetScan == nil {
			return nil, fmt.Errorf("failed to create an asset scan: no asset scan data. status code=%v", http.StatusConflict)
		}
		return nil, AssetScanConflictError{
			ConflictingAssetScan: resp.JSON409.AssetScan,
			Message:              "conflict",
		}
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to create an asset scan. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to create an asset scan. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) PatchScan(ctx context.Context, scanID models.ScanID, scan *models.Scan) error {
	params := models.PatchScansScanIDParams{}
	resp, err := b.apiClient.PatchScansScanIDWithResponse(ctx, scanID, &params, *scan)
	if err != nil {
		return fmt.Errorf("failed to update a scan: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return fmt.Errorf("failed to update a scan: empty body")
		}
		return nil
	case http.StatusBadRequest:
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			return fmt.Errorf("failed to update a scan. status code=%v: %v", resp.StatusCode(), *resp.JSON400.Message)
		}
		return fmt.Errorf("failed to update a scan. status code=%v", resp.StatusCode())
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return fmt.Errorf("failed to update a scan: empty body on not found")
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return fmt.Errorf("failed to update a scan, not found: %v", *resp.JSON404.Message)
		}
		return fmt.Errorf("failed to update a scan, not found")
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return fmt.Errorf("failed to update scan. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return fmt.Errorf("failed to update scan. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) GetAssetScanSummary(ctx context.Context, assetScanID string) (*models.ScanFindingsSummary, error) {
	params := models.GetAssetScansAssetScanIDParams{
		Select: utils.PointerTo("summary"),
	}
	assetScan, err := b.GetAssetScan(ctx, assetScanID, params)
	if err != nil {
		return nil, err
	}
	return assetScan.Summary, nil
}

func (b *BackendClient) GetAssetScanStatus(ctx context.Context, assetScanID string) (*models.AssetScanStatus, error) {
	params := models.GetAssetScansAssetScanIDParams{
		Select: utils.PointerTo("status"),
	}
	assetScan, err := b.GetAssetScan(ctx, assetScanID, params)
	if err != nil {
		return nil, err
	}
	return assetScan.Status, nil
}

func (b *BackendClient) PatchAssetScanStatus(ctx context.Context, assetScanID string, status *models.AssetScanStatus) error {
	assetScan := models.AssetScan{
		Status: status,
	}
	params := models.PatchAssetScansAssetScanIDParams{}
	resp, err := b.apiClient.PatchAssetScansAssetScanIDWithResponse(ctx, assetScanID, &params, assetScan)
	if err != nil {
		return fmt.Errorf("failed to update an asset scan status: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return fmt.Errorf("failed to update an asset scan status: empty body")
		}
		return nil
	case http.StatusBadRequest:
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			return fmt.Errorf("failed to update asset scan status. status code=%v: %v", resp.StatusCode(), *resp.JSON400.Message)
		}
		return fmt.Errorf("failed to update asset scan status. status code=%v", resp.StatusCode())
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return fmt.Errorf("failed to update an asset scan status: empty body on not found")
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return fmt.Errorf("failed to update asset scan status, not found: %v", *resp.JSON404.Message)
		}
		return fmt.Errorf("failed to update asset scan status, not found")
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return fmt.Errorf("failed to update asset scan status. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return fmt.Errorf("failed to update asset scan status. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) GetScan(ctx context.Context, scanID string, params models.GetScansScanIDParams) (*models.Scan, error) {
	resp, err := b.apiClient.GetScansScanIDWithResponse(ctx, scanID, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get a scan: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("failed to get a scan: empty body")
		}
		return resp.JSON200, nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return nil, fmt.Errorf("failed to get a scan: empty body on not found")
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return nil, fmt.Errorf("failed to get a scan, not found: %v", *resp.JSON404.Message)
		}
		return nil, fmt.Errorf("failed to get a scan, not found")
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get a scan status. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get a scan status. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) PostScanConfig(ctx context.Context, scanConfig models.ScanConfig) (*models.ScanConfig, error) {
	resp, err := b.apiClient.PostScanConfigsWithResponse(ctx, scanConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create scan config: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusCreated:
		if resp.JSON201 == nil {
			return nil, fmt.Errorf("failed to create a scan config: empty body. status code=%v", http.StatusCreated)
		}
		return resp.JSON201, nil
	case http.StatusBadRequest:
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			return nil, fmt.Errorf("failed to create a scan config. status code=%v: %v", resp.StatusCode(), *resp.JSON400.Message)
		}
		return nil, fmt.Errorf("failed to create a scan config. status code=%v", resp.StatusCode())
	case http.StatusConflict:
		if resp.JSON409 == nil {
			return nil, fmt.Errorf("failed to create a scan config: empty nody. status code=%v", http.StatusConflict)
		}
		if resp.JSON409.ScanConfig == nil {
			return nil, fmt.Errorf("failed to create a scan config: no scan config data. status code=%v", http.StatusConflict)
		}
		return nil, ScanConfigConflictError{
			ConflictingScanConfig: resp.JSON409.ScanConfig,
			Message:               "conflict",
		}
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to create a scan config. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to create a scan config. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) GetScanEstimation(ctx context.Context, scanEstimationID string, params models.GetScanEstimationsScanEstimationIDParams) (*models.ScanEstimation, error) {
	resp, err := b.apiClient.GetScanEstimationsScanEstimationIDWithResponse(ctx, scanEstimationID, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get a scan estimation: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("failed to get a scan estimation: empty body")
		}
		return resp.JSON200, nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return nil, fmt.Errorf("failed to get a scan estimation: empty body on not found")
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return nil, fmt.Errorf("failed to get a scan estimation, not found: %v", *resp.JSON404.Message)
		}
		return nil, fmt.Errorf("failed to get a scan estimation, not found")
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get a scan estimation status. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get a scan estimation status. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) GetScanEstimations(ctx context.Context, params models.GetScanEstimationsParams) (*models.ScanEstimations, error) {
	resp, err := b.apiClient.GetScanEstimationsWithResponse(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get scanEstimations: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("no scanEstimations: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get scanEstimations. status code=%v: %s", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get scanEstimations. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) PatchScanEstimation(ctx context.Context, scanEstimationID models.ScanEstimationID, scanEstimation *models.ScanEstimation) error {
	params := models.PatchScanEstimationsScanEstimationIDParams{}
	resp, err := b.apiClient.PatchScanEstimationsScanEstimationIDWithResponse(ctx, scanEstimationID, &params, *scanEstimation)
	if err != nil {
		return fmt.Errorf("failed to update a scan estimation: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return fmt.Errorf("failed to update a scan estimation: empty body")
		}
		return nil
	case http.StatusBadRequest:
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			return fmt.Errorf("failed to update a scan estimation. status code=%v: %v", resp.StatusCode(), *resp.JSON400.Message)
		}
		return fmt.Errorf("failed to update a scan estimation. status code=%v", resp.StatusCode())
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return fmt.Errorf("failed to update a scan estimation: empty body on not found")
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return fmt.Errorf("failed to update a scan estimation, not found: %v", *resp.JSON404.Message)
		}
		return fmt.Errorf("failed to update a scan estimation, not found")
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return fmt.Errorf("failed to update scan estimation. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return fmt.Errorf("failed to update scan estimation. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) DeleteScanEstimation(ctx context.Context, scanEstimationID models.ScanEstimationID) error {
	resp, err := b.apiClient.DeleteScanEstimationsScanEstimationIDWithResponse(ctx, scanEstimationID)
	if err != nil {
		return fmt.Errorf("failed to delete a scan estimation: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return fmt.Errorf("failed to delete a scan estimation: empty body")
		}
		return nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return fmt.Errorf("failed to delete a scan estimation: empty body on not found")
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return fmt.Errorf("failed to delete a scan estimation, not found: %v", *resp.JSON404.Message)
		}
		return fmt.Errorf("failed to delete a scan estimation, not found")
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return fmt.Errorf("failed to delete scan estimation. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return fmt.Errorf("failed to delete scan estimation. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) PostAssetScanEstimation(ctx context.Context, assetScanEstimation models.AssetScanEstimation) (*models.AssetScanEstimation, error) {
	resp, err := b.apiClient.PostAssetScanEstimationsWithResponse(ctx, assetScanEstimation)
	if err != nil {
		return nil, fmt.Errorf("failed to create an asset scan estimation: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusCreated:
		if resp.JSON201 == nil {
			return nil, fmt.Errorf("failed to create an asset scan estimation: empty body. status code=%v", http.StatusCreated)
		}
		return resp.JSON201, nil
	case http.StatusBadRequest:
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			return nil, fmt.Errorf("failed to create an asset scan estimation. status code=%v: %v", resp.StatusCode(), *resp.JSON400.Message)
		}
		return nil, fmt.Errorf("failed to create an asset scan estimation. status code=%v", resp.StatusCode())
	case http.StatusConflict:
		if resp.JSON409 == nil {
			return nil, fmt.Errorf("failed to create an asset scan estimation: empty body. status code=%v", http.StatusConflict)
		}
		if resp.JSON409.AssetScanEstimation == nil {
			return nil, fmt.Errorf("failed to create an asset scan estimation: no asset scan data. status code=%v", http.StatusConflict)
		}
		return nil, AssetScanEstimationConflictError{
			ConflictingAssetScanEstimation: resp.JSON409.AssetScanEstimation,
			Message:                        "conflict",
		}
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to create an asset scan estimation. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to create an asset scan estimation. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) DeleteAssetScanEstimation(ctx context.Context, assetScanEstimationID models.AssetScanEstimationID) error {
	resp, err := b.apiClient.DeleteAssetScanEstimationsAssetScanEstimationIDWithResponse(ctx, assetScanEstimationID)
	if err != nil {
		return fmt.Errorf("failed to delete a asset scan estimation: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return fmt.Errorf("failed to delete asset scan estimation: empty body")
		}
		return nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return fmt.Errorf("failed to delete asset scan estimation: empty body on not found")
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return fmt.Errorf("failed to delete asset scan estimation, not found: %v", *resp.JSON404.Message)
		}
		return fmt.Errorf("failed to delete asset scan estimation, not found")
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return fmt.Errorf("failed to delete asset scan estimation. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return fmt.Errorf("failed to delete asset scan estimation. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) GetAssetScanEstimations(ctx context.Context, params models.GetAssetScanEstimationsParams) (models.AssetScanEstimations, error) {
	newGetAssetScanEstimationsError := func(err error) error {
		return fmt.Errorf("failed to get asset scan estimations: %w", err)
	}

	var assetScanEstimations models.AssetScanEstimations
	resp, err := b.apiClient.GetAssetScanEstimationsWithResponse(ctx, &params)
	if err != nil {
		return assetScanEstimations, newGetAssetScanEstimationsError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return assetScanEstimations, newGetAssetScanEstimationsError(fmt.Errorf("empty body"))
		}
		return *resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return assetScanEstimations, newGetAssetScanEstimationsError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return assetScanEstimations, newGetAssetScanEstimationsError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func (b *BackendClient) GetAssetScanEstimation(ctx context.Context, assetScanEstimationID string, params models.GetAssetScanEstimationsAssetScanEstimationIDParams) (models.AssetScanEstimation, error) {
	newGetExistingError := func(err error) error {
		return fmt.Errorf("failed to get existing asset scan estimation %v: %w", assetScanEstimationID, err)
	}

	var assetScanEstimations models.AssetScanEstimation
	resp, err := b.apiClient.GetAssetScanEstimationsAssetScanEstimationIDWithResponse(ctx, assetScanEstimationID, &params)
	if err != nil {
		return assetScanEstimations, newGetExistingError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return assetScanEstimations, newGetExistingError(fmt.Errorf("empty body"))
		}
		return *resp.JSON200, nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return assetScanEstimations, newGetExistingError(fmt.Errorf("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return assetScanEstimations, newGetExistingError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return assetScanEstimations, newGetExistingError(fmt.Errorf("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return assetScanEstimations, newGetExistingError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return assetScanEstimations, newGetExistingError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func (b *BackendClient) PatchAssetScanEstimation(ctx context.Context, assetScanEstimation models.AssetScanEstimation, assetScanEstimationID string) error {
	newUpdateAssetScanEstimationError := func(err error) error {
		return fmt.Errorf("failed to update asset scan estimation %v: %w", assetScanEstimationID, err)
	}

	params := models.PatchAssetScanEstimationsAssetScanEstimationIDParams{}
	resp, err := b.apiClient.PatchAssetScanEstimationsAssetScanEstimationIDWithResponse(ctx, assetScanEstimationID, &params, assetScanEstimation)
	if err != nil {
		return newUpdateAssetScanEstimationError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return newUpdateAssetScanEstimationError(fmt.Errorf("empty body"))
		}
		return nil
	case http.StatusBadRequest:
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			return newUpdateAssetScanEstimationError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSON400.Message))
		}
		return newUpdateAssetScanEstimationError(fmt.Errorf("status code=%v", resp.StatusCode()))
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return newUpdateAssetScanEstimationError(fmt.Errorf("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return newUpdateAssetScanEstimationError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return newUpdateAssetScanEstimationError(fmt.Errorf("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return newUpdateAssetScanEstimationError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return newUpdateAssetScanEstimationError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func (b *BackendClient) GetScanConfigs(ctx context.Context, params models.GetScanConfigsParams) (*models.ScanConfigs, error) {
	resp, err := b.apiClient.GetScanConfigsWithResponse(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get scan configs: %w", err)
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

func (b *BackendClient) GetScanConfig(ctx context.Context, scanConfigID string, params models.GetScanConfigsScanConfigIDParams) (*models.ScanConfig, error) {
	resp, err := b.apiClient.GetScanConfigsScanConfigIDWithResponse(ctx, scanConfigID, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get a scan config: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("failed to get scan config: empty body")
		}
		return resp.JSON200, nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return nil, fmt.Errorf("failed to get a scan config: empty body on not found")
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return nil, fmt.Errorf("failed to get a scan config, not found: %v", *resp.JSON404.Message)
		}
		return nil, fmt.Errorf("failed to get a scan config, not found")
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get a scan config. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get a scan config. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) PatchScanConfig(ctx context.Context, scanConfigID string, scanConfig *models.ScanConfig) error {
	newPatchScanConfigResultError := func(err error) error {
		return fmt.Errorf("failed to update scan config %v: %w", scanConfigID, err)
	}

	params := models.PatchScanConfigsScanConfigIDParams{}
	resp, err := b.apiClient.PatchScanConfigsScanConfigIDWithResponse(ctx, scanConfigID, &params, *scanConfig)
	if err != nil {
		return newPatchScanConfigResultError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return newPatchScanConfigResultError(fmt.Errorf("empty body"))
		}
		return nil
	case http.StatusBadRequest:
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			return newPatchScanConfigResultError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSON400.Message))
		}
		return newPatchScanConfigResultError(fmt.Errorf("status code=%v", resp.StatusCode()))
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
		return nil, fmt.Errorf("failed to get scans: %w", err)
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

//nolint:cyclop
func (b *BackendClient) PostAsset(ctx context.Context, asset models.Asset) (*models.Asset, error) {
	resp, err := b.apiClient.PostAssetsWithResponse(ctx, asset)
	if err != nil {
		return nil, fmt.Errorf("failed to create an asset: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusCreated:
		if resp.JSON201 == nil {
			return nil, fmt.Errorf("failed to create an asset: empty body. status code=%v", http.StatusCreated)
		}
		return resp.JSON201, nil
	case http.StatusBadRequest:
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			return nil, fmt.Errorf("failed to create an asset. status code=%v: %v", resp.StatusCode(), *resp.JSON400.Message)
		}
		return nil, fmt.Errorf("failed to create an asset. status code=%v", resp.StatusCode())
	case http.StatusConflict:
		if resp.JSON409 == nil {
			return nil, fmt.Errorf("failed to create an asset: empty body. status code=%v", http.StatusConflict)
		}
		if resp.JSON409.Asset == nil {
			return nil, fmt.Errorf("failed to create an asset: no asset data. status code=%v", http.StatusConflict)
		}
		return nil, AssetConflictError{
			ConflictingAsset: resp.JSON409.Asset,
			Message:          "conflict",
		}
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to create an asset. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to create an asset. status code=%v", resp.StatusCode())
	}
}

//nolint:cyclop
func (b *BackendClient) PatchAsset(ctx context.Context, asset models.Asset, assetID string) error {
	newUpdateAssetError := func(err error) error {
		return fmt.Errorf("failed to update asset %v: %w", assetID, err)
	}

	params := models.PatchAssetsAssetIDParams{}
	resp, err := b.apiClient.PatchAssetsAssetIDWithResponse(ctx, assetID, &params, asset)
	if err != nil {
		return newUpdateAssetError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return newUpdateAssetError(fmt.Errorf("empty body"))
		}
		return nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return newUpdateAssetError(fmt.Errorf("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return newUpdateAssetError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return newUpdateAssetError(fmt.Errorf("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return newUpdateAssetError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return newUpdateAssetError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

// nolint:cyclop
func (b *BackendClient) GetAsset(ctx context.Context, assetID string, params models.GetAssetsAssetIDParams) (models.Asset, error) {
	newGetExistingError := func(err error) error {
		return fmt.Errorf("failed to get existing asset %v: %w", assetID, err)
	}

	var asset models.Asset
	resp, err := b.apiClient.GetAssetsAssetIDWithResponse(ctx, assetID, &params)
	if err != nil {
		return asset, newGetExistingError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return asset, newGetExistingError(fmt.Errorf("empty body"))
		}
		return *resp.JSON200, nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return asset, newGetExistingError(fmt.Errorf("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return asset, newGetExistingError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return asset, newGetExistingError(fmt.Errorf("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return asset, newGetExistingError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return asset, newGetExistingError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func (b *BackendClient) GetAssets(ctx context.Context, params models.GetAssetsParams) (*models.Assets, error) {
	resp, err := b.apiClient.GetAssetsWithResponse(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get assets: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("no assets: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get assets. status code=%v: %s", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get assets. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) GetFindings(ctx context.Context, params models.GetFindingsParams) (*models.Findings, error) {
	resp, err := b.apiClient.GetFindingsWithResponse(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get findings: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("no findings: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get findings. status code=%v: %s", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get findings. status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) PatchFinding(ctx context.Context, findingID models.FindingID, finding models.Finding) error {
	resp, err := b.apiClient.PatchFindingsFindingIDWithResponse(ctx, findingID, finding)
	if err != nil {
		return fmt.Errorf("failed to update a finding: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return fmt.Errorf("failed to update a finding: empty body")
		}
		return nil
	case http.StatusBadRequest:
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			return fmt.Errorf("failed to update a finding: status code=%v: %v", resp.StatusCode(), *resp.JSON400.Message)
		}
		return fmt.Errorf("failed to update a finding: status code=%v", resp.StatusCode())
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return fmt.Errorf("failed to update a finding: empty body on not found")
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return fmt.Errorf("failed to update a finding: not found: %v", *resp.JSON404.Message)
		}
		return fmt.Errorf("failed to update a finding: not found")
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return fmt.Errorf("failed to update a finding: status code=%v: %v", resp.StatusCode(), resp.JSONDefault.Message)
		}
		return fmt.Errorf("failed to update a finding: status code=%v", resp.StatusCode())
	}
}

func (b *BackendClient) PostFinding(ctx context.Context, finding models.Finding) (*models.Finding, error) {
	resp, err := b.apiClient.PostFindingsWithResponse(ctx, finding)
	if err != nil {
		return nil, fmt.Errorf("failed to create a finding: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusCreated:
		if resp.JSON201 == nil {
			return nil, fmt.Errorf("failed to create a finding: empty body. status code=%v", http.StatusCreated)
		}
		return resp.JSON201, nil
	case http.StatusConflict:
		if resp.JSON409 == nil {
			return nil, fmt.Errorf("failed to create a finding: empty body. status code=%v", http.StatusConflict)
		}
		if resp.JSON409.Finding == nil {
			return nil, fmt.Errorf("failed to create a finding: no finding data. status code=%v", http.StatusConflict)
		}
		return nil, FindingConflictError{
			ConflictingFinding: resp.JSON409.Finding,
			Message:            "conflict",
		}
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to create a finding. status code=%v: %v", resp.StatusCode(), resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to create a finding. status code=%v", resp.StatusCode())
	}
}

//nolint:cyclop
func (b *BackendClient) PostProvider(ctx context.Context, provider models.Provider) (*models.Provider, error) {
	resp, err := b.apiClient.PostProvidersWithResponse(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create a provider: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusCreated:
		if resp.JSON201 == nil {
			return nil, fmt.Errorf("failed to create a provider: empty body. status code=%v", http.StatusCreated)
		}
		return resp.JSON201, nil
	case http.StatusBadRequest:
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			return nil, fmt.Errorf("failed to create a provider. status code=%v: %v", resp.StatusCode(), *resp.JSON400.Message)
		}
		return nil, fmt.Errorf("failed to create a provider. status code=%v", resp.StatusCode())
	case http.StatusConflict:
		if resp.JSON409 == nil {
			return nil, fmt.Errorf("failed to create a provider: empty body. status code=%v", http.StatusConflict)
		}
		if resp.JSON409.Provider == nil {
			return nil, fmt.Errorf("failed to create a provider: no provider data. status code=%v", http.StatusConflict)
		}
		return nil, ProviderConflictError{
			ConflictingProvider: resp.JSON409.Provider,
			Message:             "conflict",
		}
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to create a provider. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to create a provider. status code=%v", resp.StatusCode())
	}
}

//nolint:cyclop
func (b *BackendClient) PatchProvider(ctx context.Context, provider models.Provider, providerID string) error {
	newUpdateProviderError := func(err error) error {
		return fmt.Errorf("failed to update provider %v: %w", providerID, err)
	}

	params := models.PatchProvidersProviderIDParams{}
	resp, err := b.apiClient.PatchProvidersProviderIDWithResponse(ctx, providerID, &params, provider)
	if err != nil {
		return newUpdateProviderError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return newUpdateProviderError(fmt.Errorf("empty body"))
		}
		return nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return newUpdateProviderError(fmt.Errorf("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return newUpdateProviderError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return newUpdateProviderError(fmt.Errorf("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return newUpdateProviderError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return newUpdateProviderError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

// nolint:cyclop
func (b *BackendClient) GetProvider(ctx context.Context, providerID string, params models.GetProvidersProviderIDParams) (models.Provider, error) {
	newGetExistingError := func(err error) error {
		return fmt.Errorf("failed to get existing provider %v: %w", providerID, err)
	}

	var provider models.Provider
	resp, err := b.apiClient.GetProvidersProviderIDWithResponse(ctx, providerID, &params)
	if err != nil {
		return provider, newGetExistingError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return provider, newGetExistingError(fmt.Errorf("empty body"))
		}
		return *resp.JSON200, nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return provider, newGetExistingError(fmt.Errorf("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return provider, newGetExistingError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return provider, newGetExistingError(fmt.Errorf("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return provider, newGetExistingError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return provider, newGetExistingError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func (b *BackendClient) GetProviders(ctx context.Context, params models.GetProvidersParams) (*models.Providers, error) {
	resp, err := b.apiClient.GetProvidersWithResponse(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("no providers: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get providers. status code=%v: %s", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get providers. status code=%v", resp.StatusCode())
	}
}
