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
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cloudformationtypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/openclarity/vmclarity/testenv/aws/asset"
	"github.com/openclarity/vmclarity/testenv/types"
	"github.com/openclarity/vmclarity/testenv/utils"
	"github.com/openclarity/vmclarity/testenv/utils/docker"
)

type ContextKeyType string

const (
	AWSClientContextKey         ContextKeyType = "AWSClient"
	DefaultStackCreationTimeout                = 30 * time.Minute
	DefaultSSHPortReadyTimeout                 = 2 * time.Minute
	DefaultRemoteUser                          = "ubuntu"
)

var _ types.Environment = &AWSEnv{}

// AWS Environment.
type AWSEnv struct {
	client              *cloudformation.Client
	ec2Client           *ec2.Client
	s3Client            *s3.Client
	testAsset           *asset.Asset
	server              *Server
	stackName           string
	templateURL         string
	workDir             string
	region              string
	sshKeyPair          *utils.SSHKeyPair
	sshPortForwardInput *utils.SSHForwardInput
	sshPortForward      *utils.SSHPortForward
	meta                map[string]interface{}

	*docker.DockerHelper
}

type Server struct {
	InstanceID string
	PublicIP   string
}

// Setup AWS test environment from cloud formation template.
// * Create a new CloudFormation stack from template
// (upload template file to S3 is required since the template is larger than 51,200 bytes).
// * Create test asset.
// * When infrastructure is ready, start SSH port forwarding and remote docker client.
func (e *AWSEnv) SetUp(ctx context.Context) error {
	// Prepare stack
	err := e.prepareStack(ctx)
	if err != nil {
		return fmt.Errorf("failed to prepare stack: %w", err)
	}

	// Create a new CloudFormation stack from template
	_, err = e.client.CreateStack(
		ctx,
		&cloudformation.CreateStackInput{
			StackName:    &e.stackName,
			Capabilities: []cloudformationtypes.Capability{cloudformationtypes.CapabilityCapabilityIam},
			TemplateURL:  &e.templateURL,
			Parameters: []cloudformationtypes.Parameter{
				{
					ParameterKey:   aws.String("KeyName"),
					ParameterValue: &e.stackName,
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create stack: %w", err)
	}

	waiter := cloudformation.NewStackCreateCompleteWaiter(e.client)
	if err = waiter.Wait(ctx, &cloudformation.DescribeStacksInput{
		StackName: &e.stackName,
	}, DefaultStackCreationTimeout); err != nil {
		return fmt.Errorf("failed to wait for the stack to be created: %w", err)
	}

	// Create a new test asset
	err = e.testAsset.Create(ctx, e.ec2Client)
	if err != nil {
		return fmt.Errorf("failed to create test asset: %w", err)
	}

	if err = e.afterSetUp(ctx); err != nil {
		return fmt.Errorf("failed to run after setup: %w", err)
	}

	return nil
}

func (e *AWSEnv) TearDown(ctx context.Context) error {
	// Stop SSH port forwarding
	if e.sshPortForward != nil {
		e.sshPortForward.Stop()
	}

	// Delete the CloudFormation stack
	_, err := e.client.DeleteStack(
		ctx,
		&cloudformation.DeleteStackInput{
			StackName: &e.stackName,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to delete stack: %w", err)
	}

	// Cleanup stack
	err = e.cleanupStack(ctx)
	if err != nil {
		return fmt.Errorf("failed to cleanup stack: %w", err)
	}

	// Delete the test asset
	err = e.testAsset.Delete(ctx, e.ec2Client)
	if err != nil {
		return fmt.Errorf("failed to delete test asset: %w", err)
	}

	return nil
}

func (e *AWSEnv) ServiceLogs(ctx context.Context, _ []string, startTime time.Time, stdout, stderr io.Writer) error {
	input := &utils.SSHJournalctlInput{
		PrivateKey: e.sshKeyPair.PrivateKey,
		PublicKey:  e.sshKeyPair.PublicKey,
		User:       DefaultRemoteUser,
		Host:       e.server.PublicIP,
		WorkDir:    e.workDir,
		Service:    "docker",
	}

	err := utils.GetServiceLogs(input, startTime, stdout, stderr)
	if err != nil {
		return fmt.Errorf("failed to get service logs: %w", err)
	}

	return nil
}

func (e *AWSEnv) Endpoints(_ context.Context) (*types.Endpoints, error) {
	apiURL, err := url.Parse("http://" + e.sshPortForwardInput.LocalAddressPort() + "/api")
	if err != nil {
		return nil, fmt.Errorf("failed to parse API URL: %w", err)
	}

	uiBackendURL, err := url.Parse("http://" + e.sshPortForwardInput.LocalAddressPort() + "/ui/api")
	if err != nil {
		return nil, fmt.Errorf("failed to parse Backend API URL: %w", err)
	}

	return &types.Endpoints{
		API:       apiURL,
		UIBackend: uiBackendURL,
	}, nil
}

func (e *AWSEnv) Context(ctx context.Context) (context.Context, error) {
	return context.WithValue(ctx, AWSClientContextKey, e.client), nil
}

func New(config *Config, opts ...ConfigOptFn) (*AWSEnv, error) {
	if err := applyConfigWithOpts(config, opts...); err != nil {
		return nil, fmt.Errorf("failed to apply config options: %w", err)
	}

	// Load default AWS configuration and set region
	cfg, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}
	cfg.Region = config.Region

	// Create AWS CloudFormation client
	client := cloudformation.NewFromConfig(cfg)

	// Create AWS EC2 client
	ec2Client := ec2.NewFromConfig(cfg)

	// Create AWS S3 client
	s3Client := s3.NewFromConfig(cfg)

	sshKeyPair, err := utils.LoadOrGenerateAndSaveSSHKeyPair(config.PrivateKeyFile, config.PublicKeyFile, config.WorkDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH key pair: %w", err)
	}

	return &AWSEnv{
		client:     client,
		ec2Client:  ec2Client,
		s3Client:   s3Client,
		stackName:  config.EnvName,
		workDir:    config.WorkDir,
		region:     config.Region,
		sshKeyPair: sshKeyPair,
		testAsset:  &asset.Asset{},
		meta: map[string]interface{}{
			"environment": "aws",
			"name":        config.EnvName,
		},
	}, nil
}
