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

	"github.com/openclarity/kubeclarity/sbom_db/api/client/client"
	"github.com/openclarity/kubeclarity/sbom_db/api/client/client/operations"
	"github.com/openclarity/kubeclarity/sbom_db/api/client/models"
	"github.com/openclarity/kubeclarity/shared/pkg/utils/gzip"
)

type Setter interface {
	Set(ctx context.Context, imageHash string, sbom []byte) error
}

type SetterImpl struct {
	client *client.KubeClaritySBOMDBAPIs
}

func createSetter(client *client.KubeClaritySBOMDBAPIs) Setter {
	return &SetterImpl{
		client: client,
	}
}

func (g *SetterImpl) Set(ctx context.Context, imageHash string, sbom []byte) error {
	base64CompressedSbom, err := gzip.CompressAndEncode(sbom)
	if err != nil {
		return fmt.Errorf("failed to compress data: %v", err)
	}

	params := operations.NewPutSbomDBResourceHashParams().
		WithResourceHash(imageHash).
		WithBody(&models.SBOM{
			Sbom: base64CompressedSbom,
		}).
		WithContext(ctx)
	if _, err := g.client.Operations.PutSbomDBResourceHash(params); err != nil {
		return fmt.Errorf("failed to set sbom: %v", err)
	}

	return nil
}
