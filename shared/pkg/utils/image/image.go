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

package image

import (
	"fmt"
	"strings"

	"github.com/containers/image/v5/docker/reference"
	"github.com/sirupsen/logrus"
)

const (
	localImageIDDockerPrefix = "docker://sha256"
	localImageIDSHA256Prefix = "sha256:"
)

func GetImageRef(imageName string) (reference.Named, error) {
	ref, err := reference.ParseNormalizedNamed(imageName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image name. name=%v: %v", imageName, err)
	}

	// strip tag if image has digest and tag
	ref = imageNameWithDigestOrTag(ref)
	// add default tag "latest"
	ref = reference.TagNameOnly(ref)

	return ref, nil
}

// imageNameWithDigestOrTag strips the tag from ambiguous image references that have a digest as well (e.g. `image:tag@sha256:123...`).
// Based on https://github.com/cri-o/cri-o/pull/3060
func imageNameWithDigestOrTag(named reference.Named) reference.Named {
	_, isTagged := named.(reference.NamedTagged)
	canonical, isDigested := named.(reference.Canonical)
	if isTagged && isDigested {
		canonical, err := reference.WithDigest(reference.TrimNamed(named), canonical.Digest())
		if err != nil {
			logrus.Errorf("Failed to create canonical reference - returning the given name. name=%v, %v", named.Name(), err)
			return named
		}

		return canonical
	}

	return named
}

func IsLocalImage(imageID string) bool {
	// image ids with no name and only hash are not pullable from the registry, so we can't scan them.
	if strings.HasPrefix(imageID, localImageIDDockerPrefix) || strings.HasPrefix(imageID, localImageIDSHA256Prefix) {
		return true
	}
	return false
}
