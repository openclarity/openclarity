// Copyright © 2021 Cisco Systems, Inc. and its affiliates.
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

package common

import (
	"github.com/openclarity/kubeclarity/api/client/models"
	"testing"

	"gotest.tools/assert"

	"github.com/openclarity/kubeclarity/api/client/client"
	"github.com/openclarity/kubeclarity/api/client/client/operations"
)


func PostApplications(t *testing.T, kubeclarityAPI *client.KubeClarityAPIs, applicationInfo *models.ApplicationInfo) *operations.PostApplicationsCreated {
	t.Helper()
	params := operations.NewPostApplicationsParams().WithBody(applicationInfo)
	app, err := kubeclarityAPI.Operations.PostApplications(params)
	assert.NilError(t, err)

	return app
}

func PutRuntimeScanStart(t *testing.T, kubeclarityAPI *client.KubeClarityAPIs, config *models.RuntimeScanConfig) *operations.PutRuntimeScanStartCreated {
	t.Helper()
	params := operations.NewPutRuntimeScanStartParams().WithBody(config)
	res, err := kubeclarityAPI.Operations.PutRuntimeScanStart(params)
	assert.NilError(t, err)

	return res
}

func GetRuntimeScanProgress(t *testing.T, kubeclarityAPI *client.KubeClarityAPIs) *models.Progress {
	t.Helper()
	params := operations.NewGetRuntimeScanProgressParams()
	res, err := kubeclarityAPI.Operations.GetRuntimeScanProgress(params)
	assert.NilError(t, err)

	return res.Payload
}

func GetRuntimeScanResults(t *testing.T, kubeclarityAPI *client.KubeClarityAPIs) *models.RuntimeScanResults {
	t.Helper()
	params := operations.NewGetRuntimeScanResultsParams()
	res, err := kubeclarityAPI.Operations.GetRuntimeScanResults(params)
	assert.NilError(t, err)

	return res.Payload
}

func GetPackages(t *testing.T, kubeclarityAPI *client.KubeClarityAPIs) *operations.GetPackagesOKBody {
	t.Helper()
	params := operations.NewGetPackagesParams().WithPage(0).WithPageSize(50).WithSortKey("packageName")
	res, err := kubeclarityAPI.Operations.GetPackages(params)
	assert.NilError(t, err)

	return res.Payload
}

func GetApplicationResources(t *testing.T, kubeclarityAPI *client.KubeClarityAPIs) *operations.GetApplicationResourcesOKBody {
	t.Helper()
	params := operations.NewGetApplicationResourcesParams().WithPage(0).WithPageSize(50).WithSortKey("resourceName")
	res, err := kubeclarityAPI.Operations.GetApplicationResources(params)
	assert.NilError(t, err)

	return res.Payload
}

func GetVulnerabilities(t *testing.T, kubeclarityAPI *client.KubeClarityAPIs) *operations.GetVulnerabilitiesOKBody {
	t.Helper()
	params := operations.NewGetVulnerabilitiesParams().WithPage(0).WithPageSize(50).WithSortKey("vulnerabilityName")
	res, err := kubeclarityAPI.Operations.GetVulnerabilities(params)
	assert.NilError(t, err)

	return res.Payload
}

