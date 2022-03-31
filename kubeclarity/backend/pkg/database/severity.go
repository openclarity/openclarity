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

	"gorm.io/gorm"

	"github.com/cisco-open/kubei/api/server/models"
	"github.com/cisco-open/kubei/backend/pkg/types"
)

type Severity int

const (
	NEGLIGIBLE Severity = iota
	LOW
	MEDIUM
	HIGH
	CRITICAL
)

var SeverityStringToInt = map[string]Severity{
	string(models.VulnerabilitySeverityNEGLIGIBLE): NEGLIGIBLE,
	string(models.VulnerabilitySeverityLOW):        LOW,
	string(models.VulnerabilitySeverityMEDIUM):     MEDIUM,
	string(models.VulnerabilitySeverityHIGH):       HIGH,
	string(models.VulnerabilitySeverityCRITICAL):   CRITICAL,
}

var SeverityIntToString = map[Severity]models.VulnerabilitySeverity{
	NEGLIGIBLE: models.VulnerabilitySeverityNEGLIGIBLE,
	LOW:        models.VulnerabilitySeverityLOW,
	MEDIUM:     models.VulnerabilitySeverityMEDIUM,
	HIGH:       models.VulnerabilitySeverityHIGH,
	CRITICAL:   models.VulnerabilitySeverityCRITICAL,
}

var CVSSSeverityIntToString = map[Severity]models.VulnerabilitySeverity{
	NEGLIGIBLE: "", // There is no CVSS NEGLIGIBLE severity
	LOW:        models.VulnerabilitySeverityLOW,
	MEDIUM:     models.VulnerabilitySeverityMEDIUM,
	HIGH:       models.VulnerabilitySeverityHIGH,
	CRITICAL:   models.VulnerabilitySeverityCRITICAL,
}

var ModelsVulnerabilitySeverityToInt = map[models.VulnerabilitySeverity]Severity{
	models.VulnerabilitySeverityNEGLIGIBLE: NEGLIGIBLE,
	models.VulnerabilitySeverityLOW:        LOW,
	models.VulnerabilitySeverityMEDIUM:     MEDIUM,
	models.VulnerabilitySeverityHIGH:       HIGH,
	models.VulnerabilitySeverityCRITICAL:   CRITICAL,
}

var TypesVulnerabilitySeverityToInt = map[types.VulnerabilitySeverity]Severity{
	types.VulnerabilitySeverityNEGLIGIBLE: NEGLIGIBLE,
	types.VulnerabilitySeverityLOW:        LOW,
	types.VulnerabilitySeverityMEDIUM:     MEDIUM,
	types.VulnerabilitySeverityHIGH:       HIGH,
	types.VulnerabilitySeverityCRITICAL:   CRITICAL,
}

const (
	// NOTE: when changing one of the column names change also the gorm label in SeverityCounters.
	columnSeverityCountersTotalNegCount      = "total_neg_count"
	columnSeverityCountersTotalLowCount      = "total_low_count"
	columnSeverityCountersTotalMediumCount   = "total_medium_count"
	columnSeverityCountersTotalHighCount     = "total_high_count"
	columnSeverityCountersTotalCriticalCount = "total_critical_count"
	columnSeverityCountersHighestSeverity    = "highest_severity"
)

type SeverityCounters struct {
	TotalNegCount      int `json:"total_neg_count,omitempty" gorm:"column:total_neg_count"`
	TotalLowCount      int `json:"total_low_count,omitempty" gorm:"column:total_low_count"`
	TotalMediumCount   int `json:"total_medium_count,omitempty" gorm:"column:total_medium_count"`
	TotalHighCount     int `json:"total_high_count,omitempty" gorm:"column:total_high_count"`
	TotalCriticalCount int `json:"total_critical_count,omitempty" gorm:"column:total_critical_count"`
	HighestSeverity    int `json:"highest_severity,omitempty" gorm:"column:highest_severity"`
	LowestSeverity     int `json:"lowest_severity,omitempty" gorm:"column:lowest_severity"`
}

func SeverityFilterIs(db *gorm.DB, columnName string, values []string) *gorm.DB {
	if len(values) == 0 {
		return db
	}
	return db.Where(fmt.Sprintf("%s IN ?", columnName), getSeveritiesFromSlice(values))
}

func SeverityFilterIsNot(db *gorm.DB, columnName string, values []string) *gorm.DB {
	if len(values) == 0 {
		return db
	}
	return db.Where(fmt.Sprintf("%s NOT IN ?", columnName), getSeveritiesFromSlice(values))
}

func SeverityFilterGte(db *gorm.DB, columnName string, value *string) *gorm.DB {
	if value == nil {
		return db
	}
	return db.Where(fmt.Sprintf("%s >= %d", columnName, getSeverityFromString(*value)))
}

func SeverityFilterLte(db *gorm.DB, columnName string, value *string) *gorm.DB {
	if value == nil {
		return db
	}
	return db.Where(fmt.Sprintf("%s <= %d", columnName, getSeverityFromString(*value)))
}

func getSeveritiesFromSlice(severities []string) []Severity {
	ret := make([]Severity, len(severities))

	for i, severity := range severities {
		ret[i] = getSeverityFromString(severity)
	}

	return ret
}

func getSeverityFromString(s string) Severity {
	return SeverityStringToInt[s]
}

func getVulnerabilityCount(counters SeverityCounters) []*models.VulnerabilityCount {
	var ret []*models.VulnerabilityCount

	ret = append(ret, &models.VulnerabilityCount{
		Count:    uint32(counters.TotalCriticalCount),
		Severity: models.VulnerabilitySeverityCRITICAL,
	})
	ret = append(ret, &models.VulnerabilityCount{
		Count:    uint32(counters.TotalHighCount),
		Severity: models.VulnerabilitySeverityHIGH,
	})
	ret = append(ret, &models.VulnerabilityCount{
		Count:    uint32(counters.TotalMediumCount),
		Severity: models.VulnerabilitySeverityMEDIUM,
	})
	ret = append(ret, &models.VulnerabilityCount{
		Count:    uint32(counters.TotalLowCount),
		Severity: models.VulnerabilitySeverityLOW,
	})
	ret = append(ret, &models.VulnerabilityCount{
		Count:    uint32(counters.TotalNegCount),
		Severity: models.VulnerabilitySeverityNEGLIGIBLE,
	})

	return ret
}

func getVulnerabilityWithFixCount(counters, countersWithFix SeverityCounters) []*models.VulnerabilitiesWithFix {
	var ret []*models.VulnerabilitiesWithFix

	ret = append(ret, &models.VulnerabilitiesWithFix{
		CountWithFix: uint32(countersWithFix.TotalCriticalCount),
		CountTotal:   uint32(counters.TotalCriticalCount),
		Severity:     models.VulnerabilitySeverityCRITICAL,
	})
	ret = append(ret, &models.VulnerabilitiesWithFix{
		CountWithFix: uint32(countersWithFix.TotalHighCount),
		CountTotal:   uint32(counters.TotalHighCount),
		Severity:     models.VulnerabilitySeverityHIGH,
	})
	ret = append(ret, &models.VulnerabilitiesWithFix{
		CountWithFix: uint32(countersWithFix.TotalMediumCount),
		CountTotal:   uint32(counters.TotalMediumCount),
		Severity:     models.VulnerabilitySeverityMEDIUM,
	})
	ret = append(ret, &models.VulnerabilitiesWithFix{
		CountWithFix: uint32(countersWithFix.TotalLowCount),
		CountTotal:   uint32(counters.TotalLowCount),
		Severity:     models.VulnerabilitySeverityLOW,
	})
	ret = append(ret, &models.VulnerabilitiesWithFix{
		CountWithFix: uint32(countersWithFix.TotalNegCount),
		CountTotal:   uint32(counters.TotalNegCount),
		Severity:     models.VulnerabilitySeverityNEGLIGIBLE,
	})

	return ret
}
