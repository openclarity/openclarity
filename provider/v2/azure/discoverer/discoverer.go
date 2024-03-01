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

package discoverer

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/provider"
)

type Discoverer struct {
	VMClient    *armcompute.VirtualMachinesClient
	DisksClient *armcompute.DisksClient
}

func (d *Discoverer) DiscoverAssets(ctx context.Context) provider.AssetDiscoverer {
	assetDiscoverer := provider.NewSimpleAssetDiscoverer()

	go func() {
		defer close(assetDiscoverer.OutputChan)

		// list all vms in all resourceGroups in the subscription
		res := d.VMClient.NewListAllPager(nil)
		for res.More() {
			page, err := res.NextPage(ctx)
			if err != nil {
				assetDiscoverer.Error = fmt.Errorf("failed to get next page: %w", err)
				return
			}
			ts, err := d.processVirtualMachineListIntoAssetTypes(ctx, page.VirtualMachineListResult)
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

func (d *Discoverer) processVirtualMachineListIntoAssetTypes(ctx context.Context, vmList armcompute.VirtualMachineListResult) ([]apitypes.AssetType, error) {
	ret := make([]apitypes.AssetType, 0, len(vmList.Value))
	for _, vm := range vmList.Value {
		info, err := getVMInfoFromVirtualMachine(vm, d.getRootVolumeInfo(ctx, vm))
		if err != nil {
			return nil, fmt.Errorf("unable to convert instance to vminfo: %w", err)
		}
		ret = append(ret, info)
	}
	return ret, nil
}

func (d *Discoverer) getRootVolumeInfo(ctx context.Context, vm *armcompute.VirtualMachine) *apitypes.RootVolume {
	logger := log.GetLoggerFromContextOrDiscard(ctx)
	ret := &apitypes.RootVolume{
		SizeGB:    int(to.ValueOrZero(vm.Properties.StorageProfile.OSDisk.DiskSizeGB)),
		Encrypted: apitypes.RootVolumeEncryptedUnknown,
	}
	osDiskID, err := arm.ParseResourceID(to.ValueOrZero(vm.Properties.StorageProfile.OSDisk.ManagedDisk.ID))
	if err != nil {
		logger.Warnf("Failed to parse disk ID. DiskID=%v: %v",
			to.ValueOrZero(vm.Properties.StorageProfile.OSDisk.ManagedDisk.ID), err)
		return ret
	}
	osDisk, err := d.DisksClient.Get(ctx, osDiskID.ResourceGroupName, osDiskID.Name, nil)
	if err != nil {
		logger.Warnf("Failed to get OS disk. DiskID=%v: %v",
			to.ValueOrZero(vm.Properties.StorageProfile.OSDisk.ManagedDisk.ID), err)
		return ret
	}
	ret.Encrypted = isEncrypted(osDisk)
	ret.SizeGB = int(to.ValueOrZero(osDisk.Disk.Properties.DiskSizeGB))

	return ret
}

func getVMInfoFromVirtualMachine(vm *armcompute.VirtualMachine, rootVol *apitypes.RootVolume) (apitypes.AssetType, error) {
	assetType := apitypes.AssetType{}
	err := assetType.FromVMInfo(apitypes.VMInfo{
		ObjectType:       "VMInfo",
		InstanceProvider: to.Ptr(apitypes.Azure),
		InstanceID:       *vm.ID,
		Image:            createImageURN(vm.Properties.StorageProfile.ImageReference),
		InstanceType:     *vm.Type,
		LaunchTime:       *vm.Properties.TimeCreated,
		Location:         *vm.Location,
		Platform:         string(*vm.Properties.StorageProfile.OSDisk.OSType),
		RootVolume:       *rootVol,
		SecurityGroups:   &[]apitypes.SecurityGroup{},
		Tags:             convertTags(vm.Tags),
	})
	if err != nil {
		err = fmt.Errorf("failed to create AssetType from VMInfo: %w", err)
	}

	return assetType, err
}

func isEncrypted(disk armcompute.DisksClientGetResponse) apitypes.RootVolumeEncrypted {
	if disk.Properties.EncryptionSettingsCollection == nil {
		return apitypes.RootVolumeEncryptedNo
	}
	if *disk.Properties.EncryptionSettingsCollection.Enabled {
		return apitypes.RootVolumeEncryptedYes
	}

	return apitypes.RootVolumeEncryptedNo
}

func convertTags(tags map[string]*string) *[]apitypes.Tag {
	ret := make([]apitypes.Tag, 0, len(tags))
	for key, val := range tags {
		ret = append(ret, apitypes.Tag{
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
