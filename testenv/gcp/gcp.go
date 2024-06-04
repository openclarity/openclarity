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

package gcp

import (
	"context"
	"crypto/sha1" // nolint:gosec
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/api/iam/v1"

	"github.com/openclarity/vmclarity/testenv/gcp/asset"
	"github.com/openclarity/vmclarity/testenv/types"
	"github.com/openclarity/vmclarity/testenv/utils"
	"github.com/openclarity/vmclarity/testenv/utils/docker"
)

type ContextKeyType string

const (
	GCPClientContextKey                 ContextKeyType = "GCPClient"
	PollInterval                                       = 5 * time.Second
	DiscovererSnapshotterRoleNameFormat                = "projects/%s/roles/vmclarity_%s_discoverer_snapshotter"
	ScannerRoleNameFormat                              = "projects/%s/roles/vmclarity_%s_scanner"

	DefaultRemoteUser = "vmclarity"
	ProjectID         = "gcp-osedev-nprd-52462"
	Zone              = "us-central1-f"
	TestAssetName     = "vmclarity-test-asset"
)

type GCPEnv struct {
	workDir         string
	dm              *deploymentmanager.Service
	instancesClient *compute.InstancesClient
	iamService      *iam.Service
	serverIP        *string
	envName         string

	sshKeyPair          *utils.SSHKeyPair
	sshPortForwardInput *utils.SSHForwardInput
	sshPortForward      *utils.SSHPortForward

	*docker.DockerHelper
}

// nolint:cyclop
func (e *GCPEnv) SetUp(ctx context.Context) error {
	example, err := e.createExample()
	if err != nil {
		return fmt.Errorf("failed to create example: %w", err)
	}

	imports, err := createDeploymentImports()
	if err != nil {
		return fmt.Errorf("failed to create deployment imports: %w", err)
	}

	op, err := e.dm.Deployments.Insert(
		ProjectID,
		&deploymentmanager.Deployment{
			Name: e.envName,
			Target: &deploymentmanager.TargetConfiguration{
				Config: &deploymentmanager.ConfigFile{
					Content: example,
				},
				Imports: imports,
			},
		},
	).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to set up the deployment: %w", err)
	}

	for {
		op, err = e.dm.Operations.Get(ProjectID, op.Name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("failed to get operation status: %w", err)
		}
		if op.Status == "DONE" {
			break
		}

		time.Sleep(PollInterval)
	}

	err = asset.Create(ctx, e.instancesClient, ProjectID, Zone, TestAssetName)
	if err != nil {
		return fmt.Errorf("failed to create test asset: %w", err)
	}

	if err = e.afterSetUp(ctx); err != nil {
		return fmt.Errorf("failed to run after setup: %w", err)
	}

	return nil
}

func (e *GCPEnv) TearDown(ctx context.Context) error {
	// Stop SSH port forwarding
	if e.sshPortForward != nil {
		e.sshPortForward.Stop()
	}

	op, err := e.dm.Deployments.Delete(ProjectID, e.envName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("unable to delete deployment: %w", err)
	}
	for {
		op, err = e.dm.Operations.Get(ProjectID, op.Name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("failed to get operation status: %w", err)
		}
		if op.Status == "DONE" {
			break
		}

		time.Sleep(PollInterval)
	}

	err = asset.Delete(ctx, e.instancesClient, ProjectID, Zone, TestAssetName)
	if err != nil {
		return fmt.Errorf("failed to delete test asset: %w", err)
	}

	err = os.RemoveAll(e.workDir)
	if err != nil {
		return fmt.Errorf("failed to remove work directory: %w", err)
	}

	h := sha1.New() // nolint:gosec
	_, err = io.WriteString(h, e.envName)
	SHA1Hash := hex.EncodeToString(h.Sum(nil))[:10]
	if err != nil {
		return fmt.Errorf("failed to write string: %w", err)
	}

	_, err = e.iamService.Projects.Roles.Undelete(fmt.Sprintf(DiscovererSnapshotterRoleNameFormat, ProjectID, SHA1Hash), &iam.UndeleteRoleRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to undelete discoverer snapshotter role: %w", err)
	}
	_, err = e.iamService.Projects.Roles.Undelete(fmt.Sprintf(ScannerRoleNameFormat, ProjectID, SHA1Hash), &iam.UndeleteRoleRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to undelete scanner role: %w", err)
	}

	return nil
}

func (e *GCPEnv) ServiceLogs(_ context.Context, _ []string, startTime time.Time, stdout, stderr io.Writer) error {
	input := &utils.SSHJournalctlInput{
		PrivateKey: e.sshKeyPair.PrivateKey,
		PublicKey:  e.sshKeyPair.PublicKey,
		User:       DefaultRemoteUser,
		Host:       *e.serverIP,
		WorkDir:    e.workDir,
		Service:    "docker",
	}

	err := utils.GetServiceLogs(input, startTime, stdout, stderr)
	if err != nil {
		return fmt.Errorf("failed to get service logs: %w", err)
	}

	return nil
}

func (e *GCPEnv) Endpoints(_ context.Context) (*types.Endpoints, error) {
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

func (e *GCPEnv) Context(ctx context.Context) (context.Context, error) {
	return context.WithValue(ctx, GCPClientContextKey, e.dm), nil
}

func New(config *Config, opts ...ConfigOptFn) (*GCPEnv, error) {
	if err := applyConfigWithOpts(config, opts...); err != nil {
		return nil, fmt.Errorf("failed to apply config options: %w", err)
	}

	dm, err := deploymentmanager.NewService(config.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create deploymentmanager: %w", err)
	}

	instancesClient, err := compute.NewInstancesRESTClient(config.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance client: %w", err)
	}

	iamService, err := iam.NewService(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create IAM service: %w", err)
	}

	sshKeyPair, err := utils.LoadOrGenerateAndSaveSSHKeyPair(config.PrivateKeyFile, config.PublicKeyFile, config.WorkDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH key pair: %w", err)
	}

	return &GCPEnv{
		workDir:         config.WorkDir,
		dm:              dm,
		instancesClient: instancesClient,
		iamService:      iamService,
		sshKeyPair:      sshKeyPair,
		envName:         config.EnvName,
	}, nil
}
