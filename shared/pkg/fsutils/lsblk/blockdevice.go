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

package lsblk

import (
	"fmt"
)

type BlockDevice struct {
	Name                   string `json:"name,omitempty" mapstructure:"NAME"`             // device name
	KernelName             string `json:"kname,omitempty" mapstructure:"KNAME"`           // internal kernel device name
	Path                   string `json:"path,omitempty" mapstructure:"PATH"`             // path to the device node
	MajorMinor             string `json:"maj:min,omitempty" mapstructure:"MAJ:MIN"`       // major:minor device number
	FSAvail                Bytes  `json:"fsavail,omitempty" mapstructure:"FSAVAIL"`       // filesystem size available
	FSSize                 Bytes  `json:"fssize,omitempty" mapstructure:"FSSIZE"`         // filesystem size
	FSType                 string `json:"fstype,omitempty" mapstructure:"FSTYPE"`         // filesystem type
	FSUsed                 Bytes  `json:"fsused,omitempty" mapstructure:"FSUSED"`         // filesystem size used
	FSUsedPercent          string `json:"fsuse%,omitempty" mapstructure:"FSUSE%"`         // filesystem use percentage
	MountPoint             string `json:"mountpoint,omitempty" mapstructure:"MOUNTPOINT"` // where the device is mounted
	Label                  string `json:"label,omitempty" mapstructure:"LABEL"`           // filesystem LABEL
	UUID                   string `json:"uuid,omitempty" mapstructure:"UUID"`             // filesystem UUID
	PartitionTableUUID     string `json:"ptuuid,omitempty" mapstructure:"PTUUID"`         // partition table identifier (usually UUID)
	PartitionTableType     string `json:"pttype,omitempty" mapstructure:"PTTYPE"`         // partition table type
	PartitionType          string `json:"parttype,omitempty" mapstructure:"PARTTYPE"`     // partition type UUID
	PartitionLabel         string `json:"partlabel,omitempty" mapstructure:"PARTLABEL"`   // partition LABEL
	PartitionUUID          string `json:"partuuid,omitempty" mapstructure:"PARTUUID"`     // partition UUID
	PartitionFlags         string `json:"partflags,omitempty" mapstructure:"PARTFLAGS"`   // partition flags
	ReadAhead              Bytes  `json:"ra,omitempty" mapstructure:"RA"`                 // read-ahead of the device
	ReadOnly               bool   `json:"ro,omitempty" mapstructure:"RO"`                 // read-only device
	Removable              bool   `json:"rm,omitempty" mapstructure:"RM"`                 // removable device
	HotPlug                bool   `json:"hotplug,omitempty" mapstructure:"HOTPLUG"`       // removable or hotplug device (usb, pcmcia, ...)
	Model                  string `json:"model,omitempty" mapstructure:"MODEL"`           // device identifier
	Serial                 string `json:"serial,omitempty" mapstructure:"SERIAL"`         // disk serial number
	Size                   Bytes  `json:"size,omitempty" mapstructure:"SIZE"`             // size of the device
	State                  string `json:"state,omitempty" mapstructure:"STATE"`           // state of the device
	Owner                  string `json:"owner,omitempty" mapstructure:"OWNER"`           // user name
	Group                  string `json:"group,omitempty" mapstructure:"GROUP"`           // group name
	Mode                   string `json:"mode,omitempty" mapstructure:"MODE"`             // device node permissions
	AlignmentOffset        Bytes  `json:"alignment,omitempty" mapstructure:"ALIGNMENT"`   // alignment offset
	MinIO                  Bytes  `json:"min-io,omitempty" mapstructure:"MIN-IO"`         // minimum I/O size
	OptimalIO              Bytes  `json:"opt-io,omitempty" mapstructure:"OPT-IO"`         // optimal I/O size
	PhysicalSectorSize     Bytes  `json:"phy-sec,omitempty" mapstructure:"PHY-SEC"`       // physical sector size
	LogSecSectorSize       Bytes  `json:"log-sec,omitempty" mapstructure:"LOG-SEC"`       // logical sector size
	Rotational             bool   `json:"rota,omitempty" mapstructure:"ROTA"`             // rotational device
	Scheduler              string `json:"sched,omitempty" mapstructure:"SCHED"`           // I/O scheduler name
	RequestQueueSize       Bytes  `json:"rq-size,omitempty" mapstructure:"RQ-SIZE"`       // request queue size
	Type                   string `json:"type,omitempty" mapstructure:"TYPE"`             // device type
	DiscardAlignmentOffset Bytes  `json:"disc-aln,omitempty" mapstructure:"DISC-ALN"`     // discard alignment offset
	DiscardGranularity     Bytes  `json:"disc-gran,omitempty" mapstructure:"DISC-GRAN"`   // discard granularity
	DiscardMaxBytes        Bytes  `json:"disc-max,omitempty" mapstructure:"DISC-MAX"`     // discard max bytes
	DiscardZeroesData      bool   `json:"disc-zero,omitempty" mapstructure:"DISC-ZERO"`   // discard zeroes data
	WriteSameMaxBytes      Bytes  `json:"wsame,omitempty" mapstructure:"WSAME"`           // write same max bytes
	WWN                    string `json:"wwn,omitempty" mapstructure:"WWN"`               // unique storage identifier
	Randomness             bool   `json:"rand,omitempty" mapstructure:"RAND"`             // adds randomness
	ParentKernelName       string `json:"pkname,omitempty" mapstructure:"PKNAME"`         // internal parent kernel device name
	HostChannelTargetLun   string `json:"hctl,omitempty" mapstructure:"HCTL"`             // Host:Channel:Target:Lun for SCSI
	TransportType          string `json:"tran,omitempty" mapstructure:"TRAN"`             // device transport type
	Subsystems             string `json:"subsystems,omitempty" mapstructure:"SUBSYSTEMS"` // de-duplicated chain of subsystems
	Revision               string `json:"rev,omitempty" mapstructure:"REV"`               // device revision
	Vendor                 string `json:"vendor,omitempty" mapstructure:"VENDOR"`         // device vendor
	Zoned                  string `json:"zoned,omitempty" mapstructure:"ZONED"`           // zone model
}

func (b BlockDevice) String() string {
	return fmt.Sprintf("Device: %s Filesystem: %s MountPoint=%s", b.Path, b.FSType, b.MountPoint)
}
