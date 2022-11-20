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

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/openclarity/vmclarity/runtime_scan/pkg/types"
)

type VolumeImpl struct {
	ec2Client *ec2.Client
	id        string
	name      string
	region    string
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
