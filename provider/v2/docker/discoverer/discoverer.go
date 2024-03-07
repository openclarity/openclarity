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

	"github.com/docker/docker/client"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/provider"
)

var _ provider.Discoverer = &Discoverer{}

type Discoverer struct {
	DockerClient *client.Client
}

func (d *Discoverer) DiscoverAssets(ctx context.Context) provider.AssetDiscoverer {
	assetDiscoverer := provider.NewSimpleAssetDiscoverer()

	go func() {
		defer close(assetDiscoverer.OutputChan)

		// Get image assets
		imageAssets, err := d.getImageAssets(ctx)
		if err != nil {
			assetDiscoverer.Error = provider.FatalErrorf("failed to get images. Provider=%s: %w", apitypes.Docker, err)
			return
		}

		// Get container assets
		containerAssets, err := d.getContainerAssets(ctx)
		if err != nil {
			assetDiscoverer.Error = provider.FatalErrorf("failed to get containers. Provider=%s: %w", apitypes.Docker, err)
			return
		}

		// Combine assets
		assets := append(imageAssets, containerAssets...)

		for _, asset := range assets {
			select {
			case assetDiscoverer.OutputChan <- asset:
			case <-ctx.Done():
				assetDiscoverer.Error = ctx.Err()
				return
			}
		}
	}()

	return assetDiscoverer
}

func convertTags(tags map[string]string) *[]apitypes.Tag {
	ret := make([]apitypes.Tag, 0, len(tags))
	for key, val := range tags {
		ret = append(ret, apitypes.Tag{
			Key:   key,
			Value: val,
		})
	}
	return &ret
}
