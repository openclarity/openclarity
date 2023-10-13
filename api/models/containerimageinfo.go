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
