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

package gcp

import (
	"context"
	"errors"
	"fmt"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/orchestrator/provider"
	"github.com/openclarity/vmclarity/pkg/shared/log"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

type Provider struct {
	snapshotsClient *compute.SnapshotsClient
	disksClient     *compute.DisksClient
	instancesClient *compute.InstancesClient
	regionsClient   *compute.RegionsClient

	config *Config
}

func New(ctx context.Context) (*Provider, error) {
	config, err := NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	err = config.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate configuration: %w", err)
	}

	regionsClient, err := compute.NewRegionsRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create regions client: %w", err)
	}

	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance client: %w", err)
	}

	snapshotsClient, err := compute.NewSnapshotsRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot client: %w", err)
	}

	disksClient, err := compute.NewDisksRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create disks client: %w", err)
	}

	return &Provider{
		snapshotsClient: snapshotsClient,
		disksClient:     disksClient,
		instancesClient: instancesClient,
		regionsClient:   regionsClient,
		config:          config,
	}, nil
}

func (p *Provider) Kind() models.CloudProvider {
	return models.GCP
}

func (p *Provider) Estimate(ctx context.Context, stats models.AssetScanStats, asset *models.Asset, assetScanTemplate *models.AssetScanTemplate) (*models.Estimation, error) {
	return &models.Estimation{}, provider.FatalErrorf("Not Implemented")
}

// nolint:cyclop
func (p *Provider) RunAssetScan(ctx context.Context, config *provider.ScanJobConfig) error {
	// convert AssetInfo to vmInfo
	vminfo, err := config.AssetInfo.AsVMInfo()
	if err != nil {
		return provider.FatalErrorf("unable to get vminfo from AssetInfo: %w", err)
	}

	logger := log.GetLoggerFromContextOrDefault(ctx).WithFields(logrus.Fields{
		"AssetScanID":   config.AssetScanID,
		"AssetLocation": vminfo.Location,
		"InstanceID":    vminfo.InstanceID,
		"ScannerZone":   p.config.ScannerZone,
		"Provider":      string(p.Kind()),
	})
	logger.Debugf("Running asset scan")

	targetName := vminfo.InstanceID
	targetZone := vminfo.Location

	// get the target instance to scan from gcp.
	targetVM, err := p.instancesClient.Get(ctx, &computepb.GetInstanceRequest{
		Instance: targetName,
		Project:  p.config.ProjectID,
		Zone:     targetZone,
	})
	if err != nil {
		_, err := handleGcpRequestError(err, "getting target virtual machine %v", targetName)
		return err
	}
	logger.Debugf("Got target VM: %v", targetVM.Name)

	// get target instance boot disk
	bootDisk, err := getInstanceBootDisk(targetVM)
	if err != nil {
		return provider.FatalErrorf("unable to get instance boot disk: %w", err)
	}
	logger.Debugf("Got target boot disk: %v", bootDisk.GetSource())

	// ensure that a snapshot was created from the target instance root disk. (create if not)
	snapshot, err := p.ensureSnapshotFromAttachedDisk(ctx, config, bootDisk)
	if err != nil {
		return fmt.Errorf("failed to ensure snapshot for vm root volume: %w", err)
	}
	logger.Debugf("Created snapshot: %v", snapshot.Name)

	// create a disk from the snapshot.
	// Snapshots are global resources, so any snapshot is accessible by any resource within the same project.
	var diskFromSnapshot *computepb.Disk
	diskFromSnapshot, err = p.ensureDiskFromSnapshot(ctx, config, snapshot)
	if err != nil {
		return fmt.Errorf("failed to ensure disk created from snapshot: %w", err)
	}
	logger.Debugf("Created disk from snapshot: %v", diskFromSnapshot.Name)

	// create the scanner instance
	scannerVM, err := p.ensureScannerVirtualMachine(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure scanner virtual machine: %w", err)
	}
	logger.Debugf("Created scanner virtual machine: %v", scannerVM.Name)

	// attach the disk from snapshot to the scanner instance
	err = p.ensureDiskAttachedToScannerVM(ctx, scannerVM, diskFromSnapshot)
	if err != nil {
		return fmt.Errorf("failed to ensure target disk is attached to virtual machine: %w", err)
	}
	logger.Debugf("Attached disk to scanner virtual machine")

	return nil
}

func (p *Provider) RemoveAssetScan(ctx context.Context, config *provider.ScanJobConfig) error {
	logger := log.GetLoggerFromContextOrDefault(ctx).WithFields(logrus.Fields{
		"AssetScanID": config.AssetScanID,
		"ScannerZone": p.config.ScannerZone,
		"Provider":    string(p.Kind()),
	})

	err := p.ensureScannerVirtualMachineDeleted(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure scanner virtual machine deleted: %w", err)
	}
	logger.Debugf("Deleted scanner virtual machine")

	err = p.ensureTargetDiskDeleted(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure target disk deleted: %w", err)
	}
	logger.Debugf("Deleted disk")

	err = p.ensureSnapshotDeleted(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to ensure snapshot deleted: %w", err)
	}
	logger.Debugf("Deleted snapshot")

	return nil
}

// nolint: cyclop
func (p *Provider) DiscoverAssets(ctx context.Context) provider.AssetDiscoverer {
	assetDiscoverer := provider.NewSimpleAssetDiscoverer()

	go func() {
		defer close(assetDiscoverer.OutputChan)

		regions, err := p.listAllRegions(ctx)
		if err != nil {
			assetDiscoverer.Error = fmt.Errorf("failed to list all regions: %w", err)
			return
		}

		var zones []string
		for _, region := range regions {
			zones = append(zones, getZonesLastPart(region.Zones)...)
		}

		for _, zone := range zones {
			assets, err := p.listInstances(ctx, nil, zone)
			if err != nil {
				assetDiscoverer.Error = fmt.Errorf("failed to list instances: %w", err)
				return
			}

			for _, asset := range assets {
				select {
				case assetDiscoverer.OutputChan <- asset:
				case <-ctx.Done():
					assetDiscoverer.Error = ctx.Err()
					return
				}
			}
		}
	}()

	return assetDiscoverer
}

// getZonesLastPart converts a list of zone URLs into a list of zone IDs.
// For example input:
//
// [
//
//	https://www.googleapis.com/compute/v1/projects/gcp-etigcp-nprd-12855/zones/us-central1-c,
//	https://www.googleapis.com/compute/v1/projects/gcp-etigcp-nprd-12855/zones/us-central1-a
//
// ]
//
// returns [us-central1-c, us-central1-a].
func getZonesLastPart(zones []string) []string {
	ret := make([]string, 0, len(zones))
	for _, zone := range zones {
		z := zone
		ret = append(ret, getLastURLPart(&z))
	}
	return ret
}

func getInstanceBootDisk(vm *computepb.Instance) (*computepb.AttachedDisk, error) {
	for _, disk := range vm.Disks {
		if disk.Boot != nil && *disk.Boot {
			return disk, nil
		}
	}
	return nil, fmt.Errorf("failed to find instance boot disk")
}

func (p *Provider) listInstances(ctx context.Context, filter *string, zone string) ([]models.AssetType, error) {
	var ret []models.AssetType

	it := p.instancesClient.List(ctx, &computepb.ListInstancesRequest{
		Filter:     filter,
		MaxResults: utils.PointerTo[uint32](maxResults),
		Project:    p.config.ProjectID,
		Zone:       zone,
	})
	for {
		resp, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			_, err = handleGcpRequestError(err, "listing instances for project %s zone %s", p.config.ProjectID, zone)
			return nil, err
		}

		info, err := p.getVMInfoFromVirtualMachine(ctx, resp)
		if err != nil {
			return nil, fmt.Errorf("failed to get vminfo from virtual machine: %w", err)
		}
		ret = append(ret, info)
	}

	return ret, nil
}

func (p *Provider) listAllRegions(ctx context.Context) ([]*computepb.Region, error) {
	var ret []*computepb.Region

	it := p.regionsClient.List(ctx, &computepb.ListRegionsRequest{
		MaxResults: utils.PointerTo[uint32](maxResults),
		Project:    p.config.ProjectID,
	})
	for {
		resp, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			_, err := handleGcpRequestError(err, "list regions")
			return nil, err
		}

		ret = append(ret, resp)
	}
	return ret, nil
}

func (p *Provider) getVMInfoFromVirtualMachine(ctx context.Context, vm *computepb.Instance) (models.AssetType, error) {
	assetType := models.AssetType{}
	launchTime, err := time.Parse(time.RFC3339, *vm.CreationTimestamp)
	if err != nil {
		return models.AssetType{}, fmt.Errorf("failed to parse time: %v", *vm.CreationTimestamp)
	}
	// get boot disk name
	diskName := getLastURLPart(vm.Disks[0].Source)

	var platform string
	var image string

	// get disk from gcp
	disk, err := p.disksClient.Get(ctx, &computepb.GetDiskRequest{
		Disk:    diskName,
		Project: p.config.ProjectID,
		Zone:    getLastURLPart(vm.Zone),
	})
	if err != nil {
		logrus.Warnf("failed to get disk %v: %v", diskName, err)
	} else {
		if disk.Architecture != nil {
			platform = *disk.Architecture
		}
		image = getLastURLPart(disk.SourceImage)
	}

	err = assetType.FromVMInfo(models.VMInfo{
		InstanceProvider: utils.PointerTo(models.GCP),
		InstanceID:       *vm.Name,
		Image:            image,
		InstanceType:     getLastURLPart(vm.MachineType),
		LaunchTime:       launchTime,
		Location:         getLastURLPart(vm.Zone),
		Platform:         platform,
		SecurityGroups:   &[]models.SecurityGroup{},
		Tags:             utils.PointerTo(convertLabelsToTags(vm.Labels)),
	})
	if err != nil {
		return models.AssetType{}, provider.FatalErrorf("failed to create AssetType from VMInfo: %w", err)
	}

	return assetType, nil
}

func convertLabelsToTags(labels map[string]string) []models.Tag {
	tags := make([]models.Tag, 0, len(labels))

	for k, v := range labels {
		tags = append(
			tags,
			models.Tag{
				Key:   k,
				Value: v,
			},
		)
	}

	return tags
}
