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

package secret

import (
	"context"
	"fmt"
	"os"

	"github.com/containers/image/v5/docker/reference"
	corev1 "k8s.io/api/core/v1"

	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/utils/k8s"
)

const ImagePullSecretEnvVar = "K8S_IMAGE_PULL_SECRET" // nolint:gosec

type ImagePullSecret struct {
	body string
}

func (s *ImagePullSecret) Name() string {
	return "ImagePullSecret"
}

func (s *ImagePullSecret) IsSupported(_ reference.Named) bool {
	s.body = os.Getenv(ImagePullSecretEnvVar)
	return s.body != ""
}

func (s *ImagePullSecret) GetCredentials(_ context.Context, named reference.Named) (username, password string, err error) {
	secretDataMap := make(map[string][]byte)

	secretDataMap[corev1.DockerConfigJsonKey] = []byte(s.body)
	secrets := []corev1.Secret{{
		Data: secretDataMap,
		Type: corev1.SecretTypeDockerConfigJson,
	}}

	authorization, err := k8s.GetAuthConfig(secrets, named)
	if err != nil {
		return "", "", fmt.Errorf("failed to create authorization: %v", err)
	}

	return authorization.Username, authorization.Password, nil
}
