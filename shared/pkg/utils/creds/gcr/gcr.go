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

package gcr

import (
	"context"
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/docker-credential-gcr/config"
	"github.com/GoogleCloudPlatform/docker-credential-gcr/credhelper"
	"github.com/GoogleCloudPlatform/docker-credential-gcr/store"
	"github.com/containers/image/v5/docker/reference"
)

const gcrURL = "gcr.io"

type GCR struct{}

func (g *GCR) Name() string {
	return "gcr"
}

func (g *GCR) IsSupported(named reference.Named) bool {
	return strings.HasSuffix(reference.Domain(named), gcrURL)
}

func (g *GCR) GetCredentials(_ context.Context, named reference.Named) (username, password string, err error) {
	credStore, err := store.DefaultGCRCredStore()
	if err != nil {
		return "", "", fmt.Errorf("failed to create default GCR cred store: %v", err)
	}

	userCfg, err := config.LoadUserConfig()
	if err != nil {
		return "", "", fmt.Errorf("failed to load user config: %v", err)
	}

	username, password, err = credhelper.NewGCRCredentialHelper(credStore, userCfg).Get(reference.Domain(named))
	if err != nil {
		return "", "", fmt.Errorf("failed to retrieve credentials from the store: %v", err)
	}

	return username, password, nil
}
