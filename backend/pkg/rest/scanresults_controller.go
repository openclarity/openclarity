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

package rest

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/backend/pkg/database"
)

func (s *ServerImpl) GetTargetsTargetIDScanResults(
	ctx echo.Context,
	targetID models.TargetID,
	params models.GetTargetsTargetIDScanResultsParams,
) error {
	results, err := s.dbHandler.ScanResultsTable().ListScanResults(targetID, params)
	if err != nil {
		// TODO check errors for status code
		log.Errorf("%v", err)
		return sendError(ctx, http.StatusInternalServerError, oops)
	}
	resultsModel := []models.ScanResults{}
	for _, result := range results {
		result := result
		resultModel := createModelScanResultsFromDB(&result)
		resultsModel = append(resultsModel, *resultModel)
	}
	return sendResponse(ctx, http.StatusOK, &resultsModel)
}

func (s *ServerImpl) PostTargetsTargetIDScanResults(
	ctx echo.Context,
	targetID models.TargetID,
) error {
	var scanResults models.ScanResults
	err := ctx.Bind(&scanResults)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	newScanResults := createDBScanResultsFromModel(&scanResults)
	results, err := s.dbHandler.ScanResultsTable().CreateScanResults(targetID, newScanResults)
	if err != nil {
		// TODO check errors for status code
		return sendError(ctx, http.StatusInternalServerError, err.Error())
	}
	return sendResponse(ctx, http.StatusCreated, createModelScanResultsFromDB(results))
}

//nolint:cyclop
func (s *ServerImpl) GetTargetsTargetIDScanResultsScanID(
	ctx echo.Context,
	targetID models.TargetID,
	scanID models.ScanID,
) error {
	result, err := s.dbHandler.ScanResultsTable().GetScanResults(targetID, scanID)
	if err != nil {
		// TODO check errors for status code
		log.Errorf("%v", err)
		return sendError(ctx, http.StatusNotFound, oops)
	}
	return sendResponse(ctx, http.StatusOK, createModelScanResultsFromDB(result))
}

func (s *ServerImpl) PutTargetsTargetIDScanResultsScanID(
	ctx echo.Context,
	targetID models.TargetID,
	scanID models.ScanID,
) error {
	var scanResults models.ScanResults
	err := ctx.Bind(&scanResults)
	if err != nil {
		log.Errorf("%v", err)
		return sendError(ctx, http.StatusBadRequest, oops)
	}

	newScanResults := createDBScanResultsFromModel(&scanResults)
	results, err := s.dbHandler.ScanResultsTable().UpdateScanResults(targetID, scanID, newScanResults)
	if err != nil {
		// TODO check errors for status code
		log.Errorf("%v", err)
		return sendError(ctx, http.StatusInternalServerError, oops)
	}
	return sendResponse(ctx, http.StatusOK, createModelScanResultsFromDB(results))
}

// TODO after db design.
func createDBScanResultsFromModel(scanResults *models.ScanResults) *database.ScanResults {
	var scanResultID string
	if scanResults.Id == nil || *scanResults.Id == "" {
		scanResultID = generateScanResultsID()
	} else {
		scanResultID = *scanResults.Id
	}
	var sbomRes *database.SbomScanResults
	if scanResults.Sboms != nil {
		sbomRes = &database.SbomScanResults{
			Results: *scanResults.Sboms,
		}
	}
	var vulRs *database.VulnerabilityScanResults
	if scanResults.Vulnerabilities != nil {
		vulRs = &database.VulnerabilityScanResults{
			Results: *scanResults.Vulnerabilities,
		}
	}
	var malwareRes *database.MalwareScanResults
	if scanResults.Malwares != nil {
		malwareRes = &database.MalwareScanResults{
			Results: *scanResults.Malwares,
		}
	}
	var secretRes *database.SecretScanResults
	if scanResults.Secrets != nil {
		secretRes = &database.SecretScanResults{
			Results: *scanResults.Secrets,
		}
	}
	var rootkitRes *database.RootkitScanScanResults
	if scanResults.Rootkits != nil {
		rootkitRes = &database.RootkitScanScanResults{
			Results: *scanResults.Rootkits,
		}
	}
	var misconfigRes *database.MisconfigurationScanResults
	if scanResults.Misconfigurations != nil {
		misconfigRes = &database.MisconfigurationScanResults{
			Results: *scanResults.Misconfigurations,
		}
	}
	var exploitRes *database.ExploitScanResults
	if scanResults.Exploits != nil {
		exploitRes = &database.ExploitScanResults{
			Results: *scanResults.Exploits,
		}
	}
	return &database.ScanResults{
		ID:               scanResultID,
		Sbom:             sbomRes,
		Vulnerability:    vulRs,
		Malware:          malwareRes,
		Rootkit:          rootkitRes,
		Secret:           secretRes,
		Misconfiguration: misconfigRes,
		Exploit:          exploitRes,
	}
}

func createModelScanResultsFromDB(scanResults *database.ScanResults) *models.ScanResults {
	var sbomRes models.SbomScan
	if scanResults.Sbom != nil {
		sbomRes = scanResults.Sbom.Results
	}
	var vulRes models.VulnerabilityScan
	if scanResults.Vulnerability != nil {
		vulRes = scanResults.Vulnerability.Results
	}
	var malwareRes models.MalwareScan
	if scanResults.Malware != nil {
		malwareRes = scanResults.Malware.Results
	}
	var secretRes models.SecretScan
	if scanResults.Secret != nil {
		secretRes = scanResults.Secret.Results
	}
	var misconfigRes models.MisconfigurationScan
	if scanResults.Misconfiguration != nil {
		misconfigRes = scanResults.Misconfiguration.Results
	}
	var rootkitRes models.RootkitScan
	if scanResults.Rootkit != nil {
		rootkitRes = scanResults.Rootkit.Results
	}
	var exploitRes models.ExploitScan
	if scanResults.Exploit != nil {
		exploitRes = scanResults.Exploit.Results
	}
	return &models.ScanResults{
		Id:                &scanResults.ID,
		Sboms:             &sbomRes,
		Vulnerabilities:   &vulRes,
		Malwares:          &malwareRes,
		Rootkits:          &rootkitRes,
		Secrets:           &secretRes,
		Misconfigurations: &misconfigRes,
		Exploits:          &exploitRes,
	}
}

func generateScanResultsID() string {
	return uuid.NewString()
}
