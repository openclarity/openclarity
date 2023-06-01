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
	ScanResultsTable() ScanResultsTable
	ScanConfigsTable() ScanConfigsTable
	ScansTable() ScansTable
	TargetsTable() TargetsTable
	ScopesTable() ScopesTable
	FindingsTable() FindingsTable
}

type ScansTable interface {
	GetScans(params models.GetScansParams) (models.Scans, error)
	GetScan(scanID models.ScanID, params models.GetScansScanIDParams) (models.Scan, error)

	CreateScan(scan models.Scan) (models.Scan, error)
	UpdateScan(scan models.Scan, params models.PatchScansScanIDParams) (models.Scan, error)
	SaveScan(scan models.Scan, params models.PutScansScanIDParams) (models.Scan, error)

	DeleteScan(scanID models.ScanID) error
}

type ScanResultsTable interface {
	GetScanResults(params models.GetScanResultsParams) (models.TargetScanResults, error)
	GetScanResult(scanResultID models.ScanResultID, params models.GetScanResultsScanResultIDParams) (models.TargetScanResult, error)

	CreateScanResult(scanResults models.TargetScanResult) (models.TargetScanResult, error)
	UpdateScanResult(scanResults models.TargetScanResult, params models.PatchScanResultsScanResultIDParams) (models.TargetScanResult, error)
	SaveScanResult(scanResults models.TargetScanResult, params models.PutScanResultsScanResultIDParams) (models.TargetScanResult, error)

	// DeleteScanResult(scanResultID models.ScanResultID) error
}

type ScanConfigsTable interface {
	GetScanConfigs(params models.GetScanConfigsParams) (models.ScanConfigs, error)
	GetScanConfig(scanConfigID models.ScanConfigID, params models.GetScanConfigsScanConfigIDParams) (models.ScanConfig, error)

	CreateScanConfig(scanConfig models.ScanConfig) (models.ScanConfig, error)
	UpdateScanConfig(scanConfig models.ScanConfig, params models.PatchScanConfigsScanConfigIDParams) (models.ScanConfig, error)
	SaveScanConfig(scanConfig models.ScanConfig, params models.PutScanConfigsScanConfigIDParams) (models.ScanConfig, error)

	DeleteScanConfig(scanConfigID models.ScanConfigID) error
}

type TargetsTable interface {
	GetTargets(params models.GetTargetsParams) (models.Targets, error)
	GetTarget(targetID models.TargetID, params models.GetTargetsTargetIDParams) (models.Target, error)

	CreateTarget(target models.Target) (models.Target, error)
	UpdateTarget(target models.Target, params models.PatchTargetsTargetIDParams) (models.Target, error)
	SaveTarget(target models.Target, params models.PutTargetsTargetIDParams) (models.Target, error)

	DeleteTarget(targetID models.TargetID) error
}

type ScopesTable interface {
	GetScopes(params models.GetDiscoveryScopesParams) (models.Scopes, error)
	SetScopes(scopes models.Scopes) (models.Scopes, error)
}

type FindingsTable interface {
	GetFindings(params models.GetFindingsParams) (models.Findings, error)
	GetFinding(findingID models.FindingID, params models.GetFindingsFindingIDParams) (models.Finding, error)

	CreateFinding(finding models.Finding) (models.Finding, error)
	UpdateFinding(finding models.Finding) (models.Finding, error)
	SaveFinding(finding models.Finding) (models.Finding, error)

	DeleteFinding(findingID models.FindingID) error
}
