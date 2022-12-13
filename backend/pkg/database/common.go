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
	"sort"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/openclarity/kubeclarity/api/server/models"
)

const (
	dbArraySeparator = "|"
)

type (
	PkgVulID      string
	ResourcePkgID string
)

type TransactionParams struct {
	// map package.id + vulnerability.id to fix version
	FixVersions map[PkgVulID]string
	// map resource.id + package.id to analyzers list
	Analyzers map[ResourcePkgID][]string
	// map resource.id + package.id to scanners list
	Scanners map[ResourcePkgID][]string

	Timestamp time.Time

	VulnerabilitySource models.VulnerabilitySource
}

type CountFilters struct {
	ApplicationIDs             []string
	VulnerabilitySeverityGte   *string
	CisDockerBenchmarkLevelGte *string
}

func CreatePkgVulID(pkgID, vulID string) PkgVulID {
	return PkgVulID(uuid.NewV5(uuid.Nil, pkgID+"."+vulID).String())
}

func CreateResourcePkgID(resourceID, pkgID string) ResourcePkgID {
	return ResourcePkgID(uuid.NewV5(uuid.Nil, resourceID+"."+pkgID).String())
}

func FieldInTable(table, field string) string {
	return table + "." + field
}

func Paginate(page, pageSize int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page - 1) * pageSize
		return db.Offset(int(offset)).Limit(int(pageSize))
	}
}

func CreateTimeFilter(timeColName string, startTime, endTime strfmt.DateTime) string {
	return fmt.Sprintf("%v BETWEEN '%v' AND '%v'", timeColName, startTime, endTime)
}

func FilterIsBool(db *gorm.DB, column string, value *bool) *gorm.DB {
	if value == nil {
		return db
	}
	return db.Where(fmt.Sprintf("%s = ?", column), *value)
}

func FilterIs(db *gorm.DB, column string, values []string) *gorm.DB {
	if values == nil {
		return db
	}
	return db.Where(fmt.Sprintf("%s IN ?", column), values)
}

func FilterIsNot(db *gorm.DB, column string, values []string) *gorm.DB {
	if len(values) == 0 {
		return db
	}
	return db.Where(fmt.Sprintf("%s NOT IN ?", column), values)
}

func FilterIsNumber(db *gorm.DB, column string, values []int64) *gorm.DB {
	if len(values) == 0 {
		return db
	}
	return db.Where(fmt.Sprintf("%s IN ?", column), values)
}

func FilterIsNotNumber(db *gorm.DB, column string, values []int64) *gorm.DB {
	if len(values) == 0 {
		return db
	}
	return db.Where(fmt.Sprintf("%s NOT IN ?", column), values)
}

func FilterStartsWith(db *gorm.DB, column string, value *string) *gorm.DB {
	if value == nil {
		return db
	}
	// ex. WHERE CustomerName LIKE 'a%'	Finds any values that start with "a"
	return db.Where(fmt.Sprintf("%s LIKE ?", column), fmt.Sprintf("%s%%", *value))
}

func FilterEndsWith(db *gorm.DB, column string, value *string) *gorm.DB {
	if value == nil {
		return db
	}
	// ex. WHERE CustomerName LIKE '%a'	Finds any values that end with "a"
	return db.Where(fmt.Sprintf("%s LIKE ?", column), fmt.Sprintf("%%%s", *value))
}

func FilterContains(db *gorm.DB, column string, values []string) *gorm.DB {
	if len(values) == 0 {
		return db
	}
	for _, value := range values {
		// ex. WHERE CustomerName LIKE '%or%'	Finds any values that have "or" in any position
		db = db.Where(fmt.Sprintf("%s LIKE ?", column), fmt.Sprintf("%%%s%%", value))
	}
	return db
}

func FilterIsNotEmptyString(db *gorm.DB, column string) *gorm.DB {
	return db.Where(fmt.Sprintf("%s != ''", column))
}

func FilterIsEmptyString(db *gorm.DB, column string) *gorm.DB {
	return db.Where(fmt.Sprintf("%s = ''", column))
}

func FilterGte(db *gorm.DB, column string, value *int64) *gorm.DB {
	if value == nil {
		return db
	}
	return db.Where(fmt.Sprintf("%s >= ?", column), value)
}

func FilterLte(db *gorm.DB, column string, value *int64) *gorm.DB {
	if value == nil {
		return db
	}
	return db.Where(fmt.Sprintf("%s <= ?", column), value)
}

func FilterArrayContains(db *gorm.DB, column string, values []string) *gorm.DB {
	if len(values) == 0 {
		return db
	}
	for _, value := range values {
		db = db.Where(fmt.Sprintf("%s LIKE ?", column), fmt.Sprintf("%%%s%%", ToDBArrayElement(value)))
	}
	return db
}

func FilterArrayDoesntContain(db *gorm.DB, column string, values []string) *gorm.DB {
	if len(values) == 0 {
		return db
	}
	for _, value := range values {
		db = db.Where(fmt.Sprintf("%s NOT LIKE ?", column), fmt.Sprintf("%%%s%%", ToDBArrayElement(value)))
	}
	return db
}

// ArrayToDBArray Convert an array element to a DB array element.
func ArrayToDBArray(arr []string) string {
	sort.Strings(arr)
	ret := ""
	for _, s := range arr {
		ret += ToDBArrayElement(s)
	}
	return ret
}

func DBArrayToArray(str string) []string {
	log.Debugf("Converting DB array representation: %v", str)
	var ret []string
	spl := strings.Split(str, dbArraySeparator)
	for _, s := range spl {
		if s != "" {
			ret = append(ret, s)
		}
	}
	sort.Strings(ret)
	return ret
}

// ToDBArrayElement Convert an array element to a DB array element.
func ToDBArrayElement(s string) string {
	if strings.Contains(s, dbArraySeparator) {
		log.Warnf("Array element contains illegal character '|' : %v", s)
		s = strings.ReplaceAll(s, dbArraySeparator, "")
	}
	return dbArraySeparator + s + dbArraySeparator
}

func createVulnerabilitiesColumnSortOrder(sortDir string) (string, error) {
	return strings.Join(
		[]string{
			fmt.Sprintf("%v %v", columnSeverityCountersTotalCriticalCount, strings.ToLower(sortDir)),
			fmt.Sprintf("%v %v", columnSeverityCountersTotalHighCount, strings.ToLower(sortDir)),
			fmt.Sprintf("%v %v", columnSeverityCountersTotalMediumCount, strings.ToLower(sortDir)),
			fmt.Sprintf("%v %v", columnSeverityCountersTotalLowCount, strings.ToLower(sortDir)),
			fmt.Sprintf("%v %v", columnSeverityCountersTotalNegCount, strings.ToLower(sortDir)),
		}, ","), nil
}

func createCISDockerBenchmarkResultsColumnSortOrder(sortDir string) (string, error) {
	return strings.Join(
		[]string{
			fmt.Sprintf("%v %v", columnCISDockerBenchmarkLevelCountersTotalFatalCount, strings.ToLower(sortDir)),
			fmt.Sprintf("%v %v", columnCISDockerBenchmarkLevelCountersTotalWarnCount, strings.ToLower(sortDir)),
			fmt.Sprintf("%v %v", columnCISDockerBenchmarkLevelCountersTotalInfoCount, strings.ToLower(sortDir)),
		}, ","), nil
}
