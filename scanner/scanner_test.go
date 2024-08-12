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
	"errors"
	"testing"
	"time"

	"github.com/openclarity/vmclarity/scanner/common"
	"github.com/openclarity/vmclarity/scanner/families"
	misconfigurationtypes "github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
)

type familyNotifierSpy struct {
	Results []families.FamilyResult
}

func (n *familyNotifierSpy) FamilyStarted(context.Context, families.FamilyType) error {
	return nil
}

func (n *familyNotifierSpy) FamilyFinished(_ context.Context, res families.FamilyResult) error {
	n.Results = append(n.Results, res)

	return nil
}

func TestManagerRun(t *testing.T) {
	manager := New(&Config{
		Misconfiguration: misconfigurationtypes.Config{
			Enabled:      true,
			ScannersList: []string{"fake"},
			Inputs: []common.ScanInput{
				{
					Input:     "./",
					InputType: common.ROOTFS,
				},
			},
		},
	})

	notifier := &familyNotifierSpy{}
	errs := manager.Run(context.Background(), notifier)
	if len(errs) > 0 {
		t.Fatalf("expected manager to run successfully, got %v", errs)
	}

	if len(notifier.Results) != 1 {
		t.Fatalf("expected 1 result")
	}

	result := notifier.Results[0]
	if err := result.Err; err != nil {
		t.Fatalf("expected FamilyResult(%s) nil error, got %v", result.FamilyType, err)
	}

	if _, ok := result.Result.(*misconfigurationtypes.Result); !ok {
		t.Fatalf("expected Misconfigurations results, got %T", result.Result)
	}
}

func TestManagerRunTimeout(t *testing.T) {
	manager := New(&Config{})
	notifier := &familyNotifierSpy{}
	ctx, cancel := context.WithTimeout(context.Background(), -time.Nanosecond)
	defer cancel()

	manager.Run(ctx, notifier)

	if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded error, got %s", ctx.Err())
	}

	for _, res := range notifier.Results {
		if res.Err == nil {
			t.Fatalf("expected FamilyResult(%s) error, got nil", res.FamilyType)
		}
	}
}
