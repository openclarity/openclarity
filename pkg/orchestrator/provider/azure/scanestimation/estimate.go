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

package scanestimation

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/orchestrator/provider/common"
	familiestypes "github.com/openclarity/vmclarity/pkg/shared/families/types"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

type MarketOption string

const (
	MarketOptionSpot     MarketOption = "Spot"
	MarketOptionOnDemand MarketOption = "OnDemand"
)

type ScanEstimator struct {
	priceFetcher PriceFetcherImpl
}

type EstimateAssetScanParams struct {
	SourceRegion               string
	DestRegion                 string
	OSDiskStorageAccountType   armcompute.DiskStorageAccountTypes
	DataDiskStorageAccountType armcompute.DiskStorageAccountTypes
	StorageAccountType         armcompute.StorageAccountTypes
	ScannerVMSize              armcompute.VirtualMachineSizeTypes
	ScannerOSDiskSizeGB        int64
	Stats                      models.AssetScanStats
	Asset                      *models.Asset
	AssetScanTemplate          *models.AssetScanTemplate
}

const scanEstimatorTimeout = time.Second * 30

func New() *ScanEstimator {
	netClient := &http.Client{
		Timeout: scanEstimatorTimeout,
	}

	return &ScanEstimator{
		priceFetcher: PriceFetcherImpl{client: netClient},
	}
}

const (
	SecondsInAMonth = 86400 * 30
	SecondsInAnHour = 60 * 60
)

type recipeResource string

const (
	Snapshot      recipeResource = "Snapshot"
	ScannerVM     recipeResource = "ScannerVM"
	DataDisk      recipeResource = "DataDisk"
	ScannerOSDisk recipeResource = "ScannerOSDisk"
	DataTransfer  recipeResource = "DataTransfer"
	BlobStorage   recipeResource = "BlobStorage"
)

// scanSizesGB represents the disk used memory size on the machines that the tests were taken on.
// 1652MB and 4559MB are the sizes that were tests.
var scanSizesGB = []float64{0.01, 1.652, 4.559}

// jobCreationTime the time that took the job move from Scheduled state to InProgress.
var jobCreationTime = common.MustLogarithmicFit(scanSizesGB, []float64{0.01, 1860, 2460})

// familyScanDurationsMap Calculate the logarithmic fit of each family base on static measurements of family scan duration in seconds per scanSizesGB value.
// The tests were made on a Standard_D2s_v3 virtual machine with Standard SSD LRS os disk (30 GB)
// The times correspond to the scan size values in scanSizesGB.
// TODO add infoFinder family stats.
// nolint:gomnd
var familyScanDurationsMap = map[familiestypes.FamilyType]*common.LogarithmicFormula{
	familiestypes.SBOM:             common.MustLogarithmicFit(scanSizesGB, []float64{0.01, 16, 17}),
	familiestypes.Vulnerabilities:  common.MustLogarithmicFit(scanSizesGB, []float64{0.01, 4, 10}), // TODO check time with no sbom scan
	familiestypes.Secrets:          common.MustLogarithmicFit(scanSizesGB, []float64{0.01, 420, 780}),
	familiestypes.Exploits:         common.MustLogarithmicFit(scanSizesGB, []float64{0, 0, 0}),
	familiestypes.Rootkits:         common.MustLogarithmicFit(scanSizesGB, []float64{0, 0, 0}),
	familiestypes.Misconfiguration: common.MustLogarithmicFit(scanSizesGB, []float64{0.01, 6, 7}),
	familiestypes.Malware:          common.MustLogarithmicFit(scanSizesGB, []float64{0.01, 900, 1140}),
}

const two = 2

// nolint:cyclop
func (s *ScanEstimator) EstimateAssetScan(ctx context.Context, params EstimateAssetScanParams) (*models.Estimation, error) {
	var snapshotFromBlobMonthlyCost float64
	var err error

	if params.AssetScanTemplate == nil || params.AssetScanTemplate.ScanFamiliesConfig == nil {
		return nil, fmt.Errorf("scan families config was not provided")
	}
	familiesConfig := params.AssetScanTemplate.ScanFamiliesConfig

	// Get scan size and scan duration using previous stats and asset info.
	scanSizeMB, err := common.GetScanSize(params.Stats, params.Asset)
	if err != nil {
		return nil, fmt.Errorf("failed to get scan size: %w", err)
	}
	scanSizeGB := float64(scanSizeMB) / common.MBInGB
	scanDurationSec := common.GetScanDuration(params.Stats, familiesConfig, scanSizeMB, familyScanDurationsMap)

	marketOption := MarketOptionOnDemand
	if params.AssetScanTemplate.UseSpotInstances() {
		marketOption = MarketOptionSpot
	}

	sourceRegion := params.SourceRegion
	destRegion := params.DestRegion
	scannerVMSize := params.ScannerVMSize
	scannerOSDiskSizeGB := params.ScannerOSDiskSizeGB
	scannerOSDiskType := params.OSDiskStorageAccountType
	// The data disk that is created from the snapshot.
	dataDiskType := params.DataDiskStorageAccountType
	// jobCreationTimeSec is the time between scan state moves from Scheduled to inProgress.
	jobCreationTimeSec := jobCreationTime.Evaluate(scanSizeGB)
	// snapshotFromOSDiskIdleRunTime is the approximate amount of time that a resource is up before the scan starts (during job creation time).
	// The order in which resources are created in Azure:
	// 1. snapshot from target OS disk
	// 2. Blob storage (for different regions)
	// 3. Snapshot from blob (for different regions)
	// 4. data disk from snapshot
	// 5. Scanner VM
	// The idle run time will be given to each resource accordingly:
	snapshotFromOSDiskIdleRunTime := jobCreationTimeSec
	blobStorageIdleRunTime := jobCreationTimeSec * 0.8
	snapshotFromBlobIdleRunTime := jobCreationTimeSec * 0.6
	diskFromSnapshotIdleRunTime := jobCreationTimeSec * 0.4
	scannerVMIdleRunTime := jobCreationTimeSec * 0.2

	// The SKU for the snapshot is automatically chosen by Azure based on the source disk SKU.
	// We need to convert from the Disk SKU to the snapshot SKU to get the correct pricing.
	snapshotStorageAccountType, err := getSnapshotTypeFromDiskType(params.OSDiskStorageAccountType)
	if err != nil {
		return nil, err
	}

	// Get relevant current prices from Azure price list API
	// Fetch the snapshot from os disk monthly cost.
	snapshotFromOSDiskMonthlyCost, err := s.priceFetcher.GetSnapshotGBPerMonthCost(ctx, sourceRegion, snapshotStorageAccountType)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly cost for destination snapshot: %w", err)
	}

	// Fetch the scanner vm hourly cost.
	scannerPerHourCost, err := s.priceFetcher.GetInstancePerHourCost(ctx, destRegion, scannerVMSize, marketOption)
	if err != nil {
		return nil, fmt.Errorf("failed to get scanner per hour cost: %w", err)
	}

	// Fetch the scanner os disk monthly cost.
	scannerOSDiskMonthlyCost, err := s.priceFetcher.GetManagedDiskMonthlyCost(ctx, destRegion, scannerOSDiskType, scannerOSDiskSizeGB)
	if err != nil {
		return nil, fmt.Errorf("failed to get os disk monthly cost per GB: %w", err)
	}

	// Fetch the monthly cost of the disk that was created from the snapshot.
	// We assume that the data disk type will be the same as the os disk size.
	diskFromSnapshotMonthlyCost, err := s.priceFetcher.GetManagedDiskMonthlyCost(ctx, destRegion, dataDiskType, scannerOSDiskSizeGB)
	if err != nil {
		return nil, fmt.Errorf("failed to get data disk monthly cost per GB: %w", err)
	}

	dataTransferCost := 0.0
	snapshotFromBlobCost := 0.0
	blobStorageCost := 0.0
	if sourceRegion != destRegion {
		// if the scanner is in a different region than the scanned asset, we have another snapshot created in the
		// dest region.
		snapshotFromBlobMonthlyCost, err = s.priceFetcher.GetSnapshotGBPerMonthCost(ctx, destRegion, snapshotStorageAccountType)
		if err != nil {
			return nil, fmt.Errorf("failed to get dest snapshot monthly cost: %w", err)
		}

		snapshotFromBlobCost = snapshotFromBlobMonthlyCost * ((snapshotFromBlobIdleRunTime + float64(scanDurationSec)) / SecondsInAMonth) * scanSizeGB

		// Fetch the data transfer cost per GB (if source and dest regions are the same, this will be 0).
		dataTransferCostPerGB, err := s.priceFetcher.GetDataTransferPerGBCost(ctx, sourceRegion)
		if err != nil {
			return nil, fmt.Errorf("failed to get data transfer cost per GB: %w", err)
		}

		dataTransferCost = dataTransferCostPerGB * scanSizeGB

		// when moving a snapshot into another region, the snapshot is copied into a blob storage.
		blobStoragePerGB, err := s.priceFetcher.GetBlobStoragePerGBCost(ctx, destRegion, params.StorageAccountType)
		if err != nil {
			return nil, fmt.Errorf("failed to get blob storage cost per GB: %w", err)
		}

		blobStorageCost = blobStoragePerGB * ((blobStorageIdleRunTime + float64(scanDurationSec)) / SecondsInAMonth) * scanSizeGB
	}

	snapshotFromOSDiskCost := snapshotFromOSDiskMonthlyCost * ((snapshotFromOSDiskIdleRunTime + float64(scanDurationSec)) / SecondsInAMonth) * scanSizeGB
	diskFromSnapshotCost := diskFromSnapshotMonthlyCost * ((diskFromSnapshotIdleRunTime + float64(scanDurationSec)) / SecondsInAMonth) * scanSizeGB
	scannerCost := scannerPerHourCost * ((scannerVMIdleRunTime + float64(scanDurationSec)) / SecondsInAnHour)
	scannerOSDiskCost := scannerOSDiskMonthlyCost * ((scannerVMIdleRunTime + float64(scanDurationSec)) / SecondsInAMonth) * float64(scannerOSDiskSizeGB)

	jobTotalCost := snapshotFromBlobCost + diskFromSnapshotCost + scannerCost + scannerOSDiskCost + dataTransferCost + snapshotFromOSDiskCost + blobStorageCost

	// Create the Estimation object base on the calculated data.
	costBreakdown := []models.CostBreakdownComponent{
		{
			Cost:      float32(snapshotFromOSDiskCost),
			Operation: fmt.Sprintf("%v-%v", Snapshot, destRegion),
		},
		{
			Cost:      float32(scannerCost),
			Operation: string(ScannerVM),
		},
		{
			Cost:      float32(diskFromSnapshotCost),
			Operation: string(DataDisk),
		},
		{
			Cost:      float32(scannerOSDiskCost),
			Operation: string(ScannerOSDisk),
		},
	}
	if sourceRegion != destRegion {
		costBreakdown = append(costBreakdown, []models.CostBreakdownComponent{
			{
				Cost:      float32(snapshotFromBlobCost),
				Operation: fmt.Sprintf("%v-%v", Snapshot, sourceRegion),
			},
			{
				Cost:      float32(dataTransferCost),
				Operation: string(DataTransfer),
			},
			{
				Cost:      float32(blobStorageCost),
				Operation: string(BlobStorage),
			},
		}...)
	}

	estimation := models.Estimation{
		Cost:          utils.PointerTo(float32(jobTotalCost)),
		CostBreakdown: &costBreakdown,
		Size:          utils.PointerTo(int(scanSizeGB)),
		Duration:      utils.PointerTo(int(scanDurationSec)),
	}

	return &estimation, nil
}

// nolint:exhaustive
func getSnapshotTypeFromDiskType(diskType armcompute.DiskStorageAccountTypes) (armcompute.SnapshotStorageAccountTypes, error) {
	switch diskType {
	case armcompute.DiskStorageAccountTypesStandardSSDLRS:
		return armcompute.SnapshotStorageAccountTypesStandardLRS, nil
	default:
		return "", fmt.Errorf("unsupported disk type: %v", diskType)
	}
}
