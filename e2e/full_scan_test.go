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
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
	uibackend_models "github.com/openclarity/vmclarity/pkg/uibackend/api/models"
)

var _ = ginkgo.Describe("Running a full scan (exploits, info finder, malware, misconfigurations, rootkits, SBOM, secrets and vulnerabilities)", func() {
	reportFailedConfig := ReportFailedConfig{}

	ginkgo.Context("which scans a docker container", func() {
		ginkgo.It("should finish successfully", func(ctx ginkgo.SpecContext) {
			ginkgo.By("applying a scan configuration")
			apiScanConfig, err := client.PostScanConfig(ctx, GetFullScanConfig())
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("updating scan configuration to run now")
			updateScanConfig := UpdateScanConfigToStartNow(apiScanConfig)
			err = client.PatchScanConfig(ctx, *apiScanConfig.Id, updateScanConfig)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("waiting until scan starts")
			scanParams := models.GetScansParams{
				Filter: utils.PointerTo(fmt.Sprintf(
					"scanConfig/id eq '%s' and state ne '%s' and state ne '%s'",
					*apiScanConfig.Id,
					models.ScanStateDone,
					models.ScanStateFailed,
				)),
			}
			var scans *models.Scans
			gomega.Eventually(func() bool {
				scans, err = client.GetScans(ctx, scanParams)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				if len(*scans.Items) == 1 {
					reportFailedConfig.objects = append(
						reportFailedConfig.objects,
						APIObject{"scan", fmt.Sprintf("id eq '%s'", *(*scans.Items)[0].Id)},
					)
					return true
				}
				return false
			}, DefaultTimeout, time.Second).Should(gomega.BeTrue())

			ginkgo.By("waiting until scan state changes to done")
			scanParams = models.GetScansParams{
				Filter: utils.PointerTo(fmt.Sprintf(
					"scanConfig/id eq '%s' and state eq '%s'",
					*apiScanConfig.Id,
					models.ScanStateDone,
				)),
			}
			gomega.Eventually(func() bool {
				scans, err = client.GetScans(ctx, scanParams)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				return len(*scans.Items) == 1
			}, DefaultTimeout*10, time.Second).Should(gomega.BeTrue())

			ginkgo.By("waiting until asset is found in riskiest assets dashboard")
			gomega.Eventually(func() bool {
				riskiestAssets, err := uiClient.GetDashboardRiskiestAssets(ctx)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				if riskiestAssets == nil {
					return false
				}
				for _, v := range *riskiestAssets.Vulnerabilities {
					if strings.Contains(*v.AssetInfo.Name, "alpine") && *v.CriticalVulnerabilitiesCount > 1 {
						return true
					}
				}
				return false
			}, DefaultTimeout*2, time.Second).Should(gomega.BeTrue())

			ginkgo.By("waiting until findings trends dashboard is populated with vulnerabilities")
			gomega.Eventually(func() bool {
				findingsTrends, err := uiClient.GetDashboardFindingsTrends(
					ctx,
					uibackend_models.GetDashboardFindingsTrendsParams{
						StartTime: time.Now().Add(-time.Minute * 10),
						EndTime:   time.Now(),
					},
				)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				if findingsTrends == nil {
					return false
				}
				for _, trend := range *findingsTrends {
					if *trend.FindingType == uibackend_models.VULNERABILITY {
						for _, t := range *trend.Trends {
							if *t.Count > 0 {
								return true
							}
						}
					}
				}
				return false
			}, DefaultTimeout, time.Second).Should(gomega.BeTrue())

			ginkgo.By("waiting until findings impact dashboard is populated with vulnerabilities")
			gomega.Eventually(func() bool {
				findingsImpact, err := uiClient.GetDashboardFindingsImpact(ctx)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				return findingsImpact != nil && findingsImpact.Vulnerabilities != nil && len(*findingsImpact.Vulnerabilities) > 0
			}, DefaultTimeout*2, time.Second).Should(gomega.BeTrue())
		})
	})

	ginkgo.AfterEach(func(ctx ginkgo.SpecContext) {
		if ginkgo.CurrentSpecReport().Failed() {
			reportFailedConfig.startTime = ginkgo.CurrentSpecReport().StartTime
			ReportFailed(ctx, testEnv, client, &reportFailedConfig)
		}
	})
})
