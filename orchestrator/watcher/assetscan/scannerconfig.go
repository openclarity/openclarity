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

package assetscan

import "time"

const (
	DefaultTrivyScanTimeout   = 5 * time.Minute
	DefaultGrypeServerTimeout = 2 * time.Minute
)

type ScannerConfig struct {
	// Address that the Scanner should use to talk to the VMClarity backend
	// We use a configuration variable for this instead of discovering it
	// automatically in case VMClarity backend has multiple IPs (internal
	// traffic and external traffic for example) so we need the specific
	// address to use.
	APIServerAddress string `json:"apiserver-address,omitempty" mapstructure:"apiserver_address"`

	ExploitsDBAddress string `mapstructure:"exploitsdb_address"`

	TrivyServerAddress string        `mapstructure:"trivy_server_address"`
	TrivyScanTimeout   time.Duration `mapstructure:"trivy_scan_timeout"`

	GrypeServerAddress string        `mapstructure:"grype_server_address"`
	GrypeServerTimeout time.Duration `mapstructure:"grype_server_timeout"`

	YaraRuleServerAddress string `mapstructure:"yara_rule_server_address"`

	// The container image to use once we've booted the scanner virtual
	// machine, that contains the VMClarity CLI plus all the required
	// tools.
	ScannerImage string `mapstructure:"container_image"`

	// The freshclam mirror url to use if it's enabled
	AlternativeFreshclamMirrorURL string `mapstructure:"freshclam_mirror"`
}
