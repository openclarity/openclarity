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

package e2e

import (
	"context"
	"fmt"
	"gotest.tools/assert"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/openclarity/kubeclarity/api/client/client/operations"
	"github.com/openclarity/kubeclarity/api/client/models"
	"github.com/openclarity/kubeclarity/e2e/common"
	"github.com/openclarity/kubeclarity/shared/pkg/formatter"
)

const (
	DirectoryAnalyzeOutputSBOMFile = "dir.sbom"
	ImageAnalyzeOutputSBOMFile     = "merged.sbom"
	TestImageName                  = "erezfish/test:1.1"
	ApplicationName                = "test-app"
)

func TestCLIScan(t *testing.T) {
	t.Logf("Starting test CLI scan")

	stopCh := make(chan struct{})
	defer func() {
		stopCh <- struct{}{}
		time.Sleep(2 * time.Second)
	}()
	f1 := features.New("cli scan flow - analyze and scan").
		WithLabel("type", "cli").
		Assess("cli scan flow", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// setup env
			t.Logf("setup env...")
			assert.NilError(t, setupCLIScanTestEnv(stopCh))

			// create application
			t.Logf("create application...")
			appID, err := createApplication()
			assert.NilError(t, err)

			// analyze dir
			t.Logf("analyze dir...")
			assert.NilError(t, analyzeDir())
			validateAnalyzeDir(t)

			// analyze image with --merge-sbom directory sbom, and export to backend
			t.Logf("analyze image...")
			assert.NilError(t, analyzeImage(t, DirectoryAnalyzeOutputSBOMFile, appID))
			validateAnalyzeImage(t)

			// scan merged sbom
			t.Logf("scan merged sbom...")
			assert.NilError(t, scanSBOM(t, ImageAnalyzeOutputSBOMFile, appID))
			validateScanSBOM(t)

			// scan image
			// TODO upload a test image to github repo and use it here.
			t.Logf("scan image...")
			assert.NilError(t, scanImage(t, TestImageName, appID))
			validateScanImage(t)

			return ctx
		}).Feature()

	// test features
	testenv.Test(t, f1)
}

func getCdxSbom(t *testing.T, fileName string) *cdx.BOM {
	sbom, err := os.ReadFile(fileName)
	assert.NilError(t, err)
	input := formatter.New(formatter.CycloneDXFormat, sbom)
	assert.NilError(t, input.Decode(formatter.CycloneDXFormat))
	return input.GetSBOM().(*cdx.BOM)
}

func validateAnalyzeDir(t *testing.T) {
	sbom := getCdxSbom(t, DirectoryAnalyzeOutputSBOMFile)
	assert.Assert(t, sbom != nil)
	assert.Assert(t, sbom.Components != nil)
	//assert.Assert(t, sbom.Metadata.Component.Name == "github.com/cisco-open/kubei/e2e/test_ananlyze")
	assert.Assert(t, len(*sbom.Components) > 0)
	// TODO assert properties - analyzers = syft, gomod
}

func validateAnalyzeImage(t *testing.T) {
	sbom := getCdxSbom(t, ImageAnalyzeOutputSBOMFile)
	assert.Assert(t, sbom != nil)
	// check generated sbom
	assert.Assert(t, sbom.Components != nil)
	//assert.Assert(t, sbom.Metadata.Component.Name == TestImageName)
	assert.Assert(t, len(*sbom.Components) > 0)
	// TODO assert properties - analyzers = syft
	// TODO validate merged? how? more components?
	//assert.Assert(t, *sbom.Components[0].Properties)

	// check export to db
	packages := common.GetPackages(t, kubeclarityAPI)
	assert.Assert(t, *packages.Total > 0)

	appResources := common.GetApplicationResources(t, kubeclarityAPI)
	assert.Assert(t, *appResources.Total > 0)
}

func validateScanImage(t *testing.T) {
	vuls := common.GetVulnerabilities(t, kubeclarityAPI)
	assert.Assert(t, *vuls.Total > 0)
}

func validateScanSBOM(t *testing.T) {
	vuls := common.GetVulnerabilities(t, kubeclarityAPI)

	// TODO how to validate that vulnerabilities were added on top of scanned sbom vuls
	assert.Assert(t, *vuls.Total > 0)
}

func createApplication() (id string, err error) {
	var app *operations.PostApplicationsCreated
	appType := models.ApplicationTypePOD
	params := operations.NewPostApplicationsParams().WithBody(&models.ApplicationInfo{
		Name: common.StringPtr(ApplicationName),
		Type: &appType,
	})
	app, err = kubeclarityAPI.Operations.PostApplications(params)
	if err != nil {
		return "", fmt.Errorf("failed to post application to backend: %v", err)
	}
	id = app.Payload.ID
	return
}

func analyzeDir() error {
	dirPath := filepath.Join(common.GetCurrentDir(), "test_analyze")

	command := cliPath + " analyze " + dirPath + " --input-type dir -o " + DirectoryAnalyzeOutputSBOMFile

	cmd := exec.Command("/bin/sh", "-c", command)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute command. %v, %s", err, out)
	}
	return nil
}

var cliPath = filepath.Join(common.GetCurrentDir(), "kubeclarity-cli")

// analyze test image, merge inputSbom and export to backend
func analyzeImage(t *testing.T, inputSbom string, appID string) error {
	assert.NilError(t, os.Setenv("BACKEND_HOST", "localhost:"+common.KubeClarityPortForwardHostPort))
	assert.NilError(t, os.Setenv("BACKEND_DISABLE_TLS", "true"))

	command := fmt.Sprintf("%v analyze %v --application-id %v --input-type image --merge-sbom %v -e -o %v", cliPath, TestImageName, appID, inputSbom, ImageAnalyzeOutputSBOMFile)

	cmd := exec.Command("/bin/sh", "-c", command)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute command. %v, %s", err, out)
	}
	return nil
}

func scanSBOM(t *testing.T, inputSbom string, appID string) error {
	assert.NilError(t, os.Setenv("BACKEND_HOST", "localhost:"+common.KubeClarityPortForwardHostPort))
	assert.NilError(t, os.Setenv("BACKEND_DISABLE_TLS", "true"))

	command := fmt.Sprintf("%v scan %v --application-id %v --input-type sbom -e", cliPath, inputSbom, appID)

	cmd := exec.Command("/bin/sh", "-c", command)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute command. %v, %s", err, out)
	}
	return nil
}

func scanImage(t *testing.T, image string, appID string) error {
	assert.NilError(t, os.Setenv("BACKEND_HOST", "localhost:"+common.KubeClarityPortForwardHostPort))
	assert.NilError(t, os.Setenv("BACKEND_DISABLE_TLS", "true"))

	command := fmt.Sprintf("%v scan %v --application-id %v --input-type image -e", cliPath, image, appID)

	cmd := exec.Command("/bin/sh", "-c", command)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute command. %v, %s", err, out)
	}
	return nil
}

func setupCLIScanTestEnv(stopCh chan struct{}) error {
	println("Set up cli scan test env...")

	println("port-forward to kubeclarity...")
	common.PortForwardToKubeClarity(stopCh)

	return nil
}
