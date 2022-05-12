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
	"time"

	"gorm.io/gorm"
	_clause "gorm.io/gorm/clause"
)

const (
	schedulerTableName = "scheduler"

	// NOTE: when changing one of the column names change also the gorm label in Scheduler.
	columnSchedulerID = "id"
	columnLastScanTime = "last_scan_time"
	columnStartTime = "start_time"
	columnConfig = "config"
	columnInterval = "interval"
)

type Scheduler struct {
	ID string `gorm:"primaryke0y" faker:"-"`

	LastScanTime string `json:"last_scan_time,omitempty" gorm:"column:last_scan_time"`
	StartTime string `json:"start_time,omitempty" gorm:"column:start_time"`
	Config string `json:"config,omitempty" gorm:"column:config"`
	// Interval saved in seconds
	Interval int64 `json:"interval,omitempty" gorm:"column:interval"`
}

type SchedulerTable interface {
	Get() (*Scheduler, error)
	Set(scheduler *Scheduler) error
	UpdateLastScanTime() error
	UpdateConfig(config string) error
	UpdateInterval(interval int64) error
	UpdateStartTime(t string) error
}

type SchedulerTableHandler struct {
	table *gorm.DB
}

func (Scheduler) TableName() string {
	return schedulerTableName
}

func (s *SchedulerTableHandler) Get() (*Scheduler, error) {
	var scheduler Scheduler
	if err := s.table.First(&scheduler).Error; err != nil {
		return nil, err
	}

	return &scheduler, nil
}

func (s *SchedulerTableHandler) Set(scheduler *Scheduler) error {
	// On conflict, update record with the new one.
	clause := _clause.OnConflict{
		Columns:   []_clause.Column{{Name: columnSchedulerID}},
		UpdateAll: true,
	}

	if err := s.table.Clauses(clause).Create(scheduler).Error; err != nil {
		return fmt.Errorf("failed to set scheduler: %v", err)
	}

	return nil
}

func (s *SchedulerTableHandler) UpdateLastScanTime() error {
	return s.table.Model(&Scheduler{}).Where(columnSchedulerID, "1").Update(columnLastScanTime, time.Now().UTC().Format(time.RFC3339)).Error
}

func (s *SchedulerTableHandler) UpdateConfig(config string) error {
	return s.table.Model(&Scheduler{}).Where(columnSchedulerID, "1").Update(columnConfig, config).Error
}

func (s *SchedulerTableHandler) UpdateStartTime(t string) error {
	return s.table.Model(&Scheduler{}).Where(columnSchedulerID, "1").Update(columnStartTime, t).Error
}

func (s *SchedulerTableHandler) UpdateInterval(interval int64) error {
	return s.table.Model(&Scheduler{}).Where(columnSchedulerID, "1").Update(columnInterval, interval).Error
}
