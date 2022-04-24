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

	dockle_types "github.com/Portshift/dockle/pkg/types"
	"gorm.io/gorm"

	"github.com/cisco-open/kubei/api/server/models"
)

type Level int

const (
	CISDockerBenchmarkLevelIGNORE Level = iota
	CISDockerBenchmarkLevelINFO
	CISDockerBenchmarkLevelWARN
	CISDockerBenchmarkLevelFATAL
)

var LevelStringToInt = map[string]Level{
	string(models.CISDockerBenchmarkLevelINFO):  CISDockerBenchmarkLevelINFO,
	string(models.CISDockerBenchmarkLevelWARN):  CISDockerBenchmarkLevelWARN,
	string(models.CISDockerBenchmarkLevelFATAL): CISDockerBenchmarkLevelFATAL,
}

const (
	// NOTE: when changing one of the column names change also the gorm label in CISDockerBenchmarkLevelCounters.
	columnCISDockerBenchmarkLevelCountersTotalInfoCount  = "total_info_count"
	columnCISDockerBenchmarkLevelCountersTotalWarnCount  = "total_warn_count"
	columnCISDockerBenchmarkLevelCountersTotalFatalCount = "total_fatal_count"
	columnCISDockerBenchmarkLevelCountersHighestLevel    = "highest_level"
)

type CISDockerBenchmarkLevelCounters struct {
	TotalInfoCount                 int `json:"total_info_count,omitempty" gorm:"column:total_info_count"`
	TotalWarnCount                 int `json:"total_warn_count,omitempty" gorm:"column:total_warn_count"`
	TotalFatalCount                int `json:"total_fatal_count,omitempty" gorm:"column:total_fatal_count"`
	HighestCISDockerBenchmarkLevel int `json:"highest_level,omitempty" gorm:"column:highest_level"`
	LowestCISDockerBenchmarkLevel  int `json:"lowest_level,omitempty" gorm:"column:lowest_level"`
}

func CISDockerBenchmarkLevelFilterGte(db *gorm.DB, columnName string, value *string) *gorm.DB {
	if value == nil {
		return db
	}
	return db.Where(fmt.Sprintf("%s >= %d", columnName, getCISDockerBenchmarkLevelFromString(*value)))
}

func CISDockerBenchmarkLevelFilterLte(db *gorm.DB, columnName string, value *string) *gorm.DB {
	if value == nil {
		return db
	}
	return db.Where(fmt.Sprintf("%s <= %d", columnName, getCISDockerBenchmarkLevelFromString(*value)))
}

func getCISDockerBenchmarkLevelFromString(s string) Level {
	return LevelStringToInt[s]
}

func getCISDockerBenchmarkLevelCount(counters CISDockerBenchmarkLevelCounters) []*models.CISDockerBenchmarkLevelCount {
	var ret []*models.CISDockerBenchmarkLevelCount

	ret = append(ret, &models.CISDockerBenchmarkLevelCount{
		Count: uint32(counters.TotalInfoCount),
		Level: models.CISDockerBenchmarkLevelINFO,
	})
	ret = append(ret, &models.CISDockerBenchmarkLevelCount{
		Count: uint32(counters.TotalWarnCount),
		Level: models.CISDockerBenchmarkLevelWARN,
	})
	ret = append(ret, &models.CISDockerBenchmarkLevelCount{
		Count: uint32(counters.TotalFatalCount),
		Level: models.CISDockerBenchmarkLevelFATAL,
	})

	return ret
}

func FromDockleTypeToLevel(level int64) Level {
	switch int(level) {
	case dockle_types.IgnoreLevel, dockle_types.PassLevel, dockle_types.SkipLevel:
		return CISDockerBenchmarkLevelIGNORE
	case dockle_types.InfoLevel:
		return CISDockerBenchmarkLevelINFO
	case dockle_types.WarnLevel:
		return CISDockerBenchmarkLevelWARN
	case dockle_types.FatalLevel:
		return CISDockerBenchmarkLevelFATAL
	}

	return CISDockerBenchmarkLevelIGNORE
}
