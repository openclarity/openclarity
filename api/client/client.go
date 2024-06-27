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
package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	apiclient "github.com/openclarity/vmclarity/api/client/internal/client"
	"github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
)

type Client struct {
	api apiclient.ClientWithResponsesInterface
}

func New(serverAddress string) (*Client, error) {
	api, err := apiclient.NewClientWithResponses(serverAddress)
	if err != nil {
		return nil, fmt.Errorf("unable to create VMClarity API client. serverAddress=%v: %w", serverAddress, err)
	}

	return &Client{
		api: api,
	}, nil
}

func (c *Client) GetAssetScan(ctx context.Context, assetScanID string, params types.GetAssetScansAssetScanIDParams) (types.AssetScan, error) {
	newGetExistingError := func(err error) error {
		return fmt.Errorf("failed to get existing asset scan %v: %w", assetScanID, err)
	}

	var assetScans types.AssetScan
	resp, err := c.api.GetAssetScansAssetScanIDWithResponse(ctx, assetScanID, &params)
	if err != nil {
		return assetScans, newGetExistingError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return assetScans, newGetExistingError(errors.New("empty body"))
		}
		return *resp.JSON200, nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return assetScans, newGetExistingError(errors.New("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return assetScans, newGetExistingError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return assetScans, newGetExistingError(errors.New("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return assetScans, newGetExistingError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return assetScans, newGetExistingError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func (c *Client) GetAssetScans(ctx context.Context, params types.GetAssetScansParams) (types.AssetScans, error) {
	newGetAssetScansError := func(err error) error {
		return fmt.Errorf("failed to get asset scans: %w", err)
	}

	var assetScans types.AssetScans
	resp, err := c.api.GetAssetScansWithResponse(ctx, &params)
	if err != nil {
		return assetScans, newGetAssetScansError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return assetScans, newGetAssetScansError(errors.New("empty body"))
		}
		return *resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return assetScans, newGetAssetScansError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return assetScans, newGetAssetScansError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func (c *Client) PatchAssetScan(ctx context.Context, assetScan types.AssetScan, assetScanID string) error {
	newUpdateAssetScanError := func(err error) error {
		return fmt.Errorf("failed to update asset scan %v: %w", assetScanID, err)
	}

	params := types.PatchAssetScansAssetScanIDParams{}
	resp, err := c.api.PatchAssetScansAssetScanIDWithResponse(ctx, assetScanID, &params, assetScan)
	if err != nil {
		return newUpdateAssetScanError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return newUpdateAssetScanError(errors.New("empty body"))
		}
		return nil
	case http.StatusBadRequest:
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			return newUpdateAssetScanError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSON400.Message))
		}
		return newUpdateAssetScanError(fmt.Errorf("status code=%v", resp.StatusCode()))
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return newUpdateAssetScanError(errors.New("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return newUpdateAssetScanError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return newUpdateAssetScanError(errors.New("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return newUpdateAssetScanError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return newUpdateAssetScanError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func (c *Client) PostScan(ctx context.Context, scan types.Scan) (*types.Scan, error) {
	resp, err := c.api.PostScansWithResponse(ctx, scan)
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

func (c *Client) PostAssetScan(ctx context.Context, assetScan types.AssetScan) (*types.AssetScan, error) {
	resp, err := c.api.PostAssetScansWithResponse(ctx, assetScan)
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

func (c *Client) PatchScan(ctx context.Context, scanID types.ScanID, scan *types.Scan) error {
	params := types.PatchScansScanIDParams{}
	resp, err := c.api.PatchScansScanIDWithResponse(ctx, scanID, &params, *scan)
	if err != nil {
		return fmt.Errorf("failed to update a scan: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return errors.New("failed to update a scan: empty body")
		}
		return nil
	case http.StatusBadRequest:
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			return fmt.Errorf("failed to update a scan. status code=%v: %v", resp.StatusCode(), *resp.JSON400.Message)
		}
		return fmt.Errorf("failed to update a scan. status code=%v", resp.StatusCode())
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return errors.New("failed to update a scan: empty body on not found")
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return fmt.Errorf("failed to update a scan, not found: %v", *resp.JSON404.Message)
		}
		return errors.New("failed to update a scan, not found")
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return fmt.Errorf("failed to update scan. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return fmt.Errorf("failed to update scan. status code=%v", resp.StatusCode())
	}
}

func (c *Client) GetAssetScanSummary(ctx context.Context, assetScanID string) (*types.ScanFindingsSummary, error) {
	params := types.GetAssetScansAssetScanIDParams{
		Select: to.Ptr("summary"),
	}
	assetScan, err := c.GetAssetScan(ctx, assetScanID, params)
	if err != nil {
		return nil, err
	}
	return assetScan.Summary, nil
}

func (c *Client) GetAssetScanStatus(ctx context.Context, assetScanID string) (*types.AssetScanStatus, error) {
	params := types.GetAssetScansAssetScanIDParams{
		Select: to.Ptr("status"),
	}
	assetScan, err := c.GetAssetScan(ctx, assetScanID, params)
	if err != nil {
		return nil, err
	}
	return assetScan.Status, nil
}

func (c *Client) PatchAssetScanStatus(ctx context.Context, assetScanID string, status *types.AssetScanStatus) error {
	assetScan := types.AssetScan{
		Status: status,
	}
	params := types.PatchAssetScansAssetScanIDParams{}
	resp, err := c.api.PatchAssetScansAssetScanIDWithResponse(ctx, assetScanID, &params, assetScan)
	if err != nil {
		return fmt.Errorf("failed to update an asset scan status: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return errors.New("failed to update an asset scan status: empty body")
		}
		return nil
	case http.StatusBadRequest:
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			return fmt.Errorf("failed to update asset scan status. status code=%v: %v", resp.StatusCode(), *resp.JSON400.Message)
		}
		return fmt.Errorf("failed to update asset scan status. status code=%v", resp.StatusCode())
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return errors.New("failed to update an asset scan status: empty body on not found")
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return fmt.Errorf("failed to update asset scan status, not found: %v", *resp.JSON404.Message)
		}
		return errors.New("failed to update asset scan status, not found")
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return fmt.Errorf("failed to update asset scan status. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return fmt.Errorf("failed to update asset scan status. status code=%v", resp.StatusCode())
	}
}

func (c *Client) GetScan(ctx context.Context, scanID string, params types.GetScansScanIDParams) (*types.Scan, error) {
	resp, err := c.api.GetScansScanIDWithResponse(ctx, scanID, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get a scan: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, errors.New("failed to get a scan: empty body")
		}
		return resp.JSON200, nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return nil, errors.New("failed to get a scan: empty body on not found")
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return nil, fmt.Errorf("failed to get a scan, not found: %v", *resp.JSON404.Message)
		}
		return nil, errors.New("failed to get a scan, not found")
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get a scan status. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get a scan status. status code=%v", resp.StatusCode())
	}
}

func (c *Client) PostScanConfig(ctx context.Context, scanConfig types.ScanConfig) (*types.ScanConfig, error) {
	resp, err := c.api.PostScanConfigsWithResponse(ctx, scanConfig)
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

func (c *Client) GetScanEstimation(ctx context.Context, scanEstimationID string, params types.GetScanEstimationsScanEstimationIDParams) (*types.ScanEstimation, error) {
	resp, err := c.api.GetScanEstimationsScanEstimationIDWithResponse(ctx, scanEstimationID, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get a scan estimation: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, errors.New("failed to get a scan estimation: empty body")
		}
		return resp.JSON200, nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return nil, errors.New("failed to get a scan estimation: empty body on not found")
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return nil, fmt.Errorf("failed to get a scan estimation, not found: %v", *resp.JSON404.Message)
		}
		return nil, errors.New("failed to get a scan estimation, not found")
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get a scan estimation status. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get a scan estimation status. status code=%v", resp.StatusCode())
	}
}

func (c *Client) GetScanEstimations(ctx context.Context, params types.GetScanEstimationsParams) (*types.ScanEstimations, error) {
	resp, err := c.api.GetScanEstimationsWithResponse(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get scanEstimations: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, errors.New("no scanEstimations: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get scanEstimations. status code=%v: %s", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get scanEstimations. status code=%v", resp.StatusCode())
	}
}

func (c *Client) PatchScanEstimation(ctx context.Context, scanEstimationID types.ScanEstimationID, scanEstimation *types.ScanEstimation) error {
	params := types.PatchScanEstimationsScanEstimationIDParams{}
	resp, err := c.api.PatchScanEstimationsScanEstimationIDWithResponse(ctx, scanEstimationID, &params, *scanEstimation)
	if err != nil {
		return fmt.Errorf("failed to update a scan estimation: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return errors.New("failed to update a scan estimation: empty body")
		}
		return nil
	case http.StatusBadRequest:
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			return fmt.Errorf("failed to update a scan estimation. status code=%v: %v", resp.StatusCode(), *resp.JSON400.Message)
		}
		return fmt.Errorf("failed to update a scan estimation. status code=%v", resp.StatusCode())
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return errors.New("failed to update a scan estimation: empty body on not found")
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return fmt.Errorf("failed to update a scan estimation, not found: %v", *resp.JSON404.Message)
		}
		return errors.New("failed to update a scan estimation, not found")
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return fmt.Errorf("failed to update scan estimation. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return fmt.Errorf("failed to update scan estimation. status code=%v", resp.StatusCode())
	}
}

func (c *Client) DeleteScanEstimation(ctx context.Context, scanEstimationID types.ScanEstimationID) error {
	resp, err := c.api.DeleteScanEstimationsScanEstimationIDWithResponse(ctx, scanEstimationID)
	if err != nil {
		return fmt.Errorf("failed to delete a scan estimation: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return errors.New("failed to delete a scan estimation: empty body")
		}
		return nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return errors.New("failed to delete a scan estimation: empty body on not found")
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return fmt.Errorf("failed to delete a scan estimation, not found: %v", *resp.JSON404.Message)
		}
		return errors.New("failed to delete a scan estimation, not found")
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return fmt.Errorf("failed to delete scan estimation. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return fmt.Errorf("failed to delete scan estimation. status code=%v", resp.StatusCode())
	}
}

func (c *Client) PostAssetScanEstimation(ctx context.Context, assetScanEstimation types.AssetScanEstimation) (*types.AssetScanEstimation, error) {
	resp, err := c.api.PostAssetScanEstimationsWithResponse(ctx, assetScanEstimation)
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

func (c *Client) DeleteAssetScanEstimation(ctx context.Context, assetScanEstimationID types.AssetScanEstimationID) error {
	resp, err := c.api.DeleteAssetScanEstimationsAssetScanEstimationIDWithResponse(ctx, assetScanEstimationID)
	if err != nil {
		return fmt.Errorf("failed to delete a asset scan estimation: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return errors.New("failed to delete asset scan estimation: empty body")
		}
		return nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return errors.New("failed to delete asset scan estimation: empty body on not found")
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return fmt.Errorf("failed to delete asset scan estimation, not found: %v", *resp.JSON404.Message)
		}
		return errors.New("failed to delete asset scan estimation, not found")
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return fmt.Errorf("failed to delete asset scan estimation. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return fmt.Errorf("failed to delete asset scan estimation. status code=%v", resp.StatusCode())
	}
}

func (c *Client) GetAssetScanEstimations(ctx context.Context, params types.GetAssetScanEstimationsParams) (types.AssetScanEstimations, error) {
	newGetAssetScanEstimationsError := func(err error) error {
		return fmt.Errorf("failed to get asset scan estimations: %w", err)
	}

	var assetScanEstimations types.AssetScanEstimations
	resp, err := c.api.GetAssetScanEstimationsWithResponse(ctx, &params)
	if err != nil {
		return assetScanEstimations, newGetAssetScanEstimationsError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return assetScanEstimations, newGetAssetScanEstimationsError(errors.New("empty body"))
		}
		return *resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return assetScanEstimations, newGetAssetScanEstimationsError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return assetScanEstimations, newGetAssetScanEstimationsError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func (c *Client) GetAssetScanEstimation(ctx context.Context, assetScanEstimationID string, params types.GetAssetScanEstimationsAssetScanEstimationIDParams) (types.AssetScanEstimation, error) {
	newGetExistingError := func(err error) error {
		return fmt.Errorf("failed to get existing asset scan estimation %v: %w", assetScanEstimationID, err)
	}

	var assetScanEstimations types.AssetScanEstimation
	resp, err := c.api.GetAssetScanEstimationsAssetScanEstimationIDWithResponse(ctx, assetScanEstimationID, &params)
	if err != nil {
		return assetScanEstimations, newGetExistingError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return assetScanEstimations, newGetExistingError(errors.New("empty body"))
		}
		return *resp.JSON200, nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return assetScanEstimations, newGetExistingError(errors.New("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return assetScanEstimations, newGetExistingError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return assetScanEstimations, newGetExistingError(errors.New("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return assetScanEstimations, newGetExistingError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return assetScanEstimations, newGetExistingError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func (c *Client) PatchAssetScanEstimation(ctx context.Context, assetScanEstimation types.AssetScanEstimation, assetScanEstimationID string) error {
	newUpdateAssetScanEstimationError := func(err error) error {
		return fmt.Errorf("failed to update asset scan estimation %v: %w", assetScanEstimationID, err)
	}

	params := types.PatchAssetScanEstimationsAssetScanEstimationIDParams{}
	resp, err := c.api.PatchAssetScanEstimationsAssetScanEstimationIDWithResponse(ctx, assetScanEstimationID, &params, assetScanEstimation)
	if err != nil {
		return newUpdateAssetScanEstimationError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return newUpdateAssetScanEstimationError(errors.New("empty body"))
		}
		return nil
	case http.StatusBadRequest:
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			return newUpdateAssetScanEstimationError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSON400.Message))
		}
		return newUpdateAssetScanEstimationError(fmt.Errorf("status code=%v", resp.StatusCode()))
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return newUpdateAssetScanEstimationError(errors.New("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return newUpdateAssetScanEstimationError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return newUpdateAssetScanEstimationError(errors.New("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return newUpdateAssetScanEstimationError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return newUpdateAssetScanEstimationError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func (c *Client) GetScanConfigs(ctx context.Context, params types.GetScanConfigsParams) (*types.ScanConfigs, error) {
	resp, err := c.api.GetScanConfigsWithResponse(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get scan configs: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, errors.New("no scan configs: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get scan configs. status code=%v: %s", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get scan configs. status code=%v", resp.StatusCode())
	}
}

func (c *Client) GetScanConfig(ctx context.Context, scanConfigID string, params types.GetScanConfigsScanConfigIDParams) (*types.ScanConfig, error) {
	resp, err := c.api.GetScanConfigsScanConfigIDWithResponse(ctx, scanConfigID, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get a scan config: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, errors.New("failed to get scan config: empty body")
		}
		return resp.JSON200, nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return nil, errors.New("failed to get a scan config: empty body on not found")
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return nil, fmt.Errorf("failed to get a scan config, not found: %v", *resp.JSON404.Message)
		}
		return nil, errors.New("failed to get a scan config, not found")
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get a scan config. status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get a scan config. status code=%v", resp.StatusCode())
	}
}

func (c *Client) PatchScanConfig(ctx context.Context, scanConfigID string, scanConfig *types.ScanConfig) error {
	newPatchScanConfigResultError := func(err error) error {
		return fmt.Errorf("failed to update scan config %v: %w", scanConfigID, err)
	}

	params := types.PatchScanConfigsScanConfigIDParams{}
	resp, err := c.api.PatchScanConfigsScanConfigIDWithResponse(ctx, scanConfigID, &params, *scanConfig)
	if err != nil {
		return newPatchScanConfigResultError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return newPatchScanConfigResultError(errors.New("empty body"))
		}
		return nil
	case http.StatusBadRequest:
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			return newPatchScanConfigResultError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSON400.Message))
		}
		return newPatchScanConfigResultError(fmt.Errorf("status code=%v", resp.StatusCode()))
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return newPatchScanConfigResultError(errors.New("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return newPatchScanConfigResultError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return newPatchScanConfigResultError(errors.New("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return newPatchScanConfigResultError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return newPatchScanConfigResultError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func (c *Client) GetScans(ctx context.Context, params types.GetScansParams) (*types.Scans, error) {
	resp, err := c.api.GetScansWithResponse(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get scans: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, errors.New("no scans: empty body")
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
func (c *Client) PostAsset(ctx context.Context, asset types.Asset) (*types.Asset, error) {
	resp, err := c.api.PostAssetsWithResponse(ctx, asset)
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
func (c *Client) PatchAsset(ctx context.Context, asset types.Asset, assetID string) error {
	newUpdateAssetError := func(err error) error {
		return fmt.Errorf("failed to update asset %v: %w", assetID, err)
	}

	params := types.PatchAssetsAssetIDParams{}
	resp, err := c.api.PatchAssetsAssetIDWithResponse(ctx, assetID, &params, asset)
	if err != nil {
		return newUpdateAssetError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return newUpdateAssetError(errors.New("empty body"))
		}
		return nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return newUpdateAssetError(errors.New("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return newUpdateAssetError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return newUpdateAssetError(errors.New("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return newUpdateAssetError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return newUpdateAssetError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

// nolint:cyclop
func (c *Client) GetAsset(ctx context.Context, assetID string, params types.GetAssetsAssetIDParams) (types.Asset, error) {
	newGetExistingError := func(err error) error {
		return fmt.Errorf("failed to get existing asset %v: %w", assetID, err)
	}

	var asset types.Asset
	resp, err := c.api.GetAssetsAssetIDWithResponse(ctx, assetID, &params)
	if err != nil {
		return asset, newGetExistingError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return asset, newGetExistingError(errors.New("empty body"))
		}
		return *resp.JSON200, nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return asset, newGetExistingError(errors.New("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return asset, newGetExistingError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return asset, newGetExistingError(errors.New("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return asset, newGetExistingError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return asset, newGetExistingError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func (c *Client) GetAssets(ctx context.Context, params types.GetAssetsParams) (*types.Assets, error) {
	resp, err := c.api.GetAssetsWithResponse(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get assets: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, errors.New("no assets: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get assets. status code=%v: %s", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get assets. status code=%v", resp.StatusCode())
	}
}

func (c *Client) GetFindings(ctx context.Context, params types.GetFindingsParams) (*types.Findings, error) {
	resp, err := c.api.GetFindingsWithResponse(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get findings: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, errors.New("no findings: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get findings. status code=%v: %s", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get findings. status code=%v", resp.StatusCode())
	}
}

func (c *Client) GetFinding(ctx context.Context, findingID types.FindingID, params types.GetFindingsFindingIDParams) (*types.Finding, error) {
	newGetExistingError := func(err error) error {
		return fmt.Errorf("failed to get existing finding %v: %w", findingID, err)
	}

	resp, err := c.api.GetFindingsFindingIDWithResponse(ctx, findingID, &params)
	if err != nil {
		return nil, newGetExistingError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, newGetExistingError(errors.New("empty body"))
		}
		return resp.JSON200, nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return nil, newGetExistingError(errors.New("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return nil, newGetExistingError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return nil, newGetExistingError(errors.New("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, newGetExistingError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return nil, newGetExistingError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func (c *Client) PatchFinding(ctx context.Context, findingID types.FindingID, finding types.Finding) error {
	resp, err := c.api.PatchFindingsFindingIDWithResponse(ctx, findingID, finding)
	if err != nil {
		return fmt.Errorf("failed to update a finding: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return errors.New("failed to update a finding: empty body")
		}
		return nil
	case http.StatusBadRequest:
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			return fmt.Errorf("failed to update a finding: status code=%v: %v", resp.StatusCode(), *resp.JSON400.Message)
		}
		return fmt.Errorf("failed to update a finding: status code=%v", resp.StatusCode())
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return errors.New("failed to update a finding: empty body on not found")
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return fmt.Errorf("failed to update a finding: not found: %v", *resp.JSON404.Message)
		}
		return errors.New("failed to update a finding: not found")
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return fmt.Errorf("failed to update a finding: status code=%v: %v", resp.StatusCode(), resp.JSONDefault.Message)
		}
		return fmt.Errorf("failed to update a finding: status code=%v", resp.StatusCode())
	}
}

func (c *Client) PostFinding(ctx context.Context, finding types.Finding) (*types.Finding, error) {
	resp, err := c.api.PostFindingsWithResponse(ctx, finding)
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
func (c *Client) PostProvider(ctx context.Context, provider types.Provider) (*types.Provider, error) {
	resp, err := c.api.PostProvidersWithResponse(ctx, provider)
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
func (c *Client) PatchProvider(ctx context.Context, provider types.Provider, providerID string) error {
	newUpdateProviderError := func(err error) error {
		return fmt.Errorf("failed to update provider %v: %w", providerID, err)
	}

	params := types.PatchProvidersProviderIDParams{}
	resp, err := c.api.PatchProvidersProviderIDWithResponse(ctx, providerID, &params, provider)
	if err != nil {
		return newUpdateProviderError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return newUpdateProviderError(errors.New("empty body"))
		}
		return nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return newUpdateProviderError(errors.New("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return newUpdateProviderError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return newUpdateProviderError(errors.New("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return newUpdateProviderError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return newUpdateProviderError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

// nolint:cyclop
func (c *Client) GetProvider(ctx context.Context, providerID string, params types.GetProvidersProviderIDParams) (types.Provider, error) {
	newGetExistingError := func(err error) error {
		return fmt.Errorf("failed to get existing provider %v: %w", providerID, err)
	}

	var provider types.Provider
	resp, err := c.api.GetProvidersProviderIDWithResponse(ctx, providerID, &params)
	if err != nil {
		return provider, newGetExistingError(err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return provider, newGetExistingError(errors.New("empty body"))
		}
		return *resp.JSON200, nil
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return provider, newGetExistingError(errors.New("empty body on not found"))
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return provider, newGetExistingError(fmt.Errorf("not found: %v", *resp.JSON404.Message))
		}
		return provider, newGetExistingError(errors.New("not found"))
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return provider, newGetExistingError(fmt.Errorf("status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message))
		}
		return provider, newGetExistingError(fmt.Errorf("status code=%v", resp.StatusCode()))
	}
}

func (c *Client) GetProviders(ctx context.Context, params types.GetProvidersParams) (*types.Providers, error) {
	resp, err := c.api.GetProvidersWithResponse(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, errors.New("no providers: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get providers. status code=%v: %s", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get providers. status code=%v", resp.StatusCode())
	}
}

func (c *Client) GetAssetFindings(ctx context.Context, params types.GetAssetFindingsParams) (*types.AssetFindings, error) {
	resp, err := c.api.GetAssetFindingsWithResponse(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset findings: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, errors.New("no asset findings: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get asset findings. status code=%v: %s", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get asset findings. status code=%v", resp.StatusCode())
	}
}

func (c *Client) PatchAssetFinding(ctx context.Context, assetFindingID types.AssetFindingID, assetFinding types.AssetFinding) error {
	resp, err := c.api.PatchAssetFindingsAssetFindingIDWithResponse(ctx, assetFindingID, &types.PatchAssetFindingsAssetFindingIDParams{}, assetFinding)
	if err != nil {
		return fmt.Errorf("failed to update an asset finding: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return errors.New("failed to update an asset finding: empty body")
		}
		return nil
	case http.StatusBadRequest:
		if resp.JSON400 != nil && resp.JSON400.Message != nil {
			return fmt.Errorf("failed to update an asset finding: status code=%v: %v", resp.StatusCode(), *resp.JSON400.Message)
		}
		return fmt.Errorf("failed to update an asset finding: status code=%v", resp.StatusCode())
	case http.StatusNotFound:
		if resp.JSON404 == nil {
			return errors.New("failed to update an asset finding: empty body on not found")
		}
		if resp.JSON404 != nil && resp.JSON404.Message != nil {
			return fmt.Errorf("failed to update an asset finding: not found: %v", *resp.JSON404.Message)
		}
		return errors.New("failed to update an asset finding: not found")
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return fmt.Errorf("failed to update an asset finding: status code=%v: %v", resp.StatusCode(), resp.JSONDefault.Message)
		}
		return fmt.Errorf("failed to update an asset finding: status code=%v", resp.StatusCode())
	}
}

func (c *Client) PostAssetFinding(ctx context.Context, assetFinding types.AssetFinding) (*types.AssetFinding, error) {
	resp, err := c.api.PostAssetFindingsWithResponse(ctx, assetFinding)
	if err != nil {
		return nil, fmt.Errorf("failed to create an asset finding: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusCreated:
		if resp.JSON201 == nil {
			return nil, fmt.Errorf("failed to create an asset finding: empty body. status code=%v", http.StatusCreated)
		}
		return resp.JSON201, nil
	case http.StatusConflict:
		if resp.JSON409 == nil {
			return nil, fmt.Errorf("failed to create an asset finding: empty body. status code=%v", http.StatusConflict)
		}
		if resp.JSON409.Finding == nil {
			return nil, fmt.Errorf("failed to create an asset finding: no finding data. status code=%v", http.StatusConflict)
		}
		return nil, AssetFindingConflictError{
			ConflictingAssetFinding: resp.JSON409.Finding,
			Message:                 "conflict",
		}
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to create an asset finding. status code=%v: %v", resp.StatusCode(), resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to create an asset finding. status code=%v", resp.StatusCode())
	}
}
