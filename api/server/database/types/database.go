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

	"github.com/openclarity/vmclarity/api/types"
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
	return "Precondition failed: " + e.Reason
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
	AssetFindingsTable() AssetFindingsTable
}

type ScansTable interface {
	GetScans(params types.GetScansParams) (types.Scans, error)
	GetScan(scanID types.ScanID, params types.GetScansScanIDParams) (types.Scan, error)

	CreateScan(scan types.Scan) (types.Scan, error)
	UpdateScan(scan types.Scan, params types.PatchScansScanIDParams) (types.Scan, error)
	SaveScan(scan types.Scan, params types.PutScansScanIDParams) (types.Scan, error)

	DeleteScan(scanID types.ScanID) error
}

type AssetScansTable interface {
	GetAssetScans(params types.GetAssetScansParams) (types.AssetScans, error)
	GetAssetScan(assetScanID types.AssetScanID, params types.GetAssetScansAssetScanIDParams) (types.AssetScan, error)

	CreateAssetScan(assetScans types.AssetScan) (types.AssetScan, error)
	UpdateAssetScan(assetScans types.AssetScan, params types.PatchAssetScansAssetScanIDParams) (types.AssetScan, error)
	SaveAssetScan(assetScans types.AssetScan, params types.PutAssetScansAssetScanIDParams) (types.AssetScan, error)

	// DeleteAssetScan(assetScanID types.AssetScanID) error
}

type ScanConfigsTable interface {
	GetScanConfigs(params types.GetScanConfigsParams) (types.ScanConfigs, error)
	GetScanConfig(scanConfigID types.ScanConfigID, params types.GetScanConfigsScanConfigIDParams) (types.ScanConfig, error)

	CreateScanConfig(scanConfig types.ScanConfig) (types.ScanConfig, error)
	UpdateScanConfig(scanConfig types.ScanConfig, params types.PatchScanConfigsScanConfigIDParams) (types.ScanConfig, error)
	SaveScanConfig(scanConfig types.ScanConfig, params types.PutScanConfigsScanConfigIDParams) (types.ScanConfig, error)

	DeleteScanConfig(scanConfigID types.ScanConfigID) error
}

type AssetsTable interface {
	GetAssets(params types.GetAssetsParams) (types.Assets, error)
	GetAsset(assetID types.AssetID, params types.GetAssetsAssetIDParams) (types.Asset, error)

	CreateAsset(asset types.Asset) (types.Asset, error)
	UpdateAsset(asset types.Asset, params types.PatchAssetsAssetIDParams) (types.Asset, error)
	SaveAsset(asset types.Asset, params types.PutAssetsAssetIDParams) (types.Asset, error)

	DeleteAsset(assetID types.AssetID) error
}

type FindingsTable interface {
	GetFindings(params types.GetFindingsParams) (types.Findings, error)
	GetFinding(findingID types.FindingID, params types.GetFindingsFindingIDParams) (types.Finding, error)

	CreateFinding(finding types.Finding) (types.Finding, error)
	UpdateFinding(finding types.Finding) (types.Finding, error)
	SaveFinding(finding types.Finding) (types.Finding, error)

	DeleteFinding(findingID types.FindingID) error
}

type ScanEstimationsTable interface {
	GetScanEstimations(params types.GetScanEstimationsParams) (types.ScanEstimations, error)
	GetScanEstimation(scanEstimationID types.ScanEstimationID, params types.GetScanEstimationsScanEstimationIDParams) (types.ScanEstimation, error)

	CreateScanEstimation(scanEstimation types.ScanEstimation) (types.ScanEstimation, error)
	UpdateScanEstimation(scanEstimation types.ScanEstimation, params types.PatchScanEstimationsScanEstimationIDParams) (types.ScanEstimation, error)
	SaveScanEstimation(scanEstimation types.ScanEstimation, params types.PutScanEstimationsScanEstimationIDParams) (types.ScanEstimation, error)

	DeleteScanEstimation(scanEstimationID types.ScanEstimationID) error
}

type AssetScanEstimationsTable interface {
	GetAssetScanEstimations(params types.GetAssetScanEstimationsParams) (types.AssetScanEstimations, error)
	GetAssetScanEstimation(assetScanEstimationID types.AssetScanEstimationID, params types.GetAssetScanEstimationsAssetScanEstimationIDParams) (types.AssetScanEstimation, error)

	CreateAssetScanEstimation(assetScanEstimations types.AssetScanEstimation) (types.AssetScanEstimation, error)
	UpdateAssetScanEstimation(assetScanEstimations types.AssetScanEstimation, params types.PatchAssetScanEstimationsAssetScanEstimationIDParams) (types.AssetScanEstimation, error)
	SaveAssetScanEstimation(assetScanEstimations types.AssetScanEstimation, params types.PutAssetScanEstimationsAssetScanEstimationIDParams) (types.AssetScanEstimation, error)

	DeleteAssetScanEstimation(assetScanEstimationID types.AssetScanEstimationID) error
}

type ProvidersTable interface {
	GetProviders(params types.GetProvidersParams) (types.Providers, error)
	GetProvider(providerID types.ProviderID, params types.GetProvidersProviderIDParams) (types.Provider, error)

	CreateProvider(provider types.Provider) (types.Provider, error)
	UpdateProvider(provider types.Provider, params types.PatchProvidersProviderIDParams) (types.Provider, error)
	SaveProvider(provider types.Provider, params types.PutProvidersProviderIDParams) (types.Provider, error)

	DeleteProvider(providerID types.ProviderID) error
}

type AssetFindingsTable interface {
	GetAssetFindings(params types.GetAssetFindingsParams) (types.AssetFindings, error)
	GetAssetFinding(assetFindingID types.AssetFindingID, params types.GetAssetFindingsAssetFindingIDParams) (types.AssetFinding, error)

	CreateAssetFinding(assetFinding types.AssetFinding) (types.AssetFinding, error)
	UpdateAssetFinding(assetFinding types.AssetFinding, params types.PatchAssetFindingsAssetFindingIDParams) (types.AssetFinding, error)
	SaveAssetFinding(assetFinding types.AssetFinding, params types.PutAssetFindingsAssetFindingIDParams) (types.AssetFinding, error)

	DeleteAssetFinding(assetFindingID types.AssetFindingID) error
}
