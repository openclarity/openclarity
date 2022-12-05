// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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

package database

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/openclarity/vmclarity/api/models"
)

// TODO after db design.
type ScanResults struct {
	ID               string
	Sbom             *SbomScanResults
	Vulnerability    *VulnerabilityScanResults
	Malware          *MalwareScanResults
	Rootkit          *RootkitScanScanResults
	Secret           *SecretScanResults
	Misconfiguration *MisconfigurationScanResults
	Exploit          *ExploitScanResults
}

// TODO after db design.
type SbomScanResults struct {
	Results models.SbomScan
}

type VulnerabilityScanResults struct {
	Results models.VulnerabilityScan
}

type MalwareScanResults struct {
	Results models.MalwareScan
}

type RootkitScanScanResults struct {
	Results models.RootkitScan
}

type SecretScanResults struct {
	Results models.SecretScan
}

type MisconfigurationScanResults struct {
	Results models.MisconfigurationScan
}

type ExploitScanResults struct {
	Results models.ExploitScan
}

//nolint:interfacebloat
type ScanResultsTable interface {
	ListScanResults(targetID models.TargetID, params models.GetTargetsTargetIDScanResultsParams) ([]ScanResults, error)
	CreateScanResults(targetID models.TargetID, scanResults *ScanResults) (*ScanResults, error)
	GetScanResults(targetID models.TargetID, scanID models.ScanID) (*ScanResults, error)
	GetSBOM(targetID models.TargetID, scanID models.ScanID) (*SbomScanResults, error)
	GetVulnerabilities(targetID models.TargetID, scanID models.ScanID) (*VulnerabilityScanResults, error)
	GetMalwares(targetID models.TargetID, scanID models.ScanID) (*MalwareScanResults, error)
	GetRootkits(targetID models.TargetID, scanID models.ScanID) (*RootkitScanScanResults, error)
	GetSecrets(targetID models.TargetID, scanID models.ScanID) (*SecretScanResults, error)
	GetMisconfigurations(targetID models.TargetID, scanID models.ScanID) (*MisconfigurationScanResults, error)
	GetExploits(targetID models.TargetID, scanID models.ScanID) (*ExploitScanResults, error)
	UpdateScanResults(targetID models.TargetID, scanID models.ScanID, scanResults *ScanResults) (*ScanResults, error)
}

type ScanResultsTableHandler struct {
	db *gorm.DB
}

func (db *Handler) ScanResultsTable() ScanResultsTable {
	return &ScanResultsTableHandler{
		db: db.DB,
	}
}

func (s *ScanResultsTableHandler) ListScanResults(targetID models.TargetID, params models.GetTargetsTargetIDScanResultsParams,
) ([]ScanResults, error) {
	return []ScanResults{}, fmt.Errorf("not implemented")
}

func (s *ScanResultsTableHandler) CreateScanResults(targetID models.TargetID, scanResults *ScanResults,
) (*ScanResults, error) {
	return &ScanResults{}, fmt.Errorf("not implemented")
}

func (s *ScanResultsTableHandler) GetScanResults(targetID models.TargetID, scanID models.ScanID) (*ScanResults, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *ScanResultsTableHandler) GetSBOM(targetID models.TargetID, scanID models.ScanID) (*SbomScanResults, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *ScanResultsTableHandler) GetVulnerabilities(targetID models.TargetID, scanID models.ScanID) (*VulnerabilityScanResults, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *ScanResultsTableHandler) GetMalwares(targetID models.TargetID, scanID models.ScanID) (*MalwareScanResults, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *ScanResultsTableHandler) GetRootkits(targetID models.TargetID, scanID models.ScanID) (*RootkitScanScanResults, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *ScanResultsTableHandler) GetSecrets(targetID models.TargetID, scanID models.ScanID) (*SecretScanResults, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *ScanResultsTableHandler) GetMisconfigurations(targetID models.TargetID, scanID models.ScanID) (*MisconfigurationScanResults, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *ScanResultsTableHandler) GetExploits(targetID models.TargetID, scanID models.ScanID) (*ExploitScanResults, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *ScanResultsTableHandler) UpdateScanResults(
	targetID models.TargetID,
	scanID models.ScanID,
	scanResults *ScanResults,
) (*ScanResults, error) {
	return &ScanResults{}, fmt.Errorf("not implemented")
}
