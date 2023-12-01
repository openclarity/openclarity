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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	envtypes "github.com/openclarity/vmclarity/e2e/testenv/types"
	sharedutils "github.com/openclarity/vmclarity/pkg/shared/utils"
)

func NewDeploymentFromConfig(config *Config) (*appsv1.Deployment, error) {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.Name,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "vmclarity-asset",
				"app.kubernetes.io/instance":   config.Name,
				"app.kubernetes.io/component":  "asset",
				"app.kubernetes.io/managed-by": "testenv",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: sharedutils.PointerTo[int32](1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name":      "vmclarity-asset",
					"app.kubernetes.io/instance":  config.Name,
					"app.kubernetes.io/component": "asset",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name":       "vmclarity-asset",
						"app.kubernetes.io/instance":   config.Name,
						"app.kubernetes.io/component":  "asset",
						"app.kubernetes.io/managed-by": "testenv",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "alpine",
							Image:   "alpine:3.18.2",
							Command: []string{"sleep", "infinity"},
						},
					},
				},
			},
		},
	}, nil
}

type DeploymentOptFn = func(d *appsv1.Deployment) error

var applyDeploymentOpts = envtypes.WithOpts[appsv1.Deployment, DeploymentOptFn]

func WithReplicas(r int32) DeploymentOptFn {
	return func(d *appsv1.Deployment) error {
		if d == nil {
			return nil
		}

		d.Spec.Replicas = sharedutils.PointerTo(r)

		return nil
	}
}
