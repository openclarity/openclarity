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

//go:build linux || freebsd || openbsd || darwin
// +build linux freebsd openbsd darwin

package mount

import (
	"context"
	"fmt"
	"strings"

	"github.com/moby/sys/mountinfo"

	"github.com/openclarity/vmclarity/pkg/shared/command"
)

const DefaultBinaryPath = "mount"

const (
	ErrUnknown                          MountErrorKind = -1
	ErrSuccess                          MountErrorKind = 0
	ErrIncorrectInvocationOrPermissions MountErrorKind = 1
	ErrSystemError                      MountErrorKind = 2
	ErrInternalBug                      MountErrorKind = 4
	ErrUserInterrupt                    MountErrorKind = 8
	ErrProblemsWritingOrLocking         MountErrorKind = 16
	ErrMountFailure                     MountErrorKind = 32
	ErrSomeMountSucceeded               MountErrorKind = 64
)

type MountErrorKind int

// nolint:cyclop
func (k MountErrorKind) String() string {
	switch k {
	case ErrSuccess:
		return "Success"
	case ErrIncorrectInvocationOrPermissions:
		return "Incorrect invocation or permissions"
	case ErrSystemError:
		return "System error (out of memory, cannot fork, no more loop devices)"
	case ErrInternalBug:
		return "Internal mount bug"
	case ErrUserInterrupt:
		return "User interrupt"
	case ErrProblemsWritingOrLocking:
		return "Problems writing or locking /etc/mtab"
	case ErrMountFailure:
		return "Mount failure"
	case ErrSomeMountSucceeded:
		return "Some mount succeeded"
	case ErrUnknown:
		fallthrough
	default:
		return "Unknown error"
	}
}

func NewMountErrorKind(errCode int) MountErrorKind {
	switch errCode {
	case int(ErrSuccess):
		return ErrSuccess
	case int(ErrIncorrectInvocationOrPermissions):
		return ErrIncorrectInvocationOrPermissions
	case int(ErrSystemError):
		return ErrIncorrectInvocationOrPermissions
	case int(ErrInternalBug):
		return ErrIncorrectInvocationOrPermissions
	case int(ErrUserInterrupt):
		return ErrIncorrectInvocationOrPermissions
	case int(ErrProblemsWritingOrLocking):
		return ErrIncorrectInvocationOrPermissions
	case int(ErrMountFailure):
		return ErrIncorrectInvocationOrPermissions
	case int(ErrSomeMountSucceeded):
		return ErrIncorrectInvocationOrPermissions
	default:
		return ErrIncorrectInvocationOrPermissions
	}
}

type MountError struct {
	Kind MountErrorKind
	Msg  string
}

func (e MountError) Error() string {
	if e.Msg != "" {
		return fmt.Sprintf("%s: %s", e.Kind, e.Msg)
	}

	return e.Kind.String()
}

func NewMountError(errCode int, errMsg string) error {
	return MountError{
		Kind: NewMountErrorKind(errCode),
		Msg:  errMsg,
	}
}

func Mount(ctx context.Context, source, target, fsType string, opts []string) error {
	args := []string{"--types", fsType, "--source", source, "--target", target}
	if opts != nil {
		args = append(args, []string{"--options", strings.Join(opts, ",")}...)
	}

	cmd := &command.Command{
		Cmd:  DefaultBinaryPath,
		Args: args,
	}

	result, err := cmd.Run(ctx)
	if err != nil {
		err = NewMountError(result.ExitCode(), result.StdErr.String())
		return fmt.Errorf("failed to run mount command: %w", err)
	}

	return nil
}

func List(_ context.Context, filter mountinfo.FilterFunc) ([]*mountinfo.Info, error) {
	mountpoints, err := mountinfo.GetMounts(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of mountpoints: %w", err)
	}

	return mountpoints, nil
}
