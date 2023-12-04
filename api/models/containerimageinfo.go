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

package models

import (
	"fmt"
	"strings"
)

// GetFirstRepoTag returns the first repo tag if it exists. Otherwise, returns false.
func (c *ContainerImageInfo) GetFirstRepoTag() (string, bool) {
	var tag string
	var ok bool

	if c.RepoTags != nil && len(*c.RepoTags) > 0 {
		tag, ok = (*c.RepoTags)[0], true
	}

	return tag, ok
}

// GetFirstRepoDigest returns the first repo digest if it exists. Otherwise, returns false.
func (c *ContainerImageInfo) GetFirstRepoDigest() (string, bool) {
	var digest string
	var ok bool

	if c.RepoDigests != nil && len(*c.RepoDigests) > 0 {
		digest, ok = (*c.RepoDigests)[0], true
	}

	return digest, ok
}

// Merge merges target and c together and then returns the merged result. c is
// not a pointer so that:
// a) a non-pointer ContainerImageInfo can be merged.
// b) the source ContainerimageInfo can not be modified by this function.
func (c ContainerImageInfo) Merge(target ContainerImageInfo) (ContainerImageInfo, error) {
	id, err := CoalesceComparable(c.ImageID, target.ImageID)
	if err != nil {
		return c, fmt.Errorf("failed to merge Id field: %w", err)
	}

	// NOTE(sambetts) Size seems to depend on the host and container
	// runtime; there doesn't seem to be a clean way to get a repeatable
	// size. If the sizes conflict we'll pick the larger of the two sizes
	// as it is likely to be more accurate to the real size not taking into
	// account deduplication in the CRI etc. The sizes are normally within
	// a few kilobytes of each other.
	size, err := CoalesceComparable(*c.Size, *target.Size)
	if err != nil {
		if *c.Size > *target.Size {
			size = *c.Size
		} else {
			size = *target.Size
		}
	}

	os, err := CoalesceComparable(*c.Os, *target.Os)
	if err != nil {
		return c, fmt.Errorf("failed to merge Os field: %w", err)
	}

	architecture, err := CoalesceComparable(*c.Architecture, *target.Architecture)
	if err != nil {
		return c, fmt.Errorf("failed to merge Architecture field: %w", err)
	}

	labels := UnionSlices(*c.Labels, *target.Labels)

	repoDigests := UnionSlices(*c.RepoDigests, *target.RepoDigests)

	repoTags := UnionSlices(*c.RepoTags, *target.RepoTags)

	return ContainerImageInfo{
		ImageID:      id,
		Size:         &size,
		Labels:       &labels,
		Os:           &os,
		Architecture: &architecture,
		RepoDigests:  &repoDigests,
		RepoTags:     &repoTags,
	}, nil
}

const nilString = "nil"

func (c ContainerImageInfo) String() string {
	size := nilString
	if c.Size != nil {
		size = fmt.Sprintf("%d", *c.Size)
	}

	labels := nilString
	if c.Labels != nil {
		l := make([]string, len(*c.Labels))
		for i, label := range *c.Labels {
			l[i] = fmt.Sprintf("{Key: \"%s\", Value: \"%s\"}", label.Key, label.Value)
		}
		labels = fmt.Sprintf("[%s]", strings.Join(l, ", "))
	}

	os := nilString
	if c.Os != nil {
		os = *c.Os
	}

	architecture := nilString
	if c.Architecture != nil {
		architecture = *c.Architecture
	}

	repoDigests := nilString
	if c.RepoDigests != nil {
		repoDigests = fmt.Sprintf("[%s]", strings.Join(*c.RepoDigests, ", "))
	}

	repoTags := nilString
	if c.RepoTags != nil {
		repoTags = fmt.Sprintf("[%s]", strings.Join(*c.RepoTags, ", "))
	}

	return fmt.Sprintf("{ImageID: %s, Size: %s, Labels: %s, Arch: %s, OS: %s, Digests: %s, Tags: %s}", c.ImageID, size, labels, architecture, os, repoDigests, repoTags)
}
