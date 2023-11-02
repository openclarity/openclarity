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

package e2e

import (
	"time"

	"github.com/google/uuid"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

const (
	DefaultScope   string        = "assetInfo/labels/any(t: t/key eq 'scanconfig' and t/value eq 'test')"
	DefaultTimeout time.Duration = 60 * time.Second
)

var FullScanFamiliesConfig = models.ScanFamiliesConfig{
	Exploits: &models.ExploitsConfig{
		Enabled: utils.PointerTo(true),
	},
	InfoFinder: &models.InfoFinderConfig{
		Enabled: utils.PointerTo(true),
	},
	Malware: &models.MalwareConfig{
		Enabled: utils.PointerTo(true),
	},
	Misconfigurations: &models.MisconfigurationsConfig{
		Enabled: utils.PointerTo(true),
	},
	Rootkits: &models.RootkitsConfig{
		Enabled: utils.PointerTo(true),
	},
	Sbom: &models.SBOMConfig{
		Enabled: utils.PointerTo(true),
	},
	Secrets: &models.SecretsConfig{
		Enabled: utils.PointerTo(true),
	},
	Vulnerabilities: &models.VulnerabilitiesConfig{
		Enabled: utils.PointerTo(true),
	},
}

func GetFullScanConfig() models.ScanConfig {
	return GetCustomScanConfig(
		&FullScanFamiliesConfig,
		DefaultScope,
		600, // nolint:gomnd
	)
}

func GetCustomScanConfig(scanFamiliesConfig *models.ScanFamiliesConfig, scope string, timeoutSeconds int) models.ScanConfig {
	return models.ScanConfig{
		Name: utils.PointerTo(uuid.New().String()),
		ScanTemplate: &models.ScanTemplate{
			AssetScanTemplate: &models.AssetScanTemplate{
				ScanFamiliesConfig: scanFamiliesConfig,
			},
			Scope:          utils.PointerTo(scope),
			TimeoutSeconds: utils.PointerTo(timeoutSeconds),
		},
		Scheduled: &models.RuntimeScheduleScanConfig{
			CronLine: utils.PointerTo("0 */4 * * *"),
			OperationTime: utils.PointerTo(
				time.Date(2023, 1, 20, 15, 46, 18, 0, time.UTC),
			),
		},
	}
}

func UpdateScanConfigToStartNow(config *models.ScanConfig) *models.ScanConfig {
	return &models.ScanConfig{
		Name: config.Name,
		ScanTemplate: &models.ScanTemplate{
			AssetScanTemplate: &models.AssetScanTemplate{
				ScanFamiliesConfig: config.ScanTemplate.AssetScanTemplate.ScanFamiliesConfig,
			},
			MaxParallelScanners: config.ScanTemplate.MaxParallelScanners,
			Scope:               config.ScanTemplate.Scope,
			TimeoutSeconds:      config.ScanTemplate.TimeoutSeconds,
		},
		Scheduled: &models.RuntimeScheduleScanConfig{
			CronLine:      config.Scheduled.CronLine,
			OperationTime: utils.PointerTo(time.Now()),
		},
	}
}
