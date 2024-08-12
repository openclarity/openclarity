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
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/pkg/orchestrator/provider"
	"github.com/openclarity/vmclarity/pkg/shared/log"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

type Volume struct {
	ID     string
	Region string

	BlockDeviceName string
	Metadata        provider.ScanMetadata

	ec2Client *ec2.Client
}

func (v *Volume) CreateSnapshot(ctx context.Context) (*Snapshot, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithFields(logrus.Fields{
		"VolumeID":  v.ID,
		"Operation": "CreateSnapshot",
	})

	options := func(options *ec2.Options) {
		options.Region = v.Region
	}

	ec2TagsForSnapshot := EC2TagsFromScanMetadata(v.Metadata)
	ec2TagsForSnapshot = append(ec2TagsForSnapshot, ec2types.Tag{
		Key:   utils.PointerTo(EC2TagKeyAssetVolumeID),
		Value: utils.PointerTo(v.ID),
	})

	ec2Filters := EC2FiltersFromEC2Tags(ec2TagsForSnapshot)

	describeParams := &ec2.DescribeSnapshotsInput{
		Filters: ec2Filters,
	}
	describeOut, err := v.ec2Client.DescribeSnapshots(ctx, describeParams, options)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch snapshots for volume. VolumeID=%s: %w", v.ID, err)
	}

	if len(describeOut.Snapshots) > 1 {
		logger.Warnf("Multiple snapshots found for volume: %d", len(describeOut.Snapshots))
	}

	for _, snap := range describeOut.Snapshots {
		switch snap.State {
		case ec2types.SnapshotStateError, ec2types.SnapshotStateRecoverable:
			// We want to recreate the snapshot if it is in error or recoverable state. Cleanup will take care of
			// removing these as well.
		case ec2types.SnapshotStateRecovering, ec2types.SnapshotStatePending, ec2types.SnapshotStateCompleted:
			fallthrough
		default:
			return &Snapshot{
				ec2Client: v.ec2Client,
				ID:        *snap.SnapshotId,
				Region:    v.Region,
				Metadata:  v.Metadata,
				VolumeID:  v.ID,
			}, nil
		}
	}

	createParams := ec2.CreateSnapshotInput{
		VolumeId:    &v.ID,
		Description: utils.PointerTo(EC2SnapshotDescription),
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeSnapshot,
				Tags:         ec2TagsForSnapshot,
			},
		},
	}
	createOut, err := v.ec2Client.CreateSnapshot(ctx, &createParams, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot for volume. VolumeID=%s: %w", v.ID, err)
	}

	return &Snapshot{
		ec2Client: v.ec2Client,
		ID:        *createOut.SnapshotId,
		Region:    v.Region,
		Metadata:  v.Metadata,
		VolumeID:  v.ID,
	}, nil
}

func (v *Volume) WaitForReady(ctx context.Context, timeout time.Duration, interval time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	timer := time.NewTicker(interval)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			ready, err := v.IsReady(ctx)
			if err != nil {
				return fmt.Errorf("failed to get volume state. VolumeID=%s: %w", v.ID, err)
			}
			if ready {
				return nil
			}
		case <-ctx.Done():
			return fmt.Errorf("failed to wait until VM Instance is in ready state. VolumeID=%s: %w", v.ID, ctx.Err())
		}
	}
}

func (v *Volume) IsReady(ctx context.Context) (bool, error) {
	out, err := v.ec2Client.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{
		VolumeIds: []string{v.ID},
	}, func(options *ec2.Options) {
		options.Region = v.Region
	})
	if err != nil {
		return false, fmt.Errorf("failed to describe volumes. VolumeID=%s: %w", v.ID, err)
	}

	if len(out.Volumes) != 1 {
		return false, fmt.Errorf("got unexcpected number of volumes (%d). Excpecting 1. VolumeID=%s",
			len(out.Volumes), v.ID)
	}

	volState := out.Volumes[0].State
	switch volState {
	case ec2types.VolumeStateAvailable, ec2types.VolumeStateInUse:
		return true, nil
	case ec2types.VolumeStateCreating:
		return false, nil
	case ec2types.VolumeStateDeleted, ec2types.VolumeStateDeleting, ec2types.VolumeStateError:
		return false, FatalError{
			Err: fmt.Errorf("volume is not ready due to its state: %s", volState),
		}
	default:
	}

	return false, nil
}

func (v *Volume) IsAttached(ctx context.Context) (bool, error) {
	out, err := v.ec2Client.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{
		VolumeIds: []string{v.ID},
	}, func(options *ec2.Options) {
		options.Region = v.Region
	})
	if err != nil {
		return false, fmt.Errorf("failed to describe volumes. VolumeID=%s: %w", v.ID, err)
	}

	if len(out.Volumes) != 1 {
		return false, fmt.Errorf("got unexcpected number of volumes (%d). Excpecting 1. VolumeID=%s",
			len(out.Volumes), v.ID)
	}

	for _, attachment := range out.Volumes[0].Attachments {
		if attachment.State == ec2types.VolumeAttachmentStateAttached {
			return true, nil
		}
	}

	return false, nil
}

func (v *Volume) Delete(ctx context.Context) error {
	if v == nil {
		return nil
	}

	_, err := v.ec2Client.DeleteVolume(ctx, &ec2.DeleteVolumeInput{
		VolumeId: &v.ID,
	}, func(options *ec2.Options) {
		options.Region = v.Region
	})
	if err != nil {
		return fmt.Errorf("failed to delete volume: %w", err)
	}

	return nil
}
