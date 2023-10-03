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
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

var _ = ginkgo.Describe("Posting and getting a provider", func() {
	reportFailedConfig := ReportFailedConfig{}

	ginkgo.Context("", func() {
		ginkgo.It("should work successfully", func(ctx ginkgo.SpecContext) {
			ginkgo.By("applying a provider")
			apiProvider, err := client.PostProvider(
				ctx,
				models.Provider{
					DisplayName: utils.PointerTo("test-provider"),
					Status: utils.PointerTo(models.ProviderStatus{
						State:              models.ProviderStatusStateHealthy,
						Reason:             models.HeartbeatReceived,
						LastTransitionTime: time.Now(),
					}),
				})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			reportFailedConfig.objects = append(
				reportFailedConfig.objects,
				APIObject{"provider", fmt.Sprintf("id eq '%s'", *apiProvider.Id)},
			)

			ginkgo.By("waiting until test provider is found")
			gomega.Eventually(func() bool {
				providers, err := client.GetProviders(
					ctx,
					models.GetProvidersParams{
						Filter: utils.PointerTo(fmt.Sprintf("id eq '%s'", *apiProvider.Id)),
					})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				return len(*providers.Items) == 1
			}, DefaultTimeout, time.Second).Should(gomega.BeTrue())
		})
	})

	ginkgo.AfterEach(func(ctx ginkgo.SpecContext) {
		if ginkgo.CurrentSpecReport().Failed() {
			reportFailedConfig.startTime = ginkgo.CurrentSpecReport().StartTime
			ReportFailed(ctx, testEnv, client, &reportFailedConfig)
		}
	})
})
