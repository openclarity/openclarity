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

package discoverer

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/sirupsen/logrus"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/provider/aws/types"
	"github.com/openclarity/vmclarity/provider/aws/utils"
)

var _ provider.Discoverer = &Discoverer{}

type Discoverer struct {
	Ec2Client *ec2.Client
}

func (d *Discoverer) DiscoverAssets(ctx context.Context) provider.AssetDiscoverer {
	assetDiscoverer := provider.NewSimpleAssetDiscoverer()

	go func() {
		defer close(assetDiscoverer.OutputChan)

		regions, err := d.ListAllRegions(ctx)
		if err != nil {
			assetDiscoverer.Error = fmt.Errorf("failed to get regions: %w", err)
			return
		}

		logger := log.GetLoggerFromContextOrDiscard(ctx)

		for _, region := range regions {
			instances, err := d.GetInstances(ctx, []ec2types.Filter{}, region.Name)
			if err != nil {
				logger.Warnf("Failed to get instances. region=%v: %v", region, err)
				continue
			}

			for _, instance := range instances {
				asset, err := getVMInfoFromInstance(instance)
				if err != nil {
					assetDiscoverer.Error = utils.FatalError{
						Err: fmt.Errorf("failed convert EC2 Instance to AssetType: %w", err),
					}
					return
				}
				select {
				case assetDiscoverer.OutputChan <- asset:
				case <-ctx.Done():
					assetDiscoverer.Error = ctx.Err()
					return
				}
			}
		}
	}()

	return assetDiscoverer
}

func (d *Discoverer) ListAllRegions(ctx context.Context) ([]types.Region, error) {
	ret := make([]types.Region, 0)
	out, err := d.Ec2Client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{
		AllRegions: nil, // display also disabled regions?
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe regions: %w", err)
	}

	for _, region := range out.Regions {
		ret = append(ret, types.Region{
			Name: *region.RegionName,
		})
	}

	return ret, nil
}

func (d *Discoverer) GetInstances(ctx context.Context, filters []ec2types.Filter, regionID string) ([]types.Instance, error) {
	ret := make([]types.Instance, 0)

	input := &ec2.DescribeInstancesInput{
		MaxResults: to.Ptr[int32](types.MaxResults), // TODO what will be a good number?
	}
	if len(filters) > 0 {
		input.Filters = filters
	}

	out, err := d.Ec2Client.DescribeInstances(ctx, input, func(options *ec2.Options) {
		options.Region = regionID
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %w", err)
	}
	ret = append(ret, d.getInstancesFromDescribeInstancesOutput(ctx, out, regionID)...)

	// use pagination
	// TODO we can make it better by not saving all results in memory. See https://github.com/openclarity/vmclarity/pull/3#discussion_r1021656861
	for out.NextToken != nil {
		input := &ec2.DescribeInstancesInput{
			MaxResults: to.Ptr[int32](types.MaxResults), // TODO what will be a good number?
			NextToken:  out.NextToken,
		}
		if len(filters) > 0 {
			input.Filters = filters
		}

		out, err = d.Ec2Client.DescribeInstances(ctx, input, func(options *ec2.Options) {
			options.Region = regionID
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe instances: %w", err)
		}
		ret = append(ret, d.getInstancesFromDescribeInstancesOutput(ctx, out, regionID)...)
	}

	return ret, nil
}

func getVMInfoFromInstance(i types.Instance) (apitypes.AssetType, error) {
	assetType := apitypes.AssetType{}
	err := assetType.FromVMInfo(apitypes.VMInfo{
		Image:            i.Image,
		InstanceID:       i.ID,
		InstanceProvider: to.Ptr(apitypes.AWS),
		InstanceType:     i.InstanceType,
		LaunchTime:       i.LaunchTime,
		Location:         i.Location(),
		ObjectType:       "VMInfo",
		Platform:         i.Platform,
		RootVolume: apitypes.RootVolume{
			Encrypted: i.RootVolumeEncrypted,
			SizeGB:    int(i.RootVolumeSizeGB),
		},
		SecurityGroups: to.Ptr(i.SecurityGroups),
		Tags:           to.Ptr(i.Tags),
	})
	if err != nil {
		err = fmt.Errorf("failed to create AssetType from VMInfo: %w", err)
	}

	return assetType, err
}

func (d *Discoverer) getInstancesFromDescribeInstancesOutput(ctx context.Context, result *ec2.DescribeInstancesOutput, regionID string) []types.Instance {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	var ret []types.Instance
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			// Ignore terminated instances they are destroyed and
			// will be garbage collected by AWS.
			if instance.State.Name == ec2types.InstanceStateNameTerminated {
				continue
			}

			if err := validateInstanceFields(instance); err != nil {
				logger.Errorf("Instance validation failed. instance id=%v: %v", to.ValueOrZero(instance.InstanceId), err)
				continue
			}
			rootVol, err := getRootVolumeInfo(ctx, d.Ec2Client, instance, regionID)
			if err != nil {
				logger.Warnf("Couldn't get root volume info. instance id=%v: %v", to.ValueOrZero(instance.InstanceId), err)
				rootVol = &apitypes.RootVolume{
					SizeGB:    0,
					Encrypted: apitypes.RootVolumeEncryptedUnknown,
				}
			}

			ret = append(ret, types.Instance{
				ID:                  *instance.InstanceId,
				Region:              regionID,
				AvailabilityZone:    *instance.Placement.AvailabilityZone,
				Image:               *instance.ImageId,
				InstanceType:        string(instance.InstanceType),
				Platform:            *instance.PlatformDetails,
				Tags:                utils.GetTagsFromECTags(instance.Tags),
				LaunchTime:          *instance.LaunchTime,
				VpcID:               *instance.VpcId,
				SecurityGroups:      getSecurityGroupsIDs(instance.SecurityGroups),
				RootDeviceName:      to.ValueOrZero(instance.RootDeviceName),
				RootVolumeSizeGB:    int32(rootVol.SizeGB),
				RootVolumeEncrypted: rootVol.Encrypted,
				Ec2Client:           d.Ec2Client,
			})
		}
	}
	return ret
}

func validateInstanceFields(instance ec2types.Instance) error {
	if instance.InstanceId == nil {
		return errors.New("instance id does not exist")
	}
	if instance.Placement == nil {
		return errors.New("insatnce Placement does not exist")
	}
	if instance.Placement.AvailabilityZone == nil {
		return errors.New("insatnce AvailabilityZone does not exist")
	}
	if instance.ImageId == nil {
		return errors.New("instance ImageId does not exist")
	}
	if instance.PlatformDetails == nil {
		return errors.New("instance PlatformDetails does not exist")
	}
	if instance.LaunchTime == nil {
		return errors.New("instance LaunchTime does not exist")
	}
	if instance.VpcId == nil {
		return errors.New("instance VpcId does not exist")
	}
	return nil
}

func getRootVolumeInfo(ctx context.Context, client *ec2.Client, i ec2types.Instance, region string) (*apitypes.RootVolume, error) {
	if i.RootDeviceName == nil || *i.RootDeviceName == "" {
		return nil, errors.New("RootDeviceName is not set")
	}
	logger := log.GetLoggerFromContextOrDiscard(ctx)
	for _, mapping := range i.BlockDeviceMappings {
		if to.ValueOrZero(mapping.DeviceName) == to.ValueOrZero(i.RootDeviceName) {
			if mapping.Ebs == nil {
				return nil, errors.New("EBS of the root volume is nil")
			}
			if mapping.Ebs.VolumeId == nil {
				return nil, errors.New("volume ID of the root volume is nil")
			}
			descParams := &ec2.DescribeVolumesInput{
				VolumeIds: []string{*mapping.Ebs.VolumeId},
			}

			describeOut, err := client.DescribeVolumes(ctx, descParams, func(options *ec2.Options) {
				options.Region = region
			})
			if err != nil {
				return nil, errors.New("failed to describe the root volume")
			}

			if len(describeOut.Volumes) == 0 {
				return nil, errors.New("volume list is empty")
			}
			if len(describeOut.Volumes) > 1 {
				logger.WithFields(logrus.Fields{
					"VolumeId":    *mapping.Ebs.VolumeId,
					"Volumes num": len(describeOut.Volumes),
				}).Warnf("Found more than 1 root volume, using the first")
			}

			return &apitypes.RootVolume{
				SizeGB:    int(to.ValueOrZero(describeOut.Volumes[0].Size)),
				Encrypted: encryptedToAPI(describeOut.Volumes[0].Encrypted),
			}, nil
		}
	}

	return nil, errors.New("instance doesn't have a root volume block device mapping")
}

func getSecurityGroupsIDs(sg []ec2types.GroupIdentifier) []apitypes.SecurityGroup {
	securityGroups := make([]apitypes.SecurityGroup, 0, len(sg))
	for _, s := range sg {
		if s.GroupId == nil {
			continue
		}
		securityGroups = append(securityGroups, apitypes.SecurityGroup{
			Id: *s.GroupId,
		})
	}

	return securityGroups
}

func encryptedToAPI(encrypted *bool) apitypes.RootVolumeEncrypted {
	if encrypted == nil {
		return apitypes.RootVolumeEncryptedUnknown
	}
	if *encrypted {
		return apitypes.RootVolumeEncryptedYes
	}
	return apitypes.RootVolumeEncryptedNo
}
