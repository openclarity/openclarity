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

package scanner

import (
	"context"
	"fmt"
	"net"

	"gopkg.in/yaml.v3"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	apitypes "github.com/openclarity/vmclarity/api/types"
	discoveryclient "github.com/openclarity/vmclarity/containerruntimediscovery/client"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/scanner/families"
)

var _ provider.Scanner = &Scanner{}

type Scanner struct {
	ClientSet                          kubernetes.Interface
	ContainerRuntimeDiscoveryNamespace string
	ScannerNamespace                   string
	CrDiscovererLabels                 map[string]string
}

func (s *Scanner) RunAssetScan(ctx context.Context, t *provider.ScanJobConfig) error {
	discoverers, err := s.ClientSet.CoreV1().Pods(s.ContainerRuntimeDiscoveryNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Set(s.CrDiscovererLabels).String(),
	})
	if err != nil {
		return fmt.Errorf("unable to list discoverers: %w", err)
	}

	objectType, err := t.AssetInfo.ValueByDiscriminator()
	if err != nil {
		return fmt.Errorf("failed to get asset object type: %w", err)
	}

	clients := []*discoveryclient.Client{}
	for _, discoverer := range discoverers.Items {
		discovererEndpoint := net.JoinHostPort(discoverer.Status.PodIP, "8080")
		clients = append(clients, discoveryclient.NewClient(discovererEndpoint))
	}

	switch value := objectType.(type) {
	case apitypes.ContainerImageInfo:
		var pickedClient *discoveryclient.Client
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

		err := s.runScannerJob(ctx, t, pickedClient.ExportImageURL(ctx, value.ImageID))
		if err != nil {
			// TODO(sambetts) Make runScannerJob idempotent and
			// change this to a normal Errorf.
			return provider.FatalErrorf("unable to run scanner job: %w", err)
		}
	case apitypes.ContainerInfo:
		var pickedClient *discoveryclient.Client
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

		err := s.runScannerJob(ctx, t, pickedClient.ExportContainerURL(ctx, value.ContainerID))
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

func (s *Scanner) RemoveAssetScan(ctx context.Context, t *provider.ScanJobConfig) error {
	jobName := "vmclarity-scan-" + t.AssetScanID

	err := s.ClientSet.BatchV1().Jobs(s.ScannerNamespace).Delete(ctx, jobName, metav1.DeleteOptions{
		PropagationPolicy: to.Ptr(metav1.DeletePropagationBackground),
	})
	if err != nil && !k8sErrors.IsNotFound(err) {
		return fmt.Errorf("unable to delete job: %w", err)
	}

	err = s.ClientSet.CoreV1().ConfigMaps(s.ScannerNamespace).Delete(ctx, jobName, metav1.DeleteOptions{})
	if err != nil && !k8sErrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete config map: %w", err)
	}

	return nil
}

// mountPointPath defines the location in the container where assets will be mounted.
var (
	mountPointPath  = "/mnt/snapshot"
	archiveLocation = mountPointPath + "/image.tar"
)

func (s *Scanner) generateScanConfig(config *provider.ScanJobConfig) ([]byte, error) {
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

func (s *Scanner) runScannerJob(ctx context.Context, config *provider.ScanJobConfig, sourceURL string) error {
	configBytes, err := s.generateScanConfig(config)
	if err != nil {
		return fmt.Errorf("failed to generate scanner config yaml: %w", err)
	}

	jobName := "vmclarity-scan-" + config.AssetScanID

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: s.ScannerNamespace,
		},
		BinaryData: map[string][]byte{
			"config.yaml": configBytes,
		},
	}
	_, err = s.ClientSet.CoreV1().ConfigMaps(s.ScannerNamespace).Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create config map: %w", err)
	}

	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: s.ScannerNamespace,
		},
		Spec: batchv1.JobSpec{
			// TTLSecondsAfterFinished: to.Ptr(int32(120)),
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
			BackoffLimit: to.Ptr[int32](0),
		},
	}

	_, err = s.ClientSet.BatchV1().Jobs(s.ScannerNamespace).Create(ctx, jobSpec, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("unable to create job: %w", err)
	}

	return nil
}
