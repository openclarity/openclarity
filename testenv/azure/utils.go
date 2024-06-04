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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"

	"github.com/openclarity/vmclarity/installation"
	"github.com/openclarity/vmclarity/testenv/utils"
	dockerhelper "github.com/openclarity/vmclarity/testenv/utils/docker"
)

func checkAzureEnvVars() error {
	var missingEnvs []string

	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		missingEnvs = append(missingEnvs, "AZURE_SUBSCRIPTION_ID")
	}

	tenantID := os.Getenv("AZURE_TENANT_ID")
	if len(tenantID) == 0 {
		missingEnvs = append(missingEnvs, "AZURE_TENANT_ID")
	}

	clientID := os.Getenv("AZURE_CLIENT_ID")
	if len(clientID) == 0 {
		missingEnvs = append(missingEnvs, "AZURE_CLIENT_ID")
	}

	clientSecret := os.Getenv("AZURE_CLIENT_SECRET")
	if len(clientSecret) == 0 {
		missingEnvs = append(missingEnvs, "AZURE_CLIENT_SECRET")
	}

	if len(missingEnvs) > 0 {
		return fmt.Errorf("missing environment variables: %s", strings.Join(missingEnvs, ", "))
	}

	return nil
}

func (e *AzureEnv) createAzureResources(ctx context.Context) error {
	templateFile, err := installation.AzureManifestBundle.ReadFile("vmclarity.json")
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}
	template := make(map[string]interface{})
	if err := json.Unmarshal(templateFile, &template); err != nil {
		return fmt.Errorf("failed to unmarshal template file: %w", err)
	}

	deploymentResp, err := e.deploymentsClient.BeginCreateOrUpdateAtSubscriptionScope(
		ctx,
		fmt.Sprintf(bicepFileResourcesFormat, e.postfix),
		armresources.Deployment{
			Location: &e.location,
			Properties: &armresources.DeploymentProperties{
				Template:   template,
				Parameters: e.createTestParams(),
				Mode:       to.Ptr(armresources.DeploymentModeIncremental),
			},
		},
		nil)
	if err != nil {
		return fmt.Errorf("failed to begin creating deployment: %w", err)
	}

	_, err = deploymentResp.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to create deployment: %w", err)
	}

	validateResp, err := e.deploymentsClient.BeginValidateAtSubscriptionScope(
		ctx,
		fmt.Sprintf(bicepFileResourcesFormat, e.postfix),
		armresources.Deployment{
			Location: &e.location,
			Properties: &armresources.DeploymentProperties{
				Template:   template,
				Parameters: e.createTestParams(),
				Mode:       to.Ptr(armresources.DeploymentModeIncremental),
			},
		},
		nil)
	if err != nil {
		return fmt.Errorf("failed to begin validating deployment: %w", err)
	}

	_, err = validateResp.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to validate deployment: %w", err)
	}

	return nil
}

func (e *AzureEnv) createTestParams() map[string]interface{} {
	params := make(map[string]interface{})

	params["location"] = map[string]interface{}{"value": e.location}
	params["adminUsername"] = map[string]interface{}{"value": defaultRemoteUser}
	params["adminSSHKey"] = map[string]interface{}{"value": strings.TrimSpace(string(e.sshKeyPair.PublicKey))}
	params["deployPostfix"] = map[string]interface{}{"value": e.postfix}

	return params
}

func (e *AzureEnv) setServerConnection(ctx context.Context) error {
	// get the public IP address of the VMClarity server
	serverPublicIP, err := e.publicIPAddressesClient.Get(ctx, fmt.Sprintf(bicepFileResourcesFormat, e.postfix), fmt.Sprintf(bicepFileResourcesFormat, "server-public-ip"), &armnetwork.PublicIPAddressesClientGetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get the public IP address of the server: %w", err)
	}

	e.serverHost = serverPublicIP.Properties.DNSSettings.Fqdn

	e.sshPortForwardInput = &utils.SSHForwardInput{
		PrivateKey:    e.sshKeyPair.PrivateKey,
		User:          defaultRemoteUser,
		Host:          *e.serverHost,
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

func (e *AzureEnv) cleanupDeployments(ctx context.Context, discoverRoleDeploymentName string) error {
	deploymentPollerResp, err := e.deploymentsClient.BeginDeleteAtSubscriptionScope(ctx, fmt.Sprintf(bicepFileResourcesFormat, e.postfix), nil)
	if err != nil {
		return fmt.Errorf("failed to begin deleting deployment: %w", err)
	}

	_, err = deploymentPollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}

	discoverRoleDeploymentPollerResp, err := e.deploymentsClient.BeginDeleteAtSubscriptionScope(ctx, discoverRoleDeploymentName, nil)
	if err != nil {
		return fmt.Errorf("failed to begin deleting discover role deployment: %w", err)
	}

	_, err = discoverRoleDeploymentPollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete discover role deployment: %w", err)
	}

	return nil
}

func (e *AzureEnv) cleanupDiscovererSnapshotterRole(ctx context.Context, discoverRoleDeploymentName string) error {
	// Get the role assignment and role definition IDs from the deployment they were created in
	discoverRoleDeployment, err := e.deploymentsClient.GetAtSubscriptionScope(ctx, discoverRoleDeploymentName, nil)
	if err != nil {
		return fmt.Errorf("failed to get deployment %s: %w", discoverRoleDeploymentName, err)
	}

	var discoverRoleAssignmentID string
	var discoverRoleDefinitionID string

	for _, resource := range discoverRoleDeployment.Properties.OutputResources {
		if strings.Contains(*resource.ID, "roleAssignments") {
			discoverRoleAssignmentID = *resource.ID
		} else if strings.Contains(*resource.ID, "roleDefinitions") {
			discoverRoleDefinitionID = *resource.ID
		}
	}

	if discoverRoleAssignmentID == "" || discoverRoleDefinitionID == "" {
		return fmt.Errorf("cannot find role assignment or role definition for %s", discoverRoleDeploymentName)
	}

	_, err = e.roleAssignmentsClient.DeleteByID(ctx, discoverRoleAssignmentID, nil)
	if err != nil {
		return fmt.Errorf("failed to delete role assignment: %w", err)
	}

	// roleDefinitionsClient does not have DeleteByID method (yet)
	roleDefinition, err := e.roleDefinitionsClient.GetByID(ctx, discoverRoleDefinitionID, nil)
	if err != nil {
		return fmt.Errorf("failed to get role definition: %w", err)
	}

	_, err = e.roleDefinitionsClient.Delete(ctx, "/subscriptions/"+e.subscriptionID, *roleDefinition.Name, nil)
	if err != nil {
		return fmt.Errorf("failed to delete role definition: %w", err)
	}

	return nil
}

func (e *AzureEnv) cleanupResourceGroup(ctx context.Context) error {
	resourceGroupPollerResp, err := e.resourceGroupClient.BeginDelete(ctx, fmt.Sprintf(bicepFileResourcesFormat, e.postfix), nil)
	if err != nil {
		return fmt.Errorf("failed to begin deleting resource group: %w", err)
	}

	_, err = resourceGroupPollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete resource group: %w", err)
	}

	return nil
}
