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
	"fmt"

	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/provider/external/utils"
	provider_service "github.com/openclarity/vmclarity/provider/external/utils/proto"
)

var _ provider.Discoverer = &Discoverer{}

type Discoverer struct {
	Client provider_service.ProviderClient
}

func (d *Discoverer) DiscoverAssets(ctx context.Context) provider.AssetDiscoverer {
	assetDiscoverer := provider.NewSimpleAssetDiscoverer()

	go func() {
		defer close(assetDiscoverer.OutputChan)

		res, err := d.Client.DiscoverAssets(ctx, &provider_service.DiscoverAssetsParams{})
		if err != nil {
			assetDiscoverer.Error = fmt.Errorf("failed to discover assets: %w", err)
			return
		}

		assets := res.GetAssets()
		for _, asset := range assets {
			modelsAsset, err := utils.ConvertAssetToModels(asset)
			if err != nil {
				assetDiscoverer.Error = fmt.Errorf("failed to convert asset to models asset: %w", err)
				return
			}

			select {
			case assetDiscoverer.OutputChan <- *modelsAsset.AssetInfo:
			case <-ctx.Done():
				assetDiscoverer.Error = ctx.Err()
				return
			}
		}
	}()

	return assetDiscoverer
}
