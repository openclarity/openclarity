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

package families

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/shared/pkg/families/exploits"
	"github.com/openclarity/vmclarity/shared/pkg/families/interfaces"
	"github.com/openclarity/vmclarity/shared/pkg/families/malware"
	"github.com/openclarity/vmclarity/shared/pkg/families/misconfiguration"
	"github.com/openclarity/vmclarity/shared/pkg/families/results"
	"github.com/openclarity/vmclarity/shared/pkg/families/rootkits"
	"github.com/openclarity/vmclarity/shared/pkg/families/sbom"
	"github.com/openclarity/vmclarity/shared/pkg/families/secrets"
	"github.com/openclarity/vmclarity/shared/pkg/families/types"
	"github.com/openclarity/vmclarity/shared/pkg/families/vulnerabilities"
)

type Manager struct {
	config   *Config
	families []interfaces.Family
}

func New(logger *log.Entry, config *Config) *Manager {
	manager := &Manager{
		config: config,
	}

	// Analyzers.
	// SBOM MUST come before vulnerabilities.
	if config.SBOM.Enabled {
		manager.families = append(manager.families, sbom.New(logger, config.SBOM))
	}

	// Scanners.
	// Vulnerabilities MUST be after SBOM to support the case it is configured to use the output from sbom.
	if config.Vulnerabilities.Enabled {
		manager.families = append(manager.families, vulnerabilities.New(logger, config.Vulnerabilities))
	}
	if config.Secrets.Enabled {
		manager.families = append(manager.families, secrets.New(logger, config.Secrets))
	}
	if config.Rootkits.Enabled {
		manager.families = append(manager.families, rootkits.New(logger, config.Rootkits))
	}
	if config.Malware.Enabled {
		manager.families = append(manager.families, malware.New(logger, config.Malware))
	}
	if config.Misconfiguration.Enabled {
		manager.families = append(manager.families, misconfiguration.New(logger, config.Misconfiguration))
	}

	// Enrichers.
	// Exploits MUST be after Vulnerabilities to support the case it is configured to use the output from Vulnerabilities.
	if config.Exploits.Enabled {
		manager.families = append(manager.families, exploits.New(logger, config.Exploits))
	}

	return manager
}

type RunErrors map[types.FamilyType]error

func (m *Manager) Run() (*results.Results, RunErrors) {
	familyErrors := make(map[types.FamilyType]error)

	familyResults := results.New()

	for _, family := range m.families {
		ret, err := family.Run(familyResults)
		if err != nil {
			familyErrors[family.GetType()] = fmt.Errorf("failed to run family %v: %w", family.GetType(), err)
		} else {
			familyResults.SetResults(ret)
		}
	}

	return familyResults, familyErrors
}
