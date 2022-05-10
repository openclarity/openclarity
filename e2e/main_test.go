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
	"testing"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/third_party/helm"

	"github.com/openclarity/kubeclarity/api/client/client"
	"github.com/openclarity/kubeclarity/e2e/common"
)

var (
	testenv        env.Environment
	KubeconfigFile string
	kubeclarityAPI *client.KubeClarityAPIs
	k8sClient      klient.Client
	helmManager    *helm.Manager
)

func TestMain(m *testing.M) {
	testenv = env.New()
	kindClusterName := envconf.RandomName("my-cluster", 16)

	testenv.Setup(
		envfuncs.CreateKindClusterWithConfig(kindClusterName, "kindest/node:v1.22.2", "kind-config.yaml"),
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			println("Setup")
			k8sClient = cfg.Client()

			tag := os.Getenv("DOCKER_TAG")

			println("DOCKER_TAG=", tag)

			if err := common.LoadDockerImagesToCluster(kindClusterName, tag); err != nil {
				fmt.Printf("Failed to load docker images to cluster: %v", err)
				return nil, err
			}

			clientTransport := httptransport.New("localhost:"+common.KubeClarityPortForwardHostPort, client.DefaultBasePath, []string{"http"})
			kubeclarityAPI = client.New(clientTransport, strfmt.Default)

			KubeconfigFile = cfg.KubeconfigFile()
			helmManager = helm.New(KubeconfigFile)

			return ctx, nil
		},
	)
	testenv.Finish(
		envfuncs.DestroyKindCluster(kindClusterName),
	)
	testenv.BeforeEachTest(
		func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
			t.Logf("BeforeEachTest")

			t.Logf("deploying kubeclarity...")
			if err := common.InstallKubeClarity(helmManager, "--create-namespace --wait"); err != nil {
				return nil, fmt.Errorf("failed to install kubeclarity: %v", err)
			}

			t.Logf("waiting for kubeclarity to run...")
			if err := common.WaitForPodRunning(k8sClient, common.KubeClarityNamespace, "app=kubeclarity-kubeclarity"); err != nil {
				common.DescribeKubeClarityDeployment()
				common.DescribeKubeClarityPods()
				return nil, fmt.Errorf("failed to wait for kubeclarity pod to be running: %v", err)
			}

			return ctx, nil
		},
	)
	testenv.AfterEachTest(
		func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
			t.Logf("AfterEachTest")

			t.Logf("uninstalling kubeclarity...")
			assert.NilError(t, common.UninstallKubeClarity())
			return ctx, nil
		},
	)

	// launch package tests
	os.Exit(testenv.Run(m))
}
