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

package types

type FamilyConfigEnabler interface {
	IsEnabled() bool
}

func (c *VulnerabilitiesConfig) IsEnabled() bool {
	return c != nil && c.Enabled != nil && *c.Enabled
}

func (c *VulnerabilitiesConfig) GetScannersList() []string {
	if c != nil && c.Scanners != nil && len(*c.Scanners) != 0 {
		return *c.Scanners
	}

	return []string{"grype", "trivy"}
}

func (c *SecretsConfig) IsEnabled() bool {
	return c != nil && c.Enabled != nil && *c.Enabled
}

func (c *SecretsConfig) GetScannersList() []string {
	if c != nil && c.Scanners != nil && len(*c.Scanners) != 0 {
		return *c.Scanners
	}

	return []string{"gitleaks"}
}

func (c *SBOMConfig) IsEnabled() bool {
	return c != nil && c.Enabled != nil && *c.Enabled
}

func (c *SBOMConfig) GetAnalyzersList() []string {
	if c != nil && c.Analyzers != nil && len(*c.Analyzers) != 0 {
		return *c.Analyzers
	}

	return []string{"syft", "trivy", "windows"}
}

func (c *RootkitsConfig) IsEnabled() bool {
	return c != nil && c.Enabled != nil && *c.Enabled
}

func (c *RootkitsConfig) GetScannersList() []string {
	if c != nil && c.Scanners != nil && len(*c.Scanners) != 0 {
		return *c.Scanners
	}

	return []string{"chkrootkit"}
}

func (c *MisconfigurationsConfig) IsEnabled() bool {
	return c != nil && c.Enabled != nil && *c.Enabled
}

func (c *MisconfigurationsConfig) GetScannersList() []string {
	if c != nil && c.Scanners != nil && len(*c.Scanners) != 0 {
		return *c.Scanners
	}

	return []string{"lynis", "cisdocker"}
}

func (c *MalwareConfig) IsEnabled() bool {
	return c != nil && c.Enabled != nil && *c.Enabled
}

func (c *MalwareConfig) GetScannersList() []string {
	if c != nil && c.Scanners != nil && len(*c.Scanners) != 0 {
		return *c.Scanners
	}

	return []string{"clam", "yara"}
}

func (c *MalwareConfig) GetYaraDirectoriesToScan() []string {
	if c != nil && c.YaraDirectoriesToScan != nil && len(*c.YaraDirectoriesToScan) != 0 {
		return *c.YaraDirectoriesToScan
	}

	return []string{"/home", "/opt", "/tmp", "/var/log"}
}

func (c *ExploitsConfig) IsEnabled() bool {
	return c != nil && c.Enabled != nil && *c.Enabled
}

func (c *ExploitsConfig) GetScannersList() []string {
	if c != nil && c.Scanners != nil && len(*c.Scanners) != 0 {
		return *c.Scanners
	}

	return []string{"exploitdb"}
}

func (c *InfoFinderConfig) IsEnabled() bool {
	return c != nil && c.Enabled != nil && *c.Enabled
}

func (c *InfoFinderConfig) GetScannersList() []string {
	if c != nil && c.Scanners != nil && len(*c.Scanners) != 0 {
		return *c.Scanners
	}

	return []string{"sshTopology"}
}

func (c *PluginsConfig) IsEnabled() bool {
	return c != nil && c.Enabled != nil && *c.Enabled
}
