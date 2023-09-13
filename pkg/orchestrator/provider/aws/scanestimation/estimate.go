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
	kubeclarityUtils "github.com/openclarity/kubeclarity/shared/pkg/utils"

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
	MBInGB          = 1000
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

// FamilyScanDurationsMap Calculate the logarithmic fit of each family base on static measurements of family scan duration in seconds per scanSizesGB value.
// The tests were made on a t2.large instance with a gp2 volume.
// The times correspond to the scan size values in scanSizesGB.
// TODO add infoFinder family stats.
// nolint:gomnd
var FamilyScanDurationsMap = map[familiestypes.FamilyType]*common.LogarithmicFormula{
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
	scanSizeMB, err := getScanSize(params.Stats, params.Asset)
	if err != nil {
		return nil, fmt.Errorf("failed to get scan size: %w", err)
	}
	scanSizeGB := float64(scanSizeMB) / MBInGB
	scanDurationSec := getScanDuration(params.Stats, familiesConfig, scanSizeMB)

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

// Search in all the families stats and look for the first family (by random order) that has scan size stats for ROOTFS scan.
// nolint:cyclop
func getScanSize(stats models.AssetScanStats, asset *models.Asset) (int64, error) {
	var scanSizeMB int64
	const half = 2

	sbomStats, ok := findMatchingStatsForInputTypeRootFS(stats.Sbom)
	if ok {
		if sbomStats.Size != nil && *sbomStats.Size > 0 {
			return *sbomStats.Size, nil
		}
	}

	vulStats, ok := findMatchingStatsForInputTypeRootFS(stats.Vulnerabilities)
	if ok {
		if vulStats.Size != nil && *vulStats.Size > 0 {
			return *vulStats.Size, nil
		}
	}

	secretsStats, ok := findMatchingStatsForInputTypeRootFS(stats.Secrets)
	if ok {
		if secretsStats.Size != nil && *secretsStats.Size > 0 {
			return *secretsStats.Size, nil
		}
	}

	exploitsStats, ok := findMatchingStatsForInputTypeRootFS(stats.Exploits)
	if ok {
		if exploitsStats.Size != nil && *exploitsStats.Size > 0 {
			return *exploitsStats.Size, nil
		}
	}

	rootkitsStats, ok := findMatchingStatsForInputTypeRootFS(stats.Rootkits)
	if ok {
		if rootkitsStats.Size != nil && *rootkitsStats.Size > 0 {
			return *rootkitsStats.Size, nil
		}
	}

	misconfigurationsStats, ok := findMatchingStatsForInputTypeRootFS(stats.Misconfigurations)
	if ok {
		if misconfigurationsStats.Size != nil && *misconfigurationsStats.Size > 0 {
			return *misconfigurationsStats.Size, nil
		}
	}

	malwareStats, ok := findMatchingStatsForInputTypeRootFS(stats.Malware)
	if ok {
		if malwareStats.Size != nil && *malwareStats.Size > 0 {
			return *malwareStats.Size, nil
		}
	}

	// if scan size was not found from the previous scan stats, estimate the scan size from the asset root volume size
	vminfo, err := asset.AssetInfo.AsVMInfo()
	if err != nil {
		return 0, fmt.Errorf("failed to use asset info as vminfo: %w", err)
	}
	sourceVolumeSizeMB := int64(vminfo.RootVolume.SizeGB * MBInGB)
	scanSizeMB = sourceVolumeSizeMB / half // Volumes are normally only about 50% full

	return scanSizeMB, nil
}

// findMatchingStatsForInputTypeRootFS will find the first stats for rootfs scan.
func findMatchingStatsForInputTypeRootFS(stats *[]models.AssetScanInputScanStats) (models.AssetScanInputScanStats, bool) {
	if stats == nil {
		return models.AssetScanInputScanStats{}, false
	}
	for i, scanStats := range *stats {
		if *scanStats.Type == string(kubeclarityUtils.ROOTFS) {
			ret := *stats
			return ret[i], true
		}
	}
	return models.AssetScanInputScanStats{}, false
}

// nolint:cyclop
func getScanDuration(stats models.AssetScanStats, familiesConfig *models.ScanFamiliesConfig, scanSizeMB int64) int64 {
	var totalScanDuration int64

	scanSizeGB := float64(scanSizeMB) / MBInGB

	if familiesConfig.Sbom.IsEnabled() {
		scanDuration := getScanDurationFromStats(stats.Sbom)
		if scanDuration != 0 {
			totalScanDuration += scanDuration
		} else {
			// if we didn't find the duration from the stats, take it from our static scan duration map.
			totalScanDuration += int64(FamilyScanDurationsMap[familiestypes.SBOM].Evaluate(scanSizeGB))
		}
	}

	if familiesConfig.Vulnerabilities.IsEnabled() {
		scanDuration := getScanDurationFromStats(stats.Vulnerabilities)
		if scanDuration != 0 {
			totalScanDuration += scanDuration
		} else {
			totalScanDuration += int64(FamilyScanDurationsMap[familiestypes.Vulnerabilities].Evaluate(scanSizeGB))
		}
	}

	if familiesConfig.Secrets.IsEnabled() {
		scanDuration := getScanDurationFromStats(stats.Secrets)
		if scanDuration != 0 {
			totalScanDuration += scanDuration
		} else {
			totalScanDuration += int64(FamilyScanDurationsMap[familiestypes.Secrets].Evaluate(scanSizeGB))
		}
	}

	if familiesConfig.Exploits.IsEnabled() {
		scanDuration := getScanDurationFromStats(stats.Exploits)
		if scanDuration != 0 {
			totalScanDuration += scanDuration
		} else {
			totalScanDuration += int64(FamilyScanDurationsMap[familiestypes.Exploits].Evaluate(scanSizeGB))
		}
	}

	if familiesConfig.Rootkits.IsEnabled() {
		scanDuration := getScanDurationFromStats(stats.Rootkits)
		if scanDuration != 0 {
			totalScanDuration += scanDuration
		} else {
			totalScanDuration += int64(FamilyScanDurationsMap[familiestypes.Rootkits].Evaluate(scanSizeGB))
		}
	}

	if familiesConfig.Misconfigurations.IsEnabled() {
		scanDuration := getScanDurationFromStats(stats.Misconfigurations)
		if scanDuration != 0 {
			totalScanDuration += scanDuration
		} else {
			totalScanDuration += int64(FamilyScanDurationsMap[familiestypes.Misconfiguration].Evaluate(scanSizeGB))
		}
	}

	if familiesConfig.Malware.IsEnabled() {
		scanDuration := getScanDurationFromStats(stats.Malware)
		if scanDuration != 0 {
			totalScanDuration += scanDuration
		} else {
			totalScanDuration += int64(FamilyScanDurationsMap[familiestypes.Malware].Evaluate(scanSizeGB))
		}
	}

	return totalScanDuration
}

func getScanDurationFromStats(stats *[]models.AssetScanInputScanStats) int64 {
	stat, ok := findMatchingStatsForInputTypeRootFS(stats)
	if !ok {
		return 0
	}

	if stat.ScanTime == nil {
		return 0
	}
	if stat.ScanTime.EndTime == nil || stat.ScanTime.StartTime == nil {
		return 0
	}

	dur := stat.ScanTime.EndTime.Sub(*stat.ScanTime.StartTime)

	return int64(dur.Seconds())
}
