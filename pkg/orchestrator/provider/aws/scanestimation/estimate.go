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
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/pricing"

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
	priceFetcher PriceFetcher
}

type EstimateAssetScanParams struct {
	SourceRegion            string
	DestRegion              string
	ScannerVolumeType       ec2types.VolumeType
	FromSnapshotVolumeType  ec2types.VolumeType
	ScannerInstanceType     ec2types.InstanceType
	JobCreationTimeSec      int64
	ScannerRootVolumeSizeGB int64
	Stats                   models.AssetScanStats
	Asset                   *models.Asset
	AssetScanTemplate       *models.AssetScanTemplate
}

func New(pricingClient *pricing.Client, ec2Client *ec2.Client) *ScanEstimator {
	return &ScanEstimator{
		priceFetcher: &PriceFetcherImpl{pricingClient: pricingClient, ec2Client: ec2Client},
	}
}

const (
	SecondsInAMonth = 86400 * 30
	SecondsInAnHour = 60 * 60
)

type recipeResource string

const (
	Snapshot           recipeResource = "Snapshot"
	ScannerInstance    recipeResource = "ScannerInstance"
	VolumeFromSnapshot recipeResource = "VolumeFromSnapshot"
	ScannerRootVolume  recipeResource = "ScannerRootVolume"
	DataTransfer       recipeResource = "DataTransfer"
)

// scanSizesGB represents the memory sizes on the machines that the tests were taken on.
var scanSizesGB = []float64{0.01, 2.5, 8.1}

// familyScanDurationsMap Calculate the logarithmic fit of each family base on static measurements of family scan duration in seconds per scanSizesGB value.
// The tests were made on a t2.large instance with a gp2 volume.
// The times correspond to the scan size values in scanSizesGB.
// TODO add infoFinder family stats.
// nolint:gomnd
var familyScanDurationsMap = map[familiestypes.FamilyType]*common.LogarithmicFormula{
	familiestypes.SBOM:             common.MustLogarithmicFit(scanSizesGB, []float64{0.01, 11, 37}),
	familiestypes.Vulnerabilities:  common.MustLogarithmicFit(scanSizesGB, []float64{0.01, 1, 11}), // TODO check time with no sbom scan
	familiestypes.Secrets:          common.MustLogarithmicFit(scanSizesGB, []float64{0.01, 720, 1320}),
	familiestypes.Exploits:         common.MustLogarithmicFit(scanSizesGB, []float64{0, 0, 0}),
	familiestypes.Rootkits:         common.MustLogarithmicFit(scanSizesGB, []float64{0, 0, 0}),
	familiestypes.Misconfiguration: common.MustLogarithmicFit(scanSizesGB, []float64{0.01, 4, 5}),
	familiestypes.Malware:          common.MustLogarithmicFit(scanSizesGB, []float64{0.01, 840, 2460}),
}

// Reserved Instances are not physical instances, but rather a billing discount that is applied to the running On-Demand Instances in your account.
// The On-Demand Instances must match certain specifications of the Reserved Instances in order to benefit from the billing discount.
// A decade after launching Reserved Instances (RIs), Amazon Web Services (AWS) introduced Savings Plans as a more flexible alternative to RIs. AWS Savings Plans are not meant to replace Reserved Instances; they are complementary.

// We are not taking into account Reserved Instances (RIs) or Saving Plans (SPs) since we don't know the exact OnDemand configuration in order to launch them.
// In the future, we can let the user choose to use RI's or SP's as the scanner instances.

// nolint:cyclop
func (s *ScanEstimator) EstimateAssetScan(ctx context.Context, params EstimateAssetScanParams) (*models.Estimation, error) {
	var sourceSnapshotMonthlyCost float64
	var err error

	if params.AssetScanTemplate == nil || params.AssetScanTemplate.ScanFamiliesConfig == nil {
		return nil, errors.New("scan families config was not provided")
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
	fromSnapshotVolumeType := params.FromSnapshotVolumeType
	jobCreationTimeSec := params.JobCreationTimeSec
	scannerInstanceType := params.ScannerInstanceType
	scannerRootVolumeSizeGB := params.ScannerRootVolumeSizeGB
	scannerVolumeType := params.ScannerVolumeType

	// Get relevant current prices from AWS price list API
	// Fetch the dest snapshot monthly cost.
	destSnapshotMonthlyCost, err := s.priceFetcher.GetSnapshotMonthlyCostPerGB(ctx, destRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly cost for destination snapshot: %w", err)
	}

	// Fetch the scanner instance hourly cost.
	scannerPerHourCost, err := s.priceFetcher.GetInstancePerHourCost(ctx, destRegion, scannerInstanceType, marketOption)
	if err != nil {
		return nil, fmt.Errorf("failed to get scanner per hour cost: %w", err)
	}

	// Fetch the scanner job volume monthly cost.
	scannerRootVolumeMonthlyCost, err := s.priceFetcher.GetVolumeMonthlyCostPerGB(ctx, destRegion, scannerVolumeType)
	if err != nil {
		return nil, fmt.Errorf("failed to get volume monthly cost per GB: %w", err)
	}

	// Fetch the monthly cost of the volume that was created from the snapshot.
	volumeFromSnapshotMonthlyCost, err := s.priceFetcher.GetVolumeMonthlyCostPerGB(ctx, destRegion, fromSnapshotVolumeType)
	if err != nil {
		return nil, fmt.Errorf("failed to get volume monthly cost per GB: %w", err)
	}

	dataTransferCost := 0.0
	sourceSnapshotCost := 0.0
	if sourceRegion != destRegion {
		// if the scanner is in a different region then the scanned asset, we have another snapshot created in the
		// source region.
		sourceSnapshotMonthlyCost, err = s.priceFetcher.GetSnapshotMonthlyCostPerGB(ctx, sourceRegion)
		if err != nil {
			return nil, fmt.Errorf("failed to get source snapshot monthly cost: %w", err)
		}

		sourceSnapshotCost = sourceSnapshotMonthlyCost * ((float64(jobCreationTimeSec + scanDurationSec)) / SecondsInAMonth) * scanSizeGB

		// Fetch the data transfer cost per GB (if source and dest regions are the same, this will be 0).
		dataTransferCostPerGB, err := s.priceFetcher.GetDataTransferCostPerGB(sourceRegion, destRegion)
		if err != nil {
			return nil, fmt.Errorf("failed to get data transfer cost per GB: %w", err)
		}

		dataTransferCost = dataTransferCostPerGB * scanSizeGB
	}

	destSnapshotCost := destSnapshotMonthlyCost * ((float64(jobCreationTimeSec + scanDurationSec)) / SecondsInAMonth) * scanSizeGB
	volumeFromSnapshotCost := volumeFromSnapshotMonthlyCost * ((float64(jobCreationTimeSec + scanDurationSec)) / SecondsInAMonth) * scanSizeGB
	scannerCost := scannerPerHourCost * ((float64(jobCreationTimeSec + scanDurationSec)) / SecondsInAnHour)
	scannerRootVolumeCost := scannerRootVolumeMonthlyCost * ((float64(jobCreationTimeSec + scanDurationSec)) / SecondsInAMonth) * float64(scannerRootVolumeSizeGB)

	jobTotalCost := sourceSnapshotCost + volumeFromSnapshotCost + scannerCost + scannerRootVolumeCost + dataTransferCost + destSnapshotCost

	// Create the Estimation object base on the calculated data.
	costBreakdown := []models.CostBreakdownComponent{
		{
			Cost:      float32(destSnapshotCost),
			Operation: fmt.Sprintf("%v-%v", Snapshot, destRegion),
		},
		{
			Cost:      float32(scannerCost),
			Operation: string(ScannerInstance),
		},
		{
			Cost:      float32(volumeFromSnapshotCost),
			Operation: string(VolumeFromSnapshot),
		},
		{
			Cost:      float32(scannerRootVolumeCost),
			Operation: string(ScannerRootVolume),
		},
	}
	if sourceRegion != destRegion {
		costBreakdown = append(costBreakdown, []models.CostBreakdownComponent{
			{
				Cost:      float32(sourceSnapshotCost),
				Operation: fmt.Sprintf("%v-%v", Snapshot, sourceRegion),
			},
			{
				Cost:      float32(dataTransferCost),
				Operation: string(DataTransfer),
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
