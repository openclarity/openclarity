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
	"fmt"
	"net/http"
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/stretchr/testify/assert"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/backend/pkg/database"
)

var scanResultsPath = fmt.Sprintf("%s/targets/%s/scanResults", baseURL, testID)

func TestScanResultsController(t *testing.T) {
	restServer := createTestRestServer(t)

	fakeHandler := database.NewFakeDatabase()
	restServer.RegisterHandlers(fakeHandler)

	targetID := testID
	targetType := models.TargetType("VM")
	scanResults := []string{}
	instanceName := "instance"
	instanceProvider := models.CloudProvider("AWS")
	location := "eu-central2"
	vmInfo := models.VMInfo{
		InstanceName:     &instanceName,
		InstanceProvider: &instanceProvider,
		Location:         &location,
	}
	targetInfo := &models.Target_TargetInfo{}
	if err := targetInfo.FromVMInfo(vmInfo); err != nil {
		t.Errorf("failed to create target info")
	}

	newTarget := models.Target{
		Id:          &targetID,
		ScanResults: &scanResults,
		TargetType:  &targetType,
		TargetInfo:  targetInfo,
	}

	// Create new target.
	result := testutil.NewRequest().Post(fmt.Sprintf("%s/targets", baseURL)).WithJsonBody(newTarget).Go(t, restServer.echoServer)
	assert.Equal(t, http.StatusCreated, result.Code())
	target := models.Target{}
	if err := result.UnmarshalBodyToObject(&target); err != nil {
		t.Errorf("failed to unmarshal response body")
	}
	assert.Equal(t, newTarget, target)

	packageID := "packageID"
	packageName := "testPackage"
	packageVersion := "testVersion"
	scanResID := testID
	newScanResults := models.ScanResults{
		Id: &scanResID,
		Sboms: &models.SbomScan{
			Packages: &[]models.Package{
				{
					Id: &packageID,
					PackageInfo: &models.PackageInfo{
						Id:             &packageID,
						PackageName:    &packageName,
						PackageVersion: &packageVersion,
					},
				},
			},
		},
		Vulnerabilities: &models.VulnerabilityScan{
			Vulnerabilities: &[]models.Vulnerability{},
		},
		Malwares: &models.MalwareScan{
			Malwares: &[]models.MalwareInfo{},
		},
		Misconfigurations: &models.MisconfigurationScan{
			Misconfigurations: &[]models.MisconfigurationInfo{},
		},
		Secrets: &models.SecretScan{
			Secrets: &[]models.SecretInfo{},
		},
		Rootkits: &models.RootkitScan{
			Rootkits: &[]models.RootkitInfo{},
		},
		Exploits: &models.ExploitScan{
			Exploits: &[]models.ExploitInfo{},
		},
	}

	// POST new scan results to target
	result = testutil.NewRequest().Post(scanResultsPath).WithJsonBody(newScanResults).Go(t, restServer.echoServer)
	assert.Equal(t, http.StatusCreated, result.Code())
	got := models.ScanResults{}
	if err := result.UnmarshalBodyToObject(&got); err != nil {
		t.Errorf("failed to unmarshal response body")
	}

	assert.Equal(t, newScanResults, got)

	// Get scan results for specified target
	result = testutil.NewRequest().Get(fmt.Sprintf("%s?page=1&pageSize=1", scanResultsPath)).Go(t, restServer.echoServer)
	assert.Equal(t, http.StatusOK, result.Code())
	var gotList []models.ScanResults
	if err := result.UnmarshalBodyToObject(&gotList); err != nil {
		t.Errorf("failed to unmarshal response body")
	}
	wantList := []models.ScanResults{newScanResults}
	assert.Equal(t, wantList, gotList)

	// Get scanResults with ID
	result = testutil.NewRequest().Get(fmt.Sprintf("%s/%s", scanResultsPath, testID)).Go(t, restServer.echoServer)
	assert.Equal(t, http.StatusOK, result.Code())
	scanRes := models.ScanResults{}
	if err := result.UnmarshalBodyToObject(&scanRes); err != nil {
		t.Errorf("failed to unmarshal response body")
	}
	assert.Equal(t, newScanResults, scanRes)

	// Get scanResults with wrong ID
	result = testutil.NewRequest().Get(fmt.Sprintf("%s/wrongID", scanResultsPath)).Go(t, restServer.echoServer)
	assert.Equal(t, http.StatusNotFound, result.Code())

	// Update scan results
	vulnerabilityID := "vulnerabilityID"
	vulnerabilityName := "testVulName"
	vulnerabilityDesc := "Description"
	updatedScanResults := newScanResults
	updatedScanResults.Vulnerabilities = &models.VulnerabilityScan{
		Vulnerabilities: &[]models.Vulnerability{
			{
				Id: &vulnerabilityID,
				VulnerabilityInfo: &models.VulnerabilityInfo{
					Id:                &vulnerabilityID,
					VulnerabilityName: &vulnerabilityName,
					Description:       &vulnerabilityDesc,
				},
			},
		},
	}
	result = testutil.NewRequest().Put(fmt.Sprintf("%s/%s", scanResultsPath, testID)).WithJsonBody(updatedScanResults).Go(t, restServer.echoServer)
	assert.Equal(t, http.StatusOK, result.Code())

	// Get specified Vulnerability scan results
	result = testutil.NewRequest().Get(fmt.Sprintf("%s/%s", scanResultsPath, testID)).Go(t, restServer.echoServer)
	assert.Equal(t, http.StatusOK, result.Code())
	var updatedRes models.ScanResults
	if err := result.UnmarshalBodyToObject(&updatedRes); err != nil {
		t.Errorf("failed to unmarshal response body")
	}
	assert.Equal(t, updatedScanResults, updatedRes)

	scanResultsWithoutID := newScanResults
	scanResultsWithoutID.Id = nil
	// Get scan results for specified target
	result = testutil.NewRequest().Get(fmt.Sprintf("%s?page=1&pageSize=1", scanResultsPath)).Go(t, restServer.echoServer)
	assert.Equal(t, http.StatusOK, result.Code())
	if err := result.UnmarshalBodyToObject(&gotList); err != nil {
		t.Errorf("failed to unmarshal response body")
	}
	wantList = []models.ScanResults{scanResultsWithoutID}
	assert.Equal(t, wantList[0].Sboms, gotList[0].Sboms)
}
