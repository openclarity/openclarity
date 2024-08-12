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

package diskutil

import (
	"fmt"
)

type Bytes float64

type BlockDevice struct {
	DeviceIdentifier string `json:"deviceIdentifier,omitempty" mapstructure:"DEVICE_IDENTIFIER"`
	Path             string `json:"path,omitempty" mapstructure:"PATH"` // path to the device node
	Whole            bool   `json:"whole,omitempty" mapstructure:"WHOLE"`
	PartOfWhole      string `json:"partOfWhole,omitempty" mapstructure:"PART_OF_WHOLE"`
	DeviceMediaName  string `json:"deviceMediaName,omitempty" mapstructure:"DEVICE_MEDIA_NAME"`

	VolumeName string `json:"volumeName,omitempty" mapstructure:"VOLUME_NAME"`
	Mounted    bool   `json:"mounted,omitempty" mapstructure:"MOUNTED"`
	MountPoint string `json:"mountPoint,omitempty" mapstructure:"MOUNT_POINT"`
	FileSystem string `json:"fileSystem,omitempty" mapstructure:"FILE_SYSTEM"`

	PartitionType   string `json:"partitionType,omitempty" mapstructure:"PARTITION_TYPE"`
	FSType          string `json:"fstype,omitempty" mapstructure:"FSTYPE"` // filesystem type
	TypeBundle      string `json:"typeBundle,omitempty" mapstructure:"TYPE_BUNDLE"`
	NameUserVisible string `json:"nameUserVisible,omitempty" mapstructure:"NAME_USER_VISIBLE"`
	Owners          string `json:"owners,omitempty" mapstructure:"OWNERS"`

	Content           string `json:"content,omitempty" mapstructure:"CONTENT"`
	OSCanBeInstalled  bool   `json:"osCanBeInstalled,omitempty" mapstructure:"OS_CAN_BE_INSTALLED"`
	BooterDisk        string `json:"booterDisk,omitempty" mapstructure:"BOOTER_DISK"`
	RecoveryDisk      string `json:"recoveryDisk,omitempty" mapstructure:"RECVOERY_DISK"`
	MediaType         string `json:"mediaType,omitempty" mapstructure:"MEDIA_TYPE"`
	Protocol          string `json:"protocol,omitempty" mapstructure:"PROTOCOL"`
	SMARTStatus       string `json:"smartStatus,omitempty" mapstructure:"SMART_STATUS"`
	VolumeUUID        string `json:"volumeUUID,omitempty" mapstructure:"VOLUME_UUID"`
	DiskPartitionUUID string `json:"diskPartitionUUID,omitempty" mapstructure:"DISK_PARTITION_UUID"`
	PartitionOffset   Bytes  `json:"partitionOffset,omitempty" mapstructure:"PARTITION_OFFSET"`

	DiskSize        Bytes `json:"diskSize,omitempty" mapstructure:"DISK_SIZE"`
	DeviceBlockSize Bytes `json:"deviceBlockSize,omitempty" mapstructure:"DEVICE_BLOCK_SIZE"`

	VolumeUsedSpace     Bytes `json:"volumeUsedSpace,omitempty" mapstructure:"VOLUME_USED_SPACE"`
	ContainerTotalSpace Bytes `json:"containerTotalSpace,omitempty" mapstructure:"CONTAINER_TOTAL_SPACE"`
	ContainerFreeSpace  Bytes `json:"containerFreeSpace,omitempty" mapstructure:"CONTAINER_FREE_SPACE"`
	AllocationBlockSize Bytes `json:"allocationBlockSize,omitempty" mapstructure:"ALLOCATION_BLOCK_SIZE"`

	MediaOSUseOnly bool `json:"mediaOSUseOnly,omitempty" mapstructure:"MEDIA_OS_USE_ONLY"`
	MediaReadOnly  bool `json:"mediaReadOnly,omitempty" mapstructure:"MEDIA_READ_ONLY"`
	VolumeReadOnly bool `json:"volumeReadOnly,omitempty" mapstructure:"VOLUME_READ_ONLY"`

	DeviceLocation string `json:"deviceLocation,omitempty" mapstructure:"DEVICE_LOCATION"`
	RemovableMedia string `json:"removableMedia,omitempty" mapstructure:"REMOVABLE_MEDIA"`

	SolidState         bool `json:"solidState,omitempty" mapstructure:"SOLID_STATE"`
	Virtual            bool `json:"virtual,omitempty" mapstructure:"VIRTUAL"`
	HardwareAESSupport bool `json:"hardwareAESSupport,omitempty" mapstructure:"HARDWARE_AES_SUPPORT"`
}

func (b BlockDevice) String() string {
	return fmt.Sprintf("Device: %s Filesystem: %s MountPoint: %s", b.DeviceIdentifier, b.FSType, b.MountPoint)
}
