// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/openclarity/vmclarity/runtime_scan/pkg/types"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

type InstanceImpl struct {
	ec2Client        *ec2.Client
	id               string
	region           string
	availabilityZone string
}

func (i *InstanceImpl) GetID() string {
	return i.id
}

func (i *InstanceImpl) GetLocation() string {
	return i.region
}

func (i *InstanceImpl) GetAvailabilityZone() string {
	return i.availabilityZone
}

func (i *InstanceImpl) GetRootVolume(ctx context.Context) (types.Volume, error) {
	out, err := i.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{i.id},
	}, func(options *ec2.Options) {
		options.Region = i.region
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %v", err)
	}

	if len(out.Reservations) == 0 {
		return nil, fmt.Errorf("no reservations were found")
	}
	if len(out.Reservations) > 1 {
		return nil, fmt.Errorf("more than one reservations were found")
	}
	if len(out.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("no instances were found")
	}
	if len(out.Reservations[0].Instances) > 1 {
		return nil, fmt.Errorf("more than one instances were found")
	}

	outInstance := out.Reservations[0].Instances[0]
	rootDeviceName := *outInstance.RootDeviceName

	// find root volume of the instance
	for _, blkDevice := range outInstance.BlockDeviceMappings {
		if strings.Compare(*blkDevice.DeviceName, rootDeviceName) == 0 {
			return &VolumeImpl{
				ec2Client: i.ec2Client,
				id:        *blkDevice.Ebs.VolumeId,
				region:    i.region,
			}, nil
		}
	}
	return nil, fmt.Errorf("failed to find root device volume")
}

func (i *InstanceImpl) WaitForReady(ctx context.Context) error {
	// nolint:govet
	ctxWithTimeout, _ := context.WithTimeout(context.Background(), utils.DefaultResourceReadyWaitTimeoutMin*time.Minute)

	for {
		select {
		case <-time.After(utils.DefaultResourceReadyCheckIntervalSec * time.Second):
			out, err := i.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
				InstanceIds: []string{i.id},
			}, func(options *ec2.Options) {
				options.Region = i.region
			})
			if err != nil {
				return fmt.Errorf("failed to describe instance. instanceID=%v: %v", i.id, err)
			}
			state := getInstanceState(out, i.id)
			if state == ec2types.InstanceStateNameRunning {
				return nil
			}
		case <-ctxWithTimeout.Done():
			return fmt.Errorf("timeout: %v", ctxWithTimeout.Err())
		}
	}
}

func (i *InstanceImpl) Delete(ctx context.Context) error {
	if i == nil {
		return nil
	}

	_, err := i.ec2Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{i.id},
	}, func(options *ec2.Options) {
		options.Region = i.region
	})
	if err != nil {
		return fmt.Errorf("failed to terminate instances: %v", err)
	}

	return nil
}

func (i *InstanceImpl) AttachVolume(ctx context.Context, volume types.Volume, deviceName string) error {
	_, err := i.ec2Client.AttachVolume(ctx, &ec2.AttachVolumeInput{
		Device:     utils.StringPtr(deviceName),
		InstanceId: utils.StringPtr(i.GetID()),
		VolumeId:   utils.StringPtr(volume.GetID()),
	}, func(options *ec2.Options) {
		options.Region = i.GetLocation()
	})
	if err != nil {
		return fmt.Errorf("failed to attach volume: %v", err)
	}

	return nil
}
