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

package external

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/orchestrator/provider"
	provider_service "github.com/openclarity/vmclarity/pkg/orchestrator/provider/external/proto"
)

type Provider struct {
	client provider_service.ProviderClient
	config *Config
	conn   *grpc.ClientConn
}

func New(_ context.Context) (*Provider, error) {
	config, err := NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	err = config.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate configuration: %w", err)
	}

	var opts []grpc.DialOption
	// TODO secure connections
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.Dial(config.ProviderPluginAddress, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial grpc. address=%v: %w", config.ProviderPluginAddress, err)
	}

	return &Provider{
		client: provider_service.NewProviderClient(conn),
		config: config,
		conn:   conn,
	}, nil
}

func (p *Provider) Kind() models.CloudProvider {
	return models.External
}

func (p *Provider) Estimate(ctx context.Context, stats models.AssetScanStats, asset *models.Asset, assetScanTemplate *models.AssetScanTemplate) (*models.Estimation, error) {
	return &models.Estimation{}, provider.FatalErrorf("Not Implemented")
}

func (p *Provider) DiscoverAssets(ctx context.Context) provider.AssetDiscoverer {
	assetDiscoverer := provider.NewSimpleAssetDiscoverer()

	go func() {
		defer close(assetDiscoverer.OutputChan)

		res, err := p.client.DiscoverAssets(ctx, &provider_service.DiscoverAssetsParams{})
		if err != nil {
			assetDiscoverer.Error = fmt.Errorf("failed to discover assets: %w", err)
			return
		}

		assets := res.GetAssets()
		for _, asset := range assets {
			modelsAsset, err := convertAssetToModels(asset)
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

func (p *Provider) RunAssetScan(ctx context.Context, config *provider.ScanJobConfig) error {
	scanJobConfig, err := convertScanJobConfig(config)
	if err != nil {
		return fmt.Errorf("failed to convert scan job config: %w", err)
	}

	res, err := p.client.RunAssetScan(ctx, &provider_service.RunAssetScanParams{
		ScanJobConfig: scanJobConfig,
	})
	if err != nil {
		return provider.FatalErrorf("failed to run asset scan: %v", err)
	}

	if res.Err == nil {
		return provider.FatalErrorf("failed to run asset scan: an error type must be set")
	}

	switch res.Err.ErrorType.(type) {
	case *provider_service.Error_ErrNone:
		return nil
	case *provider_service.Error_ErrRetry:
		retryableErr := res.GetErr().GetErrRetry()
		return provider.RetryableErrorf(time.Second*time.Duration(retryableErr.After), retryableErr.Err)
	case *provider_service.Error_ErrFatal:
		fatalErr := res.GetErr().GetErrFatal()
		return provider.FatalErrorf("failed to run asset scan: %v", fatalErr.Err)
	default:
		return provider.FatalErrorf("failed to run asset scan: error type is not supported: %t", res.Err.GetErrorType())
	}
}

func (p *Provider) RemoveAssetScan(ctx context.Context, config *provider.ScanJobConfig) error {
	scanJobConfig, err := convertScanJobConfig(config)
	if err != nil {
		return fmt.Errorf("failed to convert scan job config: %w", err)
	}

	_, err = p.client.RemoveAssetScan(ctx, &provider_service.RemoveAssetScanParams{
		ScanJobConfig: scanJobConfig,
	})
	if err != nil {
		return fmt.Errorf("failed to remove asset scan: %w", err)
	}
	return nil
}
