// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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

package main

import (
	"log/slog"
	"time"

	"github.com/openclarity/vmclarity/plugins/sdk-go/plugin"
	"github.com/openclarity/vmclarity/plugins/sdk-go/types"
)

//nolint:containedctx
type Scanner struct {
	status *types.Status
}

func (s *Scanner) Metadata() *types.Metadata {
	return &types.Metadata{
		Name:    types.Ptr("Example scanner"),
		Version: types.Ptr("v0.1.2"),
	}
}

func (s *Scanner) Start(config types.Config) {
	logger := plugin.GetLogger()

	logger.Info("Starting scanner with config", slog.Any("config", config))

	go func(config types.Config) {
		// Mark scan started
		logger.Info("Scanner is running...")
		s.SetStatus(types.NewScannerStatus(types.StateRunning, types.Ptr("Scanner is running...")))

		// Example scanning
		time.Sleep(5 * time.Second) //nolint:mnd

		result := types.Result{
			Vmclarity: types.VMClarityData{
				Vulnerabilities: types.Ptr([]types.Vulnerability{
					{
						VulnerabilityName: types.Ptr("vulnerability #1"),
						Description:       types.Ptr("some vulnerability"),
					},
				}),
			},
		}
		if err := result.Export(config.OutputFile); err != nil {
			logger.Error("Failed to save result to output file")
			s.SetStatus(types.NewScannerStatus(types.StateFailed, types.Ptr("Scanner failed saving result.")))
			return
		}

		// Mark scan done
		logger.Info("Scanner finished running.")
		s.SetStatus(types.NewScannerStatus(types.StateDone, types.Ptr("Scanner finished running.")))
	}(config)
}

func (s *Scanner) Stop(stop types.Stop) {
	// cleanup logic
}

func (s *Scanner) GetStatus() *types.Status {
	return s.status
}

func (s *Scanner) SetStatus(newStatus *types.Status) {
	s.status = types.NewScannerStatus(newStatus.State, newStatus.Message)
}

func main() {
	plugin.Run(&Scanner{
		status: types.NewScannerStatus(types.StateReady, types.Ptr("Scanner ready")),
	})
}
