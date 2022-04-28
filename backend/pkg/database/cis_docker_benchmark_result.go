package database

import (
	"fmt"
	"github.com/cisco-open/kubei/api/server/models"
	"gorm.io/gorm"
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
	//var table []CISDockerBenchmarkLevelCounters

	tx, err := c.setCountFilters(c.table, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to set count filters: %v", err)
	}

	if err := tx.Select(totalLevelCountStmnt).Scan(&counters).Error; err != nil {
		return nil, err
	}
	//
	//for _, result := range table {
	//	counters.TotalInfoCount += result.TotalFatalCount
	//	counters.TotalInfoCount += result.TotalFatalCount
	//	counters.TotalInfoCount += result.TotalFatalCount
	//}

	return getCISDockerBenchmarkLevelCount(counters), nil
}

func (c *CISDockerBenchmarkResultTableHandler) setCountFilters(tx *gorm.DB, filters *CountFilters) (*gorm.DB, error) {
	if filters == nil {
		return tx, nil
	}

	tx = FilterIs(tx, columnApplicationCisDockerBenchmarkChecksViewApplicationID, filters.ApplicationIDs)

	tx = CISDockerBenchmarkLevelFilterGte(tx, columnCISDockerBenchmarkLevelCountersHighestLevel, filters.CisDockerBenchmarkLevelGte)

	return tx, nil
}
