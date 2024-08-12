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

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
	uitypes "github.com/openclarity/vmclarity/uibackend/types"
)

var _ = ginkgo.Describe("Running a full scan (exploits, info finder, malware, misconfigurations, rootkits, SBOM, secrets and vulnerabilities)", func() {
	reportFailedConfig := ReportFailedConfig{}

	ginkgo.Context("which scans a docker container", func() {
		ginkgo.It("should finish successfully", func(ctx ginkgo.SpecContext) {
			ginkgo.By("waiting until test asset is found")
			reportFailedConfig.objects = append(
				reportFailedConfig.objects,
				APIObject{"asset", cfg.TestSuiteParams.Scope},
			)
			assetsParams := apitypes.GetAssetsParams{
				Filter: to.Ptr(cfg.TestSuiteParams.Scope),
			}
			gomega.Eventually(func() bool {
				assets, err := client.GetAssets(ctx, assetsParams)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				return len(*assets.Items) == 1
			}, DefaultTimeout, DefaultPeriod).Should(gomega.BeTrue())

			ginkgo.By("applying a scan configuration")
			apiScanConfig, err := client.PostScanConfig(
				ctx,
				GetCustomScanConfig(
					cfg.TestSuiteParams.FamiliesConfig,
					cfg.TestSuiteParams.Scope,
					cfg.TestSuiteParams.ScanTimeout,
				),
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("updating scan configuration to run now")
			updateScanConfig := UpdateScanConfigToStartNow(apiScanConfig)
			err = client.PatchScanConfig(ctx, *apiScanConfig.Id, updateScanConfig)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("waiting until scan starts")
			scanParams := apitypes.GetScansParams{
				Filter: to.Ptr(fmt.Sprintf(
					"scanConfig/id eq '%s' and status/state ne '%s' and status/state ne '%s'",
					*apiScanConfig.Id,
					apitypes.ScanStatusStateDone,
					apitypes.ScanStatusStateFailed,
				)),
			}
			var scans *apitypes.Scans
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
			}, DefaultTimeout, DefaultPeriod).Should(gomega.BeTrue())

			reportFailedConfig.objects = append(
				reportFailedConfig.objects,
				APIObject{"assetScan", fmt.Sprintf("scan/id eq '%s'", *apiScanConfig.Id)},
			)

			ginkgo.By("waiting until scan state changes to done")
			scanParams = apitypes.GetScansParams{
				Filter: to.Ptr(fmt.Sprintf(
					"scanConfig/id eq '%s' and status/state eq '%s' and status/reason eq '%s'",
					*apiScanConfig.Id,
					apitypes.ScanStatusStateDone,
					apitypes.ScanStatusReasonSuccess,
				)),
			}
			gomega.Eventually(func() bool {
				scans, err = client.GetScans(ctx, scanParams)
				gomega.Expect(skipDBLockedErr(err)).NotTo(gomega.HaveOccurred())
				return len(*scans.Items) == 1
			}, cfg.TestSuiteParams.ScanTimeout, DefaultPeriod).Should(gomega.BeTrue())

			ginkgo.By("waiting until asset is found in riskiest assets dashboard")
			gomega.Eventually(func() bool {
				riskiestAssets, err := uiClient.GetDashboardRiskiestAssets(ctx)
				gomega.Expect(skipDBLockedErr(err)).NotTo(gomega.HaveOccurred()) // nolint:wrapcheck
				if riskiestAssets == nil {
					return false
				}
				for _, v := range *riskiestAssets.Vulnerabilities {
					if *v.CriticalVulnerabilitiesCount > 1 {
						return true
					}
				}
				return false
			}, DefaultTimeout*2, DefaultPeriod).Should(gomega.BeTrue())

			ginkgo.By("waiting until findings trends dashboard is populated with vulnerabilities")
			gomega.Eventually(func() bool {
				findingsTrends, err := uiClient.GetDashboardFindingsTrends(
					ctx,
					uitypes.GetDashboardFindingsTrendsParams{
						StartTime: time.Now().Add(-time.Minute * 10),
						EndTime:   time.Now(),
					},
				)
				gomega.Expect(skipDBLockedErr(err)).NotTo(gomega.HaveOccurred()) // nolint:wrapcheck
				if findingsTrends == nil {
					return false
				}
				for _, trend := range *findingsTrends {
					if *trend.FindingType == uitypes.VULNERABILITY {
						for _, t := range *trend.Trends {
							if *t.Count > 0 {
								return true
							}
						}
					}
				}
				return false
			}, DefaultTimeout, DefaultPeriod).Should(gomega.BeTrue())

			ginkgo.By("waiting until findings impact dashboard is populated with vulnerabilities")
			gomega.Eventually(func() bool {
				findingsImpact, err := uiClient.GetDashboardFindingsImpact(ctx)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				return findingsImpact != nil && findingsImpact.Vulnerabilities != nil && len(*findingsImpact.Vulnerabilities) > 0
			}, DefaultTimeout*2, DefaultPeriod).Should(gomega.BeTrue())

			ginkgo.By("waiting until at least one findings summary has been updated")
			vulnerabilityCountFilter := strings.Join([]string{
				"summary/totalVulnerabilities/totalCriticalVulnerabilities gt 0",
				"summary/totalVulnerabilities/totalHighVulnerabilities gt 0",
				"summary/totalVulnerabilities/totalMediumVulnerabilities gt 0",
				"summary/totalVulnerabilities/totalLowVulnerabilities gt 0",
				"summary/totalVulnerabilities/totalNegligibleVulnerabilities gt 0",
			}, " or ")
			gomega.Eventually(func() bool {
				findings, err := client.GetFindings(ctx, apitypes.GetFindingsParams{
					Filter: to.Ptr(fmt.Sprintf("summary/updatedAt ne null and (%s)", vulnerabilityCountFilter)),
					Top:    to.Ptr(1),
					Count:  to.Ptr(true),
				})
				gomega.Expect(skipDBLockedErr(err)).NotTo(gomega.HaveOccurred())
				return *findings.Count > 0
			}, DefaultTimeout*2, DefaultPeriod).Should(gomega.BeTrue())
		})
	})

	ginkgo.AfterEach(func(ctx ginkgo.SpecContext) {
		if ginkgo.CurrentSpecReport().Failed() {
			reportFailedConfig.startTime = ginkgo.CurrentSpecReport().StartTime
			ReportFailed(ctx, testEnv, client, &reportFailedConfig)
		}
	})
})
