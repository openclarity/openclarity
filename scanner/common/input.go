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

package common

type InputType string

const (
	SBOM          InputType = "sbom"
	IMAGE         InputType = "image"
	DOCKERARCHIVE InputType = "docker-archive"
	OCIARCHIVE    InputType = "oci-archive"
	OCIDIR        InputType = "oci-dir"
	DIR           InputType = "dir"
	ROOTFS        InputType = "rootfs"
	FILE          InputType = "file"
	CSV           InputType = "csv"
)

func (s InputType) GetSource(localImage bool) string {
	switch s {
	case IMAGE:
		return getImageSource(localImage)
	case ROOTFS, DIR:
		return string(DIR)
	case DOCKERARCHIVE, OCIARCHIVE, OCIDIR, FILE, SBOM, CSV:
		fallthrough
	default:
		return string(s)
	}
}

// IsOnFilesystem returns true if the InputType can be found on the filesystem.
func (s InputType) IsOnFilesystem() bool {
	switch s {
	case IMAGE:
		return false
	case ROOTFS, DIR, DOCKERARCHIVE, OCIARCHIVE, OCIDIR, FILE, SBOM, CSV:
		fallthrough
	default:
		return true
	}
}

// IsOneOf returns true if one of provided input types matches the actual type.
func (s InputType) IsOneOf(types ...InputType) bool {
	for _, typ := range types {
		if s == typ {
			return true
		}
	}

	return false
}

func getImageSource(local bool) string {
	if local {
		return "docker"
	}
	return "registry"
}
