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

// nolint:wrapcheck
package scanner

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"

	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/provider/azure/utils"
	"github.com/openclarity/vmclarity/provider/cloudinit"
)

var (
	VMCreateEstimateProvisionTime = 2 * time.Minute
	VMDiskAttachEstimateTime      = 2 * time.Minute
	VMDeleteEstimateTime          = 2 * time.Minute
)

func scannerVMNameFromJobConfig(config *provider.ScanJobConfig) string {
	return "vmclarity-scanner-" + config.AssetScanID
}

func (s *Scanner) ensureScannerVirtualMachine(ctx context.Context, config *provider.ScanJobConfig, networkInterface armnetwork.Interface) (armcompute.VirtualMachine, error) {
	vmName := scannerVMNameFromJobConfig(config)

	vmResp, err := s.VMClient.Get(ctx, s.ScannerResourceGroup, vmName, nil)
	if err == nil {
		if *vmResp.VirtualMachine.Properties.ProvisioningState != provisioningStateSucceeded {
			return vmResp.VirtualMachine, provider.RetryableErrorf(VMCreateEstimateProvisionTime, "VM is not ready yet, provisioning state: %s", *vmResp.VirtualMachine.Properties.ProvisioningState)
		}
		return vmResp.VirtualMachine, nil
	}

	notFound, err := utils.HandleAzureRequestError(err, "getting scanner virtual machine: %s", vmName)
	if !notFound {
		return armcompute.VirtualMachine{}, err
	}

	userData, err := cloudinit.New(config)
	if err != nil {
		return armcompute.VirtualMachine{}, fmt.Errorf("failed to generate cloud-init: %w", err)
	}
	userDataBase64 := base64.StdEncoding.EncodeToString([]byte(userData))

	parameters := armcompute.VirtualMachine{
		Location: to.Ptr(s.ScannerLocation),
		Identity: &armcompute.VirtualMachineIdentity{
			// Scanners don't need access to Azure so no need for an Identity
			Type: to.Ptr(armcompute.ResourceIdentityTypeNone),
		},
		Properties: &armcompute.VirtualMachineProperties{
			HardwareProfile: &armcompute.HardwareProfile{
				VMSize: to.Ptr(armcompute.VirtualMachineSizeTypes(s.ScannerVMSize)),
			},
			StorageProfile: &armcompute.StorageProfile{
				ImageReference: &armcompute.ImageReference{
					Publisher: to.Ptr(s.ScannerImagePublisher),
					SKU:       to.Ptr(s.ScannerImageSKU),
					Version:   to.Ptr(s.ScannerImageVersion),
					Offer:     to.Ptr(s.ScannerImageOffer),
				},
				OSDisk: &armcompute.OSDisk{
					Name:         to.Ptr(vmName + "-rootvolume"),
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

	if s.ScannerPublicKey != "" {
		parameters.Properties.OSProfile.LinuxConfiguration.SSH = &armcompute.SSHConfiguration{
			PublicKeys: []*armcompute.SSHPublicKey{
				{
					Path:    to.Ptr(fmt.Sprintf("/home/%s/.ssh/authorized_keys", "vmclarity")),
					KeyData: to.Ptr(s.ScannerPublicKey),
				},
			},
		}
	}

	_, err = s.VMClient.BeginCreateOrUpdate(ctx, s.ScannerResourceGroup, vmName, parameters, nil)
	if err != nil {
		_, err = utils.HandleAzureRequestError(err, "creating virtual machine")
		return armcompute.VirtualMachine{}, err
	}

	return armcompute.VirtualMachine{}, provider.RetryableErrorf(VMCreateEstimateProvisionTime, "vm created")
}

func (s *Scanner) ensureScannerVirtualMachineDeleted(ctx context.Context, config *provider.ScanJobConfig) error {
	vmName := scannerVMNameFromJobConfig(config)

	return utils.EnsureDeleted(
		"virtual machine",
		func() error {
			_, err := s.VMClient.Get(ctx, s.ScannerResourceGroup, vmName, nil)
			return err
		},
		func() error {
			_, err := s.VMClient.BeginDelete(ctx, s.ScannerResourceGroup, vmName, nil)
			return err
		},
		VMDeleteEstimateTime,
	)
}

func (s *Scanner) ensureDiskAttachedToScannerVM(ctx context.Context, vm armcompute.VirtualMachine, disk armcompute.Disk) error {
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
				CreateOption: to.Ptr(armcompute.DiskCreateOptionTypesAttach),
				Lun:          to.Ptr[int32](0),
				ManagedDisk: &armcompute.ManagedDiskParameters{
					ID: disk.ID,
				},
				Name: disk.Name,
			},
		}

		_, err := s.VMClient.BeginCreateOrUpdate(ctx, s.ScannerResourceGroup, *vm.Name, vm, nil)
		if err != nil {
			_, err := utils.HandleAzureRequestError(err, "attaching disk %s to VM %s", *disk.Name, *vm.Name)
			return err
		}
	}

	diskResp, err := s.DisksClient.Get(ctx, s.ScannerResourceGroup, *disk.Name, nil)
	if err != nil {
		_, err := utils.HandleAzureRequestError(err, "getting disk %s", *disk.Name)
		return err
	}

	if *diskResp.Disk.Properties.DiskState != armcompute.DiskStateAttached {
		return provider.RetryableErrorf(VMDiskAttachEstimateTime, "volume is not yet attached, disk is in state: %v", *diskResp.Disk.Properties.DiskState)
	}

	return nil
}
