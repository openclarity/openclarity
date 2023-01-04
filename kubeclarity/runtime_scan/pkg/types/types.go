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
	"k8s.io/apimachinery/pkg/labels"

	"github.com/openclarity/kubeclarity/runtime_scan/api/server/models"
)

type Status string

const (
	Idle            Status = "Idle"
	ScanInit        Status = "ScanInit"
	ScanInitFailure Status = "ScanInitFailure"
	NothingToScan   Status = "NothingToScan"
	Scanning        Status = "Scanning"
	DoneScanning    Status = "DoneScanning"
	ScanAborted     Status = "Aborted"
)

type ScanProgress struct {
	ImagesToScan          uint32
	ImagesStartedToScan   uint32
	ImagesCompletedToScan uint32
	Status                Status
}

func (s *ScanProgress) SetStatus(status Status) {
	s.Status = status
}

type ImageScanResult struct {
	// Pod data
	PodName      string
	PodNamespace string
	PodUID       string
	PodLabels    labels.Set
	// Container data
	ContainerName string
	ImageName     string
	ImageHash     string
	// Scan results
	CISDockerBenchmarkResult []*models.CISDockerBenchmarkCodeInfo
	Vulnerabilities          []*models.PackageVulnerabilityScan
	LayerCommands            []*models.ResourceLayerCommand
	Success                  bool
	ScanErrors               []*ScanError
}

type ScanResults struct {
	ImageScanResults []*ImageScanResult
	Progress         ScanProgress
}
