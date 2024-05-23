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
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/provider/azure/discoverer"
	"github.com/openclarity/vmclarity/provider/azure/estimator"
	"github.com/openclarity/vmclarity/provider/azure/scanner"
)

type Provider struct {
	*discoverer.Discoverer
	*scanner.Scanner
	*estimator.Estimator
}

func (p *Provider) Kind() apitypes.CloudProvider {
	return apitypes.Azure
}

func New(_ context.Context) (*Provider, error) {
	config, err := NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	err = config.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate configuration: %w", err)
	}

	cred, err := azidentity.NewManagedIdentityCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed create managed identity credential: %w", err)
	}

	networkClientFactory, err := armnetwork.NewClientFactory(config.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create network client factory: %w", err)
	}

	computeClientFactory, err := armcompute.NewClientFactory(config.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client factory: %w", err)
	}

	return &Provider{
		Discoverer: &discoverer.Discoverer{
			VMClient:    computeClientFactory.NewVirtualMachinesClient(),
			DisksClient: computeClientFactory.NewDisksClient(),
		},
		Scanner: &scanner.Scanner{
			Cred:             cred,
			VMClient:         computeClientFactory.NewVirtualMachinesClient(),
			SnapshotsClient:  computeClientFactory.NewSnapshotsClient(),
			DisksClient:      computeClientFactory.NewDisksClient(),
			InterfacesClient: networkClientFactory.NewInterfacesClient(),

			SubscriptionID:              config.SubscriptionID,
			ScannerLocation:             config.ScannerLocation,
			ScannerResourceGroup:        config.ScannerResourceGroup,
			ScannerSubnet:               config.ScannerSubnet,
			ScannerPublicKey:            string(config.ScannerPublicKey),
			ScannerVMSize:               config.ScannerVMSize,
			ScannerImagePublisher:       config.ScannerImagePublisher,
			ScannerImageOffer:           config.ScannerImageOffer,
			ScannerImageSKU:             config.ScannerImageSKU,
			ScannerImageVersion:         config.ScannerImageVersion,
			ScannerSecurityGroup:        config.ScannerSecurityGroup,
			ScannerStorageAccountName:   config.ScannerStorageAccountName,
			ScannerStorageContainerName: config.ScannerStorageContainerName,
		},
		Estimator: &estimator.Estimator{},
	}, nil
}
