// Copyright © 2023 Cisco Systems, Inc. and its affiliates.
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

package gorm

import (
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/openclarity/vmclarity/backend/pkg/database/odatasql"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
)

type ODataObject struct {
	gorm.Model
	Data datatypes.JSON
}

var schemaMetas = map[string]odatasql.SchemaMeta{
	"ScanConfig": {
		Table: "scan_configs",
		Fields: odatasql.Schema{
			"id":   odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"name": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"scanFamiliesConfig": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScanFamiliesConfig"},
			},
			"scheduled": odatasql.FieldMeta{
				FieldType:             odatasql.ComplexFieldType,
				ComplexFieldSchemas:   []string{"SingleScheduleScanConfig"},
				DiscriminatorProperty: "objectType",
			},
			"scope": odatasql.FieldMeta{
				FieldType:             odatasql.ComplexFieldType,
				ComplexFieldSchemas:   []string{"AwsScanScope"},
				DiscriminatorProperty: "objectType",
			},
		},
	},
	"ScanFamiliesConfig": {
		Fields: odatasql.Schema{
			"exploits": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ExploitsConfig"},
			},
			"malware": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"MalwareConfig"},
			},
			"misconfigurations": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"MisconfigurationsConfig"},
			},
			"rootkits": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"RootkitsConfig"},
			},
			"sbom": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"SBOMConfig"},
			},
			"secrets": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"SecretsConfig"},
			},
			"vulnerabilities": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"VulnerabilitiesConfig"},
			},
		},
	},
	"ExploitsConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"MalwareConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"MisconfigurationsConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"RootkitsConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"SBOMConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"SecretsConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"VulnerabilitiesConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"SingleScheduleScanConfig": {
		Fields: odatasql.Schema{
			"objectType":    odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"operationTime": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"AwsScanScope": {
		Fields: odatasql.Schema{
			"objectType":                 odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"all":                        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"shouldScanStoppedInstances": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"instanceTagExclusion": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Tag"},
				},
			},
			"instanceTagSelector": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Tag"},
				},
			},
			"regions": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"AwsRegion"},
				},
			},
		},
	},
	"Tag": {
		Fields: odatasql.Schema{
			"key":   odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"value": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"AwsRegion": {
		Fields: odatasql.Schema{
			"id":   odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"name": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"vpcs": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"AwsVPC"},
				},
			},
		},
	},
	"AwsVPC": {
		Fields: odatasql.Schema{
			"id":   odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"name": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"securityGroups": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"AwsSecurityGroup"},
				},
			},
		},
	},
	"AwsSecurityGroup": {
		Fields: odatasql.Schema{
			"id":   odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"name": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
}

func ODataQuery(db *gorm.DB, schema string, filterString, selectString, expandString *string, top, skip *int, collection bool, result interface{}) error {
	// If we're not getting a collection, make sure the result is limited
	// to 1 item.
	if !collection {
		top = utils.IntPtr(1)
		skip = nil
	}

	// Build the raw SQL query using the odatasql library, this will also
	// parse and validate the ODATA query params.
	query, err := odatasql.BuildSQLQuery(schemaMetas, schema, filterString, selectString, expandString, top, skip)
	if err != nil {
		return fmt.Errorf("failed to build query for DB: %w", err)
	}

	// Use the query to populate "result" using the gorm finalisers so that
	// the gorm error handling processes things like no results found.
	if collection {
		if err := db.Raw(query).Find(result).Error; err != nil {
			return fmt.Errorf("failed to query DB: %w", err)
		}
	} else {
		if err := db.Raw(query).First(result).Error; err != nil {
			return fmt.Errorf("failed to query DB: %w", err)
		}
	}
	return nil
}

func ODataCount(db *gorm.DB, schema string, filterString *string) (int, error) {
	query, err := odatasql.BuildCountQuery(schemaMetas, schema, filterString)
	if err != nil {
		return 0, fmt.Errorf("failed to build query to count objects: %w", err)
	}

	var count int
	if err := db.Raw(query).Scan(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to query DB: %w", err)
	}
	return count, nil
}
