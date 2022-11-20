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
)

type SnapshotImpl struct {
	ec2Client *ec2.Client
	id        string
	region    string
}

func (s *SnapshotImpl) GetID() string {
	return s.id
}

func (s *SnapshotImpl) GetRegion() string {
	return s.region
}

func (s *SnapshotImpl) Copy(ctx context.Context, dstRegion string) (types.Snapshot, error) {
	snap, err := s.ec2Client.CopySnapshot(ctx, &ec2.CopySnapshotInput{
		SourceRegion:     &s.region,
		SourceSnapshotId: &s.id,
		Description:      &snapshotDescription,
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeSnapshot,
				Tags:         vmclarityTags,
			},
		},
	}, func(options *ec2.Options) {
		options.Region = dstRegion
	})
	if err != nil {
		return nil, fmt.Errorf("failed to copy snapshot: %v", err)
	}

	return &SnapshotImpl{
		ec2Client: s.ec2Client,
		id:        *snap.SnapshotId,
		region:    dstRegion,
	}, nil
}

func (s *SnapshotImpl) Delete(ctx context.Context) error {
	if s == nil {
		return nil
	}

	_, err := s.ec2Client.DeleteSnapshot(ctx, &ec2.DeleteSnapshotInput{
		SnapshotId: &s.id,
	}, func(options *ec2.Options) {
		options.Region = s.region
	})
	if err != nil {
		return fmt.Errorf("failed to delete snapshot: %v", err)
	}

	return nil
}

func (s *SnapshotImpl) WaitForReady(ctx context.Context) error {
	// nolint:govet
	ctxWithTimeout, _ := context.WithTimeout(context.Background(), waitTimeout*time.Minute)

	for {
		select {
		case <-time.After(checkInterval * time.Second):
			out, err := s.ec2Client.DescribeSnapshots(ctx, &ec2.DescribeSnapshotsInput{
				SnapshotIds: []string{s.id},
			}, func(options *ec2.Options) {
				options.Region = s.region
			})
			if err != nil {
				return fmt.Errorf("failed to describe snapshot. snapshotID=%v: %v", s.id, err)
			}
			if len(out.Snapshots) != 1 {
				return fmt.Errorf("got unexcpected number of snapshots (%v) with snapshot id %v. excpecting 1", len(out.Snapshots), s.id)
			}
			if out.Snapshots[0].State == ec2types.SnapshotStateCompleted {
				return nil
			}
		case <-ctxWithTimeout.Done():
			return fmt.Errorf("timeout: %v", ctxWithTimeout.Err())
		}
	}
}
