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

package discoverer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apitypes "github.com/openclarity/vmclarity/api/types"
	discoverytypes "github.com/openclarity/vmclarity/containerruntimediscovery/types"
	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/core/to"
)

const (
	PodNamespaceLabel = "io.kubernetes.pod.namespace"
	PodNameLabel      = "io.kubernetes.pod.name"
)

// nolint:cyclop
func (d *Discoverer) discoverContainers(ctx context.Context, outputChan chan apitypes.AssetType, crDiscoverers []corev1.Pod) error {
	for _, discoverer := range crDiscoverers {
		err := d.discoverContainersFromDiscoverer(ctx, outputChan, discoverer)
		if err != nil {
			return fmt.Errorf("failed to discover containers from discoverer %s: %w", discoverer.Name, err)
		}
	}

	return nil
}

func (d *Discoverer) discoverContainersFromDiscoverer(ctx context.Context, outputChan chan apitypes.AssetType, discoverer corev1.Pod) error {
	discovererEndpoint := net.JoinHostPort(discoverer.Status.PodIP, "8080")
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s/containers", discovererEndpoint), nil)
	if err != nil {
		return fmt.Errorf("unable to create request to discoverer: %w", err)
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("unable to contact discoverer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected error from discoverer status %s: %s", resp.Status, b)
	}

	var containerResponse discoverytypes.ListContainersResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&containerResponse)
	if err != nil {
		return fmt.Errorf("unable to decode response from discoverer: %w", err)
	}

	logger := log.GetLoggerFromContextOrDiscard(ctx)
	for _, container := range containerResponse.Containers {
		// Update asset location based on the discoverer that
		// we found it from
		container.Location = to.Ptr(discoverer.Spec.NodeName)

		if err = d.enrichContainerInfo(ctx, &container); err != nil {
			logger.WithFields(map[string]interface{}{
				"discoverer":  discoverer.Spec.NodeName,
				"containerId": container.ContainerID,
			}).Warnf("failed to enrich ContainerInfo: %v", err)
		}

		// Convert to asset
		asset := apitypes.AssetType{}
		err = asset.FromContainerInfo(container)
		if err != nil {
			return fmt.Errorf("failed to create AssetType from ContainerInfo: %w", err)
		}

		outputChan <- asset
	}

	return nil
}

func (d *Discoverer) enrichContainerInfo(ctx context.Context, c *apitypes.ContainerInfo) error {
	// Get namespace and Pod name for container
	var ns, podName string
	if c.Labels != nil {
		for _, label := range *c.Labels {
			switch label.Key {
			case PodNamespaceLabel:
				ns = label.Value
			case PodNameLabel:
				podName = label.Value
			}
		}
	}

	if ns == "" || podName == "" {
		return nil
	}

	// Get Pod from K8s API
	pod, err := d.ClientSet.CoreV1().Pods(ns).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get Pod for container. ContainerId=%s: %w", c.ContainerID, err)
	}

	if pod == nil {
		return nil
	}

	// Merge labels set for Pod into container labels where the former has higher precedence
	c.Labels = apitypes.MergeTags(c.Labels, apitypes.MapToTags(pod.Labels))

	return nil
}
