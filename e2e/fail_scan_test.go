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

var _ = ginkgo.Describe("Detecting scan failures", func() {
	reportFailedConfig := ReportFailedConfig{}

	ginkgo.Context("when a scan stops without assets to scan", func() {
		ginkgo.It("should detect failure reason successfully", func(ctx ginkgo.SpecContext) {
			ginkgo.By("applying a scan configuration with not existing label")
			apiScanConfig, err := client.PostScanConfig(
				ctx,
				GetCustomScanConfig(
					&FullScanFamiliesConfig,
					"assetInfo/labels/any(t: t/key eq 'notexisting' and t/value eq 'label')",
					600,
				))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("updating scan configuration to run now")
			updateScanConfig := UpdateScanConfigToStartNow(apiScanConfig)
			err = client.PatchScanConfig(ctx, *apiScanConfig.Id, updateScanConfig)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("waiting until scan starts")
			scanParams := models.GetScansParams{
				Filter: utils.PointerTo(fmt.Sprintf(
					"scanConfig/id eq '%s'",
					*apiScanConfig.Id,
				)),
			}
			gomega.Eventually(func() bool {
				scans, err := client.GetScans(ctx, scanParams)
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

			ginkgo.By("waiting until scan state changes to failed with nothing to scan as state reason")
			params := models.GetScansParams{
				Filter: utils.PointerTo(fmt.Sprintf(
					"scanConfig/id eq '%s' and state eq '%s' and stateReason eq '%s'",
					*apiScanConfig.Id,
					models.ScanStateDone,
					models.ScanRelationshipStateReasonNothingToScan,
				)),
			}
			var scans *models.Scans
			gomega.Eventually(func() bool {
				scans, err = client.GetScans(ctx, params)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				return len(*scans.Items) == 1
			}, DefaultTimeout, time.Second).Should(gomega.BeTrue())
		})
	})

	ginkgo.Context("when a scan stops with timeout", func() {
		ginkgo.It("should detect failure reason successfully", func(ctx ginkgo.SpecContext) {
			ginkgo.By("applying a scan configuration with short timeout")
			apiScanConfig, err := client.PostScanConfig(
				ctx,
				GetCustomScanConfig(
					&FullScanFamiliesConfig,
					DefaultScope,
					2,
				))
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("updating scan configuration to run now")
			updateScanConfig := UpdateScanConfigToStartNow(apiScanConfig)
			err = client.PatchScanConfig(ctx, *apiScanConfig.Id, updateScanConfig)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("waiting until scan starts")
			scanParams := models.GetScansParams{
				Filter: utils.PointerTo(fmt.Sprintf(
					"scanConfig/id eq '%s'",
					*apiScanConfig.Id,
				)),
			}
			gomega.Eventually(func() bool {
				scans, err := client.GetScans(ctx, scanParams)
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

			ginkgo.By("waiting until scan state changes to failed with timed out as state reason")
			params := models.GetScansParams{
				Filter: utils.PointerTo(fmt.Sprintf(
					"scanConfig/id eq '%s' and state eq '%s' and stateReason eq '%s'",
					*apiScanConfig.Id,
					models.ScanStateFailed,
					models.ScanStateReasonTimedOut,
				)),
			}
			var scans *models.Scans
			gomega.Eventually(func() bool {
				scans, err = client.GetScans(ctx, params)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				return len(*scans.Items) == 1
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
