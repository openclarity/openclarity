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
	"os"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	localDBPath = "./db.db"
)

const (
	DBDriverTypeLocal = "LOCAL"
)

type Database interface {
	TargetsTable() TargetsTable
	ScanResultsTable() ScanResultsTable
}

type Handler struct {
	DB *gorm.DB
}

type DBConfig struct {
	EnableInfoLogs bool
	DriverType     string
	DBPassword     string
	DBUser         string
	DBHost         string
	DBPort         string
	DBName         string
}

func Init(config *DBConfig) *Handler {
	databaseHandler := Handler{}

	databaseHandler.DB = initDataBase(config)

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

func initDataBase(config *DBConfig) *gorm.DB {
	dbDriver := config.DriverType
	dbLogger := logger.Default
	if config.EnableInfoLogs {
		dbLogger = dbLogger.LogMode(logger.Info)
	}

	db := initDB(config, dbDriver, dbLogger)

	// recreate views from scratch
	dropAllViews(db)

	// this will ensure table is created
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run auto migration: %v", err)
	}

	createAllViews(db)

	return db
}

func initDB(_ *DBConfig, dbDriver string, dbLogger logger.Interface) *gorm.DB {
	var db *gorm.DB
	switch dbDriver {
	case DBDriverTypeLocal:
		db = initSqlite(dbLogger)
	default:
		log.Fatalf("DB driver is not supported: %v", dbDriver)
	}
	return db
}

// nolint:cyclop
func createAllViews(db *gorm.DB) {
}

func dropAllViews(db *gorm.DB) {
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
