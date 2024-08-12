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

package families

import (
	"context"
	"errors"
	"testing"
	"time"

	misconfigurationTypes "github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
	"github.com/openclarity/vmclarity/scanner/families/types"
	"github.com/openclarity/vmclarity/scanner/utils"
)

type familyNotifierSpy struct {
	Results []FamilyResult
}

func (n *familyNotifierSpy) FamilyStarted(context.Context, types.FamilyType) error {
	return nil
}

func (n *familyNotifierSpy) FamilyFinished(_ context.Context, res FamilyResult) error {
	n.Results = append(n.Results, res)

	return nil
}

func TestManagerRunTimeout(t *testing.T) {
	conf := &Config{
		Misconfiguration: misconfigurationTypes.Config{
			Enabled:      true,
			ScannersList: []string{"fake"},
			Inputs: []types.Input{
				{
					Input:     "./",
					InputType: string(utils.ROOTFS),
				},
			},
		},
	}

	manager := New(conf)
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
