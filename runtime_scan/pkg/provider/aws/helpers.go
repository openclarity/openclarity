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
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

func EC2TagsFromScanMetadata(meta provider.ScanMetadata) []ec2types.Tag {
	return []ec2types.Tag{
		{
			Key:   utils.PointerTo(EC2TagKeyOwner),
			Value: utils.PointerTo(EC2TagValueOwner),
		},
		{
			Key:   utils.PointerTo(EC2TagKeyName),
			Value: utils.PointerTo(fmt.Sprintf(EC2TagValueNamePattern, meta.ScanResultID)),
		},
		{
			Key:   utils.PointerTo(EC2TagKeyScanID),
			Value: utils.PointerTo(meta.ScanID),
		},
		{
			Key:   utils.PointerTo(EC2TagKeyScanResultID),
			Value: utils.PointerTo(meta.ScanResultID),
		},
		{
			Key:   utils.PointerTo(EC2TagKeyTargetID),
			Value: utils.PointerTo(meta.TargetID),
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

func convertFromAPIScanScope(scope *models.AwsScanScope) *ScanScope {
	var tagSelector []models.Tag
	if scope.InstanceTagSelector != nil {
		tagSelector = *scope.InstanceTagSelector
	}

	var excludeTags []models.Tag
	if scope.InstanceTagExclusion != nil {
		excludeTags = *scope.InstanceTagExclusion
	}

	return &ScanScope{
		AllRegions:  convertBool(scope.AllRegions),
		Regions:     convertFromAPIRegions(scope.Regions),
		ScanStopped: convertBool(scope.ShouldScanStoppedInstances),
		TagSelector: tagSelector,
		ExcludeTags: excludeTags,
	}
}

func convertFromAPIRegions(regions *[]models.AwsRegion) []Region {
	var ret []Region
	if regions != nil {
		for _, region := range *regions {
			ret = append(ret, Region{
				Name: region.Name,
				VPCs: convertFromAPIVPCs(region.Vpcs),
			})
		}
	}

	return ret
}

func convertFromAPIVPCs(vpcs *[]models.AwsVPC) []VPC {
	if vpcs == nil {
		return nil
	}
	ret := make([]VPC, len(*vpcs))
	for i, vpc := range *vpcs {
		ret[i] = VPC{
			ID:             vpc.Id,
			SecurityGroups: *vpc.SecurityGroups,
		}
	}

	return ret
}

func convertToAPIRegions(regions []Region) *[]models.AwsRegion {
	ret := make([]models.AwsRegion, len(regions))
	for i := range regions {
		ret[i] = models.AwsRegion{
			Name: regions[i].Name,
			Vpcs: convertToAPIVPCs(regions[i].VPCs),
		}
	}

	return &ret
}

func convertToAPIVPCs(vpcs []VPC) *[]models.AwsVPC {
	ret := make([]models.AwsVPC, len(vpcs))
	for i := range vpcs {
		ret[i] = models.AwsVPC{
			Id:             vpcs[i].ID,
			SecurityGroups: utils.PointerTo(vpcs[i].SecurityGroups),
		}
	}

	return &ret
}

func convertBool(all *bool) bool {
	if all != nil {
		return *all
	}
	return false
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

func getPointerValOrEmpty(val *string) string {
	if val == nil {
		return ""
	}
	return *val
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

func getVPCSecurityGroupsIDs(vpc VPC) []string {
	sgs := make([]string, len(vpc.SecurityGroups))
	for i, sg := range vpc.SecurityGroups {
		sgs[i] = sg.Id
	}
	return sgs
}

func createVPCFilters(vpc VPC) []ec2types.Filter {
	ret := make([]ec2types.Filter, 0)

	// create per vpc filters
	ret = append(ret, ec2types.Filter{
		Name:   utils.PointerTo(VpcIDFilterName),
		Values: []string{vpc.ID},
	})
	sgs := getVPCSecurityGroupsIDs(vpc)
	if len(sgs) > 0 {
		ret = append(ret, ec2types.Filter{
			Name:   utils.PointerTo(SecurityGroupIDFilterName),
			Values: sgs,
		})
	}

	log.Infof("VPC filter created: %+v", ret)

	return ret
}

func EC2FiltersFromInstanceState(states ...ec2types.InstanceStateName) []ec2types.Filter {
	values := make([]string, 0, len(states))
	for _, state := range states {
		values = append(values, string(state))
	}

	return []ec2types.Filter{
		{
			Name:   utils.PointerTo(InstanceStateFilterName),
			Values: values,
		},
	}
}

// If you specify multiple filters, the filters are joined with an AND, and the request returns
// only results that match all the specified filters.
func createInclusionTagsFilters(tags []models.Tag) []ec2types.Filter {
	return EC2FiltersFromTags(tags)
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

func getSecurityGroupsFromEC2SecurityGroups(groups []ec2types.SecurityGroup) []models.SecurityGroup {
	var ret []models.SecurityGroup

	for _, securityGroup := range groups {
		if securityGroup.GroupId != nil {
			ret = append(ret, models.SecurityGroup{
				Id: *securityGroup.GroupId,
			})
		}
	}

	return ret
}

func convertAwsVPCs(vpcs []ec2types.Vpc) []VPC {
	var ret []VPC
	for _, vpc := range vpcs {
		if vpc.VpcId != nil {
			ret = append(ret, VPC{
				ID:             *vpc.VpcId,
				SecurityGroups: nil,
			})
		}
	}

	return ret
}

// AND logic - if excludeTags = {tag1:val1, tag2:val2},
// then an instance will be excluded only if it has ALL these tags ({tag1:val1, tag2:val2}).
func hasExcludeTags(excludeTags []models.Tag, instanceTags []ec2types.Tag) bool {
	instanceTagsMap := make(map[string]string)

	if len(excludeTags) == 0 {
		return false
	}
	if len(instanceTags) == 0 {
		return false
	}

	for _, tag := range instanceTags {
		instanceTagsMap[*tag.Key] = *tag.Value
	}

	for _, tag := range excludeTags {
		val, ok := instanceTagsMap[tag.Key]
		if !ok {
			return false
		}
		if !(strings.Compare(val, tag.Value) == 0) {
			return false
		}
	}
	return true
}

func getVMInfoFromInstance(i Instance) (models.TargetType, error) {
	targetType := models.TargetType{}
	err := targetType.FromVMInfo(models.VMInfo{
		Image:            i.Image,
		InstanceID:       i.ID,
		InstanceProvider: utils.PointerTo(AWSProvider),
		InstanceType:     i.InstanceType,
		LaunchTime:       i.LaunchTime,
		Location:         i.Location(),
		ObjectType:       "VMInfo",
		Platform:         i.Platform,
		SecurityGroups:   utils.PointerTo(i.SecurityGroups),
		Tags:             utils.PointerTo(i.Tags),
	})
	if err != nil {
		err = fmt.Errorf("failed to create TargetType from VMInfo: %w", err)
	}

	return targetType, err
}
