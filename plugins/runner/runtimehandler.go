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

//go:build !linux

package runner

import (
	"context"
	"errors"
	"fmt"

	"github.com/openclarity/vmclarity/plugins/runner/internal/runtimehandler"
	"github.com/openclarity/vmclarity/plugins/runner/internal/runtimehandler/docker"
	"github.com/openclarity/vmclarity/plugins/runner/types"
)

func getPluginRuntimeHandler(ctx context.Context, config types.PluginConfig) (runtimehandler.PluginRuntimeHandler, error) {
	if config.BinaryMode {
		return nil, errors.New("binary mode not supported on this platform")
	}

	handler, err := docker.New(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize docker runtime handler: %w", err)
	}

	return handler, nil
}
