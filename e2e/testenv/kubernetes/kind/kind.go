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

package kind

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	kinddefaults "sigs.k8s.io/kind/pkg/apis/config/defaults"
	kindcluster "sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/log"

	k8senvtypes "github.com/openclarity/vmclarity/e2e/testenv/kubernetes/types"
	envtypes "github.com/openclarity/vmclarity/e2e/testenv/types"
)

const (
	ImageLoadTimeout = 30 * time.Minute
)

//go:embed kind-cluster.yml
var ClusterConfigFile []byte

type Provider struct {
	config         *k8senvtypes.ProviderConfig
	kind           *kindcluster.Provider
	kindOpts       []kindcluster.ProviderOption
	kubeConfigPath string
	loader         ImageLoader

	generatedClusterName string
}

func (p *Provider) SetUp(ctx context.Context) error {
	// List available clusters
	clusters, err := p.kind.List()
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	// Skip creating cluster if it is already created
	for _, clusterName := range clusters {
		if clusterName == p.generatedClusterName {
			return nil
		}
	}

	nodeImage, ok := NodeImagesByVersion[p.config.KubernetesVersion]
	if !ok {
		nodeImage = kinddefaults.Image
	}

	err = p.kind.Create(
		p.generatedClusterName,
		kindcluster.CreateWithRawConfig(ClusterConfigFile),
		kindcluster.CreateWithNodeImage(nodeImage),
		kindcluster.CreateWithWaitForReady(p.config.ClusterCreationTimeout),
		kindcluster.CreateWithKubeconfigPath(p.kubeConfigPath),
		kindcluster.CreateWithDisplaySalutation(false),
		kindcluster.CreateWithDisplayUsage(false),
	)
	if err != nil {
		return fmt.Errorf("failed to create cluster with name %s: %w", p.config.ClusterName, err)
	}

	if p.loader == nil {
		return nil
	}

	nodes, err := p.kind.ListNodes(p.generatedClusterName)
	if err != nil {
		return fmt.Errorf("failed to list nodes in cluster: %w", err)
	}

	if err = p.loader.Load(ctx, nodes); err != nil {
		return fmt.Errorf("failed to load images to cluster: %w", err)
	}

	return nil
}

func (p *Provider) TearDown(_ context.Context) error {
	// List available clusters
	clusters, err := p.kind.List()
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	deleteCluster := func() error {
		err = p.kind.Delete(
			p.generatedClusterName,
			p.kubeConfigPath,
		)
		if err != nil {
			return fmt.Errorf("failed to delete cluster with name %s: %w", p.generatedClusterName, err)
		}

		return nil
	}

	// Delete cluster if exists
	for _, clusterName := range clusters {
		if clusterName == p.generatedClusterName {
			if err = deleteCluster(); err != nil {
				return fmt.Errorf("failed to delete cluster: %w", err)
			}

			return nil
		}
	}

	return nil
}

func (p *Provider) KubeConfig(_ context.Context) (string, error) {
	kubeConfig, err := p.kind.KubeConfig(p.generatedClusterName, false)
	if err != nil {
		return "", fmt.Errorf("failed to get KubeConfig: %w", err)
	}

	return kubeConfig, nil
}

func (p *Provider) ExportKubeConfig(_ context.Context, explicitPath string) error {
	if err := p.kind.ExportKubeConfig(p.generatedClusterName, explicitPath, false); err != nil {
		return fmt.Errorf("failed to export KubeConfig for cluster %s to path %s: %w", p.generatedClusterName, p.kubeConfigPath, err)
	}

	return nil
}

func (p *Provider) ClusterName() string {
	return p.generatedClusterName
}

func New(config *k8senvtypes.ProviderConfig, opts ...ProviderOptFn) (*Provider, error) {
	provider := &Provider{
		config:               config,
		kubeConfigPath:       config.KubeConfigPath,
		kindOpts:             []kindcluster.ProviderOption{},
		generatedClusterName: generatedClusterName(config.ClusterName, ClusterConfigFile),
	}

	if err := applyProviderWithOpts(provider, opts...); err != nil {
		return nil, fmt.Errorf("failed to apply options to ProviderConfig: %w", err)
	}

	nodeProvider, err := kindcluster.DetectNodeProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to detect Kind node provider: %w", err)
	}
	provider.kindOpts = append(provider.kindOpts, nodeProvider)
	provider.kind = kindcluster.NewProvider(provider.kindOpts...)

	return provider, nil
}

// ProviderOptFn defines transformer function for Provider.
type ProviderOptFn func(*Provider) error

var applyProviderWithOpts = envtypes.WithOpts[Provider, ProviderOptFn]

func WithKindLogger(logger *logrus.Entry) ProviderOptFn {
	return func(p *Provider) error {
		opt := kindcluster.ProviderWithLogger(log.NoopLogger{})
		if logger != nil {
			opt = kindcluster.ProviderWithLogger(NewLogger(logger))
		}

		if p.kindOpts == nil {
			p.kindOpts = []kindcluster.ProviderOption{
				opt,
			}
		} else {
			p.kindOpts = append(p.kindOpts, opt)
		}

		return nil
	}
}

func WithLoadingImages(images []string) ProviderOptFn {
	return func(p *Provider) error {
		loader, err := NewDockerImageLoader(images, ImageLoadTimeout)
		if err != nil {
			return fmt.Errorf("failed to initialize container image loader: %w", err)
		}
		p.loader = loader

		return nil
	}
}

// generatedClusterName returns a cluster name from name and the content has of the cluster configuration.
func generatedClusterName(name string, clusterConfig []byte) string {
	h := sha256.New()
	h.Write(clusterConfig)

	return fmt.Sprintf("%s-%.4x", name, h.Sum(nil))
}
