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
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/compose-spec/compose-go/cli"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/e2e/testenv"
	"github.com/openclarity/vmclarity/pkg/shared/backendclient"
	"github.com/openclarity/vmclarity/pkg/shared/log"
)

var (
	testEnv *testenv.Environment
	client  *backendclient.BackendClient
)

func TestIntegrationTest(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Run integration tests")
}

func beforeSuite(ctx context.Context) {
	var err error

	ginkgo.By("creating test environment")
	log.InitLogger(logrus.DebugLevel.String(), os.Stderr)
	logger := logrus.WithContext(ctx)
	ctx = log.SetLoggerForContext(ctx, logger)

	opts, err := cli.NewProjectOptions(
		[]string{"../installation/docker/docker-compose.yml", "docker-compose.override.yml"},
		cli.WithName("vmclarity-e2e"),
		cli.WithResolvedPaths(true),
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	err = cli.WithOsEnv(opts)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	var reuseEnv bool
	if reuseEnv, _ = strconv.ParseBool(os.Getenv("USE_EXISTING")); reuseEnv {
		logger.Info("reusing existing environment...", "use_existing", reuseEnv)
	}

	testEnv, err = testenv.New(opts, reuseEnv)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	ginkgo.By("starting test environment")
	err = testEnv.Start(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	gomega.Eventually(func() bool {
		ready, err := testEnv.ServicesReady(ctx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return ready
	}, time.Second*5).Should(gomega.BeTrue())

	u, err := testEnv.VMClarityURL()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	client, err = backendclient.Create(fmt.Sprintf("%s://%s/%s", u.Scheme, u.Host, u.Path))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

var _ = ginkgo.BeforeSuite(beforeSuite)

func afterSuite(ctx context.Context) {
	ginkgo.By("tearing down test environment")
	err := testEnv.Stop(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

var _ = ginkgo.AfterSuite(afterSuite)
