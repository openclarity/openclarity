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

package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/orchestrator/provider"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

func EC2TagsFromScanMetadata(meta provider.ScanMetadata) []ec2types.Tag {
	return []ec2types.Tag{
		{
			Key:   utils.PointerTo(EC2TagKeyOwner),
			Value: utils.PointerTo(EC2TagValueOwner),
		},
		{
			Key:   utils.PointerTo(EC2TagKeyName),
			Value: utils.PointerTo(fmt.Sprintf(EC2TagValueNamePattern, meta.AssetScanID)),
		},
		{
			Key:   utils.PointerTo(EC2TagKeyScanID),
			Value: utils.PointerTo(meta.ScanID),
		},
		{
			Key:   utils.PointerTo(EC2TagKeyAssetScanID),
			Value: utils.PointerTo(meta.AssetScanID),
		},
		{
			Key:   utils.PointerTo(EC2TagKeyAssetID),
			Value: utils.PointerTo(meta.AssetID),
		},
	}
}

func EC2FiltersFromEC2Tags(tags []ec2types.Tag) []ec2types.Filter {
	filters := make([]ec2types.Filter, 0, len(tags))
	for _, tag := range tags {
		var name string
		var value []string

		if tag.Key == nil || *tag.Key == "" {
			continue
		}
		name = fmt.Sprintf("tag:%s", *tag.Key)

		if tag.Value != nil {
			value = []string{*tag.Value}
		}

		filters = append(filters, ec2types.Filter{
			Name:   utils.PointerTo(name),
			Values: value,
		})
	}

	return filters
}

func EC2FiltersFromTags(tags []models.Tag) []ec2types.Filter {
	filters := make([]ec2types.Filter, 0, len(tags))
	for _, tag := range tags {
		var name string
		var value []string

		if tag.Key == "" {
			continue
		}

		name = fmt.Sprintf("tag:%s", tag.Key)
		if tag.Value != "" {
			value = []string{tag.Value}
		}

		filters = append(filters, ec2types.Filter{
			Name:   utils.PointerTo(name),
			Values: value,
		})
	}

	return filters
}

func instanceFromEC2Instance(i *ec2types.Instance, client *ec2.Client, region string, config *provider.ScanJobConfig) *Instance {
	securityGroups := getSecurityGroupsFromEC2GroupIdentifiers(i.SecurityGroups)
	tags := getTagsFromECTags(i.Tags)

	volumes := make([]Volume, len(i.BlockDeviceMappings))
	for idx, blkDevice := range i.BlockDeviceMappings {
		var blockDeviceName string

		if blkDevice.DeviceName != nil {
			blockDeviceName = *blkDevice.DeviceName
		}

		volumes[idx] = Volume{
			ec2Client:       client,
			ID:              *blkDevice.Ebs.VolumeId,
			Region:          region,
			BlockDeviceName: blockDeviceName,
			Metadata:        config.ScanMetadata,
		}
	}

	return &Instance{
		ID:               *i.InstanceId,
		Region:           region,
		VpcID:            *i.VpcId,
		SecurityGroups:   securityGroups,
		AvailabilityZone: *i.Placement.AvailabilityZone,
		Image:            *i.ImageId,
		InstanceType:     string(i.InstanceType),
		Platform:         string(i.Platform),
		Tags:             tags,
		LaunchTime:       *i.LaunchTime,
		RootDeviceName:   *i.RootDeviceName,
		Volumes:          volumes,
		Metadata:         config.ScanMetadata,

		ec2Client: client,
	}
}

func getTagsFromECTags(tags []ec2types.Tag) []models.Tag {
	if len(tags) == 0 {
		return nil
	}

	ret := make([]models.Tag, len(tags))
	for i, tag := range tags {
		ret[i] = models.Tag{
			Key:   *tag.Key,
			Value: *tag.Value,
		}
	}
	return ret
}

func getInstanceState(result *ec2.DescribeInstancesOutput, instanceID string) ec2types.InstanceStateName {
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			if strings.Compare(*instance.InstanceId, instanceID) == 0 {
				if instance.State != nil {
					return instance.State.Name
				}
			}
		}
	}
	return ec2types.InstanceStateNamePending
}

func validateInstanceFields(instance ec2types.Instance) error {
	if instance.InstanceId == nil {
		return fmt.Errorf("instance id does not exist")
	}
	if instance.Placement == nil {
		return fmt.Errorf("insatnce Placement does not exist")
	}
	if instance.Placement.AvailabilityZone == nil {
		return fmt.Errorf("insatnce AvailabilityZone does not exist")
	}
	if instance.ImageId == nil {
		return fmt.Errorf("instance ImageId does not exist")
	}
	if instance.PlatformDetails == nil {
		return fmt.Errorf("instance PlatformDetails does not exist")
	}
	if instance.LaunchTime == nil {
		return fmt.Errorf("instance LaunchTime does not exist")
	}
	if instance.VpcId == nil {
		return fmt.Errorf("instance VpcId does not exist")
	}
	return nil
}

func getSecurityGroupsIDs(sg []ec2types.GroupIdentifier) []models.SecurityGroup {
	securityGroups := make([]models.SecurityGroup, 0, len(sg))
	for _, s := range sg {
		if s.GroupId == nil {
			continue
		}
		securityGroups = append(securityGroups, models.SecurityGroup{
			Id: *s.GroupId,
		})
	}

	return securityGroups
}

func getSecurityGroupsFromEC2GroupIdentifiers(identifiers []ec2types.GroupIdentifier) []models.SecurityGroup {
	var ret []models.SecurityGroup

	for _, identifier := range identifiers {
		if identifier.GroupId != nil {
			ret = append(ret, models.SecurityGroup{
				Id: *identifier.GroupId,
			})
		}
	}

	return ret
}

func getVMInfoFromInstance(i Instance) (models.AssetType, error) {
	assetType := models.AssetType{}
	err := assetType.FromVMInfo(models.VMInfo{
		Image:            i.Image,
		InstanceID:       i.ID,
		InstanceProvider: utils.PointerTo(models.AWS),
		InstanceType:     i.InstanceType,
		LaunchTime:       i.LaunchTime,
		Location:         i.Location(),
		ObjectType:       "VMInfo",
		Platform:         i.Platform,
		RootVolume: models.RootVolume{
			Encrypted: i.RootVolumeEncrypted,
			SizeGB:    int(i.RootVolumeSizeGB),
		},
		SecurityGroups: utils.PointerTo(i.SecurityGroups),
		Tags:           utils.PointerTo(i.Tags),
	})
	if err != nil {
		err = fmt.Errorf("failed to create AssetType from VMInfo: %w", err)
	}

	return assetType, err
}
