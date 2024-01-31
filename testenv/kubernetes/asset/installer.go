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

package asset

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type AssetInstaller struct {
	config *Config
	opts   []DeploymentOptFn
}

func (i *AssetInstaller) Install(ctx context.Context) error {
	deployment, err := NewDeploymentFromConfig(i.config)
	if err != nil {
		return fmt.Errorf("invalid Deployment spec: %w", err)
	}

	if err = applyDeploymentOpts(deployment, i.opts...); err != nil {
		return fmt.Errorf("failed to apply options to Deployment: %w", err)
	}

	client, err := NewClientFromKubeConfig(i.config.KubeConfigPath)
	if err != nil {
		return fmt.Errorf("failed to get KubeClient: %w", err)
	}

	_, err = client.AppsV1().Deployments(i.config.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to apply asset deployment: %w", err)
	}

	return nil
}

func (i *AssetInstaller) Uninstall(ctx context.Context) error {
	client, err := NewClientFromKubeConfig(i.config.KubeConfigPath)
	if err != nil {
		return fmt.Errorf("failed to get KubeClient: %w", err)
	}

	err = client.AppsV1().Deployments(i.config.Namespace).Delete(ctx, i.config.Name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to apply asset deployment: %w", err)
	}

	return nil
}

func (i *AssetInstaller) Deployments() []string {
	return []string{
		i.config.Name,
	}
}

func NewAssetInstaller(config *Config, opts ...DeploymentOptFn) (*AssetInstaller, error) {
	return &AssetInstaller{
		config: config,
		opts:   opts,
	}, nil
}

func NewClientFromKubeConfig(path string) (kubernetes.Interface, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("invalid KubeConfig path: %w", err)
	}

	kubeConfig, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read KubeConfig: %w", err)
	}

	config, err := clientcmd.RESTConfigFromKubeConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get RESTConfig for KubeConfig: %w", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to get KubeClient from config: %w", err)
	}

	return client, nil
}
