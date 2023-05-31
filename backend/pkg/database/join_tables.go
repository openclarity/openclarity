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

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openclarity/kubeclarity/api/server/models"
	"github.com/openclarity/kubeclarity/api/server/restapi/operations"
)

const (
	ApplicationResourcesJoinTableName        = "application_resources"
	ResourcePackagesJoinTableName            = "resource_packages"
	PackageVulnerabilitiesJoinTableName      = "package_vulnerabilities"
	ResourceCISDockerBenchmarkCheckTableName = "resource_cis_d_b_checks"

	// NOTE: when changing one of the column names change also the gorm label in JoinTables below.
	columnJoinTableApplicationID   = "application_id"
	columnJoinTableResourceID      = "resource_id"
	columnJoinTablePackageID       = "package_id"
	columnJoinTableVulnerabilityID = "vulnerability_id"

	// NOTE: when changing one of the column names change also the gorm label in ResourcePackages below.
	columnResourcePackagesAnalyzers = "analyzers"

	fixVersionsContextValueName = "fix_versions"
	analyzersContextValueName   = "analyzers"
)

// ApplicationResources join table of Application and Resource.
type ApplicationResources struct {
	ApplicationID string `json:"application_id,omitempty" gorm:"primarykey;column:application_id"`
	ResourceID    string `json:"resource_id,omitempty" gorm:"primarykey;column:resource_id"`
}

// ResourcePackages join table of Resource and Package.
type ResourcePackages struct {
	ResourceID string `json:"resource_id,omitempty" gorm:"primarykey;column:resource_id"`
	PackageID  string `json:"package_id,omitempty" gorm:"primarykey;column:package_id"`
	Analyzers  string `json:"analyzers,omitempty" gorm:"column:analyzers"`
}

// ResourceCISDBChecks join table of Resource and CISDockerBenchmarkCheck.
type ResourceCISDBChecks struct {
	CISDockerBenchmarkCheckID string `json:"cis_docker_benchmark_check_id,omitempty" gorm:"primarykey;column:cis_docker_benchmark_check_id"`
	ResourceID                string `json:"resource_id,omitempty" gorm:"primarykey;column:resource_id"`
}

func (ResourceCISDBChecks) TableName() string {
	return ResourceCISDockerBenchmarkCheckTableName
}

func (rp *ResourcePackages) BeforeSave(db *gorm.DB) error {
	analyzers, ok := db.Statement.Context.Value(analyzersContextValueName).(map[ResourcePkgID][]string)
	if ok {
		rp.Analyzers = ArrayToDBArray(analyzers[CreateResourcePkgID(rp.ResourceID, rp.PackageID)])
	}

	// Update analyzers list on conflict.
	db.Statement.AddClause(clause.OnConflict{
		Columns: []clause.Column{
			{Name: columnJoinTableResourceID},
			{Name: columnJoinTablePackageID},
		},
		UpdateAll: true,
	})

	return nil
}

func PackageApplicationResourcesFromDB(pri *PackageResourcesInfoView) *models.PackageApplicationResources {
	return &models.PackageApplicationResources{
		PackageID:              pri.PackageID,
		ReportingSBOMAnalyzers: DBArrayToArray(pri.Analyzers),
		ResourceHash:           pri.ResourceHash,
		ResourceID:             pri.ResourceID,
		ResourceName:           pri.ResourceName,
	}
}

// PackageVulnerabilities join table of Package and Vulnerability.
type PackageVulnerabilities struct {
	PackageID       string `json:"package_id,omitempty" gorm:"primarykey;column:package_id"`
	VulnerabilityID string `json:"vulnerability_id,omitempty" gorm:"primarykey;column:vulnerability_id"`
	FixVersion      string `json:"fix_version,omitempty" gorm:"column:fix_version"`
}

func (pv *PackageVulnerabilities) BeforeSave(db *gorm.DB) error {
	fv, ok := db.Statement.Context.Value(fixVersionsContextValueName).(map[PkgVulID]string)
	if ok {
		pv.FixVersion = fv[CreatePkgVulID(pv.PackageID, pv.VulnerabilityID)]
	}
	return nil
}

type JoinTables interface {
	DeleteRelationships(params DeleteRelationshipsParams) error
	// GetResourcePackageIDToAnalyzers returns a map of ResourcePkgID to analyzers list for the given `resourceIDs`,
	// retrieved from the PackageResources join table.
	// ResourcePackage is a package that is associated to a resource.
	// ResourcePackageID is a combination of resource and package ID.
	GetResourcePackageIDToAnalyzers(resourceIDs []string) (map[ResourcePkgID][]string, error)
	GetPackageResourcesAndTotal(params operations.GetPackagesIDApplicationResourcesParams) ([]PackageResourcesInfoView, int64, error)
	GetResourcePackagesByResources(resourceIDs []string) ([]ResourcePackages, error)
	GetResourcePackagesByPackages(packageIDs []string) ([]ResourcePackages, error)
	GetPackageVulnerabilitiesByPackages(packageIDs []string) ([]PackageVulnerabilities, error)
	GetPackageVulnerabilitiesByVulnerabilities(vulnerabilityIDs []string) ([]PackageVulnerabilities, error)
}

type JoinTablesHandler struct {
	db                 *gorm.DB
	viewRefreshHandler *ViewRefreshHandler
}

type DeleteRelationshipsParams struct {
	ApplicationIDsToRemove []string
	ResourceIDsToRemove    []string
	PackageIDsToRemove     []string
}

func (j *JoinTablesHandler) DeleteRelationships(params DeleteRelationshipsParams) error {
	err := j.db.Transaction(func(tx *gorm.DB) error {
		if len(params.PackageIDsToRemove) > 0 {
			pv := FilterIs(tx.Model(&PackageVulnerabilities{}),
				FieldInTable(PackageVulnerabilitiesJoinTableName, columnJoinTablePackageID), params.PackageIDsToRemove)
			if err := pv.Delete(&PackageVulnerabilities{}).Error; err != nil {
				return fmt.Errorf("failed to delete relationships in %q join table: %v",
					PackageVulnerabilitiesJoinTableName, err)
			}
		}

		if len(params.ResourceIDsToRemove) > 0 {
			rp := FilterIs(tx.Model(&ResourcePackages{}),
				FieldInTable(ResourcePackagesJoinTableName, columnJoinTableResourceID), params.ResourceIDsToRemove)
			if err := rp.Delete(&ResourcePackages{}).Error; err != nil {
				return fmt.Errorf("failed to delete relationships in %q join table: %v",
					ResourcePackagesJoinTableName, err)
			}
			ar := FilterIs(tx.Model(&ApplicationResources{}),
				FieldInTable(ApplicationResourcesJoinTableName, columnJoinTableResourceID), params.ResourceIDsToRemove)
			if err := ar.Delete(&ApplicationResources{}).Error; err != nil {
				return fmt.Errorf("failed to delete relationships in %q join table: %v",
					ApplicationResourcesJoinTableName, err)
			}
			rc := FilterIs(tx.Model(&ResourceCISDBChecks{}),
				FieldInTable(ResourceCISDockerBenchmarkCheckTableName, columnJoinTableResourceID), params.ResourceIDsToRemove)
			if err := rc.Delete(&ResourceCISDBChecks{}).Error; err != nil {
				return fmt.Errorf("failed to delete relationships in %q join table: %v",
					ResourceCISDockerBenchmarkCheckTableName, err)
			}
		}

		if len(params.ApplicationIDsToRemove) > 0 {
			ar := FilterIs(tx.Model(&ApplicationResources{}),
				FieldInTable(ApplicationResourcesJoinTableName, columnJoinTableApplicationID), params.ApplicationIDsToRemove)
			if err := ar.Delete(&ApplicationResources{}).Error; err != nil {
				return fmt.Errorf("failed to delete relationships in %q join table: %v",
					ApplicationResourcesJoinTableName, err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to delete relationships: %v", err)
	}

	j.tableChanged(packageTableName, resourceTableName, applicationTableName, vulnerabilityTableName)

	return nil
}

func (j *JoinTablesHandler) GetResourcePackageIDToAnalyzers(resourceIDs []string) (map[ResourcePkgID][]string, error) {
	// nolint:nilnil
	if len(resourceIDs) == 0 {
		return nil, nil
	}

	var resourcePackages []ResourcePackages
	resourcePkgIDToAnalyzers := make(map[ResourcePkgID][]string)

	tx := j.db.Model(&ResourcePackages{})
	tx = FilterIs(tx, FieldInTable(ResourcePackagesJoinTableName, columnJoinTableResourceID), resourceIDs)
	tx = FilterIsNotEmptyString(tx, FieldInTable(ResourcePackagesJoinTableName, columnResourcePackagesAnalyzers))
	if err := tx.Find(&resourcePackages).Error; err != nil {
		return nil, fmt.Errorf("failed to find relationships in %q join table: %v",
			ResourcePackagesJoinTableName, err)
	}

	for _, rp := range resourcePackages {
		resourcePkgIDToAnalyzers[CreateResourcePkgID(rp.ResourceID, rp.PackageID)] = DBArrayToArray(rp.Analyzers)
	}

	log.Debugf("Returning resourcePkgIDToAnalyzers=%+v", resourcePkgIDToAnalyzers)
	return resourcePkgIDToAnalyzers, nil
}

func (j *JoinTablesHandler) GetResourcePackagesByResources(resourceIDs []string) ([]ResourcePackages, error) {
	// nolint:nilnil
	if len(resourceIDs) == 0 {
		return nil, nil
	}

	var resourcePackages []ResourcePackages
	tx := j.db.Model(&ResourcePackages{})
	tx = FilterIs(tx, FieldInTable(ResourcePackagesJoinTableName, columnJoinTableResourceID), resourceIDs)
	if err := tx.Find(&resourcePackages).Error; err != nil {
		return nil, fmt.Errorf("failed to find relationships in %q join table: %v",
			ResourcePackagesJoinTableName, err)
	}

	return resourcePackages, nil
}

func (j *JoinTablesHandler) GetResourcePackagesByPackages(packageIDs []string) ([]ResourcePackages, error) {
	// nolint:nilnil
	if len(packageIDs) == 0 {
		return nil, nil
	}

	var resourcePackages []ResourcePackages
	tx := j.db.Model(&ResourcePackages{})
	tx = FilterIs(tx, FieldInTable(ResourcePackagesJoinTableName, columnJoinTablePackageID), packageIDs)
	if err := tx.Find(&resourcePackages).Error; err != nil {
		return nil, fmt.Errorf("failed to find relationships in %q join table: %v",
			ResourcePackagesJoinTableName, err)
	}

	return resourcePackages, nil
}

func (j *JoinTablesHandler) GetPackageVulnerabilitiesByPackages(packageIDs []string) ([]PackageVulnerabilities, error) {
	// nolint:nilnil
	if len(packageIDs) == 0 {
		return nil, nil
	}

	var packageVulnerabilities []PackageVulnerabilities
	tx := j.db.Model(&PackageVulnerabilities{})
	tx = FilterIs(tx, FieldInTable(PackageVulnerabilitiesJoinTableName, columnJoinTablePackageID), packageIDs)
	if err := tx.Find(&packageVulnerabilities).Error; err != nil {
		return nil, fmt.Errorf("failed to find relationships in %q join table: %v",
			PackageVulnerabilitiesJoinTableName, err)
	}

	return packageVulnerabilities, nil
}

func (j *JoinTablesHandler) GetPackageVulnerabilitiesByVulnerabilities(vulnerabilityIDs []string) ([]PackageVulnerabilities, error) {
	// nolint:nilnil
	if len(vulnerabilityIDs) == 0 {
		return nil, nil
	}

	var packageVulnerabilities []PackageVulnerabilities
	tx := j.db.Model(&PackageVulnerabilities{})
	tx = FilterIs(tx, FieldInTable(PackageVulnerabilitiesJoinTableName, columnJoinTableVulnerabilityID), vulnerabilityIDs)
	if err := tx.Find(&packageVulnerabilities).Error; err != nil {
		return nil, fmt.Errorf("failed to find relationships in %q join table: %v",
			PackageVulnerabilitiesJoinTableName, err)
	}

	return packageVulnerabilities, nil
}

func (j *JoinTablesHandler) GetPackageResourcesAndTotal(params operations.GetPackagesIDApplicationResourcesParams) ([]PackageResourcesInfoView, int64, error) {
	var count int64
	var packageResourcesInfoViews []PackageResourcesInfoView

	// Select PackageResourcesInfoView by package ID.
	tx := j.db.Model(&PackageResourcesInfoView{}).
		Where(FieldInTable(packageResourcesInfoView, columnPackageResourcesInfoViewPackageID)+" = ?", params.ID)

	// Apply filters.
	tx = j.setPackageResourcesFilters(tx, params)

	// Get total item count with the set filters.
	if err := tx.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count total: %w", err)
	}

	sortOrder, err := createPackageResourcesInfoSortOrder(params.SortKey, params.SortDir)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create sort order: %v", err)
	}

	// Get specific page ordered items with the current filters.
	if err := tx.Scopes(Paginate(params.Page, params.PageSize)).
		Order(sortOrder).
		Find(&packageResourcesInfoViews).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get package resources info total: %w", err)
	}

	return packageResourcesInfoViews, count, nil
}

func createPackageResourcesInfoSortOrder(sortKey string, sortDir *string) (string, error) {
	sortKeyColumnName, err := getPackageResourcesInfoSortKeyColumnName(sortKey)
	if err != nil {
		return "", fmt.Errorf("failed to get sort key column name: %v", err)
	}

	return fmt.Sprintf("%v %v", sortKeyColumnName, strings.ToLower(*sortDir)), nil
}

func getPackageResourcesInfoSortKeyColumnName(key string) (string, error) {
	switch models.PackagesApplicationResourcesSortKey(key) {
	case models.PackagesApplicationResourcesSortKeyResourceName:
		return columnPackageResourcesInfoViewResourceName, nil
	case models.PackagesApplicationResourcesSortKeyResourceHash:
		return columnPackageResourcesInfoViewResourceHash, nil
	}

	return "", fmt.Errorf("unknown sort key (%v)", key)
}

func (j *JoinTablesHandler) setPackageResourcesFilters(tx *gorm.DB, params operations.GetPackagesIDApplicationResourcesParams) *gorm.DB {
	// name filters
	tx = FilterIs(tx, columnPackageResourcesInfoViewResourceName, params.ResourceNameIs)
	tx = FilterIsNot(tx, columnPackageResourcesInfoViewResourceName, params.ResourceNameIsNot)
	tx = FilterContains(tx, columnPackageResourcesInfoViewResourceName, params.ResourceNameContains)
	tx = FilterStartsWith(tx, columnPackageResourcesInfoViewResourceName, params.ResourceNameStart)
	tx = FilterEndsWith(tx, columnPackageResourcesInfoViewResourceName, params.ResourceNameEnd)

	// hash filters
	tx = FilterIs(tx, columnPackageResourcesInfoViewResourceHash, params.ResourceHashIs)
	tx = FilterIsNot(tx, columnPackageResourcesInfoViewResourceHash, params.ResourceHashIsNot)
	tx = FilterContains(tx, columnPackageResourcesInfoViewResourceHash, params.ResourceHashContains)
	tx = FilterStartsWith(tx, columnPackageResourcesInfoViewResourceHash, params.ResourceHashStart)
	tx = FilterEndsWith(tx, columnPackageResourcesInfoViewResourceHash, params.ResourceHashEnd)

	// analyzers filter
	tx = FilterArrayContains(tx, columnPackageResourcesInfoViewAnalyzers, params.ReportingSBOMAnalyzersContainElements)
	tx = FilterArrayDoesntContain(tx, columnPackageResourcesInfoViewAnalyzers, params.ReportingSBOMAnalyzersDoesntContainElements)

	return tx
}

func (j *JoinTablesHandler) tableChanged(tables ...string) {
	for _, table := range tables {
		j.viewRefreshHandler.TableChanged(table)
	}
}
