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

package types

import (
	"context"
	"fmt"
	"strings"
)

type Installer interface {
	// Install deploys manifests specific to the platform
	Install(ctx context.Context) error
	// Uninstall deploys manifests specific to the platform
	Uninstall(ctx context.Context) error
	// Deployments list of names for deployments
	Deployments() []string
}

type InstallerType string

const (
	InstallerTypeHelm InstallerType = "helm"
)

func (i *InstallerType) UnmarshalText(text []byte) error {
	var installer InstallerType

	switch strings.ToLower(string(text)) {
	case strings.ToLower(string(InstallerTypeHelm)):
		installer = InstallerTypeHelm
	default:
		return fmt.Errorf("failed to unmarshal text into Installer: %s", text)
	}

	*i = installer

	return nil
}

var _ Installer = &NoOpInstaller{}

type NoOpInstaller struct{}

func (i *NoOpInstaller) Install(_ context.Context) error {
	return nil
}

func (i *NoOpInstaller) Uninstall(_ context.Context) error {
	return nil
}

func (i *NoOpInstaller) Deployments() []string {
	return nil
}

var _ Installer = &MultiInstaller{}

type MultiInstaller struct {
	installers []Installer
}

func (m *MultiInstaller) Install(ctx context.Context) error {
	for _, installer := range m.installers {
		if installer == nil {
			continue
		}

		if err := installer.Install(ctx); err != nil {
			return fmt.Errorf("failed to install: %w", err)
		}
	}

	return nil
}

func (m *MultiInstaller) Uninstall(ctx context.Context) error {
	for idx := len(m.installers) - 1; idx >= 0; idx-- {
		installer := m.installers[idx]

		if installer == nil {
			continue
		}

		if err := installer.Uninstall(ctx); err != nil {
			return fmt.Errorf("failed to uninstall: %w", err)
		}
	}

	return nil
}

func (m *MultiInstaller) Deployments() []string {
	var deployments []string
	for _, installer := range m.installers {
		d := installer.Deployments()
		if len(d) > 0 {
			deployments = append(deployments, d...)
		}
	}

	return deployments
}

func NewMultiInstaller(installers ...Installer) *MultiInstaller {
	return &MultiInstaller{
		installers: installers,
	}
}
