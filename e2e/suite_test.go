// Copyright © 2023 Cisco Systems, Inc. and its affiliates.
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
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/e2e/testenv"
	"github.com/openclarity/vmclarity/e2e/testenv/types"
	"github.com/openclarity/vmclarity/pkg/shared/backendclient"
	"github.com/openclarity/vmclarity/pkg/shared/log"
	"github.com/openclarity/vmclarity/pkg/shared/uibackendclient"
)

var (
	testEnv  types.Environment
	client   *backendclient.BackendClient
	uiClient *uibackendclient.UIBackendClient
	config   *types.Config
)

func TestEndToEnd(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Run end-to-end tests")
}

func beforeSuite(ctx context.Context) {
	var err error

	ginkgo.By("initializing test environment")
	log.InitLogger(logrus.DebugLevel.String(), os.Stderr)
	logger := logrus.WithContext(ctx)
	ctx = log.SetLoggerForContext(ctx, logger)

	config, err = testenv.NewConfig()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	testEnv, err = testenv.New(config)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	if !config.ReuseEnv {
		ginkgo.By("setup test environment")
		err = testEnv.SetUp(ctx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By("starting test environment")
		err = testEnv.Start(ctx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	} else {
		ginkgo.By("re-using test environment")
	}

	ginkgo.By("waiting for services to become ready")
	gomega.Eventually(func() bool {
		ready, err := testEnv.ServicesReady(ctx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return ready
	}, time.Second*5).Should(gomega.BeTrue())

	u, err := testEnv.GetGatewayServiceURL()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	base := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	clientURL, err := url.JoinPath(base, "api")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	client, err = backendclient.Create(clientURL)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	uiClientURL, err := url.JoinPath(base, "ui", "api")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	uiClient, err = uibackendclient.Create(uiClientURL)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

var _ = ginkgo.BeforeSuite(beforeSuite)

func afterSuite(ctx context.Context) {
	if !config.ReuseEnv {
		ginkgo.By("stopping test environment")
		err := testEnv.Stop(ctx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By("tearing down test environment")
		err = testEnv.TearDown(ctx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}
}

var _ = ginkgo.AfterSuite(afterSuite)
