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

	compute "cloud.google.com/go/compute/apiv1"

	apitypes "github.com/openclarity/openclarity/api/types"
	"github.com/openclarity/openclarity/provider/gcp/discoverer"
	"github.com/openclarity/openclarity/provider/gcp/estimator"
	"github.com/openclarity/openclarity/provider/gcp/scanner"
)

type Provider struct {
	*discoverer.Discoverer
	*scanner.Scanner
	*estimator.Estimator
}

func (p *Provider) Kind() apitypes.CloudProvider {
	return apitypes.GCP
}

func New(ctx context.Context) (*Provider, error) {
	config, err := NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	err = config.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate configuration: %w", err)
	}

	regionsClient, err := compute.NewRegionsRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create regions client: %w", err)
	}

	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance client: %w", err)
	}

	snapshotsClient, err := compute.NewSnapshotsRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot client: %w", err)
	}

	disksClient, err := compute.NewDisksRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create disks client: %w", err)
	}

	architecture, err := config.ScannerMachineArchitecture.MarshalText()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ScannerInstanceArchitecture into text: %w", err)
	}

	scannerMachineType, ok := config.ScannerMachineArchitectureToTypeMapping[architecture]
	if !ok {
		return nil, fmt.Errorf("failed to find machine type for architecture %s", config.ScannerMachineArchitecture)
	}

	scannerSourceImage, ok := config.ScannerMachineArchitectureToSourceImageMapping[architecture]
	if !ok {
		return nil, fmt.Errorf("failed to find source image for architecture %s", config.ScannerMachineArchitecture)
	}
	scannerSourceImage = config.ScannerSourceImagePrefix + scannerSourceImage

	return &Provider{
		Discoverer: &discoverer.Discoverer{
			DisksClient:     disksClient,
			InstancesClient: instancesClient,
			RegionsClient:   regionsClient,

			ProjectID: config.ProjectID,
		},
		Scanner: &scanner.Scanner{
			InstancesClient: instancesClient,
			SnapshotsClient: snapshotsClient,
			DisksClient:     disksClient,

			ScannerZone:         config.ScannerZone,
			ProjectID:           config.ProjectID,
			ScannerSourceImage:  scannerSourceImage,
			ScannerMachineType:  scannerMachineType,
			ScannerSubnetwork:   config.ScannerSubnetwork,
			ScannerSSHPublicKey: config.ScannerSSHPublicKey,
		},
		Estimator: &estimator.Estimator{},
	}, nil
}
