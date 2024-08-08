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

// Run start the command and waits until it is exited. It always returns a CommandStatus even if the command has failed.
// nolint:gosec,wrapcheck
func (c *Command) Run(ctx context.Context) (*State, error) {
	state, err := c.Start(ctx)
	if err != nil {
		return state, fmt.Errorf("failed to start command: %w", err)
	}

	if err = state.Wait(); err != nil {
		return state, fmt.Errorf("failed to run command: %w", err)
	}

	return state, nil
}

// Start starts the command and immediately returns a CommandStatus object.
// nolint:gosec,wrapcheck
func (c *Command) Start(ctx context.Context) (*State, error) {
	cmd := exec.CommandContext(ctx, c.Cmd, c.Args...)
	if c.Env == nil || len(c.Env) <= 0 {
		cmd.Env = cmd.Environ()
	}
	cmd.Env = c.Env

	if c.WorkDir != "" {
		cmd.Dir = c.WorkDir
	}

	var stdOut, stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	err := cmd.Start()

	return &State{
		cmd:    cmd,
		StdOut: &stdOut,
		StdErr: &stdErr,
	}, err
}

func (c *Command) String() string {
	return fmt.Sprintf("Command{Cmd: %s, Args: %s, Env: %s, WorkDir: %s}", c.Cmd, c.Args, c.Env, c.WorkDir)
}

type State struct {
	cmd    *exec.Cmd
	StdOut *bytes.Buffer
	StdErr *bytes.Buffer
}

func (s *State) Kill() error {
	if s.cmd.ProcessState != nil {
		return nil
	}

	if err := s.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill process: %w", err)
	}

	return nil
}

func (s *State) Wait() error {
	if s.cmd.ProcessState != nil {
		return nil
	}

	if err := s.cmd.Wait(); err != nil {
		return fmt.Errorf("failed to wait process: %w", err)
	}

	return nil
}

func (s *State) Pid() int {
	if s.cmd.ProcessState != nil {
		return s.cmd.ProcessState.Pid()
	}

	return s.cmd.Process.Pid
}

func (s *State) ExitCode() int {
	if s.cmd.ProcessState != nil {
		return s.cmd.ProcessState.ExitCode()
	}

	return -1
}

func (s *State) Exited() bool {
	if s.cmd.ProcessState != nil {
		return s.cmd.ProcessState.Exited()
	}

	return false
}

func (s *State) Success() bool {
	if s.cmd.ProcessState != nil {
		return s.cmd.ProcessState.Success()
	}

	return false
}

func (s *State) String() string {
	return s.cmd.String()
}
