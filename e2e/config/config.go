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

package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"

	"github.com/openclarity/vmclarity/e2e/testenv"
	k8senv "github.com/openclarity/vmclarity/e2e/testenv/kubernetes"
	"github.com/openclarity/vmclarity/e2e/testenv/kubernetes/helm"
	k8senvtypes "github.com/openclarity/vmclarity/e2e/testenv/kubernetes/types"
)

const (
	DefaultEnvPrefix       = "VMCLARITY_E2E"
	DefaultReuseFlag       = false
	DefaultEnvSetupTimeout = 30 * time.Minute
)

type Config struct {
	// ReuseEnv determines if the test environment needs to be set-up/started or not before running the test suite.
	ReuseEnv bool `mapstructure:"use_existing"`
	// EnvSetupTimeout defines the time period before the test environment setup is marked as failed due to timeout.
	EnvSetupTimeout time.Duration `mapstructure:"env_setup_timeout"`
	// TestEnvConfig contains the configuration for testenv library.
	TestEnvConfig testenv.Config `mapstructure:",squash"`
}

func NewConfig() (*Config, error) {
	v := viper.NewWithOptions(
		viper.KeyDelimiter("."),
		viper.EnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_")),
	)

	v.SetEnvPrefix(DefaultEnvPrefix)
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	//
	// E2E configuration
	//

	_ = v.BindEnv("use_existing")
	v.SetDefault("use_existing", DefaultReuseFlag)

	_ = v.BindEnv("env_setup_timeout")
	v.SetDefault("env_setup_timeout", DefaultEnvSetupTimeout)

	//
	// TestEnv configuration
	//

	_ = v.BindEnv("platform")
	v.SetDefault("platform", testenv.DefaultPlatform)

	_ = v.BindEnv("env_name")
	v.SetDefault("env_name", testenv.DefaultEnvName)
	v.RegisterAlias("docker.env_name", "env_name")
	v.RegisterAlias("kubernetes.env_name", "env_name")

	_ = v.BindEnv("apiserver_image")
	v.SetDefault("apiserver_image", testenv.DefaultAPIServer)
	v.RegisterAlias("docker.apiserver_image", "apiserver_image")
	v.RegisterAlias("kubernetes.apiserver_image", "apiserver_image")

	_ = v.BindEnv("orchestrator_image")
	v.SetDefault("orchestrator_image", testenv.DefaultOrchestrator)
	v.RegisterAlias("docker.orchestrator_image", "orchestrator_image")
	v.RegisterAlias("kubernetes.orchestrator_image", "orchestrator_image")

	_ = v.BindEnv("ui_image")
	v.SetDefault("ui_image", testenv.DefaultUI)
	v.RegisterAlias("docker.ui_image", "ui_image")
	v.RegisterAlias("kubernetes.ui_image", "ui_image")

	_ = v.BindEnv("uibackend_image")
	v.SetDefault("uibackend_image", testenv.DefaultUIBackend)
	v.RegisterAlias("docker.uibackend_image", "uibackend_image")
	v.RegisterAlias("kubernetes.uibackend_image", "uibackend_image")

	_ = v.BindEnv("scanner_image")
	v.SetDefault("scanner_image", testenv.DefaultScanner)
	v.RegisterAlias("docker.scanner_image", "scanner_image")
	v.RegisterAlias("kubernetes.scanner_image", "scanner_image")

	_ = v.BindEnv("docker.compose_files")

	_ = v.BindEnv("kubernetes.provider")
	v.SetDefault("kubernetes.provider", k8senv.DefaultProvider)

	_ = v.BindEnv("kubernetes.version")
	v.SetDefault("kubernetes.version", k8senv.DefaultKubernetesVersion)

	_ = v.BindEnv("kubernetes.cluster_name")
	v.SetDefault("kubernetes.cluster_name", testenv.DefaultEnvName)

	_ = v.BindEnv("kubernetes.cluster_creation_timeout")
	v.SetDefault("kubernetes.cluster_creation_timeout", k8senv.DefaultClusterCreationTimeout)

	_ = v.BindEnv("kubernetes.namespace")
	v.SetDefault("kubernetes.namespace", k8senv.DefaultNamespace)
	v.RegisterAlias("kubernetes.helm.namespace", "kubernetes.namespace")

	_ = v.BindEnv("kubernetes.kubeconfig")
	v.SetDefault("kubernetes.kubeconfig", k8senvtypes.DefaultKubeConfigPath())

	_ = v.BindEnv("kubernetes.installer")
	v.SetDefault("kubernetes.installer", k8senv.DefaultInstaller)

	_ = v.BindEnv("kubernetes.helm.chart_dir")

	_ = v.BindEnv("kubernetes.helm.namespace")
	v.SetDefault("kubernetes.helm.namespace", helm.DefaultHelmNamespace)

	_ = v.BindEnv("kubernetes.helm.driver")
	v.SetDefault("kubernetes.helm.driver", helm.DefaultHelmDriver)

	decodeHooks := mapstructure.ComposeDecodeHookFunc(
		// TextUnmarshallerHookFunc is needed to decode custom types
		mapstructure.TextUnmarshallerHookFunc(),
		// Default decoders
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	)

	config := &Config{}
	if err := v.Unmarshal(config, viper.DecodeHook(decodeHooks)); err != nil {
		return nil, fmt.Errorf("failed to parse end-to-end test suite configuration: %w", err)
	}

	return config, nil
}
