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
	"os"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	localDBPath = "./db.db"
)

const (
	DBDriverTypePostgres = "POSTGRES"
	DBDriverTypeLocal    = "LOCAL"
)

// In the case of materialized views slow queries can happen when refreshing the views,
// but reading these views will not cause slow queries when navigating on the UI.
// Materialized view is only supported by PostgreSQL.
// order is important, need to drop views in a reverse order of the creation.
var viewsList = []string{
	"cis_d_b_checks_view",
	vulnerabilityViewName, // MATERIALIZED in case of postgres.
	"package_resources_info_view",
	packageViewName, // MATERIALIZED in case of postgres.
	"package_severities",
	resourceViewName, // MATERIALIZED in case of postgres.
	"resource_cis_d_b_checks_view",
	"resource_severities",
	"new_vulnerabilities_view",
	applicationViewName, // MATERIALIZED in case of postgres.
	"application_cis_d_b_checks",
	"application_severities",
	"ids_view",
	"licenses_view",
}

const (
	createViewCommand                          = "CREATE VIEW"
	createMaterializedViewCommand              = "CREATE MATERIALIZED VIEW"
	dropViewCommand                            = "DROP VIEW IF EXISTS %s"
	dropMaterializedViewCommand                = "DROP MATERIALIZED VIEW IF EXISTS %s;"
	refreshMaterializedViewConcurrentlyCommand = "REFRESH MATERIALIZED VIEW CONCURRENTLY %s;"
	refreshMaterializedViewCommand             = "REFRESH MATERIALIZED VIEW %s;"
)

var (
	// In the views creations below we use LEFT JOIN and not INNER JOIN
	// to get info from multiple tables even if one is empty.
	// (https://stackoverflow.com/questions/16120789/join-three-tables-even-if-one-is-empty)

	// COALESCE is being used to get the first non null value,
	// when p.license is null an empty string will be selected.
	// ISNULL doesn't work with sqllite.

	licensesViewQuery = `
CREATE VIEW licenses_view AS
SELECT applications.id AS application_id,
       ar.resource_id,
       rp.package_id,
       COALESCE( p.license , '' ) AS license
FROM applications
         LEFT OUTER JOIN application_resources ar ON applications.id = ar.application_id
         LEFT OUTER JOIN resource_packages rp ON ar.resource_id = rp.resource_id
         LEFT OUTER JOIN packages p ON p.id = rp.package_id;
`

	// IDViewQuery ids_view is base on `resources` table to support detached resources (resource not related to any application).
	IDViewQuery = `
CREATE VIEW ids_view AS
SELECT ar.application_id,
       resources.id AS resource_id,
       rp.package_id,
       pv.vulnerability_id
FROM resources
         LEFT OUTER JOIN application_resources ar ON resources.id = ar.resource_id
         LEFT OUTER JOIN resource_packages rp ON resources.id = rp.resource_id
         LEFT OUTER JOIN package_vulnerabilities pv ON rp.package_id = pv.package_id;
`

	applicationsSeveritiesViewQuery = `
CREATE VIEW application_severities AS
SELECT apvs.application_id AS application_id,
       SUM(apvs.total_neg_count) AS total_neg_count,
       SUM(apvs.total_low_count) AS total_low_count,
       SUM(apvs.total_medium_count) AS total_medium_count,
       SUM(apvs.total_high_count) AS total_high_count,
       SUM(apvs.total_critical_count) AS total_critical_count,
       MAX(apvs.highest_severity) AS highest_severity,
       MIN(apvs.lowest_severity) AS lowest_severity
FROM (SELECT applications.id AS application_id,
       CASE WHEN v.severity = 0 THEN 1 ELSE 0 END AS total_neg_count,
       CASE WHEN v.severity = 1 THEN 1 ELSE 0 END AS total_low_count,
       CASE WHEN v.severity = 2 THEN 1 ELSE 0 END AS total_medium_count,
       CASE WHEN v.severity = 3 THEN 1 ELSE 0 END AS total_high_count,
       CASE WHEN v.severity = 4 THEN 1 ELSE 0 END AS total_critical_count,
       MAX(v.severity) AS highest_severity,
       MIN(v.severity) AS lowest_severity
FROM applications
         LEFT OUTER JOIN application_resources ar ON applications.id = ar.application_id
         LEFT OUTER JOIN resource_packages rp ON ar.resource_id = rp.resource_id
         LEFT OUTER JOIN package_vulnerabilities pv ON rp.package_id = pv.package_id
         LEFT OUTER JOIN vulnerabilities v ON v.id = pv.vulnerability_id
GROUP BY applications.id, pv.package_id, pv.vulnerability_id, v.severity) AS apvs
GROUP BY apvs.application_id;
`

	applicationsCISDockerBenchmarkChecksViewQuery = `
CREATE VIEW application_cis_d_b_checks AS
SELECT applications.id AS application_id,
       SUM(CASE WHEN c.level = 1 THEN 1 ELSE 0 END) AS total_info_count,
       SUM(CASE WHEN c.level = 2 THEN 1 ELSE 0 END) AS total_warn_count,
       SUM(CASE WHEN c.level = 3 THEN 1 ELSE 0 END) AS total_fatal_count,
       MAX(c.level) AS highest_level,
       MIN(c.level) AS lowest_level
FROM applications
         LEFT OUTER JOIN application_resources ar ON applications.id = ar.application_id
         LEFT OUTER JOIN resource_cis_d_b_checks rc ON ar.resource_id = rc.resource_id
         LEFT OUTER JOIN cis_d_b_checks c ON c.id = rc.cis_docker_benchmark_check_id
GROUP BY applications.id;
`

	applicationsViewQuery = `
%s applications_view AS
SELECT applications.*,
       COUNT(distinct ar.resource_id) AS resources,
       COUNT(distinct rp.package_id) AS packages,
       aps.total_neg_count,
       aps.total_low_count,
       aps.total_medium_count,
       aps.total_high_count,
       aps.total_critical_count,
       aps.highest_severity,
       aps.lowest_severity,
       apl.total_info_count,
       apl.total_warn_count,
       apl.total_fatal_count,
       apl.lowest_level,
       apl.highest_level
FROM applications
         LEFT OUTER JOIN application_resources ar ON applications.id = ar.application_id
         LEFT OUTER JOIN resource_packages rp ON ar.resource_id = rp.resource_id
         LEFT OUTER JOIN application_severities aps ON applications.id = aps.application_id
         LEFT OUTER JOIN application_cis_d_b_checks apl ON applications.id = apl.application_id
GROUP BY applications.id,
         aps.total_neg_count,
         aps.total_low_count,
         aps.total_medium_count,
         aps.total_high_count,
         aps.total_critical_count,
         aps.highest_severity,
         aps.lowest_severity,
         apl.total_info_count,
         apl.total_warn_count,
         apl.total_fatal_count,
         apl.lowest_level,
         apl.highest_level;
`

	createApplicationsViewIndex = `
CREATE UNIQUE INDEX applications_view_id ON applications_view (id);
`

	newVulnerabilitiesViewQuery = `
CREATE VIEW new_vulnerabilities_view AS
SELECT new_vulnerabilities.added_at AS added_at,
       SUM(CASE WHEN v.severity = 0 THEN 1 ELSE 0 END) AS total_neg_count,
       SUM(CASE WHEN v.severity = 1 THEN 1 ELSE 0 END) AS total_low_count,
       SUM(CASE WHEN v.severity = 2 THEN 1 ELSE 0 END) AS total_medium_count,
       SUM(CASE WHEN v.severity = 3 THEN 1 ELSE 0 END) AS total_high_count,
       SUM(CASE WHEN v.severity = 4 THEN 1 ELSE 0 END) AS total_critical_count
FROM new_vulnerabilities
         LEFT OUTER JOIN vulnerabilities v ON v.id = new_vulnerabilities.vul_id
GROUP BY added_at;
`

	resourcesSeveritiesViewQuery = `
CREATE VIEW resource_severities AS
SELECT resources.id AS resource_id,
       SUM(CASE WHEN v.severity = 0 THEN 1 ELSE 0 END) AS total_neg_count,
       SUM(CASE WHEN v.severity = 1 THEN 1 ELSE 0 END) AS total_low_count,
       SUM(CASE WHEN v.severity = 2 THEN 1 ELSE 0 END) AS total_medium_count,
       SUM(CASE WHEN v.severity = 3 THEN 1 ELSE 0 END) AS total_high_count,
       SUM(CASE WHEN v.severity = 4 THEN 1 ELSE 0 END) AS total_critical_count,
       MAX(v.severity) AS highest_severity,
       MIN(v.severity) AS lowest_severity
FROM resources
         LEFT OUTER JOIN resource_packages rp ON resources.id = rp.resource_id
         LEFT OUTER JOIN package_vulnerabilities pv ON rp.package_id = pv.package_id
         LEFT OUTER JOIN vulnerabilities v ON v.id = pv.vulnerability_id
GROUP BY resources.id;
`

	resourcesCISDockerBenchmarkChecksViewQuery = `
CREATE VIEW resource_cis_d_b_checks_view AS
SELECT resources.id AS resource_id,
	  SUM(CASE WHEN c.level = 1 THEN 1 ELSE 0 END) AS total_info_count,
	  SUM(CASE WHEN c.level = 2 THEN 1 ELSE 0 END) AS total_warn_count,
	  SUM(CASE WHEN c.level = 3 THEN 1 ELSE 0 END) AS total_fatal_count,
	  MAX(c.level) AS highest_level,
	  MIN(c.level) AS lowest_level
FROM resources
		LEFT OUTER JOIN resource_cis_d_b_checks rc ON resources.id = rc.resource_id
		LEFT OUTER JOIN cis_d_b_checks c ON c.id = rc.cis_docker_benchmark_check_id
GROUP BY resources.id;
	`

	resourcesViewQuery = `
%s resources_view AS
SELECT resources.*,
       COUNT(distinct ar.application_id) AS applications,
       COUNT(distinct rp.package_id) AS packages,
       rs.total_neg_count,
       rs.total_low_count,
       rs.total_medium_count,
       rs.total_high_count,
       rs.total_critical_count,
       rs.highest_severity,
       rl.total_info_count,
       rl.total_warn_count,
       rl.total_fatal_count,
       rl.lowest_level,
       rl.highest_level
FROM resources
         LEFT OUTER JOIN resource_packages rp ON resources.id = rp.resource_id
         LEFT OUTER JOIN application_resources ar ON resources.id = ar.resource_id
         LEFT OUTER JOIN resource_severities rs ON resources.id = rs.resource_id
         LEFT OUTER JOIN resource_cis_d_b_checks_view rl ON resources.id = rl.resource_id
GROUP BY resources.id,
         rs.total_neg_count,
         rs.total_low_count,
         rs.total_medium_count,
         rs.total_high_count,
         rs.total_critical_count,
         rs.highest_severity,
         rs.lowest_severity,
         rl.total_info_count,
         rl.total_warn_count,
         rl.total_fatal_count,
         rl.lowest_level,
         rl.highest_level;
`

	createResourcessViewIndex = `
CREATE UNIQUE INDEX resources_view_id ON resources_view (id);
`

	packagesSeveritiesViewQuery = `
CREATE VIEW package_severities AS
SELECT packages.id AS package_id,
       SUM(CASE WHEN v.severity = 0 THEN 1 ELSE 0 END) AS total_neg_count,
       SUM(CASE WHEN v.severity = 1 THEN 1 ELSE 0 END) AS total_low_count,
       SUM(CASE WHEN v.severity = 2 THEN 1 ELSE 0 END) AS total_medium_count,
       SUM(CASE WHEN v.severity = 3 THEN 1 ELSE 0 END) AS total_high_count,
       SUM(CASE WHEN v.severity = 4 THEN 1 ELSE 0 END) AS total_critical_count,
       MAX(v.severity) AS highest_severity,
       MIN(v.severity) AS lowest_severity
FROM packages
         LEFT OUTER JOIN package_vulnerabilities pv ON packages.id = pv.package_id
         LEFT OUTER JOIN vulnerabilities v ON v.id = pv.vulnerability_id
GROUP BY packages.id;
	`

	packagesViewQuery = `
%s packages_view AS
SELECT packages.*,
       COUNT(distinct ar.application_id) AS applications,
       COUNT(distinct rp.resource_id) AS resources,
       ps.total_neg_count,
       ps.total_low_count,
       ps.total_medium_count,
       ps.total_high_count,
       ps.total_critical_count,
       ps.highest_severity,
       ps.lowest_severity
FROM packages
         LEFT OUTER JOIN resource_packages rp ON packages.id = rp.package_id
         LEFT OUTER JOIN application_resources ar ON rp.resource_id = ar.resource_id
         LEFT OUTER JOIN package_severities ps ON packages.id = ps.package_id
GROUP BY packages.id,
         ps.total_neg_count,
         ps.total_low_count,
         ps.total_medium_count,
         ps.total_high_count,
         ps.total_critical_count,
         ps.highest_severity,
         ps.lowest_severity;
	`

	createPackagesViewIndex = `
CREATE UNIQUE INDEX packages_view_id ON packages_view (id);
`

	packageResourcesInfoViewQuery = `
CREATE VIEW package_resources_info_view AS
SELECT resource_packages.*,
       resources.name AS resource_name,
       resources.hash AS resource_hash
FROM resource_packages
         LEFT OUTER JOIN resources ON resources.id = resource_packages.resource_id;
	`

	vulnerabilitiesViewQuery = `
%s vulnerabilities_view AS
SELECT vulnerabilities.*,
       COUNT(distinct ar.application_id) AS applications,
       COUNT(distinct ar.resource_id) AS resources,
       p.name AS package_name,
       p.version AS package_version,
       p.id AS package_id,
       pv.fix_version AS fix_version
FROM vulnerabilities
         LEFT OUTER JOIN package_vulnerabilities pv ON vulnerabilities.id = pv.vulnerability_id
         LEFT OUTER JOIN resource_packages rp ON pv.package_id = rp.package_id
         LEFT OUTER JOIN application_resources ar ON rp.resource_id = ar.resource_id
         LEFT OUTER JOIN packages p ON p.id = pv.package_id
GROUP BY vulnerabilities.id,
         p.name,
         p.version,
         p.id,
         pv.fix_version;
  `

	createVulnerabilitiesViewIndex = `
CREATE UNIQUE INDEX vulnerabilities_view_id ON vulnerabilities_view (id, package_id);
`

	cisDockerBenchmarkResultsViewQuery = `
CREATE VIEW cis_d_b_checks_view AS
SELECT resource_cis_d_b_checks.resource_id AS resource_id,
       cis_d_b_checks.id AS id,
       cis_d_b_checks.code AS code,
       cis_d_b_checks.level AS level,
       cis_d_b_checks.descriptions AS descriptions
FROM resource_cis_d_b_checks
         LEFT OUTER JOIN cis_d_b_checks ON cis_d_b_checks.id = resource_cis_d_b_checks.cis_docker_benchmark_check_id;
  `
)

// nolint:interfacebloat
type Database interface {
	ApplicationTable() ApplicationTable
	ResourceTable() ResourceTable
	PackageTable() PackageTable
	VulnerabilityTable() VulnerabilityTable
	NewVulnerabilityTable() NewVulnerabilityTable
	JoinTables() JoinTables
	IDsView() IDsView
	ObjectTree() ObjectTree
	QuickScanConfigTable() QuickScanConfigTable
	SchedulerTable() SchedulerTable
	CISDockerBenchmarkResultTable() CISDockerBenchmarkResultTable
}

type Handler struct {
	DriverType         string
	DB                 *gorm.DB
	ViewRefreshHandler *ViewRefreshHandler
}

type DBConfig struct {
	EnableInfoLogs            bool
	DriverType                string
	DBPassword                string
	DBUser                    string
	DBHost                    string
	DBPort                    string
	DBName                    string
	ViewRefreshIntervalSecond int
}

func (db *Handler) ObjectTree() ObjectTree {
	return &ObjectTreeHandler{
		DriverType:         db.DriverType,
		db:                 db.DB,
		viewRefreshHandler: db.ViewRefreshHandler,
	}
}

func (db *Handler) JoinTables() JoinTables {
	return &JoinTablesHandler{
		db:                 db.DB,
		viewRefreshHandler: db.ViewRefreshHandler,
	}
}

func (db *Handler) VulnerabilityTable() VulnerabilityTable {
	return &VulnerabilityTableHandler{
		vulnerabilitiesTable: db.DB.Table(vulnerabilityTableName),
		vulnerabilitiesView:  db.DB.Table(vulnerabilityViewName),
		IDsView:              db.IDsView(),
		viewRefreshHandler:   db.ViewRefreshHandler,
	}
}

func (db *Handler) NewVulnerabilityTable() NewVulnerabilityTable {
	return &NewVulnerabilityTableHandler{
		newVulnerabilitiesTable: db.DB.Table(newVulnerabilityTableName),
		newVulnerabilitiesView:  db.DB.Table(newVulnerabilitiesViewName),
	}
}

func (db *Handler) ResourceTable() ResourceTable {
	return &ResourceTableHandler{
		resourcesTable:     db.DB.Table(resourceTableName),
		resourcesView:      db.DB.Table(resourceViewName),
		licensesView:       db.DB.Table(licensesViewName),
		IDsView:            db.IDsView(),
		viewRefreshHandler: db.ViewRefreshHandler,
	}
}

func (db *Handler) ApplicationTable() ApplicationTable {
	return &ApplicationTableHandler{
		applicationsTable:  db.DB.Table(applicationTableName),
		applicationsView:   db.DB.Table(applicationViewName),
		licensesView:       db.DB.Table(licensesViewName),
		IDsView:            db.IDsView(),
		db:                 db.DB,
		viewRefreshHandler: db.ViewRefreshHandler,
	}
}

func (db *Handler) PackageTable() PackageTable {
	return &PackageTableHandler{
		packagesTable:      db.DB.Table(packageTableName),
		packagesView:       db.DB.Table(packageViewName),
		IDsView:            db.IDsView(),
		viewRefreshHandler: db.ViewRefreshHandler,
	}
}

func (db *Handler) IDsView() IDsView {
	return &IDsViewHandler{
		IDsView: db.DB.Table(idsViewName),
	}
}

func (db *Handler) QuickScanConfigTable() QuickScanConfigTable {
	return &QuickScanConfigTableHandler{
		table: db.DB.Table(quickScanConfigTableName),
	}
}

func (db *Handler) SchedulerTable() SchedulerTable {
	return &SchedulerTableHandler{
		table:              db.DB.Table(schedulerTableName),
		viewRefreshHandler: db.ViewRefreshHandler,
	}
}

func (db *Handler) CISDockerBenchmarkResultTable() CISDockerBenchmarkResultTable {
	return &CISDockerBenchmarkResultTableHandler{
		applicationsCisDockerBenchmarkChecksView: db.DB.Table(applicationCisDockerBenchmarkChecksViewName),
		cisDockerBenchmarkChecksView:             db.DB.Table(cisDockerBenchmarkChecksViewName),
	}
}

func Init(config *DBConfig) *Handler {
	databaseHandler := Handler{}

	databaseHandler.initDataBase(config)

	// Set defaults.
	err := databaseHandler.QuickScanConfigTable().SetDefault()
	if err != nil {
		log.Fatalf("Failed to set default quick scan config: %v", err)
	}
	databaseHandler.DriverType = config.DriverType

	return &databaseHandler
}

func cleanLocalDataBase(databasePath string) {
	if _, err := os.Stat(databasePath); !os.IsNotExist(err) {
		log.Debug("deleting db...")
		if err := os.Remove(databasePath); err != nil {
			log.Warnf("failed to delete db file (%v): %v", databasePath, err)
		}
	}
}

func (db *Handler) initDataBase(config *DBConfig) {
	dbDriver := config.DriverType
	dbLogger := logger.Default
	if config.EnableInfoLogs {
		dbLogger = dbLogger.LogMode(logger.Info)
	}

	db.DB = initDB(config, dbDriver, dbLogger)
	db.SetMaterializedViewHandler(config)

	setupJoinTables(db.DB)

	// drop views before auto migrate
	db.dropAllViews(dbDriver)

	// this will ensure table is created
	if err := db.DB.AutoMigrate(Application{}, Resource{}, Package{}, Vulnerability{}, NewVulnerability{}, QuickScanConfig{}, Scheduler{}, CISDockerBenchmarkCheck{}); err != nil {
		log.Fatalf("Failed to run auto migration: %v", err)
	}

	// recreate views from scratch
	db.createAllViews(dbDriver)
}

func initDB(config *DBConfig, dbDriver string, dbLogger logger.Interface) *gorm.DB {
	var db *gorm.DB
	switch dbDriver {
	case DBDriverTypePostgres:
		db = initPostgres(config, dbLogger)
	case DBDriverTypeLocal:
		db = initSqlite(dbLogger)
	default:
		log.Fatalf("DB driver is not supported: %v", dbDriver)
	}
	return db
}

func setupJoinTables(db *gorm.DB) {
	err := db.SetupJoinTable(&Package{}, "Vulnerabilities", &PackageVulnerabilities{})
	if err != nil {
		log.Fatalf("Failed to setup join table package_vulnerabilities: %v", err)
	}

	err = db.SetupJoinTable(&Resource{}, "Packages", &ResourcePackages{})
	if err != nil {
		log.Fatalf("Failed to setup join table resource_packages: %v", err)
	}
}

// nolint:cyclop
func (db *Handler) createAllViews(dbDriver string) {
	var createCommand string
	switch dbDriver {
	case DBDriverTypePostgres:
		createCommand = createMaterializedViewCommand
	default:
		createCommand = createViewCommand
	}

	if err := db.DB.Exec(licensesViewQuery).Error; err != nil {
		log.Fatalf("Failed to create licenses_view: %v", err)
	}

	if err := db.DB.Exec(IDViewQuery).Error; err != nil {
		log.Fatalf("Failed to create ids_view: %v", err)
	}

	if err := db.DB.Exec(applicationsSeveritiesViewQuery).Error; err != nil {
		log.Fatalf("Failed to create application_severities: %v", err)
	}

	if err := db.DB.Exec(applicationsCISDockerBenchmarkChecksViewQuery).Error; err != nil {
		log.Fatalf("Failed to create application_cis_d_b_checks: %v", err)
	}

	if err := db.DB.Exec(fmt.Sprintf(applicationsViewQuery, createCommand)).Error; err != nil {
		log.Fatalf("Failed to create applications_view: %v", err)
	}

	if err := db.DB.Exec(newVulnerabilitiesViewQuery).Error; err != nil {
		log.Fatalf("Failed to create new vulnerabilities trends view query: %v", err)
	}

	if err := db.DB.Exec(resourcesSeveritiesViewQuery).Error; err != nil {
		log.Fatalf("Failed to create resource_severities: %v", err)
	}

	if err := db.DB.Exec(resourcesCISDockerBenchmarkChecksViewQuery).Error; err != nil {
		log.Fatalf("Failed to create resource_cis_d_b_checks_view: %v", err)
	}

	if err := db.DB.Exec(fmt.Sprintf(resourcesViewQuery, createCommand)).Error; err != nil {
		log.Fatalf("Failed to create resources_view: %v", err)
	}

	if err := db.DB.Exec(packagesSeveritiesViewQuery).Error; err != nil {
		log.Fatalf("Failed to create package_severities: %v", err)
	}

	if err := db.DB.Exec(fmt.Sprintf(packagesViewQuery, createCommand)).Error; err != nil {
		log.Fatalf("Failed to create packages_view: %v", err)
	}

	if err := db.DB.Exec(packageResourcesInfoViewQuery).Error; err != nil {
		log.Fatalf("Failed to create package_resources_info_view: %v", err)
	}

	if err := db.DB.Exec(fmt.Sprintf(vulnerabilitiesViewQuery, createCommand)).Error; err != nil {
		log.Fatalf("Failed to create vulnerabilities_view: %v", err)
	}

	if err := db.DB.Exec(cisDockerBenchmarkResultsViewQuery).Error; err != nil {
		log.Fatalf("Failed to create cis_d_b_checks_view: %v", err)
	}

	if dbDriver == DBDriverTypePostgres {
		createIndexForMaterializedViews(db.DB)
		db.initPostgresMaterializedViews()
	}
}

func createIndexForMaterializedViews(db *gorm.DB) {
	if err := db.Exec(createApplicationsViewIndex).Error; err != nil {
		log.Fatalf("Failed to create index for %s: %v", applicationViewName, err)
	}

	if err := db.Exec(createResourcessViewIndex).Error; err != nil {
		log.Fatalf("Failed to create index for %s: %v", resourceViewName, err)
	}

	if err := db.Exec(createPackagesViewIndex).Error; err != nil {
		log.Fatalf("Failed to create index for %s: %v", packageViewName, err)
	}
	if err := db.Exec(createVulnerabilitiesViewIndex).Error; err != nil {
		log.Fatalf("Failed to create index for %s: %v", vulnerabilityViewName, err)
	}
}

func (db *Handler) dropAllViews(dbDriver string) {
	for _, view := range viewsList {
		dropViewIfExists(db.DB, view, dbDriver)
	}
}

func dropViewIfExists(db *gorm.DB, viewName, dbDriver string) {
	dropCommand := dropViewCommand
	if dbDriver == DBDriverTypePostgres {
		switch viewName {
		case packageViewName, applicationViewName, resourceViewName, vulnerabilityViewName:
			dropCommand = dropMaterializedViewCommand
		default:
			dropCommand = dropViewCommand
		}
	}
	if err := db.Exec(fmt.Sprintf(dropCommand, viewName)).Error; err != nil {
		log.Fatalf("Failed to drop %s: %v", viewName, err)
	}
}

func initPostgres(config *DBConfig, dbLogger logger.Interface) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		log.Fatalf("Failed to open %s db: %v", config.DBName, err)
	}

	return db
}

func initSqlite(dbLogger logger.Interface) *gorm.DB {
	cleanLocalDataBase(localDBPath)

	db, err := gorm.Open(sqlite.Open(localDBPath), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		log.Fatalf("Failed to open db: %v", err)
	}

	return db
}
