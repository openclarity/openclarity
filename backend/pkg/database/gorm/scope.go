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

package gorm

import (
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/backend/pkg/database/types"
)

const (
	scopesSchemaName = "Scopes"
)

type Scopes struct {
	ODataObject
}

type ScopesTableHandler struct {
	DB *gorm.DB
}

func (db *Handler) ScopesTable() types.ScopesTable {
	return &ScopesTableHandler{
		DB: db.DB,
	}
}

func (s ScopesTableHandler) GetScopes(params models.GetDiscoveryScopesParams) (models.Scopes, error) {
	var dbScopes Scopes
	err := ODataQuery(s.DB, scopesSchemaName, params.Filter, params.Select, nil, nil, nil, nil, false, &dbScopes)
	if err != nil {
		return models.Scopes{}, err
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Scopes{}, types.ErrNotFound
		}
		return models.Scopes{}, err
	}

	var scopes models.Scopes
	err = json.Unmarshal(dbScopes.Data, &scopes)
	if err != nil {
		return models.Scopes{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return scopes, nil
}

func (s ScopesTableHandler) SetScopes(scopes models.Scopes) (models.Scopes, error) {
	marshaled, err := json.Marshal(scopes)
	if err != nil {
		return models.Scopes{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	var dbScopes Scopes
	dbScopes.Data = marshaled

	if err = s.DB.Save(&dbScopes).Error; err != nil {
		return models.Scopes{}, fmt.Errorf("failed to save scopes in db: %w", err)
	}

	// TODO(sambetts) Maybe this isn't required now because the DB isn't
	// creating any of the data (like the ID) so we can just return the
	// target pre-marshal above.
	var apiScopes models.Scopes
	if err = json.Unmarshal(dbScopes.Data, &apiScopes); err != nil {
		return models.Scopes{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return apiScopes, nil
}
