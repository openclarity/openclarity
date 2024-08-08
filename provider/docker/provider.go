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

package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/provider/docker/discoverer"
	"github.com/openclarity/vmclarity/provider/docker/estimator"
	"github.com/openclarity/vmclarity/provider/docker/scanner"
)

var _ provider.Provider = &Provider{}

type Provider struct {
	*discoverer.Discoverer
	*scanner.Scanner
	*estimator.Estimator
}

func (p *Provider) Kind() apitypes.CloudProvider {
	return apitypes.Docker
}

func New(_ context.Context) (*Provider, error) {
	config, err := NewConfig()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration. Provider=%s: %w", apitypes.Docker, err)
	}

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to load provider configuration. Provider=%s: %w", apitypes.Docker, err)
	}

	return &Provider{
		Discoverer: &discoverer.Discoverer{
			DockerClient: dockerClient,
		},
		Scanner: &scanner.Scanner{
			DockerClient: dockerClient,
			HelperImage:  config.HelperImage,
			NetworkName:  config.NetworkName,
		},
		Estimator: &estimator.Estimator{},
	}, nil
}
