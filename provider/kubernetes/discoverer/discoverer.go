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
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	"github.com/openclarity/vmclarity/provider"
)

var _ provider.Discoverer = &Discoverer{}

type Discoverer struct {
	ClientSet                          kubernetes.Interface
	ContainerRuntimeDiscoveryNamespace string
	CrDiscovererLabels                 map[string]string
}

func (d *Discoverer) DiscoverAssets(ctx context.Context) provider.AssetDiscoverer {
	assetDiscoverer := provider.NewSimpleAssetDiscoverer()

	go func() {
		defer close(assetDiscoverer.OutputChan)

		discoverers, err := d.ClientSet.CoreV1().Pods(d.ContainerRuntimeDiscoveryNamespace).List(ctx, metav1.ListOptions{
			LabelSelector: labels.Set(d.CrDiscovererLabels).String(),
		})
		if err != nil {
			assetDiscoverer.Error = fmt.Errorf("unable to list discoverers: %w", err)
			return
		}

		var errs []error

		err = d.discoverImages(ctx, assetDiscoverer.OutputChan, discoverers.Items)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to discover images: %w", err))
		}

		err = d.discoverContainers(ctx, assetDiscoverer.OutputChan, discoverers.Items)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to discover containers: %w", err))
		}

		assetDiscoverer.Error = errors.Join(errs...)
	}()

	return assetDiscoverer
}
