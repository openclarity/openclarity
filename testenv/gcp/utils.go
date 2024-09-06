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
	"fmt"
	"path/filepath"
	"strings"

	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/deploymentmanager/v2"

	"github.com/openclarity/openclarity/installation"
	"github.com/openclarity/openclarity/testenv/utils"
	"github.com/openclarity/openclarity/testenv/utils/docker"
)

const (
	FileOpenClarityInstallScriptSchema = "components/openclarity_install_script.py.schema"
	FileOpenClarityServerSchema        = "components/openclarity-server.py.schema"
	FileOpenClaritySchema              = "openclarity.py.schema"
	FileOpenClarity                    = "openclarity.py"
	FileOpenClarityInstallScript       = "components/openclarity_install_script.py"
	FileOpenClarityInstall             = "components/openclarity-install.sh"
	FileOpenClarityServer              = "components/openclarity-server.py"
	FileNetwork                        = "components/network.py"
	FileFirewallRules                  = "components/firewall-rules.py"
	FileStaticIP                       = "components/static-ip.py"
	FileServiceAccount                 = "components/service-account.py"
	FileRoles                          = "components/roles.py"
	FileCloudRouter                    = "components/cloud-router.py"
)

func (e *GCPEnv) afterSetUp(ctx context.Context) error {
	req := &computepb.GetInstanceRequest{
		Project:  ProjectID,
		Zone:     Zone,
		Instance: "openclarity-" + e.envName + "-server",
	}
	server, err := e.instancesClient.Get(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to get server instance: %w", err)
	}

	e.serverIP = server.NetworkInterfaces[0].AccessConfigs[0].NatIP
	e.sshPortForwardInput = &utils.SSHForwardInput{
		PrivateKey:    e.sshKeyPair.PrivateKey,
		User:          DefaultRemoteUser,
		Host:          *e.serverIP,
		Port:          utils.DefaultSSHPort,
		LocalPort:     8083, //nolint:mnd
		RemoteAddress: "localhost",
		RemotePort:    80, //nolint:mnd
	}

	e.sshPortForward, err = utils.NewSSHPortForward(e.sshPortForwardInput)
	if err != nil {
		return fmt.Errorf("failed to setup SSH port forwarding: %w", err)
	}

	if err = e.sshPortForward.Start(context.Background()); err != nil { //nolint:contextcheck
		return fmt.Errorf("failed to wait for SSH port to become ready: %w", err)
	}

	clientOpts, err := docker.ClientOptsWithSSHConn(ctx, e.workDir, e.sshKeyPair, e.sshPortForwardInput)
	if err != nil {
		return fmt.Errorf("failed to get options for docker client: %w", err)
	}

	e.DockerHelper, err = docker.New(clientOpts)
	if err != nil {
		return fmt.Errorf("failed to create Docker helper: %w", err)
	}

	err = e.DockerHelper.WaitForDockerReady(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if Docker client is ready: %w", err)
	}

	return nil
}

func (e *GCPEnv) createExample() (string, error) {
	openclarityConfigExampleYaml, err := installation.GCPManifestBundle.ReadFile("openclarity-config.example.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return strings.Replace(string(openclarityConfigExampleYaml), "<SSH Public Key>", string(e.sshKeyPair.PublicKey), -1), nil
}

func createDeploymentImports() ([]*deploymentmanager.ImportFile, error) {
	openclarityPy, err := installation.GCPManifestBundle.ReadFile(FileOpenClarity)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	openclarityPySchema, err := installation.GCPManifestBundle.ReadFile(FileOpenClaritySchema)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	openclarityServerPy, err := installation.GCPManifestBundle.ReadFile(FileOpenClarityServer)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	networkPy, err := installation.GCPManifestBundle.ReadFile(FileNetwork)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	firewallRulesPy, err := installation.GCPManifestBundle.ReadFile(FileFirewallRules)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	staticIPPy, err := installation.GCPManifestBundle.ReadFile(FileStaticIP)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	serviceAccountPy, err := installation.GCPManifestBundle.ReadFile(FileServiceAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	rolesPy, err := installation.GCPManifestBundle.ReadFile(FileRoles)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	cloudRouterPy, err := installation.GCPManifestBundle.ReadFile(FileCloudRouter)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	openclarityInstallScriptPy, err := installation.GCPManifestBundle.ReadFile(FileOpenClarityInstallScript)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	openclarityInstallSh, err := installation.GCPManifestBundle.ReadFile(FileOpenClarityInstall)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	openclarityServerPySchema, err := installation.GCPManifestBundle.ReadFile(FileOpenClarityServerSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	openclarityInstallScriptPySchema, err := installation.GCPManifestBundle.ReadFile(FileOpenClarityInstallScriptSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return []*deploymentmanager.ImportFile{
		{
			Content: string(openclarityPy),
			Name:    FileOpenClarity,
		},
		{
			Content: string(openclarityPySchema),
			Name:    FileOpenClaritySchema,
		},
		{
			Content: string(openclarityServerPy),
			Name:    FileOpenClarityServer,
		},
		{
			Content: string(networkPy),
			Name:    FileNetwork,
		},
		{
			Content: string(firewallRulesPy),
			Name:    FileFirewallRules,
		},
		{
			Content: string(staticIPPy),
			Name:    FileStaticIP,
		},
		{
			Content: string(serviceAccountPy),
			Name:    FileServiceAccount,
		},
		{
			Content: string(rolesPy),
			Name:    FileRoles,
		},
		{
			Content: string(cloudRouterPy),
			Name:    FileCloudRouter,
		},
		{
			Content: string(openclarityInstallScriptPy),
			Name:    filepath.Base(FileOpenClarityInstallScript),
		},
		{
			Content: string(openclarityInstallSh),
			Name:    filepath.Base(FileOpenClarityInstall),
		},
		{
			Content: string(openclarityServerPySchema),
			Name:    FileOpenClarityServerSchema,
		},
		{
			Content: string(openclarityInstallScriptPySchema),
			Name:    FileOpenClarityInstallScriptSchema,
		},
	}, nil
}
