package database

import (
	"fmt"

	"gorm.io/gorm"
	_clause "gorm.io/gorm/clause"

	"github.com/cisco-open/kubei/api/server/models"
)

const (
	quickScanConfigTableName = "quickScanConfig"

	// NOTE: when changing one of the column names change also the gorm label in QuickScanConfig.
	columnQuickScanConfigID = "id"
)

type QuickScanConfig struct {
	ID string `gorm:"primarykey" faker:"-"`

	CISDockerBenchmarkEnabled bool `json:"cis_docker_benchmark_enabled,omitempty" gorm:"column:cis_docker_benchmark_enabled"`
}

type QuickScanConfigTable interface {
	Get() (*models.RuntimeQuickScanConfig, error)
	Set(conf *models.RuntimeQuickScanConfig) error
	SetDefault() error
}

type QuickScanConfigTableHandler struct {
	table *gorm.DB
}

func (QuickScanConfig) TableName() string {
	return quickScanConfigTableName
}

func DBQuickScanConfigFromAPI(runtimeQuickScanConfig *models.RuntimeQuickScanConfig) *QuickScanConfig {
	return &QuickScanConfig{
		ID:                        "1", // We want to keep a single quick scan config at a time.
		CISDockerBenchmarkEnabled: runtimeQuickScanConfig.CisDockerBenchmarkScanEnabled,
	}
}

func RuntimeQuickScanConfigFromDB(config *QuickScanConfig) *models.RuntimeQuickScanConfig {
	return &models.RuntimeQuickScanConfig{
		CisDockerBenchmarkScanEnabled: config.CISDockerBenchmarkEnabled,
	}
}

func (q *QuickScanConfigTableHandler) Get() (*models.RuntimeQuickScanConfig, error) {
	var conf QuickScanConfig
	if err := q.table.First(&conf).Error; err != nil {
		return nil, fmt.Errorf("failed to get runtime quick scan config: %v", err)
	}

	return RuntimeQuickScanConfigFromDB(&conf), nil
}

func (q *QuickScanConfigTableHandler) Set(conf *models.RuntimeQuickScanConfig) error {
	// On conflict, update record with the new one.
	clause := _clause.OnConflict{
		Columns:   []_clause.Column{{Name: columnQuickScanConfigID}},
		UpdateAll: true,
	}

	if err := q.table.Clauses(clause).Create(DBQuickScanConfigFromAPI(conf)).Error; err != nil {
		return fmt.Errorf("failed to set runtime quick scan config: %v", err)
	}

	return nil
}

func (q *QuickScanConfigTableHandler) SetDefault() error {
	err := q.Set(createDefaultRuntimeQuickScanConfig())
	if err != nil {
		return fmt.Errorf("failed to set default runtime quick scan config: %v", err)
	}

	return nil
}

func createDefaultRuntimeQuickScanConfig() *models.RuntimeQuickScanConfig {
	return &models.RuntimeQuickScanConfig{
		CisDockerBenchmarkScanEnabled: false,
	}
}
