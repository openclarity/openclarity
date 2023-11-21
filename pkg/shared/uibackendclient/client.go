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

package uibackendclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/openclarity/vmclarity/pkg/uibackend/api/client"
	"github.com/openclarity/vmclarity/pkg/uibackend/api/models"
)

type UIBackendClient struct {
	apiClient client.ClientWithResponsesInterface
}

func Create(serverAddress string) (*UIBackendClient, error) {
	apiClient, err := client.NewClientWithResponses(serverAddress)
	if err != nil {
		return nil, fmt.Errorf("unable to create VMClarity API client. serverAddress=%v: %w", serverAddress, err)
	}

	return &UIBackendClient{
		apiClient: apiClient,
	}, nil
}

func (b *UIBackendClient) GetDashboardRiskiestAssets(ctx context.Context) (*models.RiskiestAssets, error) {
	resp, err := b.apiClient.GetDashboardRiskiestAssetsWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard riskiest assets: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("failed to get dashboard riskiest assets: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get dashboard riskiest assets: status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get dashboard riskiest assets: status code=%v", resp.StatusCode())
	}
}

func (b *UIBackendClient) GetDashboardRiskiestRegions(ctx context.Context) (*models.RiskiestRegions, error) {
	resp, err := b.apiClient.GetDashboardRiskiestRegionsWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard riskiest regions: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("failed to get dashboard riskiest regions: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get dashboard riskiest regions: status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get dashboard riskiest regions: status code=%v", resp.StatusCode())
	}
}

func (b *UIBackendClient) GetDashboardFindingsImpact(ctx context.Context) (*models.FindingsImpact, error) {
	resp, err := b.apiClient.GetDashboardFindingsImpactWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard findings trends: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("failed to get dashboard findings trends: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get dashboard findings trends: status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get dashboard findings trends: status code=%v", resp.StatusCode())
	}
}

func (b *UIBackendClient) GetDashboardFindingsTrends(ctx context.Context, params models.GetDashboardFindingsTrendsParams) (*[]models.FindingTrends, error) {
	resp, err := b.apiClient.GetDashboardFindingsTrendsWithResponse(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard findings trends: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, fmt.Errorf("failed to get dashboard findings trends: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get dashboard findings trends: status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get dashboard findings trends: status code=%v", resp.StatusCode())
	}
}
