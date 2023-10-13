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

var _ = ginkgo.Describe("Running a basic scan (only SBOM)", func() {
	reportFailedConfig := ReportFailedConfig{}
	var imageID string

	ginkgo.Context("which scans a docker container", func() {
		ginkgo.It("should finish successfully", func(ctx ginkgo.SpecContext) {
			var assets *models.Assets
			var err error

			ginkgo.By("waiting until test asset is found")
			reportFailedConfig.objects = append(
				reportFailedConfig.objects,
				APIObject{"asset", DefaultScope},
			)
			assetsParams := models.GetAssetsParams{
				Filter: utils.PointerTo(DefaultScope),
			}
			gomega.Eventually(func() bool {
				assets, err = client.GetAssets(ctx, assetsParams)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				return len(*assets.Items) == 1
			}, DefaultTimeout, time.Second).Should(gomega.BeTrue())

			containerInfo, err := (*assets.Items)[0].AssetInfo.AsContainerInfo()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			imageID = containerInfo.Image.ImageID

			RunSuccessfulScan(ctx, &reportFailedConfig, DefaultScope)
		})
	})

	reportFailedConfig = ReportFailedConfig{}

	ginkgo.Context("which scans a docker image", func() {
		ginkgo.It("should finish successfully", func(ctx ginkgo.SpecContext) {
			ginkgo.By("waiting until test asset is found")
			filter := fmt.Sprintf("assetInfo/objectType eq 'ContainerImageInfo' and assetInfo/imageID eq '%s'", imageID)
			reportFailedConfig.objects = append(
				reportFailedConfig.objects,
				APIObject{"asset", filter},
			)
			gomega.Eventually(func() bool {
				assets, err := client.GetAssets(ctx, models.GetAssetsParams{
					Filter: utils.PointerTo(filter),
				})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				return len(*assets.Items) == 1
			}, DefaultTimeout, time.Second).Should(gomega.BeTrue())

			RunSuccessfulScan(ctx, &reportFailedConfig, filter)
		})
	})

	ginkgo.AfterEach(func(ctx ginkgo.SpecContext) {
		if ginkgo.CurrentSpecReport().Failed() {
			reportFailedConfig.startTime = ginkgo.CurrentSpecReport().StartTime
			ReportFailed(ctx, testEnv, client, &reportFailedConfig)
		}
	})
})

// nolint:gomnd
func RunSuccessfulScan(ctx ginkgo.SpecContext, report *ReportFailedConfig, filter string) {
	ginkgo.By("applying a scan configuration")
	apiScanConfig, err := client.PostScanConfig(
		ctx,
		GetCustomScanConfig(
			&models.ScanFamiliesConfig{
				Sbom: &models.SBOMConfig{
					Enabled: utils.PointerTo(true),
				},
			},
			filter,
			600,
		))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	report.objects = append(
		report.objects,
		APIObject{"scanConfig", fmt.Sprintf("id eq '%s'", *apiScanConfig.Id)},
	)

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
			report.objects = append(
				report.objects,
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
	}, time.Second*120, time.Second).Should(gomega.BeTrue())
}
