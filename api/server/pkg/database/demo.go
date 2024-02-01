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
	"context"
	"encoding/json"
	"time"

	"golang.org/x/exp/rand"

	dbtypes "github.com/openclarity/vmclarity/api/server/pkg/database/types"
	"github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
	"github.com/openclarity/vmclarity/utils/log"
)

const (
	awsRegionEUCentral1    = "eu-central-1"
	awsRegionUSEast1       = "us-east-1"
	awsVPCEUCentral11      = "vpc-1-from-eu-central-1"
	awsVPCUSEast11         = "vpc-1-from-us-east-1"
	awsSGUSEast111         = "sg-1-from-vpc-1-from-us-east-1"
	awsSGEUCentral111      = "sg-1-from-vpc-1-from-eu-central-1"
	awsInstanceEUCentral11 = "i-instance-1-from-eu-central-1"
	awsInstanceEUCentral12 = "i-instance-2-from-eu-central-1"
	awsInstanceUSEast11    = "i-instance-1-from-us-east-1"
)

// nolint:gomnd,maintidx,cyclop
func CreateDemoData(ctx context.Context, db dbtypes.Database) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	// Create scan configs:
	scanConfigs := createScanConfigs(ctx)
	for i, scanConfig := range scanConfigs {
		ret, err := db.ScanConfigsTable().CreateScanConfig(scanConfig)
		if err != nil {
			logger.Fatalf("failed to create scan config [%d]: %v", i, err)
		}
		scanConfigs[i] = ret
	}

	// Create assets:
	assets := createAssets()
	for i, asset := range assets {
		retAsset, err := db.AssetsTable().CreateAsset(asset)
		if err != nil {
			logger.Fatalf("failed to create asset [%d]: %v", i, err)
		}
		assets[i] = retAsset
	}

	// Create scans:
	scans := createScans(assets, scanConfigs)
	for i, scan := range scans {
		ret, err := db.ScansTable().CreateScan(scan)
		if err != nil {
			logger.Fatalf("failed to create scan [%d]: %v", i, err)
		}
		scans[i] = ret
	}

	// Create asset scans:
	assetScans := createAssetScans(scans)
	for i, assetScan := range assetScans {
		ret, err := db.AssetScansTable().CreateAssetScan(assetScan)
		if err != nil {
			logger.Fatalf("failed to create asset scan [%d]: %v", i, err)
		}
		assetScans[i] = ret
	}

	// Create findings
	findings := createFindings(ctx, assetScans)
	for i, finding := range findings {
		ret, err := db.FindingsTable().CreateFinding(finding)
		if err != nil {
			logger.Fatalf("failed to create finding [%d]: %v", i, err)
		}
		findings[i] = ret
	}
}

// nolint:gocognit,prealloc,cyclop,gomnd
func createFindings(ctx context.Context, assetScans []types.AssetScan) []types.Finding {
	var ret []types.Finding
	rand.Seed(uint64(time.Now().Unix()))

	for _, assetScan := range assetScans {
		var foundOn *time.Time
		if assetScan.Scan.StartTime != nil {
			foundOn = assetScan.Scan.StartTime
		} else {
			randMin := rand.Intn(59) + 1
			foundOn = types.PointerTo(time.Now().Add(time.Duration(-randMin) * time.Minute))
		}
		findingBase := types.Finding{
			Asset: &types.AssetRelationship{
				Id: assetScan.Asset.Id,
			},
			FoundBy: &types.AssetScanRelationship{
				Id: *assetScan.Id,
			},
			FindingInfo: nil,
			FoundOn:     foundOn,
			// InvalidatedOn: types.PointerTo(foundOn.Add(2 * time.Minute)),
		}
		if assetScan.Sbom != nil && assetScan.Sbom.Packages != nil {
			ret = append(ret, createPackageFindings(ctx, findingBase, *assetScan.Sbom.Packages)...)
		}
		if assetScan.Vulnerabilities != nil && assetScan.Vulnerabilities.Vulnerabilities != nil {
			ret = append(ret, createVulnerabilityFindings(ctx, findingBase, *assetScan.Vulnerabilities.Vulnerabilities)...)
		}
		if assetScan.Exploits != nil && assetScan.Exploits.Exploits != nil {
			ret = append(ret, createExploitFindings(ctx, findingBase, *assetScan.Exploits.Exploits)...)
		}
		if assetScan.Malware != nil && assetScan.Malware.Malware != nil {
			ret = append(ret, createMalwareFindings(ctx, findingBase, *assetScan.Malware.Malware)...)
		}
		if assetScan.Secrets != nil && assetScan.Secrets.Secrets != nil {
			ret = append(ret, createSecretFindings(ctx, findingBase, *assetScan.Secrets.Secrets)...)
		}
		if assetScan.Misconfigurations != nil && assetScan.Misconfigurations.Misconfigurations != nil {
			ret = append(ret, createMisconfigurationFindings(ctx, findingBase, *assetScan.Misconfigurations.Misconfigurations)...)
		}
		if assetScan.Rootkits != nil && assetScan.Rootkits.Rootkits != nil {
			ret = append(ret, createRootkitFindings(ctx, findingBase, *assetScan.Rootkits.Rootkits)...)
		}
		if assetScan.InfoFinder != nil && assetScan.InfoFinder.Infos != nil {
			ret = append(ret, createInfoFinderFindings(ctx, findingBase, *assetScan.InfoFinder.Infos)...)
		}
	}

	return ret
}

// nolint:gocognit,prealloc
func createExploitFindings(ctx context.Context, base types.Finding, exploits []types.Exploit) []types.Finding {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	var ret []types.Finding
	for _, exploit := range exploits {
		val := base
		convB, err := json.Marshal(exploit)
		if err != nil {
			logger.Errorf("Failed to marshal: %v", err)
			continue
		}
		conv := types.ExploitFindingInfo{}
		err = json.Unmarshal(convB, &conv)
		if err != nil {
			logger.Errorf("Failed to unmarshal: %v", err)
			continue
		}
		val.FindingInfo = &types.Finding_FindingInfo{}
		err = val.FindingInfo.FromExploitFindingInfo(conv)
		if err != nil {
			logger.Errorf("Failed to convert FromExploitFindingInfo: %v", err)
			continue
		}
		ret = append(ret, val)
	}

	return ret
}

// nolint:gocognit,prealloc
func createPackageFindings(ctx context.Context, base types.Finding, packages []types.Package) []types.Finding {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	var ret []types.Finding
	for _, pkg := range packages {
		val := base
		convB, err := json.Marshal(pkg)
		if err != nil {
			logger.Errorf("Failed to marshal: %v", err)
			continue
		}
		conv := types.PackageFindingInfo{}
		err = json.Unmarshal(convB, &conv)
		if err != nil {
			logger.Errorf("Failed to unmarshal: %v", err)
			continue
		}
		val.FindingInfo = &types.Finding_FindingInfo{}
		err = val.FindingInfo.FromPackageFindingInfo(conv)
		if err != nil {
			logger.Errorf("Failed to convert FromPackageFindingInfo: %v", err)
			continue
		}
		ret = append(ret, val)
	}

	return ret
}

// nolint:gocognit,prealloc
func createMalwareFindings(ctx context.Context, base types.Finding, malware []types.Malware) []types.Finding {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	var ret []types.Finding
	for _, mal := range malware {
		val := base
		convB, err := json.Marshal(mal)
		if err != nil {
			logger.Errorf("Failed to marshal: %v", err)
			continue
		}
		conv := types.MalwareFindingInfo{}
		err = json.Unmarshal(convB, &conv)
		if err != nil {
			logger.Errorf("Failed to unmarshal: %v", err)
			continue
		}
		val.FindingInfo = &types.Finding_FindingInfo{}
		err = val.FindingInfo.FromMalwareFindingInfo(conv)
		if err != nil {
			logger.Errorf("Failed to convert FromMalwareFindingInfo: %v", err)
			continue
		}
		ret = append(ret, val)
	}

	return ret
}

// nolint:gocognit,prealloc
func createSecretFindings(ctx context.Context, base types.Finding, secrets []types.Secret) []types.Finding {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	var ret []types.Finding
	for _, secret := range secrets {
		val := base
		convB, err := json.Marshal(secret)
		if err != nil {
			logger.Errorf("Failed to marshal: %v", err)
			continue
		}
		conv := types.SecretFindingInfo{}
		err = json.Unmarshal(convB, &conv)
		if err != nil {
			logger.Errorf("Failed to unmarshal: %v", err)
			continue
		}
		val.FindingInfo = &types.Finding_FindingInfo{}
		err = val.FindingInfo.FromSecretFindingInfo(conv)
		if err != nil {
			logger.Errorf("Failed to convert FromSecretFindingInfo: %v", err)
			continue
		}
		ret = append(ret, val)
	}

	return ret
}

// nolint:gocognit,prealloc
func createMisconfigurationFindings(ctx context.Context, base types.Finding, misconfigurations []types.Misconfiguration) []types.Finding {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	var ret []types.Finding
	for _, misconfiguration := range misconfigurations {
		val := base
		convB, err := json.Marshal(misconfiguration)
		if err != nil {
			logger.Errorf("Failed to marshal: %v", err)
			continue
		}
		conv := types.MisconfigurationFindingInfo{}
		err = json.Unmarshal(convB, &conv)
		if err != nil {
			logger.Errorf("Failed to unmarshal: %v", err)
			continue
		}
		val.FindingInfo = &types.Finding_FindingInfo{}
		err = val.FindingInfo.FromMisconfigurationFindingInfo(conv)
		if err != nil {
			logger.Errorf("Failed to convert FromMisconfigurationFindingInfo: %v", err)
			continue
		}
		ret = append(ret, val)
	}

	return ret
}

// nolint:gocognit,prealloc
func createInfoFinderFindings(ctx context.Context, base types.Finding, infos []types.InfoFinderInfo) []types.Finding {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	var ret []types.Finding
	for _, info := range infos {
		val := base
		convB, err := json.Marshal(info)
		if err != nil {
			logger.Errorf("Failed to marshal: %v", err)
			continue
		}
		conv := types.InfoFinderFindingInfo{}
		err = json.Unmarshal(convB, &conv)
		if err != nil {
			logger.Errorf("Failed to unmarshal: %v", err)
			continue
		}
		val.FindingInfo = &types.Finding_FindingInfo{}
		err = val.FindingInfo.FromInfoFinderFindingInfo(conv)
		if err != nil {
			logger.Errorf("Failed to convert FromInfoFinderFindingInfo: %v", err)
			continue
		}
		ret = append(ret, val)
	}

	return ret
}

// nolint:gocognit,prealloc
func createRootkitFindings(ctx context.Context, base types.Finding, rootkits []types.Rootkit) []types.Finding {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	var ret []types.Finding
	for _, rootkit := range rootkits {
		val := base
		convB, err := json.Marshal(rootkit)
		if err != nil {
			logger.Errorf("Failed to marshal: %v", err)
			continue
		}
		conv := types.RootkitFindingInfo{}
		err = json.Unmarshal(convB, &conv)
		if err != nil {
			logger.Errorf("Failed to unmarshal: %v", err)
			continue
		}
		val.FindingInfo = &types.Finding_FindingInfo{}
		err = val.FindingInfo.FromRootkitFindingInfo(conv)
		if err != nil {
			logger.Errorf("Failed to convert FromRootkitFindingInfo: %v", err)
			continue
		}
		ret = append(ret, val)
	}

	return ret
}

// nolint:gocognit,prealloc
func createVulnerabilityFindings(ctx context.Context, base types.Finding, vulnerabilities []types.Vulnerability) []types.Finding {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	var ret []types.Finding
	for _, vulnerability := range vulnerabilities {
		val := base
		convB, err := json.Marshal(vulnerability)
		if err != nil {
			logger.Errorf("Failed to marshal: %v", err)
			continue
		}
		conv := types.VulnerabilityFindingInfo{}
		err = json.Unmarshal(convB, &conv)
		if err != nil {
			logger.Errorf("Failed to unmarshal: %v", err)
			continue
		}
		val.FindingInfo = &types.Finding_FindingInfo{}
		err = val.FindingInfo.FromVulnerabilityFindingInfo(conv)
		if err != nil {
			logger.Errorf("Failed to convert FromVulnerabilityFindingInfo: %v", err)
			continue
		}
		ret = append(ret, val)
	}

	return ret
}

func createVMInfo(instanceID, location, image, instanceType, platform string,
	tags []types.Tag, launchTime time.Time, instanceProvider types.CloudProvider, rootVolumeSizeGB int, rootVolumeEncrypted types.RootVolumeEncrypted,
) *types.AssetType {
	info := types.AssetType{}
	err := info.FromVMInfo(types.VMInfo{
		Image:            image,
		InstanceID:       instanceID,
		InstanceProvider: &instanceProvider,
		InstanceType:     instanceType,
		LaunchTime:       launchTime,
		Location:         location,
		Platform:         platform,
		Tags:             &tags,
		RootVolume: types.RootVolume{
			Encrypted: rootVolumeEncrypted,
			SizeGB:    rootVolumeSizeGB,
		},
	})
	if err != nil {
		panic(err)
	}
	return &info
}

// nolint:gomnd
func createAssets() []types.Asset {
	return []types.Asset{
		{
			ScansCount:   types.PointerTo(1),
			FirstSeen:    types.PointerTo(time.Now()),
			LastSeen:     types.PointerTo(time.Now()),
			TerminatedOn: types.PointerTo(time.Now()),
			Summary: &types.ScanFindingsSummary{
				TotalExploits:          types.PointerTo(0),
				TotalMalware:           types.PointerTo(0),
				TotalMisconfigurations: types.PointerTo(0),
				TotalPackages:          types.PointerTo(2),
				TotalRootkits:          types.PointerTo(0),
				TotalSecrets:           types.PointerTo(3),
				TotalVulnerabilities: &types.VulnerabilityScanSummary{
					TotalCriticalVulnerabilities:   types.PointerTo(1),
					TotalHighVulnerabilities:       types.PointerTo(1),
					TotalLowVulnerabilities:        types.PointerTo(1),
					TotalMediumVulnerabilities:     types.PointerTo(0),
					TotalNegligibleVulnerabilities: types.PointerTo(0),
				},
			},
			AssetInfo: createVMInfo(awsInstanceEUCentral11, awsRegionEUCentral1+"/"+awsVPCEUCentral11+"/"+awsSGEUCentral111,
				"ami-111", "t2.large", "Linux", []types.Tag{{Key: "Name", Value: "asset1"}}, time.Now(), types.AWS, 8, types.RootVolumeEncryptedNo),
		},
		{
			ScansCount: types.PointerTo(1),
			FirstSeen:  types.PointerTo(time.Now()),
			LastSeen:   types.PointerTo(time.Now()),
			Summary: &types.ScanFindingsSummary{
				TotalExploits:          types.PointerTo(0),
				TotalMalware:           types.PointerTo(0),
				TotalMisconfigurations: types.PointerTo(0),
				TotalPackages:          types.PointerTo(2),
				TotalRootkits:          types.PointerTo(0),
				TotalSecrets:           types.PointerTo(3),
				TotalVulnerabilities: &types.VulnerabilityScanSummary{
					TotalCriticalVulnerabilities:   types.PointerTo(1),
					TotalHighVulnerabilities:       types.PointerTo(1),
					TotalLowVulnerabilities:        types.PointerTo(1),
					TotalMediumVulnerabilities:     types.PointerTo(0),
					TotalNegligibleVulnerabilities: types.PointerTo(0),
				},
			},
			AssetInfo: createVMInfo(awsInstanceEUCentral12, awsRegionEUCentral1+"/"+awsVPCEUCentral11+"/"+awsSGEUCentral111,
				"ami-111", "t2.large", "Linux", []types.Tag{{Key: "Name", Value: "asset2"}}, time.Now(), types.AWS, 25, types.RootVolumeEncryptedYes),
		},
		{
			ScansCount: types.PointerTo(1),
			FirstSeen:  types.PointerTo(time.Now()),
			Summary: &types.ScanFindingsSummary{
				TotalExploits:          types.PointerTo(2),
				TotalMalware:           types.PointerTo(3),
				TotalMisconfigurations: types.PointerTo(3),
				TotalPackages:          types.PointerTo(0),
				TotalRootkits:          types.PointerTo(3),
				TotalSecrets:           types.PointerTo(0),
				TotalVulnerabilities: &types.VulnerabilityScanSummary{
					TotalCriticalVulnerabilities:   types.PointerTo(0),
					TotalHighVulnerabilities:       types.PointerTo(0),
					TotalLowVulnerabilities:        types.PointerTo(0),
					TotalMediumVulnerabilities:     types.PointerTo(0),
					TotalNegligibleVulnerabilities: types.PointerTo(0),
				},
			},
			AssetInfo: createVMInfo(awsInstanceUSEast11, awsRegionUSEast1+"/"+awsVPCUSEast11+"/"+awsSGUSEast111,
				"ami-112", "t2.micro", "Linux", []types.Tag{{Key: "Name", Value: "asset3"}}, time.Now(), types.AWS, 512, types.RootVolumeEncryptedUnknown),
		},
	}
}

// nolint:gomnd
func createScanConfigs(_ context.Context) []types.ScanConfig {
	// Scan config 1
	scanFamiliesConfig1 := types.ScanFamiliesConfig{
		Exploits: &types.ExploitsConfig{
			Enabled: types.PointerTo(false),
		},
		InfoFinder: &types.InfoFinderConfig{
			Enabled: types.PointerTo(false),
		},
		Malware: &types.MalwareConfig{
			Enabled: types.PointerTo(false),
		},
		Misconfigurations: &types.MisconfigurationsConfig{
			Enabled: types.PointerTo(false),
		},
		Rootkits: &types.RootkitsConfig{
			Enabled: types.PointerTo(false),
		},
		Sbom: &types.SBOMConfig{
			Enabled: types.PointerTo(true),
		},
		Secrets: &types.SecretsConfig{
			Enabled: types.PointerTo(true),
		},
		Vulnerabilities: &types.VulnerabilitiesConfig{
			Enabled: types.PointerTo(true),
		},
	}

	// Scan config 2
	scanFamiliesConfig2 := types.ScanFamiliesConfig{
		Exploits: &types.ExploitsConfig{
			Enabled: types.PointerTo(true),
		},
		InfoFinder: &types.InfoFinderConfig{
			Enabled: types.PointerTo(true),
		},
		Malware: &types.MalwareConfig{
			Enabled: types.PointerTo(true),
		},
		Misconfigurations: &types.MisconfigurationsConfig{
			Enabled: types.PointerTo(true),
		},
		Rootkits: &types.RootkitsConfig{
			Enabled: types.PointerTo(true),
		},
		Sbom: &types.SBOMConfig{
			Enabled: types.PointerTo(false),
		},
		Secrets: &types.SecretsConfig{
			Enabled: types.PointerTo(false),
		},
		Vulnerabilities: &types.VulnerabilitiesConfig{
			Enabled: types.PointerTo(false),
		},
	}

	return []types.ScanConfig{
		{
			Name: types.PointerTo("Scan Config 1"),
			ScanTemplate: &types.ScanTemplate{
				Scope:               types.PointerTo("startswith(targetInfo.location, 'eu-central-1')"),
				MaxParallelScanners: types.PointerTo(2),
				AssetScanTemplate: &types.AssetScanTemplate{
					ScanFamiliesConfig: &scanFamiliesConfig1,
				},
			},
			Scheduled: &types.RuntimeScheduleScanConfig{
				OperationTime: types.PointerTo(time.Now().Add(5 * time.Hour)),
			},
		},
		{
			Name: types.PointerTo("Scan Config 2"),
			ScanTemplate: &types.ScanTemplate{
				Scope:               types.PointerTo("startswith(targetInfo.location, 'us-east-1')"),
				MaxParallelScanners: types.PointerTo(3),
				AssetScanTemplate: &types.AssetScanTemplate{
					ScanFamiliesConfig: &scanFamiliesConfig2,
					ScannerInstanceCreationConfig: &types.ScannerInstanceCreationConfig{
						MaxPrice:         types.PointerTo("1000000"),
						RetryMaxAttempts: types.PointerTo(4),
						UseSpotInstances: true,
					},
				},
			},
			Scheduled: &types.RuntimeScheduleScanConfig{
				CronLine: types.PointerTo("0 */4 * * *"),
			},
		},
	}
}

// nolint:gomnd
func createScans(assets []types.Asset, scanConfigs []types.ScanConfig) []types.Scan {
	// Create scan 1: already ended
	scan1Start := time.Now().Add(-10 * time.Hour)
	scan1End := scan1Start.Add(5*time.Hour + 27*time.Minute + 56*time.Second)
	scan1Assets := []string{*assets[0].Id, *assets[1].Id}

	scan1Summary := &types.ScanSummary{
		JobsCompleted:          types.PointerTo[int](2),
		JobsLeftToRun:          types.PointerTo[int](0),
		TotalExploits:          types.PointerTo[int](0),
		TotalInfoFinder:        types.PointerTo[int](0),
		TotalMalware:           types.PointerTo[int](0),
		TotalMisconfigurations: types.PointerTo[int](0),
		TotalPackages:          types.PointerTo[int](4),
		TotalRootkits:          types.PointerTo[int](0),
		TotalSecrets:           types.PointerTo[int](6),
		TotalVulnerabilities: &types.VulnerabilityScanSummary{
			TotalCriticalVulnerabilities:   types.PointerTo[int](2),
			TotalHighVulnerabilities:       types.PointerTo[int](2),
			TotalLowVulnerabilities:        types.PointerTo[int](2),
			TotalMediumVulnerabilities:     types.PointerTo[int](0),
			TotalNegligibleVulnerabilities: types.PointerTo[int](0),
		},
	}

	// Create scan 2: Running
	scan2Start := time.Now().Add(-5 * time.Minute)
	scan2Assets := []string{*assets[2].Id}

	scan2Summary := &types.ScanSummary{
		JobsCompleted:          types.PointerTo[int](1),
		JobsLeftToRun:          types.PointerTo[int](1),
		TotalExploits:          types.PointerTo[int](2),
		TotalInfoFinder:        types.PointerTo[int](2),
		TotalMalware:           types.PointerTo[int](3),
		TotalMisconfigurations: types.PointerTo[int](3),
		TotalPackages:          types.PointerTo[int](0),
		TotalRootkits:          types.PointerTo[int](3),
		TotalSecrets:           types.PointerTo[int](0),
		TotalVulnerabilities: &types.VulnerabilityScanSummary{
			TotalCriticalVulnerabilities:   types.PointerTo[int](0),
			TotalHighVulnerabilities:       types.PointerTo[int](0),
			TotalLowVulnerabilities:        types.PointerTo[int](0),
			TotalMediumVulnerabilities:     types.PointerTo[int](0),
			TotalNegligibleVulnerabilities: types.PointerTo[int](0),
		},
	}

	return []types.Scan{
		{
			EndTime: &scan1End,
			ScanConfig: &types.ScanConfigRelationship{
				Id: *scanConfigs[0].Id,
			},
			Scope:             scanConfigs[0].ScanTemplate.Scope,
			AssetScanTemplate: scanConfigs[0].ScanTemplate.AssetScanTemplate,
			StartTime:         &scan1Start,
			Status:            types.NewScanStatus(types.ScanStatusStateDone, types.ScanStatusReasonSuccess, types.PointerTo("Scan was completed successfully")),
			Summary:           scan1Summary,
			AssetIDs:          &scan1Assets,
		},
		{
			ScanConfig: &types.ScanConfigRelationship{
				Id: *scanConfigs[1].Id,
			},
			Scope:             scanConfigs[1].ScanTemplate.Scope,
			AssetScanTemplate: scanConfigs[1].ScanTemplate.AssetScanTemplate,
			StartTime:         &scan2Start,
			Status:            types.NewScanStatus(types.ScanStatusStateInProgress, types.ScanStatusReasonAssetScansRunning, types.PointerTo("Scan is in progress")),
			Summary:           scan2Summary,
			AssetIDs:          &scan2Assets,
		},
	}
}

// nolint:gocognit,maintidx,cyclop,gomnd
func createAssetScans(scans []types.Scan) []types.AssetScan {
	timeNow := time.Now()

	var assetScans []types.AssetScan
	for _, scan := range scans {
		for _, assetID := range *scan.AssetIDs {
			result := types.AssetScan{
				Id: nil,
				Scan: &types.ScanRelationship{
					Id: *scan.Id,
				},
				Secrets: nil,
				Status: types.NewAssetScanStatus(
					types.AssetScanStatusStateInProgress,
					types.AssetScanStatusReasonScannerIsRunning,
					nil,
				),
				Summary: &types.ScanFindingsSummary{},
				Stats: &types.AssetScanStats{
					General: &types.AssetScanGeneralStats{
						ScanTime: &types.AssetScanScanTime{
							EndTime:   &timeNow,
							StartTime: types.PointerTo(timeNow.Add(-10 * time.Second)),
						},
					},
				},
				Asset: &types.AssetRelationship{
					Id: assetID,
				},
				ScanFamiliesConfig:            scan.AssetScanTemplate.ScanFamiliesConfig,
				ScannerInstanceCreationConfig: scan.AssetScanTemplate.ScannerInstanceCreationConfig,
			}
			// Create Exploits if needed
			if *result.ScanFamiliesConfig.Exploits.Enabled {
				result.Exploits = &types.ExploitScan{
					Exploits: createExploitsResult(),
					Status:   types.NewScannerStatus(types.ScannerStatusStateDone, types.ScannerStatusReasonSuccess, nil),
				}
				result.Stats.Exploits = &[]types.AssetScanInputScanStats{
					{
						Path: types.PointerTo("/mnt"),
						ScanTime: &types.AssetScanScanTime{
							EndTime:   &timeNow,
							StartTime: types.PointerTo(timeNow.Add(-5 * time.Second)),
						},
						Size: types.PointerTo(int64(300)),
						Type: types.PointerTo("rootfs"),
					},
					{
						Path: types.PointerTo("/data"),
						ScanTime: &types.AssetScanScanTime{
							EndTime:   &timeNow,
							StartTime: types.PointerTo(timeNow.Add(-10 * time.Second)),
						},
						Size: types.PointerTo(int64(30)),
						Type: types.PointerTo("dir"),
					},
				}
				result.Summary.TotalExploits = types.PointerTo(len(*result.Exploits.Exploits))
			} else {
				result.Exploits = &types.ExploitScan{
					Exploits: nil,
					Status:   types.NewScannerStatus(types.ScannerStatusStateSkipped, types.ScannerStatusReasonNotScheduled, nil),
				}
				result.Summary.TotalExploits = types.PointerTo(0)
			}

			// Create Malware if needed
			if *result.ScanFamiliesConfig.Malware.Enabled {
				result.Malware = &types.MalwareScan{
					Malware:  createMalwareResult(),
					Metadata: nil,
					Status: types.NewScannerStatus(
						types.ScannerStatusStateFailed,
						types.ScannerStatusReasonError,
						types.PointerTo("failed to scan malware"),
					),
				}
				result.Stats.Malware = &[]types.AssetScanInputScanStats{
					{
						Path: types.PointerTo("/mnt"),
						ScanTime: &types.AssetScanScanTime{
							EndTime:   &timeNow,
							StartTime: types.PointerTo(timeNow.Add(-5 * time.Second)),
						},
						Size: types.PointerTo(int64(300)),
						Type: types.PointerTo("rootfs"),
					},
					{
						Path: types.PointerTo("/data"),
						ScanTime: &types.AssetScanScanTime{
							EndTime:   &timeNow,
							StartTime: types.PointerTo(timeNow.Add(-10 * time.Second)),
						},
						Size: types.PointerTo(int64(30)),
						Type: types.PointerTo("dir"),
					},
				}
				result.Summary.TotalMalware = types.PointerTo(len(*result.Malware.Malware))
			} else {
				result.Malware = &types.MalwareScan{
					Malware:  nil,
					Metadata: nil,
					Status:   types.NewScannerStatus(types.ScannerStatusStateSkipped, types.ScannerStatusReasonNotScheduled, nil),
				}
				result.Summary.TotalMalware = types.PointerTo(0)
			}

			// Create Misconfigurations if needed
			if *result.ScanFamiliesConfig.Misconfigurations.Enabled {
				result.Misconfigurations = &types.MisconfigurationScan{
					Misconfigurations: createMisconfigurationsResult(),
					Scanners:          nil,
					Status:            types.NewScannerStatus(types.ScannerStatusStateInProgress, types.ScannerStatusReasonScanning, nil),
				}
				result.Stats.Misconfigurations = &[]types.AssetScanInputScanStats{
					{
						Path: types.PointerTo("/mnt"),
						ScanTime: &types.AssetScanScanTime{
							EndTime:   &timeNow,
							StartTime: types.PointerTo(timeNow.Add(-5 * time.Second)),
						},
						Size: types.PointerTo(int64(300)),
						Type: types.PointerTo("rootfs"),
					},
				}
				result.Summary.TotalMisconfigurations = types.PointerTo(len(*result.Misconfigurations.Misconfigurations))
			} else {
				result.Misconfigurations = &types.MisconfigurationScan{
					Misconfigurations: nil,
					Scanners:          nil,
					Status:            types.NewScannerStatus(types.ScannerStatusStateSkipped, types.ScannerStatusReasonNotScheduled, nil),
				}
				result.Summary.TotalMisconfigurations = types.PointerTo(0)
			}

			// Create Packages if needed
			if *result.ScanFamiliesConfig.Sbom.Enabled {
				result.Sbom = &types.SbomScan{
					Packages: createPackagesResult(),
					Status:   types.NewScannerStatus(types.ScannerStatusStatePending, types.ScannerStatusReasonScheduled, nil),
				}
				result.Stats.Sbom = &[]types.AssetScanInputScanStats{
					{
						Path: types.PointerTo("/mnt"),
						ScanTime: &types.AssetScanScanTime{
							EndTime:   &timeNow,
							StartTime: types.PointerTo(timeNow.Add(-5 * time.Second)),
						},
						Size: types.PointerTo(int64(300)),
						Type: types.PointerTo("rootfs"),
					},
				}
				result.Summary.TotalPackages = types.PointerTo(len(*result.Sbom.Packages))
			} else {
				result.Sbom = &types.SbomScan{
					Packages: nil,
					Status:   types.NewScannerStatus(types.ScannerStatusStateSkipped, types.ScannerStatusReasonNotScheduled, nil),
				}
				result.Summary.TotalPackages = types.PointerTo(0)
			}

			// Create Rootkits if needed
			if *result.ScanFamiliesConfig.Rootkits.Enabled {
				result.Rootkits = &types.RootkitScan{
					Rootkits: createRootkitsResult(),
					Status:   types.NewScannerStatus(types.ScannerStatusStateDone, types.ScannerStatusReasonSuccess, nil),
				}
				result.Stats.Rootkits = &[]types.AssetScanInputScanStats{
					{
						Path: types.PointerTo("/mnt"),
						ScanTime: &types.AssetScanScanTime{
							EndTime:   &timeNow,
							StartTime: types.PointerTo(timeNow.Add(-5 * time.Second)),
						},
						Size: types.PointerTo(int64(300)),
						Type: types.PointerTo("rootfs"),
					},
				}
				result.Summary.TotalRootkits = types.PointerTo(len(*result.Rootkits.Rootkits))
			} else {
				result.Rootkits = &types.RootkitScan{
					Rootkits: nil,
					Status:   types.NewScannerStatus(types.ScannerStatusStateSkipped, types.ScannerStatusReasonNotScheduled, nil),
				}
				result.Summary.TotalRootkits = types.PointerTo(0)
			}

			// Create Secrets if needed
			if *result.ScanFamiliesConfig.Secrets.Enabled {
				result.Secrets = &types.SecretScan{
					Secrets: createSecretsResult(),
					Status:  types.NewScannerStatus(types.ScannerStatusStateDone, types.ScannerStatusReasonSuccess, nil),
				}
				result.Stats.Secrets = &[]types.AssetScanInputScanStats{
					{
						Path: types.PointerTo("/mnt"),
						ScanTime: &types.AssetScanScanTime{
							EndTime:   &timeNow,
							StartTime: types.PointerTo(timeNow.Add(-5 * time.Second)),
						},
						Size: types.PointerTo(int64(300)),
						Type: types.PointerTo("rootfs"),
					},
					{
						Path: types.PointerTo("/data"),
						ScanTime: &types.AssetScanScanTime{
							EndTime:   &timeNow,
							StartTime: types.PointerTo(timeNow.Add(-10 * time.Second)),
						},
						Size: types.PointerTo(int64(30)),
						Type: types.PointerTo("dir"),
					},
				}
				result.Summary.TotalSecrets = types.PointerTo(len(*result.Secrets.Secrets))
			} else {
				result.Secrets = &types.SecretScan{
					Secrets: nil,
					Status:  types.NewScannerStatus(types.ScannerStatusStateSkipped, types.ScannerStatusReasonNotScheduled, nil),
				}
				result.Summary.TotalSecrets = types.PointerTo(0)
			}

			// Create Vulnerabilities if needed
			if *result.ScanFamiliesConfig.Vulnerabilities.Enabled {
				result.Vulnerabilities = &types.VulnerabilityScan{
					Vulnerabilities: createVulnerabilitiesResult(),
					Status: types.NewScannerStatus(
						types.ScannerStatusStateDone,
						types.ScannerStatusReasonSuccess,
						nil,
					),
				}
				result.Stats.Vulnerabilities = &[]types.AssetScanInputScanStats{
					{
						Path: types.PointerTo("/mnt"),
						ScanTime: &types.AssetScanScanTime{
							EndTime:   &timeNow,
							StartTime: types.PointerTo(timeNow.Add(-5 * time.Second)),
						},
						Size: types.PointerTo(int64(300)),
						Type: types.PointerTo("rootfs"),
					},
				}
				result.Summary.TotalVulnerabilities = utils.GetVulnerabilityTotalsPerSeverity(result.Vulnerabilities.Vulnerabilities)
			} else {
				result.Vulnerabilities = &types.VulnerabilityScan{
					Vulnerabilities: nil,
					Status:          types.NewScannerStatus(types.ScannerStatusStateSkipped, types.ScannerStatusReasonNotScheduled, nil),
				}
				result.Summary.TotalVulnerabilities = utils.GetVulnerabilityTotalsPerSeverity(nil)
			}

			// Create InfoFinder if needed
			if *result.ScanFamiliesConfig.InfoFinder.Enabled {
				result.InfoFinder = &types.InfoFinderScan{
					Infos:    creatInfoFinderInfos(),
					Scanners: nil,
					Status:   types.NewScannerStatus(types.ScannerStatusStateInProgress, types.ScannerStatusReasonScanning, nil),
				}
				result.Stats.InfoFinder = &[]types.AssetScanInputScanStats{
					{
						Path: types.PointerTo("/mnt"),
						ScanTime: &types.AssetScanScanTime{
							EndTime:   &timeNow,
							StartTime: types.PointerTo(timeNow.Add(-5 * time.Second)),
						},
						Size: types.PointerTo(int64(300)),
						Type: types.PointerTo("rootfs"),
					},
				}
				result.Summary.TotalInfoFinder = types.PointerTo(len(*result.InfoFinder.Infos))
			} else {
				result.InfoFinder = &types.InfoFinderScan{
					Infos:    nil,
					Scanners: nil,
					Status:   types.NewScannerStatus(types.ScannerStatusStateSkipped, types.ScannerStatusReasonNotScheduled, nil),
				}
				result.Summary.TotalInfoFinder = types.PointerTo(0)
			}

			assetScans = append(assetScans, result)
		}
	}
	return assetScans
}

func creatInfoFinderInfos() *[]types.InfoFinderInfo {
	return &[]types.InfoFinderInfo{
		{
			Data:        types.PointerTo("2048 SHA256:YQuPOM8ld6FOA9HbKCgkCJWHuGt4aTRD7hstjJpRhxc xxxx (RSA)"),
			Path:        types.PointerTo("/home/ec2-user/.ssh/authorized_keys"),
			ScannerName: types.PointerTo("sshTopology"),
			Type:        types.PointerTo(types.InfoTypeSSHAuthorizedKeyFingerprint),
		},
		{
			Data:        types.PointerTo("256 SHA256:gv6snCwAl5+6fY2g5VkmETWb9Mv0zLRkMz8aQyQWAVc xxxx (ED25519)"),
			Path:        types.PointerTo("/etc/ssh/ssh_host_ed25519_key"),
			ScannerName: types.PointerTo("sshTopology"),
			Type:        types.PointerTo(types.InfoTypeSSHDaemonKeyFingerprint),
		},
	}
}

// nolint:gomnd
func createSecretsResult() *[]types.Secret {
	return &[]types.Secret{
		{
			Description: types.PointerTo("AWS Credentials"),
			EndColumn:   types.PointerTo(8),
			EndLine:     types.PointerTo(43),
			FilePath:    types.PointerTo("/.aws/credentials"),
			Fingerprint: types.PointerTo("credentials:aws-access-token:4"),
			StartColumn: types.PointerTo(7),
			StartLine:   types.PointerTo(43),
		},
		{
			Description: types.PointerTo("export BUNDLE_ENTERPRISE__CONTRIBSYS__COM=cafebabe:deadbeef"),
			EndColumn:   types.PointerTo(10),
			EndLine:     types.PointerTo(26),
			FilePath:    types.PointerTo("cmd/generate/config/rules/sidekiq.go"),
			Fingerprint: types.PointerTo("cd5226711335c68be1e720b318b7bc3135a30eb2:cmd/generate/config/rules/sidekiq.go:sidekiq-secret:23"),
			StartColumn: types.PointerTo(7),
			StartLine:   types.PointerTo(23),
		},
		{
			Description: types.PointerTo("GitLab Personal Access Token"),
			EndColumn:   types.PointerTo(22),
			EndLine:     types.PointerTo(7),
			FilePath:    types.PointerTo("Applications/Firefox.app/Contents/Resources/browser/omni.ja"),
			Fingerprint: types.PointerTo("Applications/Firefox.app/Contents/Resources/browser/omni.ja:generic-api-key:sfs2"),
			StartColumn: types.PointerTo(20),
			StartLine:   types.PointerTo(7),
		},
	}
}

func createRootkitsResult() *[]types.Rootkit {
	return &[]types.Rootkit{
		{
			Message:     types.PointerTo("/usr/lwp-request"),
			RootkitName: types.PointerTo("Ambient's Rootkit (ARK)"),
			RootkitType: types.PointerTo(types.RootkitType("ARK")),
		},
		{
			Message:     types.PointerTo("Possible Linux/Ebury 1.4 - Operation Windigo installed"),
			RootkitName: types.PointerTo("Linux/Ebury - Operation Windigo ssh"),
			RootkitType: types.PointerTo(types.RootkitType("Malware")),
		},
		{
			Message:     types.PointerTo("/var/adm/wtmpx"),
			RootkitName: types.PointerTo("Mumblehard backdoor/botnet"),
			RootkitType: types.PointerTo(types.RootkitType("Botnet")),
		},
	}
}

func createPackagesResult() *[]types.Package {
	return &[]types.Package{
		{
			Cpes:     types.PointerTo([]string{"cpe:2.3:a:curl:curl:7.74.0-1.3+deb11u3:*:*:*:*:*:*:*"}),
			Language: types.PointerTo(""),
			Licenses: types.PointerTo([]string{"BSD-3-Clause", "BSD-4-Clause"}),
			Name:     types.PointerTo("curl"),
			Purl:     types.PointerTo("pkg:deb/debian/curl@7.74.0-1.3+deb11u3?arch=amd64&distro=debian-11"),
			Type:     types.PointerTo("deb"),
			Version:  types.PointerTo("7.74.0-1.3+deb11u3"),
		},
		{
			Cpes:     types.PointerTo([]string{"cpe:2.3:a:libtasn1-6:libtasn1-6:4.16.0-2:*:*:*:*:*:*:*", "cpe:2.3:a:libtasn1-6:libtasn1_6:4.16.0-2:*:*:*:*:*:*:*"}),
			Language: types.PointerTo("python"),
			Licenses: types.PointerTo([]string{"GFDL-1.3-only", "GPL-3.0-only", "LGPL-2.1-only"}),
			Name:     types.PointerTo("libtasn1-6"),
			Purl:     types.PointerTo("pkg:deb/debian/libtasn1-6@4.16.0-2?arch=amd64&distro=debian-11"),
			Type:     types.PointerTo("deb"),
			Version:  types.PointerTo("4.16.0-2"),
		},
	}
}

func createExploitsResult() *[]types.Exploit {
	return &[]types.Exploit{
		{
			CveID:       types.PointerTo("CVE-2009-4091"),
			Description: types.PointerTo("Simplog 0.9.3.2 - Multiple Vulnerabilities"),
			Name:        types.PointerTo("10180"),
			SourceDB:    types.PointerTo("OffensiveSecurity"),
			Title:       types.PointerTo("10180"),
			Urls:        types.PointerTo([]string{"https://www.exploit-db.com/exploits/10180"}),
		},
		{
			CveID:       types.PointerTo("CVE-2006-2896"),
			Description: types.PointerTo("FunkBoard CF0.71 - 'profile.php' Remote User Pass Change"),
			Name:        types.PointerTo("1875"),
			SourceDB:    types.PointerTo("OffensiveSecurity"),
			Title:       types.PointerTo("1875"),
			Urls:        types.PointerTo([]string{"https://gitlab.com/exploit-database/exploitdb/-/tree/main/exploits/php/webapps/1875.html"}),
		},
	}
}

func createMalwareResult() *[]types.Malware {
	return &[]types.Malware{
		{
			MalwareName: types.PointerTo("Pdf.Exploit.CVE_2009_4324-1"),
			MalwareType: types.PointerTo("WORM"),
			Path:        types.PointerTo("/test/metasploit-framework/modules/exploits/windows/browser/asus_net4switch_ipswcom.rb"),
		},
		{
			MalwareName: types.PointerTo("Xml.Malware.Squiblydoo-6728833-0"),
			MalwareType: types.PointerTo("SPYWARE"),
			Path:        types.PointerTo("/test/metasploit-framework/modules/exploits/windows/fileformat/office_ms17_11882.rb"),
		},
		{
			MalwareName: types.PointerTo("Unix.Trojan.MSShellcode-27"),
			MalwareType: types.PointerTo("TROJAN"),
			Path:        types.PointerTo("/test/metasploit-framework/documentation/modules/exploit/multi/http/makoserver_cmd_exec.md"),
		},
	}
}

func createMisconfigurationsResult() *[]types.Misconfiguration {
	return &[]types.Misconfiguration{
		{
			Message:         types.PointerTo("Install a PAM module for password strength testing like pam_cracklib or pam_passwdqc. Details: /lib/x86_64-linux-gnu/security/pam_access.so"),
			Remediation:     types.PointerTo("remediation2"),
			ScannedPath:     types.PointerTo("/home/ubuntu/debian11"),
			ScannerName:     types.PointerTo("scanner2"),
			Severity:        types.PointerTo(types.MisconfigurationHighSeverity),
			TestCategory:    types.PointerTo("AUTH"),
			TestDescription: types.PointerTo("Checking presence password strength testing tools (PAM)"),
			TestID:          types.PointerTo("AUTH-9262"),
		},
		{
			Message:         types.PointerTo("Set the sticky bit on /home/ubuntu/debian11/tmp, to prevent users deleting (by other owned) files in the /tmp directory. Details: /tmp"),
			Remediation:     types.PointerTo("remediation1"),
			ScannedPath:     types.PointerTo("/home/ubuntu/debian11"),
			ScannerName:     types.PointerTo("scanner1"),
			Severity:        types.PointerTo(types.MisconfigurationMediumSeverity),
			TestCategory:    types.PointerTo("FILE"),
			TestDescription: types.PointerTo("Checking /tmp sticky bit"),
			TestID:          types.PointerTo("FILE-6362"),
		},
		{
			Message:         types.PointerTo("Disable drivers like USB storage when not used, to prevent unauthorized storage or data theft. Details: /etc/cron.d/e2scrub_all"),
			Remediation:     types.PointerTo("remediation1"),
			ScannedPath:     types.PointerTo("/home/ubuntu/debian11"),
			ScannerName:     types.PointerTo("scanner1"),
			Severity:        types.PointerTo(types.MisconfigurationLowSeverity),
			TestCategory:    types.PointerTo("USB"),
			TestDescription: types.PointerTo("Check if USB storage is disabled"),
			TestID:          types.PointerTo("USB-1000"),
		},
	}
}

// nolint:gomnd
func createVulnerabilitiesResult() *[]types.Vulnerability {
	return &[]types.Vulnerability{
		{
			Cvss: types.PointerTo([]types.VulnerabilityCvss{
				{
					Metrics: &types.VulnerabilityCvssMetrics{
						BaseScore:           types.PointerTo[float32](7.5),
						ExploitabilityScore: types.PointerTo[float32](3.9),
						ImpactScore:         types.PointerTo[float32](3.6),
					},
					Vector:  types.PointerTo("CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:N"),
					Version: types.PointerTo("3.1"),
				},
			}),
			Description: types.PointerTo("A vulnerability exists in curl <7.87.0 HSTS check that could be bypassed to trick it to keep using HTTP. Using its HSTS support, curl can be instructed to use HTTPS instead of using an insecure clear-text HTTP step even when HTTP is provided in the\nURL. However, the HSTS mechanism could be bypassed if the host name in the given URL first uses IDN characters that get replaced to ASCII counterparts as part of the IDN conversion. Like using the character UTF-8 U+3002 (IDEOGRAPHIC FULL STOP) instead of the common ASCI\nI full stop (U+002E) `.`. Then in a subsequent request, it does not detect the HSTS state and makes a clear text transfer. Because it would store the info IDN encoded but look for it IDN decoded."),
			Distro: &types.VulnerabilityDistro{
				IDLike:  types.PointerTo([]string{"debian"}),
				Name:    types.PointerTo("ubuntu"),
				Version: types.PointerTo("11"),
			},
			Fix: &types.VulnerabilityFix{
				State:    types.PointerTo("wont-fix"),
				Versions: types.PointerTo([]string{}),
			},
			LayerId: types.PointerTo(""),
			Links:   types.PointerTo([]string{"https://security-tracker.debian.org/tracker/CVE-2022-43551"}),
			Package: &types.Package{
				Cpes:     types.PointerTo([]string{"cpe:2.3:a:curl:curl:7.74.0-1.3+deb11u3:*:*:*:*:*:*:*"}),
				Language: types.PointerTo("pl1"),
				Licenses: types.PointerTo([]string{"BSD-3-Clause", "BSD-4-Clause"}),
				Name:     types.PointerTo("curl"),
				Purl:     types.PointerTo("pkg:deb/debian/curl@7.74.0-1.3+deb11u3?arch=amd64&distro=debian-11"),
				Type:     types.PointerTo("deb"),
				Version:  types.PointerTo("7.74.0-1.3+deb11u3"),
			},
			Path:              types.PointerTo("/var/lib/dpkg/status"),
			Severity:          types.PointerTo[types.VulnerabilitySeverity](types.HIGH),
			VulnerabilityName: types.PointerTo("CVE-2022-43551"),
		},
		{
			Cvss: types.PointerTo([]types.VulnerabilityCvss{
				{
					Metrics: &types.VulnerabilityCvssMetrics{
						BaseScore:           types.PointerTo[float32](9.1),
						ExploitabilityScore: types.PointerTo[float32](3.9),
						ImpactScore:         types.PointerTo[float32](5.2),
					},
					Vector:  types.PointerTo("CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:H"),
					Version: types.PointerTo("3.1"),
				},
				{
					Metrics: &types.VulnerabilityCvssMetrics{
						BaseScore:           types.PointerTo[float32](4),
						ExploitabilityScore: types.PointerTo[float32](4.1),
						ImpactScore:         types.PointerTo[float32](4.2),
					},
					Vector:  types.PointerTo("AV:N/AC:L/Au:N/C:P/I:P/A:P"),
					Version: types.PointerTo("2.0"),
				},
			}),
			Description: types.PointerTo("GNU Libtasn1 before 4.19.0 has an ETYPE_OK off-by-one array size check that affects asn1_encode_simple_der."),
			Distro: &types.VulnerabilityDistro{
				IDLike:  types.PointerTo([]string{"debian"}),
				Name:    types.PointerTo("ubuntu"),
				Version: types.PointerTo("11"),
			},
			Fix: &types.VulnerabilityFix{
				State:    types.PointerTo("fixed"),
				Versions: types.PointerTo([]string{"4.16.0-2+deb11u1"}),
			},
			LayerId: types.PointerTo(""),
			Links:   types.PointerTo([]string{"https://security-tracker.debian.org/tracker/CVE-2021-46848", "https://security-tracker.debian.org/tracker/CVE-2021-46848_new"}),
			Package: &types.Package{
				Cpes:     types.PointerTo([]string{"cpe:2.3:a:libtasn1-6:libtasn1-6:4.16.0-2:*:*:*:*:*:*:*", "cpe:2.3:a:libtasn1-6:libtasn1_6:4.16.0-2:*:*:*:*:*:*:*"}),
				Language: types.PointerTo(""),
				Licenses: types.PointerTo([]string{"GFDL-1.3-only", "GPL-3.0-only", "LGPL-2.1-only"}),
				Name:     types.PointerTo("libtasn1-6"),
				Purl:     types.PointerTo("pkg:deb/debian/libtasn1-6@4.16.0-2?arch=amd64&distro=debian-11"),
				Type:     types.PointerTo("deb"),
				Version:  types.PointerTo("4.16.0-2"),
			},
			Path:              types.PointerTo("/var/lib/dpkg/status"),
			Severity:          types.PointerTo[types.VulnerabilitySeverity](types.CRITICAL),
			VulnerabilityName: types.PointerTo("CVE-2021-46848"),
		},
		{
			Cvss: types.PointerTo([]types.VulnerabilityCvss{
				{
					Metrics: &types.VulnerabilityCvssMetrics{
						BaseScore:           types.PointerTo[float32](9.8),
						ExploitabilityScore: types.PointerTo[float32](3.9),
						ImpactScore:         types.PointerTo[float32](5.9),
					},
					Vector:  types.PointerTo("CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H"),
					Version: types.PointerTo("3.1"),
				},
				{
					Metrics: &types.VulnerabilityCvssMetrics{
						BaseScore:           types.PointerTo[float32](7.5),
						ExploitabilityScore: types.PointerTo[float32](10),
						ImpactScore:         types.PointerTo[float32](6.4),
					},
					Vector:  types.PointerTo("AV:N/AC:L/Au:N/C:P/I:P/A:P"),
					Version: types.PointerTo("2.0"),
				},
			}),
			Description: types.PointerTo("SQLite3 from 3.6.0 to and including 3.27.2 is vulnerable to heap out-of-bound read in the rtreenode() function when handling invalid rtree tables."),
			Distro: &types.VulnerabilityDistro{
				IDLike:  types.PointerTo([]string{"debian"}),
				Name:    types.PointerTo("ubuntu"),
				Version: types.PointerTo("11"),
			},
			Fix: &types.VulnerabilityFix{
				State:    types.PointerTo("wont-fix"),
				Versions: types.PointerTo([]string{""}),
			},
			LayerId: types.PointerTo(""),
			Links:   types.PointerTo([]string{"https://security-tracker.debian.org/tracker/CVE-2019-8457"}),
			Package: &types.Package{
				Cpes:     types.PointerTo([]string{"cpe:2.3:a:libdb5.3:libdb5.3:5.3.28+dfsg1-0.8:*:*:*:*:*:*:*"}),
				Language: types.PointerTo(""),
				Licenses: types.PointerTo([]string{}),
				Name:     types.PointerTo("libdb5.3"),
				Purl:     types.PointerTo("pkg:deb/debian/libdb5.3@5.3.28+dfsg1-0.8?arch=amd64&upstream=db5.3&distro=debian-11"),
				Type:     types.PointerTo("deb"),
				Version:  types.PointerTo("5.3.28+dfsg1-0.8"),
			},
			Path:              types.PointerTo("/var/lib/dpkg/status"),
			Severity:          types.PointerTo[types.VulnerabilitySeverity](types.LOW),
			VulnerabilityName: types.PointerTo("CVE-2019-8457"),
		},
	}
}
