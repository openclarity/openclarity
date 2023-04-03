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

package sbomdb

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/kubeclarity/sbom_db/api/client/client"
	"github.com/openclarity/kubeclarity/sbom_db/api/client/client/operations"
	"github.com/openclarity/kubeclarity/shared/pkg/utils/gzip"
)

type Getter interface {
	// Get will return a decoded and uncompressed sbom.
	Get(ctx context.Context, imageHash string) ([]byte, error)
	// GetCompressed will return an encoded and compressed sbom.
	GetCompressed(ctx context.Context, imageHash string) (string, error)
}

type GetterImpl struct {
	client *client.KubeClaritySBOMDBAPIs
}

func createGetter(client *client.KubeClaritySBOMDBAPIs) Getter {
	return &GetterImpl{
		client: client,
	}
}

func (g *GetterImpl) Get(ctx context.Context, imageHash string) ([]byte, error) {
	compressedSbom, err := g.GetCompressed(ctx, imageHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get compress data: %v", err)
	}

	uncompressedSbom, err := gzip.DecodeAndUncompress(compressedSbom)
	if err != nil {
		return nil, fmt.Errorf("failed to uncompress data: %v", err)
	}

	return uncompressedSbom, nil
}

func (g *GetterImpl) GetCompressed(ctx context.Context, imageHash string) (string, error) {
	params := operations.NewGetSbomDBResourceHashParams().
		WithResourceHash(imageHash).
		WithContext(ctx)
	sbom, err := g.client.Operations.GetSbomDBResourceHash(params)
	if err != nil {
		// nolint:errorlint
		switch err.(type) {
		case *operations.GetSbomDBResourceHashNotFound:
			log.Infof("SBOM for image hash %q was not found", imageHash)
			return "", nil
		default:
			return "", fmt.Errorf("failed to get sbom: %v", err)
		}
	}

	return sbom.Payload.Sbom, nil
}
