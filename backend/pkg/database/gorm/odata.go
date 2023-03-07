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

	"github.com/openclarity/vmclarity/backend/pkg/database/odatasql"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
)

type ODataObject struct {
	gorm.Model
	Data datatypes.JSON
}

var schemaMetas = map[string]odatasql.SchemaMeta{
	targetScanResultsSchemaName: {
		Table: "scan_results",
		Fields: odatasql.Schema{
			"id": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"target": odatasql.FieldMeta{
				FieldType:          odatasql.RelationshipFieldType,
				RelationshipSchema: "Target",
			},
			"scan": odatasql.FieldMeta{
				FieldType:          odatasql.RelationshipFieldType,
				RelationshipSchema: "Scan",
			},
			"status": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"TargetScanStatus"},
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
			"summary": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScanFindingsSummary"},
			},
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
			"id": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"packageInfo": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"PackageInfo"},
			},
		},
	},
	"PackageInfo": {
		Fields: odatasql.Schema{
			"packageName":    odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"packageVersion": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
			"id": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"vulnerabilityInfo": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"VulnerabilityInfo"},
			},
		},
	},
	"VulnerabilityInfo": {
		Fields: odatasql.Schema{
			"vulnerabilityName": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"description":       odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"severity":          odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
			"id": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"malwareInfo": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"MalwareInfo"},
			},
		},
	},
	"MalwareInfo": {
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
			"id": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"secretInfo": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"SecretInfo"},
			},
		},
	},
	"SecretInfo": {
		Fields: odatasql.Schema{
			"description": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"filePath":    odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"startLine":   odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"endLine":     odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"fingerprint": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"MisconfigurationScan": {
		Fields: odatasql.Schema{
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
			"id": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"misconfigurationInfo": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"MisconfigurationInfo"},
			},
		},
	},
	"MisconfigurationInfo": {
		Fields: odatasql.Schema{
			"description": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"path":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
			"id": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"rootkitInfo": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"RootkitInfo"},
			},
		},
	},
	"RootkitInfo": {
		Fields: odatasql.Schema{
			"rootKitName": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"rootKitType": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"path":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
			"id": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"exploitInfo": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ExploitInfo"},
			},
		},
	},
	"ExploitInfo": {
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
	"Scan": {
		Fields: odatasql.Schema{
			"id":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"startTime": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"endTime":   odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"scanConfig": odatasql.FieldMeta{
				FieldType:          odatasql.RelationshipFieldType,
				RelationshipSchema: "ScanConfig",
			},
			"scanConfigSnapshot": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"ScanConfigData"},
			},
			"targetIDs": odatasql.FieldMeta{
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
			"totalVulnerabilities": odatasql.FieldMeta{
				FieldType:           odatasql.ComplexFieldType,
				ComplexFieldSchemas: []string{"VulnerabilityScanSummary"},
			},
		},
	},
	targetSchemaName: {
		Table: "targets",
		Fields: odatasql.Schema{
			"id":         odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"scansCount": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"targetInfo": odatasql.FieldMeta{
				FieldType:             odatasql.ComplexFieldType,
				ComplexFieldSchemas:   []string{"VMInfo"},
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
			"objectType":       odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"instanceID":       odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"location":         odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"instanceProvider": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
	"ScanConfigData": {
		Fields: odatasql.Schema{
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
	scopesSchemaName: {
		Table: "scopes",
		Fields: odatasql.Schema{
			"scopeInfo": odatasql.FieldMeta{
				FieldType:             odatasql.ComplexFieldType,
				ComplexFieldSchemas:   []string{"AwsAccountScope"},
				DiscriminatorProperty: "objectType",
			},
		},
	},
	"AwsAccountScope": {
		Fields: odatasql.Schema{
			"objectType": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"regions": odatasql.FieldMeta{
				FieldType: odatasql.CollectionFieldType,
				CollectionItemMeta: &odatasql.FieldMeta{
					FieldType:           odatasql.ComplexFieldType,
					ComplexFieldSchemas: []string{"AwsRegion"},
				},
			},
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
			"id": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
			"id": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"Finding": {
		Table: "findings",
		Fields: odatasql.Schema{
			"id": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			// TODO(sambetts) Make this a relationship once Scans support ODATA
			"scan": odatasql.FieldMeta{FieldType: odatasql.ComplexFieldType, ComplexFieldSchemas: []string{"Scan"}},
			// TODO(sambetts) Make this a relationship once Targets support ODATA
			"asset":         odatasql.FieldMeta{FieldType: odatasql.ComplexFieldType, ComplexFieldSchemas: []string{"Target"}},
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
				},
				DiscriminatorProperty: "objectType",
			},
		},
	},
	"PackageFindingInfo": {
		Fields: odatasql.Schema{
			"objectType":     odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"packageName":    odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"packageVersion": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"VulnerabilityFindingInfo": {
		Fields: odatasql.Schema{
			"objectType":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"vulnerabilityName": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"description":       odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"severity":          odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
			"objectType":  odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"description": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"path":        odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	"RootkitFindingInfo": {
		Fields: odatasql.Schema{
			"objectType":  odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"rootKitName": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
			"rootKitType": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
	// TODO(sambetts) Add table and other fields here when Scans are switched over to ODATA backend
	"Scan": {
		Fields: odatasql.Schema{
			"id": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
		},
	},
	// TODO(sambetts) Add table and other fields here then Targets are switched over to ODATA backend
	"Target": {
		Fields: odatasql.Schema{
			"id": odatasql.FieldMeta{FieldType: odatasql.PrimitiveFieldType},
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
