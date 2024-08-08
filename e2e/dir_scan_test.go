// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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
	"os"
	"path/filepath"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/testenv/types"
)

var _ = ginkgo.Describe("Running a SBOM and plugin scan", func() {
	reportFailedConfig := ReportFailedConfig{
		services: []string{"orchestrator"},
	}

	ginkgo.Context("which scans a directory", func() {
		ginkgo.It("should finish successfully", func(ctx ginkgo.SpecContext) {
			if cfg.TestEnvConfig.Platform != types.EnvironmentTypeDocker {
				ginkgo.Skip("skipping test because it's not running on docker")
			}

			var assets *apitypes.Assets
			var err error

			wd, err := os.Getwd()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			assetType := apitypes.AssetType{}
			err = assetType.FromDirInfo(apitypes.DirInfo{
				ObjectType: "DirInfo",
				DirName:    to.Ptr("test"),
				Location:   to.Ptr(filepath.Join(wd, "testdata")),
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			ginkgo.By("add dir asset")
			_, err = client.PostAsset(
				ctx,
				apitypes.Asset{
					AssetInfo: &assetType,
					FirstSeen: to.Ptr(time.Now()),
					// Set to future time so asset is not terminated by discoverer.
					LastSeen: to.Ptr(time.Now().Add(time.Minute * 10)),
				},
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			scope := "assetInfo/objectType eq 'DirInfo' and assetInfo/dirName eq 'test'"

			ginkgo.By("waiting until test asset is found")
			reportFailedConfig.objects = append(
				reportFailedConfig.objects,
				APIObject{"asset", scope},
			)
			assetsParams := apitypes.GetAssetsParams{
				Filter: to.Ptr(scope),
			}
			gomega.Eventually(func() bool {
				assets, err = client.GetAssets(ctx, assetsParams)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				return len(*assets.Items) == 1
			}, DefaultTimeout, DefaultPeriod).Should(gomega.BeTrue())

			ginkgo.By("applying a scan configuration")
			apiScanConfig, err := client.PostScanConfig(
				ctx,
				GetCustomScanConfig(
					&apitypes.ScanFamiliesConfig{
						Sbom: &apitypes.SBOMConfig{
							Enabled: to.Ptr(true),
						},
						Plugins: &apitypes.PluginsConfig{
							Enabled:      to.Ptr(true),
							ScannersList: to.Ptr([]string{"kics"}),
							ScannersConfig: &map[string]apitypes.PluginScannerConfig{
								"kics": {
									Config:    to.Ptr(""),
									ImageName: to.Ptr(cfg.TestEnvConfig.Docker.Images.PluginKics),
								},
							},
						},
					},
					scope,
					600*time.Second,
				),
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			reportFailedConfig.objects = append(
				reportFailedConfig.objects,
				APIObject{"scanConfig", fmt.Sprintf("id eq '%s'", *apiScanConfig.Id)},
			)

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

			ginkgo.By("waiting until scan state changes to done")
			scanParams = apitypes.GetScansParams{
				Filter: to.Ptr(fmt.Sprintf(
					"scanConfig/id eq '%s' and status/state eq '%s' and status/reason eq '%s'",
					*apiScanConfig.Id,
					apitypes.AssetScanStatusStateDone,
					apitypes.AssetScanStatusReasonSuccess,
				)),
			}
			gomega.Eventually(func() bool {
				scans, err = client.GetScans(ctx, scanParams)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				return len(*scans.Items) == 1
			}, DefaultTimeout, DefaultPeriod).Should(gomega.BeTrue())

			ginkgo.By("verifying that at least one package was found")
			gomega.Eventually(func() bool {
				totalPackages := (*scans.Items)[0].Summary.TotalPackages
				return *totalPackages > 0
			}, DefaultTimeout, DefaultPeriod).Should(gomega.BeTrue())

			ginkgo.By("verifying that at least one plugin finding was found")
			gomega.Eventually(func() bool {
				totalPlugins := (*scans.Items)[0].Summary.TotalPlugins
				return totalPlugins != nil && *totalPlugins > 0
			}, DefaultTimeout, DefaultPeriod).Should(gomega.BeTrue())
		})
	})

	ginkgo.AfterEach(func(ctx ginkgo.SpecContext) {
		if ginkgo.CurrentSpecReport().Failed() {
			reportFailedConfig.startTime = ginkgo.CurrentSpecReport().StartTime
			ReportFailed(ctx, testEnv, client, &reportFailedConfig)
		}
	})
})
