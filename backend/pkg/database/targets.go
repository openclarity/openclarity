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

	"gorm.io/gorm"

	"github.com/openclarity/vmclarity/api/models"
)

// TODO after db design.
type Target struct {
	ID         string
	TargetInfo *models.TargetType
}

type TargetsTable interface {
	ListTargets(params models.GetTargetsParams) ([]Target, error)
	GetTarget(targetID models.TargetID) (*Target, error)
	CreateTarget(target *Target) (*Target, error)
	UpdateTarget(target *Target, targetID models.TargetID) (*Target, error)
	DeleteTarget(targetID models.TargetID) error
}

type TargetsTableHandler struct {
	db *gorm.DB
}

func (db *Handler) TargetsTable() TargetsTable {
	return &TargetsTableHandler{
		db: db.DB,
	}
}

func (t *TargetsTableHandler) ListTargets(params models.GetTargetsParams) ([]Target, error) {
	return []Target{}, fmt.Errorf("not implemented")
}

func (t *TargetsTableHandler) GetTarget(targetID models.TargetID) (*Target, error) {
	return &Target{}, fmt.Errorf("not implemented")
}

func (t *TargetsTableHandler) CreateTarget(target *Target) (*Target, error) {
	return &Target{}, fmt.Errorf("not implemented")
}

func (t *TargetsTableHandler) UpdateTarget(target *Target, targetID models.TargetID) (*Target, error) {
	return &Target{}, fmt.Errorf("not implemented")
}

func (t *TargetsTableHandler) DeleteTarget(targetID models.TargetID) error {
	return fmt.Errorf("not implemented")
}
