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
	SBOM  SourceType = "sbom"
	IMAGE SourceType = "image"
	DIR   SourceType = "dir"
	FILE  SourceType = "file"
)

func ValidateInputType(inputType string) (SourceType, error) {
	switch inputType {
	case "sbom", "SBOM":
		return SBOM, nil
	case "image", "IMAGE", "":
		return IMAGE, nil
	case "dir", "DIR", "directory":
		return DIR, nil
	case "file", "FILE":
		return FILE, nil
	default:
		return "", fmt.Errorf("unsupported input type: %s", inputType)
	}
}

func CreateSource(sourceType SourceType, src string, localImage bool) string {
	if sourceType != IMAGE {
		src = fmt.Sprintf("%s:%s", sourceType, src)
	}
	return setImageSource(localImage, src)
}

func setImageSource(local bool, source string) string {
	if local {
		return "docker:" + source
	}
	return "registry:" + source
}
