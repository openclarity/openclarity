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
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

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
									InstanceId: utils.PointerTo("instance-1"),
								},
								{
									InstanceId: utils.PointerTo("instance-2"),
								},
							},
						},
						{
							Instances: []ec2types.Instance{
								{
									InstanceId: utils.PointerTo("instance-3"),
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
									InstanceId: utils.PointerTo("instance-1"),
								},
								{
									InstanceId: utils.PointerTo("instance-2"),
									State: &ec2types.InstanceState{
										Name: ec2types.InstanceStateNamePending,
									},
								},
							},
						},
						{
							Instances: []ec2types.Instance{
								{
									InstanceId: utils.PointerTo("instance-3"),
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
									InstanceId: utils.PointerTo("instance-1"),
								},
								{
									InstanceId: utils.PointerTo("instance-2"),
									State: &ec2types.InstanceState{
										Name: ec2types.InstanceStateNamePending,
									},
								},
							},
						},
						{
							Instances: []ec2types.Instance{
								{
									InstanceId: utils.PointerTo("instance-3"),
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
func TestProvider_getInstancesFromDescribeInstancesOutput(t *testing.T) {
	launchTime := time.Now()

	type fields struct{}
	type args struct {
		result   *ec2.DescribeInstancesOutput
		regionID string
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
				regionID: "region-1",
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
									InstanceId: utils.PointerTo("instance-1"),
									Tags: []ec2types.Tag{
										{
											Key:   utils.PointerTo("key-1"),
											Value: utils.PointerTo("val-1"),
										},
									},
									State: &ec2types.InstanceState{
										Name: ec2types.InstanceStateNameRunning,
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
									InstanceId: utils.PointerTo("instance-2"),
									Tags: []ec2types.Tag{
										{
											Key:   utils.PointerTo("key-2"),
											Value: utils.PointerTo("val-2"),
										},
									},
									State: &ec2types.InstanceState{
										Name: ec2types.InstanceStateNameRunning,
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
									InstanceId: utils.PointerTo("instance-3"),
									VpcId:      utils.PointerTo("vpc3"),
									ImageId:    utils.PointerTo("image3"),
									Placement: &ec2types.Placement{
										AvailabilityZone: utils.PointerTo("az3"),
									},
									State: &ec2types.InstanceState{
										Name: ec2types.InstanceStateNameRunning,
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
				regionID: "region-1",
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
					LaunchTime:          launchTime,
					RootVolumeEncrypted: models.RootVolumeEncryptedUnknown,
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
					LaunchTime:          launchTime,
					RootVolumeEncrypted: models.RootVolumeEncryptedUnknown,
				},
				{
					ID:     "instance-3",
					Region: "region-1",
					VpcID:  "vpc3",
					SecurityGroups: []models.SecurityGroup{
						{Id: "group3"},
					},
					AvailabilityZone:    "az3",
					Image:               "image3",
					InstanceType:        "t2.large",
					Platform:            "linux",
					Tags:                nil,
					LaunchTime:          launchTime,
					RootVolumeEncrypted: models.RootVolumeEncryptedUnknown,
				},
			},
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{}
			got := p.getInstancesFromDescribeInstancesOutput(ctx, tt.args.result, tt.args.regionID)

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

func Test_encryptedToAPI(t *testing.T) {
	type args struct {
		encrypted *bool
	}
	tests := []struct {
		name string
		args args
		want models.RootVolumeEncrypted
	}{
		{
			name: "unknown",
			args: args{
				encrypted: nil,
			},
			want: models.RootVolumeEncryptedUnknown,
		},
		{
			name: "no",
			args: args{
				encrypted: utils.PointerTo(false),
			},
			want: models.RootVolumeEncryptedNo,
		},
		{
			name: "yes",
			args: args{
				encrypted: utils.PointerTo(true),
			},
			want: models.RootVolumeEncryptedYes,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := encryptedToAPI(tt.args.encrypted); got != tt.want {
				t.Errorf("encryptedToAPI() = %v, want %v", got, tt.want)
			}
		})
	}
}
