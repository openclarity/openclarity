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

package helm

import (
	"testing"

	. "github.com/onsi/gomega"
)

var HelmChartDir = "../../../installation/kubernetes/helm/vmclarity"

func TestLoader(t *testing.T) {
	tests := []struct {
		Name    string
		EnvVars map[string]string

		HemConfig *Config

		ExpectedChartName string
	}{
		{
			Name:    "Load chart from manifest FS",
			EnvVars: map[string]string{},
			HemConfig: &Config{
				Namespace:      "default",
				ReleaseName:    "testenv-k8s-test",
				ChartPath:      "",
				StorageDriver:  "secret",
				KubeConfigPath: "kubeconfig",
			},
			ExpectedChartName: "vmclarity",
		},
		{
			Name:    "Load chart from dir",
			EnvVars: map[string]string{},
			HemConfig: &Config{
				Namespace:      "default",
				ReleaseName:    "testenv-k8s-test",
				ChartPath:      HelmChartDir,
				StorageDriver:  "secret",
				KubeConfigPath: "kubeconfig",
			},
			ExpectedChartName: "vmclarity",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			chart, err := LoadPathOrEmbedded(test.HemConfig.ChartPath)
			g.Expect(err).Should(Not(HaveOccurred()))

			name := chart.Name()
			g.Expect(name).Should(Equal(test.ExpectedChartName))
		})
	}
}
