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

package discovery

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider"
	"github.com/openclarity/vmclarity/shared/pkg/backendclient"
)

const (
	discoveryInterval = 2 * time.Minute
)

type ScopeDiscoverer struct {
	backendClient  *backendclient.BackendClient
	providerClient provider.Client
}

func CreateScopeDiscoverer(backendClient *backendclient.BackendClient, providerClient provider.Client) *ScopeDiscoverer {
	return &ScopeDiscoverer{
		backendClient:  backendClient,
		providerClient: providerClient,
	}
}

func (sd *ScopeDiscoverer) Start(ctx context.Context) {
	go func() {
		for {
			log.Debug("Discovering available scopes")
			// nolint:contextcheck
			scopes, err := sd.providerClient.DiscoverScopes(ctx)
			if err != nil {
				log.Warnf("Failed to discover scopes: %v", err)
			} else {
				_, err := sd.backendClient.PutDiscoveryScopes(ctx, scopes)
				if err != nil {
					log.Warnf("Failed to set scopes: %v", err)
				}
			}
			select {
			case <-time.After(discoveryInterval):
				log.Debug("Discovery interval elapsed")
			case <-ctx.Done():
				log.Infof("Stop watching scan configs.")
				return
			}
		}
	}()
}
