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
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	envtypes "github.com/openclarity/vmclarity/testenv/types"
)

// https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
const (
	AppComponentLabel = "app.kubernetes.io/component"
	AppNameLabel      = "app.kubernetes.io/name"
)

func ServiceFromMeta(m *metav1.ObjectMeta) *Service {
	if m == nil {
		return &Service{}
	}

	s := &Service{
		ID:        m.Name,
		Namespace: m.Namespace,
	}

	if application, ok := m.Labels[AppNameLabel]; ok {
		s.Application = application
	}

	if component, ok := m.Labels[AppComponentLabel]; ok {
		s.Component = component
	}

	return s
}

func ServiceStateFromReplicas(ready, total int32) envtypes.ServiceState {
	switch {
	case ready <= 0:
		return envtypes.ServiceStateNotReady
	case total > ready:
		return envtypes.ServiceStateDegraded
	default:
		return envtypes.ServiceStateReady
	}
}

func ServiceFromDeployment(d *appsv1.Deployment) *Service {
	s := ServiceFromMeta(&d.ObjectMeta)
	s.State = ServiceStateFromReplicas(d.Status.ReadyReplicas, d.Status.Replicas)

	return s
}

func ServiceFromStatefulSet(d *appsv1.StatefulSet) *Service {
	s := ServiceFromMeta(&d.ObjectMeta)
	s.State = ServiceStateFromReplicas(d.Status.ReadyReplicas, d.Status.Replicas)

	return s
}

func ServiceFromDaemonSet(d *appsv1.DaemonSet) *Service {
	s := ServiceFromMeta(&d.ObjectMeta)
	s.State = ServiceStateFromReplicas(d.Status.NumberReady, d.Status.DesiredNumberScheduled)

	return s
}
