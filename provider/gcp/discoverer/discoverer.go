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

package discoverer

import (
	"context"
	"errors"
	"fmt"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/provider/gcp/utils"
)

const (
	maxResults = 500
)

type Discoverer struct {
	DisksClient     *compute.DisksClient
	InstancesClient *compute.InstancesClient
	RegionsClient   *compute.RegionsClient

	ProjectID string
}

// nolint: cyclop
func (d *Discoverer) DiscoverAssets(ctx context.Context) provider.AssetDiscoverer {
	assetDiscoverer := provider.NewSimpleAssetDiscoverer()

	go func() {
		defer close(assetDiscoverer.OutputChan)

		regions, err := d.listAllRegions(ctx)
		if err != nil {
			assetDiscoverer.Error = fmt.Errorf("failed to list all regions: %w", err)
			return
		}

		var zones []string
		for _, region := range regions {
			zones = append(zones, getZonesLastPart(region.Zones)...)
		}

		for _, zone := range zones {
			assets, err := d.listInstances(ctx, nil, zone)
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

func (d *Discoverer) listInstances(ctx context.Context, filter *string, zone string) ([]apitypes.AssetType, error) {
	var ret []apitypes.AssetType

	it := d.InstancesClient.List(ctx, &computepb.ListInstancesRequest{
		Filter:     filter,
		MaxResults: to.Ptr[uint32](maxResults),
		Project:    d.ProjectID,
		Zone:       zone,
	})
	for {
		resp, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			_, err = utils.HandleGcpRequestError(err, "listing instances for project %s zone %s", d.ProjectID, zone)
			return nil, err // nolint: wrapcheck
		}

		info, err := d.getVMInfoFromVirtualMachine(ctx, resp)
		if err != nil {
			return nil, fmt.Errorf("failed to get vminfo from virtual machine: %w", err)
		}
		ret = append(ret, info)
	}

	return ret, nil
}

func (d *Discoverer) listAllRegions(ctx context.Context) ([]*computepb.Region, error) {
	var ret []*computepb.Region

	it := d.RegionsClient.List(ctx, &computepb.ListRegionsRequest{
		MaxResults: to.Ptr[uint32](maxResults),
		Project:    d.ProjectID,
	})
	for {
		resp, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			_, err := utils.HandleGcpRequestError(err, "list regions")
			return nil, err // nolint: wrapcheck
		}

		ret = append(ret, resp)
	}
	return ret, nil
}

func (d *Discoverer) getVMInfoFromVirtualMachine(ctx context.Context, vm *computepb.Instance) (apitypes.AssetType, error) {
	assetType := apitypes.AssetType{}
	launchTime, err := time.Parse(time.RFC3339, *vm.CreationTimestamp)
	if err != nil {
		return apitypes.AssetType{}, fmt.Errorf("failed to parse time: %v", *vm.CreationTimestamp)
	}
	// get boot disk name
	diskName := utils.GetLastURLPart(vm.Disks[0].Source)

	var platform string
	var image string

	// get disk from gcp
	disk, err := d.DisksClient.Get(ctx, &computepb.GetDiskRequest{
		Disk:    diskName,
		Project: d.ProjectID,
		Zone:    utils.GetLastURLPart(vm.Zone),
	})
	if err != nil {
		logrus.Warnf("failed to get disk %v: %v", diskName, err)
	} else {
		if disk.Architecture != nil {
			platform = *disk.Architecture
		}
		image = utils.GetLastURLPart(disk.SourceImage)
	}

	err = assetType.FromVMInfo(apitypes.VMInfo{
		InstanceProvider: to.Ptr(apitypes.GCP),
		InstanceID:       *vm.Name,
		Image:            image,
		InstanceType:     utils.GetLastURLPart(vm.MachineType),
		LaunchTime:       launchTime,
		Location:         utils.GetLastURLPart(vm.Zone),
		Platform:         platform,
		SecurityGroups:   &[]apitypes.SecurityGroup{},
		Tags:             to.Ptr(convertLabelsToTags(vm.Labels)),
	})
	if err != nil {
		return apitypes.AssetType{}, provider.FatalErrorf("failed to create AssetType from VMInfo: %w", err)
	}

	return assetType, nil
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
		ret = append(ret, utils.GetLastURLPart(&z))
	}
	return ret
}

func convertLabelsToTags(labels map[string]string) []apitypes.Tag {
	tags := make([]apitypes.Tag, 0, len(labels))

	for k, v := range labels {
		tags = append(
			tags,
			apitypes.Tag{
				Key:   k,
				Value: v,
			},
		)
	}

	return tags
}
