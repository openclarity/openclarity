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

type Database interface {
	SBOMTable() SBOMTable
}

type Handler struct {
	DB *gorm.DB
}

const (
	localDBPath = "/tmp/db.db"
)

func InitDataBase(logLevel logger.LogLevel) *Handler {
	var db *gorm.DB
	dbLogger := logger.Default
	dbLogger = dbLogger.LogMode(logLevel)

	db = initSqlite(dbLogger)

	// this will ensure table is created
	if err := db.AutoMigrate(&SBOM{}); err != nil {
		log.Fatalf("Failed to run auto migration: %v", err)
	}

	return &Handler{DB: db}
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

func cleanLocalDataBase(databasePath string) {
	if _, err := os.Stat(databasePath); !os.IsNotExist(err) {
		log.Debug("deleting db...")
		if err := os.Remove(databasePath); err != nil {
			log.Warnf("failed to delete db file (%v): %v", databasePath, err)
		}
	}
}
