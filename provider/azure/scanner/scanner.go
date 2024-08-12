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
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"

	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/workflow"
	workflowTypes "github.com/openclarity/vmclarity/workflow/types"
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

	RunAssetScanTasks    []*workflowTypes.Task[*AssetScanState]
	RemoveAssetScanTasks []*workflowTypes.Task[*AssetScanState]
}

// nolint:cyclop
func (s *Scanner) RunAssetScan(ctx context.Context, config *provider.ScanJobConfig) error {
	workflow, err := workflow.New[*AssetScanState, *workflowTypes.Task[*AssetScanState]](s.RunAssetScanTasks)
	if err != nil {
		return fmt.Errorf("failed to create RunAssetScan workflow: %w", err)
	}

	err = workflow.Run(ctx, &AssetScanState{config: config, mu: &sync.RWMutex{}})
	if err != nil {
		return fmt.Errorf("failed to run RunAssetScan workflow: %w", err)
	}

	return nil
}

func (s *Scanner) RemoveAssetScan(ctx context.Context, config *provider.ScanJobConfig) error {
	workflow, err := workflow.New[*AssetScanState, *workflowTypes.Task[*AssetScanState]](s.RemoveAssetScanTasks)
	if err != nil {
		return fmt.Errorf("failed to create RemoveAssetScan workflow: %w", err)
	}

	err = workflow.Run(ctx, &AssetScanState{config: config, mu: &sync.RWMutex{}})
	if err != nil {
		return fmt.Errorf("failed to run RemoveAssetScan workflow: %w", err)
	}

	return nil
}
