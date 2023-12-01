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

package types

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Provider interface {
	// SetUp the test environment by installing the necessary components.
	// Returns error if it fails to set up the environment.
	SetUp(ctx context.Context) error
	// TearDown the test environment by uninstalling components installed via Setup.
	// Returns error if it fails to clean up the environment.
	TearDown(ctx context.Context) error
	// KubeConfig returns the KUBECONFIG for the cluster.
	KubeConfig(ctx context.Context) (string, error)
	// ExportKubeConfig exports the KUBECONFIG for the cluster, merging
	// it into the selected file, following the rules from
	// https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#config
	// where explicitPath is the --kubeconfig value.
	ExportKubeConfig(ctx context.Context, explicitPath string) error
	// ClusterName returns the name of the cluster the provider operates on.
	ClusterName() string
}

type ProviderConfig struct {
	ClusterName            string        `mapstructure:"cluster_name"`
	ClusterCreationTimeout time.Duration `mapstructure:"cluster_creation_timeout"`
	KubeConfigPath         string        `mapstructure:"kubeconfig"`
	KubernetesVersion      string        `mapstructure:"version"`
}

// DefaultKubeConfigPath returns the default path for KUBECONFIG: ~/.kube/config.
func DefaultKubeConfigPath() string {
	return filepath.Join(homedir.HomeDir(), clientcmd.RecommendedHomeDir, clientcmd.RecommendedFileName)
}

type ProviderType string

const (
	ProviderTypeKind     ProviderType = "kind"
	ProviderTypeExternal ProviderType = "external"
)

func (p *ProviderType) UnmarshalText(text []byte) error {
	var provider ProviderType

	switch strings.ToLower(string(text)) {
	case strings.ToLower(string(ProviderTypeKind)):
		provider = ProviderTypeKind
	case strings.ToLower(string(ProviderTypeExternal)):
		provider = ProviderTypeExternal
	default:
		return fmt.Errorf("failed to unmarshal text into Kubernetes Provider: %s", text)
	}

	*p = provider

	return nil
}
