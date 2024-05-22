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
	"io"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/testenv/kubernetes/asset"
	"github.com/openclarity/vmclarity/testenv/kubernetes/helm"
	"github.com/openclarity/vmclarity/testenv/kubernetes/kind"
	"github.com/openclarity/vmclarity/testenv/kubernetes/types"
	envtypes "github.com/openclarity/vmclarity/testenv/types"
	"github.com/openclarity/vmclarity/testenv/utils"
)

type ContextKeyType string

const KubeClientContextKey ContextKeyType = "KubeClient"

const (
	GatewayHostPort      = 30080
	ListOperationTimeout = 30
)

type KubernetesEnv struct {
	// Provider used for deploying Kubernetes cluster
	types.Provider
	// Installer used for deploying applications to Kubernetes cluster
	types.Installer
	// Config stores all the configuration for Kubernetes environment
	*Config

	// meta is a collection of metadata information. Currently, it is used for structured logging.
	meta map[string]interface{}
}

func (e *KubernetesEnv) SetUp(ctx context.Context) error {
	if err := e.Provider.SetUp(ctx); err != nil {
		return fmt.Errorf("failed to set up environment: %w", err)
	}

	if err := e.Install(ctx); err != nil {
		return fmt.Errorf("failed to install manifest(s): %w", err)
	}

	return nil
}

func (e *KubernetesEnv) ServicesReady(ctx context.Context) (bool, error) {
	logger := utils.GetLoggerFromContextOrDiscard(ctx).WithFields(e.meta)

	services, err := e.Services(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve list of services: %w", err)
	}

	var result bool
	for _, service := range services {
		logger.Debugf("checking service readiness. Service=%s State=%s", service.GetID(), service.GetState())

		switch service.GetState() {
		case envtypes.ServiceStateReady, envtypes.ServiceStateDegraded:
			result = true
		case envtypes.ServiceStateNotReady, envtypes.ServiceStateUnknown:
			fallthrough
		default:
			result = false
		}
	}

	return result, nil
}

func (e *KubernetesEnv) ServiceLogs(ctx context.Context, services []string, startTime time.Time, stdout, _ io.Writer) error {
	logger := utils.GetLoggerFromContextOrDiscard(ctx).WithFields(e.meta)

	kubeConfig, err := e.Provider.KubeConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get KubeConfig from cluster %s: %w", e.ProviderConfig.ClusterName, err)
	}

	client, err := NewClientFromKubeConfig([]byte(kubeConfig))
	if err != nil {
		return fmt.Errorf("failed to get KubeClient from Kubeconfig: %w", err)
	}

	labelSelectors := []string{
		fmt.Sprintf("app.kubernetes.io/instance in (%s)", strings.Join(e.Installer.Deployments(), ", ")),
		fmt.Sprintf("app.kubernetes.io/name in (%s)", strings.Join(services, ", ")),
	}
	labelSelector := strings.Join(labelSelectors, ",")

	pods, err := client.CoreV1().Pods(e.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return fmt.Errorf("failed to get Pods from cluster: %w", err)
	}

	for _, pod := range pods.Items {
		log := client.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
			SinceTime: &metav1.Time{
				Time: startTime,
			},
			Timestamps: true,
		})

		logData, err := log.Stream(ctx)
		if err != nil {
			logger.Errorf("failed to retrieve Pod logs. Pod=%s Namespace=%s: %v", pod.Name, pod.Namespace, err)
			continue
		}

		stream := func() error {
			defer func(r io.ReadCloser) {
				err = r.Close()
				if err != nil {
					logger.Errorf("failed to close Pod logs stream. Pod=%s Namespace=%s: %v", pod.Name, pod.Namespace, err)
				}
			}(logData)

			_, err := io.Copy(stdout, logData)
			return fmt.Errorf("failed to copy log data: %w", err)
		}

		if err := stream(); err != nil {
			logger.Errorf("failed to copy Pod logs stream. Pod=%s Namespace=%s: %v", pod.Name, pod.Namespace, err)
			continue
		}
	}

	return nil
}

//nolint:cyclop
func (e *KubernetesEnv) Services(ctx context.Context) (envtypes.Services, error) {
	logger := utils.GetLoggerFromContextOrDiscard(ctx).WithFields(e.meta)

	kubeConfig, err := e.Provider.KubeConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get KubeConfig from cluster %s: %w", e.ProviderConfig.ClusterName, err)
	}

	services := make(envtypes.Services, 0, 100) //nolint:mnd

	client, err := NewClientFromKubeConfig([]byte(kubeConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to get KubeClient from Kubeconfig: %w", err)
	}

	labelSelector := fmt.Sprintf("app.kubernetes.io/instance in (%s)", strings.Join(e.Installer.Deployments(), ", "))

	deployments, err := client.AppsV1().Deployments(e.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector:  labelSelector,
		TimeoutSeconds: to.Ptr[int64](ListOperationTimeout),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Deployments from cluster: %w", err)
	}

	for _, deployment := range deployments.Items {
		service := ServiceFromDeployment(&deployment)
		if service == nil {
			logger.Warnf("failed to get service from Deployment: %v", deployment)
			continue
		}

		services = append(services, service)
	}

	statefulSets, err := client.AppsV1().StatefulSets(e.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector:  labelSelector,
		TimeoutSeconds: to.Ptr[int64](ListOperationTimeout),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get StatefulSets from cluster: %w", err)
	}

	for _, statefulSet := range statefulSets.Items {
		service := ServiceFromStatefulSet(&statefulSet)
		if service == nil {
			logger.Warnf("failed to get service from StatefulSet: %v", statefulSet)
			continue
		}

		services = append(services, service)
	}

	daemonSets, err := client.AppsV1().DaemonSets(e.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector:  labelSelector,
		TimeoutSeconds: to.Ptr[int64](ListOperationTimeout),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get DaemonSets from cluster: %w", err)
	}

	for _, daemonSet := range daemonSets.Items {
		service := ServiceFromDaemonSet(&daemonSet)
		if service == nil {
			logger.Warnf("failed to get service for DaemonSet: %v", daemonSet)
			continue
		}

		services = append(services, service)
	}

	return services, nil
}

func (e *KubernetesEnv) Endpoints(_ context.Context) (*envtypes.Endpoints, error) {
	endpoints := new(envtypes.Endpoints)
	endpoints.SetAPI("http", "localhost", strconv.Itoa(GatewayHostPort), "/api")
	endpoints.SetUIBackend("http", "localhost", strconv.Itoa(GatewayHostPort), "/ui/api/")

	return endpoints, nil
}

func (e *KubernetesEnv) Context(ctx context.Context) (context.Context, error) {
	kubeConfig, err := e.Provider.KubeConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get KubeConfig from cluster %s: %w", e.ProviderConfig.ClusterName, err)
	}

	client, err := NewClientFromKubeConfig([]byte(kubeConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to get KubeClient for cluster %s: %w", e.ProviderConfig.ClusterName, err)
	}

	return context.WithValue(ctx, KubeClientContextKey, client), nil
}

func New(config *Config, opts ...ConfigOptFn) (*KubernetesEnv, error) {
	opts = append(opts, withResolvedKubeConfigPath())
	if err := applyConfigWithOpts(config, opts...); err != nil {
		return nil, fmt.Errorf("failed to apply options to ProviderConfig: %w", err)
	}

	var provider types.Provider
	var err error
	switch config.Provider {
	case types.ProviderTypeKind:
		images, err := config.Images.AsStringSlice()
		if err != nil {
			return nil, fmt.Errorf("failed to get container images: %w", err)
		}

		provider, err = kind.New(&config.ProviderConfig,
			kind.WithLoadingImages(images),
			kind.WithKindLogger(config.logger),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create provider %s: %w", config.Provider, err)
		}
	case types.ProviderTypeExternal:
		fallthrough
	default:
		return nil, fmt.Errorf("unsupported Kubernetes provider: %s", config.Provider)
	}

	var installer types.Installer
	switch config.Installer {
	case types.InstallerTypeHelm:
		valuesOpts := []helm.ValuesOpts{
			helm.WithContainerImages(&config.Images),
			helm.WithKubernetesProvider(),
			helm.WithGatewayNodePort(GatewayHostPort),
		}

		installer, err = helm.New(config.HelmConfig,
			helm.WithDefaultKubeConfigPath(config.KubeConfigPath),
			helm.WithValuesOpts(valuesOpts),
			helm.WithDefaultReleaseName(config.EnvName),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create %s installer: %w", config.Installer, err)
		}
	default:
		installer = &types.NoOpInstaller{}
	}

	if !config.SkipAssetInstall {
		assetInstaller, err := asset.NewAssetInstaller(&asset.Config{
			KubeConfigPath: config.KubeConfigPath,
			Namespace:      config.Namespace,
			Name:           config.EnvName + "-vmclarity-asset",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create Asset installer: %w", err)
		}

		installer = types.NewMultiInstaller(installer, assetInstaller)
	}

	return &KubernetesEnv{
		Provider:  provider,
		Installer: installer,
		Config:    config,
		meta: map[string]interface{}{
			"environment": "kubernetes",
			"name":        config.EnvName,
			"namespace":   config.Namespace,
		},
	}, nil
}
