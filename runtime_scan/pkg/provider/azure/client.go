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
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v3"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

const (
	instanceIDPartsLength = 9
	resourceGroupPartIdx  = 4
	vmNamePartIdx         = 8
)

type Client struct {
	cred             azcore.TokenCredential
	rgClient         *armresources.ResourceGroupsClient
	vmClient         *armcompute.VirtualMachinesClient
	snapshotsClient  *armcompute.SnapshotsClient
	disksClient      *armcompute.DisksClient
	interfacesClient *armnetwork.InterfacesClient

	azureConfig Config
}

func New(_ context.Context) (*Client, error) {
	config, err := NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	err = config.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate configuration: %w", err)
	}

	client := Client{
		azureConfig: config,
	}

	cred, err := azidentity.NewManagedIdentityCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed create managed identity credential: %w", err)
	}
	client.cred = cred

	client.rgClient, err = armresources.NewResourceGroupsClient(config.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource group client: %w", err)
	}

	networkClientFactory, err := armnetwork.NewClientFactory(config.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create network client factory: %w", err)
	}
	client.interfacesClient = networkClientFactory.NewInterfacesClient()

	computeClientFactory, err := armcompute.NewClientFactory(config.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client factory: %w", err)
	}
	client.vmClient = computeClientFactory.NewVirtualMachinesClient()
	client.disksClient = computeClientFactory.NewDisksClient()
	client.snapshotsClient = computeClientFactory.NewSnapshotsClient()

	return &client, nil
}

func (c Client) Kind() models.CloudProvider {
	return models.Azure
}

// nolint:cyclop
func (c *Client) RunAssetScan(ctx context.Context, config *provider.ScanJobConfig) error {
	vmInfo, err := config.AssetInfo.AsVMInfo()
	if err != nil {
		return provider.FatalErrorf("unable to get vminfo from asset: %w", err)
	}

	resourceGroup, vmName, err := resourceGroupAndNameFromInstanceID(vmInfo.InstanceID)
	if err != nil {
		return err
	}

	assetVM, err := c.vmClient.Get(ctx, resourceGroup, vmName, nil)
	if err != nil {
		_, err = handleAzureRequestError(err, "getting asset virtual machine %s", vmName)
		return err
	}

	snapshot, err := c.ensureSnapshotForVMRootVolume(ctx, config, assetVM.VirtualMachine)
	if err != nil {
		return fmt.Errorf("failed to ensure snapshot for vm root volume: %w", err)
	}

	var disk armcompute.Disk
	if *assetVM.Location == c.azureConfig.ScannerLocation {
		disk, err = c.ensureManagedDiskFromSnapshot(ctx, config, snapshot)
		if err != nil {
			return fmt.Errorf("failed to ensure managed disk created from snapshot: %w", err)
		}
	} else {
		disk, err = c.ensureManagedDiskFromSnapshotInDifferentRegion(ctx, config, snapshot)
		if err != nil {
			return fmt.Errorf("failed to ensure managed disk from snapshot in different region: %w", err)
		}
	}

	networkInterface, err := c.ensureNetworkInterface(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure scanner network interface: %w", err)
	}

	scannerVM, err := c.ensureScannerVirtualMachine(ctx, config, networkInterface)
	if err != nil {
		return fmt.Errorf("failed to ensure scanner virtual machine: %w", err)
	}

	err = c.ensureDiskAttachedToScannerVM(ctx, scannerVM, disk)
	if err != nil {
		return fmt.Errorf("failed to ensure asset disk is attached to virtual machine: %w", err)
	}

	return nil
}

func (c *Client) RemoveAssetScan(ctx context.Context, config *provider.ScanJobConfig) error {
	err := c.ensureScannerVirtualMachineDeleted(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure scanner virtual machine deleted: %w", err)
	}

	err = c.ensureNetworkInterfaceDeleted(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure network interface deleted: %w", err)
	}

	err = c.ensureTargetDiskDeleted(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure asset disk deleted: %w", err)
	}

	err = c.ensureBlobDeleted(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure snapshot copy blob deleted: %w", err)
	}

	err = c.ensureSnapshotDeleted(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure snapshot deleted: %w", err)
	}

	return nil
}

// nolint: cyclop
func (c *Client) DiscoverAssets(ctx context.Context) ([]models.AssetType, error) {
	var ret []models.AssetType
	// list all vms in all resourceGroups in the subscription
	res := c.vmClient.NewListAllPager(nil)
	for res.More() {
		page, err := res.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get next page: %w", err)
		}
		ts, err := processVirtualMachineListIntoAssetTypes(page.VirtualMachineListResult)
		if err != nil {
			return nil, err
		}
		ret = append(ret, ts...)
	}
	return ret, nil
}

// Example Instance ID:
//
// /subscriptions/ecad88af-09d5-4725-8d80-906e51fddf02/resourceGroups/vmclarity-sambetts-dev/providers/Microsoft.Compute/virtualMachines/vmclarity-server
//
// Will return "vmclarity-sambetts-dev" and "vmclarity-server".
func resourceGroupAndNameFromInstanceID(instanceID string) (string, string, error) {
	idParts := strings.Split(instanceID, "/")
	if len(idParts) != instanceIDPartsLength {
		return "", "", provider.FatalErrorf("asset instance id in unexpected format got: %s", idParts)
	}
	return idParts[resourceGroupPartIdx], idParts[vmNamePartIdx], nil
}

func processVirtualMachineListIntoAssetTypes(vmList armcompute.VirtualMachineListResult) ([]models.AssetType, error) {
	ret := make([]models.AssetType, 0, len(vmList.Value))
	for _, vm := range vmList.Value {
		info, err := getVMInfoFromVirtualMachine(vm)
		if err != nil {
			return nil, fmt.Errorf("unable to convert instance to vminfo: %w", err)
		}
		ret = append(ret, info)
	}
	return ret, nil
}

func getVMInfoFromVirtualMachine(vm *armcompute.VirtualMachine) (models.AssetType, error) {
	assetType := models.AssetType{}
	err := assetType.FromVMInfo(models.VMInfo{
		ObjectType:       "VMInfo",
		InstanceProvider: utils.PointerTo(models.Azure),
		InstanceID:       *vm.ID,
		Image:            createImageURN(vm.Properties.StorageProfile.ImageReference),
		InstanceType:     *vm.Type,
		LaunchTime:       *vm.Properties.TimeCreated,
		Location:         *vm.Location,
		Platform:         string(*vm.Properties.StorageProfile.OSDisk.OSType),
		SecurityGroups:   &[]models.SecurityGroup{},
		Tags:             convertTags(vm.Tags),
	})
	if err != nil {
		err = fmt.Errorf("failed to create AssetType from VMInfo: %w", err)
	}

	return assetType, err
}

func convertTags(tags map[string]*string) *[]models.Tag {
	ret := make([]models.Tag, 0, len(tags))
	for key, val := range tags {
		ret = append(ret, models.Tag{
			Key:   key,
			Value: *val,
		})
	}
	return &ret
}

// https://learn.microsoft.com/en-us/azure/virtual-machines/linux/tutorial-manage-vm#understand-vm-images
func createImageURN(reference *armcompute.ImageReference) string {
	// ImageReference is required only when using platform images, marketplace images, or
	// virtual machine images, but is not used in other creation operations (like managed disks).
	if reference == nil {
		return ""
	}
	return *reference.Publisher + "/" + *reference.Offer + "/" + *reference.SKU + "/" + *reference.Version
}
