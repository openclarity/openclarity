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

package utils

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	kubeclarityutils "github.com/openclarity/kubeclarity/shared/pkg/utils"

	"github.com/openclarity/vmclarity/pkg/shared/families/types"
)

// InputSizesCache global cache of already calculated input sizes. If input type is a DIR/ROOTFS/FILE than the key is the input path.
var InputSizesCache = make(map[string]int64)

func TrimMountPath(toTrim string, mountPath string) string {
	// avoid representing root directory as empty string
	if toTrim == mountPath {
		return "/"
	}
	return strings.TrimPrefix(toTrim, mountPath)
}

func RemoveMountPathSubStringIfNeeded(toTrim string, mountPath string) string {
	if strings.Contains(toTrim, mountPath) {
		// assume prefix is /mnt:
		// first address cases like /mnt/foo -> should be /foo
		toTrim = strings.ReplaceAll(toTrim, mountPath+"/", "/")
		// then address cases like /mnt -> should be /
		return strings.ReplaceAll(toTrim, mountPath, "/")
	}
	return toTrim
}

func ShouldStripInputPath(inputShouldStrip *bool, familyShouldStrip bool) bool {
	if inputShouldStrip == nil {
		return familyShouldStrip
	}
	return *inputShouldStrip
}

func GetInputSize(input types.Input) (int64, error) {
	switch input.InputType {
	case string(kubeclarityutils.ROOTFS), string(kubeclarityutils.DIR), string(kubeclarityutils.FILE):
		// check if already exists in cache
		sizeFromCache, ok := InputSizesCache[input.Input]
		if ok {
			return sizeFromCache, nil
		}

		// calculate the size and add it to the cache
		size, err := DirSizeMB(input.Input)
		if err != nil {
			return 0, fmt.Errorf("failed to get dir size: %v", err)
		}
		InputSizesCache[input.InputType] = size
		return size, nil
	default:
		// currently other input types are not supported for size benchmarking.
		return 0, nil
	}
}

const megaBytesToBytes = 1000 * 1000

func DirSizeMB(path string) (int64, error) {
	var dirSizeBytes int64 = 0

	readSize := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("unable to evaluate path %s: %w", path, err)
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("unable to stat path %s: %w", path, err)
		}
		dirSizeBytes += info.Size()
		return nil
	}

	if err := filepath.WalkDir(path, readSize); err != nil {
		return 0, fmt.Errorf("failed to walk dir: %v", err)
	}

	return dirSizeBytes / megaBytesToBytes, nil
}
