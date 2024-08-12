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

package azure

import (
	"context"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/provider/v2/aws/discoverer"
	"github.com/openclarity/vmclarity/provider/v2/aws/estimator"
	"github.com/openclarity/vmclarity/provider/v2/aws/scanner"
)

var _ provider.Provider = &Provider{}

type Provider struct {
	*discoverer.Discoverer
	*scanner.Scanner
	*estimator.Estimator
}

func (p *Provider) Kind() apitypes.CloudProvider {
	// TODO implement me
	panic("implement me")
}

func New(_ context.Context) (*Provider, error) {
	return &Provider{
		Discoverer: &discoverer.Discoverer{},
		Scanner:    &scanner.Scanner{},
		Estimator:  &estimator.Estimator{},
	}, nil
}
