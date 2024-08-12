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
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// Parse diskutil output into a list of BlockDevice objects.
// Example output:
//    Device Identifier:         disk3s5
//    Device Node:               /dev/disk3s5
//    Whole:                     No
//    Part of Whole:             disk3

//    Volume Name:               Data
//    Mounted:                   Yes
//    Mount Point:               /System/Volumes/Data

//    Partition Type:            41504653-0000-11AA-AA11-00306543ECAC
//    File System Personality:   APFS
//    Type (Bundle):             apfs
//    Name (User Visible):       APFS
//    Owners:                    Enabled

//    OS Can Be Installed:       Yes
//    Booter Disk:               disk3s2
//    Recovery Disk:             disk3s3
//    Media Type:                Generic
//    Protocol:                  Apple Fabric
//    SMART Status:              Verified
//    Volume UUID:               10F7AAC0-5602-48CD-8BB3-80DBB7A03D91
//    Disk / Partition UUID:     10F7AAC0-5602-48CD-8BB3-80DBB7A03D91

//    Disk Size:                 494.4 GB (494384795648 Bytes) (exactly 965595304 512-Byte-Units)
//    Device Block Size:         4096 Bytes

//    Volume Used Space:         192.7 GB (192652795904 Bytes) (exactly 376274992 512-Byte-Units)
//    Container Total Space:     494.4 GB (494384795648 Bytes) (exactly 965595304 512-Byte-Units)
//    Container Free Space:      282.0 GB (281962233856 Bytes) (exactly 550707488 512-Byte-Units)
//    Allocation Block Size:     4096 Bytes

//    Media OS Use Only:         No
//    Media Read-Only:           No
//    Volume Read-Only:          No

//    Device Location:           Internal
//    Removable Media:           Fixed

//    Solid State:               Yes
//    Hardware AES Support:      Yes

//    This disk is an APFS Volume.  APFS Information:
//    APFS Container:            disk3
//    APFS Physical Store:       disk0s2
//    Fusion Drive:              No
//    APFS Volume Group:         10F7AAC0-5602-48CD-8BB3-80DBB7A03D91
//    FileVault:                 Yes
//    Sealed:                    No
//    Locked:                    No

// **********

//    Device Identifier:         disk3s6
//    Device Node:               /dev/disk3s6
//    Whole:                     No
//    Part of Whole:             disk3

//    Volume Name:               VM
//    Mounted:                   Yes
//    Mount Point:               /System/Volumes/VM

//    Partition Type:            41504653-0000-11AA-AA11-00306543ECAC
//    File System Personality:   APFS
//    Type (Bundle):             apfs
//    Name (User Visible):       APFS
//    Owners:                    Enabled

//    OS Can Be Installed:       No
//    Booter Disk:               disk3s2
//    Recovery Disk:             disk3s3
//    Media Type:                Generic
//    Protocol:                  Apple Fabric
//    SMART Status:              Verified
//    Volume UUID:               7E185E38-F7E8-41D2-B5E9-48A7BF23577F
//    Disk / Partition UUID:     7E185E38-F7E8-41D2-B5E9-48A7BF23577F

//    Disk Size:                 494.4 GB (494384795648 Bytes) (exactly 965595304 512-Byte-Units)
//    Device Block Size:         4096 Bytes

//    Volume Used Space:         2.1 GB (2147520512 Bytes) (exactly 4194376 512-Byte-Units)
//    Container Total Space:     494.4 GB (494384795648 Bytes) (exactly 965595304 512-Byte-Units)
//    Container Free Space:      282.0 GB (281962233856 Bytes) (exactly 550707488 512-Byte-Units)
//    Allocation Block Size:     4096 Bytes

//    Media OS Use Only:         No
//    Media Read-Only:           No
//    Volume Read-Only:          No

//    Device Location:           Internal
//    Removable Media:           Fixed

//    Solid State:               Yes
//    Hardware AES Support:      Yes

//    This disk is an APFS Volume.  APFS Information:
//    APFS Container:            disk3
//    APFS Physical Store:       disk0s2
//    Fusion Drive:              No
//    Encrypted:                 No
//    FileVault:                 No
//    Sealed:                    No
//    Locked:                    No

// **********

func parse(b *bytes.Buffer) ([]BlockDevice, error) {
	if b == nil {
		return nil, nil
	}

	var blockDevices []BlockDevice
	var currentBlockDevice BlockDevice

	for {
		line, err := b.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("failed to read line: %w", err)
		}

		if len(line) == 0 {
			continue
		}

		if line == "**********\n" {
			blockDevices = append(blockDevices, currentBlockDevice)
			currentBlockDevice = BlockDevice{}
			continue
		}

		if err := parseLine(line, &currentBlockDevice); err != nil {
			return nil, fmt.Errorf("failed to parse line: %w", err)
		}
	}

	return blockDevices, nil
}

// nolint:goconst,gocyclo,cyclop
func parseLine(line string, bd *BlockDevice) error {
	var err error
	switch {
	case strings.Contains(line, "Device Identifier:"):
		bd.DeviceIdentifier = parseString(line)
	case strings.Contains(line, "Device Node:"):
		bd.Path = parseString(line)
	case strings.Contains(line, "Part of Whole:"):
		bd.PartOfWhole = parseString(line)
	case strings.Contains(line, "Whole:"):
		bd.Whole = strings.TrimSpace(strings.Split(line, ":")[1]) == "Yes"
	case strings.Contains(line, "Device / Media Name:"):
		bd.DeviceMediaName = parseString(line)

	case strings.Contains(line, "Volume Name:"):
		bd.VolumeName = parseString(line)
	case strings.Contains(line, "Mounted:"):
		bd.Mounted = strings.TrimSpace(strings.Split(line, ":")[1]) == "Yes"
	case strings.Contains(line, "Mount Point:"):
		bd.MountPoint = parseString(line)
	case strings.Contains(line, "File System:"):
		bd.FileSystem = parseString(line)

	case strings.Contains(line, "Partition Type:"):
		bd.PartitionType = parseString(line)
	case strings.Contains(line, "File System Personality:"):
		bd.FSType = parseString(line)
	case strings.Contains(line, "Type (Bundle):"):
		bd.TypeBundle = parseString(line)
	case strings.Contains(line, "Name (User Visible):"):
		bd.NameUserVisible = parseString(line)
	case strings.Contains(line, "Owners:"):
		bd.Owners = parseString(line)

	case strings.Contains(line, "Content (IOContent):"):
		bd.Content = parseString(line)
	case strings.Contains(line, "OS Can Be Installed:"):
		bd.OSCanBeInstalled = strings.TrimSpace(strings.Split(line, ":")[1]) == "Yes"
	case strings.Contains(line, "Booter Disk:"):
		bd.BooterDisk = parseString(line)
	case strings.Contains(line, "Recovery Disk:"):
		bd.RecoveryDisk = parseString(line)
	case strings.Contains(line, "Media Type:"):
		bd.MediaType = parseString(line)
	case strings.Contains(line, "Protocol:"):
		bd.Protocol = parseString(line)
	case strings.Contains(line, "SMART Status:"):
		bd.SMARTStatus = parseString(line)
	case strings.Contains(line, "Volume UUID:"):
		bd.VolumeUUID = parseString(line)
	case strings.Contains(line, "Disk / Partition UUID:"):
		bd.DiskPartitionUUID = parseString(line)
	case strings.Contains(line, "Partition Offset:"):
		bd.PartitionOffset, err = parseBytes(strings.TrimSpace(strings.Split(line, ":")[1]))
		if err != nil {
			return fmt.Errorf("failed to parse partition offset: %w", err)
		}

	case strings.Contains(line, "Disk Size:"):
		bd.DiskSize, err = parseBytes(strings.TrimSpace(strings.Split(line, ":")[1]))
		if err != nil {
			return fmt.Errorf("failed to parse disk size: %w", err)
		}
	case strings.Contains(line, "Device Block Size:"):
		bd.DeviceBlockSize, err = parseBytes(strings.TrimSpace(strings.Split(line, ":")[1]))
		if err != nil {
			return fmt.Errorf("failed to parse device block size: %w", err)
		}

	case strings.Contains(line, "Volume Used Space:"):
		bd.VolumeUsedSpace, err = parseBytes(strings.TrimSpace(strings.Split(line, ":")[1]))
		if err != nil {
			return fmt.Errorf("failed to parse volume used space: %w", err)
		}
	case strings.Contains(line, "Container Total Space:"):
		bd.ContainerTotalSpace, err = parseBytes(strings.TrimSpace(strings.Split(line, ":")[1]))
		if err != nil {
			return fmt.Errorf("failed to parse container total space: %w", err)
		}
	case strings.Contains(line, "Container Free Space:"):
		bd.ContainerFreeSpace, err = parseBytes(strings.TrimSpace(strings.Split(line, ":")[1]))
		if err != nil {
			return fmt.Errorf("failed to parse container free space: %w", err)
		}
	case strings.Contains(line, "Allocation Block Size:"):
		bd.AllocationBlockSize, err = parseBytes(strings.TrimSpace(strings.Split(line, ":")[1]))
		if err != nil {
			return fmt.Errorf("failed to parse allocation block size: %w", err)
		}

	case strings.Contains(line, "Media OS Use Only:"):
		bd.MediaOSUseOnly = strings.TrimSpace(strings.Split(line, ":")[1]) == "Yes"
	case strings.Contains(line, "Media Read-Only:"):
		bd.MediaReadOnly = strings.TrimSpace(strings.Split(line, ":")[1]) == "Yes"
	case strings.Contains(line, "Volume Read-Only:"):
		bd.VolumeReadOnly = strings.TrimSpace(strings.Split(line, ":")[1]) == "Yes"

	case strings.Contains(line, "Device Location:"):
		bd.DeviceLocation = parseString(line)
	case strings.Contains(line, "Removable Media:"):
		bd.RemovableMedia = parseString(line)

	case strings.Contains(line, "Solid State:"):
		bd.SolidState = strings.TrimSpace(strings.Split(line, ":")[1]) == "Yes"
	case strings.Contains(line, "Virtual:"):
		bd.Virtual = strings.TrimSpace(strings.Split(line, ":")[1]) == "Yes"
	case strings.Contains(line, "Hardware AES Support:"):
		bd.HardwareAESSupport = strings.TrimSpace(strings.Split(line, ":")[1]) == "Yes"
	default:
		// Ignore unknown lines
	}

	return nil
}

func parseString(s string) string {
	ss := strings.TrimSpace(strings.Split(s, ":")[1])

	if ss == "None" || strings.Contains(ss, "Not applicable") {
		return ""
	}
	return ss
}

// nolint:mnd
func parseBytes(s string) (Bytes, error) {
	ss := strings.Fields(s)

	value, err := strconv.ParseFloat(ss[0], 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse value: %w", err)
	}

	switch ss[1] {
	case "KB":
		value *= 1024
	case "MB":
		value *= 1024 * 1024
	case "GB":
		value *= 1024 * 1024 * 1024
	case "TB":
		value *= 1024 * 1024 * 1024 * 1024
	case "Bytes":
		// Do nothing
	default:
		return 0, fmt.Errorf("unknown unit: %s", ss[1])
	}

	return Bytes(value), nil
}
