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
	"encoding/base64"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v3"

	"github.com/openclarity/vmclarity/pkg/orchestrator/provider"
	"github.com/openclarity/vmclarity/pkg/orchestrator/provider/cloudinit"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

var (
	VMCreateEstimateProvisionTime = 2 * time.Minute
	VMDiskAttachEstimateTime      = 2 * time.Minute
	VMDeleteEstimateTime          = 2 * time.Minute
)

func scannerVMNameFromJobConfig(config *provider.ScanJobConfig) string {
	return fmt.Sprintf("vmclarity-scanner-%s", config.AssetScanID)
}

func (p *Provider) ensureScannerVirtualMachine(ctx context.Context, config *provider.ScanJobConfig, networkInterface armnetwork.Interface) (armcompute.VirtualMachine, error) {
	vmName := scannerVMNameFromJobConfig(config)

	vmResp, err := p.vmClient.Get(ctx, p.config.ScannerResourceGroup, vmName, nil)
	if err == nil {
		if *vmResp.VirtualMachine.Properties.ProvisioningState != ProvisioningStateSucceeded {
			return vmResp.VirtualMachine, provider.RetryableErrorf(VMCreateEstimateProvisionTime, "VM is not ready yet, provisioning state: %s", *vmResp.VirtualMachine.Properties.ProvisioningState)
		}
		return vmResp.VirtualMachine, nil
	}

	notFound, err := handleAzureRequestError(err, "getting scanner virtual machine: %s", vmName)
	if !notFound {
		return armcompute.VirtualMachine{}, err
	}

	userData, err := cloudinit.New(config)
	if err != nil {
		return armcompute.VirtualMachine{}, fmt.Errorf("failed to generate cloud-init: %w", err)
	}
	userDataBase64 := base64.StdEncoding.EncodeToString([]byte(userData))

	parameters := armcompute.VirtualMachine{
		Location: to.Ptr(p.config.ScannerLocation),
		Identity: &armcompute.VirtualMachineIdentity{
			// Scanners don't need access to Azure so no need for an Identity
			Type: to.Ptr(armcompute.ResourceIdentityTypeNone),
		},
		Properties: &armcompute.VirtualMachineProperties{
			HardwareProfile: &armcompute.HardwareProfile{
				VMSize: to.Ptr(armcompute.VirtualMachineSizeTypes(p.config.ScannerVMSize)),
			},
			StorageProfile: &armcompute.StorageProfile{
				ImageReference: &armcompute.ImageReference{
					Offer:     to.Ptr(p.config.ScannerImageOffer),
					Publisher: to.Ptr(p.config.ScannerImagePublisher),
					SKU:       to.Ptr(p.config.ScannerImageSKU),
					Version:   to.Ptr(p.config.ScannerImageVersion),
				},
				OSDisk: &armcompute.OSDisk{
					Name:         to.Ptr(fmt.Sprintf("%s-rootvolume", vmName)),
					CreateOption: to.Ptr(armcompute.DiskCreateOptionTypesFromImage),
					// Delete disk on VM delete
					DeleteOption: to.Ptr(armcompute.DiskDeleteOptionTypesDelete),
					Caching:      to.Ptr(armcompute.CachingTypesReadWrite),
					ManagedDisk: &armcompute.ManagedDiskParameters{
						// OSDisk type Standard/Premium HDD/SSD
						StorageAccountType: to.Ptr(armcompute.StorageAccountTypesStandardLRS),
					},
					// DiskSizeGB: to.Ptr[int32](100), // default 127G
				},
			},
			OSProfile: &armcompute.OSProfile{ // use username/password
				ComputerName:  to.Ptr(vmName),
				AdminUsername: to.Ptr("vmclarity"),
				LinuxConfiguration: &armcompute.LinuxConfiguration{
					DisablePasswordAuthentication: to.Ptr(true),
				},
			},
			NetworkProfile: &armcompute.NetworkProfile{
				NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
					{
						ID: networkInterface.ID,
					},
				},
			},
			UserData: &userDataBase64,
		},
	}

	if p.config.ScannerPublicKey != "" {
		parameters.Properties.OSProfile.LinuxConfiguration.SSH = &armcompute.SSHConfiguration{
			PublicKeys: []*armcompute.SSHPublicKey{
				{
					Path:    to.Ptr(fmt.Sprintf("/home/%s/.ssh/authorized_keys", "vmclarity")),
					KeyData: to.Ptr(string(p.config.ScannerPublicKey)),
				},
			},
		}
	}

	_, err = p.vmClient.BeginCreateOrUpdate(ctx, p.config.ScannerResourceGroup, vmName, parameters, nil)
	if err != nil {
		_, err = handleAzureRequestError(err, "creating virtual machine")
		return armcompute.VirtualMachine{}, err
	}

	return armcompute.VirtualMachine{}, provider.RetryableErrorf(VMCreateEstimateProvisionTime, "vm created")
}

func (p *Provider) ensureScannerVirtualMachineDeleted(ctx context.Context, config *provider.ScanJobConfig) error {
	vmName := scannerVMNameFromJobConfig(config)

	return ensureDeleted(
		"virtual machine",
		func() error {
			_, err := p.vmClient.Get(ctx, p.config.ScannerResourceGroup, vmName, nil)
			return err // nolint: wrapcheck
		},
		func() error {
			_, err := p.vmClient.BeginDelete(ctx, p.config.ScannerResourceGroup, vmName, nil)
			return err // nolint: wrapcheck
		},
		VMDeleteEstimateTime,
	)
}

func (p *Provider) ensureDiskAttachedToScannerVM(ctx context.Context, vm armcompute.VirtualMachine, disk armcompute.Disk) error {
	var vmAttachedToDisk bool
	for _, dataDisk := range vm.Properties.StorageProfile.DataDisks {
		if dataDisk.ManagedDisk.ID == disk.ID {
			vmAttachedToDisk = true
			break
		}
	}

	if !vmAttachedToDisk {
		vm.Properties.StorageProfile.DataDisks = []*armcompute.DataDisk{
			{
				CreateOption: utils.PointerTo(armcompute.DiskCreateOptionTypesAttach),
				Lun:          utils.PointerTo[int32](0),
				ManagedDisk: &armcompute.ManagedDiskParameters{
					ID: disk.ID,
				},
				Name: disk.Name,
			},
		}

		_, err := p.vmClient.BeginCreateOrUpdate(ctx, p.config.ScannerResourceGroup, *vm.Name, vm, nil)
		if err != nil {
			_, err := handleAzureRequestError(err, "attaching disk %s to VM %s", *disk.Name, *vm.Name)
			return err
		}
	}

	diskResp, err := p.disksClient.Get(ctx, p.config.ScannerResourceGroup, *disk.Name, nil)
	if err != nil {
		_, err := handleAzureRequestError(err, "getting disk %s", *disk.Name)
		return err
	}

	if *diskResp.Disk.Properties.DiskState != armcompute.DiskStateAttached {
		return provider.RetryableErrorf(VMDiskAttachEstimateTime, "volume is not yet attached, disk is in state: %v", *diskResp.Disk.Properties.DiskState)
	}

	return nil
}
