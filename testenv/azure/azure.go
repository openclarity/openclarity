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

package azure

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"

	"github.com/openclarity/vmclarity/testenv/types"
	"github.com/openclarity/vmclarity/testenv/utils"
	"github.com/openclarity/vmclarity/testenv/utils/docker"
)

const (
	bicepFileResourcesFormat                  = "vmclarity-%s"
	bicepFileDiscoverRoleDeploymentNameFormat = "%s-discover-role"
	defaultRemoteUser                         = "vmclarity"
)

var _ types.Environment = &AzureEnv{}

type AzureEnv struct {
	// clients to manage Azure resources
	deploymentsClient       *armresources.DeploymentsClient
	resourceGroupClient     *armresources.ResourceGroupsClient
	roleDefinitionsClient   *armauthorization.RoleDefinitionsClient
	roleAssignmentsClient   *armauthorization.RoleAssignmentsClient
	virtualNetworksClient   *armnetwork.VirtualNetworksClient
	subnetsClient           *armnetwork.SubnetsClient
	interfacesClient        *armnetwork.InterfacesClient
	publicIPAddressesClient *armnetwork.PublicIPAddressesClient
	virtualMachinesClient   *armcompute.VirtualMachinesClient

	// Azure resource information
	subscriptionID      string
	postfix             string
	location            string
	sshKeyPair          *utils.SSHKeyPair
	sshPortForwardInput *utils.SSHForwardInput
	sshPortForward      *utils.SSHPortForward
	serverHost          *string

	// test environment information
	workDir string
	meta    map[string]interface{}

	*docker.DockerHelper
}

// SetUp Azure test environment
// * Install necessary components via the built bicep files.
// * Create test asset.
// * When infrastructure is ready, start SSH port forwarding and remote docker client.
// Returns error if it fails to set up the environment.
func (e *AzureEnv) SetUp(ctx context.Context) error {
	// Create and validate VMClarity deployment
	if err := e.createAzureResources(ctx); err != nil {
		return fmt.Errorf("failed to create Azure resources: %w", err)
	}

	// Create an asset VM to be scanned
	if err := e.createAssetVM(ctx, e.location, fmt.Sprintf(bicepFileResourcesFormat, e.postfix)); err != nil {
		return fmt.Errorf("failed to create asset VM: %w", err)
	}

	// Set connection with VMClarity server
	if err := e.setServerConnection(ctx); err != nil {
		return fmt.Errorf("failed to set server connection: %w", err)
	}

	return nil
}

// TearDown the test environment by uninstalling components installed via Setup.
// Returns error if it fails to clean up the environment.
func (e *AzureEnv) TearDown(ctx context.Context) error {
	// Stop SSH port forwarding
	if e.sshPortForward != nil {
		e.sshPortForward.Stop()
	}

	discoverRoleDeploymentName := fmt.Sprintf(bicepFileDiscoverRoleDeploymentNameFormat, fmt.Sprintf(bicepFileResourcesFormat, e.postfix))

	// Cleanup VMClarity Discoverer Snapshotter role assignment and role definition
	// They are at subscription scope so removing the resource group will not remove them
	if err := e.cleanupDiscovererSnapshotterRole(ctx, discoverRoleDeploymentName); err != nil {
		return fmt.Errorf("failed to cleanup VMClarity Discoverer Snapshotter role: %w", err)
	}

	// Cleanup resource group
	if err := e.cleanupResourceGroup(ctx); err != nil {
		return fmt.Errorf("failed to cleanup resource group: %w", err)
	}

	// Cleanup deployments
	if err := e.cleanupDeployments(ctx, discoverRoleDeploymentName); err != nil {
		return fmt.Errorf("failed to cleanup deployments: %w", err)
	}

	// Cleanup work directory
	err := os.RemoveAll(e.workDir)
	if err != nil {
		return fmt.Errorf("failed to clean up workdir: %w", err)
	}

	return nil
}

// ServiceLogs writes service logs to io.Writer for list of services from startTime timestamp.
// Returns error if it cannot retrieve logs.
func (e *AzureEnv) ServiceLogs(ctx context.Context, services []string, startTime time.Time, stdout, stderr io.Writer) error {
	input := &utils.SSHJournalctlInput{
		PrivateKey: e.sshKeyPair.PrivateKey,
		PublicKey:  e.sshKeyPair.PublicKey,
		User:       defaultRemoteUser,
		Host:       *e.serverHost,
		WorkDir:    e.workDir,
		Service:    "docker",
	}

	err := utils.GetServiceLogs(input, startTime, stdout, stderr)
	if err != nil {
		return fmt.Errorf("failed to get service logs: %w", err)
	}

	return nil
}

// Endpoints returns an Endpoints object containing API endpoints for services.
func (e *AzureEnv) Endpoints(ctx context.Context) (*types.Endpoints, error) {
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

// Context updates the provided ctx with environment specific data like with initialized client data allowing tests
// to interact with the underlying infrastructure.
func (e *AzureEnv) Context(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func New(config *Config, opts ...ConfigOptFn) (*AzureEnv, error) {
	if err := applyConfigWithOpts(config, opts...); err != nil {
		return nil, fmt.Errorf("failed to apply config options: %w", err)
	}

	if err := checkAzureEnvVars(); err != nil {
		return nil, err
	}

	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	resourcesClientFactory, err := armresources.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure resources client factory: %w", err)
	}

	authorizationClientFactory, err := armauthorization.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure authorization client factory: %w", err)
	}

	networkClientFactory, err := armnetwork.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure network client factory: %w", err)
	}

	computeClientFactory, err := armcompute.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure compute client factory: %w", err)
	}

	sshKeyPair, err := utils.LoadOrGenerateAndSaveSSHKeyPair(config.PrivateKeyFile, config.PublicKeyFile, config.WorkDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH key pair: %w", err)
	}

	return &AzureEnv{
		deploymentsClient:       resourcesClientFactory.NewDeploymentsClient(),
		resourceGroupClient:     resourcesClientFactory.NewResourceGroupsClient(),
		roleDefinitionsClient:   authorizationClientFactory.NewRoleDefinitionsClient(),
		roleAssignmentsClient:   authorizationClientFactory.NewRoleAssignmentsClient(),
		virtualNetworksClient:   networkClientFactory.NewVirtualNetworksClient(),
		subnetsClient:           networkClientFactory.NewSubnetsClient(),
		interfacesClient:        networkClientFactory.NewInterfacesClient(),
		publicIPAddressesClient: networkClientFactory.NewPublicIPAddressesClient(),
		virtualMachinesClient:   computeClientFactory.NewVirtualMachinesClient(),
		subscriptionID:          subscriptionID,
		postfix:                 strings.TrimPrefix(config.EnvName, "vmclarity-"),
		location:                config.Region,
		sshKeyPair:              sshKeyPair,
		workDir:                 config.WorkDir,
		meta: map[string]interface{}{
			"environment": "azure",
			"name":        config.EnvName,
		},
	}, nil
}
