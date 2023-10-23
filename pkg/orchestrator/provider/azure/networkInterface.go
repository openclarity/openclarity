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

package azure

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v3"

	"github.com/openclarity/vmclarity/pkg/orchestrator/provider"
)

var (
	NetworkInterfaceEstimateProvisionTime = 10 * time.Second
	NetworkInterfaceDeleteEstimateTime    = 10 * time.Second
)

func networkInterfaceNameFromJobConfig(config *provider.ScanJobConfig) string {
	return fmt.Sprintf("scanner-nic-%s", config.AssetScanID)
}

func (p *Provider) ensureNetworkInterface(ctx context.Context, config *provider.ScanJobConfig) (armnetwork.Interface, error) {
	nicName := networkInterfaceNameFromJobConfig(config)

	nicResp, err := p.interfacesClient.Get(ctx, p.config.ScannerResourceGroup, nicName, nil)
	if err == nil {
		if *nicResp.Interface.Properties.ProvisioningState != ProvisioningStateSucceeded {
			return nicResp.Interface, provider.RetryableErrorf(NetworkInterfaceEstimateProvisionTime, "interface is not ready yet, provisioning state: %s", *nicResp.Interface.Properties.ProvisioningState)
		}

		return nicResp.Interface, nil
	}

	notFound, err := handleAzureRequestError(err, "getting interface %s", nicName)
	if !notFound {
		return armnetwork.Interface{}, err
	}

	parameters := armnetwork.Interface{
		Location: to.Ptr(p.config.ScannerLocation),
		Properties: &armnetwork.InterfacePropertiesFormat{
			IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
				{
					Name: to.Ptr(fmt.Sprintf("%s-ipconfig", nicName)),
					Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
						PrivateIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodDynamic),
						Subnet: &armnetwork.Subnet{
							ID: to.Ptr(p.config.ScannerSubnet),
						},
					},
				},
			},
			NetworkSecurityGroup: &armnetwork.SecurityGroup{
				ID: to.Ptr(p.config.ScannerSecurityGroup),
			},
		},
	}

	_, err = p.interfacesClient.BeginCreateOrUpdate(ctx, p.config.ScannerResourceGroup, nicName, parameters, nil)
	if err != nil {
		_, err := handleAzureRequestError(err, "creating interface %s", nicName)
		return armnetwork.Interface{}, err
	}

	return armnetwork.Interface{}, provider.RetryableErrorf(NetworkInterfaceEstimateProvisionTime, "interface creating")
}

func (p *Provider) ensureNetworkInterfaceDeleted(ctx context.Context, config *provider.ScanJobConfig) error {
	nicName := networkInterfaceNameFromJobConfig(config)

	return ensureDeleted(
		"interface",
		func() error {
			_, err := p.interfacesClient.Get(ctx, p.config.ScannerResourceGroup, nicName, nil)
			return err // nolint: wrapcheck
		},
		func() error {
			_, err := p.interfacesClient.BeginDelete(ctx, p.config.ScannerResourceGroup, nicName, nil)
			return err // nolint: wrapcheck
		},
		NetworkInterfaceDeleteEstimateTime,
	)
}
