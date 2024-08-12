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
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/provider/aws/types"
)

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
		want   []types.Instance
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
									InstanceId: to.Ptr("instance-1"),
									Tags: []ec2types.Tag{
										{
											Key:   to.Ptr("key-1"),
											Value: to.Ptr("val-1"),
										},
									},
									State: &ec2types.InstanceState{
										Name: ec2types.InstanceStateNameRunning,
									},
									VpcId:   to.Ptr("vpc1"),
									ImageId: to.Ptr("image1"),
									Placement: &ec2types.Placement{
										AvailabilityZone: to.Ptr("az1"),
									},
									InstanceType:    "t2.large",
									PlatformDetails: to.Ptr("linux"),
									LaunchTime:      to.Ptr(launchTime),
									SecurityGroups: []ec2types.GroupIdentifier{
										{
											GroupId: to.Ptr("group1"),
										},
									},
								},
								{
									InstanceId: to.Ptr("instance-2"),
									Tags: []ec2types.Tag{
										{
											Key:   to.Ptr("key-2"),
											Value: to.Ptr("val-2"),
										},
									},
									State: &ec2types.InstanceState{
										Name: ec2types.InstanceStateNameRunning,
									},
									VpcId:   to.Ptr("vpc2"),
									ImageId: to.Ptr("image2"),
									Placement: &ec2types.Placement{
										AvailabilityZone: to.Ptr("az2"),
									},
									InstanceType:    "t2.large",
									PlatformDetails: to.Ptr("linux"),
									LaunchTime:      to.Ptr(launchTime),
									SecurityGroups: []ec2types.GroupIdentifier{
										{
											GroupId: to.Ptr("group2"),
										},
									},
								},
							},
						},
						{
							Instances: []ec2types.Instance{
								{
									InstanceId: to.Ptr("instance-3"),
									VpcId:      to.Ptr("vpc3"),
									ImageId:    to.Ptr("image3"),
									Placement: &ec2types.Placement{
										AvailabilityZone: to.Ptr("az3"),
									},
									State: &ec2types.InstanceState{
										Name: ec2types.InstanceStateNameRunning,
									},
									InstanceType:    "t2.large",
									PlatformDetails: to.Ptr("linux"),
									LaunchTime:      to.Ptr(launchTime),
									SecurityGroups: []ec2types.GroupIdentifier{
										{
											GroupId: to.Ptr("group3"),
										},
									},
								},
							},
						},
					},
				},
				regionID: "region-1",
			},
			want: []types.Instance{
				{
					ID:     "instance-1",
					Region: "region-1",
					VpcID:  "vpc1",
					SecurityGroups: []apitypes.SecurityGroup{
						{Id: "group1"},
					},
					AvailabilityZone: "az1",
					Image:            "image1",
					InstanceType:     "t2.large",
					Platform:         "linux",
					Tags: []apitypes.Tag{
						{
							Key:   "key-1",
							Value: "val-1",
						},
					},
					LaunchTime:          launchTime,
					RootVolumeEncrypted: apitypes.RootVolumeEncryptedUnknown,
				},
				{
					ID:     "instance-2",
					Region: "region-1",
					VpcID:  "vpc2",
					SecurityGroups: []apitypes.SecurityGroup{
						{Id: "group2"},
					},
					AvailabilityZone: "az2",
					Image:            "image2",
					InstanceType:     "t2.large",
					Platform:         "linux",
					Tags: []apitypes.Tag{
						{
							Key:   "key-2",
							Value: "val-2",
						},
					},
					LaunchTime:          launchTime,
					RootVolumeEncrypted: apitypes.RootVolumeEncryptedUnknown,
				},
				{
					ID:     "instance-3",
					Region: "region-1",
					VpcID:  "vpc3",
					SecurityGroups: []apitypes.SecurityGroup{
						{Id: "group3"},
					},
					AvailabilityZone:    "az3",
					Image:               "image3",
					InstanceType:        "t2.large",
					Platform:            "linux",
					Tags:                nil,
					LaunchTime:          launchTime,
					RootVolumeEncrypted: apitypes.RootVolumeEncryptedUnknown,
				},
			},
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Discoverer{}
			got := d.getInstancesFromDescribeInstancesOutput(ctx, tt.args.result, tt.args.regionID)

			var gotInstances []types.Instance
			for _, instance := range got {
				instance.Ec2Client = nil
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
		want apitypes.RootVolumeEncrypted
	}{
		{
			name: "unknown",
			args: args{
				encrypted: nil,
			},
			want: apitypes.RootVolumeEncryptedUnknown,
		},
		{
			name: "no",
			args: args{
				encrypted: to.Ptr(false),
			},
			want: apitypes.RootVolumeEncryptedNo,
		},
		{
			name: "yes",
			args: args{
				encrypted: to.Ptr(true),
			},
			want: apitypes.RootVolumeEncryptedYes,
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
