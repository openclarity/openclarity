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

	log "github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/openclarity/vmclarity/api/server/database/odatasql"
	"github.com/openclarity/vmclarity/api/server/database/odatasql/jsonsql"
	"github.com/openclarity/vmclarity/core/to"
)

var SQLVariant jsonsql.Variant

type ODataObject struct {
	ID   uint `gorm:"primarykey"`
	Data datatypes.JSON
}

var schemaMetas = map[string]odatasql.SchemaMeta{
	"AssetScanTemplate": {
		Fields: odatasql.Schema{
			"scanFamiliesConfig": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScanFamiliesConfig"},
			},
			"scannerInstanceCreationConfig": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScannerInstanceCreationConfig"},
			},
		},
	},
	"ScanTemplate": {
		Fields: odatasql.Schema{
			"scope": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"maxParallelScanners": odatasql.FieldMeta{
				FieldType: odatasql.NumberFieldType,
			},
			"timeoutSeconds": odatasql.FieldMeta{
				FieldType: odatasql.NumberFieldType,
			},
			"assetScanTemplate": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanTemplate"},
			},
		},
	},
	assetScansSchemaName: {
		Table: "asset_scans",
		Fields: odatasql.Schema{
			"id":       odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"revision": odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"asset": odatasql.FieldMeta{
				FieldType:            odatasql.RelationshipFieldType,
				RelationshipSchema:   assetSchemaName,
				RelationshipProperty: "id",
			},
			"scan": odatasql.FieldMeta{
				FieldType:            odatasql.RelationshipFieldType,
				RelationshipSchema:   "Scan",
				RelationshipProperty: "id",
			},
			"scanFamiliesConfig": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScanFamiliesConfig"},
			},
			"scannerInstanceCreationConfig": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScannerInstanceCreationConfig"},
			},
			"status": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanStatus"},
			},
			"sbom": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"SbomScan"},
			},
			"vulnerabilities": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"VulnerabilityScan"},
			},
			"malware": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"MalwareScan"},
			},
			"rootkits": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"RootkitScan"},
			},
			"secrets": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"SecretScan"},
			},
			"misconfigurations": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"MisconfigurationScan"},
			},
			"exploits": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ExploitScan"},
			},
			"infoFinder": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"InfoFinderScan"},
			},
			"stats": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanStats"},
			},
			"summary": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScanFindingsSummary"},
			},
			"findingsProcessed": odatasql.FieldMeta{
				FieldType: odatasql.BooleanFieldType,
			},
			"resourceCleanupStatus": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ResourceCleanupStatus"},
			},
			"annotations": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Annotation"},
				},
			},
			"provider": odatasql.FieldMeta{
				FieldType:            odatasql.RelationshipFieldType,
				RelationshipSchema:   providerSchemaName,
				RelationshipProperty: "id",
			},
		},
	},
	"AssetScanStats": {
		Fields: odatasql.Schema{
			"general": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanGeneralStats"},
			},
			"sbom": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanInputScanStats"},
			},
			"vulnerabilities": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanInputScanStats"},
			},
			"malware": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanInputScanStats"},
			},
			"rootkits": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanInputScanStats"},
			},
			"secrets": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanInputScanStats"},
			},
			"misconfigurations": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanInputScanStats"},
			},
			"exploits": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanInputScanStats"},
			},
			"infoFinder": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanInputScanStats"},
			},
		},
	},
	"AssetScanGeneralStats": {
		Fields: odatasql.Schema{
			"scanTime": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"AssetScanScanTime"},
				},
			},
		},
	},
	"AssetScanInputScanStats": {
		Fields: odatasql.Schema{
			"type": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"path": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"size": odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"scanTime": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"AssetScanScanTime"},
				},
			},
		},
	},
	"AssetScanScanTime": {
		Fields: odatasql.Schema{
			"startTime": odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"endTime":   odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
		},
	},
	"SbomScan": {
		Fields: odatasql.Schema{
			"packages": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Package"},
				},
			},
			"status": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScannerStatus"},
			},
		},
	},
	"Package": {
		Fields: odatasql.Schema{
			"name":     odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"version":  odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"language": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"purl":     odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"type":     odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"cpes": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			},
			"licenses": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			},
		},
	},
	"VulnerabilityScan": {
		Fields: odatasql.Schema{
			"vulnerabilities": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Vulnerability"},
				},
			},
			"status": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScannerStatus"},
			},
		},
	},
	"Vulnerability": {
		Fields: odatasql.Schema{
			"vulnerabilityName": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"description":       odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"severity":          odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"layerId":           odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"path":              odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"links": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			},
			"cvss": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"VulnerabilityCvss"},
				},
			},
			"distro": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"VulnerabilityDistro"},
			},
			"package": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"Package"},
			},
		},
	},
	"MalwareScan": {
		Fields: odatasql.Schema{
			"malware": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Malware"},
				},
			},
			"status": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScannerStatus"},
			},
		},
	},
	"Malware": {
		Fields: odatasql.Schema{
			"malwareName": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"malwareType": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"path":        odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"ruleName":    odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
		},
	},
	"SecretScan": {
		Fields: odatasql.Schema{
			"secrets": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Secret"},
				},
			},
			"status": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScannerStatus"},
			},
		},
	},
	"Secret": {
		Fields: odatasql.Schema{
			"description": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"filePath":    odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"startLine":   odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"endLine":     odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"startColumn": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"endColumn":   odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"fingerprint": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
		},
	},
	"MisconfigurationScan": {
		Fields: odatasql.Schema{
			"scanners": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			},
			"misconfigurations": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Misconfiguration"},
				},
			},
			"status": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScannerStatus"},
			},
		},
	},
	"Misconfiguration": {
		Fields: odatasql.Schema{
			"scannerName": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"location":    odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"category":    odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"id":          odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"description": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"severity":    odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"message":     odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"remediation": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
		},
	},
	"RootkitScan": {
		Fields: odatasql.Schema{
			"rootkits": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Rootkit"},
				},
			},
			"status": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScannerStatus"},
			},
		},
	},
	"Rootkit": {
		Fields: odatasql.Schema{
			"rootkitName": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"rootkitType": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"path":        odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
		},
	},
	"InfoFinderScan": {
		Fields: odatasql.Schema{
			"scanners": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			},
			"infos": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"InfoFinderInfo"},
				},
			},
			"status": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScannerStatus"},
			},
		},
	},
	"InfoFinderInfo": {
		Fields: odatasql.Schema{
			"scannerName": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"type":        odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"path":        odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"data":        odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
		},
	},
	"ExploitScan": {
		Fields: odatasql.Schema{
			"exploits": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Exploit"},
				},
			},
			"status": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScannerStatus"},
			},
		},
	},

	"Exploit": {
		Fields: odatasql.Schema{
			"name":        odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"title":       odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"description": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"cveID":       odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"sourceDB":    odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"urls": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			},
		},
	},
	scanSchemaName: {
		Table: "scans",
		Fields: odatasql.Schema{
			"id":        odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"name":      odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"revision":  odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"startTime": odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"endTime":   odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"scanConfig": odatasql.FieldMeta{
				FieldType:            odatasql.RelationshipFieldType,
				RelationshipSchema:   "ScanConfig",
				RelationshipProperty: "id",
			},
			"scope": odatasql.FieldMeta{
				FieldType:             odatasql.ComplexFieldType,
				ComplexFieldSchemas:   []string{"AwsScanScope"},
				DiscriminatorProperty: "objectType",
			},
			"timeoutSeconds": odatasql.FieldMeta{
				FieldType: odatasql.NumberFieldType,
			},
			"maxParallelScanners": odatasql.FieldMeta{
				FieldType: odatasql.NumberFieldType,
			},
			"assetScanTemplate": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanTemplate"},
			},
			"assetIDs": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType: odatasql.StringFieldType,
				},
			},
			"status": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScanStatus"},
			},
			"summary": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScanSummary"},
			},
			"annotations": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Annotation"},
				},
			},
		},
	},
	"ScanSummary": {
		Fields: odatasql.Schema{
			"jobsLeftToRun":          odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"jobsCompleted":          odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"totalPackages":          odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"totalExploits":          odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"totalMalware":           odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"totalMisconfigurations": odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"totalRootkits":          odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"totalSecrets":           odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"totalInfoFinder":        odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"totalVulnerabilities": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"VulnerabilitySeveritySummary"},
			},
			"totalPlugins": odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
		},
	},
	assetSchemaName: {
		Table: "assets",
		Fields: odatasql.Schema{
			"id":           odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"revision":     odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"firstSeen":    odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"lastSeen":     odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"terminatedOn": odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"scansCount":   odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"assetInfo": odatasql.FieldMeta{
				FieldType:             odatasql.ComplexFieldType,
				ComplexFieldSchemas:   []string{"VMInfo", "ContainerInfo", "ContainerImageInfo", "DirInfo", "PodInfo"},
				DiscriminatorProperty: "objectType",
			},
			"summary": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScanFindingsSummary"},
			},
			"providers": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:            odatasql.RelationshipFieldType,
					RelationshipSchema:   providerSchemaName,
					RelationshipProperty: "id",
				},
			},
			"annotations": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Annotation"},
				},
			},
		},
	},
	"VMInfo": {
		Fields: odatasql.Schema{
			"objectType":   odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"instanceID":   odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"location":     odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"launchTime":   odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"platform":     odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"instanceType": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"image":        odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"rootVolume": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"RootVolume"},
			},
			"tags": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Tag"},
				},
			},
			"securityGroups": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"SecurityGroup"},
				},
			},
			"instanceProvider": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
		},
	},
	"RootVolume": {
		Fields: odatasql.Schema{
			"sizeGB":    odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"encrypted": odatasql.FieldMeta{FieldType: odatasql.BooleanFieldType},
		},
	},
	"SecurityGroup": {
		Fields: odatasql.Schema{
			"id": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
		},
	},
	"ContainerImageInfo": {
		Fields: odatasql.Schema{
			"architecture": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"imageID":      odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"labels": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Tag"},
				},
			},
			"repoTags": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			},
			"repoDigests": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			},
			"objectType": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"os":         odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"size":       odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
		},
	},
	"ContainerInfo": {
		Fields: odatasql.Schema{
			"containerName": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"createdAt":     odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"containerID":   odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"image":         odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"labels": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Tag"},
				},
			},
			"objectType": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
		},
	},
	"DirInfo": {
		Fields: odatasql.Schema{
			"dirName":    odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"location":   odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"objectType": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
		},
	},
	"PodInfo": {
		Fields: odatasql.Schema{
			"podName":    odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"location":   odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"objectType": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
		},
	},
	"ScanFindingsSummary": {
		Fields: odatasql.Schema{
			"totalPackages":          odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"totalExploits":          odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"totalMalware":           odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"totalMisconfigurations": odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"totalRootkits":          odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"totalSecrets":           odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"totalInfoFinder":        odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"totalVulnerabilities": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"VulnerabilitySeveritySummary"},
			},
			"totalPlugins": odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
		},
	},
	"VulnerabilitySeveritySummary": {
		Fields: odatasql.Schema{
			"totalCriticalVulnerabilities":   odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"totalHighVulnerabilities":       odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"totalMediumVulnerabilities":     odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"totalLowVulnerabilities":        odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"totalNegligibleVulnerabilities": odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
		},
	},
	"ScanConfig": {
		Table: "scan_configs",
		Fields: odatasql.Schema{
			"id":       odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"revision": odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"name":     odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"disabled": odatasql.FieldMeta{FieldType: odatasql.BooleanFieldType},
			"scheduled": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"RuntimeScheduleScanConfig"},
			},
			"scanTemplate": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScanTemplate"},
			},
			"annotations": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Annotation"},
				},
			},
		},
	},
	"ScannerInstanceCreationConfig": {
		Fields: odatasql.Schema{
			"useSpotInstances": odatasql.FieldMeta{FieldType: odatasql.BooleanFieldType},
			"maxPrice":         odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"retryMaxAttempts": odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
		},
	},
	"RuntimeScheduleScanConfig": {
		Fields: odatasql.Schema{
			"cronLine":      odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"operationTime": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
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
			"infoFinder": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"InfoFinderConfig"},
			},
		},
	},
	"InfoFinderConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.BooleanFieldType},
			"scanners": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType: odatasql.StringFieldType,
				},
			},
		},
	},
	"ExploitsConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.BooleanFieldType},
			"scanners": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType: odatasql.StringFieldType,
				},
			},
		},
	},
	"MalwareConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.BooleanFieldType},
			"scanners": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType: odatasql.StringFieldType,
				},
			},
			"yara_directories_to_scan": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType: odatasql.StringFieldType,
				},
			},
		},
	},
	"MisconfigurationsConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.BooleanFieldType},
			"scanners": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType: odatasql.StringFieldType,
				},
			},
		},
	},
	"RootkitsConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.BooleanFieldType},
			"scanners": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType: odatasql.StringFieldType,
				},
			},
		},
	},
	"SBOMConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.BooleanFieldType},
			"analyzers": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType: odatasql.StringFieldType,
				},
			},
		},
	},
	"SecretsConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.BooleanFieldType},
			"scanners": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType: odatasql.StringFieldType,
				},
			},
		},
	},
	"VulnerabilitiesConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.BooleanFieldType},
			"scanners": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType: odatasql.StringFieldType,
				},
			},
		},
	},
	"Tag": {
		Fields: odatasql.Schema{
			"key":   odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"value": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
		},
	},
	findingSchemaName: {
		Table: "findings",
		Fields: odatasql.Schema{
			"id":        odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"revision":  odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"firstSeen": odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"lastSeen":  odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"lastSeenBy": odatasql.FieldMeta{
				FieldType:            odatasql.RelationshipFieldType,
				RelationshipSchema:   assetScansSchemaName,
				RelationshipProperty: "id",
			},
			"findingInfo": odatasql.FieldMeta{
				FieldType: odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{
					"PackageFindingInfo",
					"VulnerabilityFindingInfo",
					"MalwareFindingInfo",
					"SecretFindingInfo",
					"MisconfigurationFindingInfo",
					"RootkitFindingInfo",
					"ExploitFindingInfo",
					"InfoFinderFindingInfo",
				},
				DiscriminatorProperty: "objectType",
				DiscriminatorSchemaMapping: map[string]string{
					"PackageFindingInfo":          "Package",
					"VulnerabilityFindingInfo":    "Vulnerability",
					"MalwareFindingInfo":          "Malware",
					"SecretFindingInfo":           "Secret",
					"MisconfigurationFindingInfo": "Misconfiguration",
					"RootkitFindingInfo":          "Rootkit",
					"ExploitFindingInfo":          "Exploit",
					"InfoFinderFindingInfo":       "InfoFinder",
				},
			},
			"annotations": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Annotation"},
				},
			},
			"summary": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"FindingSummary"},
			},
		},
	},
	"FindingSummary": {
		Fields: odatasql.Schema{
			"updatedAt": odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"totalVulnerabilities": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"VulnerabilitySeveritySummary"},
			},
		},
	},
	"PackageFindingInfo": {
		Fields: odatasql.Schema{
			"objectType": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"name":       odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"version":    odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"language":   odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"purl":       odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"type":       odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"cpes": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			},
			"licenses": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			},
		},
	},
	"VulnerabilityFindingInfo": {
		Fields: odatasql.Schema{
			"objectType":        odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"vulnerabilityName": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"description":       odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"severity":          odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"layerId":           odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"path":              odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"links": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			},
			"cvss": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"VulnerabilityCvss"},
				},
			},
			"distro": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"VulnerabilityDistro"},
			},
			"package": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"Package"},
			},
		},
	},
	"VulnerabilityCvss": {
		Fields: odatasql.Schema{
			"metrics": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"VulnerabilityCvssMetrics"},
			},
			"vector":  odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"version": odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
		},
	},
	"VulnerabilityCvssMetrics": {
		Fields: odatasql.Schema{
			"baseScore":           odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"exploitabilityScore": odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"impactScore":         odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
		},
	},
	"VulnerabilityDistro": {
		Fields: odatasql.Schema{
			"iDLike": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			},
			"name":    odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"version": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
		},
	},
	"MalwareFindingInfo": {
		Fields: odatasql.Schema{
			"objectType":  odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"malwareName": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"malwareType": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"path":        odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"ruleName":    odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
		},
	},
	"SecretFindingInfo": {
		Fields: odatasql.Schema{
			"objectType":  odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"description": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"filePath":    odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"startLine":   odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"endLine":     odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"fingerprint": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
		},
	},
	"MisconfigurationFindingInfo": {
		Fields: odatasql.Schema{
			"objectType":  odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"scannerName": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"location":    odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"category":    odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"id":          odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"description": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"severity":    odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"message":     odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"remediation": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
		},
	},
	"InfoFinderFindingInfo": {
		Fields: odatasql.Schema{
			"objectType":  odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"scannerName": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"type":        odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"path":        odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"data":        odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
		},
	},
	"RootkitFindingInfo": {
		Fields: odatasql.Schema{
			"objectType":  odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"rootkitName": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"rootkitType": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"path":        odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
		},
	},
	"ExploitFindingInfo": {
		Fields: odatasql.Schema{
			"objectType":  odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"name":        odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"title":       odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"description": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"cveID":       odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"sourceDB":    odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"urls": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			},
		},
	},
	scanEstimationSchemaName: {
		Table: "scan_estimations",
		Fields: odatasql.Schema{
			"id":                      odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"startTime":               odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"endTime":                 odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"deleteAfter":             odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"ttlSecondsAfterFinished": odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"assetIDs": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType: odatasql.StringFieldType,
				},
			},
			"status": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScanEstimationStatus"},
			},
			"estimation": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"Estimation"},
			},
			"scanTemplate": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScanTemplate"},
			},
			"assetScanEstimations": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:            odatasql.RelationshipFieldType,
					RelationshipSchema:   assetScanEstimationsSchemaName,
					RelationshipProperty: "id",
				},
			},
			"summary": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScanEstimationSummary"},
			},
			"annotations": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Annotation"},
				},
			},
		},
	},
	"Estimation": {
		Fields: odatasql.Schema{
			"duration": odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"size":     odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"cost":     odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"costBreakdown": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"CostBreakdownComponent"},
				},
			},
		},
	},
	"CostBreakdownComponent": {
		Fields: odatasql.Schema{
			"operation": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"cost":      odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
		},
	},
	"ScanEstimationSummary": {
		Fields: odatasql.Schema{
			"totalScanTime": odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"totalScanSize": odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"totalScanCost": odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"jobsLeftToRun": odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"jobsCompleted": odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
		},
	},
	assetScanEstimationsSchemaName: {
		Table: "asset_scan_estimations",
		Fields: odatasql.Schema{
			"id": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"asset": odatasql.FieldMeta{
				FieldType:            odatasql.RelationshipFieldType,
				RelationshipSchema:   assetSchemaName,
				RelationshipProperty: "id",
			},
			"scanEstimation": odatasql.FieldMeta{
				FieldType:            odatasql.RelationshipFieldType,
				RelationshipSchema:   scanEstimationSchemaName,
				RelationshipProperty: "id",
			},
			"estimation": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"Estimation"},
			},
			"assetScanTemplate": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanTemplate"},
			},
			"status": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanEstimationStatus"},
			},
			"revision":                odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"startTime":               odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"endTime":                 odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"deleteAfter":             odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"ttlSecondsAfterFinished": odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"annotations": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Annotation"},
				},
			},
		},
	},
	"AssetScanEstimationStatus": {
		Fields: odatasql.Schema{
			"state":              odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"reason":             odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"message":            odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"lastTransitionTime": odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
		},
	},
	"ScanEstimationStatus": {
		Fields: odatasql.Schema{
			"state":              odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"reason":             odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"message":            odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"lastTransitionTime": odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
		},
	},
	"ResourceCleanupStatus": {
		Fields: odatasql.Schema{
			"lastTransitionTime": odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"message":            odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"reason":             odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"state":              odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
		},
	},
	"ScannerStatus": {
		Fields: odatasql.Schema{
			"state":              odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"reason":             odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"message":            odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"lastTransitionTime": odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
		},
	},
	"AssetScanStatus": {
		Fields: odatasql.Schema{
			"state":              odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"reason":             odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"message":            odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"lastTransitionTime": odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
		},
	},
	"ScanStatus": {
		Fields: odatasql.Schema{
			"state":              odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"reason":             odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"message":            odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"lastTransitionTime": odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
		},
	},
	"Annotation": {
		Fields: odatasql.Schema{
			"key":   odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"value": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
		},
	},
	providerSchemaName: {
		Table: "providers",
		Fields: odatasql.Schema{
			"id":                odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"revision":          odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
			"displayName":       odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"lastHeartbeatTime": odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"providerStatus": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ProviderStatus"},
			},
			"providerRuntimeVersion": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"annotations": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Annotation"},
				},
			},
		},
	},
	"ProviderStatus": {
		Fields: odatasql.Schema{
			"state":              odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"reason":             odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"message":            odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"lastTransitionTime": odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
		},
	},
	assetFindingSchemaName: {
		Table: "asset_findings",
		Fields: odatasql.Schema{
			"id": odatasql.FieldMeta{FieldType: odatasql.StringFieldType},
			"asset": odatasql.FieldMeta{
				FieldType:            odatasql.RelationshipFieldType,
				RelationshipSchema:   assetSchemaName,
				RelationshipProperty: "id",
			},
			"finding": odatasql.FieldMeta{
				FieldType:            odatasql.RelationshipFieldType,
				RelationshipSchema:   "Finding",
				RelationshipProperty: "id",
			},
			"firstSeen":     odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"lastSeen":      odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"invalidatedOn": odatasql.FieldMeta{FieldType: odatasql.DateTimeFieldType},
			"revision":      odatasql.FieldMeta{FieldType: odatasql.NumberFieldType},
		},
	},
}

func ODataQuery(db *gorm.DB, schema string, filterString, selectString, expandString, orderby *string, top, skip *int, collection bool, result interface{}) error {
	// If we're not getting a collection, make sure the result is limited
	// to 1 item.
	if !collection {
		top = to.Ptr(1)
		skip = nil
	}

	// Build the raw SQL query using the odatasql library, this will also
	// parse and validate the ODATA query params.
	query, err := odatasql.BuildSQLQuery(SQLVariant, schemaMetas, schema, filterString, selectString, expandString, orderby, top, skip)
	if err != nil {
		return fmt.Errorf("failed to build query for DB: %w", err)
	}

	log.Debugf("Running query - %q", query)

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
	query, err := odatasql.BuildCountQuery(SQLVariant, schemaMetas, schema, filterString)
	if err != nil {
		return 0, fmt.Errorf("failed to build query to count objects: %w", err)
	}

	var count int
	if err := db.Raw(query).Scan(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to query DB: %w", err)
	}
	return count, nil
}
