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
	"fmt"
	"os"
	"strconv"

	"github.com/ghodss/yaml"
	"helm.sh/helm/v3/pkg/chartutil"

	envtypes "github.com/openclarity/vmclarity/testenv/types"
)

type (
	ValuesOpts func(values *Values) error
	Values     = chartutil.Values
)

var applyValuesWithOpts = envtypes.WithOpts[Values, ValuesOpts]

func NewEmptyValues() Values {
	return make(Values)
}

func WithContainerImages(images *ContainerImages) ValuesOpts {
	return func(values *Values) error {
		v := Values{
			"apiserver": map[string]interface{}{
				"image": map[string]interface{}{
					"registry":   images.APIServer.Domain(),
					"repository": images.APIServer.Path(),
					"tag":        images.APIServer.Tag(),
					"digest":     images.APIServer.Digest(),
				},
			},
			"orchestrator": map[string]interface{}{
				"image": map[string]interface{}{
					"registry":   images.Orchestrator.Domain(),
					"repository": images.Orchestrator.Path(),
					"tag":        images.Orchestrator.Tag(),
					"digest":     images.Orchestrator.Digest(),
				},
				"scannerImage": map[string]interface{}{
					"registry":   images.Scanner.Domain(),
					"repository": images.Scanner.Path(),
					"tag":        images.Scanner.Tag(),
					"digest":     images.Scanner.Digest(),
				},
			},
			"ui": map[string]interface{}{
				"image": map[string]interface{}{
					"registry":   images.UI.Domain(),
					"repository": images.UI.Path(),
					"tag":        images.UI.Tag(),
					"digest":     images.UI.Digest(),
				},
			},
			"uibackend": map[string]interface{}{
				"image": map[string]interface{}{
					"registry":   images.UIBackend.Domain(),
					"repository": images.UIBackend.Path(),
					"tag":        images.UIBackend.Tag(),
					"digest":     images.UIBackend.Digest(),
				},
			},
			"crDiscoveryServer": map[string]interface{}{
				"image": map[string]interface{}{
					"registry":   images.CRDiscoveryServer.Domain(),
					"repository": images.CRDiscoveryServer.Path(),
					"tag":        images.CRDiscoveryServer.Tag(),
					"digest":     images.CRDiscoveryServer.Digest(),
				},
			},
		}

		*values = chartutil.MergeTables(v, *values)

		return nil
	}
}

func WithKubernetesProvider() ValuesOpts {
	return func(values *Values) error {
		v := Values{
			"orchestrator": map[string]interface{}{
				"serviceAccount": map[string]interface{}{
					"automountServiceAccountToken": "true",
				},
				"provider": "kubernetes",
			},
		}

		*values = chartutil.MergeTables(v, *values)

		return nil
	}
}

func WithGatewayNodePort(port int) ValuesOpts {
	return func(values *Values) error {
		v := Values{
			"gateway": map[string]interface{}{
				"service": map[string]interface{}{
					"type": "NodePort",
					"nodePorts": map[string]interface{}{
						"http": strconv.Itoa(port),
					},
				},
			},
		}

		*values = chartutil.MergeTables(v, *values)

		return nil
	}
}

func WithValues(v Values) ValuesOpts {
	return func(values *Values) error {
		*values = chartutil.MergeTables(v, *values)

		return nil
	}
}

func WithValuesFile(path string) ValuesOpts {
	return func(values *Values) error {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to open values file. Values=%s: %w", path, err)
		}

		v := Values{}
		if err = yaml.Unmarshal(data, v); err != nil {
			return fmt.Errorf("failed to unmarshal values file. Values=%s: %w", path, err)
		}
		*values = chartutil.MergeTables(v, *values)

		return nil
	}
}
