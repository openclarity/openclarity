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

package database

import (
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/backend/pkg/database/types"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

// nolint:gomnd,maintidx
func CreateDemoData(db types.Database) {
	// Create scopes
	scopesType := models.ScopeType{}
	err := scopesType.FromAwsAccountScope(models.AwsAccountScope{
		Regions: utils.PointerTo([]models.AwsRegion{
			{
				Name: "eu-central-1",
				Vpcs: utils.PointerTo([]models.AwsVPC{
					{
						Id: "vpc-1-from-eu-central-1",
						SecurityGroups: utils.PointerTo([]models.AwsSecurityGroup{
							{
								Id: "sg-1-from-vpc-1-from-eu-central-1",
							},
						}),
					},
					{
						Id: "vpc-2-from-eu-central-1",
						SecurityGroups: utils.PointerTo([]models.AwsSecurityGroup{
							{
								Id: "sg-1-from-vpc-2-from-eu-central-1",
							},
						}),
					},
				}),
			},
			{
				Name: "us-east-1",
				Vpcs: utils.PointerTo([]models.AwsVPC{
					{
						Id: "vpc-1-from-us-east-1",
						SecurityGroups: utils.PointerTo([]models.AwsSecurityGroup{
							{
								Id: "sg-1-from-vpc-1-from-us-east-1",
							},
						}),
					},
					{
						Id: "vpc-2-from-us-east-1",
						SecurityGroups: utils.PointerTo([]models.AwsSecurityGroup{
							{
								Id: "sg-1-from-vpc-2-from-us-east-1",
							},
						}),
					},
					{
						Id: "vpc-2-from-us-east-1",
						SecurityGroups: utils.PointerTo([]models.AwsSecurityGroup{
							{
								Id: "sg-2-from-vpc-2-from-us-east-1",
							},
						}),
					},
				}),
			},
		}),
	})
	if err != nil {
		log.Fatalf("failed to create scopes FromAwsScope: %v", err)
	}
	scopes := models.Scopes{
		ScopeInfo: &scopesType,
	}
	if _, err := db.ScopesTable().SetScopes(scopes); err != nil {
		log.Fatalf("failed to save scopes: %v", err)
	}

	//// Create scan configs
	//
	//// Scan config 1
	//scanConfig1Families := &models.ScanFamiliesConfig{
	//	Exploits: &models.ExploitsConfig{
	//		Enabled: utils.BoolPtr(false),
	//	},
	//	Malware: &models.MalwareConfig{
	//		Enabled: utils.BoolPtr(false),
	//	},
	//	Misconfigurations: &models.MisconfigurationsConfig{
	//		Enabled: utils.BoolPtr(false),
	//	},
	//	Rootkits: &models.RootkitsConfig{
	//		Enabled: utils.BoolPtr(false),
	//	},
	//	Sbom: &models.SBOMConfig{
	//		Enabled: utils.BoolPtr(true),
	//	},
	//	Secrets: &models.SecretsConfig{
	//		Enabled: utils.BoolPtr(true),
	//	},
	//	Vulnerabilities: &models.VulnerabilitiesConfig{
	//		Enabled: utils.BoolPtr(true),
	//	},
	//}
	//scanConfig1FamiliesB, err := json.Marshal(scanConfig1Families)
	//if err != nil {
	//	log.Fatalf("failed marshal scanConfig1Families: %v", err)
	//}
	//tag1 := models.Tag{
	//	Key:   utils.StringPtr("app"),
	//	Value: utils.StringPtr("my-app1"),
	//}
	//tag2 := models.Tag{
	//	Key:   utils.StringPtr("app"),
	//	Value: utils.StringPtr("my-app2"),
	//}
	//tag3 := models.Tag{
	//	Key:   utils.StringPtr("system"),
	//	Value: utils.StringPtr("sys1"),
	//}
	//tag4 := models.Tag{
	//	Key:   utils.StringPtr("system"),
	//	Value: utils.StringPtr("sys2"),
	//}
	//ScanConfig1SecurityGroups := []models.AwsSecurityGroup{
	//	{
	//		Id: utils.StringPtr("sg-1-from-vpc-1-from-eu-central-1"),
	//	},
	//}
	//ScanConfig1VPCs := []models.AwsVPC{
	//	{
	//		Id:             utils.StringPtr("vpc-1-from-eu-central-1"),
	//		SecurityGroups: &ScanConfig1SecurityGroups,
	//	},
	//}
	//ScanConfig1Regions := []models.AwsRegion{
	//	{
	//		Id:   utils.StringPtr("eu-central-1"),
	//		Vpcs: &ScanConfig1VPCs,
	//	},
	//}
	//scanConfig1SelectorTags := []models.Tag{tag1, tag2}
	//scanConfig1ExclusionTags := []models.Tag{tag3, tag4}
	//scanConfig1Scope := models.AwsScanScope{
	//	All:                        utils.BoolPtr(false),
	//	InstanceTagExclusion:       &scanConfig1ExclusionTags,
	//	InstanceTagSelector:        &scanConfig1SelectorTags,
	//	ObjectType:                 "AwsScanScope",
	//	Regions:                    &ScanConfig1Regions,
	//	ShouldScanStoppedInstances: utils.BoolPtr(false),
	//}
	//
	//var scanConfig1ScopeType models.ScanScopeType
	//
	//err = scanConfig1ScopeType.FromAwsScanScope(scanConfig1Scope)
	//if err != nil {
	//	log.Fatalf("failed to convert scanConfig1Scope: %v", err)
	//}
	//
	//scanConfig1ScopeB, err := scanConfig1ScopeType.MarshalJSON()
	//if err != nil {
	//	log.Fatalf("failed to marshal scanConfig1ScopeType: %v", err)
	//}
	//
	//single1 := models.SingleScheduleScanConfig{
	//	OperationTime: time.Now(),
	//}
	//var scanConfig1Scheduled models.RuntimeScheduleScanConfigType
	//err = scanConfig1Scheduled.FromSingleScheduleScanConfig(single1)
	//if err != nil {
	//	log.Fatalf("failed to create FromSingleScheduleScanConfig: %v", err)
	//}
	//scanConfig1ScheduledB, err := scanConfig1Scheduled.MarshalJSON()
	//
	//// Scan config 2
	//scanConfig2Families := &models.ScanFamiliesConfig{
	//	Exploits: &models.ExploitsConfig{
	//		Enabled: utils.BoolPtr(true),
	//	},
	//	Malware: &models.MalwareConfig{
	//		Enabled: utils.BoolPtr(true),
	//	},
	//	Misconfigurations: &models.MisconfigurationsConfig{
	//		Enabled: utils.BoolPtr(true),
	//	},
	//	Rootkits: &models.RootkitsConfig{
	//		Enabled: utils.BoolPtr(true),
	//	},
	//	Sbom: &models.SBOMConfig{
	//		Enabled: utils.BoolPtr(false),
	//	},
	//	Secrets: &models.SecretsConfig{
	//		Enabled: utils.BoolPtr(false),
	//	},
	//	Vulnerabilities: &models.VulnerabilitiesConfig{
	//		Enabled: utils.BoolPtr(false),
	//	},
	//}
	//scanConfig2FamiliesB, err := json.Marshal(scanConfig2Families)
	//if err != nil {
	//	log.Fatalf("failed marshal scanConfig2Families: %v", err)
	//}
	//ScanConfig2SecurityGroups := []models.AwsSecurityGroup{
	//	{
	//		Id: utils.StringPtr("sg-1-from-vpc-1-from-us-east-1"),
	//	},
	//}
	//ScanConfig2VPCs := []models.AwsVPC{
	//	{
	//		Id:             utils.StringPtr("vpc-1-from-us-east-1"),
	//		SecurityGroups: &ScanConfig2SecurityGroups,
	//	},
	//}
	//ScanConfig2Regions := []models.AwsRegion{
	//	{
	//		Id:   utils.StringPtr("us-east-1"),
	//		Vpcs: &ScanConfig2VPCs,
	//	},
	//}
	//scanConfig2SelectorTags := []models.Tag{tag2}
	//scanConfig2ExclusionTags := []models.Tag{tag4}
	//scanConfig2Scope := models.AwsScanScope{
	//	All:                        utils.BoolPtr(false),
	//	InstanceTagExclusion:       &scanConfig2ExclusionTags,
	//	InstanceTagSelector:        &scanConfig2SelectorTags,
	//	ObjectType:                 "AwsScanScope",
	//	Regions:                    &ScanConfig2Regions,
	//	ShouldScanStoppedInstances: utils.BoolPtr(true),
	//}
	//
	//var scanConfig2ScopeType models.ScanScopeType
	//
	//err = scanConfig2ScopeType.FromAwsScanScope(scanConfig2Scope)
	//if err != nil {
	//	log.Fatalf("failed to convert scanConfig2Scope: %v", err)
	//}
	//
	//scanConfig2ScopeB, err := scanConfig2ScopeType.MarshalJSON()
	//if err != nil {
	//	log.Fatalf("failed to marshal scanConfig2ScopeType: %v", err)
	//}
	//
	//single2 := models.SingleScheduleScanConfig{
	//	OperationTime: time.Now(),
	//}
	//var scanConfig2Scheduled models.RuntimeScheduleScanConfigType
	//err = scanConfig2Scheduled.FromSingleScheduleScanConfig(single2)
	//if err != nil {
	//	log.Fatalf("failed to create FromSingleScheduleScanConfig: %v", err)
	//}
	//scanConfig2ScheduledB, err := scanConfig2Scheduled.MarshalJSON()
	//
	//scanConfigs := []ScanConfig{
	//	{
	//		Base: Base{
	//			ID: uuid.NewV5(uuid.Nil, "1"),
	//		},
	//		Name:               utils.StringPtr("demo scan 1"),
	//		ScanFamiliesConfig: scanConfig1FamiliesB,
	//		Scheduled:          scanConfig1ScheduledB,
	//		Scope:              scanConfig1ScopeB,
	//	},
	//	{
	//		Base: Base{
	//			ID: uuid.NewV5(uuid.Nil, "2"),
	//		},
	//		Name:               utils.StringPtr("demo scan 2"),
	//		ScanFamiliesConfig: scanConfig2FamiliesB,
	//		Scheduled:          scanConfig2ScheduledB,
	//		Scope:              scanConfig2ScopeB,
	//	},
	//}
	//if _, err := db.ScanConfigsTable().SaveScanConfig(&scanConfigs[0]); err != nil {
	//	log.Fatalf("failed to save scan config 1: %v", err)
	//}
	//if _, err := db.ScanConfigsTable().SaveScanConfig(&scanConfigs[1]); err != nil {
	//	log.Fatalf("failed to save scan config 2: %v", err)
	//}

	// Create targets
	targets := []models.Target{
		{
			ScansCount: utils.PointerTo(100),
			Summary: &models.ScanFindingsSummary{
				TotalExploits:          utils.PointerTo(1),
				TotalMalware:           utils.PointerTo(2),
				TotalMisconfigurations: utils.PointerTo(3),
				TotalPackages:          utils.PointerTo(4),
				TotalRootkits:          utils.PointerTo(5),
				TotalSecrets:           utils.PointerTo(6),
				TotalVulnerabilities: &models.VulnerabilityScanSummary{
					TotalCriticalVulnerabilities:   utils.PointerTo(7),
					TotalHighVulnerabilities:       utils.PointerTo(8),
					TotalLowVulnerabilities:        utils.PointerTo(9),
					TotalMediumVulnerabilities:     utils.PointerTo(10),
					TotalNegligibleVulnerabilities: utils.PointerTo(11),
				},
			},
			TargetInfo: createVMInfo("i-instance-1-from-eu-central-1", "eu-central-1", models.AWS),
		},
		{
			ScansCount: utils.PointerTo(102),
			Summary: &models.ScanFindingsSummary{
				TotalExploits:          utils.PointerTo(12),
				TotalMalware:           utils.PointerTo(22),
				TotalMisconfigurations: utils.PointerTo(32),
				TotalPackages:          utils.PointerTo(42),
				TotalRootkits:          utils.PointerTo(52),
				TotalSecrets:           utils.PointerTo(62),
				TotalVulnerabilities: &models.VulnerabilityScanSummary{
					TotalCriticalVulnerabilities:   utils.PointerTo(72),
					TotalHighVulnerabilities:       utils.PointerTo(82),
					TotalLowVulnerabilities:        utils.PointerTo(92),
					TotalMediumVulnerabilities:     utils.PointerTo(102),
					TotalNegligibleVulnerabilities: utils.PointerTo(112),
				},
			},
			TargetInfo: createVMInfo("i-instance-2-from-eu-central-1", "eu-central-1", models.AWS),
		},
		{
			ScansCount: utils.PointerTo(103),
			Summary: &models.ScanFindingsSummary{
				TotalExploits:          utils.PointerTo(13),
				TotalMalware:           utils.PointerTo(23),
				TotalMisconfigurations: utils.PointerTo(33),
				TotalPackages:          utils.PointerTo(43),
				TotalRootkits:          utils.PointerTo(53),
				TotalSecrets:           utils.PointerTo(63),
				TotalVulnerabilities: &models.VulnerabilityScanSummary{
					TotalCriticalVulnerabilities:   utils.PointerTo(73),
					TotalHighVulnerabilities:       utils.PointerTo(83),
					TotalLowVulnerabilities:        utils.PointerTo(93),
					TotalMediumVulnerabilities:     utils.PointerTo(103),
					TotalNegligibleVulnerabilities: utils.PointerTo(113),
				},
			},
			TargetInfo: createVMInfo("i-instance-1-from-us-east-1", "us-east-1", models.AWS),
		},
	}
	for i, target := range targets {
		retTarget, err := db.TargetsTable().CreateTarget(target)
		if err != nil {
			log.Fatalf("failed to save target [%d]: %v", i, err)
		}
		targets[i] = retTarget
	}

	// Create scans
	//
	//// Create scan 1
	//scan1Start := time.Now()
	//scan1End := scan1Start.Add(10 * time.Hour)
	//scan1Targets := []string{targets[0].ID.String(), targets[1].ID.String()}
	//scan1TargetsB, err := json.Marshal(scan1Targets)
	//if err != nil {
	//	log.Fatalf("failed to marshal scan1Targets: %v", err)
	//}
	//scan1Summary := &models.ScanSummary{
	//	JobsCompleted:          utils.PointerTo[int](23),
	//	JobsLeftToRun:          utils.PointerTo[int](107),
	//	TotalExploits:          utils.PointerTo[int](14),
	//	TotalMalware:           utils.PointerTo[int](44),
	//	TotalMisconfigurations: utils.PointerTo[int](9),
	//	TotalPackages:          utils.PointerTo[int](4221),
	//	TotalRootkits:          utils.PointerTo[int](1),
	//	TotalSecrets:           utils.PointerTo[int](0),
	//	TotalVulnerabilities: &models.VulnerabilityScanSummary{
	//		TotalCriticalVulnerabilities:   utils.PointerTo[int](9),
	//		TotalHighVulnerabilities:       utils.PointerTo[int](12),
	//		TotalLowVulnerabilities:        utils.PointerTo[int](424),
	//		TotalMediumVulnerabilities:     utils.PointerTo[int](1551),
	//		TotalNegligibleVulnerabilities: utils.PointerTo[int](132),
	//	},
	//}
	//scan1SummaryB, err := json.Marshal(scan1Summary)
	//if err != nil {
	//	log.Fatalf("failed to marshal scan1Summary: %v", err)
	//}
	//
	//scan1ConfigSnapshot := &models.ScanConfigData{
	//	Name:               utils.PointerTo[string]("Scan Config 1"),
	//	ScanFamiliesConfig: scanConfig1Families,
	//	Scheduled:          &scanConfig1Scheduled,
	//	Scope:              &scanConfig1ScopeType,
	//}
	//scan1ConfigSnapshotB, err := json.Marshal(scan1ConfigSnapshot)
	//if err != nil {
	//	log.Fatalf("failed to marshal scan1ConfigSnapshot: %v", err)
	//}
	//
	//// Create scan 2
	//scan2Start := time.Now()
	//scan2Targets := []string{targets[2].ID.String()}
	//scan2TargetsB, err := json.Marshal(scan2Targets)
	//if err != nil {
	//	log.Fatalf("failed to marshal scan2TargetsB: %v", err)
	//}
	//
	//scan2Summary := &models.ScanSummary{
	//	JobsCompleted:          utils.PointerTo[int](77),
	//	JobsLeftToRun:          utils.PointerTo[int](98),
	//	TotalExploits:          utils.PointerTo[int](6),
	//	TotalMalware:           utils.PointerTo[int](0),
	//	TotalMisconfigurations: utils.PointerTo[int](75),
	//	TotalPackages:          utils.PointerTo[int](9778),
	//	TotalRootkits:          utils.PointerTo[int](5),
	//	TotalSecrets:           utils.PointerTo[int](557),
	//	TotalVulnerabilities: &models.VulnerabilityScanSummary{
	//		TotalCriticalVulnerabilities:   utils.PointerTo[int](11),
	//		TotalHighVulnerabilities:       utils.PointerTo[int](52),
	//		TotalLowVulnerabilities:        utils.PointerTo[int](241),
	//		TotalMediumVulnerabilities:     utils.PointerTo[int](8543),
	//		TotalNegligibleVulnerabilities: utils.PointerTo[int](73),
	//	},
	//}
	//scan2SummaryB, err := json.Marshal(scan2Summary)
	//if err != nil {
	//	log.Fatalf("failed to marshal scan2Summary: %v", err)
	//}
	//
	//scan2ConfigSnapshot := &models.ScanConfigData{
	//	Name:               utils.PointerTo[string]("Scan Config 2"),
	//	ScanFamiliesConfig: scanConfig2Families,
	//	Scheduled:          &scanConfig2Scheduled,
	//	Scope:              &scanConfig2ScopeType,
	//}
	//scan2ConfigSnapshotB, err := json.Marshal(scan2ConfigSnapshot)
	//if err != nil {
	//	log.Fatalf("failed to marshal scan2ConfigSnapshot: %v", err)
	//}
	//
	//scans := []Scan{
	//	{
	//		Base: Base{
	//			ID: uuid.NewV5(uuid.Nil, "1"),
	//		},
	//		ScanStartTime:      &scan1Start,
	//		ScanEndTime:        &scan1End,
	//		ScanConfigID:       utils.StringPtr(scanConfigs[0].ID.String()),
	//		ScanConfigSnapshot: scan1ConfigSnapshotB,
	//		State:              string(models.Done),
	//		StateMessage:       "Scan was completed successfully",
	//		StateReason:        string(models.ScanStateReasonSuccess),
	//		Summary:            scan1SummaryB,
	//		TargetIDs:          scan1TargetsB,
	//	},
	//	{
	//		Base: Base{
	//			ID: uuid.NewV5(uuid.Nil, "2"),
	//		},
	//		ScanStartTime:      &scan2Start,
	//		ScanEndTime:        nil, // not ended
	//		ScanConfigID:       utils.StringPtr(scanConfigs[1].ID.String()),
	//		ScanConfigSnapshot: scan2ConfigSnapshotB,
	//		State:              string(models.InProgress),
	//		StateMessage:       "Scan is in progress",
	//		StateReason:        string(models.ScanStateReasonSuccess),
	//		Summary:            scan2SummaryB,
	//		TargetIDs:          scan2TargetsB,
	//	},
	//}
	//if _, err := db.ScansTable().SaveScan(&scans[0]); err != nil {
	//	log.Fatalf("failed to save scan 1: %v", err)
	//}
	//if _, err := db.ScansTable().SaveScan(&scans[1]); err != nil {
	//	log.Fatalf("failed to save scan 2: %v", err)
	//}
	//
	// Create scan results
	scanResults := []models.TargetScanResult{
		{
			Scan: models.ScanRelationship{
				Id: "demo-id",
				//EndTime: utils.PointerTo(time.Now().Add(24 * time.Hour)),
				//ScanConfig: &models.ScanConfigRelationship{
				//	Id: uuid.NewV4().String(),
				//},
				//ScanConfigSnapshot: &models.ScanConfigData{
				//	Name: utils.PointerTo("ScanConfigSnapshot-1-Name"),
				//	ScanFamiliesConfig: &models.ScanFamiliesConfig{
				//		Exploits: &models.ExploitsConfig{
				//			Enabled: utils.PointerTo(true),
				//		},
				//		Malware: &models.MalwareConfig{
				//			Enabled: utils.PointerTo(true),
				//		},
				//		Misconfigurations: &models.MisconfigurationsConfig{
				//			Enabled: utils.PointerTo(false),
				//		},
				//		Rootkits: &models.RootkitsConfig{
				//			Enabled: utils.PointerTo(true),
				//		},
				//		Sbom: &models.SBOMConfig{
				//			Enabled: utils.PointerTo(false),
				//		},
				//		Secrets: &models.SecretsConfig{
				//			Enabled: utils.PointerTo(true),
				//		},
				//		Vulnerabilities: &models.VulnerabilitiesConfig{
				//			Enabled: utils.PointerTo(true),
				//		},
				//	},
				//	Scheduled: createSingleScheduleScanConfig(time.Now()),
				//	Scope: createAWSScanScopeType(models.AwsScanScope{
				//		AllRegions: utils.PointerTo(true),
				//		InstanceTagSelector: utils.PointerTo([]models.Tag{
				//			{
				//				Key:   "key",
				//				Value: "value",
				//			},
				//		}),
				//	}),
				//},
				//StartTime:   utils.PointerTo(time.Now()),
				//State:       utils.PointerTo(models.Done),
				//StateReason: utils.PointerTo(models.ScanStateReasonSuccess),
				//Summary: &models.ScanSummary{
				//	JobsCompleted:          utils.PointerTo(77),
				//	JobsLeftToRun:          utils.PointerTo(98),
				//	TotalExploits:          utils.PointerTo(6),
				//	TotalMalware:           utils.PointerTo(0),
				//	TotalMisconfigurations: utils.PointerTo(75),
				//	TotalPackages:          utils.PointerTo(9778),
				//	TotalRootkits:          utils.PointerTo(5),
				//	TotalSecrets:           utils.PointerTo(557),
				//	TotalVulnerabilities: &models.VulnerabilityScanSummary{
				//		TotalCriticalVulnerabilities:   utils.PointerTo(11),
				//		TotalHighVulnerabilities:       utils.PointerTo(52),
				//		TotalLowVulnerabilities:        utils.PointerTo(241),
				//		TotalMediumVulnerabilities:     utils.PointerTo(8543),
				//		TotalNegligibleVulnerabilities: utils.PointerTo(73),
				//	},
				//},
				//TargetIDs: utils.PointerTo([]string{*targets[0].Id}),
			},
			Summary: &models.ScanFindingsSummary{
				TotalExploits:          utils.PointerTo(6),
				TotalMalware:           utils.PointerTo(0),
				TotalMisconfigurations: utils.PointerTo(75),
				TotalPackages:          utils.PointerTo(9778),
				TotalRootkits:          utils.PointerTo(5),
				TotalSecrets:           utils.PointerTo(557),
				TotalVulnerabilities: &models.VulnerabilityScanSummary{
					TotalCriticalVulnerabilities:   utils.PointerTo(11),
					TotalHighVulnerabilities:       utils.PointerTo(52),
					TotalLowVulnerabilities:        utils.PointerTo(241),
					TotalMediumVulnerabilities:     utils.PointerTo(8543),
					TotalNegligibleVulnerabilities: utils.PointerTo(73),
				},
			},
			Target: models.TargetRelationship{
				Id: *targets[0].Id,
			},
		},
		{
			Scan: models.ScanRelationship{
				Id: "demo-id2",
				//EndTime: utils.PointerTo(time.Now().Add(24 * time.Hour)),
				//ScanConfig: &models.ScanConfigRelationship{
				//	Id: uuid.NewV4().String(),
				//},
				//ScanConfigSnapshot: &models.ScanConfigData{
				//	Name: utils.PointerTo("ScanConfigSnapshot-2-Name"),
				//	ScanFamiliesConfig: &models.ScanFamiliesConfig{
				//		Exploits: &models.ExploitsConfig{
				//			Enabled: utils.PointerTo(true),
				//		},
				//		Malware: &models.MalwareConfig{
				//			Enabled: utils.PointerTo(true),
				//		},
				//		Misconfigurations: &models.MisconfigurationsConfig{
				//			Enabled: utils.PointerTo(false),
				//		},
				//		Rootkits: &models.RootkitsConfig{
				//			Enabled: utils.PointerTo(true),
				//		},
				//		Sbom: &models.SBOMConfig{
				//			Enabled: utils.PointerTo(false),
				//		},
				//		Secrets: &models.SecretsConfig{
				//			Enabled: utils.PointerTo(true),
				//		},
				//		Vulnerabilities: &models.VulnerabilitiesConfig{
				//			Enabled: utils.PointerTo(true),
				//		},
				//	},
				//	Scheduled: createSingleScheduleScanConfig(time.Now()),
				//	Scope: createAWSScanScopeType(models.AwsScanScope{
				//		AllRegions: utils.PointerTo(true),
				//		InstanceTagSelector: utils.PointerTo([]models.Tag{
				//			{
				//				Key:   "key2",
				//				Value: "value2",
				//			},
				//		}),
				//	}),
				//},
				//StartTime:   utils.PointerTo(time.Now()),
				//State:       utils.PointerTo(models.Done),
				//StateReason: utils.PointerTo(models.ScanStateReasonSuccess),
				//Summary: &models.ScanSummary{
				//	JobsCompleted:          utils.PointerTo(77),
				//	JobsLeftToRun:          utils.PointerTo(98),
				//	TotalExploits:          utils.PointerTo(6),
				//	TotalMalware:           utils.PointerTo(0),
				//	TotalMisconfigurations: utils.PointerTo(75),
				//	TotalPackages:          utils.PointerTo(9778),
				//	TotalRootkits:          utils.PointerTo(5),
				//	TotalSecrets:           utils.PointerTo(557),
				//	TotalVulnerabilities: &models.VulnerabilityScanSummary{
				//		TotalCriticalVulnerabilities:   utils.PointerTo(11),
				//		TotalHighVulnerabilities:       utils.PointerTo(52),
				//		TotalLowVulnerabilities:        utils.PointerTo(241),
				//		TotalMediumVulnerabilities:     utils.PointerTo(8543),
				//		TotalNegligibleVulnerabilities: utils.PointerTo(73),
				//	},
				//},
				//TargetIDs: utils.PointerTo([]string{*targets[1].Id}),
			},
			Summary: &models.ScanFindingsSummary{
				TotalExploits:          utils.PointerTo(6),
				TotalMalware:           utils.PointerTo(0),
				TotalMisconfigurations: utils.PointerTo(75),
				TotalPackages:          utils.PointerTo(9778),
				TotalRootkits:          utils.PointerTo(5),
				TotalSecrets:           utils.PointerTo(557),
				TotalVulnerabilities: &models.VulnerabilityScanSummary{
					TotalCriticalVulnerabilities:   utils.PointerTo(11),
					TotalHighVulnerabilities:       utils.PointerTo(52),
					TotalLowVulnerabilities:        utils.PointerTo(241),
					TotalMediumVulnerabilities:     utils.PointerTo(8543),
					TotalNegligibleVulnerabilities: utils.PointerTo(73),
				},
			},
			Target: models.TargetRelationship{
				Id: *targets[1].Id,
			},
		},
	}
	for i, scanResult := range scanResults {
		if _, err := db.ScanResultsTable().CreateScanResult(scanResult); err != nil {
			log.Fatalf("failed to save scan results [%d]: %v", i, err)
		}
	}
}

//func createAWSScanScopeType(scope models.AwsScanScope) *models.ScanScopeType {
//	var scopeType models.ScanScopeType
//
//	if err := scopeType.FromAwsScanScope(scope); err != nil {
//		panic(err)
//	}
//	return &scopeType
//}
//
//func createSingleScheduleScanConfig(operationTime time.Time) *models.RuntimeScheduleScanConfigType {
//	var scanConfigScheduled models.RuntimeScheduleScanConfigType
//	if err := scanConfigScheduled.FromSingleScheduleScanConfig(models.SingleScheduleScanConfig{
//		OperationTime: operationTime,
//	}); err != nil {
//		panic(err)
//	}
//	return &scanConfigScheduled
//}

func createVMInfo(instanceID, location string, instanceProvider models.CloudProvider) *models.TargetType {
	info := models.TargetType{}
	err := info.FromVMInfo(models.VMInfo{
		InstanceID:       instanceID,
		InstanceProvider: &instanceProvider,
		Location:         location,
	})
	if err != nil {
		panic(err)
	}
	return &info
}
