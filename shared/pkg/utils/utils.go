// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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

package utils

import (
	"bytes"
	"fmt"
	"os/exec"
)

const (
	DefaultResourceReadyWaitTimeoutMin   = 3
	DefaultResourceReadyCheckIntervalSec = 3
)

func RunCommand(cmd *exec.Cmd) ([]byte, error) {
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run command %v, error: %w, stdout: %v, stderr: %v", cmd.String(), err, outb.String(), errb.String())
	}
	return outb.Bytes(), nil
}

func StringPtr(val string) *string {
	ret := val
	return &ret
}

func BoolPtr(val bool) *bool {
	ret := val
	return &ret
}

func Int32Ptr(val int32) *int32 {
	ret := val
	return &ret
}

func PointerTo[T any](value T) *T {
	return &value
}
