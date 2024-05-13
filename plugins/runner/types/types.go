// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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
	"time"

	plugintypes "github.com/openclarity/vmclarity/plugins/sdk-go/types"
)

const (
	WaitReadyTimeout    = 1 * time.Minute
	GracefulStopTimeout = 10 * time.Second
	ScanTimeout         = 15 * time.Minute
)

type PluginConfig struct {
	// Name is the name of the plugin scanner. This should be unique as only one PluginRunner with the same Name can exist.
	Name string `yaml:"name" mapstructure:"name"`
	// ImageName is the name of the docker image that will be used to run the plugin scanner
	ImageName string `yaml:"image_name" mapstructure:"image_name"`
	// InputDir is a directory where the plugin scanner will read the asset filesystem
	InputDir string `yaml:"input_dir" mapstructure:"input_dir"`
	// ScannerConfig is a json string that will be passed to the scanner in the plugin
	ScannerConfig string `yaml:"scanner_config" mapstructure:"scanner_config"`
	// TimeoutSeconds defines the number of seconds before the scan is marked failed due to timeout
	TimeoutSeconds int `yaml:"timeout_seconds" mapstructure:"timeout_seconds"`
}

type PluginRunner interface {
	Start(ctx context.Context) error
	WaitReady(ctx context.Context) error
	Metadata(ctx context.Context) (*plugintypes.Metadata, error)
	Run(ctx context.Context) error
	WaitDone(ctx context.Context) error
	Stop(ctx context.Context) error
	Logs(ctx context.Context) (io.ReadCloser, error)
	Result(ctx context.Context) (io.ReadCloser, error)
	Remove(ctx context.Context) error
}
