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
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/openclarity/vmclarity/cmd/vmclarity-cli/logutil"
	cliutils "github.com/openclarity/vmclarity/pkg/cli/utils"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/backendclient"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

// AssetScanCreateCmd represents the standalone command.
var AssetScanCreateCmd = &cobra.Command{
	Use:   "asset-scan-create",
	Short: "Create asset scan",
	Long:  `It creates asset scan. It's useful in the CI/CD mode without VMClarity orchestration`,
	Run: func(cmd *cobra.Command, args []string) {
		logutil.Logger.Infof("asset-scan-create called")
		assetID, err := cmd.Flags().GetString("asset-id")
		if err != nil {
			logutil.Logger.Fatalf("Unable to get asset id: %v", err)
		}
		server, err := cmd.Flags().GetString("server")
		if err != nil {
			logutil.Logger.Fatalf("Unable to get VMClarity server address: %v", err)
		}
		jsonPath, err := cmd.Flags().GetString("jsonpath")
		if err != nil {
			logutil.Logger.Fatalf("Unable to get jsonpath: %v", err)
		}

		assetScan, err := createAssetScan(context.TODO(), server, assetID)
		if err != nil {
			logutil.Logger.Fatalf("Failed to create asset scan: %v", err)
		}

		if err := cliutils.PrintJSONData(assetScan, jsonPath); err != nil {
			logutil.Logger.Fatalf("Failed to print jsonpath: %v", err)
		}
	},
}

func init() {
	AssetScanCreateCmd.Flags().String("server", "", "VMClarity server to create asset to, for example: http://localhost:9999/api")
	AssetScanCreateCmd.Flags().String("asset-id", "", "Asset ID for asset scan")
	AssetScanCreateCmd.Flags().String("jsonpath", "", "print selected value of asset scan")
	if err := AssetScanCreateCmd.MarkFlagRequired("server"); err != nil {
		logutil.Logger.Fatalf("Failed to mark server flag as required: %v", err)
	}
	if err := AssetScanCreateCmd.MarkFlagRequired("asset-id"); err != nil {
		logutil.Logger.Fatalf("Failed to mark asset-id flag as required: %v", err)
	}
}

func createAssetScan(ctx context.Context, server, assetID string) (*models.AssetScan, error) {
	client, err := backendclient.Create(server)
	if err != nil {
		return nil, fmt.Errorf("failed to create VMClarity API client: %w", err)
	}

	asset, err := client.GetAsset(ctx, assetID, models.GetAssetsAssetIDParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to get asset %s: %w", assetID, err)
	}
	assetScanData := createEmptyAssetScanForAsset(asset)

	assetScan, err := client.PostAssetScan(ctx, assetScanData)
	if err != nil {
		var conErr backendclient.AssetScanConflictError
		if errors.As(err, &conErr) {
			assetScanID := *conErr.ConflictingAssetScan.Id
			logutil.Logger.WithField("AssetScanID", assetScanID).Debug("AssetScan already exist.")
			return conErr.ConflictingAssetScan, nil
		}
		return nil, fmt.Errorf("failed to post AssetScan to backend API: %w", err)
	}

	return assetScan, nil
}

func createEmptyAssetScanForAsset(asset models.Asset) models.AssetScan {
	return models.AssetScan{
		Asset: &models.AssetRelationship{
			Id: *asset.Id,
		},
		Status: &models.AssetScanStatus{
			General: &models.AssetScanState{
				State: utils.PointerTo(models.AssetScanStateStateReadyToScan),
			},
		},
		ResourceCleanupStatus: models.NewResourceCleanupStatus(
			models.ResourceCleanupStatusStateSkipped,
			models.ResourceCleanupStatusReasonNotApplicable,
			nil,
		),
	}
}
