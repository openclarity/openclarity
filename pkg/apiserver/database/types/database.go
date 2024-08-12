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

package types

import (
	"errors"
	"fmt"

	"github.com/openclarity/vmclarity/api/models"
)

const (
	DBDriverTypeLocal    = "LOCAL"
	DBDriverTypePostgres = "POSTGRES"
)

var ErrNotFound = errors.New("not found")

type PreconditionFailedError struct {
	Reason string
}

func (e *PreconditionFailedError) Error() string {
	return fmt.Sprintf("Precondition failed: %s", e.Reason)
}

type DBConfig struct {
	EnableInfoLogs bool   `json:"enable-info-logs"`
	DriverType     string `json:"driver-type,omitempty"`
	DBPassword     string `json:"-"`
	DBUser         string `json:"db-user,omitempty"`
	DBHost         string `json:"db-host,omitempty"`
	DBPort         string `json:"db-port,omitempty"`
	DBName         string `json:"db-name,omitempty"`

	LocalDBPath string `json:"local-db-path,omitempty"`
}

type Database interface {
	AssetScansTable() AssetScansTable
	ScanConfigsTable() ScanConfigsTable
	ScansTable() ScansTable
	AssetsTable() AssetsTable
	FindingsTable() FindingsTable
	ScanEstimationsTable() ScanEstimationsTable
	AssetScanEstimationsTable() AssetScanEstimationsTable
	ProvidersTable() ProvidersTable
}

type ScansTable interface {
	GetScans(params models.GetScansParams) (models.Scans, error)
	GetScan(scanID models.ScanID, params models.GetScansScanIDParams) (models.Scan, error)

	CreateScan(scan models.Scan) (models.Scan, error)
	UpdateScan(scan models.Scan, params models.PatchScansScanIDParams) (models.Scan, error)
	SaveScan(scan models.Scan, params models.PutScansScanIDParams) (models.Scan, error)

	DeleteScan(scanID models.ScanID) error
}

type AssetScansTable interface {
	GetAssetScans(params models.GetAssetScansParams) (models.AssetScans, error)
	GetAssetScan(assetScanID models.AssetScanID, params models.GetAssetScansAssetScanIDParams) (models.AssetScan, error)

	CreateAssetScan(assetScans models.AssetScan) (models.AssetScan, error)
	UpdateAssetScan(assetScans models.AssetScan, params models.PatchAssetScansAssetScanIDParams) (models.AssetScan, error)
	SaveAssetScan(assetScans models.AssetScan, params models.PutAssetScansAssetScanIDParams) (models.AssetScan, error)

	// DeleteAssetScan(assetScanID models.AssetScanID) error
}

type ScanConfigsTable interface {
	GetScanConfigs(params models.GetScanConfigsParams) (models.ScanConfigs, error)
	GetScanConfig(scanConfigID models.ScanConfigID, params models.GetScanConfigsScanConfigIDParams) (models.ScanConfig, error)

	CreateScanConfig(scanConfig models.ScanConfig) (models.ScanConfig, error)
	UpdateScanConfig(scanConfig models.ScanConfig, params models.PatchScanConfigsScanConfigIDParams) (models.ScanConfig, error)
	SaveScanConfig(scanConfig models.ScanConfig, params models.PutScanConfigsScanConfigIDParams) (models.ScanConfig, error)

	DeleteScanConfig(scanConfigID models.ScanConfigID) error
}

type AssetsTable interface {
	GetAssets(params models.GetAssetsParams) (models.Assets, error)
	GetAsset(assetID models.AssetID, params models.GetAssetsAssetIDParams) (models.Asset, error)

	CreateAsset(asset models.Asset) (models.Asset, error)
	UpdateAsset(asset models.Asset, params models.PatchAssetsAssetIDParams) (models.Asset, error)
	SaveAsset(asset models.Asset, params models.PutAssetsAssetIDParams) (models.Asset, error)

	DeleteAsset(assetID models.AssetID) error
}

type FindingsTable interface {
	GetFindings(params models.GetFindingsParams) (models.Findings, error)
	GetFinding(findingID models.FindingID, params models.GetFindingsFindingIDParams) (models.Finding, error)

	CreateFinding(finding models.Finding) (models.Finding, error)
	UpdateFinding(finding models.Finding) (models.Finding, error)
	SaveFinding(finding models.Finding) (models.Finding, error)

	DeleteFinding(findingID models.FindingID) error
}

type ScanEstimationsTable interface {
	GetScanEstimations(params models.GetScanEstimationsParams) (models.ScanEstimations, error)
	GetScanEstimation(scanEstimationID models.ScanEstimationID, params models.GetScanEstimationsScanEstimationIDParams) (models.ScanEstimation, error)

	CreateScanEstimation(scanEstimation models.ScanEstimation) (models.ScanEstimation, error)
	UpdateScanEstimation(scanEstimation models.ScanEstimation, params models.PatchScanEstimationsScanEstimationIDParams) (models.ScanEstimation, error)
	SaveScanEstimation(scanEstimation models.ScanEstimation, params models.PutScanEstimationsScanEstimationIDParams) (models.ScanEstimation, error)

	DeleteScanEstimation(scanEstimationID models.ScanEstimationID) error
}

type AssetScanEstimationsTable interface {
	GetAssetScanEstimations(params models.GetAssetScanEstimationsParams) (models.AssetScanEstimations, error)
	GetAssetScanEstimation(assetScanEstimationID models.AssetScanEstimationID, params models.GetAssetScanEstimationsAssetScanEstimationIDParams) (models.AssetScanEstimation, error)

	CreateAssetScanEstimation(assetScanEstimations models.AssetScanEstimation) (models.AssetScanEstimation, error)
	UpdateAssetScanEstimation(assetScanEstimations models.AssetScanEstimation, params models.PatchAssetScanEstimationsAssetScanEstimationIDParams) (models.AssetScanEstimation, error)
	SaveAssetScanEstimation(assetScanEstimations models.AssetScanEstimation, params models.PutAssetScanEstimationsAssetScanEstimationIDParams) (models.AssetScanEstimation, error)

	DeleteAssetScanEstimation(assetScanEstimationID models.AssetScanEstimationID) error
}

type ProvidersTable interface {
	GetProviders(params models.GetProvidersParams) (models.Providers, error)
	GetProvider(providerID models.ProviderID, params models.GetProvidersProviderIDParams) (models.Provider, error)

	CreateProvider(provider models.Provider) (models.Provider, error)
	UpdateProvider(provider models.Provider, params models.PatchProvidersProviderIDParams) (models.Provider, error)
	SaveProvider(provider models.Provider, params models.PutProvidersProviderIDParams) (models.Provider, error)

	DeleteProvider(providerID models.ProviderID) error
}
