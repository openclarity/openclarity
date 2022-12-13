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

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	idsViewName = "ids_view"

	columnIDsViewResourceID      = "resource_id"
	columnIDsViewApplicationID   = "application_id"
	columnIDsViewPackageID       = "package_id"
	columnIDsViewVulnerabilityID = "vulnerability_id"
)

type idType string

const (
	ApplicationIDType   idType = "application"
	ResourceIDType      idType = "resource"
	PackageIDType       idType = "package"
	VulnerabilityIDType idType = "vulnerability"
)

type GetIDsParams struct {
	FilterIDs    []string // The IDs to filter by
	FilterIDType idType   // The ID type to filter by
	LookupIDType idType   // The ID type to lookup for
}

type IDsView interface {
	GetIDs(params GetIDsParams, idsShouldMatch bool) ([]string, error)
}

type IDsViewHandler struct {
	IDsView *gorm.DB
}

func (i *IDsViewHandler) GetIDs(params GetIDsParams, idsShouldMatch bool) ([]string, error) {
	ids := []string{}

	if len(params.FilterIDs) == 0 {
		// nothing to filter by, so return nil so that the caller knows
		// the difference between a query wasn't made, and there were
		// no results
		return nil, nil
	}

	lookupIDColumnName := getIDViewColumnNameByIDType(params.LookupIDType)

	// we will use Session(&gorm.Session{}) here so every call to the handler will start from scratch
	tx := i.IDsView.
		Session(&gorm.Session{}).
		Select("distinct " + lookupIDColumnName)

	filterIDColumnName := getIDViewColumnNameByIDType(params.FilterIDType)

	if idsShouldMatch {
		for _, id := range params.FilterIDs {
			// for each OR filter we need to verify that lookup id column is not null to avoid failing during Find
			tx.Or(fmt.Sprintf("%s = '%s' AND %s is not null", filterIDColumnName, id,
				lookupIDColumnName))
		}
	} else {
		for _, id := range params.FilterIDs {
			tx.Not(filterIDColumnName+" = ?", id)
		}
		// after we filter out all ids need to verify that lookup id column is not null to avoid failing during Find
		tx.Where(lookupIDColumnName + " is not null")
	}

	// Verify that query will return at least one result.
	var count int64
	if err := tx.Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to count IDs: %v", err)
	}
	if count <= 0 {
		// No need to run query, but return empty list of strings
		// because there are no results, not nil
		log.Debugf("no IDs found. pararms=%+v", params)
		return ids, nil
	}

	// Run query.
	if err := tx.Find(&ids).Error; err != nil {
		return nil, fmt.Errorf("failed to get IDs: %v", err)
	}

	return ids, nil
}

func getIDViewColumnNameByIDType(idType idType) string {
	switch idType {
	case ApplicationIDType:
		return columnIDsViewApplicationID
	case ResourceIDType:
		return columnIDsViewResourceID
	case PackageIDType:
		return columnIDsViewPackageID
	case VulnerabilityIDType:
		return columnIDsViewVulnerabilityID
	default:
		panic(fmt.Sprintf("unsupported id type: %v", idType))
	}
}
