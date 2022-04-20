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
	log "github.com/sirupsen/logrus"

	"github.com/cisco-open/kubei/shared/pkg/utils/creds/ecr"
	"github.com/cisco-open/kubei/shared/pkg/utils/creds/gcr"
	"github.com/cisco-open/kubei/shared/pkg/utils/creds/secret"
	"github.com/cisco-open/kubei/shared/pkg/utils/image"
)

type CredExtractor struct {
	extractors []Extractor
}

type Extractor interface {
	// Name Prints the name of the extractor.
	Name() string
	// IsSupported Returns true if extractor is supported for extracting credentials for the given image.
	IsSupported(named reference.Named) bool
	// GetCredentials Returns the proper credentials for the given image.
	GetCredentials(ctx context.Context, named reference.Named) (username, password string, err error)
}

func CreateCredExtractor() *CredExtractor {
	return &CredExtractor{
		extractors: []Extractor{
			// Note: ImagePullSecret must be first
			&secret.ImagePullSecret{},
			&gcr.GCR{},
			&ecr.ECR{},
		},
	}
}

func (c *CredExtractor) GetCredentials(ctx context.Context, named reference.Named) (username, password string, err error) {
	// Found the matched extractor and get credential
	for _, extractor := range c.extractors {
		if !extractor.IsSupported(named) {
			continue
		}

		username, password, err = extractor.GetCredentials(ctx, named)
		if err != nil {
			log.Debugf("Failed to get credentials. image name=%v, extractor=%v: %v", named.String(), extractor.Name(), err)
			continue
		}

		// Verify that no empty username and/or password retrieved.
		if username == "" || password == "" {
			log.Debugf("Credentials not found by extractor. image name=%v, extractor=%v", named.String(), extractor.Name())
			continue
		}

		log.Debugf("Credentials found. image name=%v, extractor=%v", named.String(), extractor.Name())
		return username, password, nil
	}

	log.Debugf("Credentials not found. image name=%v.", named.String())
	return "", "", nil
}

func ExtractCredentials(imageName string) (username string, password string, err error) {
	ref, err := image.GetImageRef(imageName)
	if err != nil {
		return "", "", fmt.Errorf("failed to get image ref. image name=%v: %v", imageName, err)
	}

	credExtractor := CreateCredExtractor()
	if username, password, err = credExtractor.GetCredentials(context.Background(), ref); err != nil {
		return "", "", fmt.Errorf("failed to get credentials. image name=%v: %v", imageName, err)
	}

	return username, password, nil
}
