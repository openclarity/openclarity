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
	"io"
	"net/url"
	"time"
)

type Environment interface {
	// Start the test environment by ensuring that all the services are running.
	// Returns error if it fails to start services.
	Start(ctx context.Context) error
	// Stop the test environment by ensuring that all the services are stopped.
	// Returns error if it fails to stop services.
	Stop(ctx context.Context) error
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
	Services() []string
	// VMClarityAPIURL returns the URL for communicating with VMClarity API.
	// Returns error if it fails to determine the URL.
	VMClarityAPIURL() (*url.URL, error)
	// Context updates the provided ctx with environment specific data like with initialized client data allowing tests
	// to interact with the underlying infrastructure.
	Context(ctx context.Context) context.Context
}

type Platform string

const (
	Docker     Platform = "docker"
	Kubernetes Platform = "kubernetes"
	AWS        Platform = "aws"
	GCP        Platform = "gpc"
	Azure      Platform = "azure"
)

type Config struct {
	// Platform defines the platform to be used for running end-to-end test suite.
	Platform Platform `mapstructure:"platform"`
	// ReuseEnv determines if the test environment needs to be set-up/started or not before running the test suite.
	ReuseEnv bool `mapstructure:"use_existing"`
}
