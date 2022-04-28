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
	"gorm.io/gorm"

	"github.com/cisco-open/kubei/api/server/models"
)

const (
	cisDockerBenchmarkCheckTableName            = "cis_docker_benchmark_checks"
	applicationCisDockerBenchmarkChecksViewName = "application_cis_docker_benchmark_checks"

	columnApplicationCisDockerBenchmarkChecksViewApplicationID = "application_id"
)

type CISDockerBenchmarkCheck struct {
	ID string `gorm:"primarykey" faker:"-"` // consists of the Code name

	Code         string `json:"code,omitempty" gorm:"column:code" faker:"oneof: code3, code2, code1"`
	Level        int    `json:"level,omitempty" gorm:"column:level" faker:"oneof: 3, 2, 1"`
	Descriptions string `json:"descriptions" gorm:"column:descriptions" faker:"oneof: desc3, desc2, desc1"`
}

type CISDockerBenchmarkResultTable interface {
	CountPerLevel(filters *CountFilters) ([]*models.CISDockerBenchmarkLevelCount, error)
}

type CISDockerBenchmarkResultTableHandler struct {
	table *gorm.DB
}

func (CISDockerBenchmarkCheck) TableName() string {
	return cisDockerBenchmarkCheckTableName
}

const totalLevelCountStmnt = "SUM(total_info_count) AS total_info_count," +
	" SUM(total_warn_count) AS total_warn_count," +
	" SUM(total_fatal_count) AS total_fatal_count"

func (c *CISDockerBenchmarkResultTableHandler) CountPerLevel(filters *CountFilters) ([]*models.CISDockerBenchmarkLevelCount, error) {
	var counters CISDockerBenchmarkLevelCounters

	tx := c.setCountFilters(c.table, filters)

	if err := tx.Select(totalLevelCountStmnt).Scan(&counters).Error; err != nil {
		return nil, err
	}

	return getCISDockerBenchmarkLevelCount(counters), nil
}

func (c *CISDockerBenchmarkResultTableHandler) setCountFilters(tx *gorm.DB, filters *CountFilters) *gorm.DB {
	if filters == nil {
		return tx
	}

	tx = FilterIs(tx, columnApplicationCisDockerBenchmarkChecksViewApplicationID, filters.ApplicationIDs)

	tx = CISDockerBenchmarkLevelFilterGte(tx, columnCISDockerBenchmarkLevelCountersHighestLevel, filters.CisDockerBenchmarkLevelGte)

	return tx
}
