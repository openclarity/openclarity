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
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/openclarity/kubeclarity/api/client/client/operations"
	"github.com/openclarity/kubeclarity/api/client/models"
	"github.com/openclarity/kubeclarity/e2e/common"
)

func TestRuntimeScan(t *testing.T) {
	stopCh := make(chan struct{})
	defer func() {
		stopCh <- struct{}{}
		time.Sleep(2 * time.Second)
	}()
	f1 := features.New("assert results").
		WithLabel("type", "assert").
		Assess("vulnerability in DB", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Logf("Setup runtime env...")
			assert.NilError(t, setupRuntimeScanTestEnv(stopCh))

			t.Logf("start runtime scan...")
			assert.NilError(t, startRuntimeScan([]string{"test"}))
			// wait for progress DONE
			t.Logf("wait for scan done...")
			assert.NilError(t, waitForScanDone())

			t.Logf("get runtime scan results...")
			results := common.GetRuntimeScanResults(t, kubeclarityAPI)
			assert.Assert(t, results.Counters.Resources > 0)
			assert.Assert(t, results.Counters.Vulnerabilities > 0)
			assert.Assert(t, results.Counters.Packages > 0)
			assert.Assert(t, results.Counters.Applications > 0)

			assert.Assert(t, len(results.VulnerabilityPerSeverity) > 0)
			return ctx
		}).Feature()

	// test features
	testenv.Test(t, f1)
}

func startRuntimeScan(namespaces []string) error {
	params := operations.NewPutRuntimeScanStartParams().WithBody(&models.RuntimeScanConfig{
		Namespaces: namespaces,
	})
	_, err := kubeclarityAPI.Operations.PutRuntimeScanStart(params)
	return err
}

func waitForScanDone() error {
	timer := time.NewTimer(3 * time.Minute)
	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-timer.C:
			return fmt.Errorf("timeout reached")
		case <-ticker.C:
			params := operations.NewGetRuntimeScanProgressParams()
			res, err := kubeclarityAPI.Operations.GetRuntimeScanProgress(params)
			if err != nil {
				return err
			}
			if res.Payload.Status == models.RuntimeScanStatusDONE {
				return nil
			}
		}
	}
}

func setupRuntimeScanTestEnv(stopCh chan struct{}) error {
	println("Set up runtime scan test env...")

	println("creating namespace test...")
	if err := common.CreateNamespace(k8sClient, "test"); err != nil {
		return fmt.Errorf("failed to create test namepsace: %v", err)
	}

	println("deploying test image to test namespace...")
	if err := common.InstallTest("test"); err != nil {
		return fmt.Errorf("failed to install test image: %v", err)
	}

	if err := common.WaitForPodRunning(k8sClient, "test", "app=test"); err != nil {
		return fmt.Errorf("failed to wait for test pod running: %v", err)
	}

	println("port-forward to kubeclarity...")
	common.PortForwardToKubeClarity(stopCh)

	return nil
}
