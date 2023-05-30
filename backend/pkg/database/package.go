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
	"fmt"
	"strings"

	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"

	"github.com/openclarity/kubeclarity/api/server/models"
	"github.com/openclarity/kubeclarity/api/server/restapi/operations"
	"github.com/openclarity/kubeclarity/backend/pkg/types"
	runtime_scan_models "github.com/openclarity/kubeclarity/runtime_scan/api/server/models"
)

const (
	packageTableName = "packages"
	packageViewName  = "packages_view"

	// NOTE: when changing one of the column names change also the gorm label in Package.
	columnPkgID       = "id"
	columnPkgName     = "name"
	columnPkgVersion  = "version"
	columnPkgLicense  = "license"
	columnPkgLanguage = "language"

	// NOTE: when changing one of the column names change also the gorm label in PackageView.
	columnPkgViewApplications = "applications"
	columnPkgViewResources    = "resources"
)

type Package struct {
	ID string `gorm:"primarykey" faker:"-"` // consists of the package name + version

	Name            string          `json:"name,omitempty" gorm:"column:name" faker:"oneof: pkg1, pkg2, pkg3"`
	Version         string          `json:"version,omitempty" gorm:"column:version" faker:"oneof: v1, v2, v3"`
	License         string          `json:"license,omitempty" gorm:"column:license" faker:"oneof: MIT, , Apache 2.0"`
	Language        string          `json:"language,omitempty" gorm:"column:language" faker:"oneof: go, , java, python"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities,omitempty" gorm:"many2many:package_vulnerabilities" faker:"-"`
}

type PackageView struct {
	Package
	Applications int `json:"applications,omitempty" gorm:"column:applications"`
	Resources    int `json:"resources,omitempty" gorm:"column:resources"`
	SeverityCounters
}

type GetPackagesParams struct {
	operations.GetPackagesParams
	// List of application IDs that were affected by the last runtime scan.
	RuntimeScanApplicationIDs []string
}

type PackageTable interface {
	Create(pkg *Package) error
	GetPackagesAndTotal(params GetPackagesParams) ([]PackageView, int64, error)
	GetPackage(id string) (*models.Package, error)
	GetPackagesCountPerLanguage() ([]*models.PackagesCountPerLanguage, error)
	GetPackagesCountPerLicense() ([]*models.PackagesCountPerLicense, error)
	Count(filters *CountFilters) (int64, error)
	GetMostVulnerable(limit int) ([]*models.Package, error)
	DeleteByIDs(pkgIDs []string) error
	GetDBPackage(id string) (*Package, error)
}

type PackageTableHandler struct {
	packagesTable      *gorm.DB
	packagesView       *gorm.DB
	IDsView            IDsView
	viewRefreshHandler *ViewRefreshHandler
}

func (Package) TableName() string {
	return packageTableName
}

func CreatePackageFromContentAnalysis(pkgInfo *models.PackageInfo) *Package {
	return CreatePackage(types.PkgFromBackendAPI(pkgInfo), nil)
}

func CreatePackageFromRuntimeContentAnalysis(pkgInfo *runtime_scan_models.PackageInfo) *Package {
	return CreatePackage(types.PkgFromRuntimeScan(pkgInfo), nil)
}

func CreatePackage(pkg *types.PackageInfo, vuls []Vulnerability) *Package {
	return &Package{
		ID:              CreatePackageID(pkg),
		Name:            pkg.Name,
		Version:         pkg.Version,
		License:         pkg.License,
		Language:        pkg.Language,
		Vulnerabilities: vuls,
	}
}

func CreatePackageID(pkgInfo *types.PackageInfo) string {
	return uuid.NewV5(uuid.Nil, fmt.Sprintf("%s.%s", pkgInfo.Name, pkgInfo.Version)).String()
}

func (p *PackageTableHandler) Create(pkg *Package) error {
	if err := p.packagesTable.Create(pkg).Error; err != nil {
		return fmt.Errorf("failed to create package: %v", err)
	}

	p.viewRefreshHandler.TableChanged(packageTableName)

	return nil
}

func (p *PackageTableHandler) DeleteByIDs(pkgIDs []string) error {
	if len(pkgIDs) == 0 {
		return nil
	}
	if err := p.packagesTable.Delete(&Package{}, pkgIDs).Error; err != nil {
		return fmt.Errorf("failed to delete packages by ID: %v", err)
	}

	p.viewRefreshHandler.TableChanged(packageTableName)

	return nil
}

func (p *PackageTableHandler) GetDBPackage(id string) (*Package, error) {
	var pkg Package

	tx := p.packagesTable.
		Where(packageTableName+"."+columnPkgID+" = ?", id)

	if err := tx.First(&pkg).Error; err != nil {
		return nil, fmt.Errorf("failed to get package by id %q: %v", id, err)
	}

	return &pkg, nil
}

func (p *PackageTableHandler) setPackagesFilters(params GetPackagesParams) (*gorm.DB, error) {
	tx := p.packagesView

	// name filters
	tx = FilterIs(tx, columnPkgName, params.PackageNameIs)
	tx = FilterIsNot(tx, columnPkgName, params.PackageNameIsNot)
	tx = FilterContains(tx, columnPkgName, params.PackageNameContains)
	tx = FilterStartsWith(tx, columnPkgName, params.PackageNameStart)
	tx = FilterEndsWith(tx, columnPkgName, params.PackageNameEnd)

	// version filters
	tx = FilterIs(tx, columnPkgVersion, params.PackageVersionIs)
	tx = FilterIsNot(tx, columnPkgVersion, params.PackageVersionIsNot)
	tx = FilterContains(tx, columnPkgVersion, params.PackageVersionContains)
	tx = FilterStartsWith(tx, columnPkgVersion, params.PackageVersionStart)
	tx = FilterEndsWith(tx, columnPkgVersion, params.PackageVersionEnd)

	// license filters
	tx = FilterIs(tx, columnPkgLicense, params.LicenseIs)
	tx = FilterIsNot(tx, columnPkgLicense, params.LicenseIsNot)
	tx = FilterContains(tx, columnPkgLicense, params.LicenseContains)
	tx = FilterStartsWith(tx, columnPkgLicense, params.LicenseStart)
	tx = FilterEndsWith(tx, columnPkgLicense, params.LicenseEnd)

	// language filters
	tx = FilterIs(tx, columnPkgLanguage, params.LanguageIs)
	tx = FilterIsNot(tx, columnPkgLanguage, params.LanguageIsNot)
	tx = FilterContains(tx, columnPkgLanguage, params.LanguageContains)
	tx = FilterStartsWith(tx, columnPkgLanguage, params.LanguageStart)
	tx = FilterEndsWith(tx, columnPkgLanguage, params.LanguageEnd)

	// applications filter
	tx = FilterIsNumber(tx, columnPkgViewApplications, params.ApplicationsIs)
	tx = FilterIsNotNumber(tx, columnPkgViewApplications, params.ApplicationsIsNot)
	tx = FilterGte(tx, columnPkgViewApplications, params.ApplicationsGte)
	tx = FilterLte(tx, columnPkgViewApplications, params.ApplicationsLte)

	// resources filter
	tx = FilterIsNumber(tx, columnPkgViewResources, params.ApplicationResourcesIs)
	tx = FilterIsNotNumber(tx, columnPkgViewResources, params.ApplicationResourcesIsNot)
	tx = FilterGte(tx, columnPkgViewResources, params.ApplicationResourcesGte)
	tx = FilterLte(tx, columnPkgViewResources, params.ApplicationResourcesLte)

	// vulnerabilities filter
	tx = SeverityFilterGte(tx, columnSeverityCountersHighestSeverity, params.VulnerabilitySeverityGte)
	tx = SeverityFilterLte(tx, columnSeverityCountersHighestSeverity, params.VulnerabilitySeverityLte)

	ids, err := p.getPackageIDs(params)
	if err != nil {
		return nil, fmt.Errorf("failed to get package IDs: %v", err)
	}
	tx = FilterIs(tx, columnPkgID, ids)

	return tx, nil
}

func PackageFromDB(view *PackageView) *models.Package {
	return &models.Package{
		ApplicationResources: uint32(view.Resources),
		Applications:         uint32(view.Applications),
		ID:                   view.ID,
		Language:             view.Language,
		License:              view.License,
		PackageName:          view.Name,
		Version:              view.Version,
		Vulnerabilities:      getVulnerabilityCount(view.SeverityCounters),
	}
}

func (p *PackageTableHandler) GetPackagesCountPerLicense() ([]*models.PackagesCountPerLicense, error) {
	tx := p.packagesView

	var pkgCount []*models.PackagesCountPerLicense
	var count uint32
	var license string

	// nolint:sqlclosecheck
	// If Next is called and returns false and there are no further result sets,
	// the Rows are closed automatically.
	rows, err := FilterIsNotEmptyString(tx, columnPkgLicense).Select(columnPkgLicense, "COUNT(id) AS count").Group(columnPkgLicense).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to group rows: %v", err)
	}

	for rows.Next() {
		if err := rows.Scan(&license, &count); err != nil {
			return nil, fmt.Errorf("failed to get fields: %v", err)
		}

		pkgCount = append(pkgCount, &models.PackagesCountPerLicense{
			Count:   count,
			License: license,
		})
	}
	return pkgCount, nil
}

func (p *PackageTableHandler) GetPackagesCountPerLanguage() ([]*models.PackagesCountPerLanguage, error) {
	tx := p.packagesView

	var pkgCount []*models.PackagesCountPerLanguage
	var count uint32
	var language string

	// nolint:sqlclosecheck
	// If Next is called and returns false and there are no further result sets,
	// the Rows are closed automatically.
	rows, err := FilterIsNotEmptyString(tx, columnPkgLanguage).Select(columnPkgLanguage, "COUNT(id) AS count").Group(columnPkgLanguage).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to group rows: %v", err)
	}

	for rows.Next() {
		if err := rows.Scan(&language, &count); err != nil {
			return nil, fmt.Errorf("failed to get fields: %v", err)
		}

		pkgCount = append(pkgCount, &models.PackagesCountPerLanguage{
			Count:    count,
			Language: language,
		})
	}
	return pkgCount, nil
}

func (p *PackageTableHandler) GetMostVulnerable(limit int) ([]*models.Package, error) {
	tx := p.packagesView

	var views []PackageView

	sortOrder, err := createVulnerabilitiesColumnSortOrder("desc")
	if err != nil {
		return nil, fmt.Errorf("failed to create sort order: %v", err)
	}

	if err := tx.Order(sortOrder).Limit(limit).Scan(&views).Error; err != nil {
		return nil, err
	}

	pkgs := make([]*models.Package, len(views))
	for i := range views {
		pkgs[i] = PackageFromDB(&views[i])
	}

	return pkgs, nil
}

func (p *PackageTableHandler) Count(filters *CountFilters) (int64, error) {
	var count int64
	var err error

	tx := p.packagesView

	tx, err = p.setCountFilters(tx, filters)
	if err != nil {
		return 0, fmt.Errorf("failed to set count filters: %v", err)
	}

	if err := tx.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count total: %v", err)
	}
	return count, nil
}

func (p *PackageTableHandler) GetPackagesAndTotal(params GetPackagesParams) ([]PackageView, int64, error) {
	var count int64
	var Packages []PackageView

	tx, err := p.setPackagesFilters(params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to set filters: %v", err)
	}

	// get total item count with the set filters
	if err := tx.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count total: %v", err)
	}

	sortOrder, err := createApplicationPackagesSortOrder(params.SortKey, params.SortDir)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create sort order: %v", err)
	}

	// get specific page ordered items with the current filters
	if err := tx.Scopes(Paginate(params.Page, params.PageSize)).
		Order(sortOrder).
		Find(&Packages).Error; err != nil {
		return nil, 0, err
	}

	return Packages, count, nil
}

func (p *PackageTableHandler) GetPackage(id string) (*models.Package, error) {
	var packageView PackageView

	if err := p.packagesView.
		Where(packageViewName+"."+columnPkgID+" = ?", id).
		First(&packageView).Error; err != nil {
		return nil, fmt.Errorf("failed to get package by id %q: %w", id, err)
	}

	return PackageFromDB(&packageView), nil
}

func (p *PackageTableHandler) setCountFilters(tx *gorm.DB, filters *CountFilters) (*gorm.DB, error) {
	if filters == nil {
		return tx, nil
	}

	// application ids filter
	pkgIds, err := p.IDsView.GetIDs(GetIDsParams{
		FilterIDs:    filters.ApplicationIDs,
		FilterIDType: ApplicationIDType,
		LookupIDType: PackageIDType,
	}, true)
	if err != nil {
		return tx, fmt.Errorf("failed to get package ids by app ids %v: %v", filters.ApplicationIDs, err)
	}
	tx = FilterIs(tx, columnPkgID, pkgIds)

	// severity filter
	tx = SeverityFilterGte(tx, columnSeverityCountersHighestSeverity, filters.VulnerabilitySeverityGte)

	return tx, nil
}

func createApplicationPackagesSortOrder(sortKey string, sortDir *string) (string, error) {
	if models.PackagesSortKey(sortKey) == models.PackagesSortKeyVulnerabilities {
		return createVulnerabilitiesColumnSortOrder(*sortDir)
	}

	sortKeyColumnName, err := getApplicationPackagesSortKeyColumnName(sortKey)
	if err != nil {
		return "", fmt.Errorf("failed to get sort key column name: %v", err)
	}

	return fmt.Sprintf("%v %v", sortKeyColumnName, strings.ToLower(*sortDir)), nil
}

func getApplicationPackagesSortKeyColumnName(key string) (string, error) {
	switch models.PackagesSortKey(key) {
	case models.PackagesSortKeyPackageName:
		return columnPkgName, nil
	case models.PackagesSortKeyVersion:
		return columnPkgVersion, nil
	case models.PackagesSortKeyLanguage:
		return columnPkgLanguage, nil
	case models.PackagesSortKeyLicense:
		return columnPkgLicense, nil
	case models.PackagesSortKeyApplications:
		return columnPkgViewApplications, nil
	case models.PackagesSortKeyApplicationResources:
		return columnPkgViewResources, nil
	case models.PackagesSortKeyVulnerabilities:
		return "", fmt.Errorf("unsupported key (%v)", key)
	}

	return "", fmt.Errorf("unknown sort key (%v)", key)
}

func (p *PackageTableHandler) getPackageIDs(params GetPackagesParams) ([]string, error) {
	ids, err := p.IDsView.GetIDs(p.createGetIDsParams(params), true)
	if err != nil {
		return nil, fmt.Errorf("failed to get IDs: %v", err)
	}

	return ids, nil
}

func (p *PackageTableHandler) createGetIDsParams(params GetPackagesParams) GetIDsParams {
	retParams := GetIDsParams{
		LookupIDType: PackageIDType,
	}

	// system filters - only one is allowed
	if params.CurrentRuntimeScan != nil && *params.CurrentRuntimeScan {
		retParams.FilterIDType = ApplicationIDType
		retParams.FilterIDs = params.RuntimeScanApplicationIDs
	} else if params.ApplicationID != nil {
		retParams.FilterIDType = ApplicationIDType
		retParams.FilterIDs = []string{*params.ApplicationID}
	} else if params.ApplicationResourceID != nil {
		retParams.FilterIDType = ResourceIDType
		retParams.FilterIDs = []string{*params.ApplicationResourceID}
	}

	return retParams
}
