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

package assetscanwatcher

import (
	"time"

	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider"
	"github.com/openclarity/vmclarity/shared/pkg/backendclient"
)

const (
	DefaultPollInterval     = time.Minute
	DefaultReconcileTimeout = 5 * time.Minute
	DefaultAbortTimeout     = 10 * time.Minute
)

type Config struct {
	Backend          *backendclient.BackendClient
	Provider         provider.Provider
	PollPeriod       time.Duration
	ReconcileTimeout time.Duration
	ScannerConfig    ScannerConfig
	AbortTimeout     time.Duration
}

func (c Config) WithBackendClient(b *backendclient.BackendClient) Config {
	c.Backend = b
	return c
}

func (c Config) WithProviderClient(p provider.Provider) Config {
	c.Provider = p
	return c
}

func (c Config) WithReconcileTimeout(t time.Duration) Config {
	c.ReconcileTimeout = t
	return c
}

func (c Config) WithPollPeriod(t time.Duration) Config {
	c.PollPeriod = t
	return c
}

func (c Config) WithScannerConfig(s ScannerConfig) Config {
	c.ScannerConfig = s
	return c
}

type ScannerConfig struct {
	// Address that the Scanner should use to talk to the VMClarity backend
	// We use a configuration variable for this instead of discovering it
	// automatically in case VMClarity backend has multiple IPs (internal
	// traffic and external traffic for example) so we need the specific
	// address to use.
	ScannerBackendAddress string

	ExploitsDBAddress string

	TrivyServerAddress string
	TrivyServerTimeout time.Duration

	GrypeServerAddress string
	GrypeServerTimeout time.Duration

	DeleteJobPolicy DeleteJobPolicyType

	// The container image to use once we've booted the scanner virtual
	// machine, that contains the VMClarity CLI plus all the required
	// tools.
	ScannerImage string

	// The gitleaks binary path in the scanner image container.
	GitleaksBinaryPath string

	// The clam binary path in the scanner image container.
	ClamBinaryPath string

	// The freshclam binary path in the scanner image container
	FreshclamBinaryPath string

	// The freshclam mirror url to use if it's enabled
	AlternativeFreshclamMirrorURL string

	// The location where Lynis is installed in the scanner image
	LynisInstallPath string

	// The chkrootkit binary path in the scanner image container.
	ChkrootkitBinaryPath string
}
