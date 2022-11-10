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

type viewNameSet map[string]struct{}

type ViewRefreshHandler struct {
	mu                        sync.Mutex
	viewsToRefresh            map[string]viewNameSet // map of tables that shows which views should be refreshed due to table changes
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
	var wg sync.WaitGroup
	for viewName := range viewNames {
		wg.Add(1)
		go func() {
			defer wg.Done()
			vh.refreshFunc[viewName](db)
		}()
	}
	wg.Wait()
}

// getViewsToRefresh creates a list of views that should be refreshed due to table changes.
// The viewsToRefresh is the map of tables that shows which views should be refreshed due to table changes.
func (vh *ViewRefreshHandler) getViewsToRefresh() map[string]bool {
	viewToRefresh := make(map[string]bool)
	tables := vh.GetAndClearChanges()
	for table := range tables {
		for viewName := range vh.viewsToRefresh[table] {
			viewToRefresh[viewName] = true
		}
	}

	return viewToRefresh
}

func (vh *ViewRefreshHandler) registerPostgresViewRefreshHandlers() {
	vh.refreshFunc[applicationViewName] = getPostgresViewRefreshHandlerFunc(applicationViewName)
	vh.addViewsToRefreshByTable(
		applicationViewName,
		applicationTableName,
		resourceTableName,
		packageTableName,
		vulnerabilityTableName,
	)
	vh.refreshFunc[resourceViewName] = getPostgresViewRefreshHandlerFunc(resourceViewName)
	vh.addViewsToRefreshByTable(
		resourceViewName,
		applicationTableName,
		resourceTableName,
		packageTableName,
		vulnerabilityTableName,
	)
	vh.refreshFunc[packageViewName] = getPostgresViewRefreshHandlerFunc(packageViewName)
	vh.addViewsToRefreshByTable(
		packageViewName,
		applicationTableName,
		resourceTableName,
		packageTableName,
		vulnerabilityTableName,
	)
	vh.refreshFunc[vulnerabilityViewName] = getPostgresViewRefreshHandlerFunc(vulnerabilityViewName)
	vh.addViewsToRefreshByTable(
		vulnerabilityViewName,
		applicationTableName,
		resourceTableName,
		packageTableName,
		vulnerabilityTableName,
	)
}

func (vh *ViewRefreshHandler) IsSetViewRefreshHandler() bool {
	return len(vh.refreshFunc) > 0
}

func (vh *ViewRefreshHandler) addViewsToRefreshByTable(viewName string, tables ...string) {
	for _, views := range vh.viewsToRefresh {
		delete(views, viewName)
	}
	for _, table := range tables {
		if _, ok := vh.viewsToRefresh[table]; !ok {
			vh.viewsToRefresh[table] = viewNameSet{}
		}
		vh.viewsToRefresh[table][viewName] = struct{}{}
	}
}

func (db *Handler) SetMaterializedViewHandler(config *DBConfig) {
	db.ViewRefreshHandler = &ViewRefreshHandler{
		tableChanged:              make(map[string]bool),
		viewsToRefresh:            make(map[string]viewNameSet),
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

func (db *Handler) initPostgresMaterializedViews() {
	var materializedViews = []string{
		applicationViewName,
		resourceViewName,
		packageViewName,
		vulnerabilityViewName,
	}
	for _, viewName := range materializedViews {
		if err := db.DB.Exec(fmt.Sprintf(refreshMaterializedViewCommand, viewName)).Error; err != nil {
			log.Fatalf("Failed to init materialized %s: %v", viewName, err)
		}
	}
	db.ViewRefreshHandler.registerPostgresViewRefreshHandlers()
}
