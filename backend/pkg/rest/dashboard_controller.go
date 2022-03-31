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
	"fmt"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/cisco-open/kubei/api/server/models"
	"github.com/cisco-open/kubei/api/server/restapi/operations"
)

const mostVulnerableLimit = 5

func (s *Server) GetDashboardVulnerabilitiesWithFix(_ operations.GetDashboardVulnerabilitiesWithFixParams) middleware.Responder {
	vulsCount, err := s.dbHandler.VulnerabilityTable().CountVulnerabilitiesWithFix()
	if err != nil {
		log.Errorf("Failed to count vulnerabilities with a fix: %v", err)
		return operations.NewGetDashboardVulnerabilitiesWithFixDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	return operations.NewGetDashboardVulnerabilitiesWithFixOK().WithPayload(vulsCount)
}

func (s *Server) GetDashboardPackagesPerLanguage(_ operations.GetDashboardPackagesPerLanguageParams) middleware.Responder {
	pkgCount, err := s.dbHandler.PackageTable().GetPackagesCountPerLanguage()
	if err != nil {
		log.Errorf("Failed to get package count per language: %v", err)
		return operations.NewGetDashboardPackagesPerLanguageDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	return operations.NewGetDashboardPackagesPerLanguageOK().WithPayload(pkgCount)
}

func (s *Server) GetDashboardPackagesPerLicense(_ operations.GetDashboardPackagesPerLicenseParams) middleware.Responder {
	pkgCount, err := s.dbHandler.PackageTable().GetPackagesCountPerLicense()
	if err != nil {
		log.Errorf("Failed to get package count per license: %v", err)
		return operations.NewGetDashboardPackagesPerLicenseDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	return operations.NewGetDashboardPackagesPerLicenseOK().WithPayload(pkgCount)
}

func (s *Server) getDashboardCounters() (*models.DashboardCounters, error) {
	pkgCount, err := s.dbHandler.PackageTable().Count(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to count packages: %v", err)
	}
	appCount, err := s.dbHandler.ApplicationTable().Count(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to count applications: %v", err)
	}
	resCount, err := s.dbHandler.ResourceTable().Count(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to count resources: %v", err)
	}

	return &models.DashboardCounters{
		Applications: uint32(appCount),
		Packages:     uint32(pkgCount),
		Resources:    uint32(resCount),
	}, nil
}

func (s *Server) GetDashboardCounters(_ operations.GetDashboardCountersParams) middleware.Responder {
	counters, err := s.getDashboardCounters()
	if err != nil {
		log.Errorf("Failed to get dashboard counters: %v", err)
		return operations.NewGetDashboardCountersDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	return operations.NewGetDashboardCountersOK().WithPayload(counters)
}

func (s *Server) getMostVulnerable() (*models.MostVulnerable, error) {
	apps, err := s.dbHandler.ApplicationTable().GetMostVulnerable(mostVulnerableLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get most vulnerable applications: %v", err)
	}
	pkgs, err := s.dbHandler.PackageTable().GetMostVulnerable(mostVulnerableLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get most vulnerable packages: %v", err)
	}
	res, err := s.dbHandler.ResourceTable().GetMostVulnerable(mostVulnerableLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get most vulnerable resources: %v", err)
	}

	return &models.MostVulnerable{
		Applications: apps,
		Packages:     pkgs,
		Resources:    res,
	}, nil
}

func (s *Server) GetDashboardMostVulnerable(_ operations.GetDashboardMostVulnerableParams) middleware.Responder {
	mostVulnerable, err := s.getMostVulnerable()
	if err != nil {
		log.Errorf("Failed to get most vulnerable: %v", err)
		return operations.NewGetDashboardMostVulnerableDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	return operations.NewGetDashboardMostVulnerableOK().WithPayload(mostVulnerable)
}

func (s *Server) GetDashboardTrendsVulnerabilities(params operations.GetDashboardTrendsVulnerabilitiesParams) middleware.Responder {
	trends, err := s.dbHandler.NewVulnerabilityTable().GetNewVulnerabilitiesTrends(params)
	if err != nil {
		log.Errorf("Failed to get new vulnerabilities trends: %v", err)
		return operations.NewGetDashboardTrendsVulnerabilitiesDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	return operations.NewGetDashboardTrendsVulnerabilitiesOK().WithPayload(trends)
}
