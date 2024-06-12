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

package diskutil

import (
	"context"
	"fmt"
	"os"

	"github.com/openclarity/vmclarity/utils/command"
)

type Diskutil struct {
	BinaryPath  string
	Environment []string
}

func (d *Diskutil) List(ctx context.Context, devPaths ...string) ([]BlockDevice, error) {
	args := []string{"info", "-all"}
	if devPaths != nil {
		args = append(args, devPaths...)
	}

	o, err := d.run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to run diskutil command: %w", err)
	}

	blockDevices, err := parse(o.StdOut)
	if err != nil {
		return nil, fmt.Errorf("failed to parse diskutil output: %w", err)
	}

	return blockDevices, nil
}

func (d *Diskutil) run(ctx context.Context, args ...string) (*command.State, error) {
	cmd := &command.Command{
		Cmd:  d.BinaryPath,
		Args: args,
		Env:  d.Environment,
	}

	result, err := cmd.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run diskutil command: %w", err)
	}

	return result, nil
}

func New() *Diskutil {
	return &Diskutil{
		BinaryPath:  "diskutil",
		Environment: os.Environ(),
	}
}
