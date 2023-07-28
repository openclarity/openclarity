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
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/openclarity/vmclarity/pkg/shared/command"
)

type LsBlk struct {
	BinaryPath  string
	Environment []string
}

func (l *LsBlk) List(ctx context.Context) ([]BlockDevice, error) {
	result, err := l.help(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run lsblk command: %w", err)
	}

	if strings.Contains(string(result), "--json") {
		return l.listWithJSON(ctx)
	}

	return l.listWithPairs(ctx)
}

func (l *LsBlk) help(ctx context.Context) ([]byte, error) {
	cmd := &command.Command{
		Cmd:  l.BinaryPath,
		Args: []string{"--help"},
		Env:  l.Environment,
	}

	result, err := cmd.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run lsblk command: %w", err)
	}

	return result.StdOut.Bytes(), nil
}

func (l *LsBlk) listWithJSON(ctx context.Context) ([]BlockDevice, error) {
	cmd := &command.Command{
		Cmd:  l.BinaryPath,
		Args: []string{"--json", "--output-all", "--bytes", "--all"},
		Env:  l.Environment,
	}

	result, err := cmd.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run lsblk command: %w", err)
	}

	blockDevices, err := parseJSONFormat(result.StdOut)
	if err != nil {
		return nil, fmt.Errorf("failed to parse lsblk output: %w", err)
	}

	return blockDevices, nil
}

func (l *LsBlk) listWithPairs(ctx context.Context) ([]BlockDevice, error) {
	cmd := &command.Command{
		Cmd:  l.BinaryPath,
		Args: []string{"--pairs", "--output-all", "--bytes", "--all"},
		Env:  l.Environment,
	}

	result, err := cmd.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run lsblk command: %w", err)
	}

	blockDevices, err := parsePairsFormat(result.StdOut)
	if err != nil {
		return nil, fmt.Errorf("failed to parse lsblk output: %w", err)
	}

	return blockDevices, nil
}

func New() *LsBlk {
	return &LsBlk{
		BinaryPath:  "lsblk",
		Environment: os.Environ(),
	}
}
