package database

const (
	cisDockerBenchmarkResultTableName = "cis_docker_benchmark_results"
)

type CISDockerBenchmarkResult struct {
	ID string `gorm:"primarykey" faker:"-"`

	Code         string `json:"code,omitempty" gorm:"column:code"`
	Level        int64  `json:"level,omitempty" gorm:"column:level" faker:"oneof: 2, 1, 0"`
	Descriptions string `json:"descriptions" gorm:"column:descriptions"`
}

func (CISDockerBenchmarkResult) TableName() string {
	return cisDockerBenchmarkResultTableName
}
