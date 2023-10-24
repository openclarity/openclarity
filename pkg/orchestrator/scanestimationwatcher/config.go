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

package scanestimationwatcher

import (
	"time"

	"github.com/openclarity/vmclarity/pkg/orchestrator/provider"
	"github.com/openclarity/vmclarity/pkg/shared/backendclient"
)

const (
	DefaultPollInterval             = 5 * time.Second
	DefaultReconcileTimeout         = 2 * time.Minute
	DefaultScanEstimationTimeout    = 48 * time.Hour
	DefaultScanEstimationTTLSeconds = 7200 // // 2 hours
)

type Config struct {
	Backend               *backendclient.BackendClient
	Provider              provider.Provider
	PollPeriod            time.Duration `mapstructure:"poll_period"`
	ReconcileTimeout      time.Duration `mapstructure:"reconcile_timeout"`
	ScanEstimationTimeout time.Duration `mapstructure:"estimation_timeout"`
}

func (c Config) WithBackendClient(b *backendclient.BackendClient) Config {
	c.Backend = b
	return c
}

func (c Config) WithProviderClient(p provider.Provider) Config {
	c.Provider = p
	return c
}
