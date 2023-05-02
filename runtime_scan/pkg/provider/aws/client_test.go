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
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/openclarity/vmclarity/runtime_scan/pkg/types"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
)

func Test_createVPCFilters(t *testing.T) {
	var (
		vpcID = "vpc-1"
		sgID1 = "sg-1"
		sgID2 = "sg-2"

		vpcIDFilterName = vpcIDFilterName
		sgIDFilterName  = sgIDFilterName
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
					Name:   &vpcIDFilterName,
					Values: []string{vpcID},
				},
			},
		},
		{
			name: "vpc with one security group",
			args: args{
				vpc: VPC{
					ID: vpcID,
					SecurityGroups: []SecurityGroup{
						{
							ID: sgID1,
						},
					},
				},
			},
			want: []ec2types.Filter{
				{
					Name:   &vpcIDFilterName,
					Values: []string{vpcID},
				},
				{
					Name:   &sgIDFilterName,
					Values: []string{sgID1},
				},
			},
		},
		{
			name: "vpc with two security groups",
			args: args{
				vpc: VPC{
					ID: vpcID,
					SecurityGroups: []SecurityGroup{
						{
							ID: sgID1,
						},
						{
							ID: sgID2,
						},
					},
				},
			},
			want: []ec2types.Filter{
				{
					Name:   &vpcIDFilterName,
					Values: []string{vpcID},
				},
				{
					Name:   &sgIDFilterName,
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
		tags []types.Tag
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
			want: nil,
		},
		{
			name: "1 tag",
			args: args{
				tags: []types.Tag{
					{
						Key: tagName,
						Val: tagVal,
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
		excludeTags  []types.Tag
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
				excludeTags: []types.Tag{
					{
						Key: tagName1,
						Val: tagVal1,
					},
					{
						Key: "stam1",
						Val: "stam2",
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
				excludeTags: []types.Tag{
					{
						Key: tagName1,
						Val: tagVal1,
					},
					{
						Key: "stam1",
						Val: "stam2",
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
				excludeTags: []types.Tag{
					{
						Key: tagName1,
						Val: tagVal1,
					},
					{
						Key: tagName2,
						Val: tagVal2,
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
				excludeTags: []types.Tag{
					{
						Key: "stam1",
						Val: "stam2",
					},
					{
						Key: "stam3",
						Val: "stam4",
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

func Test_createInstanceStateFilters(t *testing.T) {
	type args struct {
		scanStopped bool
	}
	tests := []struct {
		name string
		args args
		want []ec2types.Filter
	}{
		{
			name: "should scan stopped",
			args: args{
				scanStopped: true,
			},
			want: []ec2types.Filter{
				{
					Name:   utils.StringPtr(instanceStateFilterName),
					Values: []string{"running", "stopped"},
				},
			},
		},
		{
			name: "should not scan stopped",
			args: args{
				scanStopped: false,
			},
			want: []ec2types.Filter{
				{
					Name:   utils.StringPtr(instanceStateFilterName),
					Values: []string{"running"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createInstanceStateFilters(tt.args.scanStopped); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createInstanceStateFilters() = %v, want %v", got, tt.want)
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
		excludeTags []types.Tag
		regionID    string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*InstanceImpl
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
			want: []*InstanceImpl{
				{
					id:               "instance-1",
					region:           "region-1",
					vpcID:            "vpc1",
					securityGroups:   []string{"group1"},
					availabilityZone: "az1",
					image:            "image1",
					ec2Type:          "t2.large",
					platform:         "linux",
					tags: []types.Tag{
						{
							Key: "key-1",
							Val: "val-1",
						},
					},
					launchTime: launchTime,
				},
				{
					id:               "instance-2",
					region:           "region-1",
					vpcID:            "vpc2",
					securityGroups:   []string{"group2"},
					availabilityZone: "az2",
					image:            "image2",
					ec2Type:          "t2.large",
					platform:         "linux",
					tags: []types.Tag{
						{
							Key: "key-2",
							Val: "val-2",
						},
					},
					launchTime: launchTime,
				},
				{
					id:               "instance-3",
					region:           "region-1",
					vpcID:            "vpc3",
					securityGroups:   []string{"group3"},
					availabilityZone: "az3",
					image:            "image3",
					ec2Type:          "t2.large",
					platform:         "linux",
					tags:             nil,
					launchTime:       launchTime,
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
				excludeTags: []types.Tag{
					{
						Key: "key-1",
						Val: "val-1",
					},
				},
				regionID: "region-1",
			},
			want: []*InstanceImpl{
				{
					id:               "instance-2",
					region:           "region-1",
					vpcID:            "vpc2",
					securityGroups:   []string{"group2"},
					availabilityZone: "az2",
					image:            "image2",
					ec2Type:          "t2.large",
					platform:         "linux",
					tags: []types.Tag{
						{
							Key: "key-2",
							Val: "val-2",
						},
					},
					launchTime: launchTime,
				},
				{
					id:               "instance-3",
					region:           "region-1",
					vpcID:            "vpc3",
					securityGroups:   []string{"group3"},
					availabilityZone: "az3",
					image:            "image3",
					ec2Type:          "t2.large",
					platform:         "linux",
					tags:             nil,
					launchTime:       launchTime,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{}
			got := c.getInstancesFromDescribeInstancesOutput(tt.args.result, tt.args.excludeTags, tt.args.regionID)

			var gotInstances []*InstanceImpl
			for _, instance := range got {
				var instanceImpl *InstanceImpl
				var ok bool
				if instanceImpl, ok = instance.(*InstanceImpl); !ok {
					t.Errorf("failed to convert type")
				}
				instanceImpl.ec2Client = nil
				gotInstances = append(gotInstances, instanceImpl)
			}
			sort.Slice(gotInstances, func(i, j int) bool {
				return gotInstances[i].id > gotInstances[j].id
			})
			sort.Slice(tt.want, func(i, j int) bool {
				return tt.want[i].id > tt.want[j].id
			})

			if !reflect.DeepEqual(gotInstances, tt.want) {
				t.Errorf("getInstancesFromDescribeInstancesOutput() = %v, want %v", gotInstances, tt.want)
			}
		})
	}
}
