// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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

package asset

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const (
	DefaultAmazonMachineImage     = "ami-03484a09b43a06725"
	DefaultInstanceType           = ec2types.InstanceTypeT2Micro
	DefaultInstanceRunningTimeout = 2 * time.Minute
)

// Create a new asset to be scanned in the test environment.
// This asset will be an EC2 instance tagged with {"scanconfig": "test"}.
type Asset struct {
	InstanceID string
}

func (a *Asset) Create(ctx context.Context, ec2Client *ec2.Client) error {
	// Create an EC2 instance
	ec2Output, err := ec2Client.RunInstances(
		ctx,
		&ec2.RunInstancesInput{
			ImageId:      aws.String(DefaultAmazonMachineImage),
			InstanceType: DefaultInstanceType,
			MinCount:     aws.Int32(1),
			MaxCount:     aws.Int32(1),
		})
	if err != nil {
		return fmt.Errorf("failed to create test instance: %w", err)
	}

	a.InstanceID = *ec2Output.Instances[0].InstanceId

	// Add tag to the created instance
	_, err = ec2Client.CreateTags(
		ctx,
		&ec2.CreateTagsInput{
			Resources: []string{a.InstanceID},
			Tags: []ec2types.Tag{
				{
					Key:   aws.String("scanconfig"),
					Value: aws.String("test"),
				},
			},
		})
	if err != nil {
		return fmt.Errorf("failed to create tags for instance: %w", err)
	}

	err = ec2.NewInstanceRunningWaiter(ec2Client).Wait(
		ctx,
		&ec2.DescribeInstancesInput{
			InstanceIds: []string{a.InstanceID},
		},
		DefaultInstanceRunningTimeout,
	)
	if err != nil {
		return fmt.Errorf("failed to wait for instance to be running: %w", err)
	}

	return nil
}

func (a *Asset) Delete(ctx context.Context, ec2Client *ec2.Client) error {
	if a.InstanceID == "" {
		return nil
	}

	// Delete the EC2 instance
	_, err := ec2Client.TerminateInstances(
		ctx,
		&ec2.TerminateInstancesInput{
			InstanceIds: []string{a.InstanceID},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to delete test instance: %w", err)
	}

	return nil
}
