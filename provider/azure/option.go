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

package azure

type Option func(*Provider)

type ScannerTaskOption Option

func WithEnsureVMInfo(deps ...string) ScannerTaskOption {
	return func(s *Provider) {
		s.EnsureAssetVMInfo(deps)
	}
}

func WithEnsureSnapshot(deps ...string) ScannerTaskOption {
	return func(s *Provider) {
		s.EnsureSnapshotWithCleanup(deps)
	}
}

func WithEnsureDisk(deps ...string) ScannerTaskOption {
	return func(s *Provider) {
		s.EnsureDiskWithCleanup(deps)
	}
}

func WithEnsureScannerVM(deps ...string) ScannerTaskOption {
	return func(s *Provider) {
		s.EnsureScannerVMWithCleanup(deps)
	}
}

func WithEnsureAttachDiskToScannerVM(deps ...string) ScannerTaskOption {
	return func(s *Provider) {
		s.EnsureAttachDiskToScannerVM(deps)
	}
}
