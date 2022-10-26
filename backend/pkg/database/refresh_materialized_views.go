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

const materializedViewRefreshIntervalSecond = 5

type ViewRefreshHandler struct {
	mu                 sync.Mutex
	shouldRefreshViews bool
}

func (vh *ViewRefreshHandler) SetTrue() {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	vh.shouldRefreshViews = true
}

func (vh *ViewRefreshHandler) GetAndSetFalse() bool {
	vh.mu.Lock()
	defer vh.mu.Unlock()
	shouldRefresh := vh.shouldRefreshViews
	vh.shouldRefreshViews = false
	return shouldRefresh
}

func (db *Handler) SetMaterializedViewHandler() {
	db.ViewRefreshHandler = &ViewRefreshHandler{}
}

func (db *Handler) RefreshMaterializedViews() {
	ticker := time.NewTicker(materializedViewRefreshIntervalSecond * time.Second)
	done := make(chan bool)

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if db.ViewRefreshHandler.GetAndSetFalse() {
				refreshMaterializedViews(db.DB)
			}
		}
	}
}

func refreshMaterializedViews(db *gorm.DB) {
	if err := db.Exec(fmt.Sprintf(refreshMaterializedViewCommand, applicationsView)).Error; err != nil {
		log.Fatalf("Failed to refresh materialized %s: %v", applicationsView, err)
	}
	if err := db.Exec(fmt.Sprintf(refreshMaterializedViewCommand, resourcesView)).Error; err != nil {
		log.Fatalf("Failed to refresh materialized %s: %v", resourcesView, err)
	}
	if err := db.Exec(fmt.Sprintf(refreshMaterializedViewCommand, packagesView)).Error; err != nil {
		log.Fatalf("Failed to refresh materialized %s: %v", packagesView, err)
	}
	if err := db.Exec(fmt.Sprintf(refreshMaterializedViewCommand, vulnerabilitiesView)).Error; err != nil {
		log.Fatalf("Failed to refresh materialized %s: %v", vulnerabilitiesView, err)
	}
}
