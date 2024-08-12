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
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/orchestrator/provider"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

type jobConfigInput struct {
	config    *ScannerConfig
	assetScan *models.AssetScan
	asset     *models.Asset
}

func (i *jobConfigInput) Validate() error {
	if i.config == nil {
		return errors.New("invalid JobConfigInput as ScannerConfig is nil")
	}

	if i.assetScan == nil {
		return errors.New("invalid JobConfigInput as AssetScan is nil")
	}

	if i.asset == nil {
		return errors.New("invalid JobConfigInput as Asset is nil")
	}

	return nil
}

func newJobConfig(i *jobConfigInput) (*provider.ScanJobConfig, error) {
	if err := i.Validate(); err != nil {
		return nil, fmt.Errorf("faield to create JobConfig: %w", err)
	}

	instanceCreationConfig := models.ScannerInstanceCreationConfig{
		MaxPrice:         nil,
		RetryMaxAttempts: utils.PointerTo(1),
		UseSpotInstances: false,
	}
	if i.assetScan.ScannerInstanceCreationConfig != nil {
		instanceCreationConfig = *i.assetScan.ScannerInstanceCreationConfig
	}

	scannerConfig := NewFamiliesConfigFrom(i.config, i.assetScan.ScanFamiliesConfig)
	scannerConfigYAML, err := yaml.Marshal(scannerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to convert ScannerConfig to YAML for AssetScan with %s id",
			*i.assetScan.Id)
	}

	return &provider.ScanJobConfig{
		ScannerImage:     i.config.ScannerImage,
		ScannerCLIConfig: string(scannerConfigYAML),
		VMClarityAddress: i.config.APIServerAddress,
		ScanMetadata: provider.ScanMetadata{
			AssetScanID: *i.assetScan.Id,
			AssetID:     i.assetScan.Asset.Id,
		},
		ScannerInstanceCreationConfig: instanceCreationConfig,
		Asset:                         *i.asset,
	}, nil
}
