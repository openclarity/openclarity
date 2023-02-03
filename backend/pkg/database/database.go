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
	"time"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	DBDriverTypeLocal = "LOCAL"
)

type Database interface {
	ScanResultsTable() ScanResultsTable
	ScanConfigsTable() ScanConfigsTable
	ScansTable() ScansTable
	TargetsTable() TargetsTable
}

type Handler struct {
	DB *gorm.DB
}

type DBConfig struct {
	EnableInfoLogs bool   `json:"enable-info-logs"`
	DriverType     string `json:"driver-type,omitempty"`
	DBPassword     string `json:"-"`
	DBUser         string `json:"db-user,omitempty"`
	DBHost         string `json:"db-host,omitempty"`
	DBPort         string `json:"db-port,omitempty"`
	DBName         string `json:"db-name,omitempty"`

	LocalDBPath string `json:"local-db-path,omitempty"`
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

func Init(config *DBConfig) *Handler {
	databaseHandler := Handler{}

	databaseHandler.DB = initDataBase(config)

	return &databaseHandler
}

func initDataBase(config *DBConfig) *gorm.DB {
	dbDriver := config.DriverType
	dbLogger := logger.Default
	if config.EnableInfoLogs {
		dbLogger = dbLogger.LogMode(logger.Info)
	}

	db := initDB(config, dbDriver, dbLogger)

	// this will ensure table is created
	if err := db.AutoMigrate(Target{}, ScanResult{}, ScanConfig{}, Scan{}); err != nil {
		log.Fatalf("Failed to run auto migration: %v", err)
	}

	return db
}

func initDB(config *DBConfig, dbDriver string, dbLogger logger.Interface) *gorm.DB {
	var db *gorm.DB
	switch dbDriver {
	case DBDriverTypeLocal:
		db = initSqlite(config, dbLogger)
	default:
		log.Fatalf("DB driver is not supported: %v", dbDriver)
	}
	return db
}

func initSqlite(config *DBConfig, dbLogger logger.Interface) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(config.LocalDBPath), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		log.Fatalf("Failed to open db: %v", err)
	}

	return db
}
