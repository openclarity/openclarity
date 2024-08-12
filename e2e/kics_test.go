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
	"context"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/openclarity/vmclarity/scanner/families"
	"github.com/openclarity/vmclarity/scanner/families/plugins"
	"github.com/openclarity/vmclarity/scanner/families/plugins/common"
	"github.com/openclarity/vmclarity/scanner/families/plugins/runner/config"
	"github.com/openclarity/vmclarity/scanner/families/types"
	"github.com/openclarity/vmclarity/scanner/utils"
)

const scannerPluginName = "kics"

type Notifier struct {
	Results []families.FamilyResult
}

func (n *Notifier) FamilyStarted(context.Context, types.FamilyType) error { return nil }

func (n *Notifier) FamilyFinished(_ context.Context, res families.FamilyResult) error {
	n.Results = append(n.Results, res)

	return nil
}

var _ = ginkgo.Describe("Running KICS scan", func() {
	ginkgo.Context("which scans an openapi.yaml file", func() {
		ginkgo.It("should finish successfully", func(ctx ginkgo.SpecContext) {
			if cfg.TestEnvConfig.Images.PluginKics == "" {
				ginkgo.Skip("KICS plugin image not provided")
			}

			input, err := filepath.Abs("./testdata")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			notifier := &Notifier{}

			errs := families.New(&families.Config{
				Plugins: plugins.Config{
					Enabled:      true,
					ScannersList: []string{scannerPluginName},
					Inputs: []types.Input{
						{
							Input:     input,
							InputType: string(utils.ROOTFS),
						},
					},
					ScannersConfig: &common.ScannersConfig{
						scannerPluginName: config.Config{
							Name:          scannerPluginName,
							ImageName:     cfg.TestEnvConfig.Images.PluginKics,
							InputDir:      "",
							ScannerConfig: "",
						},
					},
				},
			}).Run(ctx, notifier)
			gomega.Expect(errs).To(gomega.BeEmpty())

			gomega.Eventually(func() bool {
				if len(notifier.Results) != 1 {
					return false
				}

				results := notifier.Results[0].Result.(*plugins.Results)                             // nolint:forcetypeassert
				rawData := results.PluginOutputs[scannerPluginName].RawJSON.(map[string]interface{}) // nolint:forcetypeassert

				if rawData["total_counter"] != float64(23) {
					return false
				}

				if len(results.Findings) != 23 {
					return false
				}

				return true
			}, DefaultTimeout, DefaultPeriod).Should(gomega.BeTrue())
		})
	})
})
