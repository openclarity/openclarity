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
	"strings"

	log "github.com/sirupsen/logrus"

	dockle_types "github.com/Portshift/dockle/pkg/types"
	"gorm.io/gorm"

	"github.com/openclarity/kubeclarity/api/server/models"
	"github.com/openclarity/kubeclarity/api/server/restapi/operations"
)

const (
	cisDockerBenchmarkCheckTableName            = "cis_d_b_checks"
	cisDockerBenchmarkChecksViewName            = "cis_d_b_checks_view"
	applicationCisDockerBenchmarkChecksViewName = "application_cis_d_b_checks"

	columnApplicationCisDockerBenchmarkChecksViewApplicationID = "application_id"
	columnCisDockerBenchmarkChecksViewResourceID               = "resource_id"
	columnCode                                                 = "code"
	columnLevel                                                = "level"
)

type CISDockerBenchmarkCheck struct {
	ID string `gorm:"primarykey" faker:"-"` // consists of the Code name

	Code         string `json:"code,omitempty" gorm:"column:code" faker:"oneof: CIS-DI-0006, CIS-DI-0005, CIS-DI-0001"`
	Level        int    `json:"level,omitempty" gorm:"column:level" faker:"oneof: 3, 2, 1"`
	Descriptions string `json:"descriptions" gorm:"column:descriptions" faker:"oneof: desc3, desc2, desc1"`
}

type CISDockerBenchmarkCheckView struct {
	CISDockerBenchmarkCheck
	ResourceID string `json:"resource_id,omitempty" gorm:"column:resource_id"`
}

type CISDockerBenchmarkResultTable interface {
	CountPerLevel(filters *CountFilters) ([]*models.CISDockerBenchmarkLevelCount, error)
	GetCISDockerBenchmarkResultsAndTotal(params operations.GetCisdockerbenchmarkresultsIDParams) ([]CISDockerBenchmarkCheckView, int64, error)
}

type CISDockerBenchmarkResultTableHandler struct {
	applicationsCisDockerBenchmarkChecksView *gorm.DB
	cisDockerBenchmarkChecksView             *gorm.DB
}

func (CISDockerBenchmarkCheck) TableName() string {
	return cisDockerBenchmarkCheckTableName
}

const totalLevelCountStmnt = "SUM(total_info_count) AS total_info_count," +
	" SUM(total_warn_count) AS total_warn_count," +
	" SUM(total_fatal_count) AS total_fatal_count"

func (c *CISDockerBenchmarkResultTableHandler) CountPerLevel(filters *CountFilters) ([]*models.CISDockerBenchmarkLevelCount, error) {
	var counters CISDockerBenchmarkLevelCounters

	tx := c.setCountFilters(c.applicationsCisDockerBenchmarkChecksView, filters)

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

func (c *CISDockerBenchmarkResultTableHandler) GetCISDockerBenchmarkResultsAndTotal(params operations.GetCisdockerbenchmarkresultsIDParams) ([]CISDockerBenchmarkCheckView, int64, error) {
	var count int64
	var cisDockerBenchmarkResults []CISDockerBenchmarkCheckView

	tx := FilterIs(c.cisDockerBenchmarkChecksView, columnCisDockerBenchmarkChecksViewResourceID, []string{params.ID})

	// get total item count with the set filters
	if err := tx.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count total: %v", err)
	}

	sortOrder, err := createCISDockerBenchmarkResultsSortOrder(params.SortKey, params.SortDir)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create sort order: %v", err)
	}

	// get specific ordered items
	if err := tx.Scopes().
		Order(sortOrder).
		Find(&cisDockerBenchmarkResults).Error; err != nil {
		return nil, 0, err
	}

	return cisDockerBenchmarkResults, count, nil
}

func CISDockerBenchmarkResultFromDB(result *CISDockerBenchmarkCheckView) *models.CISDockerBenchmarkResultsEX {
	return &models.CISDockerBenchmarkResultsEX{
		Code:  result.Code,
		Title: dockle_types.TitleMap[result.Code],
		Desc:  result.Descriptions,
		Level: convertToAPILevel(result.Level),
	}
}

func convertToAPILevel(level int) models.CISDockerBenchmarkLevel {
	switch level {
	case int(CISDockerBenchmarkLevelINFO):
		return models.CISDockerBenchmarkLevelINFO
	case int(CISDockerBenchmarkLevelWARN):
		return models.CISDockerBenchmarkLevelWARN
	case int(CISDockerBenchmarkLevelFATAL):
		return models.CISDockerBenchmarkLevelFATAL
	default:
		log.Errorf("Invalid level: %v", level)
		return models.CISDockerBenchmarkLevelINFO
	}
}

func createCISDockerBenchmarkResultsSortOrder(sortKey string, sortDir *string) (string, error) {
	sortKeyColumnName, err := getCISDockerBenchmarkResultsSortKeyColumnName(sortKey)
	if err != nil {
		return "", fmt.Errorf("failed to get sort key column name: %v", err)
	}

	return fmt.Sprintf("%v %v", sortKeyColumnName, strings.ToLower(*sortDir)), nil
}

func getCISDockerBenchmarkResultsSortKeyColumnName(key string) (string, error) {
	switch models.CISDockerBenchmarkResultsSortKey(key) {
	case models.CISDockerBenchmarkResultsSortKeyCode:
		return columnCode, nil
	case models.CISDockerBenchmarkResultsSortKeyLevel:
		return columnLevel, nil
	default:
		return "", fmt.Errorf("unknown sort key (%v)", key)
	}
}
