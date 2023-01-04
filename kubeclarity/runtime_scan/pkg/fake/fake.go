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

package fake

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	runtime_scan_models "github.com/openclarity/kubeclarity/runtime_scan/api/server/models"
	_config "github.com/openclarity/kubeclarity/runtime_scan/pkg/config"
	"github.com/openclarity/kubeclarity/runtime_scan/pkg/types"
)

const imagesToScan = 1

type Orchestrator struct {
	scanProgress *types.ScanProgress
	killSwitch   chan struct{}
	sync.Mutex
}

func Create() *Orchestrator {
	return &Orchestrator{
		scanProgress: &types.ScanProgress{},
		killSwitch:   make(chan struct{}),
	}
}

func (o *Orchestrator) Start(_ chan struct{}) {
}

func (o *Orchestrator) Scan(_ *_config.ScanConfig) (chan struct{}, error) {
	ks := make(chan struct{})
	scanProgress := &types.ScanProgress{}
	o.Lock()
	if o.killSwitch != nil {
		return nil, fmt.Errorf("found existing killSwitch which most likely means there is an existing scan, call Clear to abort this scan before calling Scan again")
	}
	o.scanProgress = scanProgress
	o.killSwitch = ks
	o.Unlock()

	// Create channel for caller to use to track doneness
	done := make(chan struct{})
	go func() {
		log.Infof("Start scanning %v images", imagesToScan)

		scanProgress.ImagesToScan = imagesToScan

		ticker := time.NewTicker(30 * time.Second) // nolint:gomnd
		defer ticker.Stop()

		for {
			scanProgress.ImagesStartedToScan++
			log.Infof("Started image #%v", scanProgress.ImagesStartedToScan)
			select {
			case <-ks:
				close(done)
				return
			case <-ticker.C:
				scanProgress.ImagesCompletedToScan++
				log.Infof("Completed image #%v", scanProgress.ImagesCompletedToScan)
				if scanProgress.ImagesCompletedToScan == scanProgress.ImagesToScan {
					close(done)
					return
				}
			}
		}
	}()
	return done, nil
}

func (o *Orchestrator) ScanProgress() types.ScanProgress {
	o.Lock()
	defer o.Unlock()

	return *o.scanProgress
}

func (o *Orchestrator) Results() *types.ScanResults {
	o.Lock()
	defer o.Unlock()

	return &types.ScanResults{
		ImageScanResults: createFakeImageScanResults(),
		Progress:         *o.scanProgress,
	}
}

func (o *Orchestrator) Clear() {
	o.Lock()
	defer o.Unlock()

	if o.killSwitch != nil {
		close(o.killSwitch)
		o.killSwitch = nil
	}
}

func (o *Orchestrator) Stop() {
}

// nolint:gomnd
func createFakeImageScanResults() []*types.ImageScanResult {
	var results []*types.ImageScanResult

	for i := 0; i < imagesToScan; i++ {
		results = append(results, &types.ImageScanResult{
			PodName:       fmt.Sprintf("pod-%v", i),
			PodNamespace:  fmt.Sprintf("namespace-%v", i),
			PodUID:        fmt.Sprintf("pod-id-%v", i),
			ContainerName: fmt.Sprintf("container-%v", i),
			ImageName:     fmt.Sprintf("image-name-%v", i),
			ImageHash:     fmt.Sprintf("image-hash-%v", i),
			Vulnerabilities: []*runtime_scan_models.PackageVulnerabilityScan{
				{
					Cvss: &runtime_scan_models.CVSS{
						CvssV3Metrics: &runtime_scan_models.CVSSV3Metrics{
							BaseScore:           99.8,
							ExploitabilityScore: 99.8,
							ImpactScore:         99.8,
						},
						CvssV3Vector: &runtime_scan_models.CVSSV3Vector{
							AttackComplexity:   runtime_scan_models.AttackComplexityHIGH,
							AttackVector:       runtime_scan_models.AttackVectorNETWORK,
							Availability:       runtime_scan_models.AvailabilityHIGH,
							Confidentiality:    runtime_scan_models.ConfidentialityHIGH,
							Integrity:          runtime_scan_models.IntegrityHIGH,
							PrivilegesRequired: runtime_scan_models.PrivilegesRequiredHIGH,
							Scope:              runtime_scan_models.ScopeUNCHANGED,
							UserInteraction:    runtime_scan_models.UserInteractionNONE,
							Vector:             "CVSS:3.1/AV:N/AC:H/PR:H/UI:N/S:U/C:H/I:H/A:H",
						},
					},
					Description: fmt.Sprintf("description-%v", i),
					FixVersion:  fmt.Sprintf("fix-version-%v", i),
					Links:       []string{"link1"},
					Package: &runtime_scan_models.PackageInfo{
						Language: "go",
						License:  "MIT",
						Name:     fmt.Sprintf("pkg-name-%v", i),
						Version:  fmt.Sprintf("pkg-version-%v", i),
					},
					Scanners:          []string{"scanner1, scanner2"},
					Severity:          runtime_scan_models.VulnerabilitySeverityHIGH,
					VulnerabilityName: fmt.Sprintf("CVE-%v", i),
				},
				{
					Cvss: &runtime_scan_models.CVSS{
						CvssV3Metrics: &runtime_scan_models.CVSSV3Metrics{
							BaseScore:           99.8,
							ExploitabilityScore: 99.8,
							ImpactScore:         99.8,
						},
						CvssV3Vector: &runtime_scan_models.CVSSV3Vector{
							AttackComplexity:   runtime_scan_models.AttackComplexityHIGH,
							AttackVector:       runtime_scan_models.AttackVectorNETWORK,
							Availability:       runtime_scan_models.AvailabilityHIGH,
							Confidentiality:    runtime_scan_models.ConfidentialityHIGH,
							Integrity:          runtime_scan_models.IntegrityHIGH,
							PrivilegesRequired: runtime_scan_models.PrivilegesRequiredHIGH,
							Scope:              runtime_scan_models.ScopeUNCHANGED,
							UserInteraction:    runtime_scan_models.UserInteractionNONE,
							Vector:             "CVSS:3.1/AV:N/AC:H/PR:H/UI:N/S:U/C:H/I:H/A:H",
						},
					},
					Description: fmt.Sprintf("description-%v", i+1),
					FixVersion:  fmt.Sprintf("fix-version-%v", i+1),
					Links:       []string{"link1, link2"},
					Package: &runtime_scan_models.PackageInfo{
						Language: "go",
						License:  "MIT",
						Name:     fmt.Sprintf("pkg-name-%v", i+1),
						Version:  fmt.Sprintf("pkg-version-%v", i+1),
					},
					Scanners:          []string{"scanner2"},
					Severity:          runtime_scan_models.VulnerabilitySeverityHIGH,
					VulnerabilityName: fmt.Sprintf("CVE-%v", i),
				},
			},
			Success:    true,
			ScanErrors: nil,
		})
	}

	return results
}
