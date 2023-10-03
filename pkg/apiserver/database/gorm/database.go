// Copyright © 2023 Cisco Systems, Inc. and its affiliates.
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

package gorm

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/openclarity/vmclarity/pkg/apiserver/database/odatasql/jsonsql"
	"github.com/openclarity/vmclarity/pkg/apiserver/database/types"
)

func NewDatabase(config types.DBConfig) (types.Database, error) {
	db, err := initDataBase(config)
	if err != nil {
		return nil, fmt.Errorf("unable to create new GORM database: %w", err)
	}
	return &Handler{DB: db}, nil
}

type Handler struct {
	DB *gorm.DB
}

// Base contains common columns for all tables.
type Base struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (base *Base) BeforeCreate(db *gorm.DB) error {
	base.ID = uuid.New()
	return nil
}

// nolint:cyclop
func initDataBase(config types.DBConfig) (*gorm.DB, error) {
	dbDriver := config.DriverType
	dbLogger := logger.Default
	if config.EnableInfoLogs {
		dbLogger = dbLogger.LogMode(logger.Info)
	}

	db, err := initDB(config, dbDriver, dbLogger)
	if err != nil {
		return nil, err
	}

	// this will ensure table is created
	if err := db.AutoMigrate(
		Asset{},
		AssetScan{},
		ScanConfig{},
		Scan{},
		Finding{},
		AssetScanEstimation{},
		ScanEstimation{},
		Provider{},
	); err != nil {
		return nil, fmt.Errorf("failed to run auto migration: %w", err)
	}

	// Create indexes for our objects
	//
	// First for all objects index the ID field this speeds up anywhere
	// we're getting a single object out of the DB, including in PATCH/PUT
	// etc.
	idb := db.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS assets_id_idx ON assets((%s))", SQLVariant.JSONExtract("Data", "$.id")))
	if idb.Error != nil {
		return nil, fmt.Errorf("failed to create index assets_id_idx: %w", idb.Error)
	}

	idb = db.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS asset_scans_id_idx ON asset_scans((%s))", SQLVariant.JSONExtract("Data", "$.id")))
	if idb.Error != nil {
		return nil, fmt.Errorf("failed to create index asset_scans_id_idx: %w", idb.Error)
	}

	idb = db.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS asset_scan_estimations_id_idx ON asset_scan_estimations((%s))", SQLVariant.JSONExtract("Data", "$.id")))
	if idb.Error != nil {
		return nil, fmt.Errorf("failed to create index asset_scan_estimations_id_idx: %w", idb.Error)
	}

	idb = db.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS scan_configs_id_idx ON scan_configs((%s))", SQLVariant.JSONExtract("Data", "$.id")))
	if idb.Error != nil {
		return nil, fmt.Errorf("failed to create index scan_configs_id_idx: %w", idb.Error)
	}

	idb = db.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS scans_id_idx ON scans((%s))", SQLVariant.JSONExtract("Data", "$.id")))
	if idb.Error != nil {
		return nil, fmt.Errorf("failed to create index scans_id_idx: %w", idb.Error)
	}

	idb = db.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS scan_estimations_id_idx ON scan_estimations((%s))", SQLVariant.JSONExtract("Data", "$.id")))
	if idb.Error != nil {
		return nil, fmt.Errorf("failed to create index scan_estimations_id_idx: %w", idb.Error)
	}

	idb = db.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS findings_id_idx ON findings((%s))", SQLVariant.JSONExtract("Data", "$.id")))
	if idb.Error != nil {
		return nil, fmt.Errorf("failed to create index findings_id_idx: %w", idb.Error)
	}

	idb = db.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS providers_id_idx ON providers((%s))", SQLVariant.JSONExtract("Data", "$.id")))
	if idb.Error != nil {
		return nil, fmt.Errorf("failed to create index providers_id_idx: %w", idb.Error)
	}

	// For processing asset scans to findings we need to find all the scan
	// results by general status and findingsProcessed, so add an index for
	// that.
	idb = db.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS asset_scans_findings_processed_idx ON asset_scans((%s), (%s))", SQLVariant.JSONExtract("Data", "$.findingsProcessed"), SQLVariant.JSONExtract("Data", "$.status.general.state")))
	if idb.Error != nil {
		return nil, fmt.Errorf("failed to create index asset_scans_findings_processed_idx: %w", idb.Error)
	}

	// The UI needs to find all the findings for a specific finding type
	// and the asset scan processor needs to filter that list by a
	// specific asset scan. So add a combined index for those cases.
	idb = db.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS findings_by_type_and_assetscan_idx ON findings((%s), (%s))", SQLVariant.JSONExtract("Data", "$.findingInfo.objectType"), SQLVariant.JSONExtract("Data", "$.assetScan.id")))
	if idb.Error != nil {
		return nil, fmt.Errorf("failed to create index findings_by_type_and_assetscan_idx: %w", idb.Error)
	}

	// The finding trends widget in the backend UI needs to count all the findings for a specific finding type
	// that was active during a given time point. So add a combined index for those cases.
	// Example query:
	//	SELECT COUNT(*) FROM findings WHERE ((findings.Data->'$.findingInfo.objectType' = JSON_QUOTE('Vulnerability') AND datetime(findings.Data->>'$.foundOn') <= datetime('2023-06-11T14:24:28Z')) AND (findings.Data->'$.invalidatedOn' is NULL OR datetime(findings.Data->>'$.invalidatedOn') > datetime('2023-06-11T14:24:28Z')))
	idb = db.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS findings_by_type_and_foundOn_and_invalidatedOn_idx ON findings((%s), (%s), (%s), (%s))",
		SQLVariant.JSONExtract("Data", "$.findingInfo.objectType"),
		SQLVariant.JSONExtractText("Data", "$.foundOn"),
		SQLVariant.JSONExtractText("Data", "$.invalidatedOn"),
		SQLVariant.JSONExtract("Data", "$.invalidatedOn")))
	if idb.Error != nil {
		return nil, fmt.Errorf("failed to create index findings_by_type_and_foundOn_and_invalidatedOn_idx: %w", idb.Error)
	}

	// TODO(sambetts) Add indexes for all the uniqueness checks we need to
	// do for each object

	return db, nil
}

func initDB(config types.DBConfig, dbDriver string, dbLogger logger.Interface) (*gorm.DB, error) {
	switch dbDriver {
	case types.DBDriverTypeLocal:
		SQLVariant = jsonsql.SQLite
		return initSqlite(config, dbLogger)
	case types.DBDriverTypePostgres:
		SQLVariant = jsonsql.Postgres
		return initPostgres(config, dbLogger)
	default:
		return nil, fmt.Errorf("driver type %s is not supported by GORM driver", dbDriver)
	}
}

func initSqlite(config types.DBConfig, dbLogger logger.Interface) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(config.LocalDBPath), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	return db, nil
}

func initPostgres(config types.DBConfig, dbLogger logger.Interface) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open %s db: %w", config.DBName, err)
	}

	return db, nil
}
