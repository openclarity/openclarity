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

import "fmt"

type SourceType string

const (
	SBOM          SourceType = "sbom"
	IMAGE         SourceType = "image"
	DOCKERARCHIVE SourceType = "docker-archive"
	OCIARCHIVE    SourceType = "oci-archive"
	OCIDIR        SourceType = "oci-dir"
	DIR           SourceType = "dir"
	ROOTFS        SourceType = "rootfs"
	FILE          SourceType = "file"
)

func CreateSource(sourceType SourceType, src string, localImage bool) string {
	switch sourceType {
	case IMAGE:
		return setImageSource(localImage, src)
	case DOCKERARCHIVE:
		return "docker-archive:" + src
	case OCIARCHIVE:
		return "oci-archive:" + src
	case OCIDIR:
		return "oci-dir:" + src
	case ROOTFS, DIR:
		return fmt.Sprintf("%s:%s", DIR, src)
	case FILE, SBOM:
		fallthrough
	default:
		return fmt.Sprintf("%s:%s", sourceType, src)
	}
}

func setImageSource(local bool, source string) string {
	if local {
		return "docker:" + source
	}
	return "registry:" + source
}
