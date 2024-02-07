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

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/cli/pkg/utils"
)

const (
	DefaultScope   string = "assetInfo/labels/any(t: t/key eq 'scanconfig' and t/value eq 'test')"
	DefaultTimeout        = 60 * time.Second
)

var FullScanFamiliesConfig = apitypes.ScanFamiliesConfig{
	Exploits: &apitypes.ExploitsConfig{
		Enabled: utils.PointerTo(true),
	},
	InfoFinder: &apitypes.InfoFinderConfig{
		Enabled: utils.PointerTo(true),
	},
	Malware: &apitypes.MalwareConfig{
		Enabled: utils.PointerTo(true),
	},
	Misconfigurations: &apitypes.MisconfigurationsConfig{
		Enabled: utils.PointerTo(true),
	},
	Rootkits: &apitypes.RootkitsConfig{
		Enabled: utils.PointerTo(true),
	},
	Sbom: &apitypes.SBOMConfig{
		Enabled: utils.PointerTo(true),
	},
	Secrets: &apitypes.SecretsConfig{
		Enabled: utils.PointerTo(true),
	},
	Vulnerabilities: &apitypes.VulnerabilitiesConfig{
		Enabled: utils.PointerTo(true),
	},
}

func GetFullScanConfig() apitypes.ScanConfig {
	return GetCustomScanConfig(
		&FullScanFamiliesConfig,
		DefaultScope,
		600, // nolint:gomnd
	)
}

func GetCustomScanConfig(scanFamiliesConfig *apitypes.ScanFamiliesConfig, scope string, timeoutSeconds int) apitypes.ScanConfig {
	return apitypes.ScanConfig{
		Name: utils.PointerTo(uuid.New().String()),
		ScanTemplate: &apitypes.ScanTemplate{
			AssetScanTemplate: &apitypes.AssetScanTemplate{
				ScanFamiliesConfig: scanFamiliesConfig,
			},
			Scope:          utils.PointerTo(scope),
			TimeoutSeconds: utils.PointerTo(timeoutSeconds),
		},
		Scheduled: &apitypes.RuntimeScheduleScanConfig{
			CronLine: utils.PointerTo("0 */4 * * *"),
			OperationTime: utils.PointerTo(
				time.Date(2023, 1, 20, 15, 46, 18, 0, time.UTC),
			),
		},
	}
}

func UpdateScanConfigToStartNow(config *apitypes.ScanConfig) *apitypes.ScanConfig {
	return &apitypes.ScanConfig{
		Name: config.Name,
		ScanTemplate: &apitypes.ScanTemplate{
			AssetScanTemplate: &apitypes.AssetScanTemplate{
				ScanFamiliesConfig: config.ScanTemplate.AssetScanTemplate.ScanFamiliesConfig,
			},
			MaxParallelScanners: config.ScanTemplate.MaxParallelScanners,
			Scope:               config.ScanTemplate.Scope,
			TimeoutSeconds:      config.ScanTemplate.TimeoutSeconds,
		},
		Scheduled: &apitypes.RuntimeScheduleScanConfig{
			CronLine:      config.Scheduled.CronLine,
			OperationTime: utils.PointerTo(time.Now()),
		},
	}
}
