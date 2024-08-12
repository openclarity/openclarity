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

import "strings"

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
