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

package types

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/sirupsen/logrus"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/provider/aws/utils"
)

type Instance struct {
	ID                  string
	Region              string
	VpcID               string
	SecurityGroups      []apitypes.SecurityGroup
	AvailabilityZone    string
	Image               string
	InstanceType        string
	Platform            string
	Tags                []apitypes.Tag
	LaunchTime          time.Time
	RootDeviceName      string
	RootVolumeSizeGB    int32
	RootVolumeEncrypted apitypes.RootVolumeEncrypted
	Volumes             []Volume

	Metadata provider.ScanMetadata

	Ec2Client *ec2.Client
}

func (i *Instance) Location() string {
	return Location{
		Region: i.Region,
		Vpc:    i.VpcID,
	}.String()
}

func (i *Instance) RootVolume() *Volume {
	for _, vol := range i.Volumes {
		if vol.BlockDeviceName == i.RootDeviceName {
			return &vol
		}
	}

	return nil
}

func (i *Instance) IsReady(ctx context.Context) (bool, error) {
	var ready bool

	out, err := i.Ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{i.ID},
	}, func(options *ec2.Options) {
		options.Region = i.Region
	})
	if err != nil {
		return ready, fmt.Errorf("failed to get VM instance. InstanceID=%s: %w", i.ID, err)
	}

	state := utils.GetInstanceState(out, i.ID)
	if state == ec2types.InstanceStateNameRunning {
		ready = true
	}

	return ready, nil
}

func (i *Instance) Delete(ctx context.Context) error {
	if i == nil {
		return nil
	}

	_, err := i.Ec2Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{i.ID},
	}, func(options *ec2.Options) {
		options.Region = i.Region
	})
	if err != nil {
		return fmt.Errorf("failed to terminate instances: %w", err)
	}

	return nil
}

// nolint:cyclop
func (i *Instance) AttachVolume(ctx context.Context, volume *Volume, deviceName string) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithFields(logrus.Fields{
		"InstanceID": i.ID,
		"Operation":  "AttachVolume",
		"VolumeID":   volume.ID,
	})

	options := func(options *ec2.Options) {
		options.Region = volume.Region
	}

	descParams := &ec2.DescribeVolumesInput{
		VolumeIds: []string{volume.ID},
	}

	describeOut, err := i.Ec2Client.DescribeVolumes(ctx, descParams, options)
	if err != nil {
		return fmt.Errorf("failed to fetch volume. VolumeID=%s: %w", volume.ID, err)
	}

	logger.Tracef("Found %d volumes", len(describeOut.Volumes))

	for _, vol := range describeOut.Volumes {
		logger.WithFields(logrus.Fields{
			"VolumeState": vol.State,
		}).Trace("Found volume")

		switch vol.State {
		case ec2types.VolumeStateInUse:
			for _, attachment := range vol.Attachments {
				if *attachment.VolumeId == volume.ID && *attachment.InstanceId == i.ID {
					logger.Trace("Volume is already attached to the instance")
					return nil
				}
			}
		case ec2types.VolumeStateAvailable:
			logger.Trace("Attaching volume to instance")

			attachVolParams := &ec2.AttachVolumeInput{
				Device:     to.Ptr(deviceName),
				InstanceId: to.Ptr(i.ID),
				VolumeId:   to.Ptr(volume.ID),
			}
			_, err := i.Ec2Client.AttachVolume(ctx, attachVolParams, options)
			if err != nil {
				return fmt.Errorf("failed to attach volume: %w", err)
			}
			return nil
		case ec2types.VolumeStateDeleted, ec2types.VolumeStateDeleting, ec2types.VolumeStateError:
			return utils.FatalError{
				Err: fmt.Errorf("cannot attach volume with state: %s", vol.State),
			}
		case ec2types.VolumeStateCreating:
			return utils.RetryableError{
				Err:   fmt.Errorf("cannot attach volume with state: %s", vol.State),
				After: utils.VolumeReadynessAfter,
			}
		}
	}

	return utils.FatalError{
		Err: fmt.Errorf("failed to find volume. VolumeID=%s", volume.ID),
	}
}
