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

package helm

import (
	"context"
	"fmt"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"

	"github.com/openclarity/vmclarity/e2e/testenv/utils"
)

const (
	DefaultInstallTimeout   = 15 * time.Minute
	DefaultUninstallTimeout = 5 * time.Minute
)

type Installer struct {
	config *Config
	getter genericclioptions.RESTClientGetter

	chart   *chart.Chart
	values  Values
	release *release.Release

	meta map[string]interface{}
}

func (i *Installer) Install(ctx context.Context) error {
	logger := utils.GetLoggerFromContextOrDiscard(ctx).WithFields(i.meta)

	// Init action.Configuration
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(i.getter, i.config.Namespace, i.config.StorageDriver, logger.Debugf); err != nil {
		return fmt.Errorf("failed to initialize action: %w", err)
	}

	if yaml, err := i.values.YAML(); err != nil {
		logger.Warn("failed to encode values to YAML")
	} else {
		logger.Debugf("installing Helm Chart with values:\n%s\n", yaml)
	}

	install := action.NewInstall(actionConfig)
	install.Namespace = i.config.Namespace
	install.ReleaseName = i.config.ReleaseName
	install.Wait = true
	install.Timeout = DefaultInstallTimeout
	r, err := install.RunWithContext(ctx, i.chart, i.values)
	if err != nil {
		return fmt.Errorf("failed to install Helm chart: %w", err)
	}
	logger.Debugf("Helm chart has been installed. ChartName=%s ReleaseName=%s", i.chart.Name(), r.Name)
	i.release = r

	return nil
}

func (i *Installer) Uninstall(ctx context.Context) error {
	logger := utils.GetLoggerFromContextOrDiscard(ctx).WithFields(i.meta)

	// Init action.Configuration
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(i.getter, i.config.Namespace, i.config.StorageDriver, logger.Debugf); err != nil {
		return fmt.Errorf("failed to initialize action: %w", err)
	}

	uninstall := action.NewUninstall(actionConfig)
	uninstall.IgnoreNotFound = true
	uninstall.Timeout = DefaultUninstallTimeout
	uninstall.Wait = true

	_, err := uninstall.Run(i.config.ReleaseName)
	if err != nil {
		return fmt.Errorf("failed to uninstall Helm chart. ReleaseName=%s: %w", i.config.ReleaseName, err)
	}
	logger.Debugf("Helm chart has been uninstalled. ReleaseName=%s", i.config.ReleaseName)

	return nil
}

func (i *Installer) Deployments() []string {
	return []string{i.config.ReleaseName}
}

func New(config *Config, opts ...ConfigOptFn) (*Installer, error) {
	opts = append(opts,
		withResolvedChartPath(),
		withResolvedKubeConfigPath(),
	)
	if err := applyConfigWithOpts(config, opts...); err != nil {
		return nil, fmt.Errorf("failed to apply options to Config: %w", err)
	}

	restClientGetter := &genericclioptions.ConfigFlags{
		Namespace:  &config.Namespace,
		KubeConfig: &config.KubeConfigPath,
		WrapConfigFn: func(config *rest.Config) *rest.Config {
			config.Burst = 100
			config.UserAgent = "testenv/helm"
			return config
		},
	}

	// Load Helm Chart
	helmChart, err := LoadPathOrEmbedded(config.ChartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load helm chart(s): %w", err)
	}

	// Load values
	values := NewEmptyValues()
	if err = applyValuesWithOpts(&values, config.valuesOpts...); err != nil {
		return nil, err
	}

	return &Installer{
		getter: restClientGetter,
		config: config,
		chart:  helmChart,
		values: values,
		meta: map[string]interface{}{
			"installer": "helm",
			"namespace": config.Namespace,
		},
	}, nil
}
