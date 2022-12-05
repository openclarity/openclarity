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

package database

import (
	"fmt"

	"github.com/openclarity/vmclarity/api/models"
)

func (ft *FakeTargetsTable) ListTargets(params models.GetTargetsParams) ([]Target, error) {
	targets := make([]Target, 0)
	for _, target := range *ft.targets {
		targets = append(targets, *target)
	}
	return targets, nil
}

func (ft *FakeTargetsTable) GetTarget(targetID models.TargetID) (*Target, error) {
	targets := *ft.targets
	if _, ok := targets[targetID]; !ok {
		return nil, fmt.Errorf("target not exists with ID: %s", targetID)
	}
	return targets[targetID], nil
}

func (ft *FakeTargetsTable) CreateTarget(target *Target) (*Target, error) {
	targets := *ft.targets
	targets[target.ID] = target
	ft.targets = &targets
	return target, nil
}

func (ft *FakeTargetsTable) UpdateTarget(target *Target, targetID models.TargetID) (*Target, error) {
	targets := *ft.targets
	if _, ok := targets[targetID]; !ok {
		return nil, fmt.Errorf("target not exist with ID: %s", targetID)
	}
	targets[targetID] = target
	ft.targets = &targets
	return target, nil
}

func (ft *FakeTargetsTable) DeleteTarget(targetID models.TargetID) error {
	targets := *ft.targets
	if _, ok := targets[targetID]; !ok {
		return fmt.Errorf("target not exists with ID: %s", targetID)
	}
	delete(targets, targetID)
	ft.targets = &targets
	return nil
}
