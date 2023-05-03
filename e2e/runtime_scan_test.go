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
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"gotest.tools/assert"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/openclarity/kubeclarity/api/client/models"
	"github.com/openclarity/kubeclarity/e2e/common"
)

func TestRuntimeScan(t *testing.T) {
	stopCh := make(chan struct{})
	defer func() {
		stopCh <- struct{}{}
		time.Sleep(2 * time.Second)
	}()
	f1 := features.New("runtime scan").
		WithLabel("type", "runtime").
		Assess("runtime scan flow", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Logf("Setup runtime env...")
			assert.NilError(t, setupRuntimeScanTestEnv(stopCh))

			t.Logf("start runtime scan...")
			startRuntimeScan(t)

			// wait for progress DONE
			t.Logf("wait for scan done...")
			assert.NilError(t, waitForScanDone(t))

			t.Logf("get runtime scan results...")
			// wait for refreshing materialized views
			time.Sleep(common.WaitForMaterializedViewRefreshSecond * time.Second * 2)
			results := common.GetRuntimeScanResults(t, kubeclarityAPI)
			resultsB, err := json.Marshal(results)
			assert.NilError(t, err)
			t.Logf("Got runtime scan results: %+v", string(resultsB))

			assert.Assert(t, results.Counters.Resources > 0)
			assert.Assert(t, results.Counters.Vulnerabilities > 0)
			assert.Assert(t, results.Counters.Packages > 0)
			assert.Assert(t, results.Counters.Applications > 0)
			assert.Assert(t, results.CisDockerBenchmarkCounters.Resources > 0)
			assert.Assert(t, results.CisDockerBenchmarkCounters.Applications > 0)

			assert.Assert(t, len(results.VulnerabilityPerSeverity) > 0)
			assert.Assert(t, len(results.CisDockerBenchmarkCountPerLevel) > 0)

			// radial/busyboxplus:curl is not supported as it uses V1 image manifest, which should result in an error.
			assert.Assert(t, len(results.Failures) > 0)
			assert.Assert(t, strings.Contains(results.Failures[0].Message, "radial/busyboxplus:curl"))
			return ctx
		}).Feature()

	// test features
	testenv.Test(t, f1)
}

func waitForScanDone(t *testing.T) error {
	t.Helper()
	timer := time.NewTimer(3 * time.Minute)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			return fmt.Errorf("timeout reached")
		case <-ticker.C:
			progress := common.GetRuntimeScanProgress(t, kubeclarityAPI)
			if progress.Status == models.RuntimeScanStatusDONE {
				return nil
			}
		}
	}
}

func startRuntimeScan(t *testing.T) {
	t.Helper()
	// configuration values are based on createDefaultRuntimeQuickScanConfig()
	_ = common.PutRuntimeQuickScanConfig(t, kubeclarityAPI, &models.RuntimeQuickScanConfig{
		CisDockerBenchmarkScanEnabled: true,
		MaxScanParallelism:            10,
	})
	_ = common.PutRuntimeScanStart(t, kubeclarityAPI, &models.RuntimeScanConfig{
		Namespaces: []string{"test"},
	})
}

func setupRuntimeScanTestEnv(stopCh chan struct{}) error {
	println("Set up runtime scan test env...")

	println("creating namespace test...")
	if err := common.CreateNamespace(k8sClient, "test"); err != nil {
		return fmt.Errorf("failed to create test namepsace: %v", err)
	}

	println("deploying nginx and curl to test namespace...")
	if err := common.Deploy("test", "curl_nginx.yaml"); err != nil {
		return fmt.Errorf("failed to install curl_nginx.yaml: %v", err)
	}

	if err := common.WaitForPodRunning(k8sClient, "test", "app=nginx"); err != nil {
		return fmt.Errorf("failed to wait for nginx test pod running: %v", err)
	}

	if err := common.WaitForPodRunning(k8sClient, "test", "app=curl"); err != nil {
		return fmt.Errorf("failed to wait for curl test pod running: %v", err)
	}

	println("port-forward to kubeclarity...")
	common.PortForwardToKubeClarity(stopCh)

	return nil
}
