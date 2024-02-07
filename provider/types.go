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

	apitypes "github.com/openclarity/vmclarity/api/types"
)

type Provider interface {
	// Kind returns apitypes.CloudProvider
	Kind() apitypes.CloudProvider

	Discoverer
	Estimator
	Scanner
}

type Discoverer interface {
	// DiscoverAssets returns AssetDiscoverer.
	// It is the responsibility of DiscoverAssets to feed the AssetDiscoverer channel with assets, and to close the channel when done.
	// In addition, any error should be reported on the AssetDiscoverer error part.
	DiscoverAssets(ctx context.Context) AssetDiscoverer
}

// Estimator estimates asset scans cost, time and size without running any asset scan.
type Estimator interface {
	// Estimate returns Estimation containing asset scan estimation data according to the AssetScanTemplate.
	// The cost estimation takes into account all the resources that are being used during a scan, and includes a detailed bill of the cost of each resource.
	// If AssetScanTemplate contains several input paths to scan, the cost will be calculated for each input and will be summed up to a total cost.
	// When exists, the scan size and scan time will be taken from the AssetScanStats. Otherwise, they will be estimated base on lab tests and Asset volume size.
	Estimate(context.Context, apitypes.AssetScanStats, *apitypes.Asset, *apitypes.AssetScanTemplate) (*apitypes.Estimation, error)
}

type Scanner interface {
	// RunAssetScan is a non-blocking call which takes a ScanJobConfig and creates resources for performing Scan.
	// It may return FatalError or RetryableError to indicate if the error is permanent or transient.
	// It is expected to return RetryableError in case the resources are not ready or are still being created.
	// It must return nil if all the resources are created and ready.
	// It also must be idempotent.
	RunAssetScan(context.Context, *ScanJobConfig) error
	// RemoveAssetScan is a non-blocking call which takes a ScanJobConfig and remove resources created for Scan.
	// It may return FatalError or RetryableError to indicate if the error is permanent or transient.
	// It is expected to return RetryableError in case the resources are still being deleted.
	// It must return nil if all the resources are deleted.
	// It also must be idempotent.
	RemoveAssetScan(context.Context, *ScanJobConfig) error
}

type ScanMetadata struct {
	ScanID      string
	AssetScanID string
	AssetID     string
}

type ScanJobConfig struct {
	ScannerImage     string // Scanner Container Image to use containing the vmclarity-cli and tools
	ScannerCLIConfig string // Scanner CLI config yaml (families config yaml)
	VMClarityAddress string // The backend address for the scanner CLI to export too

	ScanMetadata
	apitypes.ScannerInstanceCreationConfig
	apitypes.Asset
}

// AssetDiscoverer is used to discover assets in a buffered manner.
type AssetDiscoverer interface {
	Chan() chan apitypes.AssetType
	Err() error
}

type SimpleAssetDiscoverer struct {
	OutputChan chan apitypes.AssetType
	Error      error
}

func (ad *SimpleAssetDiscoverer) Chan() chan apitypes.AssetType {
	return ad.OutputChan
}

func (ad *SimpleAssetDiscoverer) Err() error {
	return ad.Error
}

func NewSimpleAssetDiscoverer() *SimpleAssetDiscoverer {
	output := make(chan apitypes.AssetType)
	return &SimpleAssetDiscoverer{
		OutputChan: output,
		Error:      nil,
	}
}
