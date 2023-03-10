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
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/openclarity/vmclarity/runtime_scan/pkg/types"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

type VolumeImpl struct {
	ec2Client *ec2.Client
	id        string
	region    string
}

func (v *VolumeImpl) GetID() string {
	return v.id
}

func (v *VolumeImpl) TakeSnapshot(ctx context.Context) (types.Snapshot, error) {
	params := ec2.CreateSnapshotInput{
		VolumeId:    &v.id,
		Description: &snapshotDescription,
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeSnapshot,
				Tags:         vmclarityTags,
			},
		},
	}
	out, err := v.ec2Client.CreateSnapshot(ctx, &params, func(options *ec2.Options) {
		options.Region = v.region
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %v", err)
	}
	return &SnapshotImpl{
		ec2Client: v.ec2Client,
		id:        *out.SnapshotId,
		region:    v.region,
	}, nil
}

func (v *VolumeImpl) WaitForReady(ctx context.Context) error {
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), utils.DefaultResourceReadyWaitTimeoutMin*time.Minute)
	defer cancel()

	for {
		select {
		case <-time.After(utils.DefaultResourceReadyCheckIntervalSec * time.Second):
			out, err := v.ec2Client.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{
				VolumeIds: []string{v.id},
			}, func(options *ec2.Options) {
				options.Region = v.region
			})
			if err != nil {
				return fmt.Errorf("failed to describe volumes. volumeID=%v: %v", v.id, err)
			}
			if len(out.Volumes) != 1 {
				return fmt.Errorf("got unexcpected number of volumes (%v) with volume id %v. excpecting 1", len(out.Volumes), v.id)
			}
			if out.Volumes[0].State == ec2types.VolumeStateAvailable {
				return nil
			}
		case <-ctxWithTimeout.Done():
			return fmt.Errorf("waiting for volume ready was canceled: %v", ctx.Err())
		}
	}
}

func (v *VolumeImpl) WaitForAttached(ctx context.Context) error {
	// nolint:govet
	ctxWithTimeout, _ := context.WithTimeout(context.Background(), utils.DefaultResourceReadyWaitTimeoutMin*time.Minute)

	for {
		select {
		case <-time.After(utils.DefaultResourceReadyCheckIntervalSec * time.Second):
			out, err := v.ec2Client.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{
				VolumeIds: []string{v.id},
			}, func(options *ec2.Options) {
				options.Region = v.region
			})
			if err != nil {
				return fmt.Errorf("failed to describe volumes. volumeID=%v: %v", v.id, err)
			}
			if len(out.Volumes) != 1 {
				return fmt.Errorf("got unexcpected number of volumes (%v) with volume id %v. excpecting 1", len(out.Volumes), v.id)
			}
			if out.Volumes[0].Attachments[0].State == ec2types.VolumeAttachmentStateAttached {
				return nil
			}
		case <-ctxWithTimeout.Done():
			return fmt.Errorf("waiting for volume ready was canceled: %v", ctxWithTimeout.Err())
		}
	}
}

func (v *VolumeImpl) Delete(ctx context.Context) error {
	if v == nil {
		return nil
	}

	_, err := v.ec2Client.DeleteVolume(ctx, &ec2.DeleteVolumeInput{
		VolumeId: &v.id,
	}, func(options *ec2.Options) {
		options.Region = v.region
	})
	if err != nil {
		return fmt.Errorf("failed to delete volume: %v", err)
	}

	return nil
}
