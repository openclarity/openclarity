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
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"

	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/provider/azure/utils"
	workflowTypes "github.com/openclarity/vmclarity/workflow/types"
)

const (
	EnsureAssetVMInfoTaskName     = "EnsureVMInfo"
	EnsureSnapshotTaskName        = "EnsureSnapshot"
	EnsureDiskTaskName            = "EnsureDisk"
	EnsureScannerVMTaskName       = "EnsureScannerVM"
	AttachDiskToScannerVMTaskName = "AttachDiskToScannerVM"

	EnsureBlobDeletedTaskName             = "EnsureBlobDeleted"
	EnsureSnapshotDeletedTaskName         = "EnsureSnapshotDeleted"
	EnsureTargetDiskDeletedTaskName       = "EnsureTargetDiskDeleted"
	EnsureScannerVMDeletedTaskName        = "EnsureScannerVMDeleted"
	EnsureNetworkInterfaceDeletedTaskName = "EnsureNetworkInterfaceDeleted"
)

type AssetScanState struct {
	config *provider.ScanJobConfig
	mu     *sync.RWMutex

	assetVM   armcompute.VirtualMachinesClientGetResponse
	scannerVM armcompute.VirtualMachine
	snapshot  armcompute.Snapshot
	disk      armcompute.Disk
}

func (s *AssetScanState) AddAssetVM(assetVM armcompute.VirtualMachinesClientGetResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.assetVM = assetVM
}

func (s *AssetScanState) AddScannerVM(scannerVM armcompute.VirtualMachine) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.scannerVM = scannerVM
}

func (s *AssetScanState) AddSnapshot(snapshot armcompute.Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.snapshot = snapshot
}

func (s *AssetScanState) AddDisk(disk armcompute.Disk) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.disk = disk
}

// Ensures that the virtual machine information is available in the workflow's state.
func (s *Scanner) EnsureAssetVMInfo(deps []string) {
	s.RunAssetScanTasks = append(s.RunAssetScanTasks, &workflowTypes.Task[*AssetScanState]{
		Name: EnsureAssetVMInfoTaskName,
		Deps: deps,
		Fn: func(ctx context.Context, state *AssetScanState) error {
			vmInfo, err := state.config.AssetInfo.AsVMInfo()
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

			state.AddAssetVM(assetVM)

			return nil
		},
	})
}

// Ensures that a snapshot is created for the asset virtual machine's root volume,
// and that the snapshot is deleted after the workflow has completed.
func (s *Scanner) EnsureSnapshotWithCleanup(deps []string) {
	s.RunAssetScanTasks = append(s.RunAssetScanTasks, &workflowTypes.Task[*AssetScanState]{
		Name: EnsureSnapshotTaskName,
		Deps: deps,
		Fn: func(ctx context.Context, state *AssetScanState) error {
			if reflect.DeepEqual(state.assetVM.VirtualMachine, armcompute.VirtualMachine{}) {
				return provider.FatalErrorf("assetVM is not available in the AssetScanState")
			}

			snapshot, err := s.ensureSnapshotForVMRootVolume(ctx, state.config, state.assetVM.VirtualMachine)
			if err != nil {
				return fmt.Errorf("failed to ensure snapshot for vm root volume: %w", err)
			}

			state.AddSnapshot(snapshot)

			return nil
		},
	})

	s.RemoveAssetScanTasks = append(s.RemoveAssetScanTasks, &workflowTypes.Task[*AssetScanState]{
		Name: EnsureSnapshotDeletedTaskName,
		Deps: nil,
		Fn: func(ctx context.Context, state *AssetScanState) error {
			err := s.ensureSnapshotDeleted(ctx, state.config)
			if err != nil {
				return fmt.Errorf("failed to ensure snapshot deleted: %w", err)
			}

			return nil
		},
	})
}

// Ensures that a managed disk is created from the snapshot,
// and that the disk and the copied blob is deleted after the workflow has completed.
func (s *Scanner) EnsureDiskWithCleanup(deps []string) {
	s.RunAssetScanTasks = append(s.RunAssetScanTasks, &workflowTypes.Task[*AssetScanState]{
		Name: EnsureDiskTaskName,
		Deps: deps,
		Fn: func(ctx context.Context, state *AssetScanState) error {
			var disk armcompute.Disk
			var err error

			if reflect.DeepEqual(state.assetVM.VirtualMachine, armcompute.VirtualMachine{}) {
				return provider.FatalErrorf("assetVM is not available in the AssetScanState")
			}

			if reflect.DeepEqual(state.snapshot, armcompute.Snapshot{}) {
				return provider.FatalErrorf("snapshot is not available in the AssetScanState")
			}

			if *state.assetVM.Location == s.ScannerLocation {
				disk, err = s.ensureManagedDiskFromSnapshot(ctx, state.config, state.snapshot)
				if err != nil {
					return fmt.Errorf("failed to ensure managed disk created from snapshot: %w", err)
				}
			} else {
				disk, err = s.ensureManagedDiskFromSnapshotInDifferentRegion(ctx, state.config, state.snapshot)
				if err != nil {
					return fmt.Errorf("failed to ensure managed disk from snapshot in different region: %w", err)
				}
			}

			state.AddDisk(disk)

			return nil
		},
	})

	s.RemoveAssetScanTasks = append(s.RemoveAssetScanTasks, &workflowTypes.Task[*AssetScanState]{
		Name: EnsureTargetDiskDeletedTaskName,
		Deps: []string{EnsureScannerVMDeletedTaskName},
		Fn: func(ctx context.Context, state *AssetScanState) error {
			err := s.ensureTargetDiskDeleted(ctx, state.config)
			if err != nil {
				return fmt.Errorf("failed to ensure asset disk deleted: %w", err)
			}

			return nil
		},
	})

	// In case of cross-region snapshot, we need to ensure the copied blob is deleted as well.
	s.RemoveAssetScanTasks = append(s.RemoveAssetScanTasks, &workflowTypes.Task[*AssetScanState]{
		Name: EnsureBlobDeletedTaskName,
		Deps: nil,
		Fn: func(ctx context.Context, state *AssetScanState) error {
			err := s.ensureBlobDeleted(ctx, state.config)
			if err != nil {
				return fmt.Errorf("failed to ensure snapshot copy blob deleted: %w", err)
			}

			return nil
		},
	})
}

// Ensures that a network interface and the scanner virtual machine is created,
// and that the network interface and the scanner virtual machine are deleted after the workflow has completed.
func (s *Scanner) EnsureScannerVMWithCleanup(deps []string) {
	s.RunAssetScanTasks = append(s.RunAssetScanTasks, &workflowTypes.Task[*AssetScanState]{
		Name: EnsureScannerVMTaskName,
		Deps: deps,
		Fn: func(ctx context.Context, state *AssetScanState) error {
			networkInterface, err := s.ensureNetworkInterface(ctx, state.config)
			if err != nil {
				return fmt.Errorf("failed to ensure scanner network interface: %w", err)
			}

			scannerVM, err := s.ensureScannerVirtualMachine(ctx, state.config, networkInterface)
			if err != nil {
				return fmt.Errorf("failed to ensure scanner virtual machine: %w", err)
			}

			state.AddScannerVM(scannerVM)

			return nil
		},
	})

	s.RemoveAssetScanTasks = append(s.RemoveAssetScanTasks, &workflowTypes.Task[*AssetScanState]{
		Name: EnsureScannerVMDeletedTaskName,
		Deps: nil,
		Fn: func(ctx context.Context, state *AssetScanState) error {
			err := s.ensureScannerVirtualMachineDeleted(ctx, state.config)
			if err != nil {
				return fmt.Errorf("failed to ensure scanner virtual machine deleted: %w", err)
			}
			return nil
		},
	})

	s.RemoveAssetScanTasks = append(s.RemoveAssetScanTasks, &workflowTypes.Task[*AssetScanState]{
		Name: EnsureNetworkInterfaceDeletedTaskName,
		Deps: []string{EnsureScannerVMDeletedTaskName},
		Fn: func(ctx context.Context, state *AssetScanState) error {
			err := s.ensureNetworkInterfaceDeleted(ctx, state.config)
			if err != nil {
				return fmt.Errorf("failed to ensure network interface deleted: %w", err)
			}

			return nil
		},
	})
}

// Ensures that the asset disk is attached to the scanner virtual machine.
func (s *Scanner) EnsureAttachDiskToScannerVM(deps []string) {
	s.RunAssetScanTasks = append(s.RunAssetScanTasks, &workflowTypes.Task[*AssetScanState]{
		Name: AttachDiskToScannerVMTaskName,
		Deps: deps,
		Fn: func(ctx context.Context, state *AssetScanState) error {
			if reflect.DeepEqual(state.assetVM.VirtualMachine, armcompute.VirtualMachine{}) {
				return provider.FatalErrorf("assetVM is not available in the AssetScanState")
			}

			if reflect.DeepEqual(state.disk, armcompute.Disk{}) {
				return provider.FatalErrorf("disk is not available in the AssetScanState")
			}

			err := s.ensureDiskAttachedToScannerVM(ctx, state.scannerVM, state.disk)
			if err != nil {
				return fmt.Errorf("failed to ensure asset disk is attached to virtual machine: %w", err)
			}

			return nil
		},
	})
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
