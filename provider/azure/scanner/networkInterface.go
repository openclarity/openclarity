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

// nolint: wrapcheck
package scanner

import (
	"context"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"

	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/provider/azure/utils"
)

var (
	NetworkInterfaceEstimateProvisionTime = 10 * time.Second
	NetworkInterfaceDeleteEstimateTime    = 10 * time.Second
)

func networkInterfaceNameFromJobConfig(config *provider.ScanJobConfig) string {
	return "scanner-nic-" + config.AssetScanID
}

func (s *Scanner) ensureNetworkInterface(ctx context.Context, config *provider.ScanJobConfig) (armnetwork.Interface, error) {
	nicName := networkInterfaceNameFromJobConfig(config)

	nicResp, err := s.InterfacesClient.Get(ctx, s.ScannerResourceGroup, nicName, nil)
	if err == nil {
		if *nicResp.Interface.Properties.ProvisioningState != provisioningStateSucceeded {
			return nicResp.Interface, provider.RetryableErrorf(NetworkInterfaceEstimateProvisionTime, "interface is not ready yet, provisioning state: %s", *nicResp.Interface.Properties.ProvisioningState)
		}

		return nicResp.Interface, nil
	}

	notFound, err := utils.HandleAzureRequestError(err, "getting interface %s", nicName)
	if !notFound {
		return armnetwork.Interface{}, err
	}

	parameters := armnetwork.Interface{
		Location: to.Ptr(s.ScannerLocation),
		Properties: &armnetwork.InterfacePropertiesFormat{
			IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
				{
					Name: to.Ptr(nicName + "-ipconfig"),
					Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
						PrivateIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodDynamic),
						Subnet: &armnetwork.Subnet{
							ID: to.Ptr(s.ScannerSubnet),
						},
					},
				},
			},
			NetworkSecurityGroup: &armnetwork.SecurityGroup{
				ID: to.Ptr(s.ScannerSecurityGroup),
			},
		},
	}

	_, err = s.InterfacesClient.BeginCreateOrUpdate(ctx, s.ScannerResourceGroup, nicName, parameters, nil)
	if err != nil {
		_, err := utils.HandleAzureRequestError(err, "creating interface %s", nicName)
		return armnetwork.Interface{}, err
	}

	return armnetwork.Interface{}, provider.RetryableErrorf(NetworkInterfaceEstimateProvisionTime, "interface creating")
}

func (s *Scanner) ensureNetworkInterfaceDeleted(ctx context.Context, config *provider.ScanJobConfig) error {
	nicName := networkInterfaceNameFromJobConfig(config)

	return utils.EnsureDeleted(
		"interface",
		func() error {
			_, err := s.InterfacesClient.Get(ctx, s.ScannerResourceGroup, nicName, nil)
			return err
		},
		func() error {
			_, err := s.InterfacesClient.BeginDelete(ctx, s.ScannerResourceGroup, nicName, nil)
			return err
		},
		NetworkInterfaceDeleteEstimateTime,
	)
}
