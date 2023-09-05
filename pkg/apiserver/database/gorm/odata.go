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

	"github.com/openclarity/vmclarity/pkg/apiserver/database/odatasql"
	"github.com/openclarity/vmclarity/pkg/apiserver/database/odatasql/jsonsql"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
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
			"scope": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"maxParallelScanners": odatasql.FieldMeta{
				FieldType: odatasql.PrimitiveFieldType,
			},
			"timeoutSeconds": odatasql.FieldMeta{
				FieldType: odatasql.PrimitiveFieldType,
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
			"id":       odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"revision": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
			"sboms": odatasql.FieldMeta{
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
			"findingsProcessed": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"resourceCleanup":   odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
			"type": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"path": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"size": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
			"startTime": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"endTime":   odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
		},
	},
	"Package": {
		Fields: odatasql.Schema{
			"name":     odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"version":  odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"language": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"purl":     odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"type":     odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"cpes": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			},
			"licenses": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
		},
	},
	"Vulnerability": {
		Fields: odatasql.Schema{
			"vulnerabilityName": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"description":       odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"severity":          odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"layerId":           odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"path":              odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"links": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
		},
	},
	"Malware": {
		Fields: odatasql.Schema{
			"malwareName": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"malwareType": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"path":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
		},
	},
	"Secret": {
		Fields: odatasql.Schema{
			"description": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"filePath":    odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"startLine":   odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"endLine":     odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"startColumn": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"endColumn":   odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"fingerprint": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"MisconfigurationScan": {
		Fields: odatasql.Schema{
			"scanners": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			},
			"misconfigurations": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Misconfiguration"},
				},
			},
		},
	},
	"Misconfiguration": {
		Fields: odatasql.Schema{
			"scannerName":     odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"scannedPath":     odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"testCategory":    odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"testID":          odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"testDescription": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"severity":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"message":         odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"remediation":     odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
		},
	},
	"Rootkit": {
		Fields: odatasql.Schema{
			"rootkitName": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"rootkitType": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"path":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"InfoFinderScan": {
		Fields: odatasql.Schema{
			"scanners": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			},
			"infos": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"InfoFinderInfo"},
				},
			},
		},
	},
	"InfoFinderInfo": {
		Fields: odatasql.Schema{
			"scannerName": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"type":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"path":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"data":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
		},
	},
	"Exploit": {
		Fields: odatasql.Schema{
			"name":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"title":       odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"description": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"cveID":       odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"sourceDB":    odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"urls": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			},
		},
	},
	scanSchemaName: {
		Table: "scans",
		Fields: odatasql.Schema{
			"id":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"name":      odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"revision":  odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"startTime": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"endTime":   odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
				FieldType: odatasql.PrimitiveFieldType,
			},
			"maxParallelScanners": odatasql.FieldMeta{
				FieldType: odatasql.PrimitiveFieldType,
			},
			"assetScanTemplate": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanTemplate"},
			},
			"assetIDs": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType: odatasql.PrimitiveFieldType,
				},
			},
			"state": odatasql.FieldMeta{
				FieldType: odatasql.PrimitiveFieldType,
			},
			"stateMessage": odatasql.FieldMeta{
				FieldType: odatasql.PrimitiveFieldType,
			},
			"stateReason": odatasql.FieldMeta{
				FieldType: odatasql.PrimitiveFieldType,
			},
			"summary": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScanSummary"},
			},
		},
	},
	"ScanSummary": {
		Fields: odatasql.Schema{
			"jobsLeftToRun":          odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"jobsCompleted":          odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"totalPackages":          odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"totalExploits":          odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"totalMalware":           odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"totalMisconfigurations": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"totalRootkits":          odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"totalSecrets":           odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"totalInfoFinder":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"totalVulnerabilities": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"VulnerabilityScanSummary"},
			},
		},
	},
	assetSchemaName: {
		Table: "assets",
		Fields: odatasql.Schema{
			"id":           odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"revision":     odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"firstSeen":    odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"lastSeen":     odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"terminatedOn": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"scansCount":   odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"assetInfo": odatasql.FieldMeta{
				FieldType:             odatasql.ComplexFieldType,
				ComplexFieldSchemas:   []string{"VMInfo", "ContainerInfo", "ContainerImageInfo"},
				DiscriminatorProperty: "objectType",
			},
			"summary": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScanFindingsSummary"},
			},
		},
	},
	"VMInfo": {
		Fields: odatasql.Schema{
			"objectType":   odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"instanceID":   odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"location":     odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"launchTime":   odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"platform":     odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"instanceType": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"image":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
			"instanceProvider": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"RootVolume": {
		Fields: odatasql.Schema{
			"sizeGB":    odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"encrypted": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"SecurityGroup": {
		Fields: odatasql.Schema{
			"id": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"ContainerImageInfo": {
		Fields: odatasql.Schema{
			"architecture": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"id":           odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"labels": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Tag"},
				},
			},
			"name":       odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"objectType": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"os":         odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"size":       odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"ContainerInfo": {
		Fields: odatasql.Schema{
			"containerName": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"createdAt":     odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"id":            odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"image":         odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"labels": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"Tag"},
				},
			},
			"objectType": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"ScanFindingsSummary": {
		Fields: odatasql.Schema{
			"totalPackages":          odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"totalExploits":          odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"totalMalware":           odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"totalMisconfigurations": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"totalRootkits":          odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"totalSecrets":           odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"totalInfoFinder":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"totalVulnerabilities": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"VulnerabilityScanSummary"},
			},
		},
	},
	"VulnerabilityScanSummary": {
		Fields: odatasql.Schema{
			"totalCriticalVulnerabilities":   odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"totalHighVulnerabilities":       odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"totalMediumVulnerabilities":     odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"totalLowVulnerabilities":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"totalNegligibleVulnerabilities": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"ScanConfig": {
		Table: "scan_configs",
		Fields: odatasql.Schema{
			"id":       odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"revision": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"name":     odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"disabled": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"scheduled": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"RuntimeScheduleScanConfig"},
			},
			"scanTemplate": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScanTemplate"},
			},
		},
	},
	"ScannerInstanceCreationConfig": {
		Fields: odatasql.Schema{
			"useSpotInstances": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"maxPrice":         odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"retryMaxAttempts": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"RuntimeScheduleScanConfig": {
		Fields: odatasql.Schema{
			"cronLine":      odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"operationTime": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
			"enabled": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"scanners": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType: odatasql.PrimitiveFieldType,
				},
			},
		},
	},
	"ExploitsConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"scanners": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType: odatasql.PrimitiveFieldType,
				},
			},
		},
	},
	"MalwareConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"scanners": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType: odatasql.PrimitiveFieldType,
				},
			},
		},
	},
	"MisconfigurationsConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"scanners": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType: odatasql.PrimitiveFieldType,
				},
			},
		},
	},
	"RootkitsConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"scanners": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType: odatasql.PrimitiveFieldType,
				},
			},
		},
	},
	"SBOMConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"analyzers": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType: odatasql.PrimitiveFieldType,
				},
			},
		},
	},
	"SecretsConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"scanners": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType: odatasql.PrimitiveFieldType,
				},
			},
		},
	},
	"VulnerabilitiesConfig": {
		Fields: odatasql.Schema{
			"enabled": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"scanners": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType: odatasql.PrimitiveFieldType,
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
	"Finding": {
		Table: "findings",
		Fields: odatasql.Schema{
			"id": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"asset": odatasql.FieldMeta{
				FieldType:            odatasql.RelationshipFieldType,
				RelationshipSchema:   assetSchemaName,
				RelationshipProperty: "id",
			},
			"foundBy": odatasql.FieldMeta{
				FieldType:            odatasql.RelationshipFieldType,
				RelationshipSchema:   assetScansSchemaName,
				RelationshipProperty: "id",
			},
			"foundOn":       odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"invalidatedOn": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
		},
	},
	"PackageFindingInfo": {
		Fields: odatasql.Schema{
			"objectType": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"name":       odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"version":    odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"language":   odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"purl":       odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"type":       odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"cpes": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			},
			"licenses": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			},
		},
	},
	"VulnerabilityFindingInfo": {
		Fields: odatasql.Schema{
			"objectType":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"vulnerabilityName": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"description":       odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"severity":          odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"layerId":           odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"path":              odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"links": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
			"vector":  odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"version": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"VulnerabilityCvssMetrics": {
		Fields: odatasql.Schema{
			"baseScore":           odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"exploitabilityScore": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"impactScore":         odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"VulnerabilityDistro": {
		Fields: odatasql.Schema{
			"iDLike": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			},
			"name":    odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"version": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"MalwareFindingInfo": {
		Fields: odatasql.Schema{
			"objectType":  odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"malwareName": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"malwareType": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"path":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"SecretFindingInfo": {
		Fields: odatasql.Schema{
			"objectType":  odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"description": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"filePath":    odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"startLine":   odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"endLine":     odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"fingerprint": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"MisconfigurationFindingInfo": {
		Fields: odatasql.Schema{
			"objectType":      odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"scannerName":     odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"scannedPath":     odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"testCategory":    odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"testID":          odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"testDescription": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"severity":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"message":         odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"remediation":     odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"InfoFinderFindingInfo": {
		Fields: odatasql.Schema{
			"objectType":  odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"scannerName": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"type":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"path":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"data":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"RootkitFindingInfo": {
		Fields: odatasql.Schema{
			"objectType":  odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"rootkitName": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"rootkitType": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"path":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"ExploitFindingInfo": {
		Fields: odatasql.Schema{
			"objectType":  odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"name":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"title":       odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"description": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"cveID":       odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"sourceDB":    odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"urls": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			},
		},
	},
	"AssetScanStatus": {
		Fields: odatasql.Schema{
			"general": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanState"},
			},
			"sbom": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanState"},
			},
			"vulnerabilities": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanState"},
			},
			"malware": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanState"},
			},
			"rootkits": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanState"},
			},
			"secrets": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanState"},
			},
			"misconfigurations": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanState"},
			},
			"exploits": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanState"},
			},
			"infoFinder": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"AssetScanState"},
			},
		},
	},
	"AssetScanState": {
		Fields: odatasql.Schema{
			"state":              odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"lastTransitionTime": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"errors": odatasql.FieldMeta{
				FieldType:          odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			},
		},
	},
}

func ODataQuery(db *gorm.DB, schema string, filterString, selectString, expandString, orderby *string, top, skip *int, collection bool, result interface{}) error {
	// If we're not getting a collection, make sure the result is limited
	// to 1 item.
	if !collection {
		top = utils.PointerTo(1)
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
