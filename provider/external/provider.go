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

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/provider/external/discoverer"
	"github.com/openclarity/vmclarity/provider/external/estimator"
	"github.com/openclarity/vmclarity/provider/external/scanner"
	provider_service "github.com/openclarity/vmclarity/provider/external/utils/proto"
)

var _ provider.Provider = &Provider{}

type Provider struct {
	*discoverer.Discoverer
	*scanner.Scanner
	*estimator.Estimator
}

func (p *Provider) Kind() apitypes.CloudProvider {
	return apitypes.External
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

	conn, err := grpc.NewClient(config.ProviderPluginAddress, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial grpc. address=%v: %w", config.ProviderPluginAddress, err)
	}

	client := provider_service.NewProviderClient(conn)

	return &Provider{
		Discoverer: &discoverer.Discoverer{
			Client: client,
		},
		Scanner: &scanner.Scanner{
			Client: client,
		},
		Estimator: &estimator.Estimator{},
	}, nil
}
