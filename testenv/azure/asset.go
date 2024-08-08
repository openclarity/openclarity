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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
)

const (
	virtualNetworkName     = "vmclarity-asset-net"
	virtualNetworkAddress  = "10.1.0.0/16"
	networkSubnetName      = "vmclarity-asset-subnet"
	networkSubnetAddress   = "10.1.0.0/24"
	networkInterfaceName   = "vmclarity-asset-net-int"
	osDiskName             = "vmclarity-asset-os-disk"
	virtualMachineName     = "vmclarity-asset"
	virtualMachineUserName = "ubuntu"
	virtualMachineSize     = "Standard_B1ls"
)

var imageReference = &armcompute.ImageReference{
	Offer:     to.Ptr("0001-com-ubuntu-minimal-focal"),
	Publisher: to.Ptr("Canonical"),
	SKU:       to.Ptr("minimal-20_04-lts"),
	Version:   to.Ptr("20.04.202004230"),
}

func (e *AzureEnv) createAssetVM(ctx context.Context, location, resourceGroupName string) error {
	_, err := e.createVirtualNetwork(ctx, location, resourceGroupName)
	if err != nil {
		return fmt.Errorf("failed to create virtual network for asset VM: %w", err)
	}

	subnet, err := e.createSubnet(ctx, resourceGroupName)
	if err != nil {
		return fmt.Errorf("failed to create subnet for asset VM: %w", err)
	}

	nic, err := e.createNIC(ctx, location, resourceGroupName, *subnet.ID)
	if err != nil {
		return fmt.Errorf("failed to create network interface for asset VM: %w", err)
	}

	_, err = e.createVirtualMachine(ctx, location, resourceGroupName, *nic.ID)
	if err != nil {
		return fmt.Errorf("failed to create asset virtual machine: %w", err)
	}

	return nil
}

func (e *AzureEnv) createVirtualNetwork(ctx context.Context, location, resourceGroupName string) (*armnetwork.VirtualNetwork, error) {
	pollerResp, err := e.virtualNetworksClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		virtualNetworkName,
		armnetwork.VirtualNetwork{
			Location: to.Ptr(location),
			Properties: &armnetwork.VirtualNetworkPropertiesFormat{
				AddressSpace: &armnetwork.AddressSpace{
					AddressPrefixes: []*string{
						to.Ptr(virtualNetworkAddress),
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to begin creating virtual network: %w", err)
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create virtual network: %w", err)
	}

	return &resp.VirtualNetwork, nil
}

func (e *AzureEnv) createSubnet(ctx context.Context, resourceGroupName string) (*armnetwork.Subnet, error) {
	pollerResp, err := e.subnetsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		virtualNetworkName,
		networkSubnetName,
		armnetwork.Subnet{
			Properties: &armnetwork.SubnetPropertiesFormat{
				AddressPrefix: to.Ptr(networkSubnetAddress),
			},
		},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to begin creating subnet: %w", err)
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create subnet: %w", err)
	}

	return &resp.Subnet, nil
}

func (e *AzureEnv) createNIC(ctx context.Context, location, resourceGroupName, subnetID string) (*armnetwork.Interface, error) {
	pollerResp, err := e.interfacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		networkInterfaceName,
		armnetwork.Interface{
			Location: to.Ptr(location),
			Properties: &armnetwork.InterfacePropertiesFormat{
				IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
					{
						Name: to.Ptr("ipConfig"),
						Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
							PrivateIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodDynamic),
							Subnet: &armnetwork.Subnet{
								ID: to.Ptr(subnetID),
							},
						},
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to begin creating network interface: %w", err)
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create network interface: %w", err)
	}

	return &resp.Interface, nil
}

func (e *AzureEnv) createVirtualMachine(ctx context.Context, location, resourceGroupName, networkInterfaceID string) (*armcompute.VirtualMachine, error) {
	parameters := armcompute.VirtualMachine{
		Location: to.Ptr(location),
		Tags:     map[string]*string{"scanconfig": to.Ptr("test")},
		Identity: &armcompute.VirtualMachineIdentity{
			Type: to.Ptr(armcompute.ResourceIdentityTypeNone),
		},
		Properties: &armcompute.VirtualMachineProperties{
			StorageProfile: &armcompute.StorageProfile{
				ImageReference: imageReference,
				OSDisk: &armcompute.OSDisk{
					Name:         to.Ptr(osDiskName),
					CreateOption: to.Ptr(armcompute.DiskCreateOptionTypesFromImage),
					ManagedDisk: &armcompute.ManagedDiskParameters{
						StorageAccountType: to.Ptr(armcompute.StorageAccountTypesStandardSSDLRS), // OSDisk type Standard/Premium HDD/SSD
					},
				},
			},
			HardwareProfile: &armcompute.HardwareProfile{
				VMSize: to.Ptr(armcompute.VirtualMachineSizeTypes(virtualMachineSize)), // VM size include vCPUs,RAM,Data Disks,Temp storage.
			},
			OSProfile: &armcompute.OSProfile{ //
				ComputerName:  to.Ptr("vmclarity-asset-compute"),
				AdminUsername: to.Ptr(virtualMachineUserName),
				LinuxConfiguration: &armcompute.LinuxConfiguration{
					DisablePasswordAuthentication: to.Ptr(true),
					SSH: &armcompute.SSHConfiguration{
						PublicKeys: []*armcompute.SSHPublicKey{
							{
								Path:    to.Ptr(fmt.Sprintf("/home/%s/.ssh/authorized_keys", virtualMachineUserName)),
								KeyData: to.Ptr(string(e.sshKeyPair.PublicKey)),
							},
						},
					},
				},
			},
			NetworkProfile: &armcompute.NetworkProfile{
				NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
					{
						ID: to.Ptr(networkInterfaceID),
						Properties: &armcompute.NetworkInterfaceReferenceProperties{
							Primary: to.Ptr(true),
						},
					},
				},
			},
		},
	}

	pollerResponse, err := e.virtualMachinesClient.BeginCreateOrUpdate(ctx, resourceGroupName, virtualMachineName, parameters, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin creating virtual machine: %w", err)
	}

	resp, err := pollerResponse.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create virtual machine: %w", err)
	}

	return &resp.VirtualMachine, nil
}
