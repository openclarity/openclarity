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
	"fmt"

	"github.com/spf13/viper"
)

const (
	DefaultEnvPrefix = "VMCLARITY_KUBERNETES"
)

type Config struct {
	// KubeConfig defines a path to a kubeconfig file to use to connect to
	// the Kubernetes API
	KubeConfig string `mapstructure:"kubeconfig"`

	// Container Runtime Discovery Namespace is the namespace where the
	// container runtime discovery daemonset is running.
	ContainerRuntimeDiscoveryNamespace string `mapstructure:"container_runtime_discovery_namespace"`
}

func NewConfig() (*Config, error) {
	// Avoid modifying the global instance
	v := viper.New()

	v.SetEnvPrefix(DefaultEnvPrefix)
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	_ = v.BindEnv("kubeconfig")
	_ = v.BindEnv("container_runtime_discovery_namespace")

	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to parse provider configuration. Provider=Kubernetes: %w", err)
	}

	return config, nil
}
