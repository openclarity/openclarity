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

package command

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

type Command struct {
	Cmd     string
	Args    []string
	Env     []string
	WorkDir string
}

// nolint:gosec,wrapcheck
func (c *Command) Run(ctx context.Context) (*CommandResult, error) {
	var stdOut, stdErr bytes.Buffer
	var err error

	cmd := exec.CommandContext(ctx, c.Cmd, c.Args...)
	if c.Env == nil || len(c.Env) <= 0 {
		cmd.Env = cmd.Environ()
	}
	cmd.Env = c.Env

	if c.WorkDir != "" {
		cmd.Dir = c.WorkDir
	}

	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	err = cmd.Run()

	return &CommandResult{
		cmd.ProcessState,
		&stdOut,
		&stdErr,
	}, err
}

func (c Command) String() string {
	return fmt.Sprintf("Command{Cmd: %s, Args: %s, Env: %s, WorkDir: %s}", c.Cmd, c.Args, c.Env, c.WorkDir)
}
