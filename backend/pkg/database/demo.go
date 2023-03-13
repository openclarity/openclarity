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
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/backend/pkg/database/types"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

const (
	awsRegionEUCentral1    = "eu-central-1"
	awsRegionUSEast1       = "us-east-1"
	awsVPCEUCentral11      = "vpc-1-from-eu-central-1"
	awsVPCEUCentral12      = "vpc-2-from-eu-central-1"
	awsVPCUSEast11         = "vpc-1-from-us-east-1"
	awsVPCUSEast12         = "vpc-2-from-us-east-1"
	awsSGUSEast111         = "sg-1-from-vpc-1-from-us-east-1"
	awsSGUSEast121         = "sg-1-from-vpc-2-from-us-east-1"
	awsSGUSEast122         = "sg-2-from-vpc-2-from-us-east-1"
	awsSGEUCentral111      = "sg-1-from-vpc-1-from-eu-central-1"
	awsSGEUCentral121      = "sg-1-from-vpc-2-from-eu-central-1"
	awsInstanceEUCentral11 = "i-instance-1-from-eu-central-1"
	awsInstanceEUCentral12 = "i-instance-2-from-eu-central-1"
	awsInstanceUSEast11    = "i-instance-1-from-us-east-1"
)

var regions = []models.AwsRegion{
	{
		Name: awsRegionEUCentral1,
		Vpcs: utils.PointerTo([]models.AwsVPC{
			{
				Id: awsVPCEUCentral11,
				SecurityGroups: utils.PointerTo([]models.AwsSecurityGroup{
					{
						Id: awsSGEUCentral111,
					},
				}),
			},
			{
				Id: awsVPCEUCentral12,
				SecurityGroups: utils.PointerTo([]models.AwsSecurityGroup{
					{
						Id: awsSGEUCentral121,
					},
				}),
			},
		}),
	},
	{
		Name: awsRegionUSEast1,
		Vpcs: utils.PointerTo([]models.AwsVPC{
			{
				Id: awsVPCUSEast11,
				SecurityGroups: utils.PointerTo([]models.AwsSecurityGroup{
					{
						Id: awsSGUSEast111,
					},
				}),
			},
			{
				Id: awsVPCUSEast12,
				SecurityGroups: utils.PointerTo([]models.AwsSecurityGroup{
					{
						Id: awsSGUSEast121,
					},
					{
						Id: awsSGUSEast122,
					},
				}),
			},
		}),
	},
}

// nolint:gomnd,maintidx,cyclop
func CreateDemoData(db types.Database) {
	// Create scopes:
	scopes, err := createScopes()
	if err != nil {
		log.Fatalf("failed to create scopes FromAwsScope: %v", err)
	}
	if _, err := db.ScopesTable().SetScopes(scopes); err != nil {
		log.Fatalf("failed to save scopes: %v", err)
	}

	// Create scan configs:
	scanConfigs := createScanConfigs()
	for i, scanConfig := range scanConfigs {
		ret, err := db.ScanConfigsTable().CreateScanConfig(scanConfig)
		if err != nil {
			log.Fatalf("failed to create scan config [%d]: %v", i, err)
		}
		scanConfigs[i] = ret
	}

	// Create targets:
	targets := createTargets()
	for i, target := range targets {
		retTarget, err := db.TargetsTable().CreateTarget(target)
		if err != nil {
			log.Fatalf("failed to create target [%d]: %v", i, err)
		}
		targets[i] = retTarget
	}

	// Create scans:
	scans := createScans(targets, scanConfigs)
	for i, scan := range scans {
		ret, err := db.ScansTable().CreateScan(scan)
		if err != nil {
			log.Fatalf("failed to create scan [%d]: %v", i, err)
		}
		scans[i] = ret
	}

	// Create scan results:
	scanResults := createScanResults(scans)
	for i, scanResult := range scanResults {
		ret, err := db.ScanResultsTable().CreateScanResult(scanResult)
		if err != nil {
			log.Fatalf("failed to create scan result [%d]: %v", i, err)
		}
		scanResults[i] = ret
	}
}

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

func createScopes() (models.Scopes, error) {
	scopesType := models.ScopeType{}
	err := scopesType.FromAwsAccountScope(models.AwsAccountScope{
		Regions: utils.PointerTo(regions),
	})
	// nolint:wrapcheck
	return models.Scopes{
		ScopeInfo: &scopesType,
	}, err
}

func createTargets() []models.Target {
	return []models.Target{
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
			TargetInfo: createVMInfo(awsInstanceEUCentral11, awsRegionEUCentral1, models.AWS),
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
			TargetInfo: createVMInfo(awsInstanceEUCentral12, awsRegionEUCentral1, models.AWS),
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
			TargetInfo: createVMInfo(awsInstanceUSEast11, awsRegionUSEast1, models.AWS),
		},
	}
}

func createScanConfigs() []models.ScanConfig {
	// Scan config 1
	scanFamiliesConfig1 := &models.ScanFamiliesConfig{
		Exploits: &models.ExploitsConfig{
			Enabled: utils.BoolPtr(false),
		},
		Malware: &models.MalwareConfig{
			Enabled: utils.BoolPtr(false),
		},
		Misconfigurations: &models.MisconfigurationsConfig{
			Enabled: utils.BoolPtr(false),
		},
		Rootkits: &models.RootkitsConfig{
			Enabled: utils.BoolPtr(false),
		},
		Sbom: &models.SBOMConfig{
			Enabled: utils.BoolPtr(true),
		},
		Secrets: &models.SecretsConfig{
			Enabled: utils.BoolPtr(true),
		},
		Vulnerabilities: &models.VulnerabilitiesConfig{
			Enabled: utils.BoolPtr(true),
		},
	}
	tag1 := models.Tag{
		Key:   "app",
		Value: "my-app1",
	}
	tag2 := models.Tag{
		Key:   "app",
		Value: "my-app2",
	}
	tag3 := models.Tag{
		Key:   "system",
		Value: "sys1",
	}
	tag4 := models.Tag{
		Key:   "system",
		Value: "sys2",
	}
	ScanConfig1SecurityGroups := []models.AwsSecurityGroup{
		{
			Id: awsSGEUCentral111,
		},
	}
	ScanConfig1VPCs := []models.AwsVPC{
		{
			Id:             awsVPCEUCentral11,
			SecurityGroups: &ScanConfig1SecurityGroups,
		},
	}
	ScanConfig1Regions := []models.AwsRegion{
		{
			Name: awsRegionEUCentral1,
			Vpcs: &ScanConfig1VPCs,
		},
	}
	scanConfig1SelectorTags := []models.Tag{tag1, tag2}
	scanConfig1ExclusionTags := []models.Tag{tag3, tag4}
	scope1 := models.AwsScanScope{
		AllRegions:                 utils.BoolPtr(false),
		InstanceTagExclusion:       &scanConfig1ExclusionTags,
		InstanceTagSelector:        &scanConfig1SelectorTags,
		ObjectType:                 "AwsScanScope",
		Regions:                    &ScanConfig1Regions,
		ShouldScanStoppedInstances: utils.BoolPtr(false),
	}

	var scanScopeType1 models.ScanScopeType

	err := scanScopeType1.FromAwsScanScope(scope1)
	if err != nil {
		log.Fatalf("failed to convert scope1: %v", err)
	}

	single1 := models.SingleScheduleScanConfig{
		OperationTime: time.Now().Add(-10 * time.Hour),
	}
	var scheduled1 models.RuntimeScheduleScanConfigType
	err = scheduled1.FromSingleScheduleScanConfig(single1)
	if err != nil {
		log.Fatalf("failed to create FromSingleScheduleScanConfig: %v", err)
	}

	// Scan config 2
	scanFamiliesConfig2 := &models.ScanFamiliesConfig{
		Exploits: &models.ExploitsConfig{
			Enabled: utils.BoolPtr(true),
		},
		Malware: &models.MalwareConfig{
			Enabled: utils.BoolPtr(true),
		},
		Misconfigurations: &models.MisconfigurationsConfig{
			Enabled: utils.BoolPtr(true),
		},
		Rootkits: &models.RootkitsConfig{
			Enabled: utils.BoolPtr(true),
		},
		Sbom: &models.SBOMConfig{
			Enabled: utils.BoolPtr(false),
		},
		Secrets: &models.SecretsConfig{
			Enabled: utils.BoolPtr(false),
		},
		Vulnerabilities: &models.VulnerabilitiesConfig{
			Enabled: utils.BoolPtr(false),
		},
	}

	ScanConfig2SecurityGroups := []models.AwsSecurityGroup{
		{
			Id: awsSGUSEast111,
		},
	}
	ScanConfig2VPCs := []models.AwsVPC{
		{
			Id:             awsVPCUSEast11,
			SecurityGroups: &ScanConfig2SecurityGroups,
		},
	}
	ScanConfig2Regions := []models.AwsRegion{
		{
			Name: awsRegionUSEast1,
			Vpcs: &ScanConfig2VPCs,
		},
	}
	scanConfig2SelectorTags := []models.Tag{tag2}
	scanConfig2ExclusionTags := []models.Tag{tag4}
	scanConfig2Scope := models.AwsScanScope{
		AllRegions:                 utils.BoolPtr(false),
		InstanceTagExclusion:       &scanConfig2ExclusionTags,
		InstanceTagSelector:        &scanConfig2SelectorTags,
		ObjectType:                 "AwsScanScope",
		Regions:                    &ScanConfig2Regions,
		ShouldScanStoppedInstances: utils.BoolPtr(true),
	}

	var scanScopeType2 models.ScanScopeType

	err = scanScopeType2.FromAwsScanScope(scanConfig2Scope)
	if err != nil {
		log.Fatalf("failed to convert scanConfig2Scope: %v", err)
	}

	single2 := models.SingleScheduleScanConfig{
		OperationTime: time.Now().Add(-5 * time.Minute),
	}
	var scanConfig2Scheduled models.RuntimeScheduleScanConfigType
	err = scanConfig2Scheduled.FromSingleScheduleScanConfig(single2)
	if err != nil {
		log.Fatalf("failed to create FromSingleScheduleScanConfig: %v", err)
	}

	return []models.ScanConfig{
		{
			Name:               "demo scan 1",
			ScanFamiliesConfig: scanFamiliesConfig1,
			Scheduled:          &scheduled1,
			Scope:              &scanScopeType1,
		},
		{
			Name:               "demo scan 2",
			ScanFamiliesConfig: scanFamiliesConfig2,
			Scheduled:          &scanConfig2Scheduled,
			Scope:              &scanScopeType2,
		},
	}
}

func createScans(targets []models.Target, scanConfigs []models.ScanConfig) []models.Scan {
	// Create scan 1: already ended
	scan1Start := time.Now().Add(-10 * time.Hour)
	scan1End := scan1Start.Add(-5 * time.Hour)
	scan1Targets := []string{*targets[0].Id, *targets[1].Id}

	scan1Summary := &models.ScanSummary{
		JobsCompleted:          utils.PointerTo[int](23),
		JobsLeftToRun:          utils.PointerTo[int](0),
		TotalExploits:          utils.PointerTo[int](14),
		TotalMalware:           utils.PointerTo[int](44),
		TotalMisconfigurations: utils.PointerTo[int](9),
		TotalPackages:          utils.PointerTo[int](4221),
		TotalRootkits:          utils.PointerTo[int](1),
		TotalSecrets:           utils.PointerTo[int](0),
		TotalVulnerabilities: &models.VulnerabilityScanSummary{
			TotalCriticalVulnerabilities:   utils.PointerTo[int](9),
			TotalHighVulnerabilities:       utils.PointerTo[int](12),
			TotalLowVulnerabilities:        utils.PointerTo[int](424),
			TotalMediumVulnerabilities:     utils.PointerTo[int](1551),
			TotalNegligibleVulnerabilities: utils.PointerTo[int](132),
		},
	}

	scan1ConfigSnapshot := &models.ScanConfigData{
		Name:               utils.PointerTo[string]("Scan Config 1"),
		ScanFamiliesConfig: scanConfigs[0].ScanFamiliesConfig,
		Scheduled:          scanConfigs[0].Scheduled,
		Scope:              scanConfigs[0].Scope,
	}

	// Create scan 2: Running
	scan2Start := time.Now().Add(-5 * time.Minute)
	scan2Targets := []string{*targets[2].Id}

	scan2Summary := &models.ScanSummary{
		JobsCompleted:          utils.PointerTo[int](77),
		JobsLeftToRun:          utils.PointerTo[int](98),
		TotalExploits:          utils.PointerTo[int](6),
		TotalMalware:           utils.PointerTo[int](0),
		TotalMisconfigurations: utils.PointerTo[int](75),
		TotalPackages:          utils.PointerTo[int](9778),
		TotalRootkits:          utils.PointerTo[int](5),
		TotalSecrets:           utils.PointerTo[int](557),
		TotalVulnerabilities: &models.VulnerabilityScanSummary{
			TotalCriticalVulnerabilities:   utils.PointerTo[int](11),
			TotalHighVulnerabilities:       utils.PointerTo[int](52),
			TotalLowVulnerabilities:        utils.PointerTo[int](241),
			TotalMediumVulnerabilities:     utils.PointerTo[int](8543),
			TotalNegligibleVulnerabilities: utils.PointerTo[int](73),
		},
	}

	scan2ConfigSnapshot := &models.ScanConfigData{
		Name:               utils.PointerTo[string]("Scan Config 2"),
		ScanFamiliesConfig: scanConfigs[1].ScanFamiliesConfig,
		Scheduled:          scanConfigs[1].Scheduled,
		Scope:              scanConfigs[1].Scope,
	}

	return []models.Scan{
		{
			EndTime: &scan1End,
			ScanConfig: &models.ScanConfigRelationship{
				Id: *scanConfigs[0].Id,
			},
			ScanConfigSnapshot: scan1ConfigSnapshot,
			StartTime:          &scan1Start,
			State:              utils.PointerTo(models.ScanStateDone),
			StateMessage:       utils.StringPtr("Scan was completed successfully"),
			StateReason:        utils.PointerTo(models.ScanStateReasonSuccess),
			Summary:            scan1Summary,
			TargetIDs:          &scan1Targets,
		},
		{
			ScanConfig: &models.ScanConfigRelationship{
				Id: *scanConfigs[1].Id,
			},
			ScanConfigSnapshot: scan2ConfigSnapshot,
			StartTime:          &scan2Start,
			State:              utils.PointerTo(models.ScanStateInProgress),
			StateMessage:       utils.StringPtr("Scan is in progress"),
			StateReason:        nil,
			Summary:            scan2Summary,
			TargetIDs:          &scan2Targets,
		},
	}
}

func createScanResults(scans []models.Scan) []models.TargetScanResult {
	var scanResults []models.TargetScanResult
	for _, scan := range scans {
		for _, targetID := range *scan.TargetIDs {
			scanResults = append(scanResults, models.TargetScanResult{
				Exploits:          nil,
				Id:                nil,
				Malware:           nil,
				Misconfigurations: nil,
				Rootkits:          nil,
				Sboms:             nil,
				Scan: models.ScanRelationship{
					Id: *scan.Id,
				},
				Secrets: nil,
				Status:  nil,
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
					Id: targetID,
				},
				Vulnerabilities: nil,
			})
		}
	}
	return scanResults
}
