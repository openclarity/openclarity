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

package lsblk

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/openclarity/vmclarity/utils/command"
)

type LsBlk struct {
	BinaryPath  string
	Environment []string
}

func (l *LsBlk) List(ctx context.Context, devPaths ...string) ([]BlockDevice, error) {
	o, err := l.help(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run lsblk command: %w", err)
	}

	args := []string{"--output-all", "--bytes", "--all"}
	var parser func(*bytes.Buffer) ([]BlockDevice, error)
	if strings.Contains(o.StdOut.String(), "--json") {
		args = append(args, "--json")
		parser = parseJSONFormat
	} else {
		args = append(args, "--pairs")
		parser = parsePairsFormat
	}

	if devPaths != nil {
		args = append(args, devPaths...)
	}

	o, err = l.run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to run lsblk command: %w", err)
	}

	blockDevices, err := parser(o.StdOut)
	if err != nil {
		return nil, fmt.Errorf("failed to parse lsblk output: %w", err)
	}

	return blockDevices, nil
}

func (l *LsBlk) help(ctx context.Context) (*command.State, error) {
	cmd := &command.Command{
		Cmd:  l.BinaryPath,
		Args: []string{"--help"},
		Env:  l.Environment,
	}

	result, err := cmd.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run lsblk command: %w", err)
	}

	return result, nil
}

func (l *LsBlk) run(ctx context.Context, args ...string) (*command.State, error) {
	cmd := &command.Command{
		Cmd:  l.BinaryPath,
		Args: args,
		Env:  l.Environment,
	}

	result, err := cmd.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run lsblk command: %w", err)
	}

	return result, nil
}

func New() *LsBlk {
	return &LsBlk{
		BinaryPath:  "lsblk",
		Environment: os.Environ(),
	}
}
