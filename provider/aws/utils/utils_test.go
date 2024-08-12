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
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/openclarity/vmclarity/core/to"
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
									InstanceId: to.Ptr("instance-1"),
								},
								{
									InstanceId: to.Ptr("instance-2"),
								},
							},
						},
						{
							Instances: []ec2types.Instance{
								{
									InstanceId: to.Ptr("instance-3"),
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
									InstanceId: to.Ptr("instance-1"),
								},
								{
									InstanceId: to.Ptr("instance-2"),
									State: &ec2types.InstanceState{
										Name: ec2types.InstanceStateNamePending,
									},
								},
							},
						},
						{
							Instances: []ec2types.Instance{
								{
									InstanceId: to.Ptr("instance-3"),
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
									InstanceId: to.Ptr("instance-1"),
								},
								{
									InstanceId: to.Ptr("instance-2"),
									State: &ec2types.InstanceState{
										Name: ec2types.InstanceStateNamePending,
									},
								},
							},
						},
						{
							Instances: []ec2types.Instance{
								{
									InstanceId: to.Ptr("instance-3"),
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
			if got := GetInstanceState(tt.args.result, tt.args.instanceID); got != tt.want {
				t.Errorf("GetInstanceState() = %v, want %v", got, tt.want)
			}
		})
	}
}
