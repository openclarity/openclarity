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
	"time"

	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/cisco-open/kubei/api/server/models"
	"github.com/cisco-open/kubei/api/server/restapi/operations"
	"github.com/cisco-open/kubei/backend/pkg/database"
	"github.com/cisco-open/kubei/backend/pkg/types"
	"github.com/cisco-open/kubei/shared/pkg/utils/slice"
)

var oopsResponse = &models.APIResponse{
	Message: "Oops",
}

func (s *Server) CreateApplication(params operations.PostApplicationsParams) middleware.Responder {
	app := database.CreateApplication(params.Body)

	err := s.dbHandler.ApplicationTable().Create(app, &database.TransactionParams{})
	if err != nil {
		// Check if the application already exist.
		appDB, getDBApplicationErr := s.dbHandler.ApplicationTable().GetDBApplication(app.ID, false)
		if getDBApplicationErr == nil {
			log.Errorf(fmt.Sprintf("Application already exist. application=%+v", appDB))
			return operations.NewPostApplicationsConflict().
				WithPayload(createApplicationInfoFromDBApplication(appDB))
		}

		// Unknown error.
		log.Errorf("Failed to create application: %v", err)
		return operations.NewPostApplicationsDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	// Application was created.
	return operations.NewPostApplicationsCreated().
		WithPayload(createApplicationInfoFromDBApplication(app))
}

func (s *Server) DeleteApplication(id string) middleware.Responder {
	// Get application info.
	application, err := s.dbHandler.ApplicationTable().GetDBApplication(id, false)
	if err != nil {
		log.Errorf(fmt.Sprintf("application not found by id %q: %v", id, err))
		return operations.NewDeleteApplicationsIDNotFound()
	}

	// Delete only application<->resources relationships.
	if err = s.dbHandler.JoinTables().DeleteRelationships(database.DeleteRelationshipsParams{
		ApplicationIDsToRemove: []string{application.ID},
	}); err != nil {
		log.Error(fmt.Sprintf("Failed to delete application relationships: %v", err))
		return operations.NewDeleteApplicationsIDDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	// Delete application instance from DB.
	if err = s.dbHandler.ApplicationTable().Delete(application); err != nil {
		log.Error(fmt.Sprintf("Failed to delete application from DB: %v", err))
		return operations.NewDeleteApplicationsIDDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	return operations.NewDeleteApplicationsIDNoContent()
}

func (s *Server) GetApplication(id string) middleware.Responder {
	applicationEx, err := s.dbHandler.ApplicationTable().GetApplication(id)
	if err != nil {
		log.Error(err)
		return operations.NewGetApplicationsIDDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	return operations.NewGetApplicationsIDOK().WithPayload(applicationEx)
}

func (s *Server) UpdateApplication(params operations.PutApplicationsIDParams) middleware.Responder {
	// Get application info.
	application, err := s.dbHandler.ApplicationTable().GetDBApplication(params.ID, false)
	if err != nil {
		log.Errorf("application not found by id %q: %v", params.ID, err)
		return operations.NewPutApplicationsIDNotFound()
	}

	// Update application instance info from DB.
	if err = s.dbHandler.ApplicationTable().UpdateInfo(application.UpdateApplicationInfo(params.Body), &database.TransactionParams{}); err != nil {
		log.Errorf("Failed to update application from DB: %v", err)
		return operations.NewPutApplicationsIDDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	return operations.NewPutApplicationsIDOK().WithPayload(createApplicationInfoFromDBApplication(application))
}

func (s *Server) GetApplications(params operations.GetApplicationsParams) middleware.Responder {
	applicationsView, total, err := s.dbHandler.ApplicationTable().GetApplicationsAndTotal(
		database.GetApplicationsParams{
			GetApplicationsParams:     params,
			RuntimeScanApplicationIDs: s.runtimeScanApplicationIDs,
		})
	if err != nil {
		log.Error(err)
		return operations.NewGetApplicationsDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	log.Debugf("GetApplications controller was invoked. "+
		"params=%+v, applicationsView=%+v, total=%+v", params, applicationsView, total)

	applications := make([]*models.Application, len(applicationsView))
	for i := range applicationsView {
		applications[i] = database.ApplicationFromDB(&applicationsView[i])
	}

	return operations.NewGetApplicationsOK().WithPayload(
		&operations.GetApplicationsOKBody{
			Items: applications,
			Total: &total,
		})
}

func (s *Server) PostApplicationsVulnerabilityScan(params operations.PostApplicationsVulnerabilityScanIDParams) middleware.Responder {
	// Get application info.
	application, err := s.dbHandler.ApplicationTable().GetDBApplication(params.ID, true)
	if err != nil {
		log.Errorf("application not found by id %q: %v", params.ID, err)
		return operations.NewPostApplicationsVulnerabilityScanIDNotFound()
	}

	if err = s.handleApplicationsVulnerabilityScan(application,
		types.ApplicationVulnerabilityScanFromBackendAPI(params.Body),
		models.VulnerabilitySourceCICD); err != nil {
		log.Errorf("failed to handle vulnerability scan report: %v", err)
		return operations.NewPostApplicationsVulnerabilityScanIDDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	return operations.NewPostApplicationsVulnerabilityScanIDCreated()
}

func (s *Server) handleApplicationsVulnerabilityScan(application *database.Application,
	applicationVulnerabilityScan *types.ApplicationVulnerabilityScan, vulnerabilitySource models.VulnerabilitySource) error {
	transactionParams := &database.TransactionParams{
		FixVersions:         make(map[database.PkgVulID]string),        // will be populated during object creation
		Analyzers:           make(map[database.ResourcePkgID][]string), // will be populated during object creation
		Scanners:            make(map[database.ResourcePkgID][]string), // will be populated during object creation
		VulnerabilitySource: vulnerabilitySource,
		Timestamp:           time.Now().UTC(),
	}

	// Create new application tree based on the current application.
	// Note: For now, for each vulnerability scan report we will replace application resources list.
	application = s.updateApplicationWithVulnerabilityScan(application, applicationVulnerabilityScan, transactionParams,
		true)

	// The analyzers map is created during updateApplicationWithVulnerabilityScan() on the transactionParams.
	// Set the updated analyzers map.
	analyzers, err := s.getUpdatedAnalyzersMap(application, transactionParams.Analyzers)
	if err != nil {
		return fmt.Errorf("failed to get updated analyzers map: %v", err)
	}
	transactionParams.Analyzers = analyzers

	// Update analyzers on the application resources
	application.Resources = database.UpdateResourceAnalyzers(application.Resources, analyzers)

	// Create new vulnerabilities trends entries.
	// NOTE: MUST run before `ApplicationTable().Update` call
	// `ApplicationTable().Update` updates newApp structure and delete the associations
	if err = s.dbHandler.NewVulnerabilityTable().CreateNewVulnerabilitiesTrends(application); err != nil {
		return fmt.Errorf("failed to save new vulnerabilities trends: %v", err)
	}

	// Update new application information.
	log.Infof("Updating application. id=%v, name=%v", application.ID, application.Name)
	log.Tracef("Updating application %+v", application)
	if err = s.dbHandler.ObjectTree().SetApplication(application, transactionParams, true); err != nil {
		return fmt.Errorf("failed to update application: %v", err)
	}

	return nil
}

func (s *Server) updateApplicationWithVulnerabilityScan(application *database.Application, applicationVulnerabilityScan *types.ApplicationVulnerabilityScan,
	transactionParams *database.TransactionParams, shouldReplaceResources bool) *database.Application {
	currentResourceIDToIndex := make(map[string]int)
	for i, resource := range application.Resources {
		currentResourceIDToIndex[resource.ID] = i
	}

	// Will hold the new application resource list if shouldReplaceResources is true.
	var newResources []database.Resource

	for _, resource := range applicationVulnerabilityScan.Resources {
		// Create a new resource tree from vulnerability scan.
		newResource := database.CreateResourceFromVulnerabilityScan(resource, transactionParams)

		if i, ok := currentResourceIDToIndex[newResource.ID]; !ok {
			// If a resource does not exist on current application try to fetch from DB.
			dbResource, err := s.dbHandler.ResourceTable().GetDBResource(newResource.ID, true)
			if err == nil {
				// If a resource exist in db, update resource tree with new findings.
				newResource = updateResource(dbResource, newResource, transactionParams)
			} else {
				// New resource, need to update analyzers list.
				updateAnalyzersForNewResource(newResource, transactionParams)
			}

			if shouldReplaceResources {
				// Add resource (resource->packages->vulnerabilities) to the new application resource list.
				newResources = append(newResources, *newResource)
			} else {
				// Add resource (resource->packages->vulnerabilities) to the updated application resource list.
				application.Resources = append(application.Resources, *newResource)
			}
		} else {
			// If a resource exist, update resource tree.
			if shouldReplaceResources {
				// Update resource from existing application resource list and append to new application resource list.
				updatedResource := updateResource(&application.Resources[i], newResource, transactionParams)
				newResources = append(newResources, *updatedResource)
			} else {
				// Update resource on existing application resource list.
				application.Resources[i] = *(updateResource(&application.Resources[i], newResource, transactionParams))
			}
		}
	}

	if shouldReplaceResources {
		// Update application with updated resources list
		application.Resources = newResources
	}

	return application
}

func updateAnalyzersForNewResource(newResource *database.Resource, transactionParams *database.TransactionParams) {
	for _, p := range newResource.Packages {
		// Set scanners list as analyzers list since it's a new resource->packages relationship
		resourcePkgID := database.CreateResourcePkgID(newResource.ID, p.ID)
		transactionParams.Analyzers[resourcePkgID] = transactionParams.Scanners[resourcePkgID]
	}
}

func updateResource(currentResource *database.Resource, newResource *database.Resource, params *database.TransactionParams) *database.Resource {
	currentPackageIDToIndex := make(map[string]int)
	for i, pkg := range currentResource.Packages {
		currentPackageIDToIndex[pkg.ID] = i
	}

	// Go over resource packages and update if needed
	for i, newPkg := range newResource.Packages {
		if currentPkgIndex, ok := currentPackageIDToIndex[newPkg.ID]; !ok {
			// If a package does not exist, create a tree for it (packages->vulnerabilities)
			currentResource.Packages = append(currentResource.Packages, newResource.Packages[i])
			// Set scanners list as analyzers list since it's a new resource->packages relationship
			resourcePkgID := database.CreateResourcePkgID(newResource.ID, newPkg.ID)
			params.Analyzers[resourcePkgID] = params.Scanners[resourcePkgID]
			log.Tracef("New resource->packages found. res-id=%v, pkg-id=%v, scanners=%v",
				newResource.ID, newPkg.ID, params.Scanners[resourcePkgID])
		} else {
			// Create vulnerabilities and packages->vulnerabilities relations
			currentResource.Packages[currentPkgIndex].Vulnerabilities = newPkg.Vulnerabilities
		}
	}

	// Update CIS docker benchmark results only if exists.
	if len(newResource.CISDockerBenchmarkResults) > 0 {
		currentResource.CISDockerBenchmarkResults = newResource.CISDockerBenchmarkResults
	}

	return currentResource
}

// getUpdatedAnalyzersMap will get current ResourcePkgID to analyzers map from DB and merge it to the new analyzers map.
func (s *Server) getUpdatedAnalyzersMap(newApp *database.Application, newResourcePkgIDAnalyzers map[database.ResourcePkgID][]string) (map[database.ResourcePkgID][]string, error) {
	resourceIDs := make([]string, len(newApp.Resources))
	for i, resource := range newApp.Resources {
		resourceIDs[i] = resource.ID
	}

	// Get current ResourcePkgID to analyzers map before updating new app
	currentResourcePkgIDToAnalyzers, err := s.dbHandler.JoinTables().GetResourcePackageIDToAnalyzers(resourceIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get resourcePkgIDToAnalyzers: %v", err)
	}

	log.Tracef("currentResourcePkgIDToAnalyzers=%+v", currentResourcePkgIDToAnalyzers)

	// Merge analyzers on newResourcePkgIDAnalyzers with a current analyzers map
	return mergeResourcePkgIDToAnalyzersMaps(newResourcePkgIDAnalyzers, currentResourcePkgIDToAnalyzers), nil
}

// mergeResourcePkgIDToAnalyzersMaps merge a and b ResourcePkgIDToAnalyzers maps into a unified map.
func mergeResourcePkgIDToAnalyzersMaps(a, b map[database.ResourcePkgID][]string) map[database.ResourcePkgID][]string {
	resourcePkgIDToAnalyzers := make(map[database.ResourcePkgID][]string)
	for id, analyzers := range a {
		resourcePkgIDToAnalyzers[id] = analyzers
	}
	for id, analyzers := range b {
		resourcePkgIDToAnalyzers[id] = slice.RemoveStringDuplicates(append(resourcePkgIDToAnalyzers[id], analyzers...))
	}

	log.Tracef("Merged resourcePkgIDToAnalyzers=%+v", resourcePkgIDToAnalyzers)
	return resourcePkgIDToAnalyzers
}

func (s *Server) PostApplicationsContentAnalysis(params operations.PostApplicationsContentAnalysisIDParams) middleware.Responder {
	application, err := s.dbHandler.ApplicationTable().GetDBApplication(params.ID, false)
	if err != nil {
		log.Errorf(fmt.Sprintf("application not found by id %q: %v", params.ID, err))
		return operations.NewPostApplicationsContentAnalysisIDNotFound()
	}

	// Create new application tree
	transactionParams := &database.TransactionParams{
		Analyzers:           make(map[database.ResourcePkgID][]string), // will be populated during object creation
		VulnerabilitySource: models.VulnerabilitySourceCICD,
	}

	for _, resource := range params.Body.Resources {
		application.Resources = append(application.Resources, *database.CreateResourceFromContentAnalysis(resource, transactionParams))
	}

	// Update new application information.
	// Since it's a content analysis report we should NOT update PackageVulnerabilities relationships.
	if err = s.dbHandler.ObjectTree().SetApplication(application, transactionParams, false); err != nil {
		log.Errorf(fmt.Sprintf("failed to update application: %v", err))
		return operations.NewPostApplicationsContentAnalysisIDDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	return operations.NewPostApplicationsContentAnalysisIDCreated()
}

func createApplicationInfoFromDBApplication(appDB *database.Application) *models.ApplicationInfo {
	return &models.ApplicationInfo{
		ID:           appDB.ID,
		Name:         &appDB.Name,
		Type:         &appDB.Type,
		Environments: database.DBArrayToArray(appDB.Environments),
		Labels:       database.DBArrayToArray(appDB.Labels),
	}
}
