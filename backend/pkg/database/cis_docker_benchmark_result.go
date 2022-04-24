package database

const (
	cisDockerBenchmarkResultTableName = "cis_docker_benchmark_results"
)

type CISDockerBenchmarkResult struct {
	ID string `gorm:"primarykey" faker:"-"`

	Code         string `json:"code,omitempty" gorm:"column:code" faker:"oneof: code3, code2, code1"`
	Level        int    `json:"level,omitempty" gorm:"column:level" faker:"oneof: 3, 2, 1"`
	Descriptions string `json:"descriptions" gorm:"column:descriptions" faker:"oneof: desc3, desc2, desc1"`
}

func (CISDockerBenchmarkResult) TableName() string {
	return cisDockerBenchmarkResultTableName
}
