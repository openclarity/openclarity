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

package state

import (
	"context"

	"github.com/openclarity/openclarity/core/log"
	"github.com/openclarity/openclarity/scanner"
	"github.com/openclarity/openclarity/scanner/families"
)

type LocalState struct{}

func (l *LocalState) WaitForReadyState(context.Context) error {
	return nil
}

func (l *LocalState) MarkInProgress(ctx context.Context, _ *scanner.Config) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)
	logger.Info("Scanning is in progress")
	return nil
}

func (l *LocalState) MarkFamilyScanInProgress(ctx context.Context, familyType families.FamilyType) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	switch familyType {
	case families.SBOM:
		logger.Info("SBOM scan is in progress")
	case families.Vulnerabilities:
		logger.Info("Vulnerabilities scan is in progress")
	case families.Secrets:
		logger.Info("Secrets scan is in progress")
	case families.Exploits:
		logger.Info("Exploits scan is in progress")
	case families.Misconfiguration:
		logger.Info("Misconfiguration scan is in progress")
	case families.Rootkits:
		logger.Info("Rootkit scan is in progress")
	case families.Malware:
		logger.Info("Malware scan is in progress")
	case families.InfoFinder:
		logger.Info("InfoFinder scan is in progress")
	case families.Plugins:
		logger.Info("Plugins scan is in progress")
	}
	return nil
}

func (l *LocalState) MarkDone(ctx context.Context) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)
	logger.Info("Scan has been completed")
	return nil
}

func (l *LocalState) MarkFailed(ctx context.Context, errors string) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)
	logger.Errorf("scan has been completed with errors: %v", errors)
	return nil
}

func (l *LocalState) IsAborted(context.Context) (bool, error) {
	return false, nil
}

func NewLocalState() (*LocalState, error) {
	return &LocalState{}, nil
}
