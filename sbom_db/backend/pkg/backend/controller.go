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

package backend

import (
	"errors"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"gorm.io/gorm"

	"github.com/openclarity/kubeclarity/sbom_db/api/server/models"
	"github.com/openclarity/kubeclarity/sbom_db/api/server/restapi/operations"
	"github.com/openclarity/kubeclarity/sbom_db/backend/pkg/database"
)

var oopsResponse = &models.APIResponse{
	Message: "Oops",
}

// GetSbomDBResourceHash Get SBOM from DB by resource hash.
func (s *Server) GetSbomDBResourceHash(params operations.GetSbomDBResourceHashParams) middleware.Responder {
	sbom, err := s.dbHandler.SBOMTable().GetSBOM(params.ResourceHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return operations.NewGetSbomDBResourceHashNotFound()
		}
		return operations.NewGetSbomDBResourceHashDefault(http.StatusInternalServerError).WithPayload(oopsResponse)
	}

	return operations.NewGetSbomDBResourceHashOK().WithPayload(&models.SBOM{
		Sbom: sbom.SBOM,
	})
}

// PutSbomDBResourceHash Put SBOM in DB by resource hash. If resource hash already exists, replace the old SBOM with the new one.
func (s *Server) PutSbomDBResourceHash(params operations.PutSbomDBResourceHashParams) middleware.Responder {
	sbom := database.SBOM{
		ID:           params.ResourceHash,
		ResourceHash: params.ResourceHash,
		SBOM:         params.Body.Sbom,
	}

	if err := s.dbHandler.SBOMTable().CreateOrUpdateSBOM(&sbom); err != nil {
		return operations.NewPutSbomDBResourceHashDefault(http.StatusInternalServerError).WithPayload(oopsResponse)
	}

	return operations.NewPutSbomDBResourceHashCreated().WithPayload(&models.SuccessResponse{Message: "Success"})
}
