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

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/provider/aws/utils"
)

type Snapshot struct {
	ID       string
	Region   string
	Metadata provider.ScanMetadata
	VolumeID string

	ec2Client *ec2.Client
}

func (s *Snapshot) Copy(ctx context.Context, region string) (*Snapshot, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithFields(logrus.Fields{
		"SnapshotID":    s.ID,
		"Operation":     "Copy",
		"AssetVolumeID": s.VolumeID,
	})

	if s.Region == region {
		logger.Debugf("Copying snapshot is skipped. SourceRegion=%s AssetRegion=%s", s.Region, region)
		return s, nil
	}

	options := func(options *ec2.Options) {
		options.Region = region
	}

	ec2TagsForSnapshot := EC2TagsFromScanMetadata(s.Metadata)
	ec2TagsForSnapshot = append(ec2TagsForSnapshot, ec2types.Tag{
		Key:   to.Ptr(EC2TagKeyAssetVolumeID),
		Value: to.Ptr(s.VolumeID),
	})
	ec2Filters := utils.EC2FiltersFromEC2Tags(ec2TagsForSnapshot)

	describeParams := &ec2.DescribeSnapshotsInput{
		Filters: ec2Filters,
	}
	describeOut, err := s.ec2Client.DescribeSnapshots(ctx, describeParams, options)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch asset snapshots. AssetRegion=%s SourceSnapshot=%s SourceRegion=%s: %w",
			region, s.ID, s.Region, err)
	}

	if len(describeOut.Snapshots) > 1 {
		logger.Warnf("Multiple snapshots found for asset volume: %d", len(describeOut.Snapshots))
	}

	for _, snap := range describeOut.Snapshots {
		switch snap.State {
		case ec2types.SnapshotStateError, ec2types.SnapshotStateRecoverable:
			return nil, utils.FatalError{
				Err: fmt.Errorf("failed to copy volume snapshot with state: %s", snap.State),
			}
		case ec2types.SnapshotStateRecovering, ec2types.SnapshotStatePending, ec2types.SnapshotStateCompleted:
			return &Snapshot{
				ec2Client: s.ec2Client,
				ID:        *snap.SnapshotId,
				Region:    region,
				Metadata:  s.Metadata,
				VolumeID:  *snap.VolumeId,
			}, nil
		}
	}

	copySnapParams := &ec2.CopySnapshotInput{
		SourceRegion:     &s.Region,
		SourceSnapshotId: &s.ID,
		Description:      to.Ptr(EC2SnapshotDescription),
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeSnapshot,
				Tags:         ec2TagsForSnapshot,
			},
		},
	}

	snap, err := s.ec2Client.CopySnapshot(ctx, copySnapParams, options)
	if err != nil {
		return nil, fmt.Errorf("failed to copy snapshot between regions. SnapshotID=%s SourceRegion=%s DestinationRegion=%s: %w",
			s.ID, s.Region, region, err)
	}

	return &Snapshot{
		ec2Client: s.ec2Client,
		ID:        *snap.SnapshotId,
		Region:    region,
		Metadata:  s.Metadata,
		VolumeID:  s.VolumeID,
	}, nil
}

func (s *Snapshot) Delete(ctx context.Context) error {
	if s == nil {
		return nil
	}

	_, err := s.ec2Client.DeleteSnapshot(ctx, &ec2.DeleteSnapshotInput{
		SnapshotId: &s.ID,
	}, func(options *ec2.Options) {
		options.Region = s.Region
	})
	if err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	return nil
}

func (s *Snapshot) IsReady(ctx context.Context) (bool, error) {
	var ready bool

	out, err := s.ec2Client.DescribeSnapshots(ctx, &ec2.DescribeSnapshotsInput{
		SnapshotIds: []string{s.ID},
	}, func(options *ec2.Options) {
		options.Region = s.Region
	})
	if err != nil {
		return ready, fmt.Errorf("failed to describe snapshot. SnapshotID=%s: %w", s.ID, err)
	}

	for _, snap := range out.Snapshots {
		if snap.SnapshotId == nil {
			continue
		}
		if *snap.SnapshotId == s.ID && snap.State == ec2types.SnapshotStateCompleted {
			return true, nil
		}
	}

	return false, nil
}

func (s *Snapshot) CreateVolume(ctx context.Context, az string) (*Volume, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithFields(logrus.Fields{
		"SnapshotID":    s.ID,
		"Operation":     "CreateVolume",
		"AssetVolumeID": s.VolumeID,
	})

	options := func(options *ec2.Options) {
		options.Region = s.Region
	}

	ec2TagsForVolume := EC2TagsFromScanMetadata(s.Metadata)
	ec2TagsForVolume = append(ec2TagsForVolume, ec2types.Tag{
		Key:   to.Ptr(EC2TagKeyAssetVolumeID),
		Value: to.Ptr(s.VolumeID),
	})

	ec2Filters := utils.EC2FiltersFromEC2Tags(ec2TagsForVolume)
	ec2Filters = append(ec2Filters, ec2types.Filter{
		Name:   to.Ptr(SnapshotIDFilterName),
		Values: []string{s.ID},
	})

	descParams := &ec2.DescribeVolumesInput{
		Filters: ec2Filters,
	}

	describeOut, err := s.ec2Client.DescribeVolumes(ctx, descParams, options)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch volume for snapshot. SnapshotID=%s: %w", s.ID, err)
	}

	if len(describeOut.Volumes) > 1 {
		logger.Warnf("Multiple volumes found for snapshot: %d", len(describeOut.Volumes))
	}

	for _, vol := range describeOut.Volumes {
		switch vol.State {
		case ec2types.VolumeStateDeleted, ec2types.VolumeStateDeleting, ec2types.VolumeStateError:
			return nil, utils.FatalError{
				Err: fmt.Errorf("found volume in unexpected state. VolumeID=%s: %s", *vol.VolumeId, vol.State),
			}
		case ec2types.VolumeStateAvailable, ec2types.VolumeStateCreating, ec2types.VolumeStateInUse:
			return &Volume{
				Ec2Client: s.ec2Client,
				ID:        *vol.VolumeId,
				Region:    s.Region,
				Metadata:  s.Metadata,
			}, nil
		}
	}

	createParams := &ec2.CreateVolumeInput{
		AvailabilityZone: &az,
		SnapshotId:       &s.ID,
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeVolume,
				Tags:         ec2TagsForVolume,
			},
		},
		VolumeType: ec2types.VolumeTypeGp2,
	}
	out, err := s.ec2Client.CreateVolume(ctx, createParams, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume from snapshot. SnapshotID=%s: %w", s.ID, err)
	}

	return &Volume{
		Ec2Client: s.ec2Client,
		ID:        *out.VolumeId,
		Region:    s.Region,
		Metadata:  s.Metadata,
	}, nil
}
