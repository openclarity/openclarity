// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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

package config

import (
	"context"
	"fmt"

	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var defaultScannerJobTemplate = []byte(`apiVersion: batch/v1
kind: Job
metadata:
  labels:
    app: scanner
    sidecar.istio.io/inject: "false"
spec:
  backoffLimit: 0
  ttlSecondsAfterFinished: 300
  template:
    metadata:
     labels:
      app: scanner
      sidecar.istio.io/inject: "false"
    spec:
      restartPolicy: Never
      containers:
      - name: vulnerability-scanner
      image: TBD
      args:
      - scan
      securityContext:
        capabilities:
          drop:
          - all
        runAsNonRoot: true
        runAsGroup: 1001
        runAsUser: 1001
        privileged: false
        allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      resources:
      requests:
        memory: "50Mi"
        cpu: "50m"
      limits:
        memory: "1000Mi"
        cpu: "1000m"
`)

func loadScannerTemplate(clientset kubernetes.Interface, configMapName, configMapNamespace string) (*batchv1.Job, error) {
	var job batchv1.Job
	var scannerTemplate []byte

	if configMapName == "" {
		// Use default scanner job template from config map.
		scannerTemplate = defaultScannerJobTemplate
	} else {
		// Get scanner job template from config map.
		cm, err := clientset.CoreV1().ConfigMaps(configMapNamespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get scanner template config map: %v", err)
		}

		config, ok := cm.Data["config"]
		if !ok {
			return nil, fmt.Errorf("no scanner template config in configmap")
		}

		scannerTemplate = []byte(config)
	}

	log.Debugf("Using scannerTemplate:\n%+v", string(scannerTemplate))

	err := yaml.Unmarshal(scannerTemplate, &job)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal scanner template: %v", err)
	}

	return &job, nil
}
