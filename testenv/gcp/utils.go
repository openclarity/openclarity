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

	"github.com/openclarity/vmclarity/installation"
	"github.com/openclarity/vmclarity/testenv/utils"
	"github.com/openclarity/vmclarity/testenv/utils/docker"
)

const (
	FileVMClarityInstallScriptSchema = "components/vmclarity_install_script.py.schema"
	FileVMClarityServerSchema        = "components/vmclarity-server.py.schema"
	FileVMClaritySchema              = "vmclarity.py.schema"
	FileVMClarity                    = "vmclarity.py"
	FileVMClarityInstallScript       = "components/vmclarity_install_script.py"
	FileVMClarityInstall             = "components/vmclarity-install.sh"
	FileVMClarityServer              = "components/vmclarity-server.py"
	FileNetwork                      = "components/network.py"
	FileFirewallRules                = "components/firewall-rules.py"
	FileStaticIP                     = "components/static-ip.py"
	FileServiceAccount               = "components/service-account.py"
	FileRoles                        = "components/roles.py"
	FileCloudRouter                  = "components/cloud-router.py"
)

func (e *GCPEnv) afterSetUp(ctx context.Context) error {
	req := &computepb.GetInstanceRequest{
		Project:  ProjectID,
		Zone:     Zone,
		Instance: "vmclarity-" + e.envName + "-server",
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
		LocalPort:     8080, //nolint:mnd
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
	vmclarityConfigExampleYaml, err := installation.GCPManifestBundle.ReadFile("vmclarity-config.example.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return strings.Replace(string(vmclarityConfigExampleYaml), "<SSH Public Key>", string(e.sshKeyPair.PublicKey), -1), nil
}

func createDeploymentImports() ([]*deploymentmanager.ImportFile, error) {
	vmclarityPy, err := installation.GCPManifestBundle.ReadFile(FileVMClarity)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	vmclarityPySchema, err := installation.GCPManifestBundle.ReadFile(FileVMClaritySchema)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	vmclarityServerPy, err := installation.GCPManifestBundle.ReadFile(FileVMClarityServer)
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
	vmclarityInstallScriptPy, err := installation.GCPManifestBundle.ReadFile(FileVMClarityInstallScript)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	vmclarityInstallSh, err := installation.GCPManifestBundle.ReadFile(FileVMClarityInstall)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	vmclarityServerPySchema, err := installation.GCPManifestBundle.ReadFile(FileVMClarityServerSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	vmclarityInstallScriptPySchema, err := installation.GCPManifestBundle.ReadFile(FileVMClarityInstallScriptSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return []*deploymentmanager.ImportFile{
		{
			Content: string(vmclarityPy),
			Name:    FileVMClarity,
		},
		{
			Content: string(vmclarityPySchema),
			Name:    FileVMClaritySchema,
		},
		{
			Content: string(vmclarityServerPy),
			Name:    FileVMClarityServer,
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
			Content: string(vmclarityInstallScriptPy),
			Name:    filepath.Base(FileVMClarityInstallScript),
		},
		{
			Content: string(vmclarityInstallSh),
			Name:    filepath.Base(FileVMClarityInstall),
		},
		{
			Content: string(vmclarityServerPySchema),
			Name:    FileVMClarityServerSchema,
		},
		{
			Content: string(vmclarityInstallScriptPySchema),
			Name:    FileVMClarityInstallScriptSchema,
		},
	}, nil
}
