// Copyright © 2021 Cisco Systems, Inc. and its affiliates.
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

package common

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/third_party/helm"
)

// EXPORTED:

func InstallTest(ns string) error {
	cmd := exec.Command("kubectl", "-n", ns, "apply", "-f", "test.yaml")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute command. %v, %s", err, out)
	}
	return nil
}

func LoadDockerImagesToCluster(cluster, tag string) error {
	if err := LoadDockerImageToCluster(cluster, fmt.Sprintf("ghcr.io/openclarity/kubeclarity:%v", tag)); err != nil {
		return fmt.Errorf("failed to load docker image to cluster: %v", err)
	}
	if err := LoadDockerImageToCluster(cluster, fmt.Sprintf("ghcr.io/openclarity/kubeclarity-sbom-db:%v", tag)); err != nil {
		return fmt.Errorf("failed to load docker image to cluster: %v", err)
	}
	// ghcr.io/openclarity/kubeclarity-cli
	if err := LoadDockerImageToCluster(cluster, fmt.Sprintf("ghcr.io/openclarity/kubeclarity-runtime-k8s-scanner:%v", tag)); err != nil {
		return fmt.Errorf("failed to load docker image to cluster: %v", err)
	}
	if err := LoadDockerImageToCluster(cluster, fmt.Sprintf("ghcr.io/openclarity/kubeclarity-cis-docker-benchmark-scanner:%v", tag)); err != nil {
		return fmt.Errorf("failed to load docker image to cluster: %v", err)
	}

	return nil
}

func LoadDockerImageToCluster(cluster, image string) error {
	cmd := exec.Command("kind", "load", "docker-image", image, "--name", cluster)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute command. %v, %s", err, out)
	}
	return nil
}

var curDir, _ = os.Getwd()
var chartPath = filepath.Join(curDir, "../charts/kubeclarity")

func GetCurrentDir() string {
	return curDir
}

func InstallKubeClarity(manager *helm.Manager, args string) error {
	if err := manager.RunInstall(helm.WithName(KubeClarityHelmReleaseName),
		helm.WithVersion("v1.1"),
		helm.WithNamespace(KubeClarityNamespace),
		helm.WithChart(chartPath),
		helm.WithArgs(args)); err != nil {
		return fmt.Errorf("failed to run helm install command with args: %v. %v", args, err)
	}
	return nil
}

func UninstallKubeClarity() error {
	// uninstall kubeclarity
	cmd := exec.Command("helm", "uninstall", KubeClarityHelmReleaseName, "-n", KubeClarityNamespace)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to uninstall kubeclarity: %s. %v", out, err)
	}

	cmd2 := exec.Command("kubectl", "-n", KubeClarityNamespace, "delete", "pvc", "data-kubeclarity-kubeclarity-postgresql-0")
	out, err = cmd2.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete pvc: %s. %v", out, err)
	}

	cmd3 := exec.Command("kubectl", "delete", "ns", KubeClarityNamespace)
	out, err = cmd3.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete kubeclarity namespace: %s. %v", out, err)
	}

	return nil
}

func PortForwardToKubeClarity(stopCh chan struct{}) {
	go func() {
		err := portForward("service", KubeClarityNamespace, KubeClarityServiceName, KubeClarityPortForwardHostPort, KubeClarityPortForwardTargetPort, stopCh)
		if err != nil {
			println("port forward failed. %v", err)
			return
		}
	}()
	time.Sleep(3 * time.Second)
}

func StringPtr(val string) *string {
	ret := val
	return &ret
}

//TODO use https://github.com/kubernetes-sigs/e2e-framework/tree/main/examples/wait_for_resources
func WaitForPodRunning(client klient.Client, ns string, labelSelector string) error {
	podList := v1.PodList{}
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	timeout := time.NewTimer(3 * time.Minute)
	for {
		select {
		case <-timeout.C:
			return fmt.Errorf("timeout reached")
		case <-ticker.C:
			if err := client.Resources(ns).List(context.TODO(), &podList, func(lo *v12.ListOptions) {
				lo.LabelSelector = labelSelector
			}); err != nil {
				return fmt.Errorf("failed to get pod in namespace: %v and labels: %v. %v", ns, labelSelector, err)
			}
			pod := podList.Items[0]
			if pod.Status.Phase == v1.PodRunning {
				return nil
			}
		}
	}
}

func CreateNamespace(client klient.Client ,name string) error {
	var ns = v1.Namespace{
		TypeMeta:   v12.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: v12.ObjectMeta{
			Name:                       name,
		},
	}
	if err := client.Resources(name).Create(context.TODO(), &ns); err != nil {
		return err
	}
	return nil
}

// NON EXPORTED:

func portForward(kind, namespace, name, hostPort, targetPort string, stopCh chan struct{}) error {
	cmd := exec.Command("kubectl", "port-forward", "-n", namespace,
		fmt.Sprintf("%s/%s", kind, name), fmt.Sprintf("%s:%s", hostPort, targetPort))

	processExitedCh := make(chan struct{})
	var output []byte
	var err error
	go func() {
		output, err = cmd.CombinedOutput()
		if err != nil {
			processExitedCh <- struct{}{}
		}
	}()

	select {
	case <-stopCh:
		if err := cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %v", err)
		}
		return nil
	case <-processExitedCh:
		return fmt.Errorf("port-forward process exited unexpectedly, output: %s", output)
	}
}
