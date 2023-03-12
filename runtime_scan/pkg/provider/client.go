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

package provider

import (
	"context"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/types"
)

type ScanningJobConfig struct {
	ScannerImage     string // Scanner Container Image to use containing the vmclarity-cli and tools
	ScannerCLIConfig string // Scanner CLI config yaml (families config yaml)
	VMClarityAddress string // The backend address for the scanner CLI to export too
	ScanResultID     string // The ID of the ScanResult that the scanner CLI should update
	KeyPairName      string // The name of the key pair to set on the instance, ignored if not set, used mainly for debugging.
}

type Client interface {
	// RunScanningJob - run a scanning job.
	RunScanningJob(ctx context.Context, region, id string, config ScanningJobConfig) (types.Instance, error)
	// DiscoverScopes - List all scopes
	DiscoverScopes(ctx context.Context) (*models.Scopes, error)
	// DiscoverInstances - list VM instances in the account according to the scan scope.
	DiscoverInstances(ctx context.Context, scanScope *models.ScanScopeType) ([]types.Instance, error)
}
