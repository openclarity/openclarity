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

package aws

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cloudformationtypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/openclarity/vmclarity/installation"
	"github.com/openclarity/vmclarity/testenv/utils"
	dockerhelper "github.com/openclarity/vmclarity/testenv/utils/docker"
)

func (e *AWSEnv) prepareStack(ctx context.Context) error {
	var err error

	// Create a new key pair
	_, err = e.ec2Client.ImportKeyPair(ctx, &ec2.ImportKeyPairInput{
		KeyName:           &e.stackName,
		PublicKeyMaterial: e.sshKeyPair.PublicKey,
	})
	if err != nil {
		return fmt.Errorf("failed to import key pair: %w", err)
	}

	// Create a new S3 bucket
	_, err = e.s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: &e.stackName,
		CreateBucketConfiguration: &s3types.CreateBucketConfiguration{
			LocationConstraint: s3types.BucketLocationConstraint(e.region),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	f, err := installation.AWSManifestBundle.Open("VmClarity.cfn")
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Upload template file to S3 bucket
	_, err = e.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &e.stackName,
		Key:    &e.stackName,
		Body:   f,
	})
	if err != nil {
		return fmt.Errorf("failed to put object: %w", err)
	}
	e.templateURL = "https://" + e.stackName + ".s3.amazonaws.com/" + e.stackName

	return nil
}

func (e *AWSEnv) afterSetUp(ctx context.Context) error {
	ready, err := e.infrastructureReady(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if infrastructure is ready: %w", err)
	}
	if !ready {
		return errors.New("infrastructure is not ready")
	}

	e.sshPortForwardInput = &utils.SSHForwardInput{
		PrivateKey:    e.sshKeyPair.PrivateKey,
		User:          DefaultRemoteUser,
		Host:          e.server.PublicIP,
		Port:          utils.DefaultSSHPort,
		LocalPort:     8080, //nolint:mnd
		RemoteAddress: "localhost",
		RemotePort:    80, //nolint:mnd
	}

	e.sshPortForward, err = utils.NewSSHPortForward(e.sshPortForwardInput)
	if err != nil {
		return fmt.Errorf("failed to setup SSH port forwarding: %w", err)
	}

	// Use non-inherited context to avoid cancelling the port forward with timeout
	if err = e.sshPortForward.Start(context.Background()); err != nil { //nolint:contextcheck
		return fmt.Errorf("failed to wait for the SSH port to become ready: %w", err)
	}

	clientOpts, err := dockerhelper.ClientOptsWithSSHConn(ctx, e.workDir, e.sshKeyPair, e.sshPortForwardInput)
	if err != nil {
		return fmt.Errorf("failed to get options for docker client: %w", err)
	}

	e.DockerHelper, err = dockerhelper.New(clientOpts)
	if err != nil {
		return fmt.Errorf("failed to create Docker helper: %w", err)
	}

	err = e.DockerHelper.WaitForDockerReady(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if Docker client is ready: %w", err)
	}

	return nil
}

func (e *AWSEnv) cleanupStack(ctx context.Context) error {
	// Delete the key pair
	_, err := e.ec2Client.DeleteKeyPair(ctx, &ec2.DeleteKeyPairInput{
		KeyName: &e.stackName,
	})
	if err != nil {
		return fmt.Errorf("failed to delete key pair: %w", err)
	}

	// Remove work directory
	err = os.RemoveAll(e.workDir)
	if err != nil {
		return fmt.Errorf("failed to clean up workdir: %w", err)
	}

	// Delete template file from S3 bucket
	_, err = e.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &e.stackName,
		Key:    &e.stackName,
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	// Delete the S3 bucket
	_, err = e.s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: &e.stackName,
	})
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}

	return nil
}

// Get stack status by stack name.
func (e *AWSEnv) getStackStatus(ctx context.Context) (cloudformationtypes.StackStatus, error) {
	stacks, err := e.client.DescribeStacks(
		ctx,
		&cloudformation.DescribeStacksInput{
			StackName: &e.stackName,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to describe stack: %w", err)
	}

	if len(stacks.Stacks) != 1 {
		return "", errors.New("failed to find stack")
	}

	return stacks.Stacks[0].StackStatus, nil
}

func (e *AWSEnv) getServer(ctx context.Context) (*Server, error) {
	// Get list of stack resources
	resources, err := e.client.ListStackResources(
		ctx,
		&cloudformation.ListStackResourcesInput{
			StackName: &e.stackName,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list stack resources: %w", err)
	}

	// Get VMClarity Server EC2 instance ID
	var serverInstanceID string
	for _, resource := range resources.StackResourceSummaries {
		if *resource.ResourceType == "AWS::EC2::Instance" {
			serverInstanceID = *resource.PhysicalResourceId

			break
		}
	}

	publicIP, err := e.getServerPublicIP(ctx, serverInstanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public IP address for instance with %s ID: %w", serverInstanceID, err)
	}

	return &Server{
		InstanceID: serverInstanceID,
		PublicIP:   publicIP,
	}, nil
}

// Check if the infrastructure is ready.
func (e *AWSEnv) infrastructureReady(ctx context.Context) (bool, error) {
	// Get stack status
	// If the stack status is not CREATE_COMPLETE, then the infrastructure is not ready
	stackStatus, err := e.getStackStatus(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get stack status: %w", err)
	}
	if stackStatus != cloudformationtypes.StackStatusCreateComplete {
		return false, nil
	}

	// Get test asset status
	// If the test asset status is not running, then the infrastructure are not ready
	ready, err := e.isEC2InstanceReady(ctx, e.testAsset.InstanceID)
	if err != nil {
		return false, fmt.Errorf("failed to get test instance status: %w", err)
	}
	if !ready {
		return false, nil
	}

	// Get VMClarity Server EC2 instance status
	// If the server status is not running, then the infrastructure are not ready
	e.server, err = e.getServer(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get VMClarity server: %w", err)
	}

	ready, err = e.isEC2InstanceReady(ctx, e.server.InstanceID)
	if err != nil {
		return false, fmt.Errorf("failed to get server instance status: %w", err)
	}

	return ready, nil
}

// Get EC2 instance status by instance ID.
func (e *AWSEnv) isEC2InstanceReady(ctx context.Context, instanceID string) (bool, error) {
	instanceStatus, err := e.ec2Client.DescribeInstanceStatus(
		ctx,
		&ec2.DescribeInstanceStatusInput{
			InstanceIds: []string{instanceID},
		},
	)
	if err != nil {
		return false, fmt.Errorf("failed to describe instance status: %w", err)
	}

	if len(instanceStatus.InstanceStatuses) != 1 {
		return false, errors.New("failed to find instance status")
	}

	return instanceStatus.InstanceStatuses[0].InstanceState.Name == ec2types.InstanceStateNameRunning, nil
}

// Get the public IP address by instance ID.
func (e *AWSEnv) getServerPublicIP(ctx context.Context, instanceID string) (string, error) {
	instances, err := e.ec2Client.DescribeInstances(
		ctx,
		&ec2.DescribeInstancesInput{
			InstanceIds: []string{instanceID},
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to describe instances: %w", err)
	}

	if len(instances.Reservations) != 1 {
		return "", errors.New("failed to find instance status")
	}

	if instances.Reservations[0].Instances[0].State.Name != ec2types.InstanceStateNameRunning {
		return "", errors.New("server instance is not running")
	}

	return *instances.Reservations[0].Instances[0].PublicIpAddress, nil
}
