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

