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

package k8s

import (
	"strings"

	"github.com/containers/image/v5/docker/reference"
	log "github.com/sirupsen/logrus"
)

const MaxK8sJobName = 63

// ParseImageHash extracts image hash from image ID
// input: docker-pullable://gcr.io/development-infra-208909/kubeclarity@sha256:6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa
// output: 6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa
func ParseImageHash(imageID string) string {
	index := strings.LastIndex(imageID, ":")
	if index == -1 {
		return ""
	}

	return imageID[index+1:]
}

// NormalizeImageID remove "docker-pullable://" prefix from imageID if exists and then normalize it.
// https://github.com/kubernetes/kubernetes/issues/95968
// input: docker-pullable://gcr.io/development-infra-208909/kubeclarity@sha256:6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa
// output: gcr.io/development-infra-208909/kubeclarity@sha256:6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa
func NormalizeImageID(imageID string) string {
	imageID = strings.TrimPrefix(imageID, "docker-pullable://")

	named, err := reference.ParseNormalizedNamed(imageID)
	if err != nil {
		log.Errorf("Failed to parse image id. image id=%v: %v", imageID, err)
		return imageID
	}
	return named.String()
}
