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

func TestTargetsController(t *testing.T) {
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

	// Create new target
	result := testutil.NewRequest().Post(fmt.Sprintf("%s/targets", baseURL)).WithJsonBody(newTarget).Go(t, restServer.echoServer)
	assert.Equal(t, http.StatusCreated, result.Code())
	got := models.Target{}
	if err := result.UnmarshalBodyToObject(&got); err != nil {
		t.Errorf("failed to unmarshal response body")
	}
	assert.Equal(t, newTarget, got)

	// List targets
	result = testutil.NewRequest().Get(fmt.Sprintf("%s/targets?page=1&pageSize=1", baseURL)).Go(t, restServer.echoServer)
	assert.Equal(t, http.StatusOK, result.Code())
	var gotList []models.Target
	if err := result.UnmarshalBodyToObject(&gotList); err != nil {
		t.Errorf("failed to unmarshal response body")
	}
	want := []models.Target{newTarget}
	assert.Equal(t, want, gotList)

	// Get target with ID
	result = testutil.NewRequest().Get(fmt.Sprintf("%s/targets/%s", baseURL, testID)).Go(t, restServer.echoServer)
	assert.Equal(t, http.StatusOK, result.Code())
	if err := result.UnmarshalBodyToObject(&got); err != nil {
		t.Errorf("failed to unmarshal response body")
	}
	assert.Equal(t, newTarget, got)

	// Get target with wrong ID
	result = testutil.NewRequest().Get(fmt.Sprintf("%s/targets/wrongID", baseURL)).Go(t, restServer.echoServer)
	assert.Equal(t, http.StatusNotFound, result.Code())

	updatedTarget := newTarget
	updatedScanResults := []string{"1", "2"}
	updatedTarget.ScanResults = &updatedScanResults
	result = testutil.NewRequest().Put(fmt.Sprintf("%s/targets/%s", baseURL, testID)).WithJsonBody(updatedTarget).Go(t, restServer.echoServer)
	assert.Equal(t, http.StatusOK, result.Code())

	// Get target with ID after update
	result = testutil.NewRequest().Get(fmt.Sprintf("%s/targets/%s", baseURL, testID)).Go(t, restServer.echoServer)
	assert.Equal(t, http.StatusOK, result.Code())
	if err := result.UnmarshalBodyToObject(&got); err != nil {
		t.Errorf("failed to unmarshal response body")
	}
	assert.Equal(t, updatedTarget, got)

	// Delete target with wrong ID
	result = testutil.NewRequest().Delete(fmt.Sprintf("%s/targets/wrongID", baseURL)).Go(t, restServer.echoServer)
	assert.Equal(t, http.StatusNotFound, result.Code())

	// Delete target
	result = testutil.NewRequest().Delete(fmt.Sprintf("%s/targets/%s", baseURL, testID)).Go(t, restServer.echoServer)
	assert.Equal(t, http.StatusNoContent, result.Code())

	// Get target with ID after delete
	result = testutil.NewRequest().Get(fmt.Sprintf("%s/targets/%s", baseURL, testID)).Go(t, restServer.echoServer)
	assert.Equal(t, http.StatusNotFound, result.Code())

	// List targets after delete
	result = testutil.NewRequest().Get(fmt.Sprintf("%s/targets?page=1&pageSize=1", baseURL)).Go(t, restServer.echoServer)
	assert.Equal(t, http.StatusOK, result.Code())
	if err := result.UnmarshalBodyToObject(&gotList); err != nil {
		t.Errorf("failed to unmarshal response body")
	}
	want = []models.Target{}
	assert.Equal(t, want, gotList)
}
