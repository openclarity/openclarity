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
	DOCKERARCHIVE SourceType = "dockerarchive"
	OCIARCHIVE    SourceType = "ociarchive"
	DIR           SourceType = "dir"
	ROOTFS        SourceType = "rootfs"
	FILE          SourceType = "file"
)

func ValidateInputType(inputType string) (SourceType, error) {
	switch inputType {
	case "sbom", "SBOM":
		return SBOM, nil
	case "image", "IMAGE", "":
		return IMAGE, nil
	case "docker-archive":
		return DOCKERARCHIVE, nil
	case "oci-archive":
		return OCIARCHIVE, nil
	case "dir", "DIR", "directory":
		return DIR, nil
	case "file", "FILE":
		return FILE, nil
	case "rootfs", "ROOTFS":
		return ROOTFS, nil
	default:
		return "", fmt.Errorf("unsupported input type: %s", inputType)
	}
}

func CreateSource(sourceType SourceType, src string, localImage bool) string {
	switch sourceType {
	case IMAGE:
		return setImageSource(localImage, src)
	case DOCKERARCHIVE:
		return fmt.Sprintf("docker-archive:%s", src)
	case OCIARCHIVE:
		return fmt.Sprintf("oci-archive:%s", src)
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
