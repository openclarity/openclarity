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

package creds

import (
	"context"
	"fmt"

	"github.com/containers/image/v5/docker/reference"
	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"github.com/google/go-containerregistry/pkg/name"

	"github.com/openclarity/kubeclarity/shared/v2/pkg/utils/image"
)

func ExtractCredentials(imageName string) (string, string, error) {
	ref, err := image.GetImageRef(imageName)
	if err != nil {
		return "", "", fmt.Errorf("failed to get image ref. image name=%v: %v", imageName, err)
	}

	return getCredsWithK8sChain(ref)
}

func getCredsWithK8sChain(namedImageRef reference.Named) (string, string, error) {
	keyChain, err := k8schain.NewNoClient(context.TODO())
	if err != nil {
		return "", "", fmt.Errorf("failed to create k8schain: %v", err)
	}

	repository, err := name.NewRepository(reference.FamiliarName(namedImageRef))
	if err != nil {
		return "", "", fmt.Errorf("failed to create repository: %v", err)
	}

	authenticator, err := keyChain.Resolve(repository)
	if err != nil {
		return "", "", fmt.Errorf("failed to create authenticator: %v", err)
	}

	authorization, err := authenticator.Authorization()
	if err != nil {
		return "", "", fmt.Errorf("failed to create authorization: %v", err)
	}

	return authorization.Username, authorization.Password, nil
}
