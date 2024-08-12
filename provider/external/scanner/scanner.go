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
	"fmt"
	"time"

	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/provider/external/utils"
	provider_service "github.com/openclarity/vmclarity/provider/external/utils/proto"
)

var _ provider.Scanner = &Scanner{}

type Scanner struct {
	Client provider_service.ProviderClient
}

func (s *Scanner) RunAssetScan(ctx context.Context, t *provider.ScanJobConfig) error {
	scanJobConfig, err := utils.ConvertScanJobConfig(t)
	if err != nil {
		return fmt.Errorf("failed to convert scan job config: %w", err)
	}

	res, err := s.Client.RunAssetScan(ctx, &provider_service.RunAssetScanParams{
		ScanJobConfig: scanJobConfig,
	})
	if err != nil {
		return provider.FatalErrorf("failed to run asset scan: %v", err)
	}

	if res.Err == nil {
		return provider.FatalErrorf("failed to run asset scan: an error type must be set")
	}

	switch res.Err.ErrorType.(type) {
	case *provider_service.Error_ErrNone:
		return nil
	case *provider_service.Error_ErrRetry:
		retryableErr := res.GetErr().GetErrRetry()
		return provider.RetryableErrorf(time.Second*time.Duration(retryableErr.After), retryableErr.Err)
	case *provider_service.Error_ErrFatal:
		fatalErr := res.GetErr().GetErrFatal()
		return provider.FatalErrorf("failed to run asset scan: %v", fatalErr.Err)
	default:
		return provider.FatalErrorf("failed to run asset scan: error type is not supported: %t", res.Err.GetErrorType())
	}
}

func (s *Scanner) RemoveAssetScan(ctx context.Context, t *provider.ScanJobConfig) error {
	scanJobConfig, err := utils.ConvertScanJobConfig(t)
	if err != nil {
		return fmt.Errorf("failed to convert scan job config: %w", err)
	}

	_, err = s.Client.RemoveAssetScan(ctx, &provider_service.RemoveAssetScanParams{
		ScanJobConfig: scanJobConfig,
	})
	if err != nil {
		return fmt.Errorf("failed to remove asset scan: %w", err)
	}
	return nil
}
