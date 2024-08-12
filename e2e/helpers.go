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
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
)

const (
	DefaultTimeout = 5 * time.Minute
	DefaultPeriod  = 5 * time.Second

	fullScanStartOffset = 5 * time.Second
)

func GetCustomScanConfig(scanFamiliesConfig *apitypes.ScanFamiliesConfig, scope string, timeout time.Duration) apitypes.ScanConfig {
	return apitypes.ScanConfig{
		Name: to.Ptr(uuid.New().String()),
		ScanTemplate: &apitypes.ScanTemplate{
			AssetScanTemplate: &apitypes.AssetScanTemplate{
				ScanFamiliesConfig: scanFamiliesConfig,
			},
			Scope:          to.Ptr(scope),
			TimeoutSeconds: to.Ptr(int(timeout.Seconds())),
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
			OperationTime: to.Ptr(time.Now().Add(fullScanStartOffset)), // ensure it won't become outdated by the time it's used
		},
	}
}

// "database is locked" errors from SQLite should be ignored during the test
// they occur frequently on Azure, but they are temporary and should not cause the test to fail.
func skipDBLockedErr(err error) error {
	if err != nil && strings.Contains(err.Error(), "database is locked") {
		logrus.Info("Database is locked, retrying...")
		return nil
	}

	return err
}
