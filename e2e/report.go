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
	"encoding/json"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/formatter"
	"github.com/onsi/gomega"

	apiclient "github.com/openclarity/vmclarity/api/client"
	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/testenv/types"
)

type APIObject struct {
	objectType string
	filter     string
}

type ReportFailedConfig struct {
	startTime time.Time
	// if not empty, print logs for services in slice. if empty, print logs for all services.
	services []string
	// if true, print all API objects.
	allAPIObjects bool
	// if not empty, print objects in slice.
	objects []APIObject
}

// ReportFailed gathers relevant API data and docker service logs for debugging purposes.
func ReportFailed(ctx ginkgo.SpecContext, testEnv types.Environment, client *apiclient.Client, config *ReportFailedConfig) {
	ginkgo.GinkgoWriter.Println("------------------------------")

	DumpAPIData(ctx, client, config)
	DumpServiceLogs(ctx, testEnv, config)

	ginkgo.GinkgoWriter.Println("------------------------------")
}

// nolint:cyclop
// DumpAPIData prints API objects filtered using test parameters (e.g. assets filtered by scope, scan configs filtered by id).
// If filter not provided, no objects are printed.
func DumpAPIData(ctx ginkgo.SpecContext, client *apiclient.Client, config *ReportFailedConfig) {
	ginkgo.GinkgoWriter.Println(formatter.F("{{red}}[FAILED] Report API Data:{{/}}"))

	if config.allAPIObjects {
		config.objects = append(config.objects, APIObject{"asset", ""}, APIObject{"scanConfigs", ""}, APIObject{"scans", ""})
	}

	for _, object := range config.objects {
		switch object.objectType {
		case "asset":
			var params apitypes.GetAssetsParams
			if object.filter == "" {
				params = apitypes.GetAssetsParams{}
			} else {
				params = apitypes.GetAssetsParams{Filter: to.Ptr(object.filter)}
			}
			assets, err := client.GetAssets(ctx, params)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			buf, err := json.MarshalIndent(*assets.Items, "", "\t")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			ginkgo.GinkgoWriter.Printf("Asset: %s\n", string(buf))

		case "scanConfig":
			var params apitypes.GetScanConfigsParams
			if object.filter == "" {
				params = apitypes.GetScanConfigsParams{}
			} else {
				params = apitypes.GetScanConfigsParams{Filter: to.Ptr(object.filter)}
			}
			scanConfigs, err := client.GetScanConfigs(ctx, params)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			buf, err := json.MarshalIndent(*scanConfigs.Items, "", "\t")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			ginkgo.GinkgoWriter.Printf("Scan Config: %s\n", string(buf))

		case "scan":
			var params apitypes.GetScansParams
			if object.filter == "" {
				params = apitypes.GetScansParams{}
			} else {
				params = apitypes.GetScansParams{Filter: to.Ptr(object.filter)}
			}
			scans, err := client.GetScans(ctx, params)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			buf, err := json.MarshalIndent(*scans.Items, "", "\t")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			ginkgo.GinkgoWriter.Printf("Scan: %s\n", string(buf))

		case "assetScan":
			var params apitypes.GetAssetScansParams
			if object.filter == "" {
				params = apitypes.GetAssetScansParams{}
			} else {
				params = apitypes.GetAssetScansParams{Filter: to.Ptr(object.filter)}
			}
			assetScans, err := client.GetAssetScans(ctx, params)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			buf, err := json.MarshalIndent(*assetScans.Items, "", "\t")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			ginkgo.GinkgoWriter.Printf("Asset Scan: %s\n", string(buf))
		}
	}
}

// DumpServiceLogs prints service logs since the test started until it failed.
func DumpServiceLogs(ctx ginkgo.SpecContext, testEnv types.Environment, config *ReportFailedConfig) {
	ginkgo.GinkgoWriter.Println(formatter.F("{{red}}[FAILED] Report Service Logs:{{/}}"))

	if len(config.services) == 0 {
		services, err := testEnv.Services(ctx)
		if err != nil {
			ginkgo.GinkgoWriter.Println(formatter.F("{{red}}failed to retrieve list of services{{/}}"))
			return
		}

		config.services = services.IDs()
	}

	err := testEnv.ServiceLogs(ctx, config.services, config.startTime, formatter.ColorableStdOut, formatter.ColorableStdErr)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}
