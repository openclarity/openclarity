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
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/containerruntimediscovery/types"
)

type Client struct {
	endpoint string
}

func NewClient(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
	}
}

func (c *Client) GetImages(ctx context.Context) ([]apitypes.ContainerImageInfo, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s/images", c.endpoint), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request to discoverer: %w", err)
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to contact discoverer: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var lir types.ListImagesResponse
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&lir)
		if err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return lir.Images, nil
	default:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("unexpected error status %d, failed to read body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("unexpected error status %d: %v", resp.StatusCode, string(body))
	}
}

func (c *Client) GetImage(ctx context.Context, imageID string) (apitypes.ContainerImageInfo, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s/images/%s", c.endpoint, imageID), nil)
	if err != nil {
		return apitypes.ContainerImageInfo{}, fmt.Errorf("unable to create request to discoverer: %w", err)
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return apitypes.ContainerImageInfo{}, fmt.Errorf("unable to contact discoverer: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var info apitypes.ContainerImageInfo
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&info)
		if err != nil {
			return apitypes.ContainerImageInfo{}, fmt.Errorf("failed to decode response: %w", err)
		}
		return info, nil
	case http.StatusNotFound:
		return apitypes.ContainerImageInfo{}, types.ErrNotFound
	default:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return apitypes.ContainerImageInfo{}, fmt.Errorf("unexpected error status %d, failed to read body: %w", resp.StatusCode, err)
		}
		return apitypes.ContainerImageInfo{}, fmt.Errorf("unexpected error status %d: %v", resp.StatusCode, string(body))
	}
}

func (c *Client) GetContainers(ctx context.Context) ([]apitypes.ContainerInfo, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s/containers", c.endpoint), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request to discoverer: %w", err)
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to contact discoverer: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var lcr types.ListContainersResponse
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&lcr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		return lcr.Containers, nil
	default:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("unexpected error status %d, failed to read body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("unexpected error status %d: %v", resp.StatusCode, string(body))
	}
}

func (c *Client) GetContainer(ctx context.Context, containerID string) (apitypes.ContainerInfo, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s/containers/%s", c.endpoint, containerID), nil)
	if err != nil {
		return apitypes.ContainerInfo{}, fmt.Errorf("unable to create request to discoverer: %w", err)
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return apitypes.ContainerInfo{}, fmt.Errorf("unable to contact discoverer: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var info apitypes.ContainerInfo
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&info)
		if err != nil {
			return apitypes.ContainerInfo{}, fmt.Errorf("failed to decode response: %w", err)
		}
		return info, nil
	case http.StatusNotFound:
		return apitypes.ContainerInfo{}, types.ErrNotFound
	default:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return apitypes.ContainerInfo{}, fmt.Errorf("unexpected error status %d, failed to read body: %w", resp.StatusCode, err)
		}
		return apitypes.ContainerInfo{}, fmt.Errorf("unexpected error status %d: %v", resp.StatusCode, string(body))
	}
}

func (c *Client) ExportImageURL(_ context.Context, imageID string) string {
	return fmt.Sprintf("http://%s/exportimage/%s", c.endpoint, imageID)
}

func (c *Client) ExportContainerURL(_ context.Context, containerID string) string {
	return fmt.Sprintf("http://%s/exportcontainer/%s", c.endpoint, containerID)
}
