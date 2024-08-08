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

package scanner

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"

	awstype "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/sirupsen/logrus"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/provider/aws/types"
	"github.com/openclarity/vmclarity/provider/aws/utils"
	"github.com/openclarity/vmclarity/provider/cloudinit"
)

var _ provider.Scanner = &Scanner{}

type Scanner struct {
	Kind                apitypes.CloudProvider
	ScannerRegion       string
	BlockDeviceName     string
	ScannerImage        string
	ScannerInstanceType string
	SecurityGroupID     string
	SubnetID            string
	KeyPairName         string
	Ec2Client           *ec2.Client
}

// nolint:cyclop,gocognit,maintidx,wrapcheck
func (s *Scanner) RunAssetScan(ctx context.Context, t *provider.ScanJobConfig) error {
	vmInfo, err := t.AssetInfo.AsVMInfo()
	if err != nil {
		return utils.FatalError{Err: err}
	}

	logger := log.GetLoggerFromContextOrDefault(ctx).WithFields(logrus.Fields{
		"AssetInstanceID": vmInfo.InstanceID,
		"AssetLocation":   vmInfo.Location,
		"ScannerLocation": s.ScannerRegion,
		"Provider":        string(s.Kind),
	})

	// Note(chrisgacsal): In order to speed up the initialization process the scanner instance and the volume are created
	//                    in parallel.

	// Create scanner instance
	numOfGoroutines := 2
	errs := make(chan error, numOfGoroutines)
	var scannnerInstance *types.Instance
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		logger.Trace("Creating scanner VM instance")

		var err error
		scannnerInstance, err = s.createInstance(ctx, s.ScannerRegion, t)
		if err != nil {
			errs <- utils.WrapError(fmt.Errorf("failed to create scanner VM instance: %w", err))
			return
		}

		ready, err := scannnerInstance.IsReady(ctx)
		if err != nil {
			errs <- utils.WrapError(fmt.Errorf("failed to get scanner VM instance state: %w", err))
			return
		}
		logger.WithFields(logrus.Fields{
			"ScannerInstanceID": scannnerInstance.ID,
		}).Debugf("Scanner instance is ready: %t", ready)
		if !ready {
			errs <- utils.RetryableError{
				Err:   errors.New("scanner instance is not ready"),
				After: utils.InstanceReadynessAfter,
			}
		}
	}()

	// Create volume snapshot from the root volume of the Asset Instance used for scanning by:
	// * fetching the Asset Instance from provider
	// * creating a volume snapshot from the root volume of the Asset Instance
	// * copying the volume snapshot to the region/location of the scanner instance if they are deployed in separate locations
	var destVolSnapshot *types.Snapshot
	wg.Add(1)
	go func() {
		defer wg.Done()

		logger.Debug("Getting asset VM instance")

		assetVMLocation, err := types.NewLocation(vmInfo.Location)
		if err != nil {
			errs <- utils.FatalError{
				Err: fmt.Errorf("failed to parse Location for asset VM instance: %w", err),
			}
			return
		}

		var SrcEC2Instance *ec2types.Instance
		SrcEC2Instance, err = s.getInstanceWithID(ctx, vmInfo.InstanceID, assetVMLocation.Region)
		if err != nil {
			errs <- utils.WrapError(fmt.Errorf("failed to fetch asset VM instance: %w", err))
			return
		}
		if SrcEC2Instance == nil {
			errs <- utils.FatalError{
				Err: fmt.Errorf("failed to find asset VM instance. InstanceID=%s", vmInfo.InstanceID),
			}
			return
		}

		srcInstance := instanceFromEC2Instance(SrcEC2Instance, s.Ec2Client, assetVMLocation.Region, t)

		logger.WithField("AssetInstanceID", srcInstance.ID).Trace("Found asset VM instance")

		srcVol := srcInstance.RootVolume()
		if srcVol == nil {
			errs <- utils.FatalError{
				Err: errors.New("failed to get root block device for asset VM instance"),
			}
			return
		}

		logger.WithField("AssetVolumeID", srcVol.ID).Debug("Creating asset volume snapshot for asset VM instance")
		srcVolSnapshot, err := srcVol.CreateSnapshot(ctx)
		if err != nil {
			errs <- utils.WrapError(fmt.Errorf("failed to create volume snapshot from asset volume. AssetVolumeID=%s: %w",
				srcVol.ID, err))
			return
		}

		ready, err := srcVolSnapshot.IsReady(ctx)
		if err != nil {
			err = fmt.Errorf("failed to get volume snapshot state. AssetVolumeSnapshotID=%s: %w",
				srcVolSnapshot.ID, err)
			errs <- utils.WrapError(err)
			return
		}

		logger.WithFields(logrus.Fields{
			"AssetVolumeID":         srcVol.ID,
			"AssetVolumeSnapshotID": srcVolSnapshot.ID,
		}).Debugf("Asset volume snapshot is ready: %t", ready)
		if !ready {
			errs <- utils.RetryableError{
				Err:   errors.New("asset volume snapshot is not ready"),
				After: utils.SnapshotReadynessAfter,
			}
			return
		}

		logger.WithFields(logrus.Fields{
			"AssetVolumeID":         srcVol.ID,
			"AssetVolumeSnapshotID": srcVolSnapshot.ID,
		}).Debug("Copying asset volume snapshot to scanner location")
		destVolSnapshot, err = srcVolSnapshot.Copy(ctx, s.ScannerRegion)
		if err != nil {
			err = fmt.Errorf("failed to copy asset volume snapshot to location. AssetVolumeSnapshotID=%s Location=%s: %w",
				srcVolSnapshot.ID, s.ScannerRegion, err)
			errs <- utils.WrapError(err)
			return
		}

		ready, err = destVolSnapshot.IsReady(ctx)
		if err != nil {
			err = fmt.Errorf("failed to get volume snapshot state. ScannerVolumeSnapshotID=%s: %w",
				srcVolSnapshot.ID, err)
			errs <- utils.WrapError(err)
			return
		}

		logger.WithFields(logrus.Fields{
			"AssetVolumeID":           srcVol.ID,
			"AssetVolumeSnapshotID":   srcVolSnapshot.ID,
			"ScannerVolumeSnapshotID": destVolSnapshot.ID,
		}).Debugf("Scanner volume snapshot is ready: %t", ready)

		if !ready {
			errs <- utils.RetryableError{
				Err:   errors.New("scanner volume snapshot is not ready"),
				After: utils.SnapshotReadynessAfter,
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
		return utils.WrapError(err)
	}

	ready, err := scannerVol.IsReady(ctx)
	if err != nil {
		err = fmt.Errorf("failed to check if scanner volume is ready. ScannerVolumeID=%s: %w", scannerVol.ID, err)
		return utils.WrapError(err)
	}
	logger.WithFields(logrus.Fields{
		"ScannerAvailabilityZone": scannerInstanceAZ,
		"ScannerVolumeSnapshotID": destVolSnapshot.ID,
		"ScannerVolumeID":         scannerVol.ID,
	}).Debugf("Scanner volume is ready: %t", ready)
	if !ready {
		return utils.RetryableError{
			Err:   fmt.Errorf("scanner volume is not ready. ScannerVolumeID=%s", scannerVol.ID),
			After: utils.VolumeReadynessAfter,
		}
	}

	// Attach volume to be scanned to scanner instance
	logger.WithFields(logrus.Fields{
		"ScannerAvailabilityZone": scannerInstanceAZ,
		"ScannerVolumeSnapshotID": destVolSnapshot.ID,
		"ScannerVolumeID":         scannerVol.ID,
		"ScannerIntanceID":        scannnerInstance.ID,
	}).Debug("Attaching scanner volume to scanner VM instance")
	err = scannnerInstance.AttachVolume(ctx, scannerVol, s.BlockDeviceName)
	if err != nil {
		err = fmt.Errorf("failed to attach volume to scanner instance. ScannerVolumeID=%s ScannerInstanceID=%s: %w",
			scannerVol.ID, scannnerInstance.ID, err)
		return utils.WrapError(err)
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
		return utils.WrapError(err)
	}
	if !ready {
		return utils.RetryableError{
			Err:   fmt.Errorf("scanner volume is not attached yet. ScannerVolumeID=%s", scannerVol.ID),
			After: utils.VolumeAttachmentReadynessAfter,
		}
	}

	return nil
}

// RemoveAssetScan removes all the cloud resources associated with a Scan defined by config parameter.
// The operation is idempotent, therefore it is safe to call it multiple times.
func (s *Scanner) RemoveAssetScan(ctx context.Context, t *provider.ScanJobConfig) error {
	vmInfo, err := t.AssetInfo.AsVMInfo()
	if err != nil {
		return utils.FatalError{Err: err}
	}

	logger := log.GetLoggerFromContextOrDefault(ctx).WithFields(logrus.Fields{
		"ScannerLocation": s.ScannerRegion,
		"Provider":        string(s.Kind),
	})

	ec2Tags := types.EC2TagsFromScanMetadata(t.ScanMetadata)
	ec2Filters := utils.EC2FiltersFromEC2Tags(ec2Tags)

	numOfGoroutines := 3
	errs := make(chan error, numOfGoroutines)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Delete scanner instance
		logger.Debug("Deleting scanner VM Instance.")
		done, err := s.deleteInstances(ctx, ec2Filters, s.ScannerRegion)
		if err != nil {
			errs <- utils.WrapError(fmt.Errorf("failed to delete scanner VM instance: %w", err))
			return
		}
		// Deleting scanner VM instance is in-progress, thus cannot proceed with deleting the scanner volume.
		if !done {
			errs <- utils.RetryableError{
				Err:   errors.New("deleting Scanner VM instance is in-progress"),
				After: utils.InstanceReadynessAfter,
			}
			return
		}

		// Delete scanner volume
		logger.Debug("Deleting scanner volume.")
		done, err = s.deleteVolumes(ctx, ec2Filters, s.ScannerRegion)
		if err != nil {
			errs <- utils.WrapError(fmt.Errorf("failed to delete scanner volume: %w", err))
			return
		}

		if !done {
			errs <- utils.RetryableError{
				Err:   errors.New("deleting Scanner volume is in-progress"),
				After: utils.VolumeReadynessAfter,
			}
			return
		}
	}()

	// Delete volume snapshot created for the scanner volume
	wg.Add(1)
	go func() {
		defer wg.Done()

		logger.Debug("Deleting scanner volume snapshot.")
		done, err := s.deleteVolumeSnapshots(ctx, ec2Filters, s.ScannerRegion)
		if err != nil {
			errs <- utils.WrapError(fmt.Errorf("failed to delete scanner volume snapshot: %w", err))
			return
		}
		if !done {
			errs <- utils.RetryableError{
				Err:   errors.New("deleting Scanner volume snapshot is in-progress"),
				After: utils.SnapshotReadynessAfter,
			}
			return
		}
	}()

	// Delete volume snapshot created from the Asset instance volume
	wg.Add(1)
	go func() {
		defer wg.Done()

		location, err := types.NewLocation(vmInfo.Location)
		if err != nil {
			errs <- utils.FatalError{
				Err: fmt.Errorf("failed to parse Location string. Location=%s: %w", vmInfo.Location, err),
			}
			return
		}

		if location.Region == s.ScannerRegion {
			return
		}

		logger.WithField("AssetLocation", vmInfo.Location).Debug("Deleting asset volume snapshot.")
		done, err := s.deleteVolumeSnapshots(ctx, ec2Filters, location.Region)
		if err != nil {
			errs <- fmt.Errorf("failed to delete asset volume snapshot: %w", err)
			return
		}

		if !done {
			errs <- utils.RetryableError{
				Err:   errors.New("deleting Asset volume snapshot is in-progress"),
				After: utils.SnapshotReadynessAfter,
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

// nolint:nilnil
func (s *Scanner) getInstanceWithID(ctx context.Context, id string, region string) (*ec2types.Instance, error) {
	out, err := s.Ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
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
func (s *Scanner) createInstance(ctx context.Context, region string, config *provider.ScanJobConfig) (*types.Instance, error) {
	options := func(options *ec2.Options) {
		options.Region = region
	}

	ec2TagsForInstance := types.EC2TagsFromScanMetadata(config.ScanMetadata)
	ec2Filters := utils.EC2FiltersFromEC2Tags(ec2TagsForInstance)

	describeParams := &ec2.DescribeInstancesInput{
		Filters: ec2Filters,
	}
	describeOut, err := s.Ec2Client.DescribeInstances(ctx, describeParams, options)
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
				return instanceFromEC2Instance(&ec2Instance, s.Ec2Client, region, config), nil
			}
		}
	}

	userData, err := cloudinit.New(config)
	if err != nil {
		return nil, utils.FatalError{
			Err: fmt.Errorf("failed to generate cloud-init: %w", err),
		}
	}
	userDataBase64 := base64.StdEncoding.EncodeToString([]byte(userData))

	runParams := &ec2.RunInstancesInput{
		MaxCount:     to.Ptr[int32](1),
		MinCount:     to.Ptr[int32](1),
		ImageId:      to.Ptr(s.ScannerImage),
		InstanceType: ec2types.InstanceType(s.ScannerInstanceType),
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
			AssociatePublicIpAddress: to.Ptr(false),
			DeleteOnTermination:      to.Ptr(true),
			DeviceIndex:              to.Ptr[int32](0),
			Groups:                   []string{s.SecurityGroupID},
			SubnetId:                 &s.SubnetID,
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

	if s.KeyPairName != "" {
		// Set a key-pair to the instance.
		runParams.KeyName = &s.KeyPairName
	}

	// if retryMaxAttempts value is 0 it will be ignored
	out, err := s.Ec2Client.RunInstances(ctx, runParams, options, func(options *ec2.Options) {
		options.RetryMaxAttempts = retryMaxAttempts
		options.RetryMode = awstype.RetryModeStandard
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}
	if len(out.Instances) < 1 {
		return nil, errors.New("failed to create instance: 0 instance in response")
	}

	return instanceFromEC2Instance(&out.Instances[0], s.Ec2Client, region, config), nil
}

// deleteInstances terminates all instances which meet the conditions defined by the filters argument.
// It returns:
//   - nil, error: if error happened during the operation
//   - false, nil: if operation is in-progress
//   - true, nil:  if the operation is done and safe to assume that related resources are released (netif, vol, etc)
func (s *Scanner) deleteInstances(ctx context.Context, filters []ec2types.Filter, region string) (bool, error) {
	options := func(options *ec2.Options) {
		options.Region = region
	}

	describeParams := &ec2.DescribeInstancesInput{
		Filters: filters,
	}
	describeOut, err := s.Ec2Client.DescribeInstances(ctx, describeParams, options)
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
		_, err = s.Ec2Client.TerminateInstances(ctx, terminateParams, options)
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
func (s *Scanner) deleteVolumes(ctx context.Context, filters []ec2types.Filter, region string) (bool, error) {
	options := func(options *ec2.Options) {
		options.Region = region
	}

	describeParams := &ec2.DescribeVolumesInput{
		Filters: filters,
	}
	describeOut, err := s.Ec2Client.DescribeVolumes(ctx, describeParams, options)
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
				VolumeId: to.Ptr(vol),
			}
			_, err = s.Ec2Client.DeleteVolume(ctx, terminateParams, options)
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
func (s *Scanner) deleteVolumeSnapshots(ctx context.Context, filters []ec2types.Filter, region string) (bool, error) {
	options := func(options *ec2.Options) {
		options.Region = region
	}

	describeParams := &ec2.DescribeSnapshotsInput{
		Filters: filters,
	}
	describeOut, err := s.Ec2Client.DescribeSnapshots(ctx, describeParams, options)
	if err != nil {
		return false, fmt.Errorf("failed to fetch volume snapshots: %w", err)
	}

	snapshots := make([]string, 0)
	for _, snap := range describeOut.Snapshots {
		snapshots = append(snapshots, *snap.SnapshotId)
	}

	for _, snap := range snapshots {
		deleteParams := &ec2.DeleteSnapshotInput{
			SnapshotId: to.Ptr(snap),
		}
		_, err = s.Ec2Client.DeleteSnapshot(ctx, deleteParams, options)
		if err != nil {
			return false, fmt.Errorf("failed to delete volume snapshot with %s id: %w", snap, err)
		}
	}

	return true, nil
}

func instanceFromEC2Instance(i *ec2types.Instance, client *ec2.Client, region string, config *provider.ScanJobConfig) *types.Instance {
	securityGroups := getSecurityGroupsFromEC2GroupIdentifiers(i.SecurityGroups)
	tags := utils.GetTagsFromECTags(i.Tags)

	volumes := make([]types.Volume, len(i.BlockDeviceMappings))
	for idx, blkDevice := range i.BlockDeviceMappings {
		var blockDeviceName string

		if blkDevice.DeviceName != nil {
			blockDeviceName = *blkDevice.DeviceName
		}

		volumes[idx] = types.Volume{
			Ec2Client:       client,
			ID:              *blkDevice.Ebs.VolumeId,
			Region:          region,
			BlockDeviceName: blockDeviceName,
			Metadata:        config.ScanMetadata,
		}
	}

	return &types.Instance{
		ID:               *i.InstanceId,
		Region:           region,
		VpcID:            *i.VpcId,
		SecurityGroups:   securityGroups,
		AvailabilityZone: *i.Placement.AvailabilityZone,
		Image:            *i.ImageId,
		InstanceType:     string(i.InstanceType),
		Platform:         string(i.Platform),
		Tags:             tags,
		LaunchTime:       *i.LaunchTime,
		RootDeviceName:   *i.RootDeviceName,
		Volumes:          volumes,
		Metadata:         config.ScanMetadata,

		Ec2Client: client,
	}
}

func getSecurityGroupsFromEC2GroupIdentifiers(identifiers []ec2types.GroupIdentifier) []apitypes.SecurityGroup {
	var ret []apitypes.SecurityGroup

	for _, identifier := range identifiers {
		if identifier.GroupId != nil {
			ret = append(ret, apitypes.SecurityGroup{
				Id: *identifier.GroupId,
			})
		}
	}

	return ret
}
