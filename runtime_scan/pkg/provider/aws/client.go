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
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	awstype "github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/cloudinit"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/config/aws"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/types"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
)

type Client struct {
	ec2Client *ec2.Client
	awsConfig *aws.Config
}

var (
	snapshotDescription = "VMClarity snapshot"
	tagKey              = "Owner"
	tagVal              = "VMClarity"
	vmclarityTags       = []ec2types.Tag{
		{
			Key:   &tagKey,
			Value: &tagVal,
		},
	}
	nameTagKey = "Name"
)

func Create(ctx context.Context, config *aws.Config) (*Client, error) {
	awsClient := Client{
		awsConfig: config,
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %v", err)
	}

	// nolint:contextcheck
	awsClient.ec2Client = ec2.NewFromConfig(cfg)

	return &awsClient, nil
}

func (c *Client) DiscoverScopes(ctx context.Context) (*models.Scopes, error) {
	regions, err := c.ListAllRegions(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to list all regions: %v", err)
	}

	scopes := models.ScopeType{}
	err = scopes.FromAwsAccountScope(models.AwsAccountScope{
		Regions: convertToAPIRegions(regions),
	})
	if err != nil {
		return nil, fmt.Errorf("FromAwsScope failed: %w", err)
	}

	return &models.Scopes{
		ScopeInfo: &scopes,
	}, nil
}

func (c *Client) DiscoverInstances(ctx context.Context, scanScope *models.ScanScopeType) ([]types.Instance, error) {
	var ret []types.Instance
	var filters []ec2types.Filter

	awsScanScope, err := scanScope.AsAwsScanScope()
	if err != nil {
		return nil, fmt.Errorf("failed to convert as aws scope: %v", err)
	}

	scope := convertFromAPIScanScope(&awsScanScope)

	regions, err := c.getRegionsToScan(ctx, scope)
	if err != nil {
		return nil, fmt.Errorf("failed to get regions to scan: %v", err)
	}
	if len(regions) == 0 {
		return nil, fmt.Errorf("no regions to scan")
	}
	filters = append(filters, createInclusionTagsFilters(scope.TagSelector)...)
	filters = append(filters, createInstanceStateFilters(scope.ScanStopped)...)

	for _, region := range regions {
		// if no vpcs, that mean that we don't need any vpc filters
		if len(region.VPCs) == 0 {
			instances, err := c.GetInstances(ctx, filters, scope.ExcludeTags, region.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to get instances: %v", err)
			}
			ret = append(ret, instances...)
			continue
		}

		// need to do a per vpc call for DescribeInstances
		for _, vpc := range region.VPCs {
			vpcFilters := append(filters, createVPCFilters(vpc)...)

			instances, err := c.GetInstances(ctx, vpcFilters, scope.ExcludeTags, region.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to get instances: %v", err)
			}
			ret = append(ret, instances...)
		}
	}
	return ret, nil
}

func convertFromAPIScanScope(scope *models.AwsScanScope) *ScanScope {
	return &ScanScope{
		AllRegions:  convertBool(scope.AllRegions),
		Regions:     convertFromAPIRegions(scope.Regions),
		ScanStopped: convertBool(scope.ShouldScanStoppedInstances),
		TagSelector: convertFromAPITags(scope.InstanceTagSelector),
		ExcludeTags: convertFromAPITags(scope.InstanceTagExclusion),
	}
}

func convertFromAPITags(tags *[]models.Tag) []Tag {
	var ret []Tag
	if tags != nil {
		for _, tag := range *tags {
			ret = append(ret, Tag{
				Key: tag.Key,
				Val: tag.Value,
			})
		}
	}

	return ret
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
			SecurityGroups: convertFromAPISecurityGroups(vpc.SecurityGroups),
		}
	}

	return ret
}

func convertFromAPISecurityGroups(securityGroups *[]models.AwsSecurityGroup) []SecurityGroup {
	if securityGroups == nil {
		return []SecurityGroup{}
	}
	ret := make([]SecurityGroup, len(*securityGroups))
	for i, securityGroup := range *securityGroups {
		ret[i] = SecurityGroup{
			ID: securityGroup.Id,
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
			SecurityGroups: convertToAPISecurityGroups(vpcs[i].SecurityGroups),
		}
	}

	return &ret
}

func convertToAPISecurityGroups(securityGroups []SecurityGroup) *[]models.AwsSecurityGroup {
	ret := make([]models.AwsSecurityGroup, len(securityGroups))
	for i := range securityGroups {
		ret[i] = models.AwsSecurityGroup{
			Id: securityGroups[i].ID,
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

func (c *Client) RunScanningJob(ctx context.Context, region, id string, config provider.ScanningJobConfig) (types.Instance, error) {
	cloudInitData := cloudinit.Data{
		ScannerCLIConfig: config.ScannerCLIConfig,
		ScannerImage:     config.ScannerImage,
		ServerAddress:    config.VMClarityAddress,
		ScanResultID:     config.ScanResultID,
	}
	userData, err := cloudinit.GenerateCloudInit(cloudInitData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate cloud-init: %v", err)
	}

	instanceTags := createInstanceTags(id)
	userDataBase64 := base64.StdEncoding.EncodeToString([]byte(userData))

	runInstancesInput := &ec2.RunInstancesInput{
		MaxCount:     utils.Int32Ptr(1),
		MinCount:     utils.Int32Ptr(1),
		ImageId:      &c.awsConfig.AmiID,
		InstanceType: ec2types.InstanceType(c.awsConfig.InstanceType),
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeInstance,
				Tags:         instanceTags,
			},
			{
				ResourceType: ec2types.ResourceTypeVolume,
				Tags:         vmclarityTags,
			},
		},
		UserData: &userDataBase64,
	}

	// Create network interface in the scanner subnet with the scanner security group.
	runInstancesInput.NetworkInterfaces = []ec2types.InstanceNetworkInterfaceSpecification{
		{
			AssociatePublicIpAddress: utils.BoolPtr(false),
			DeleteOnTermination:      utils.BoolPtr(true),
			DeviceIndex:              utils.Int32Ptr(0),
			Groups:                   []string{c.awsConfig.SecurityGroupID},
			SubnetId:                 &c.awsConfig.SubnetID,
		},
	}

	var retryMaxAttempts int
	// Use spot instances if there is a configuration for it.
	if config.ScannerInstanceCreationConfig != nil {
		if config.ScannerInstanceCreationConfig.UseSpotInstances {
			runInstancesInput.InstanceMarketOptions = &ec2types.InstanceMarketOptionsRequest{
				MarketType: ec2types.MarketTypeSpot,
				SpotOptions: &ec2types.SpotMarketOptions{
					InstanceInterruptionBehavior: ec2types.InstanceInterruptionBehaviorTerminate,
					SpotInstanceType:             ec2types.SpotInstanceTypeOneTime,
					MaxPrice:                     config.ScannerInstanceCreationConfig.MaxPrice,
				},
			}
		}
		// In the case of spot instances, we have higher probability to start an instance
		// by increasing RetryMaxAttempts
		if config.ScannerInstanceCreationConfig.RetryMaxAttempts != nil {
			retryMaxAttempts = *config.ScannerInstanceCreationConfig.RetryMaxAttempts
		}
	}

	if config.KeyPairName != "" {
		// Set a key-pair to the instance.
		runInstancesInput.KeyName = &config.KeyPairName
	}

	// if retryMaxAttempts value is 0 it will be ignored
	out, err := c.ec2Client.RunInstances(ctx, runInstancesInput, func(options *ec2.Options) {
		options.Region = region
		options.RetryMaxAttempts = retryMaxAttempts
		options.RetryMode = awstype.RetryModeStandard
	})
	if err != nil {
		return nil, fmt.Errorf("failed to run instances: %v", err)
	}

	return &InstanceImpl{
		ec2Client:        c.ec2Client,
		id:               *out.Instances[0].InstanceId,
		region:           region,
		availabilityZone: *out.Instances[0].Placement.AvailabilityZone,
	}, nil
}

func createInstanceTags(id string) []ec2types.Tag {
	nameTagValue := fmt.Sprintf("vmclarity-scanner-%s", id)

	var ret []ec2types.Tag
	ret = append(ret, vmclarityTags...)
	ret = append(ret, ec2types.Tag{
		Key:   &nameTagKey,
		Value: &nameTagValue,
	})

	return ret
}

func (c *Client) GetInstances(ctx context.Context, filters []ec2types.Filter, excludeTags []Tag, regionID string) ([]types.Instance, error) {
	ret := make([]types.Instance, 0)

	out, err := c.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters:    filters,
		MaxResults: utils.Int32Ptr(maxResults), // TODO what will be a good number?
	}, func(options *ec2.Options) {
		options.Region = regionID
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %v", err)
	}
	ret = append(ret, c.getInstancesFromDescribeInstancesOutput(out, excludeTags, regionID)...)

	// use pagination
	// TODO we can make it better by not saving all results in memory. See https://github.com/openclarity/vmclarity/pull/3#discussion_r1021656861
	for out.NextToken != nil {
		out, err = c.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
			Filters:    filters,
			MaxResults: utils.Int32Ptr(maxResults),
			NextToken:  out.NextToken,
		}, func(options *ec2.Options) {
			options.Region = regionID
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe instances: %v", err)
		}
		ret = append(ret, c.getInstancesFromDescribeInstancesOutput(out, excludeTags, regionID)...)
	}

	return ret, nil
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

func (c *Client) getInstancesFromDescribeInstancesOutput(result *ec2.DescribeInstancesOutput, excludeTags []Tag, regionID string) []types.Instance {
	var ret []types.Instance

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			if hasExcludeTags(excludeTags, instance.Tags) {
				continue
			}
			ret = append(ret, &InstanceImpl{
				ec2Client: c.ec2Client,
				id:        *instance.InstanceId,
				region:    regionID,
			})
		}
	}
	return ret
}

func getVPCSecurityGroupsIDs(vpc VPC) []string {
	sgs := make([]string, len(vpc.SecurityGroups))
	for i, sg := range vpc.SecurityGroups {
		sgs[i] = sg.ID
	}
	return sgs
}

const (
	vpcIDFilterName         = "vpc-id"
	sgIDFilterName          = "instance.group-id"
	instanceStateFilterName = "instance-state-name"
)

func createVPCFilters(vpc VPC) []ec2types.Filter {
	ret := make([]ec2types.Filter, 0)

	// create per vpc filters
	ret = append(ret, ec2types.Filter{
		Name:   utils.StringPtr(vpcIDFilterName),
		Values: []string{vpc.ID},
	})
	sgs := getVPCSecurityGroupsIDs(vpc)
	if len(sgs) > 0 {
		ret = append(ret, ec2types.Filter{
			Name:   utils.StringPtr(sgIDFilterName),
			Values: sgs,
		})
	}

	log.Infof("VPC filter created: %+v", ret)

	return ret
}

func createInstanceStateFilters(scanStopped bool) []ec2types.Filter {
	filters := make([]ec2types.Filter, 0)
	states := []string{"running"}
	if scanStopped {
		states = append(states, "stopped")
	}

	// TODO these are the states: pending | running | shutting-down | terminated | stopping | stopped
	// Do we want to scan any other state (other than running and stopped)
	filters = append(filters, ec2types.Filter{
		Name:   utils.StringPtr(instanceStateFilterName),
		Values: states,
	})
	return filters
}

func createInclusionTagsFilters(tags []Tag) []ec2types.Filter {
	// nolint:prealloc
	var filters []ec2types.Filter

	// If you specify multiple filters, the filters are joined with an AND, and the request returns
	// only results that match all of the specified filters.
	for _, tag := range tags {
		filters = append(filters, ec2types.Filter{
			Name:   utils.StringPtr("tag:" + tag.Key),
			Values: []string{tag.Val},
		})
	}

	return filters
}

func (c *Client) getRegionsToScan(ctx context.Context, scope *ScanScope) ([]Region, error) {
	if scope.AllRegions {
		return c.ListAllRegions(ctx, false)
	}

	return scope.Regions, nil
}

func (c *Client) ListAllRegions(ctx context.Context, isRecursive bool) ([]Region, error) {
	ret := make([]Region, 0)
	out, err := c.ec2Client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{
		AllRegions: nil, // display also disabled regions?
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe regions: %v", err)
	}

	for _, region := range out.Regions {
		ret = append(ret, Region{
			Name: *region.RegionName,
		})
	}

	if isRecursive {
		for i, region := range ret {
			// List region VPCs
			vpcs, err := c.ec2Client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
				MaxResults: utils.Int32Ptr(maxResults),
			}, func(options *ec2.Options) {
				options.Region = region.Name
			})
			if err != nil {
				log.Warnf("Failed to describe vpcs. region=%v: %v", region.Name, err)
				continue
			}
			ret[i].VPCs = convertAwsVPCs(vpcs.Vpcs)
			for i2, vpc := range ret[i].VPCs {
				// List VPC's security groups
				securityGroups, err := c.ec2Client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
					Filters: []ec2types.Filter{
						{
							Name:   utils.StringPtr(vpcIDFilterName),
							Values: []string{vpc.ID},
						},
					},
					MaxResults: utils.Int32Ptr(maxResults),
				}, func(options *ec2.Options) {
					options.Region = region.Name
				})
				if err != nil {
					log.Warnf("Failed to describe security groups. region=%v, vpc=%v: %v", region.Name, vpc.ID, err)
					continue
				}
				ret[i].VPCs[i2].SecurityGroups = convertAwsSecurityGroups(securityGroups.SecurityGroups)
			}
		}
	}

	return ret, nil
}

func convertAwsSecurityGroups(securityGroups []ec2types.SecurityGroup) []SecurityGroup {
	var ret []SecurityGroup
	for _, securityGroup := range securityGroups {
		if securityGroup.GroupId != nil {
			ret = append(ret, SecurityGroup{
				ID: *securityGroup.GroupId,
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
func hasExcludeTags(excludeTags []Tag, instanceTags []ec2types.Tag) bool {
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
		if !(strings.Compare(val, tag.Val) == 0) {
			return false
		}
	}
	return true
}
