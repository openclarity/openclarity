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

package scanresultwatcher

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

type jobConfigInput struct {
	config     *ScannerConfig
	scanResult *models.TargetScanResult
	scanConfig *models.ScanConfigSnapshot
	target     *models.Target
}

func (i *jobConfigInput) Validate() error {
	if i.config == nil {
		return errors.New("invalid JobConfigInput as ScannerConfig is nil")
	}

	if i.scanResult == nil {
		return errors.New("invalid JobConfigInput as ScanResult is nil")
	}

	if i.scanConfig == nil {
		return errors.New("invalid JobConfigInput as ScanConfig is nil")
	}

	if i.target == nil {
		return errors.New("invalid JobConfigInput as Target is nil")
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
	if i.scanConfig.ScannerInstanceCreationConfig != nil {
		instanceCreationConfig = *i.scanConfig.ScannerInstanceCreationConfig
	}

	scannerConfig := NewFamiliesConfigFrom(i.config, i.scanConfig)
	scannerConfigYAML, err := yaml.Marshal(scannerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to convert ScannerConfig to YAML for ScanResult with %s id",
			*i.scanResult.Id)
	}

	return &provider.ScanJobConfig{
		ScannerImage:     i.config.ScannerImage,
		ScannerCLIConfig: string(scannerConfigYAML),
		VMClarityAddress: i.config.ScannerBackendAddress,
		KeyPairName:      i.config.ScannerKeyPairName,
		ScannerRegion:    i.config.Region,
		BlockDeviceName:  i.config.DeviceName,
		ScanMetadata: provider.ScanMetadata{
			ScanID:       i.scanResult.Scan.Id,
			ScanResultID: *i.scanResult.Id,
			TargetID:     i.scanResult.Target.Id,
		},
		ScannerInstanceCreationConfig: instanceCreationConfig,
		Target:                        *i.target,
	}, nil
}
