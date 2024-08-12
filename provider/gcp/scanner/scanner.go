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

package scanner

import (
	"context"
	"errors"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/sirupsen/logrus"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/provider/gcp/utils"
)

type Scanner struct {
	InstancesClient *compute.InstancesClient
	SnapshotsClient *compute.SnapshotsClient
	DisksClient     *compute.DisksClient

	ScannerZone         string
	ProjectID           string
	ScannerSourceImage  string
	ScannerMachineType  string
	ScannerSubnetwork   string
	ScannerSSHPublicKey string
}

// nolint:cyclop
func (s *Scanner) RunAssetScan(ctx context.Context, config *provider.ScanJobConfig) error {
	// convert AssetInfo to vmInfo
	vminfo, err := config.AssetInfo.AsVMInfo()
	if err != nil {
		return provider.FatalErrorf("unable to get vminfo from AssetInfo: %w", err)
	}

	logger := log.GetLoggerFromContextOrDefault(ctx).WithFields(logrus.Fields{
		"AssetScanID":   config.AssetScanID,
		"AssetLocation": vminfo.Location,
		"InstanceID":    vminfo.InstanceID,
		"ScannerZone":   s.ScannerZone,
		"Provider":      string(apitypes.GCP),
	})
	logger.Debugf("Running asset scan")

	targetName := vminfo.InstanceID
	targetZone := vminfo.Location

	// get the target instance to scan from gcp.
	targetVM, err := s.InstancesClient.Get(ctx, &computepb.GetInstanceRequest{
		Instance: targetName,
		Project:  s.ProjectID,
		Zone:     targetZone,
	})
	if err != nil {
		_, err := utils.HandleGcpRequestError(err, "getting target virtual machine %v", targetName)
		return err // nolint: wrapcheck
	}
	logger.Debugf("Got target VM: %v", targetVM.Name)

	// get target instance boot disk
	bootDisk, err := getInstanceBootDisk(targetVM)
	if err != nil {
		return provider.FatalErrorf("unable to get instance boot disk: %w", err)
	}
	logger.Debugf("Got target boot disk: %v", bootDisk.GetSource())

	// ensure that a snapshot was created from the target instance root disk. (create if not)
	snapshot, err := s.ensureSnapshotFromAttachedDisk(ctx, config, bootDisk)
	if err != nil {
		return fmt.Errorf("failed to ensure snapshot for vm root volume: %w", err)
	}
	logger.Debugf("Created snapshot: %v", snapshot.Name)

	// create a disk from the snapshot.
	// Snapshots are global resources, so any snapshot is accessible by any resource within the same project.
	var diskFromSnapshot *computepb.Disk
	diskFromSnapshot, err = s.ensureDiskFromSnapshot(ctx, config, snapshot)
	if err != nil {
		return fmt.Errorf("failed to ensure disk created from snapshot: %w", err)
	}
	logger.Debugf("Created disk from snapshot: %v", diskFromSnapshot.Name)

	// create the scanner instance
	scannerVM, err := s.ensureScannerVirtualMachine(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure scanner virtual machine: %w", err)
	}
	logger.Debugf("Created scanner virtual machine: %v", scannerVM.Name)

	// attach the disk from snapshot to the scanner instance
	err = s.ensureDiskAttachedToScannerVM(ctx, scannerVM, diskFromSnapshot)
	if err != nil {
		return fmt.Errorf("failed to ensure target disk is attached to virtual machine: %w", err)
	}
	logger.Debugf("Attached disk to scanner virtual machine")

	return nil
}

func (s *Scanner) RemoveAssetScan(ctx context.Context, config *provider.ScanJobConfig) error {
	logger := log.GetLoggerFromContextOrDefault(ctx).WithFields(logrus.Fields{
		"AssetScanID": config.AssetScanID,
		"ScannerZone": s.ScannerZone,
		"Provider":    string(apitypes.GCP),
	})

	err := s.ensureScannerVirtualMachineDeleted(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure scanner virtual machine deleted: %w", err)
	}
	logger.Debugf("Deleted scanner virtual machine")

	err = s.ensureTargetDiskDeleted(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure target disk deleted: %w", err)
	}
	logger.Debugf("Deleted disk")

	err = s.ensureSnapshotDeleted(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure snapshot deleted: %w", err)
	}
	logger.Debugf("Deleted snapshot")

	return nil
}

func getInstanceBootDisk(vm *computepb.Instance) (*computepb.AttachedDisk, error) {
	for _, disk := range vm.Disks {
		if disk.Boot != nil && *disk.Boot {
			return disk, nil
		}
	}
	return nil, errors.New("failed to find instance boot disk")
}
