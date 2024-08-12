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

package kubernetes

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/provider/kubernetes/discoverer"
	"github.com/openclarity/vmclarity/provider/kubernetes/estimator"
	"github.com/openclarity/vmclarity/provider/kubernetes/scanner"
)

var crDiscovererLabels = map[string]string{
	"app.kubernetes.io/component": "cr-discovery-server",
	"app.kubernetes.io/name":      "vmclarity",
}

var _ provider.Provider = &Provider{}

type Provider struct {
	*discoverer.Discoverer
	*scanner.Scanner
	*estimator.Estimator
}

func (p *Provider) Kind() apitypes.CloudProvider {
	return apitypes.Kubernetes
}

func New(_ context.Context) (*Provider, error) {
	config, err := NewConfig()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	var clientConfig *rest.Config
	if config.KubeConfig == "" {
		// If KubeConfig config option not set, assume we're running in-cluster.
		clientConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("unable to load in-cluster client configuration: %w", err)
		}
	} else {
		cc, err := clientcmd.LoadFromFile(config.KubeConfig)
		if err != nil {
			return nil, fmt.Errorf("unable to load kubeconfig from %s: %w", config.KubeConfig, err)
		}
		clientConfig, err = clientcmd.NewNonInteractiveClientConfig(*cc, "", &clientcmd.ConfigOverrides{}, nil).ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("unable to create client configuration from the provided kubeconfig file: %w", err)
		}
	}

	clientSet, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create kubernetes clientset: %w", err)
	}

	return &Provider{
		Discoverer: &discoverer.Discoverer{
			ClientSet:                          clientSet,
			ContainerRuntimeDiscoveryNamespace: config.ContainerRuntimeDiscoveryNamespace,
			CrDiscovererLabels:                 crDiscovererLabels,
		},
		Scanner: &scanner.Scanner{
			ClientSet:                          clientSet,
			ContainerRuntimeDiscoveryNamespace: config.ContainerRuntimeDiscoveryNamespace,
			ScannerNamespace:                   config.ScannerNamespace,
			CrDiscovererLabels:                 crDiscovererLabels,
		},
		Estimator: &estimator.Estimator{},
	}, nil
}
