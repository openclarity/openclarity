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
	"errors"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"wwwin-github.cisco.com/eti/scan-gazr/api/server/models"
	"wwwin-github.cisco.com/eti/scan-gazr/api/server/restapi/operations"
	"wwwin-github.cisco.com/eti/scan-gazr/backend/pkg/database"
)

func (s *Server) GetPackages(params operations.GetPackagesParams) middleware.Responder {
	packagesView, total, err := s.dbHandler.PackageTable().GetPackagesAndTotal(
		database.GetPackagesParams{
			GetPackagesParams:         params,
			RuntimeScanApplicationIDs: s.runtimeScanApplicationIDs,
		})
	if err != nil {
		log.Error(err)
		return operations.NewGetPackagesDefault(http.StatusInternalServerError).
			WithPayload(
				&models.APIResponse{
					Message: "Oops",
				},
			)
	}

	log.Debugf("GetPackages controller was invoked. "+
		"params=%+v, packagesView=%+v, total=%+v", params, packagesView, total)

	packages := make([]*models.Package, len(packagesView))
	for i := range packagesView {
		packages[i] = database.PackageFromDB(&packagesView[i])
	}

	return operations.NewGetPackagesOK().WithPayload(
		&operations.GetPackagesOKBody{
			Items: packages,
			Total: &total,
		})
}

func (s *Server) GetPackage(id string) middleware.Responder {
	pkg, err := s.dbHandler.PackageTable().GetPackage(id)
	if err != nil {
		log.Error(err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return operations.NewGetPackagesIDNotFound()
		}

		return operations.NewGetPackagesIDDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	return operations.NewGetPackagesIDOK().WithPayload(pkg)
}

func (s *Server) GetPackageApplicationResources(params operations.GetPackagesIDApplicationResourcesParams) middleware.Responder {
	packageResources, total, err := s.dbHandler.JoinTables().GetPackageResourcesAndTotal(params)
	if err != nil {
		log.Error(err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return operations.NewGetPackagesIDApplicationResourcesNotFound()
		}

		return operations.NewGetPackagesIDApplicationResourcesDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	log.Debugf("GetPackageApplicationResources controller was invoked. "+
		"params=%+v, packageResources=%+v, total=%+v", params, packageResources, total)

	packageApplicationResources := make([]*models.PackageApplicationResources, len(packageResources))
	for i := range packageResources {
		packageApplicationResources[i] = database.PackageApplicationResourcesFromDB(&packageResources[i])
	}

	return operations.NewGetPackagesIDApplicationResourcesOK().WithPayload(
		&operations.GetPackagesIDApplicationResourcesOKBody{
			Items: packageApplicationResources,
			Total: &total,
		})
}
