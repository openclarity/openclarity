// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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

package types

import (
	"context"
)

type Job struct {
	Instance    Instance
	SrcSnapshot Snapshot
	DstSnapshot Snapshot
}

type ScanJobRunConfig struct {
	InstanceToScan Instance
	Region         string
	ImageID        string
	DeviceName     string
	SubnetID       string
}

type Instance interface {
	GetID() string
	GetLocation() string
	GetRootVolume(ctx context.Context) (Volume, error)
	WaitForReady(ctx context.Context) error
	Delete(ctx context.Context) error
}

type TargetInstance struct {
	TargetID string
	Instance Instance
}

type Volume interface {
	TakeSnapshot(ctx context.Context) (Snapshot, error)
}

type Snapshot interface {
	GetID() string
	GetRegion() string
	Copy(ctx context.Context, dstRegion string) (Snapshot, error)
	Delete(ctx context.Context) error
	WaitForReady(ctx context.Context) error
}

type ScannerJobConfig struct {
	DirectoryToScan      string               `json:"directory_to_scan"`
	ServerToReport       string               `json:"server_to_report"`
	SbomScan             SbomScan             `json:"sbom_scan,omitempty"`
	VulnerabilityScan    VulnerabilityScan    `json:"vulnerability_scan,omitempty"`
	RootkitScan          RootkitScan          `json:"rootkit_scan,omitempty"`
	MisconfigurationScan MisconfigurationScan `json:"misconfiguration_scan,omitempty"`
	SecretScan           SecretScan           `json:"secret_scan,omitempty"`
	MalwareScan          MalwareScan          `json:"malware_scan,omitempty"`
	ExploitScan          ExploitScan          `json:"exploit_scan,omitempty"`
}

type SbomScan struct {
	Enabled bool `json:"enabled,omitempty"`
}

type VulnerabilityScan struct {
	Enabled bool `json:"enabled,omitempty"`
}

type RootkitScan struct {
	Enabled bool `json:"enabled,omitempty"`
}

type MisconfigurationScan struct {
	Enabled bool `json:"enabled,omitempty"`
}

type SecretScan struct {
	Enabled bool `json:"enabled,omitempty"`
}

type MalwareScan struct {
	Enabled bool `json:"enabled,omitempty"`
}

type ExploitScan struct {
	Enabled bool `json:"enabled,omitempty"`
}
