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

package database

import (
	"context"
	"fmt"
	"strings"

	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/cisco-open/kubei/api/server/models"
	"github.com/cisco-open/kubei/api/server/restapi/operations"
	"github.com/cisco-open/kubei/shared/pkg/utils/slice"
)

const (
	applicationTableName = "applications"
	applicationViewName  = "applications_view"

	// NOTE: when changing one of the column names change also the gorm label in Application.
	columnAppID           = "id"
	columnAppName         = "name"
	columnAppType         = "type"
	columnAppLabels       = "labels"
	columnAppEnvironments = "environments"

	// NOTE: when changing one of the column names change also the gorm label in ApplicationView.
	columnApplicationViewResources = "resources"
	columnApplicationViewPackages  = "packages"
)

type Application struct {
	ID string `gorm:"primarykey" faker:"-"` // consists of the application name

	Name         string                 `json:"name,omitempty" gorm:"column:name"`
	Type         models.ApplicationType `json:"type,omitempty" gorm:"column:type" faker:"oneof: IMAGE, DIRECTORY, FILE"`
	Labels       string                 `json:"labels,omitempty" gorm:"column:labels" faker:"oneof: |label1|, |label1||label2|, |label1||label2||label3|"`
	Environments string                 `json:"environments,omitempty" gorm:"column:environments" faker:"oneof: |env1|, |env1||env2|, |env1||env2||env3|"`
	Resources    []Resource             `json:"resources,omitempty" gorm:"many2many:application_resources;" faker:"-"`
}

type ApplicationView struct {
	Application
	Resources int `json:"resources,omitempty" gorm:"column:resources"`
	Packages  int `json:"packages,omitempty" gorm:"column:packages"`
	SeverityCounters
	CISDockerBenchmarkLevelCounters
}

type GetApplicationsParams struct {
	operations.GetApplicationsParams
	// List of application IDs that were affected by the last runtime scan.
	RuntimeScanApplicationIDs []string
}

type ApplicationTable interface {
	Create(app *Application, params *TransactionParams) error
	UpdateInfo(app *Application, params *TransactionParams) error
	Delete(app *Application) error
	GetApplicationsAndTotal(params GetApplicationsParams) ([]ApplicationView, int64, error)
	GetApplication(id string) (*models.ApplicationEx, error)
	GetDBApplication(id string, shouldGetRelationships bool) (*Application, error)
	Count(filters *CountFilters) (int64, error)
	GetMostVulnerable(limit int) ([]*models.Application, error)
}

type ApplicationTableHandler struct {
	applicationsTable *gorm.DB
	applicationsView  *gorm.DB
	licensesView      *gorm.DB
	IDsView           IDsView
	db                *gorm.DB
}

func (Application) TableName() string {
	return applicationTableName
}

func CreateApplication(app *models.ApplicationInfo) *Application {
	return &Application{
		ID:           CreateApplicationID(app),
		Name:         *app.Name,
		Type:         *app.Type,
		Labels:       ArrayToDBArray(app.Labels),
		Environments: ArrayToDBArray(app.Environments),
	}
}

func (a *Application) UpdateApplicationInfo(app *models.ApplicationInfo) *Application {
	a.Name = *app.Name
	a.Type = *app.Type
	a.Labels = ArrayToDBArray(app.Labels)
	a.Environments = ArrayToDBArray(app.Environments)
	return a
}

func CreateApplicationID(app *models.ApplicationInfo) string {
	return uuid.NewV5(uuid.Nil, *app.Name).String()
}

func (a *ApplicationTableHandler) Create(app *Application, params *TransactionParams) error {
	if err := a.applicationsTable.
		WithContext(createContextWithValues(params)).
		Create(app).Error; err != nil {
		return fmt.Errorf("failed to create application: %v", err)
	}

	return nil
}

func (a *ApplicationTableHandler) UpdateInfo(app *Application, params *TransactionParams) error {
	if err := a.db.WithContext(createContextWithValues(params)).
		Omit("Resources").Updates(app).Error; err != nil {
		return fmt.Errorf("failed to update application info: %v", err)
	}

	return nil
}

// nolint:staticcheck
func createContextWithValues(params *TransactionParams) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, fixVersionsContextValueName, params.FixVersions)
	ctx = context.WithValue(ctx, analyzersContextValueName, params.Analyzers)
	return ctx
}

func (a *ApplicationTableHandler) Delete(app *Application) error {
	if err := a.db.Delete(app).Error; err != nil {
		return fmt.Errorf("failed to delete application: %v", err)
	}

	return nil
}

func (a *ApplicationTableHandler) setApplicationsFilters(params GetApplicationsParams) (*gorm.DB, error) {
	tx := a.applicationsView

	// name filters
	tx = FilterIs(tx, columnAppName, params.ApplicationNameIs)
	tx = FilterIsNot(tx, columnAppName, params.ApplicationNameIsNot)
	tx = FilterContains(tx, columnAppName, params.ApplicationNameContains)
	tx = FilterStartsWith(tx, columnAppName, params.ApplicationNameStart)
	tx = FilterEndsWith(tx, columnAppName, params.ApplicationNameEnd)

	// type filter
	tx = FilterIs(tx, columnAppType, params.ApplicationTypeIs)

	// labels filters
	tx = FilterArrayContains(tx, columnAppLabels, params.ApplicationLabelsContainElements)
	tx = FilterArrayDoesntContain(tx, columnAppLabels, params.ApplicationLabelsDoesntContainElements)

	// envs filters
	tx = FilterArrayContains(tx, columnAppEnvironments, params.ApplicationEnvsContainElements)
	tx = FilterArrayDoesntContain(tx, columnAppEnvironments, params.ApplicationEnvsDoesntContainElements)

	// resources filter
	tx = FilterIsNumber(tx, columnApplicationViewResources, params.ApplicationResourcesIs)
	tx = FilterIsNotNumber(tx, columnApplicationViewResources, params.ApplicationResourcesIsNot)
	tx = FilterGte(tx, columnApplicationViewResources, params.ApplicationResourcesGte)
	tx = FilterLte(tx, columnApplicationViewResources, params.ApplicationResourcesLte)

	// packages filter
	tx = FilterIsNumber(tx, columnApplicationViewPackages, params.PackagesIs)
	tx = FilterIsNotNumber(tx, columnApplicationViewPackages, params.PackagesIsNot)
	tx = FilterGte(tx, columnApplicationViewPackages, params.PackagesGte)
	tx = FilterLte(tx, columnApplicationViewPackages, params.PackagesLte)

	// vulnerabilities filter
	tx = SeverityFilterGte(tx, columnSeverityCountersHighestSeverity, params.VulnerabilitySeverityGte)
	tx = SeverityFilterLte(tx, columnSeverityCountersHighestSeverity, params.VulnerabilitySeverityLte)

	// cis docker benchmark filter
	tx = CISDockerBenchmarkLevelFilterGte(tx, columnCISDockerBenchmarkLevelCountersHighestLevel, params.CisDockerBenchmarkLevelGte)
	tx = CISDockerBenchmarkLevelFilterLte(tx, columnCISDockerBenchmarkLevelCountersHighestLevel, params.CisDockerBenchmarkLevelGte)

	// system filter
	ids, err := a.getApplicationIDs(params)
	if err != nil {
		return nil, fmt.Errorf("failed to get applications IDs: %v", err)
	}
	tx = FilterIs(tx, columnAppID, ids)

	return tx, nil
}

func ApplicationFromDB(view *ApplicationView) *models.Application {
	return &models.Application{
		ApplicationName:           view.Name,
		ApplicationResources:      uint32(view.Resources),
		ApplicationType:           view.Type,
		CisDockerBenchmarkResults: getCISDockerBenchmarkLevelCount(view.CISDockerBenchmarkLevelCounters),
		Environments:              DBArrayToArray(view.Environments),
		ID:                        view.ID,
		Labels:                    DBArrayToArray(view.Labels),
		Packages:                  uint32(view.Packages),
		Vulnerabilities:           getVulnerabilityCount(view.SeverityCounters),
	}
}

func (a *ApplicationTableHandler) GetDBApplication(id string, shouldGetRelationships bool) (*Application, error) {
	var application Application

	tx := a.applicationsTable.
		Where(applicationTableName+"."+columnAppID+" = ?", id)

	if shouldGetRelationships {
		tx.Preload("Resources.Packages.Vulnerabilities").Preload(clause.Associations)
	}

	if err := tx.First(&application).Error; err != nil {
		return nil, fmt.Errorf("failed to get application by id %q: %v", id, err)
	}

	return &application, nil
}

func (a *ApplicationTableHandler) GetApplication(id string) (*models.ApplicationEx, error) {
	var applicationView ApplicationView

	if err := a.applicationsView.Where(applicationViewName+"."+columnAppID+" = ?", id).First(&applicationView).Error; err != nil {
		return nil, fmt.Errorf("failed to get application by id %q: %v", id, err)
	}

	licenses, err := a.getLicenses(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get licenses by id %q: %v", id, err)
	}

	return &models.ApplicationEx{
		Application: ApplicationFromDB(&applicationView),
		Licenses:    licenses,
	}, nil
}

func (a *ApplicationTableHandler) GetMostVulnerable(limit int) ([]*models.Application, error) {
	tx := a.applicationsView

	var views []ApplicationView

	sortOrder, err := createVulnerabilitiesColumnSortOrder("desc")
	if err != nil {
		return nil, fmt.Errorf("failed to create sort order: %v", err)
	}

	if err := tx.Order(sortOrder).Limit(limit).Scan(&views).Error; err != nil {
		return nil, err
	}

	apps := make([]*models.Application, len(views))
	for i := range views {
		apps[i] = ApplicationFromDB(&views[i])
	}

	return apps, nil
}

func (a *ApplicationTableHandler) Count(filters *CountFilters) (int64, error) {
	var count int64

	tx := a.setCountFilters(a.applicationsView, filters)

	if err := tx.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count total applications: %v", err)
	}
	return count, nil
}

func (a *ApplicationTableHandler) GetApplicationsAndTotal(params GetApplicationsParams) ([]ApplicationView, int64, error) {
	var count int64
	var applications []ApplicationView

	tx, err := a.setApplicationsFilters(params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to set filters: %v", err)
	}

	// get total item count with the set filters
	if err := tx.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count total: %v", err)
	}

	sortOrder, err := createApplicationsSortOrder(params.SortKey, params.SortDir)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create sort order: %v", err)
	}

	// get specific page ordered items with the current filters
	if err := tx.Scopes(Paginate(params.Page, params.PageSize)).
		Order(sortOrder).
		Find(&applications).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find applications: %v", err)
	}

	return applications, count, nil
}

func createApplicationsSortOrder(sortKey string, sortDir *string) (string, error) {
	switch models.ApplicationsSortKey(sortKey) {
	case models.ApplicationsSortKeyVulnerabilities:
		return createVulnerabilitiesColumnSortOrder(*sortDir)
	case models.ApplicationsSortKeyCisDockerBenchmarkResults:
		return createCISDockerBenchmarkResultsColumnSortOrder(*sortDir)
	default:
		sortKeyColumnName, err := getApplicationsSortKeyColumnName(sortKey)
		if err != nil {
			return "", fmt.Errorf("failed to get sort key column name: %v", err)
		}

		return fmt.Sprintf("%v %v", sortKeyColumnName, strings.ToLower(*sortDir)), nil
	}
}

func getApplicationsSortKeyColumnName(key string) (string, error) {
	switch models.ApplicationsSortKey(key) {
	case models.ApplicationsSortKeyApplicationName:
		return columnAppName, nil
	case models.ApplicationsSortKeyApplicationType:
		return columnAppType, nil
	case models.ApplicationsSortKeyApplicationResources:
		return columnApplicationViewResources, nil
	case models.ApplicationsSortKeyPackages:
		return columnApplicationViewPackages, nil
	case models.ApplicationsSortKeyVulnerabilities:
		return "", fmt.Errorf("unsupported key (%v)", key)
	}

	return "", fmt.Errorf("unknown sort key (%v)", key)
}

func (a *ApplicationTableHandler) getApplicationIDs(params GetApplicationsParams) ([]string, error) {
	if params.CurrentRuntimeScan != nil && *params.CurrentRuntimeScan {
		return params.RuntimeScanApplicationIDs, nil
	}

	ids, err := a.IDsView.GetIDs(a.createGetIDsParams(params), true)
	if err != nil {
		return nil, fmt.Errorf("failed to get IDs: %v", err)
	}

	return ids, nil
}

func (a *ApplicationTableHandler) createGetIDsParams(params GetApplicationsParams) GetIDsParams {
	retParams := GetIDsParams{
		LookupIDType: ApplicationIDType,
	}

	// system filters - only one is allowed
	if params.ApplicationResourceID != nil {
		retParams.FilterIDType = ResourceIDType
		retParams.FilterIDs = []string{*params.ApplicationResourceID}
	} else if params.PackageID != nil {
		retParams.FilterIDType = PackageIDType
		retParams.FilterIDs = []string{*params.PackageID}
	}

	return retParams
}

func (a *ApplicationTableHandler) getLicenses(id string) ([]string, error) {
	var licenses []string

	if err := a.licensesView.
		Select("distinct "+columnLicensesViewLicense).
		Where(columnLicensesViewApplicationID+" = ?", id).
		Find(&licenses).Error; err != nil {
		return nil, fmt.Errorf("failed to get licenses: %v", err)
	}

	licenses = slice.RemoveEmptyStrings(licenses)
	return licenses, nil
}

func (a *ApplicationTableHandler) setCountFilters(tx *gorm.DB, filters *CountFilters) *gorm.DB {
	if filters == nil {
		return tx
	}

	tx = FilterIs(tx, columnAppID, filters.ApplicationIDs)

	tx = SeverityFilterGte(tx, columnSeverityCountersHighestSeverity, filters.VulnerabilitySeverityGte)

	return tx
}
