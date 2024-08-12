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

package aws

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"

	awstype "github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/orchestrator/provider"
	"github.com/openclarity/vmclarity/pkg/orchestrator/provider/aws/scanestimation"
	"github.com/openclarity/vmclarity/pkg/orchestrator/provider/cloudinit"
	"github.com/openclarity/vmclarity/pkg/shared/log"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

type Provider struct {
	ec2Client     *ec2.Client
	scanEstimator *scanestimation.ScanEstimator
	config        *Config
}

func New(ctx context.Context) (*Provider, error) {
	config, err := NewConfig()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration. Provider=AWS: %w", err)
	}

	if err = config.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate provider configuration. Provider=AWS: %w", err)
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	return &Provider{
		ec2Client:     ec2Client,
		scanEstimator: scanestimation.New(pricing.NewFromConfig(cfg), ec2Client),
		config:        config,
	}, nil
}

func (p *Provider) Kind() models.CloudProvider {
	return models.AWS
}

func (p *Provider) Estimate(ctx context.Context, assetScanStats models.AssetScanStats, asset *models.Asset, assetScanTemplate *models.AssetScanTemplate) (*models.Estimation, error) {
	var err error
	const jobCreationTimeConst = 2

	vminfo, err := asset.AssetInfo.AsVMInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to use asset info as vminfo: %w", err)
	}

	location, err := NewLocation(vminfo.Location)
	if err != nil {
		return nil, fmt.Errorf("failed to parse location %v: %w", vminfo.Location, err)
	}

	sourceRegion := location.Region
	destRegion := p.config.ScannerRegion
	scannerInstanceType := p.config.ScannerInstanceType

	scannerRootVolumeSizeGB := vminfo.RootVolume.SizeGB
	scannerVolumeType := ec2types.VolumeTypeGp2                          // TODO this should come from configuration once we support more than one volume type.
	fromSnapshotVolumeType := ec2types.VolumeTypeGp2                     // TODO this should come from configuration once we support more than one volume type.
	jobCreationTimeSec := jobCreationTimeConst * scannerRootVolumeSizeGB // TODO create a formula to calculate this per GB

	params := scanestimation.EstimateAssetScanParams{
		SourceRegion:            sourceRegion,
		DestRegion:              destRegion,
		ScannerVolumeType:       scannerVolumeType,
		FromSnapshotVolumeType:  fromSnapshotVolumeType,
		ScannerInstanceType:     ec2types.InstanceType(scannerInstanceType),
		JobCreationTimeSec:      int64(jobCreationTimeSec),
		ScannerRootVolumeSizeGB: int64(scannerRootVolumeSizeGB),
		Stats:                   assetScanStats,
		Asset:                   asset,
		AssetScanTemplate:       assetScanTemplate,
	}
	ret, err := p.scanEstimator.EstimateAssetScan(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate asset scan: %w", err)
	}

	return ret, nil
}

// nolint:cyclop
func (p *Provider) DiscoverAssets(ctx context.Context) provider.AssetDiscoverer {
	assetDiscoverer := provider.NewSimpleAssetDiscoverer()

	go func() {
		defer close(assetDiscoverer.OutputChan)

		regions, err := p.ListAllRegions(ctx)
		if err != nil {
			assetDiscoverer.Error = fmt.Errorf("failed to get regions: %w", err)
			return
		}

		logger := log.GetLoggerFromContextOrDiscard(ctx)

		for _, region := range regions {
			instances, err := p.GetInstances(ctx, []ec2types.Filter{}, region.Name)
			if err != nil {
				logger.Warnf("Failed to get instances. region=%v: %v", region, err)
				continue
			}

			for _, instance := range instances {
				asset, err := getVMInfoFromInstance(instance)
				if err != nil {
					assetDiscoverer.Error = FatalError{
						Err: fmt.Errorf("failed convert EC2 Instance to AssetType: %w", err),
					}
					return
				}
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

// nolint:nilnil
func (p *Provider) getInstanceWithID(ctx context.Context, id string, region string) (*ec2types.Instance, error) {
	out, err := p.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{id},
	}, func(options *ec2.Options) {
		options.Region = region
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch VM isntance. InstanceID=%s: %w", id, err)
	}

	for _, r := range out.Reservations {
		for _, i := range r.Instances {
			ec2Instance := i

			if ec2Instance.InstanceId == nil || *ec2Instance.InstanceId != id {
				continue
			}

			return &ec2Instance, nil
		}
	}

	return nil, nil
}

// nolint:cyclop
func (p *Provider) createInstance(ctx context.Context, region string, config *provider.ScanJobConfig) (*Instance, error) {
	options := func(options *ec2.Options) {
		options.Region = region
	}

	ec2TagsForInstance := EC2TagsFromScanMetadata(config.ScanMetadata)
	ec2Filters := EC2FiltersFromEC2Tags(ec2TagsForInstance)

	describeParams := &ec2.DescribeInstancesInput{
		Filters: ec2Filters,
	}
	describeOut, err := p.ec2Client.DescribeInstances(ctx, describeParams, options)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch scanner VM instances: %w", err)
	}

	for _, r := range describeOut.Reservations {
		for _, i := range r.Instances {
			ec2Instance := i

			if ec2Instance.InstanceId == nil || ec2Instance.State == nil {
				continue
			}

			ec2State := ec2Instance.State.Name
			if ec2State == ec2types.InstanceStateNameRunning || ec2State == ec2types.InstanceStateNamePending {
				return instanceFromEC2Instance(&ec2Instance, p.ec2Client, region, config), nil
			}
		}
	}

	userData, err := cloudinit.New(config)
	if err != nil {
		return nil, FatalError{
			Err: fmt.Errorf("failed to generate cloud-init: %w", err),
		}
	}
	userDataBase64 := base64.StdEncoding.EncodeToString([]byte(userData))

	runParams := &ec2.RunInstancesInput{
		MaxCount:     utils.PointerTo[int32](1),
		MinCount:     utils.PointerTo[int32](1),
		ImageId:      utils.PointerTo(p.config.ScannerImage),
		InstanceType: ec2types.InstanceType(p.config.ScannerInstanceType),
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeInstance,
				Tags:         ec2TagsForInstance,
			},
			{
				ResourceType: ec2types.ResourceTypeVolume,
				Tags:         ec2TagsForInstance,
			},
		},
		UserData: &userDataBase64,
		MetadataOptions: &ec2types.InstanceMetadataOptionsRequest{
			HttpEndpoint:         ec2types.InstanceMetadataEndpointStateEnabled,
			InstanceMetadataTags: ec2types.InstanceMetadataTagsStateEnabled,
		},
	}

	// Create network interface in the scanner subnet with the scanner security group.
	runParams.NetworkInterfaces = []ec2types.InstanceNetworkInterfaceSpecification{
		{
			AssociatePublicIpAddress: utils.PointerTo(false),
			DeleteOnTermination:      utils.PointerTo(true),
			DeviceIndex:              utils.PointerTo[int32](0),
			Groups:                   []string{p.config.SecurityGroupID},
			SubnetId:                 &p.config.SubnetID,
		},
	}

	var retryMaxAttempts int
	// Use spot instances if there is a configuration for it.
	if config.ScannerInstanceCreationConfig.UseSpotInstances {
		runParams.InstanceMarketOptions = &ec2types.InstanceMarketOptionsRequest{
			MarketType: ec2types.MarketTypeSpot,
			SpotOptions: &ec2types.SpotMarketOptions{
				InstanceInterruptionBehavior: ec2types.InstanceInterruptionBehaviorTerminate,
				SpotInstanceType:             ec2types.SpotInstanceTypeOneTime,
				MaxPrice:                     config.ScannerInstanceCreationConfig.MaxPrice,
			},
		}
	}
	// In the case of spot instances, we have higher probability to start an instance
	// by increasing RetryMaxAttempts
	if config.ScannerInstanceCreationConfig.RetryMaxAttempts != nil {
		retryMaxAttempts = *config.ScannerInstanceCreationConfig.RetryMaxAttempts
	}

	if p.config.KeyPairName != "" {
		// Set a key-pair to the instance.
		runParams.KeyName = &p.config.KeyPairName
	}

	// if retryMaxAttempts value is 0 it will be ignored
	out, err := p.ec2Client.RunInstances(ctx, runParams, options, func(options *ec2.Options) {
		options.RetryMaxAttempts = retryMaxAttempts
		options.RetryMode = awstype.RetryModeStandard
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}
	if len(out.Instances) < 1 {
		return nil, errors.New("failed to create instance: 0 instance in response")
	}

	return instanceFromEC2Instance(&out.Instances[0], p.ec2Client, region, config), nil
}

// nolint:cyclop,gocognit,maintidx
func (p *Provider) RunAssetScan(ctx context.Context, config *provider.ScanJobConfig) error {
	vmInfo, err := config.AssetInfo.AsVMInfo()
	if err != nil {
		return FatalError{Err: err}
	}

	logger := log.GetLoggerFromContextOrDefault(ctx).WithFields(logrus.Fields{
		"AssetInstanceID": vmInfo.InstanceID,
		"AssetLocation":   vmInfo.Location,
		"ScannerLocation": p.config.ScannerRegion,
		"Provider":        string(p.Kind()),
	})

	// Note(chrisgacsal): In order to speed up the initialization process the scanner instance and the volume are created
	//                    in parallel.

	// Create scanner instance
	numOfGoroutines := 2
	errs := make(chan error, numOfGoroutines)
	var scannnerInstance *Instance
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		logger.Trace("Creating scanner VM instance")

		var err error
		scannnerInstance, err = p.createInstance(ctx, p.config.ScannerRegion, config)
		if err != nil {
			errs <- WrapError(fmt.Errorf("failed to create scanner VM instance: %w", err))
			return
		}

		ready, err := scannnerInstance.IsReady(ctx)
		if err != nil {
			errs <- WrapError(fmt.Errorf("failed to get scanner VM instance state: %w", err))
			return
		}
		logger.WithFields(logrus.Fields{
			"ScannerInstanceID": scannnerInstance.ID,
		}).Debugf("Scanner instance is ready: %t", ready)
		if !ready {
			errs <- RetryableError{
				Err:   errors.New("scanner instance is not ready"),
				After: InstanceReadynessAfter,
			}
		}
	}()

	// Create volume snapshot from the root volume of the Asset Instance used for scanning by:
	// * fetching the Asset Instance from provider
	// * creating a volume snapshot from the root volume of the Asset Instance
	// * copying the volume snapshot to the region/location of the scanner instance if they are deployed in separate locations
	var destVolSnapshot *Snapshot
	wg.Add(1)
	go func() {
		defer wg.Done()

		logger.Debug("Getting asset VM instance")

		assetVMLocation, err := NewLocation(vmInfo.Location)
		if err != nil {
			errs <- FatalError{
				Err: fmt.Errorf("failed to parse Location for asset VM instance: %w", err),
			}
			return
		}

		var SrcEC2Instance *ec2types.Instance
		SrcEC2Instance, err = p.getInstanceWithID(ctx, vmInfo.InstanceID, assetVMLocation.Region)
		if err != nil {
			errs <- WrapError(fmt.Errorf("failed to fetch asset VM instance: %w", err))
			return
		}
		if SrcEC2Instance == nil {
			errs <- FatalError{
				Err: fmt.Errorf("failed to find asset VM instance. InstanceID=%s", vmInfo.InstanceID),
			}
			return
		}

		srcInstance := instanceFromEC2Instance(SrcEC2Instance, p.ec2Client, assetVMLocation.Region, config)

		logger.WithField("AssetInstanceID", srcInstance.ID).Trace("Found asset VM instance")

		srcVol := srcInstance.RootVolume()
		if srcVol == nil {
			errs <- FatalError{
				Err: errors.New("failed to get root block device for asset VM instance"),
			}
			return
		}

		logger.WithField("AssetVolumeID", srcVol.ID).Debug("Creating asset volume snapshot for asset VM instance")
		srcVolSnapshot, err := srcVol.CreateSnapshot(ctx)
		if err != nil {
			errs <- WrapError(fmt.Errorf("failed to create volume snapshot from asset volume. AssetVolumeID=%s: %w",
				srcVol.ID, err))
			return
		}

		ready, err := srcVolSnapshot.IsReady(ctx)
		if err != nil {
			err = fmt.Errorf("failed to get volume snapshot state. AssetVolumeSnapshotID=%s: %w",
				srcVolSnapshot.ID, err)
			errs <- WrapError(err)
			return
		}

		logger.WithFields(logrus.Fields{
			"AssetVolumeID":         srcVol.ID,
			"AssetVolumeSnapshotID": srcVolSnapshot.ID,
		}).Debugf("Asset volume snapshot is ready: %t", ready)
		if !ready {
			errs <- RetryableError{
				Err:   errors.New("asset volume snapshot is not ready"),
				After: SnapshotReadynessAfter,
			}
			return
		}

		logger.WithFields(logrus.Fields{
			"AssetVolumeID":         srcVol.ID,
			"AssetVolumeSnapshotID": srcVolSnapshot.ID,
		}).Debug("Copying asset volume snapshot to scanner location")
		destVolSnapshot, err = srcVolSnapshot.Copy(ctx, p.config.ScannerRegion)
		if err != nil {
			err = fmt.Errorf("failed to copy asset volume snapshot to location. AssetVolumeSnapshotID=%s Location=%s: %w",
				srcVolSnapshot.ID, p.config.ScannerRegion, err)
			errs <- WrapError(err)
			return
		}

		ready, err = destVolSnapshot.IsReady(ctx)
		if err != nil {
			err = fmt.Errorf("failed to get volume snapshot state. ScannerVolumeSnapshotID=%s: %w",
				srcVolSnapshot.ID, err)
			errs <- WrapError(err)
			return
		}

		logger.WithFields(logrus.Fields{
			"AssetVolumeID":           srcVol.ID,
			"AssetVolumeSnapshotID":   srcVolSnapshot.ID,
			"ScannerVolumeSnapshotID": destVolSnapshot.ID,
		}).Debugf("Scanner volume snapshot is ready: %t", ready)

		if !ready {
			errs <- RetryableError{
				Err:   errors.New("scanner volume snapshot is not ready"),
				After: SnapshotReadynessAfter,
			}
			return
		}
	}()
	wg.Wait()
	close(errs)

	// NOTE: make sure to drain results channel
	listOfErrors := make([]error, 0)
	for e := range errs {
		if e != nil {
			listOfErrors = append(listOfErrors, e)
		}
	}
	err = errors.Join(listOfErrors...)
	if err != nil {
		return err
	}

	// Create volume to be scanned from snapshot
	scannerInstanceAZ := scannnerInstance.AvailabilityZone
	logger.WithFields(logrus.Fields{
		"ScannerAvailabilityZone": scannerInstanceAZ,
		"ScannerVolumeSnapshotID": destVolSnapshot.ID,
	}).Debug("Creating scanner volume from volume snapshot for scanner VM instance")
	scannerVol, err := destVolSnapshot.CreateVolume(ctx, scannerInstanceAZ)
	if err != nil {
		err = fmt.Errorf("failed to create volume from snapshot. SnapshotID=%s: %w", destVolSnapshot.ID, err)
		return WrapError(err)
	}

	ready, err := scannerVol.IsReady(ctx)
	if err != nil {
		err = fmt.Errorf("failed to check if scanner volume is ready. ScannerVolumeID=%s: %w", scannerVol.ID, err)
		return WrapError(err)
	}
	logger.WithFields(logrus.Fields{
		"ScannerAvailabilityZone": scannerInstanceAZ,
		"ScannerVolumeSnapshotID": destVolSnapshot.ID,
		"ScannerVolumeID":         scannerVol.ID,
	}).Debugf("Scanner volume is ready: %t", ready)
	if !ready {
		return RetryableError{
			Err:   fmt.Errorf("scanner volume is not ready. ScannerVolumeID=%s", scannerVol.ID),
			After: VolumeReadynessAfter,
		}
	}

	// Attach volume to be scanned to scanner instance
	logger.WithFields(logrus.Fields{
		"ScannerAvailabilityZone": scannerInstanceAZ,
		"ScannerVolumeSnapshotID": destVolSnapshot.ID,
		"ScannerVolumeID":         scannerVol.ID,
		"ScannerIntanceID":        scannnerInstance.ID,
	}).Debug("Attaching scanner volume to scanner VM instance")
	err = scannnerInstance.AttachVolume(ctx, scannerVol, p.config.BlockDeviceName)
	if err != nil {
		err = fmt.Errorf("failed to attach volume to scanner instance. ScannerVolumeID=%s ScannerInstanceID=%s: %w",
			scannerVol.ID, scannnerInstance.ID, err)
		return WrapError(err)
	}

	// Wait until the volume is attached to the scanner instance
	logger.WithFields(logrus.Fields{
		"ScannerAvailabilityZone": scannerInstanceAZ,
		"ScannerVolumeSnapshotID": destVolSnapshot.ID,
		"ScannerVolumeID":         scannerVol.ID,
		"ScannerIntanceID":        scannnerInstance.ID,
	}).Debug("Checking if scanner volume is attached to scanner VM instance")
	ready, err = scannerVol.IsAttached(ctx)
	if err != nil {
		err = fmt.Errorf("failed to check if volume is attached to scanner instance. ScannerVolumeID=%s ScannerInstanceID=%s: %w",
			scannerVol.ID, scannnerInstance.ID, err)
		return WrapError(err)
	}
	if !ready {
		return RetryableError{
			Err:   fmt.Errorf("scanner volume is not attached yet. ScannerVolumeID=%s", scannerVol.ID),
			After: VolumeAttachmentReadynessAfter,
		}
	}

	return nil
}

// deleteInstances terminates all instances which meet the conditions defined by the filters argument.
// It returns:
//   - nil, error: if error happened during the operation
//   - false, nil: if operation is in-progress
//   - true, nil:  if the operation is done and safe to assume that related resources are released (netif, vol, etc)
func (p *Provider) deleteInstances(ctx context.Context, filters []ec2types.Filter, region string) (bool, error) {
	options := func(options *ec2.Options) {
		options.Region = region
	}

	describeParams := &ec2.DescribeInstancesInput{
		Filters: filters,
	}
	describeOut, err := p.ec2Client.DescribeInstances(ctx, describeParams, options)
	if err != nil {
		return false, fmt.Errorf("failed to fetch instances: %w", err)
	}

	instances := make([]string, 0)
	for _, r := range describeOut.Reservations {
		for _, i := range r.Instances {
			if i.InstanceId == nil || i.State == nil {
				continue
			}
			if i.State.Name != ec2types.InstanceStateNameTerminated {
				instances = append(instances, *i.InstanceId)
			}
		}
	}

	if len(instances) > 0 {
		terminateParams := &ec2.TerminateInstancesInput{
			InstanceIds: instances,
		}
		_, err = p.ec2Client.TerminateInstances(ctx, terminateParams, options)
		if err != nil {
			return false, fmt.Errorf("failed to terminate instances %v: %w", instances, err)
		}
		return false, nil
	}

	return true, nil
}

// deleteVolumes deletes all volumes which meet the conditions defined by the filters argument.
// It returns:
//   - nil, error: if error happened during the operation
//   - false, nil: if operation is in-progress
//   - true, nil:  if the operation is done
func (p *Provider) deleteVolumes(ctx context.Context, filters []ec2types.Filter, region string) (bool, error) {
	options := func(options *ec2.Options) {
		options.Region = region
	}

	describeParams := &ec2.DescribeVolumesInput{
		Filters: filters,
	}
	describeOut, err := p.ec2Client.DescribeVolumes(ctx, describeParams, options)
	if err != nil {
		return false, fmt.Errorf("failed to fetch volumes: %w", err)
	}

	volumes := make([]string, 0)
	for _, vol := range describeOut.Volumes {
		if vol.State == ec2types.VolumeStateDeleted || vol.State == ec2types.VolumeStateDeleting {
			continue
		}
		volumes = append(volumes, *vol.VolumeId)
	}

	if len(volumes) > 0 {
		for _, vol := range volumes {
			terminateParams := &ec2.DeleteVolumeInput{
				VolumeId: utils.PointerTo(vol),
			}
			_, err = p.ec2Client.DeleteVolume(ctx, terminateParams, options)
			if err != nil {
				return false, fmt.Errorf("failed to delete volume with %s id: %w", vol, err)
			}
		}
		return false, nil
	}

	return true, nil
}

// deleteVolumeSnapshots deletes all volume snapshots which meet the conditions defined by the filters argument.
// It returns:
//   - nil, error: if error happened during the operation
//   - false, nil: if operation is in-progress
//   - true, nil:  if the operation is done
func (p *Provider) deleteVolumeSnapshots(ctx context.Context, filters []ec2types.Filter, region string) (bool, error) {
	options := func(options *ec2.Options) {
		options.Region = region
	}

	describeParams := &ec2.DescribeSnapshotsInput{
		Filters: filters,
	}
	describeOut, err := p.ec2Client.DescribeSnapshots(ctx, describeParams, options)
	if err != nil {
		return false, fmt.Errorf("failed to fetch volume snapshots: %w", err)
	}

	snapshots := make([]string, 0)
	for _, snap := range describeOut.Snapshots {
		snapshots = append(snapshots, *snap.SnapshotId)
	}

	for _, snap := range snapshots {
		deleteParams := &ec2.DeleteSnapshotInput{
			SnapshotId: utils.PointerTo(snap),
		}
		_, err = p.ec2Client.DeleteSnapshot(ctx, deleteParams, options)
		if err != nil {
			return false, fmt.Errorf("failed to delete volume snapshot with %s id: %w", snap, err)
		}
	}

	return true, nil
}

// RemoveAssetScan removes all the cloud resources associated with a Scan defined by config parameter.
// The operation is idempotent, therefore it is safe to call it multiple times.
// nolint:cyclop,gocognit
func (p *Provider) RemoveAssetScan(ctx context.Context, config *provider.ScanJobConfig) error {
	vmInfo, err := config.AssetInfo.AsVMInfo()
	if err != nil {
		return FatalError{Err: err}
	}

	logger := log.GetLoggerFromContextOrDefault(ctx).WithFields(logrus.Fields{
		"ScannerLocation": p.config.ScannerRegion,
		"Provider":        string(p.Kind()),
	})

	ec2Tags := EC2TagsFromScanMetadata(config.ScanMetadata)
	ec2Filters := EC2FiltersFromEC2Tags(ec2Tags)

	numOfGoroutines := 3
	errs := make(chan error, numOfGoroutines)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Delete scanner instance
		logger.Debug("Deleting scanner VM Instance.")
		done, err := p.deleteInstances(ctx, ec2Filters, p.config.ScannerRegion)
		if err != nil {
			errs <- WrapError(fmt.Errorf("failed to delete scanner VM instance: %w", err))
			return
		}
		// Deleting scanner VM instance is in-progress, thus cannot proceed with deleting the scanner volume.
		if !done {
			errs <- RetryableError{
				Err:   errors.New("deleting Scanner VM instance is in-progress"),
				After: InstanceReadynessAfter,
			}
			return
		}

		// Delete scanner volume
		logger.Debug("Deleting scanner volume.")
		done, err = p.deleteVolumes(ctx, ec2Filters, p.config.ScannerRegion)

		if err != nil {
			errs <- WrapError(fmt.Errorf("failed to delete scanner volume: %w", err))
			return
		}

		if !done {
			errs <- RetryableError{
				Err:   errors.New("deleting Scanner volume is in-progress"),
				After: VolumeReadynessAfter,
			}
			return
		}
	}()

	// Delete volume snapshot created for the scanner volume
	wg.Add(1)
	go func() {
		defer wg.Done()

		logger.Debug("Deleting scanner volume snapshot.")
		done, err := p.deleteVolumeSnapshots(ctx, ec2Filters, p.config.ScannerRegion)
		if err != nil {
			errs <- WrapError(fmt.Errorf("failed to delete scanner volume snapshot: %w", err))
			return
		}
		if !done {
			errs <- RetryableError{
				Err:   errors.New("deleting Scanner volume snapshot is in-progress"),
				After: SnapshotReadynessAfter,
			}
			return
		}
	}()

	// Delete volume snapshot created from the Asset instance volume
	wg.Add(1)
	go func() {
		defer wg.Done()

		location, err := NewLocation(vmInfo.Location)
		if err != nil {
			errs <- FatalError{
				Err: fmt.Errorf("failed to parse Location string. Location=%s: %w", vmInfo.Location, err),
			}
			return
		}

		if location.Region == p.config.ScannerRegion {
			return
		}

		logger.WithField("AssetLocation", vmInfo.Location).Debug("Deleting asset volume snapshot.")
		done, err := p.deleteVolumeSnapshots(ctx, ec2Filters, location.Region)
		if err != nil {
			errs <- fmt.Errorf("failed to delete asset volume snapshot: %w", err)
			return
		}

		if !done {
			errs <- RetryableError{
				Err:   errors.New("deleting Asset volume snapshot is in-progress"),
				After: SnapshotReadynessAfter,
			}
			return
		}
	}()
	wg.Wait()
	close(errs)

	// NOTE: make sure to drain results channel
	listOfErrors := make([]error, 0)
	for e := range errs {
		if e != nil {
			listOfErrors = append(listOfErrors, e)
		}
	}
	err = errors.Join(listOfErrors...)
	return err
}

func (p *Provider) GetInstances(ctx context.Context, filters []ec2types.Filter, regionID string) ([]Instance, error) {
	ret := make([]Instance, 0)

	input := &ec2.DescribeInstancesInput{
		MaxResults: utils.PointerTo[int32](maxResults), // TODO what will be a good number?
	}
	if len(filters) > 0 {
		input.Filters = filters
	}

	out, err := p.ec2Client.DescribeInstances(ctx, input, func(options *ec2.Options) {
		options.Region = regionID
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %w", err)
	}
	ret = append(ret, p.getInstancesFromDescribeInstancesOutput(ctx, out, regionID)...)

	// use pagination
	// TODO we can make it better by not saving all results in memory. See https://github.com/openclarity/vmclarity/pull/3#discussion_r1021656861
	for out.NextToken != nil {
		input := &ec2.DescribeInstancesInput{
			MaxResults: utils.PointerTo[int32](maxResults), // TODO what will be a good number?
			NextToken:  out.NextToken,
		}
		if len(filters) > 0 {
			input.Filters = filters
		}

		out, err = p.ec2Client.DescribeInstances(ctx, input, func(options *ec2.Options) {
			options.Region = regionID
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe instances: %w", err)
		}
		ret = append(ret, p.getInstancesFromDescribeInstancesOutput(ctx, out, regionID)...)
	}

	return ret, nil
}

func (p *Provider) getInstancesFromDescribeInstancesOutput(ctx context.Context, result *ec2.DescribeInstancesOutput, regionID string) []Instance {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	var ret []Instance
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			// Ignore terminated instances they are destroyed and
			// will be garbage collected by AWS.
			if instance.State.Name == ec2types.InstanceStateNameTerminated {
				continue
			}

			if err := validateInstanceFields(instance); err != nil {
				logger.Errorf("Instance validation failed. instance id=%v: %v", utils.StringPointerValOrEmpty(instance.InstanceId), err)
				continue
			}
			rootVol, err := getRootVolumeInfo(ctx, p.ec2Client, instance, regionID)
			if err != nil {
				logger.Warnf("Couldn't get root volume info. instance id=%v: %v", utils.StringPointerValOrEmpty(instance.InstanceId), err)
				rootVol = &models.RootVolume{
					SizeGB:    0,
					Encrypted: models.RootVolumeEncryptedUnknown,
				}
			}

			ret = append(ret, Instance{
				ID:                  *instance.InstanceId,
				Region:              regionID,
				AvailabilityZone:    *instance.Placement.AvailabilityZone,
				Image:               *instance.ImageId,
				InstanceType:        string(instance.InstanceType),
				Platform:            *instance.PlatformDetails,
				Tags:                getTagsFromECTags(instance.Tags),
				LaunchTime:          *instance.LaunchTime,
				VpcID:               *instance.VpcId,
				SecurityGroups:      getSecurityGroupsIDs(instance.SecurityGroups),
				RootDeviceName:      utils.StringPointerValOrEmpty(instance.RootDeviceName),
				RootVolumeSizeGB:    int32(rootVol.SizeGB),
				RootVolumeEncrypted: rootVol.Encrypted,

				ec2Client: p.ec2Client,
			})
		}
	}
	return ret
}

func getRootVolumeInfo(ctx context.Context, client *ec2.Client, i ec2types.Instance, region string) (*models.RootVolume, error) {
	if i.RootDeviceName == nil || *i.RootDeviceName == "" {
		return nil, fmt.Errorf("RootDeviceName is not set")
	}
	logger := log.GetLoggerFromContextOrDiscard(ctx)
	for _, mapping := range i.BlockDeviceMappings {
		if utils.StringPointerValOrEmpty(mapping.DeviceName) == utils.StringPointerValOrEmpty(i.RootDeviceName) {
			if mapping.Ebs == nil {
				return nil, fmt.Errorf("EBS of the root volume is nil")
			}
			if mapping.Ebs.VolumeId == nil {
				return nil, fmt.Errorf("volume ID of the root volume is nil")
			}
			descParams := &ec2.DescribeVolumesInput{
				VolumeIds: []string{*mapping.Ebs.VolumeId},
			}

			describeOut, err := client.DescribeVolumes(ctx, descParams, func(options *ec2.Options) {
				options.Region = region
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe the root volume")
			}

			if len(describeOut.Volumes) == 0 {
				return nil, fmt.Errorf("volume list is empty")
			}
			if len(describeOut.Volumes) > 1 {
				logger.WithFields(logrus.Fields{
					"VolumeId":    *mapping.Ebs.VolumeId,
					"Volumes num": len(describeOut.Volumes),
				}).Warnf("Found more than 1 root volume, using the first")
			}

			return &models.RootVolume{
				SizeGB:    int(utils.Int32PointerValOrEmpty(describeOut.Volumes[0].Size)),
				Encrypted: encryptedToAPI(describeOut.Volumes[0].Encrypted),
			}, nil
		}
	}

	return nil, fmt.Errorf("instance doesn't have a root volume block device mapping")
}

func encryptedToAPI(encrypted *bool) models.RootVolumeEncrypted {
	if encrypted == nil {
		return models.RootVolumeEncryptedUnknown
	}
	if *encrypted {
		return models.RootVolumeEncryptedYes
	}
	return models.RootVolumeEncryptedNo
}

func (p *Provider) ListAllRegions(ctx context.Context) ([]Region, error) {
	ret := make([]Region, 0)
	out, err := p.ec2Client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{
		AllRegions: nil, // display also disabled regions?
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe regions: %w", err)
	}

	for _, region := range out.Regions {
		ret = append(ret, Region{
			Name: *region.RegionName,
		})
	}

	return ret, nil
}
