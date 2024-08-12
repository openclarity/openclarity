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

package utils

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/provider"
	provider_service "github.com/openclarity/vmclarity/provider/external/utils/proto"
)

func ConvertAssetToModels(asset *provider_service.Asset) (apitypes.Asset, error) {
	if asset == nil {
		return apitypes.Asset{}, errors.New("asset is nil")
	}

	assetType := apitypes.AssetType{}
	switch asset.AssetType.(type) {
	case *provider_service.Asset_Vminfo:
		vminfo := asset.GetVminfo()

		if err := assetType.FromVMInfo(apitypes.VMInfo{
			Image:            vminfo.Image,
			InstanceID:       vminfo.Id,
			InstanceProvider: to.Ptr(apitypes.External),
			InstanceType:     vminfo.InstanceType,
			LaunchTime:       vminfo.LaunchTime.AsTime(),
			Location:         vminfo.Location,
			Platform:         vminfo.Platform,
			SecurityGroups:   &[]apitypes.SecurityGroup{},
			Tags:             convertTagsToModels(vminfo.Tags),
		}); err != nil {
			return apitypes.Asset{}, fmt.Errorf("failed to convert asset from VMInfo: %w", err)
		}
	case *provider_service.Asset_Dirinfo:
		dirinfo := asset.GetDirinfo()

		if err := assetType.FromDirInfo(apitypes.DirInfo{
			DirName:  to.Ptr(dirinfo.DirName),
			Location: to.Ptr(dirinfo.Location),
		}); err != nil {
			return apitypes.Asset{}, fmt.Errorf("failed to convert asset from Dirinfo: %w", err)
		}
	case *provider_service.Asset_Podinfo:
		podinfo := asset.GetPodinfo()

		if err := assetType.FromPodInfo(apitypes.PodInfo{
			PodName:  to.Ptr(podinfo.PodName),
			Location: to.Ptr(podinfo.Location),
		}); err != nil {
			return apitypes.Asset{}, fmt.Errorf("failed to convert asset from Podinfo: %w", err)
		}
	default:
		return apitypes.Asset{}, fmt.Errorf("unsupported asset type: %t", asset.AssetType)
	}

	return apitypes.Asset{
		AssetInfo: &assetType,
	}, nil
}

func convertAssetFromModels(asset apitypes.Asset) (*provider_service.Asset, error) {
	value, err := asset.AssetInfo.ValueByDiscriminator()
	if err != nil {
		return nil, fmt.Errorf("failed to value by discriminator from asset info: %w", err)
	}

	switch info := value.(type) {
	case apitypes.VMInfo:
		return &provider_service.Asset{
			AssetType: &provider_service.Asset_Vminfo{Vminfo: &provider_service.VMInfo{
				Id:           info.InstanceID,
				Location:     info.Location,
				Image:        info.Image,
				InstanceType: info.InstanceType,
				Platform:     info.Platform,
				Tags:         convertTagsFromModels(info.Tags),
				LaunchTime:   timestamppb.New(info.LaunchTime),
			}},
		}, nil
	case apitypes.DirInfo:
		return &provider_service.Asset{
			AssetType: &provider_service.Asset_Dirinfo{Dirinfo: &provider_service.DirInfo{
				DirName:  *info.DirName,
				Location: *info.Location,
			}},
		}, nil
	case apitypes.PodInfo:
		return &provider_service.Asset{
			AssetType: &provider_service.Asset_Podinfo{Podinfo: &provider_service.PodInfo{
				PodName:  *info.PodName,
				Location: *info.Location,
			}},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported asset type: %t", info)
	}
}

func convertTagsToModels(tags []*provider_service.Tag) *[]apitypes.Tag {
	ret := make([]apitypes.Tag, 0)

	if len(tags) == 0 {
		return nil
	}

	for _, tag := range tags {
		ret = append(ret, apitypes.Tag{
			Key:   tag.Key,
			Value: tag.Val,
		})
	}

	return &ret
}

func convertTagsFromModels(tags *[]apitypes.Tag) []*provider_service.Tag {
	ret := make([]*provider_service.Tag, 0)

	if tags == nil {
		return nil
	}

	for _, tag := range *tags {
		ret = append(ret, &provider_service.Tag{
			Key: tag.Key,
			Val: tag.Value,
		})
	}

	return ret
}

func ConvertScanJobConfig(config *provider.ScanJobConfig) (*provider_service.ScanJobConfig, error) {
	asset, err := convertAssetFromModels(config.Asset)
	if err != nil {
		return nil, fmt.Errorf("failed to convert asset from models asset: %w", err)
	}

	ret := provider_service.ScanJobConfig{
		ScannerImage:     config.ScannerImage,
		ScannerCLIConfig: config.ScannerCLIConfig,
		VmClarityAddress: config.VMClarityAddress,
		ScanMetadata: &provider_service.ScanMetadata{
			ScanID:      config.ScanID,
			AssetScanID: config.AssetScanID,
			AssetID:     config.AssetID,
		},
		ScannerInstanceCreationConfig: &provider_service.ScannerInstanceCreationConfig{
			UseSpotInstances: config.UseSpotInstances,
		},
		Asset: asset,
	}
	if config.MaxPrice != nil {
		ret.ScannerInstanceCreationConfig.MaxPrice = *config.MaxPrice
	}
	if config.RetryMaxAttempts != nil {
		ret.ScannerInstanceCreationConfig.RetryMaxAttempts = int32(*config.RetryMaxAttempts)
	}

	return &ret, nil
}
