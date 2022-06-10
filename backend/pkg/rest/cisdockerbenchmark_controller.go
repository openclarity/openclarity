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

	"github.com/openclarity/kubeclarity/api/server/models"
	"github.com/openclarity/kubeclarity/api/server/restapi/operations"
	"github.com/openclarity/kubeclarity/backend/pkg/database"
)

func (s *Server) GetCISDockerBenchmarkResults(params operations.GetCisdockerbenchmarkresultsIDParams) middleware.Responder {
	cisDockerBenchmarkChecks, total, err := s.dbHandler.CISDockerBenchmarkResultTable().GetCISDockerBenchmarkResultsAndTotal(params)
	if err != nil {
		log.Error(err)
		return operations.NewGetCisdockerbenchmarkresultsIDDefault(http.StatusInternalServerError).
			WithPayload(
				&models.APIResponse{
					Message: "Oops",
				},
			)
	}

	log.Debugf("GetCISDockerBenchmarkResults controller was invoked. "+
		"params=%+v, cisDockerBenchmarkResults=%+v, total=%+v", params, cisDockerBenchmarkChecks, total)

	cisDockerBenchmarkResults := make([]*models.CISDockerBenchmarkResultsEX, len(cisDockerBenchmarkChecks))
	for i := range cisDockerBenchmarkResults {
		cisDockerBenchmarkResults[i] = database.CISDockerBenchmarkResultFromDB(&cisDockerBenchmarkChecks[i])
	}

	return operations.NewGetCisdockerbenchmarkresultsIDOK().WithPayload(
		&operations.GetCisdockerbenchmarkresultsIDOKBody{
			Items: cisDockerBenchmarkResults,
			Total: &total,
		})
}
