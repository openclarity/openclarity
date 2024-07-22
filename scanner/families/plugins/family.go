// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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

package plugins

import (
	"context"
	"fmt"

	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/scanner/families"
	"github.com/openclarity/vmclarity/scanner/families/plugins/runner"
	"github.com/openclarity/vmclarity/scanner/families/plugins/types"
	"github.com/openclarity/vmclarity/scanner/internal/scan_manager"
)

type Plugins struct {
	conf types.Config
}

func New(conf types.Config) families.Family[*types.Result] {
	return &Plugins{
		conf: conf,
	}
}

func (p *Plugins) GetType() families.FamilyType {
	return families.Plugins
}

func (p *Plugins) Run(ctx context.Context, _ *families.Results) (*types.Result, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	// Register plugins dynamically instead of creating a public factory. Users that
	// want to run their own plugin scanners can do so through the config unlike for
	// other families where they actually need the factory to run a custom family scanner.
	factory := scan_manager.NewFactory[types.ScannersConfig, *types.ScannerResult]()
	for _, scannerName := range p.conf.ScannersList {
		factory.Register(scannerName, runner.New)
	}

	// Top level BinaryMode overrides the individual scanner BinaryMode if set
	if p.conf.BinaryMode {
		for name := range p.conf.ScannersConfig {
			config := p.conf.ScannersConfig[name]
			config.BinaryMode = p.conf.BinaryMode
			p.conf.ScannersConfig[name] = config
		}
	}

	// Run all scanner plugins using scan manager
	manager := scan_manager.New(p.conf.ScannersList, p.conf.ScannersConfig, factory)
	scans, err := manager.Scan(ctx, p.conf.Inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to process inputs for plugins: %w", err)
	}

	plugins := types.NewResult()

	// Merge scan results
	for _, scan := range scans {
		logger.Infof("Merging result from %q", scan)

		plugins.Merge(scan.GetScanInputMetadata(len(scan.Result.Findings)), scan.Result)
	}

	return plugins, nil
}
