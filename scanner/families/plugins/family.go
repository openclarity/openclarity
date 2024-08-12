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
	"time"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/log"
	plugintypes "github.com/openclarity/vmclarity/plugins/sdk-go/types"
	"github.com/openclarity/vmclarity/scanner/families/interfaces"
	"github.com/openclarity/vmclarity/scanner/families/plugins/common"
	"github.com/openclarity/vmclarity/scanner/families/plugins/runner"
	"github.com/openclarity/vmclarity/scanner/families/results"
	"github.com/openclarity/vmclarity/scanner/families/types"
	familiesutils "github.com/openclarity/vmclarity/scanner/families/utils"
	"github.com/openclarity/vmclarity/scanner/job_manager"
	"github.com/openclarity/vmclarity/scanner/utils"
)

type Plugins struct {
	conf Config
}

var _ interfaces.Family = &Plugins{}

func (p *Plugins) Run(ctx context.Context, res *results.Results) (interfaces.IsResults, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithField("family", "plugins")
	logger.Info("Plugins Run...")

	factory := job_manager.NewJobFactory()
	for _, n := range p.conf.ScannersList {
		factory.Register(n, runner.New)
	}

	manager := job_manager.New(p.conf.ScannersList, p.conf.ScannersConfig, logger, factory)

	var pluginsResults Results
	for _, input := range p.conf.Inputs {
		startTime := time.Now()
		managerResults, err := manager.Run(ctx, utils.SourceType(input.InputType), input.Input)
		if err != nil {
			return nil, fmt.Errorf("failed to scan input %q for plugins: %w", input.Input, err)
		}
		endTime := time.Now()
		inputSize, err := familiesutils.GetInputSize(input)
		if err != nil {
			logger.Warnf("Failed to calculate input %v size: %v", input, err)
		}

		// Merge results from all plugins into the same output
		var mergedResults []apitypes.FindingInfo
		mergedPluginResult := make(map[string]plugintypes.Result)
		for name, result := range managerResults {
			logger.Infof("Merging result from %q", name)
			mergedResults = append(mergedResults, result.(*common.Results).Findings...) //nolint:forcetypeassert
			mergedPluginResult[name] = *result.(*common.Results).Output                 //nolint:forcetypeassert
		}

		pluginsResults.Findings = mergedResults
		pluginsResults.PluginOutputs = mergedPluginResult
		pluginsResults.Metadata.InputScans = append(pluginsResults.Metadata.InputScans, types.CreateInputScanMetadata(startTime, endTime, inputSize, input))
	}

	logger.Info("Plugins Done...")
	return &pluginsResults, nil
}

func (p *Plugins) GetType() types.FamilyType {
	return types.Plugins
}

func New(conf Config) *Plugins {
	return &Plugins{
		conf: conf,
	}
}
