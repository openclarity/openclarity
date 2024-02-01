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
	"errors"
	"fmt"
	"net"

	"gopkg.in/yaml.v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/pkg/containerruntimediscovery"
	"github.com/openclarity/vmclarity/pkg/orchestrator/provider"
	"github.com/openclarity/vmclarity/pkg/shared/families"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

type Provider struct {
	clientSet kubernetes.Interface
	config    *Config
}

var _ provider.Provider = &Provider{}

func New(_ context.Context) (provider.Provider, error) {
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
		clientSet: clientSet,
		config:    config,
	}, nil
}

func (p *Provider) Kind() types.CloudProvider {
	return types.Kubernetes
}

func (p *Provider) Estimate(_ context.Context, _ types.AssetScanStats, _ *types.Asset, _ *types.AssetScanTemplate) (*types.Estimation, error) {
	return &types.Estimation{}, provider.FatalErrorf("Not Implemented")
}

func (p *Provider) DiscoverAssets(ctx context.Context) provider.AssetDiscoverer {
	assetDiscoverer := provider.NewSimpleAssetDiscoverer()

	go func() {
		defer close(assetDiscoverer.OutputChan)

		discoverers, err := p.clientSet.CoreV1().Pods(p.config.ContainerRuntimeDiscoveryNamespace).List(ctx, metav1.ListOptions{
			LabelSelector: labels.Set(crDiscovererLabels).String(),
		})
		if err != nil {
			assetDiscoverer.Error = fmt.Errorf("unable to list discoverers: %w", err)
			return
		}

		var errs []error

		err = p.discoverImages(ctx, assetDiscoverer.OutputChan, discoverers.Items)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to discover images: %w", err))
		}

		err = p.discoverContainers(ctx, assetDiscoverer.OutputChan, discoverers.Items)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to discover containers: %w", err))
		}

		assetDiscoverer.Error = errors.Join(errs...)
	}()

	return assetDiscoverer
}

// nolint: cyclop
func (p *Provider) RunAssetScan(ctx context.Context, config *provider.ScanJobConfig) error {
	discoverers, err := p.clientSet.CoreV1().Pods(p.config.ContainerRuntimeDiscoveryNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Set(crDiscovererLabels).String(),
	})
	if err != nil {
		return fmt.Errorf("unable to list discoverers: %w", err)
	}

	objectType, err := config.AssetInfo.ValueByDiscriminator()
	if err != nil {
		return fmt.Errorf("failed to get asset object type: %w", err)
	}

	clients := []*containerruntimediscovery.Client{}
	for _, discoverer := range discoverers.Items {
		discovererEndpoint := net.JoinHostPort(discoverer.Status.PodIP, "8080")
		clients = append(clients, containerruntimediscovery.NewClient(discovererEndpoint))
	}

	switch value := objectType.(type) {
	case types.ContainerImageInfo:
		var pickedClient *containerruntimediscovery.Client
		for _, client := range clients {
			_, err := client.GetImage(ctx, value.ImageID)
			if err == nil {
				pickedClient = client
				break
			}
		}
		if pickedClient == nil {
			return fmt.Errorf("unable to find image ID %s in any discoverer", value.ImageID)
		}

		err := p.runScannerJob(ctx, config, pickedClient.ExportImageURL(ctx, value.ImageID))
		if err != nil {
			// TODO(sambetts) Make runScannerJob idempotent and
			// change this to a normal Errorf.
			return provider.FatalErrorf("unable to run scanner job: %w", err)
		}
	case types.ContainerInfo:
		var pickedClient *containerruntimediscovery.Client
		for _, client := range clients {
			_, err := client.GetContainer(ctx, value.ContainerID)
			if err == nil {
				pickedClient = client
				break
			}
		}
		if pickedClient == nil {
			return fmt.Errorf("unable to find container ID %s in any discoverer", value.ContainerID)
		}

		err := p.runScannerJob(ctx, config, pickedClient.ExportContainerURL(ctx, value.ContainerID))
		if err != nil {
			// TODO(sambetts) Make runScannerJob idempotent and
			// change this to a normal Errorf.
			return provider.FatalErrorf("unable to run scanner job: %w", err)
		}
	default:
		return provider.FatalErrorf("failed to scan asset object type %T: Not implemented", value)
	}

	return nil
}

// mountPointPath defines the location in the container where assets will be mounted.
var (
	mountPointPath  = "/mnt/snapshot"
	archiveLocation = mountPointPath + "/image.tar"
)

func (p *Provider) generateScanConfig(config *provider.ScanJobConfig) ([]byte, error) {
	// Add volume mount point to family configuration
	familiesConfig := families.Config{}
	err := yaml.Unmarshal([]byte(config.ScannerCLIConfig), &familiesConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal family scan configuration: %w", err)
	}

	families.SetOciArchiveForFamiliesInput([]string{archiveLocation}, &familiesConfig)
	familiesConfigByte, err := yaml.Marshal(familiesConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal family scan configuration: %w", err)
	}

	return familiesConfigByte, nil
}

func (p *Provider) runScannerJob(ctx context.Context, config *provider.ScanJobConfig, sourceURL string) error {
	configBytes, err := p.generateScanConfig(config)
	if err != nil {
		return fmt.Errorf("failed to generate scanner config yaml: %w", err)
	}

	jobName := fmt.Sprintf("vmclarity-scan-%s", config.AssetScanID)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: p.config.ScannerNamespace,
		},
		BinaryData: map[string][]byte{
			"config.yaml": configBytes,
		},
	}
	_, err = p.clientSet.CoreV1().ConfigMaps(p.config.ScannerNamespace).Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create config map: %w", err)
	}

	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: p.config.ScannerNamespace,
		},
		Spec: batchv1.JobSpec{
			// TTLSecondsAfterFinished: utils.PointerTo(int32(120)),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:    "download-asset",
							Image:   "yauritux/busybox-curl:latest",
							Command: []string{"/bin/sh", "-c"},
							Args:    []string{fmt.Sprintf("curl -v %s -o %s", sourceURL, archiveLocation)},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "asset-data",
									MountPath: mountPointPath,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            jobName,
							Image:           config.ScannerImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Args: []string{
								"scan",
								"--config",
								"/etc/vmclarity/config.yaml",
								"--server",
								config.VMClarityAddress,
								"--asset-scan-id",
								config.AssetScanID,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config",
									ReadOnly:  true,
									MountPath: "/etc/vmclarity",
								},
								{
									Name:      "asset-data",
									MountPath: mountPointPath,
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
					Volumes: []corev1.Volume{
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: jobName,
									},
								},
							},
						},
						{
							Name: "asset-data",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
			BackoffLimit: utils.PointerTo[int32](0),
		},
	}

	_, err = p.clientSet.BatchV1().Jobs(p.config.ScannerNamespace).Create(ctx, jobSpec, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("unable to create job: %w", err)
	}

	return nil
}

func (p *Provider) RemoveAssetScan(ctx context.Context, config *provider.ScanJobConfig) error {
	jobName := fmt.Sprintf("vmclarity-scan-%s", config.AssetScanID)

	err := p.clientSet.BatchV1().Jobs(p.config.ScannerNamespace).Delete(ctx, jobName, metav1.DeleteOptions{
		PropagationPolicy: utils.PointerTo(metav1.DeletePropagationBackground),
	})
	if err != nil && !k8sErrors.IsNotFound(err) {
		return fmt.Errorf("unable to delete job: %w", err)
	}

	err = p.clientSet.CoreV1().ConfigMaps(p.config.ScannerNamespace).Delete(ctx, jobName, metav1.DeleteOptions{})
	if err != nil && !k8sErrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete config map: %w", err)
	}

	return nil
}
