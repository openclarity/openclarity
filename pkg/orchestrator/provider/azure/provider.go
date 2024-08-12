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
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v3"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/orchestrator/provider"
	"github.com/openclarity/vmclarity/pkg/shared/log"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

const (
	instanceIDPartsLength = 9
	resourceGroupPartIdx  = 4
	vmNamePartIdx         = 8
)

type Provider struct {
	cred             azcore.TokenCredential
	rgClient         *armresources.ResourceGroupsClient
	vmClient         *armcompute.VirtualMachinesClient
	snapshotsClient  *armcompute.SnapshotsClient
	disksClient      *armcompute.DisksClient
	interfacesClient *armnetwork.InterfacesClient

	config *Config
}

func New(_ context.Context) (*Provider, error) {
	config, err := NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	err = config.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate configuration: %w", err)
	}

	cred, err := azidentity.NewManagedIdentityCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed create managed identity credential: %w", err)
	}

	rgClient, err := armresources.NewResourceGroupsClient(config.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource group client: %w", err)
	}

	networkClientFactory, err := armnetwork.NewClientFactory(config.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create network client factory: %w", err)
	}

	computeClientFactory, err := armcompute.NewClientFactory(config.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client factory: %w", err)
	}

	return &Provider{
		cred:             cred,
		rgClient:         rgClient,
		vmClient:         computeClientFactory.NewVirtualMachinesClient(),
		snapshotsClient:  computeClientFactory.NewSnapshotsClient(),
		disksClient:      computeClientFactory.NewDisksClient(),
		interfacesClient: networkClientFactory.NewInterfacesClient(),
		config:           config,
	}, nil
}

func (p *Provider) Kind() models.CloudProvider {
	return models.Azure
}

func (p *Provider) Estimate(ctx context.Context, stats models.AssetScanStats, asset *models.Asset, assetScanTemplate *models.AssetScanTemplate) (*models.Estimation, error) {
	return &models.Estimation{}, provider.FatalErrorf("Not Implemented")
}

// nolint:cyclop
func (p *Provider) RunAssetScan(ctx context.Context, config *provider.ScanJobConfig) error {
	vmInfo, err := config.AssetInfo.AsVMInfo()
	if err != nil {
		return provider.FatalErrorf("unable to get vminfo from asset: %w", err)
	}

	resourceGroup, vmName, err := resourceGroupAndNameFromInstanceID(vmInfo.InstanceID)
	if err != nil {
		return err
	}

	assetVM, err := p.vmClient.Get(ctx, resourceGroup, vmName, nil)
	if err != nil {
		_, err = handleAzureRequestError(err, "getting asset virtual machine %s", vmName)
		return err
	}

	snapshot, err := p.ensureSnapshotForVMRootVolume(ctx, config, assetVM.VirtualMachine)
	if err != nil {
		return fmt.Errorf("failed to ensure snapshot for vm root volume: %w", err)
	}

	var disk armcompute.Disk
	if *assetVM.Location == p.config.ScannerLocation {
		disk, err = p.ensureManagedDiskFromSnapshot(ctx, config, snapshot)
		if err != nil {
			return fmt.Errorf("failed to ensure managed disk created from snapshot: %w", err)
		}
	} else {
		disk, err = p.ensureManagedDiskFromSnapshotInDifferentRegion(ctx, config, snapshot)
		if err != nil {
			return fmt.Errorf("failed to ensure managed disk from snapshot in different region: %w", err)
		}
	}

	networkInterface, err := p.ensureNetworkInterface(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure scanner network interface: %w", err)
	}

	scannerVM, err := p.ensureScannerVirtualMachine(ctx, config, networkInterface)
	if err != nil {
		return fmt.Errorf("failed to ensure scanner virtual machine: %w", err)
	}

	err = p.ensureDiskAttachedToScannerVM(ctx, scannerVM, disk)
	if err != nil {
		return fmt.Errorf("failed to ensure asset disk is attached to virtual machine: %w", err)
	}

	return nil
}

func (p *Provider) RemoveAssetScan(ctx context.Context, config *provider.ScanJobConfig) error {
	err := p.ensureScannerVirtualMachineDeleted(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure scanner virtual machine deleted: %w", err)
	}

	err = p.ensureNetworkInterfaceDeleted(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure network interface deleted: %w", err)
	}

	err = p.ensureTargetDiskDeleted(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure asset disk deleted: %w", err)
	}

	err = p.ensureBlobDeleted(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure snapshot copy blob deleted: %w", err)
	}

	err = p.ensureSnapshotDeleted(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure snapshot deleted: %w", err)
	}

	return nil
}

// nolint: cyclop
func (p *Provider) DiscoverAssets(ctx context.Context) provider.AssetDiscoverer {
	assetDiscoverer := provider.NewSimpleAssetDiscoverer()

	go func() {
		defer close(assetDiscoverer.OutputChan)

		// list all vms in all resourceGroups in the subscription
		res := p.vmClient.NewListAllPager(nil)
		for res.More() {
			page, err := res.NextPage(ctx)
			if err != nil {
				assetDiscoverer.Error = fmt.Errorf("failed to get next page: %w", err)
				return
			}
			ts, err := p.processVirtualMachineListIntoAssetTypes(ctx, page.VirtualMachineListResult)
			if err != nil {
				assetDiscoverer.Error = err
				return
			}

			for _, asset := range ts {
				select {
				case assetDiscoverer.OutputChan <- asset:
				case <-ctx.Done():
					assetDiscoverer.Error = ctx.Err()
					return
				}
			}
		}
	}()

	return assetDiscoverer
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

func (p *Provider) processVirtualMachineListIntoAssetTypes(ctx context.Context, vmList armcompute.VirtualMachineListResult) ([]models.AssetType, error) {
	ret := make([]models.AssetType, 0, len(vmList.Value))
	for _, vm := range vmList.Value {
		info, err := getVMInfoFromVirtualMachine(vm, p.getRootVolumeInfo(ctx, vm))
		if err != nil {
			return nil, fmt.Errorf("unable to convert instance to vminfo: %w", err)
		}
		ret = append(ret, info)
	}
	return ret, nil
}

func (p *Provider) getRootVolumeInfo(ctx context.Context, vm *armcompute.VirtualMachine) *models.RootVolume {
	logger := log.GetLoggerFromContextOrDiscard(ctx)
	ret := &models.RootVolume{
		SizeGB:    int(utils.Int32PointerValOrEmpty(vm.Properties.StorageProfile.OSDisk.DiskSizeGB)),
		Encrypted: models.RootVolumeEncryptedUnknown,
	}
	osDiskID, err := arm.ParseResourceID(utils.StringPointerValOrEmpty(vm.Properties.StorageProfile.OSDisk.ManagedDisk.ID))
	if err != nil {
		logger.Warnf("Failed to parse disk ID. DiskID=%v: %v",
			utils.StringPointerValOrEmpty(vm.Properties.StorageProfile.OSDisk.ManagedDisk.ID), err)
		return ret
	}
	osDisk, err := p.disksClient.Get(ctx, osDiskID.ResourceGroupName, osDiskID.Name, nil)
	if err != nil {
		logger.Warnf("Failed to get OS disk. DiskID=%v: %v",
			utils.StringPointerValOrEmpty(vm.Properties.StorageProfile.OSDisk.ManagedDisk.ID), err)
		return ret
	}
	ret.Encrypted = isEncrypted(osDisk)
	ret.SizeGB = int(utils.Int32PointerValOrEmpty(osDisk.Disk.Properties.DiskSizeGB))

	return ret
}

func getVMInfoFromVirtualMachine(vm *armcompute.VirtualMachine, rootVol *models.RootVolume) (models.AssetType, error) {
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
		RootVolume:       *rootVol,
		SecurityGroups:   &[]models.SecurityGroup{},
		Tags:             convertTags(vm.Tags),
	})
	if err != nil {
		err = fmt.Errorf("failed to create AssetType from VMInfo: %w", err)
	}

	return assetType, err
}

func isEncrypted(disk armcompute.DisksClientGetResponse) models.RootVolumeEncrypted {
	if disk.Properties.EncryptionSettingsCollection == nil {
		return models.RootVolumeEncryptedNo
	}
	if *disk.Properties.EncryptionSettingsCollection.Enabled {
		return models.RootVolumeEncryptedYes
	}

	return models.RootVolumeEncryptedNo
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
