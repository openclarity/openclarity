package database

import (
	"fmt"
	"github.com/cisco-open/kubei/api/server/models"
	"gorm.io/gorm"
)

const (
	cisDockerBenchmarkResultTableName = "cis_docker_benchmark_results"

	// NOTE: when changing one of the column names change also the gorm label in CISDockerBenchmarkResult.
	columnDockerBenchmarkResultResourceID = "resource_id"
	columnDockerBenchmarkResultCode       = "code"
)

type CISDockerBenchmarkResult struct {
	gorm.Model `faker:"-"`
	ResourceID string `faker:"-"`

	Code         string `json:"code,omitempty" gorm:"column:code" faker:"oneof: code3, code2, code1"`
	Level        int    `json:"level,omitempty" gorm:"column:level" faker:"oneof: 3, 2, 1"`
	Descriptions string `json:"descriptions" gorm:"column:descriptions" faker:"oneof: desc3, desc2, desc1"`
}

type CISDockerBenchmarkResultTable interface {
	CountPerLevel(filters *CountFilters) ([]*models.CISDockerBenchmarkLevelCount, error)
}

type CISDockerBenchmarkResultTableHandler struct {
	table *gorm.DB
	IDsView
}

func (CISDockerBenchmarkResult) TableName() string {
	return cisDockerBenchmarkResultTableName
}

func (c *CISDockerBenchmarkResultTableHandler) CountPerLevel(filters *CountFilters) ([]*models.CISDockerBenchmarkLevelCount, error) {
	var counters CISDockerBenchmarkLevelCounters
	var table []CISDockerBenchmarkResult
	var err error

	tx, err := c.setCountFilters(c.table, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to set count filters: %v", err)
	}

	if err := tx.Scan(&table).Error; err != nil {
		return nil, err
	}

	countCode := make(map[string]bool)
	for _, result := range table {
		// count code only once.
		if countCode[result.Code] {
			continue
		}
		countCode[result.Code] = true
		switch Level(result.Level) {
		case CISDockerBenchmarkLevelINFO:
			counters.TotalInfoCount++
		case CISDockerBenchmarkLevelWARN:
			counters.TotalWarnCount++
		case CISDockerBenchmarkLevelFATAL:
			counters.TotalFatalCount++
		}
	}

	return getCISDockerBenchmarkLevelCount(counters), nil
}

func (c *CISDockerBenchmarkResultTableHandler) setCountFilters(tx *gorm.DB, filters *CountFilters) (*gorm.DB, error) {
	if filters == nil {
		return tx, nil
	}

	// set application ids filter
	resIDs, err := c.IDsView.GetIDs(GetIDsParams{
		FilterIDs:    filters.ApplicationIDs,
		FilterIDType: ApplicationIDType,
		LookupIDType: ResourceIDType,
	}, true)
	if err != nil {
		return tx, fmt.Errorf("failed to get resource ids by app ids %v: %v", filters.ApplicationIDs, err)
	}
	tx = FilterIs(tx, columnDockerBenchmarkResultResourceID, resIDs)

	tx = CISDockerBenchmarkLevelFilterGte(tx, columnCISDockerBenchmarkLevelCountersHighestLevel, filters.CisDockerBenchmarkLevelGte)

	return tx, nil
}
