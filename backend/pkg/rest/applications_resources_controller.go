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

package rest

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"

	"wwwin-github.cisco.com/eti/scan-gazr/api/server/models"
	"wwwin-github.cisco.com/eti/scan-gazr/api/server/restapi/operations"
	"wwwin-github.cisco.com/eti/scan-gazr/backend/pkg/database"
)

func (s *Server) GetApplicationResource(id string) middleware.Responder {
	resource, err := s.dbHandler.ResourceTable().GetApplicationResource(id)
	if err != nil {
		log.Error(err)
		return operations.NewGetApplicationResourcesIDDefault(http.StatusInternalServerError).
			WithPayload(
				&models.APIResponse{
					Message: "Oops",
				},
			)
	}

	return operations.NewGetApplicationResourcesIDOK().WithPayload(resource)
}

func (s *Server) GetApplicationResources(params operations.GetApplicationResourcesParams) middleware.Responder {
	resources, total, err := s.dbHandler.ResourceTable().GetApplicationResourcesAndTotal(
		database.GetApplicationResourcesParams{
			GetApplicationResourcesParams: params,
			RuntimeScanApplicationIDs:     s.runtimeScanApplicationIDs,
		})
	if err != nil {
		log.Error(err)
		return operations.NewGetApplicationResourcesDefault(http.StatusInternalServerError).
			WithPayload(
				&models.APIResponse{
					Message: "Oops",
				},
			)
	}

	log.Debugf("GetApplicationResources controller was invoked. "+
		"params=%+v, resources=%+v, total=%+v", params, resources, total)

	applicationResources := make([]*models.ApplicationResource, len(resources))
	for i := range resources {
		applicationResources[i] = database.ApplicationResourceFromDB(&resources[i])
	}

	return operations.NewGetApplicationResourcesOK().WithPayload(
		&operations.GetApplicationResourcesOKBody{
			Items: applicationResources,
			Total: &total,
		})
}
