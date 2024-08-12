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
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"

	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/provider/azure/utils"
)

const (
	provisioningStateSucceeded = "Succeeded"
	instanceIDPartsLength      = 9
	resourceGroupPartIdx       = 4
	vmNamePartIdx              = 8
)

type Scanner struct {
	Cred             azcore.TokenCredential
	VMClient         *armcompute.VirtualMachinesClient
	SnapshotsClient  *armcompute.SnapshotsClient
	DisksClient      *armcompute.DisksClient
	InterfacesClient *armnetwork.InterfacesClient

	SubscriptionID              string
	ScannerLocation             string
	ScannerResourceGroup        string
	ScannerSubnet               string
	ScannerPublicKey            string
	ScannerVMSize               string
	ScannerImagePublisher       string
	ScannerImageOffer           string
	ScannerImageSKU             string
	ScannerImageVersion         string
	ScannerSecurityGroup        string
	ScannerStorageAccountName   string
	ScannerStorageContainerName string
}

// nolint:cyclop
func (s *Scanner) RunAssetScan(ctx context.Context, config *provider.ScanJobConfig) error {
	vmInfo, err := config.AssetInfo.AsVMInfo()
	if err != nil {
		return provider.FatalErrorf("unable to get vminfo from asset: %w", err)
	}

	resourceGroup, vmName, err := resourceGroupAndNameFromInstanceID(vmInfo.InstanceID)
	if err != nil {
		return err
	}

	assetVM, err := s.VMClient.Get(ctx, resourceGroup, vmName, nil)
	if err != nil {
		_, err = utils.HandleAzureRequestError(err, "getting asset virtual machine %s", vmName)
		return err
	}

	snapshot, err := s.ensureSnapshotForVMRootVolume(ctx, config, assetVM.VirtualMachine)
	if err != nil {
		return fmt.Errorf("failed to ensure snapshot for vm root volume: %w", err)
	}

	var disk armcompute.Disk
	if *assetVM.Location == s.ScannerLocation {
		disk, err = s.ensureManagedDiskFromSnapshot(ctx, config, snapshot)
		if err != nil {
			return fmt.Errorf("failed to ensure managed disk created from snapshot: %w", err)
		}
	} else {
		disk, err = s.ensureManagedDiskFromSnapshotInDifferentRegion(ctx, config, snapshot)
		if err != nil {
			return fmt.Errorf("failed to ensure managed disk from snapshot in different region: %w", err)
		}
	}

	networkInterface, err := s.ensureNetworkInterface(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure scanner network interface: %w", err)
	}

	scannerVM, err := s.ensureScannerVirtualMachine(ctx, config, networkInterface)
	if err != nil {
		return fmt.Errorf("failed to ensure scanner virtual machine: %w", err)
	}

	err = s.ensureDiskAttachedToScannerVM(ctx, scannerVM, disk)
	if err != nil {
		return fmt.Errorf("failed to ensure asset disk is attached to virtual machine: %w", err)
	}

	return nil
}

func (s *Scanner) RemoveAssetScan(ctx context.Context, config *provider.ScanJobConfig) error {
	err := s.ensureScannerVirtualMachineDeleted(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure scanner virtual machine deleted: %w", err)
	}

	err = s.ensureNetworkInterfaceDeleted(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure network interface deleted: %w", err)
	}

	err = s.ensureTargetDiskDeleted(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure asset disk deleted: %w", err)
	}

	err = s.ensureBlobDeleted(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure snapshot copy blob deleted: %w", err)
	}

	err = s.ensureSnapshotDeleted(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure snapshot deleted: %w", err)
	}

	return nil
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
