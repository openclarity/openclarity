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
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"

	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/provider/cloudinit"
	"github.com/openclarity/vmclarity/provider/gcp/utils"
)

var (
	VMCreateEstimateProvisionTime = 2 * time.Minute
	VMDiskAttachEstimateTime      = 2 * time.Minute
	VMDeleteEstimateTime          = 2 * time.Minute
)

const (
	DiskSizeGB = 10
)

func scannerVMNameFromJobConfig(config *provider.ScanJobConfig) string {
	return "vmclarity-scanner-" + config.AssetScanID
}

func (s *Scanner) ensureScannerVirtualMachine(ctx context.Context, config *provider.ScanJobConfig) (*computepb.Instance, error) {
	vmName := scannerVMNameFromJobConfig(config)

	instanceRes, err := s.InstancesClient.Get(ctx, &computepb.GetInstanceRequest{
		Instance: vmName,
		Project:  s.ProjectID,
		Zone:     s.ScannerZone,
	})
	if err == nil {
		if *instanceRes.Status != InstanceStateRunning {
			return instanceRes, provider.RetryableErrorf(VMCreateEstimateProvisionTime, "virtual machine is not ready yet. status: %v", *instanceRes.Status)
		}

		// Everything is good, the instance exists and running.
		return instanceRes, nil
	}

	notFound, err := utils.HandleGcpRequestError(err, "getting scanner virtual machine: %v", vmName)
	// ignore not found error as it is expected
	if !notFound {
		return nil, err // nolint: wrapcheck
	}

	// create the instance if not exists
	userData, err := cloudinit.New(config)
	if err != nil {
		return nil, provider.FatalErrorf("failed to generate cloud-init: %v", err)
	}

	zone := s.ScannerZone
	instanceName := vmName

	req := &computepb.InsertInstanceRequest{
		Project: s.ProjectID,
		Zone:    zone,
		InstanceResource: &computepb.Instance{
			Metadata: &computepb.Metadata{
				Items: []*computepb.Items{
					{
						Key:   to.Ptr("user-data"),
						Value: to.Ptr(userData),
					},
				},
			},
			Description: to.Ptr("VMClarity scanner"),
			Name:        &instanceName,
			Disks: []*computepb.AttachedDisk{
				{
					InitializeParams: &computepb.AttachedDiskInitializeParams{
						DiskType:    to.Ptr(fmt.Sprintf("zones/%s/diskTypes/pd-balanced", zone)),
						DiskSizeGb:  to.Ptr[int64](DiskSizeGB),
						SourceImage: &s.ScannerSourceImage,
					},
					AutoDelete: to.Ptr(true),
					Boot:       to.Ptr(true),
					Type:       to.Ptr(computepb.AttachedDisk_PERSISTENT.String()),
				},
			},
			MachineType: to.Ptr(fmt.Sprintf("zones/%s/machineTypes/%s", zone, s.ScannerMachineType)),
			NetworkInterfaces: []*computepb.NetworkInterface{
				{
					Subnetwork: &s.ScannerSubnetwork,
				},
			},
		},
	}

	if s.ScannerSSHPublicKey != "" {
		req.InstanceResource.Metadata.Items = append(
			req.InstanceResource.Metadata.Items,
			&computepb.Items{
				Key:   to.Ptr("ssh-keys"),
				Value: to.Ptr("vmclarity:" + s.ScannerSSHPublicKey),
			},
		)
	}

	_, err = s.InstancesClient.Insert(ctx, req)
	if err != nil {
		_, err := utils.HandleGcpRequestError(err, "unable to create instance %v", vmName)
		return nil, err // nolint: wrapcheck
	}

	return nil, provider.RetryableErrorf(VMCreateEstimateProvisionTime, "vm creating")
}

func (s *Scanner) ensureScannerVirtualMachineDeleted(ctx context.Context, config *provider.ScanJobConfig) error {
	vmName := scannerVMNameFromJobConfig(config)

	return utils.EnsureDeleted( // nolint: wrapcheck
		"VirtualMachine",
		func() error {
			_, err := s.InstancesClient.Get(ctx, &computepb.GetInstanceRequest{
				Instance: vmName,
				Project:  s.ProjectID,
				Zone:     s.ScannerZone,
			})
			return err // nolint: wrapcheck
		},
		func() error {
			_, err := s.InstancesClient.Delete(ctx, &computepb.DeleteInstanceRequest{
				Instance: vmName,
				Project:  s.ProjectID,
				Zone:     s.ScannerZone,
			})
			return err // nolint: wrapcheck
		},
		VMDeleteEstimateTime,
	)
}

func (s *Scanner) ensureDiskAttachedToScannerVM(ctx context.Context, vm *computepb.Instance, disk *computepb.Disk) error {
	var diskAttached bool
	for _, attachedDisk := range vm.Disks {
		diskName := utils.GetLastURLPart(attachedDisk.Source)
		if diskName == *disk.Name {
			diskAttached = true
			break
		}
	}

	if !diskAttached {
		req := &computepb.AttachDiskInstanceRequest{
			AttachedDiskResource: &computepb.AttachedDisk{Source: to.Ptr(disk.GetSelfLink())},
			Instance:             *vm.Name,
			Project:              s.ProjectID,
			Zone:                 s.ScannerZone,
		}

		_, err := s.InstancesClient.AttachDisk(ctx, req)
		if err != nil {
			_, err = utils.HandleGcpRequestError(err, "attach disk %v to VM %v", *disk.Name, *vm.Name)
			return err // nolint: wrapcheck
		}
	}

	diskResp, err := s.DisksClient.Get(ctx, &computepb.GetDiskRequest{
		Disk:    *disk.Name,
		Project: s.ProjectID,
		Zone:    s.ScannerZone,
	})
	if err != nil {
		_, err = utils.HandleGcpRequestError(err, "get disk %v", *disk.Name)
		return err // nolint: wrapcheck
	}

	if *diskResp.Status != ProvisioningStateReady {
		return provider.RetryableErrorf(VMDiskAttachEstimateTime, "disk is not yet attached, status: %v", *disk.Status)
	}

	return nil
}
