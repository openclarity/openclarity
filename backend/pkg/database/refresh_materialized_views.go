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
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const DefaultViewRefreshIntervalSecond = 5

type refreshFunc func(db *gorm.DB)

type ViewRefreshHandler struct {
	mu                        sync.Mutex
	viewsToRefresh            map[string][]string // map of tables that shows which views should be refreshed due to table changes
	tableChanged              map[string]bool
	refreshFunc               map[string]refreshFunc // map of refresh functions for specified views
	viewRefreshIntervalSecond time.Duration
}

func (vh *ViewRefreshHandler) TableChanged(table string) {
	if !vh.IsSetViewRefreshHandler() {
		return
	}
	vh.mu.Lock()
	defer vh.mu.Unlock()
	vh.tableChanged[table] = true
}

func (vh *ViewRefreshHandler) GetAndClearChanges() map[string]bool {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	tables := vh.tableChanged
	vh.tableChanged = make(map[string]bool)
	return tables
}

func (vh *ViewRefreshHandler) runRequiredRefreshes(db *gorm.DB) {
	viewNames := vh.getViewsToRefresh()
	for viewName := range viewNames {
		vh.refreshFunc[viewName](db)
	}
}

// getViewsToRefresh creates a list of views that should be refreshed due to table changes.
// The viewsToRefresh is the map of tables that shows which views should be refreshed due to table changes.
func (vh *ViewRefreshHandler) getViewsToRefresh() map[string]bool {
	viewToRefresh := make(map[string]bool)
	tables := vh.GetAndClearChanges()
	for table := range tables {
		for _, viewName := range vh.viewsToRefresh[table] {
			viewToRefresh[viewName] = true
		}
	}

	return viewToRefresh
}

func (vh *ViewRefreshHandler) RegisterViewRefreshHandlers(dbDriver string) {
	switch dbDriver {
	case DBDriverTypePostgres:
		vh.registerViewRefreshHandler(getPostgresViewRefreshHandlerFunc(applicationViewName),
			applicationViewName,
			[]string{
				applicationTableName,
				resourceTableName,
				packageTableName,
				vulnerabilityTableName,
			})
		vh.registerViewRefreshHandler(getPostgresViewRefreshHandlerFunc(resourceViewName),
			resourceViewName,
			[]string{
				applicationTableName,
				resourceTableName,
				packageTableName,
				vulnerabilityTableName,
			})
		vh.registerViewRefreshHandler(getPostgresViewRefreshHandlerFunc(packageViewName),
			packageViewName,
			[]string{
				applicationTableName,
				resourceTableName,
				packageTableName,
				vulnerabilityTableName,
			})
		vh.registerViewRefreshHandler(getPostgresViewRefreshHandlerFunc(vulnerabilityViewName),
			vulnerabilityViewName,
			[]string{
				applicationTableName,
				resourceTableName,
				packageTableName,
				vulnerabilityTableName,
			})
	}
}

func (vh *ViewRefreshHandler) registerViewRefreshHandler(f refreshFunc, viewName string, tables []string) {
	vh.refreshFunc[viewName] = f
	vh.addViewsToRefreshByTable(viewName, tables)
}

func (vh *ViewRefreshHandler) IsSetViewRefreshHandler() bool {
	return len(vh.refreshFunc) > 0
}

func (vh *ViewRefreshHandler) addViewsToRefreshByTable(viewName string, tables []string) {
	for _, table := range tables {
		vh.viewsToRefresh[table] = append(vh.viewsToRefresh[table], viewName)
	}
}

func (db *Handler) SetMaterializedViewHandler(config *DBConfig) {
	db.ViewRefreshHandler = &ViewRefreshHandler{
		tableChanged:              make(map[string]bool),
		viewsToRefresh:            make(map[string][]string),
		viewRefreshIntervalSecond: time.Duration(config.ViewRefreshIntervalSecond) * time.Second,
		refreshFunc:               map[string]refreshFunc{},
	}
}

func (db *Handler) RefreshMaterializedViews() {
	if !db.ViewRefreshHandler.IsSetViewRefreshHandler() {
		return
	}
	for {
		<-time.After(db.ViewRefreshHandler.viewRefreshIntervalSecond)
		db.ViewRefreshHandler.runRequiredRefreshes(db.DB)
	}
}

func getPostgresViewRefreshHandlerFunc(viewName string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		if err := db.Exec(fmt.Sprintf(refreshMaterializedViewConcurrentlyCommand, viewName)).Error; err != nil {
			log.Errorf("Failed to refresh materialized %s: %v", viewName, err)
		}
	}
}

func initPostgresMaterializedViews(db *gorm.DB, viewNames []string) {
	for _, viewName := range viewNames {
		if err := db.Exec(fmt.Sprintf(refreshMaterializedViewCommand, viewName)).Error; err != nil {
			log.Fatalf("Failed to init materialized %s: %v", viewName, err)
		}
	}
}
