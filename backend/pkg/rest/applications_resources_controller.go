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

	"github.com/openclarity/kubeclarity/api/server/models"
	"github.com/openclarity/kubeclarity/api/server/restapi/operations"
	"github.com/openclarity/kubeclarity/backend/pkg/database"
	"github.com/openclarity/kubeclarity/shared/pkg/utils/slice"
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

func (s *Server) DeleteApplicationResource(id string) middleware.Responder {
	// Get application resource info.
	resource, err := s.dbHandler.ResourceTable().GetDBResource(id, false)
	if err != nil {
		log.Errorf("application resource not found by id %q: %v", id, err)
		return operations.NewDeleteApplicationResourcesIDNotFound()
	}

	// Get package IDs that has relations with the resource.
	relatedPackageIDs, err := s.getRelatedPackageIDs(resource.ID)
	if err != nil {
		log.Errorf("Failed to get packages for resource ID=%s: %v", resource.ID, err)
		return operations.NewDeleteApplicationResourcesIDDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	// Delete only applications<->resources<->packages relationships.
	if err = s.dbHandler.JoinTables().DeleteRelationships(database.DeleteRelationshipsParams{
		ResourceIDsToRemove: []string{resource.ID},
	}); err != nil {
		log.Error(fmt.Sprintf("Failed to delete resource relationships: %v", err))
		return operations.NewDeleteApplicationResourcesIDDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	// Get list of package IDs that has no any relations to other resources.
	packageIDsToDelete, err := s.getPackagesToDelete(relatedPackageIDs)
	if err != nil {
		log.Errorf("Failed to get remaining packages: %v", err)
		return operations.NewDeleteApplicationResourcesIDDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	// Get list of vulnerability IDs for the packages that needs to be deleted.
	relatedVulnerabilityIDs, err := s.getRelatedVulnerabilityIDs(packageIDsToDelete)
	if err != nil {
		log.Errorf("Failed to get vulnerabilities for packages: %v", err)
		return operations.NewDeleteApplicationResourcesIDDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	// Delete only package<->vulnerabilities relationships.
	if err = s.dbHandler.JoinTables().DeleteRelationships(database.DeleteRelationshipsParams{
		PackageIDsToRemove: packageIDsToDelete,
	}); err != nil {
		log.Errorf("Failed to delete package relationships: %v", err)
		return operations.NewDeleteApplicationResourcesIDDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	// Get list of vulnerability IDs that have no relationship for existing packages.
	vulnerabilityIDsToDelete, err := s.getVulnerabilitiesToDelete(relatedVulnerabilityIDs)
	if err != nil {
		log.Errorf("Failed to get remaining vulnerabilities: %v", err)
		return operations.NewDeleteApplicationResourcesIDDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	if err = s.dbHandler.ResourceTable().Delete(resource); err != nil {
		log.Errorf("Failed to delete resource from DB: %v", err)
		return operations.NewDeleteApplicationResourcesIDDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	s.deletePackages(packageIDsToDelete)
	s.deleteVulnerabilities(vulnerabilityIDsToDelete)

	return operations.NewDeleteApplicationsIDNoContent()
}

func (s *Server) getRelatedPackageIDs(resourceID string) ([]string, error) {
	// Get list of packages related to the resource.
	packagesResourcesByResources, err := s.dbHandler.JoinTables().GetResourcePackagesByResources([]string{resourceID})
	if err != nil {
		return nil, fmt.Errorf("failed to get packages for resource ID=%s: %v", resourceID, err)
	}
	// List of package IDs that have relations to resource
	relatedPackageIDs := []string{}
	for _, pr := range packagesResourcesByResources {
		relatedPackageIDs = append(relatedPackageIDs, pr.PackageID)
	}

	return relatedPackageIDs, nil
}

func (s *Server) getPackagesToDelete(relatedPackageIDs []string) ([]string, error) {
	// Get packagesResources by package IDs from the updated resource_packages table.
	packagesResourcesByPackages, err := s.dbHandler.JoinTables().GetResourcePackagesByPackages(relatedPackageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get remaining packages: %v", err)
	}

	remainingPackageIDs := []string{}
	for _, pr := range packagesResourcesByPackages {
		remainingPackageIDs = append(remainingPackageIDs, pr.PackageID)
	}
	// Get packages that had relation only with the deleted resource.
	// FindUnique returns a list of strings that exist in 'relatedPackageIDs' and not in 'remainingPackageIDs'.
	packageIDsToDelete := slice.FindUnique(relatedPackageIDs, remainingPackageIDs)

	return packageIDsToDelete, nil
}

func (s *Server) getRelatedVulnerabilityIDs(packageIDs []string) ([]string, error) {
	// Get list of vulnerabilities for packages.
	packageVulnerabilitiesByPackages, err := s.dbHandler.JoinTables().GetPackageVulnerabilitiesByPackages(packageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get vulnerabilities for packages: %v", err)
	}
	relatedVulnerabilityIDs := []string{}
	for _, pv := range packageVulnerabilitiesByPackages {
		relatedVulnerabilityIDs = append(relatedVulnerabilityIDs, pv.VulnerabilityID)
	}

	return relatedVulnerabilityIDs, nil
}

func (s *Server) getVulnerabilitiesToDelete(relatedVulnerabilityIDs []string) ([]string, error) {
	// Get packagesVulnerabilities by vulnerability IDs from the updated package_vulnerabilities table.
	packagesVulnerabilitiesByVulnerabilities, err := s.dbHandler.JoinTables().GetPackageVulnerabilitiesByVulnerabilities(relatedVulnerabilityIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get remaining vulnerabilities: %v", err)
	}

	remainingVulnerabilityIDs := []string{}
	for _, pv := range packagesVulnerabilitiesByVulnerabilities {
		remainingVulnerabilityIDs = append(remainingVulnerabilityIDs, pv.VulnerabilityID)
	}
	// Get packages that had relation only with the deleted resource.
	// FindUnique returns a list of strings that exist in 'relatedVulnerabilityIDs' and not in 'remainingVulnerabilityIDs'.
	vulnerabilityIDsToDelete := slice.FindUnique(relatedVulnerabilityIDs, remainingVulnerabilityIDs)

	return vulnerabilityIDsToDelete, nil
}

func (s *Server) deletePackages(packageIDs []string) {
	if err := s.dbHandler.PackageTable().DeleteByIDs(packageIDs); err != nil {
		log.Errorf("Failed to delete packages IDs=%v: %v", packageIDs, err)
	}
}

func (s *Server) deleteVulnerabilities(vulnerabilityIDs []string) {
	if err := s.dbHandler.VulnerabilityTable().DeleteByIDs(vulnerabilityIDs); err != nil {
		log.Errorf("Failed to delete vulnerabilities IDs=%v: %v", vulnerabilityIDs, err)
	}
}
