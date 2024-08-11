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

//go:build !linux && !darwin
// +build !linux,!darwin

package blockdevice

import (
	"context"
	"fmt"
	"runtime"
)

type BlockDevice struct {
	Name       string
	Path       string
	FSType     string
	MountPoint string
	Label      string
	UUID       string
}

func List(_ context.Context) ([]BlockDevice, error) {
	return nil, fmt.Errorf("mount is unsupported on %s platform", runtime.GOOS)
}
