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

func CreateSource(sourceType SourceType, localImage bool) string {
	switch sourceType {
	case IMAGE:
		return setImageSource(localImage)
	case ROOTFS, DIR:
		return string(DIR)
	case DOCKERARCHIVE, OCIARCHIVE, OCIDIR, FILE, SBOM:
		fallthrough
	default:
		return string(sourceType)
	}
}

func setImageSource(local bool) string {
	if local {
		return "docker"
	}
	return "registry"
}
