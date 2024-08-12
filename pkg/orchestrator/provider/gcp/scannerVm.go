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

package gcp

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"

	"github.com/openclarity/vmclarity/pkg/orchestrator/provider"
	"github.com/openclarity/vmclarity/pkg/orchestrator/provider/cloudinit"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
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
	return fmt.Sprintf("vmclarity-scanner-%s", config.AssetScanID)
}

func (p *Provider) ensureScannerVirtualMachine(ctx context.Context, config *provider.ScanJobConfig) (*computepb.Instance, error) {
	vmName := scannerVMNameFromJobConfig(config)

	instanceRes, err := p.instancesClient.Get(ctx, &computepb.GetInstanceRequest{
		Instance: vmName,
		Project:  p.config.ProjectID,
		Zone:     p.config.ScannerZone,
	})
	if err == nil {
		if *instanceRes.Status != InstanceStateRunning {
			return instanceRes, provider.RetryableErrorf(VMCreateEstimateProvisionTime, "virtual machine is not ready yet. status: %v", *instanceRes.Status)
		}

		// Everything is good, the instance exists and running.
		return instanceRes, nil
	}

	notFound, err := handleGcpRequestError(err, "getting scanner virtual machine: %v", vmName)
	// ignore not found error as it is expected
	if !notFound {
		return nil, err
	}

	// create the instance if not exists
	userData, err := cloudinit.New(config)
	if err != nil {
		return nil, provider.FatalErrorf("failed to generate cloud-init: %v", err)
	}

	zone := p.config.ScannerZone
	instanceName := vmName

	req := &computepb.InsertInstanceRequest{
		Project: p.config.ProjectID,
		Zone:    zone,
		InstanceResource: &computepb.Instance{
			Metadata: &computepb.Metadata{
				Items: []*computepb.Items{
					{
						Key:   utils.PointerTo("user-data"),
						Value: utils.PointerTo(userData),
					},
				},
			},
			Description: utils.PointerTo("VMClarity scanner"),
			Name:        &instanceName,
			Disks: []*computepb.AttachedDisk{
				{
					InitializeParams: &computepb.AttachedDiskInitializeParams{
						DiskType:    utils.PointerTo(fmt.Sprintf("zones/%s/diskTypes/pd-balanced", zone)),
						DiskSizeGb:  utils.PointerTo[int64](DiskSizeGB),
						SourceImage: &p.config.ScannerSourceImage,
					},
					AutoDelete: utils.PointerTo(true),
					Boot:       utils.PointerTo(true),
					Type:       utils.PointerTo(computepb.AttachedDisk_PERSISTENT.String()),
				},
			},
			MachineType: utils.PointerTo(fmt.Sprintf("zones/%s/machineTypes/%s", zone, p.config.ScannerMachineType)),
			NetworkInterfaces: []*computepb.NetworkInterface{
				{
					Subnetwork: &p.config.ScannerSubnetwork,
				},
			},
		},
	}

	if p.config.ScannerSSHPublicKey != "" {
		req.InstanceResource.Metadata.Items = append(
			req.InstanceResource.Metadata.Items,
			&computepb.Items{
				Key:   utils.PointerTo("ssh-keys"),
				Value: utils.PointerTo(fmt.Sprintf("vmclarity:%s", p.config.ScannerSSHPublicKey)),
			},
		)
	}

	_, err = p.instancesClient.Insert(ctx, req)
	if err != nil {
		_, err := handleGcpRequestError(err, "unable to create instance %v", vmName)
		return nil, err
	}

	return nil, provider.RetryableErrorf(VMCreateEstimateProvisionTime, "vm creating")
}

func (p *Provider) ensureScannerVirtualMachineDeleted(ctx context.Context, config *provider.ScanJobConfig) error {
	vmName := scannerVMNameFromJobConfig(config)

	return ensureDeleted(
		"VirtualMachine",
		func() error {
			_, err := p.instancesClient.Get(ctx, &computepb.GetInstanceRequest{
				Instance: vmName,
				Project:  p.config.ProjectID,
				Zone:     p.config.ScannerZone,
			})
			return err // nolint: wrapcheck
		},
		func() error {
			_, err := p.instancesClient.Delete(ctx, &computepb.DeleteInstanceRequest{
				Instance: vmName,
				Project:  p.config.ProjectID,
				Zone:     p.config.ScannerZone,
			})
			return err // nolint: wrapcheck
		},
		VMDeleteEstimateTime,
	)
}

func (p *Provider) ensureDiskAttachedToScannerVM(ctx context.Context, vm *computepb.Instance, disk *computepb.Disk) error {
	var diskAttached bool
	for _, attachedDisk := range vm.Disks {
		diskName := getLastURLPart(attachedDisk.Source)
		if diskName == *disk.Name {
			diskAttached = true
			break
		}
	}

	if !diskAttached {
		req := &computepb.AttachDiskInstanceRequest{
			AttachedDiskResource: &computepb.AttachedDisk{Source: utils.PointerTo(disk.GetSelfLink())},
			Instance:             *vm.Name,
			Project:              p.config.ProjectID,
			Zone:                 p.config.ScannerZone,
		}

		_, err := p.instancesClient.AttachDisk(ctx, req)
		if err != nil {
			_, err = handleGcpRequestError(err, "attach disk %v to VM %v", *disk.Name, *vm.Name)
			return err
		}
	}

	diskResp, err := p.disksClient.Get(ctx, &computepb.GetDiskRequest{
		Disk:    *disk.Name,
		Project: p.config.ProjectID,
		Zone:    p.config.ScannerZone,
	})
	if err != nil {
		_, err = handleGcpRequestError(err, "get disk %v", *disk.Name)
		return err
	}

	if *diskResp.Status != ProvisioningStateReady {
		return provider.RetryableErrorf(VMDiskAttachEstimateTime, "disk is not yet attached, status: %v", *disk.Status)
	}

	return nil
}
