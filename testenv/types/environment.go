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
	"io"
	"strings"
	"time"
)

type Environment interface {
	// SetUp the test environment by installing the necessary components.
	// Returns error if it fails to set up the environment.
	SetUp(ctx context.Context) error
	// TearDown the test environment by uninstalling components installed via Setup.
	// Returns error if it fails to clean up the environment.
	TearDown(ctx context.Context) error
	// ServicesReady returns bool based on the health status of the services.
	// Returns error if it fails to determine health status for services.
	ServicesReady(ctx context.Context) (bool, error)
	// ServiceLogs writes service logs to io.Writer for list of services from startTime timestamp.
	// Returns error if it cannot retrieve logs.
	ServiceLogs(ctx context.Context, services []string, startTime time.Time, stdout, stderr io.Writer) error
	// Services returns a list of services for the environment.
	Services(ctx context.Context) (Services, error)
	// Endpoints returns an Endpoints object containing API endpoints for services
	Endpoints(ctx context.Context) (*Endpoints, error)
	// Context updates the provided ctx with environment specific data like with initialized client data allowing tests
	// to interact with the underlying infrastructure.
	Context(ctx context.Context) (context.Context, error)
}

type EnvironmentType string

const (
	EnvironmentTypeDocker     EnvironmentType = "docker"
	EnvironmentTypeKubernetes EnvironmentType = "kubernetes"
	EnvironmentTypeAWS        EnvironmentType = "aws"
	EnvironmentTypeGCP        EnvironmentType = "gcp"
	EnvironmentTypeAzure      EnvironmentType = "azure"
)

func (p *EnvironmentType) UnmarshalText(text []byte) error {
	var platform EnvironmentType

	switch strings.ToLower(string(text)) {
	case strings.ToLower(string(EnvironmentTypeDocker)):
		platform = EnvironmentTypeDocker
	case strings.ToLower(string(EnvironmentTypeKubernetes)):
		platform = EnvironmentTypeKubernetes
	case strings.ToLower(string(EnvironmentTypeAWS)):
		platform = EnvironmentTypeAWS
	case strings.ToLower(string(EnvironmentTypeGCP)):
		platform = EnvironmentTypeGCP
	case strings.ToLower(string(EnvironmentTypeAzure)):
		platform = EnvironmentTypeAzure
	default:
		return fmt.Errorf("failed to unmarshal text into Environment: %s", text)
	}

	*p = platform

	return nil
}
