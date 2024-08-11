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

package estimator

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/provider/aws/estimator/scanestimation"
	"github.com/openclarity/vmclarity/provider/aws/types"
)

var _ provider.Estimator = &Estimator{}

type Estimator struct {
	ScannerRegion       string
	ScannerInstanceType string
	ScanEstimator       *scanestimation.ScanEstimator
	Ec2Client           *ec2.Client
}

func (e *Estimator) Estimate(ctx context.Context, stats apitypes.AssetScanStats, asset *apitypes.Asset, template *apitypes.AssetScanTemplate) (*apitypes.Estimation, error) {
	const jobCreationTimeConst = 2

	vminfo, err := asset.AssetInfo.AsVMInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to use asset info as vminfo: %w", err)
	}

	location, err := types.NewLocation(vminfo.Location)
	if err != nil {
		return nil, fmt.Errorf("failed to parse location %v: %w", vminfo.Location, err)
	}

	sourceRegion := location.Region
	destRegion := e.ScannerRegion
	scannerInstanceType := e.ScannerInstanceType

	scannerRootVolumeSizeGB := vminfo.RootVolume.SizeGB
	scannerVolumeType := ec2types.VolumeTypeGp2                          // TODO this should come from configuration once we support more than one volume type.
	fromSnapshotVolumeType := ec2types.VolumeTypeGp2                     // TODO this should come from configuration once we support more than one volume type.
	jobCreationTimeSec := jobCreationTimeConst * scannerRootVolumeSizeGB // TODO create a formula to calculate this per GB

	params := scanestimation.EstimateAssetScanParams{
		SourceRegion:            sourceRegion,
		DestRegion:              destRegion,
		ScannerVolumeType:       scannerVolumeType,
		FromSnapshotVolumeType:  fromSnapshotVolumeType,
		ScannerInstanceType:     ec2types.InstanceType(scannerInstanceType),
		JobCreationTimeSec:      int64(jobCreationTimeSec),
		ScannerRootVolumeSizeGB: int64(scannerRootVolumeSizeGB),
		Stats:                   stats,
		Asset:                   asset,
		AssetScanTemplate:       template,
	}
	ret, err := e.ScanEstimator.EstimateAssetScan(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate asset scan: %w", err)
	}

	return ret, nil
}
