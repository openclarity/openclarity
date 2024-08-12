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
	"github.com/openclarity/vmclarity/core/to"
)

const (
	DefaultScope   string = "assetInfo/labels/any(t: t/key eq 'scanconfig' and t/value eq 'test')"
	DefaultTimeout        = 60 * time.Second
)

var FullScanFamiliesConfig = apitypes.ScanFamiliesConfig{
	Exploits: &apitypes.ExploitsConfig{
		Enabled: to.Ptr(true),
	},
	InfoFinder: &apitypes.InfoFinderConfig{
		Enabled: to.Ptr(true),
	},
	Malware: &apitypes.MalwareConfig{
		Enabled: to.Ptr(true),
	},
	Misconfigurations: &apitypes.MisconfigurationsConfig{
		Enabled: to.Ptr(true),
	},
	Rootkits: &apitypes.RootkitsConfig{
		Enabled: to.Ptr(true),
	},
	Sbom: &apitypes.SBOMConfig{
		Enabled: to.Ptr(true),
	},
	Secrets: &apitypes.SecretsConfig{
		Enabled: to.Ptr(true),
	},
	Vulnerabilities: &apitypes.VulnerabilitiesConfig{
		Enabled: to.Ptr(true),
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
		Name: to.Ptr(uuid.New().String()),
		ScanTemplate: &apitypes.ScanTemplate{
			AssetScanTemplate: &apitypes.AssetScanTemplate{
				ScanFamiliesConfig: scanFamiliesConfig,
			},
			Scope:          to.Ptr(scope),
			TimeoutSeconds: to.Ptr(timeoutSeconds),
		},
		Scheduled: &apitypes.RuntimeScheduleScanConfig{
			CronLine: to.Ptr("0 */4 * * *"),
			OperationTime: to.Ptr(
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
			OperationTime: to.Ptr(time.Now()),
		},
	}
}
