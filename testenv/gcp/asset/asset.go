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

package asset

import (
	"context"
	"fmt"

	"github.com/openclarity/vmclarity/core/to"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
)

const (
	MachineType      = "zones/us-central1-f/machineTypes/e2-micro"
	SourceImage      = "projects/debian-cloud/global/images/family/debian-12"
	NetworkInterface = "global/networks/default"
)

func Create(ctx context.Context, instancesClient *compute.InstancesClient, projectID string, zone string, assetName string) error {
	var diskSizeGb int64 = 10

	req := &computepb.InsertInstanceRequest{
		Project: projectID,
		Zone:    zone,
		InstanceResource: &computepb.Instance{
			Name:        to.Ptr(assetName),
			MachineType: to.Ptr(MachineType),
			Disks: []*computepb.AttachedDisk{
				{
					InitializeParams: &computepb.AttachedDiskInitializeParams{
						DiskSizeGb:  to.Ptr(diskSizeGb),
						SourceImage: to.Ptr(SourceImage),
					},
					AutoDelete: to.Ptr(true),
					Boot:       to.Ptr(true),
				},
			},
			NetworkInterfaces: []*computepb.NetworkInterface{
				{
					Name: to.Ptr(NetworkInterface),
				},
			},
			Labels: map[string]string{
				"scanconfig": "test",
			},
		},
	}

	op, err := instancesClient.Insert(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create test instance: %w", err)
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("failed to wait for instance to be created: %w", err)
	}

	return nil
}

func Delete(ctx context.Context, instancesClient *compute.InstancesClient, projectID string, zone string, assetName string) error {
	req := &computepb.DeleteInstanceRequest{
		Project:  projectID,
		Zone:     zone,
		Instance: assetName,
	}

	op, err := instancesClient.Delete(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete test instance: %w", err)
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("failed to wait for instance to be deleted: %w", err)
	}

	return nil
}
