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
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
)

func Test_createVPCFilters(t *testing.T) {
	var (
		vpcID = "vpc-1"
		sgID1 = "sg-1"
		sgID2 = "sg-2"
	)

	type args struct {
		vpc VPC
	}
	tests := []struct {
		name string
		args args
		want []ec2types.Filter
	}{
		{
			name: "vpc with no security group",
			args: args{
				vpc: VPC{
					ID:             vpcID,
					SecurityGroups: nil,
				},
			},
			want: []ec2types.Filter{
				{
					Name:   utils.PointerTo(VpcIDFilterName),
					Values: []string{vpcID},
				},
			},
		},
		{
			name: "vpc with one security group",
			args: args{
				vpc: VPC{
					ID: vpcID,
					SecurityGroups: []models.SecurityGroup{
						{
							Id: sgID1,
						},
					},
				},
			},
			want: []ec2types.Filter{
				{
					Name:   utils.PointerTo(VpcIDFilterName),
					Values: []string{vpcID},
				},
				{
					Name:   utils.PointerTo(SecurityGroupIDFilterName),
					Values: []string{sgID1},
				},
			},
		},
		{
			name: "vpc with two security groups",
			args: args{
				vpc: VPC{
					ID: vpcID,
					SecurityGroups: []models.SecurityGroup{
						{
							Id: sgID1,
						},
						{
							Id: sgID2,
						},
					},
				},
			},
			want: []ec2types.Filter{
				{
					Name:   utils.PointerTo(VpcIDFilterName),
					Values: []string{vpcID},
				},
				{
					Name:   utils.PointerTo(SecurityGroupIDFilterName),
					Values: []string{sgID1, sgID2},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createVPCFilters(tt.args.vpc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createVPCFilters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createInclusionTagsFilters(t *testing.T) {
	var (
		tagName       = "foo"
		filterTagName = "tag:" + tagName
		tagVal        = "bar"
	)

	type args struct {
		tags []models.Tag
	}
	tests := []struct {
		name string
		args args
		want []ec2types.Filter
	}{
		{
			name: "no tags",
			args: args{
				tags: nil,
			},
			want: []ec2types.Filter{},
		},
		{
			name: "1 tag",
			args: args{
				tags: []models.Tag{
					{
						Key:   tagName,
						Value: tagVal,
					},
				},
			},
			want: []ec2types.Filter{
				{
					Name:   &filterTagName,
					Values: []string{tagVal},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createInclusionTagsFilters(tt.args.tags); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createInclusionTagsFilters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hasExcludedTags(t *testing.T) {
	var (
		tagName1 = "foo1"
		tagName2 = "foo2"
		tagVal1  = "bar1"
		tagVal2  = "bar2"
	)

	type args struct {
		excludeTags  []models.Tag
		instanceTags []ec2types.Tag
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "instance has no tags",
			args: args{
				excludeTags: []models.Tag{
					{
						Key:   tagName1,
						Value: tagVal1,
					},
					{
						Key:   "stam1",
						Value: "stam2",
					},
				},
				instanceTags: nil,
			},
			want: false,
		},
		{
			name: "empty excluded tags",
			args: args{
				excludeTags: nil,
				instanceTags: []ec2types.Tag{
					{
						Key:   &tagName1,
						Value: &tagVal1,
					},
					{
						Key:   &tagName2,
						Value: &tagVal2,
					},
				},
			},
			want: false,
		},
		{
			name: "instance does not have ALL the excluded tags (partial matching)",
			args: args{
				excludeTags: []models.Tag{
					{
						Key:   tagName1,
						Value: tagVal1,
					},
					{
						Key:   "stam1",
						Value: "stam2",
					},
				},
				instanceTags: []ec2types.Tag{
					{
						Key:   &tagName1,
						Value: &tagVal1,
					},
					{
						Key:   &tagName2,
						Value: &tagVal2,
					},
				},
			},
			want: false,
		},
		{
			name: "instance has ALL excluded tags",
			args: args{
				excludeTags: []models.Tag{
					{
						Key:   tagName1,
						Value: tagVal1,
					},
					{
						Key:   tagName2,
						Value: tagVal2,
					},
				},
				instanceTags: []ec2types.Tag{
					{
						Key:   &tagName1,
						Value: &tagVal1,
					},
					{
						Key:   &tagName2,
						Value: &tagVal2,
					},
					{
						Key:   utils.StringPtr("stam"),
						Value: utils.StringPtr("stam"),
					},
				},
			},
			want: true,
		},
		{
			name: "instance does not have excluded tags at all",
			args: args{
				excludeTags: []models.Tag{
					{
						Key:   "stam1",
						Value: "stam2",
					},
					{
						Key:   "stam3",
						Value: "stam4",
					},
				},
				instanceTags: []ec2types.Tag{
					{
						Key:   &tagName1,
						Value: &tagVal1,
					},
					{
						Key:   &tagName2,
						Value: &tagVal2,
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasExcludeTags(tt.args.excludeTags, tt.args.instanceTags); got != tt.want {
				t.Errorf("hasExcludeTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getInstanceState(t *testing.T) {
	type args struct {
		result     *ec2.DescribeInstancesOutput
		instanceID string
	}
	tests := []struct {
		name string
		args args
		want ec2types.InstanceStateName
	}{
		{
			name: "state running",
			args: args{
				result: &ec2.DescribeInstancesOutput{
					Reservations: []ec2types.Reservation{
						{
							Instances: []ec2types.Instance{
								{
									InstanceId: utils.StringPtr("instance-1"),
								},
								{
									InstanceId: utils.StringPtr("instance-2"),
								},
							},
						},
						{
							Instances: []ec2types.Instance{
								{
									InstanceId: utils.StringPtr("instance-3"),
									State: &ec2types.InstanceState{
										Name: ec2types.InstanceStateNameRunning,
									},
								},
							},
						},
					},
				},
				instanceID: "instance-3",
			},
			want: ec2types.InstanceStateNameRunning,
		},
		{
			name: "state pending",
			args: args{
				result: &ec2.DescribeInstancesOutput{
					Reservations: []ec2types.Reservation{
						{
							Instances: []ec2types.Instance{
								{
									InstanceId: utils.StringPtr("instance-1"),
								},
								{
									InstanceId: utils.StringPtr("instance-2"),
									State: &ec2types.InstanceState{
										Name: ec2types.InstanceStateNamePending,
									},
								},
							},
						},
						{
							Instances: []ec2types.Instance{
								{
									InstanceId: utils.StringPtr("instance-3"),
									State: &ec2types.InstanceState{
										Name: ec2types.InstanceStateNameRunning,
									},
								},
							},
						},
					},
				},
				instanceID: "instance-2",
			},
			want: ec2types.InstanceStateNamePending,
		},
		{
			name: "instance id not found",
			args: args{
				result: &ec2.DescribeInstancesOutput{
					Reservations: []ec2types.Reservation{
						{
							Instances: []ec2types.Instance{
								{
									InstanceId: utils.StringPtr("instance-1"),
								},
								{
									InstanceId: utils.StringPtr("instance-2"),
									State: &ec2types.InstanceState{
										Name: ec2types.InstanceStateNamePending,
									},
								},
							},
						},
						{
							Instances: []ec2types.Instance{
								{
									InstanceId: utils.StringPtr("instance-3"),
									State: &ec2types.InstanceState{
										Name: ec2types.InstanceStateNameRunning,
									},
								},
							},
						},
					},
				},
				instanceID: "instance-4",
			},
			want: ec2types.InstanceStateNamePending,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getInstanceState(tt.args.result, tt.args.instanceID); got != tt.want {
				t.Errorf("getInstanceState() = %v, want %v", got, tt.want)
			}
		})
	}
}

// nolint: maintidx
func TestClient_getInstancesFromDescribeInstancesOutput(t *testing.T) {
	launchTime := time.Now()

	type fields struct{}
	type args struct {
		result      *ec2.DescribeInstancesOutput
		excludeTags []models.Tag
		regionID    string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []Instance
	}{
		{
			name: "no reservations found",
			args: args{
				result: &ec2.DescribeInstancesOutput{
					Reservations: []ec2types.Reservation{},
				},
				excludeTags: nil,
				regionID:    "region-1",
			},
			want: nil,
		},
		{
			name: "no excluded tags",
			args: args{
				result: &ec2.DescribeInstancesOutput{
					Reservations: []ec2types.Reservation{
						{
							Instances: []ec2types.Instance{
								{
									InstanceId: utils.StringPtr("instance-1"),
									Tags: []ec2types.Tag{
										{
											Key:   utils.StringPtr("key-1"),
											Value: utils.StringPtr("val-1"),
										},
									},
									VpcId:   utils.PointerTo("vpc1"),
									ImageId: utils.PointerTo("image1"),
									Placement: &ec2types.Placement{
										AvailabilityZone: utils.PointerTo("az1"),
									},
									InstanceType:    "t2.large",
									PlatformDetails: utils.PointerTo("linux"),
									LaunchTime:      utils.PointerTo(launchTime),
									SecurityGroups: []ec2types.GroupIdentifier{
										{
											GroupId: utils.PointerTo("group1"),
										},
									},
								},
								{
									InstanceId: utils.StringPtr("instance-2"),
									Tags: []ec2types.Tag{
										{
											Key:   utils.StringPtr("key-2"),
											Value: utils.StringPtr("val-2"),
										},
									},
									VpcId:   utils.PointerTo("vpc2"),
									ImageId: utils.PointerTo("image2"),
									Placement: &ec2types.Placement{
										AvailabilityZone: utils.PointerTo("az2"),
									},
									InstanceType:    "t2.large",
									PlatformDetails: utils.PointerTo("linux"),
									LaunchTime:      utils.PointerTo(launchTime),
									SecurityGroups: []ec2types.GroupIdentifier{
										{
											GroupId: utils.PointerTo("group2"),
										},
									},
								},
							},
						},
						{
							Instances: []ec2types.Instance{
								{
									InstanceId: utils.StringPtr("instance-3"),
									VpcId:      utils.PointerTo("vpc3"),
									ImageId:    utils.PointerTo("image3"),
									Placement: &ec2types.Placement{
										AvailabilityZone: utils.PointerTo("az3"),
									},
									InstanceType:    "t2.large",
									PlatformDetails: utils.PointerTo("linux"),
									LaunchTime:      utils.PointerTo(launchTime),
									SecurityGroups: []ec2types.GroupIdentifier{
										{
											GroupId: utils.PointerTo("group3"),
										},
									},
								},
							},
						},
					},
				},
				excludeTags: nil,
				regionID:    "region-1",
			},
			want: []Instance{
				{
					ID:     "instance-1",
					Region: "region-1",
					VpcID:  "vpc1",
					SecurityGroups: []models.SecurityGroup{
						{Id: "group1"},
					},
					AvailabilityZone: "az1",
					Image:            "image1",
					InstanceType:     "t2.large",
					Platform:         "linux",
					Tags: []models.Tag{
						{
							Key:   "key-1",
							Value: "val-1",
						},
					},
					LaunchTime: launchTime,
				},
				{
					ID:     "instance-2",
					Region: "region-1",
					VpcID:  "vpc2",
					SecurityGroups: []models.SecurityGroup{
						{Id: "group2"},
					},
					AvailabilityZone: "az2",
					Image:            "image2",
					InstanceType:     "t2.large",
					Platform:         "linux",
					Tags: []models.Tag{
						{
							Key:   "key-2",
							Value: "val-2",
						},
					},
					LaunchTime: launchTime,
				},
				{
					ID:     "instance-3",
					Region: "region-1",
					VpcID:  "vpc3",
					SecurityGroups: []models.SecurityGroup{
						{Id: "group3"},
					},
					AvailabilityZone: "az3",
					Image:            "image3",
					InstanceType:     "t2.large",
					Platform:         "linux",
					Tags:             nil,
					LaunchTime:       launchTime,
				},
			},
		},
		{
			name: "one excluded instance",
			args: args{
				result: &ec2.DescribeInstancesOutput{
					Reservations: []ec2types.Reservation{
						{
							Instances: []ec2types.Instance{
								{
									InstanceId: utils.StringPtr("instance-1"),
									Tags: []ec2types.Tag{
										{
											Key:   utils.StringPtr("key-1"),
											Value: utils.StringPtr("val-1"),
										},
									},
									VpcId:   utils.PointerTo("vpc1"),
									ImageId: utils.PointerTo("image1"),
									Placement: &ec2types.Placement{
										AvailabilityZone: utils.PointerTo("az1"),
									},
									InstanceType:    "t2.large",
									PlatformDetails: utils.PointerTo("linux"),
									LaunchTime:      utils.PointerTo(launchTime),
									SecurityGroups: []ec2types.GroupIdentifier{
										{
											GroupId: utils.PointerTo("group1"),
										},
									},
								},
								{
									InstanceId: utils.StringPtr("instance-2"),
									Tags: []ec2types.Tag{
										{
											Key:   utils.StringPtr("key-2"),
											Value: utils.StringPtr("val-2"),
										},
									},
									VpcId:   utils.PointerTo("vpc2"),
									ImageId: utils.PointerTo("image2"),
									Placement: &ec2types.Placement{
										AvailabilityZone: utils.PointerTo("az2"),
									},
									InstanceType:    "t2.large",
									PlatformDetails: utils.PointerTo("linux"),
									LaunchTime:      utils.PointerTo(launchTime),
									SecurityGroups: []ec2types.GroupIdentifier{
										{
											GroupId: utils.PointerTo("group2"),
										},
									},
								},
							},
						},
						{
							Instances: []ec2types.Instance{
								{
									InstanceId: utils.StringPtr("instance-3"),
									VpcId:      utils.PointerTo("vpc3"),
									ImageId:    utils.PointerTo("image3"),
									Placement: &ec2types.Placement{
										AvailabilityZone: utils.PointerTo("az3"),
									},
									InstanceType:    "t2.large",
									PlatformDetails: utils.PointerTo("linux"),
									LaunchTime:      utils.PointerTo(launchTime),
									SecurityGroups: []ec2types.GroupIdentifier{
										{
											GroupId: utils.PointerTo("group3"),
										},
									},
								},
							},
						},
					},
				},
				excludeTags: []models.Tag{
					{
						Key:   "key-1",
						Value: "val-1",
					},
				},
				regionID: "region-1",
			},
			want: []Instance{
				{
					ID:     "instance-2",
					Region: "region-1",
					VpcID:  "vpc2",
					SecurityGroups: []models.SecurityGroup{
						{Id: "group2"},
					},
					AvailabilityZone: "az2",
					Image:            "image2",
					InstanceType:     "t2.large",
					Platform:         "linux",
					Tags: []models.Tag{
						{
							Key:   "key-2",
							Value: "val-2",
						},
					},
					LaunchTime: launchTime,
				},
				{
					ID:     "instance-3",
					Region: "region-1",
					VpcID:  "vpc3",
					SecurityGroups: []models.SecurityGroup{
						{Id: "group3"},
					},
					AvailabilityZone: "az3",
					Image:            "image3",
					InstanceType:     "t2.large",
					Platform:         "linux",
					Tags:             nil,
					LaunchTime:       launchTime,
				},
			},
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{}
			got := c.getInstancesFromDescribeInstancesOutput(ctx, tt.args.result, tt.args.excludeTags, tt.args.regionID)

			var gotInstances []Instance
			for _, instance := range got {
				instance.ec2Client = nil
				gotInstances = append(gotInstances, instance)
			}
			sort.Slice(gotInstances, func(i, j int) bool {
				return gotInstances[i].ID > gotInstances[j].ID
			})
			sort.Slice(tt.want, func(i, j int) bool {
				return tt.want[i].ID > tt.want[j].ID
			})

			if !reflect.DeepEqual(gotInstances, tt.want) {
				t.Errorf("getInstancesFromDescribeInstancesOutput() = %v, want %v", gotInstances, tt.want)
			}
		})
	}
}
