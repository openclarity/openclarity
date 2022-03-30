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

func (s *Server) GetVulnerabilitiesVulIDPkgID(vulID, pkgID string) middleware.Responder {
	vul, err := s.dbHandler.VulnerabilityTable().GetVulnerability(vulID, pkgID)
	if err != nil {
		log.Error(err)
		return operations.NewGetVulnerabilitiesVulIDPkgIDDefault(http.StatusInternalServerError).
			WithPayload(
				&models.APIResponse{
					Message: "Oops",
				},
			)
	}

	return operations.NewGetVulnerabilitiesVulIDPkgIDOK().WithPayload(vul)
}

func (s *Server) GetVulnerabilities(params operations.GetVulnerabilitiesParams) middleware.Responder {
	vulnerabilitiesView, total, err := s.dbHandler.VulnerabilityTable().GetVulnerabilitiesAndTotal(
		database.GetVulnerabilitiesParams{
			GetVulnerabilitiesParams:  params,
			RuntimeScanApplicationIDs: s.runtimeScanApplicationIDs,
		})
	if err != nil {
		log.Error(err)
		return operations.NewGetVulnerabilitiesDefault(http.StatusInternalServerError).
			WithPayload(
				&models.APIResponse{
					Message: "Oops",
				},
			)
	}

	log.Debugf("GetVulnerabilities controller was invoked. "+
		"params=%+v, vulnerabilitiesView=%+v, total=%+v", params, vulnerabilitiesView, total)

	vuls := make([]*models.Vulnerability, len(vulnerabilitiesView))
	for i := range vulnerabilitiesView {
		vuls[i] = database.VulnerabilityFromDB(&vulnerabilitiesView[i])
	}

	return operations.NewGetVulnerabilitiesOK().WithPayload(
		&operations.GetVulnerabilitiesOKBody{
			Items: vuls,
			Total: &total,
		})
}
