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

	uuid "github.com/satori/go.uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/openclarity/vmclarity/backend/pkg/database/types"
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
	base.ID = uuid.NewV4()
	return nil
}

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
		Target{},
		ScanResult{},
		ScanConfig{},
		Scan{},
		Scopes{},
		Finding{},
	); err != nil {
		return nil, fmt.Errorf("failed to run auto migration: %w", err)
	}

	return db, nil
}

func initDB(config types.DBConfig, dbDriver string, dbLogger logger.Interface) (*gorm.DB, error) {
	switch dbDriver {
	case types.DBDriverTypeLocal:
		return initSqlite(config, dbLogger)
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
