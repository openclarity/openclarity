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

package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	apiclient "github.com/openclarity/vmclarity/uibackend/client/internal/client"
	"github.com/openclarity/vmclarity/uibackend/types"
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

func (c *Client) GetDashboardRiskiestAssets(ctx context.Context) (*types.RiskiestAssets, error) {
	resp, err := c.api.GetDashboardRiskiestAssetsWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard riskiest assets: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, errors.New("failed to get dashboard riskiest assets: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get dashboard riskiest assets: status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get dashboard riskiest assets: status code=%v", resp.StatusCode())
	}
}

func (c *Client) GetDashboardRiskiestRegions(ctx context.Context) (*types.RiskiestRegions, error) {
	resp, err := c.api.GetDashboardRiskiestRegionsWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard riskiest regions: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, errors.New("failed to get dashboard riskiest regions: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get dashboard riskiest regions: status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get dashboard riskiest regions: status code=%v", resp.StatusCode())
	}
}

func (c *Client) GetDashboardFindingsImpact(ctx context.Context) (*types.FindingsImpact, error) {
	resp, err := c.api.GetDashboardFindingsImpactWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard findings trends: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, errors.New("failed to get dashboard findings trends: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get dashboard findings trends: status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get dashboard findings trends: status code=%v", resp.StatusCode())
	}
}

func (c *Client) GetDashboardFindingsTrends(ctx context.Context, params types.GetDashboardFindingsTrendsParams) (*[]types.FindingTrends, error) {
	resp, err := c.api.GetDashboardFindingsTrendsWithResponse(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard findings trends: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if resp.JSON200 == nil {
			return nil, errors.New("failed to get dashboard findings trends: empty body")
		}
		return resp.JSON200, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return nil, fmt.Errorf("failed to get dashboard findings trends: status code=%v: %v", resp.StatusCode(), *resp.JSONDefault.Message)
		}
		return nil, fmt.Errorf("failed to get dashboard findings trends: status code=%v", resp.StatusCode())
	}
}
