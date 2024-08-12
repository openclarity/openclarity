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

package image_helper // nolint:revive,stylecheck

import (
	"errors"
	"strings"

	"github.com/containers/image/v5/docker/reference"
	log "github.com/sirupsen/logrus"
)

// FsLayerCommand represents a history command of a layer in a docker image.
type FsLayerCommand struct {
	Command string
	Layer   string
}

func GetHashFromRepoDigest(repoDigests []string, imageName string) string {
	if len(repoDigests) == 0 {
		return ""
	}

	normalizedName, err := reference.ParseNormalizedNamed(imageName)
	if err != nil {
		log.Errorf("Failed to parse image name %s to normalized named: %v", imageName, err)
		return ""
	}
	familiarName := reference.FamiliarName(normalizedName)
	// iterating over RepoDigests and use RepoDigest which match to imageName
	for _, repoDigest := range repoDigests {
		normalizedRepoDigest, err := reference.ParseNormalizedNamed(repoDigest)
		if err != nil {
			log.Errorf("Failed to parse repoDigest %s, %v", repoDigest, err)
			return ""
		}
		// RepoDigests can be different based on the registry
		//        ],
		//        "RepoDigests": [
		//            "debian@sha256:2906804d2a64e8a13a434a1a127fe3f6a28bf7cf3696be4223b06276f32f1f2d",
		//            "poke/debian@sha256:a4c378901a2ba14fd331e96a49101556e91ed592d5fd68ba7405fdbf9b969e61",
		//            "poke/testdebian@sha256:a4c378901a2ba14fd331e96a49101556e91ed592d5fd68ba7405fdbf9b969e61"
		//        ],
		// Check which RegoDigest should be used
		if reference.FamiliarName(normalizedRepoDigest) == familiarName {
			return normalizedRepoDigest.(reference.Digested).Digest().Encoded() // nolint:forcetypeassert
		}
	}
	return ""
}

func GetHashFromRepoDigestsOrImageID(repoDigests []string, imageID string, imageName string) (string, error) {
	if imageID == "" && len(repoDigests) == 0 {
		return "", errors.New("RepoDigest and ImageID are missing")
	}

	hash := GetHashFromRepoDigest(repoDigests, imageName)
	if hash == "" {
		// set hash using ImageID (https://github.com/opencontainers/image-spec/blob/main/config.md#imageid) if repo digests are missing
		// image ID is represented as a hexadecimal encoding of 256 bits, e.g., sha256:a9561eb1b190625c9adb5a9513e72c4dedafc1cb2d4c5236c9a6957ec7dfd5a9
		// we need only the hash
		_, h, found := strings.Cut(imageID, ":")
		if found {
			hash = h
		} else {
			hash = imageID
		}
	}
	return hash, nil
}
