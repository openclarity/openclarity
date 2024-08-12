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

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/apiserver/common"
	"github.com/openclarity/vmclarity/pkg/apiserver/database/types"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

const (
	providerSchemaName = "Provider"
)

type Provider struct {
	ODataObject
}

type ProvidersTableHandler struct {
	DB *gorm.DB
}

func (db *Handler) ProvidersTable() types.ProvidersTable {
	return &ProvidersTableHandler{
		DB: db.DB,
	}
}

func (t *ProvidersTableHandler) GetProviders(params models.GetProvidersParams) (models.Providers, error) {
	var providers []Provider
	err := ODataQuery(t.DB, providerSchemaName, params.Filter, params.Select, params.Expand, params.OrderBy, params.Top, params.Skip, true, &providers)
	if err != nil {
		return models.Providers{}, err
	}

	items := make([]models.Provider, len(providers))
	for i, pr := range providers {
		var provider models.Provider
		err = json.Unmarshal(pr.Data, &provider)
		if err != nil {
			return models.Providers{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		items[i] = provider
	}

	output := models.Providers{Items: &items}

	if params.Count != nil && *params.Count {
		count, err := ODataCount(t.DB, providerSchemaName, params.Filter)
		if err != nil {
			return models.Providers{}, fmt.Errorf("failed to count records: %w", err)
		}
		output.Count = &count
	}

	return output, nil
}

func (t *ProvidersTableHandler) GetProvider(providerID models.ProviderID, params models.GetProvidersProviderIDParams) (models.Provider, error) {
	var dbProvider Provider
	filter := fmt.Sprintf("id eq '%s'", providerID)
	err := ODataQuery(t.DB, providerSchemaName, &filter, params.Select, params.Expand, nil, nil, nil, false, &dbProvider)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Provider{}, types.ErrNotFound
		}
		return models.Provider{}, err
	}

	var apiProvider models.Provider
	err = json.Unmarshal(dbProvider.Data, &apiProvider)
	if err != nil {
		return models.Provider{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return apiProvider, nil
}

func (t *ProvidersTableHandler) CreateProvider(provider models.Provider) (models.Provider, error) {
	// Check the user didn't provide an ID
	if provider.Id != nil {
		return models.Provider{}, &common.BadRequestError{
			Reason: "can not specify id field when creating a new Provider",
		}
	}

	if provider.Status != nil && (provider.Status.State == "" || provider.Status.Reason == "" || provider.Status.LastTransitionTime.IsZero()) {
		return models.Provider{}, &common.BadRequestError{
			Reason: "status state, status reason and status last transition time are required to create provider",
		}
	}

	// Generate a new UUID
	provider.Id = utils.PointerTo(uuid.New().String())

	// Initialise revision
	provider.Revision = utils.PointerTo(1)

	// TODO(sambetts) Lock the table here to prevent race conditions
	// checking the uniqueness.
	//
	// We might also be able to do this without locking the table by doing
	// a single query which includes the uniqueness check like:
	//
	// INSERT INTO scan_configs(data) SELECT * FROM (SELECT "<encoded json>") AS tmp WHERE NOT EXISTS (SELECT * FROM scan_configs WHERE JSON_EXTRACT(`Data`, '$.Name') = '<name from input>') LIMIT 1;
	//
	// This should return 0 affected fields if there is a conflicting
	// record in the DB, and should be treated safely by the DB without
	// locking the table.

	marshaled, err := json.Marshal(provider)
	if err != nil {
		return models.Provider{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	newProvider := Provider{}
	newProvider.Data = marshaled

	if err = t.DB.Create(&newProvider).Error; err != nil {
		return models.Provider{}, fmt.Errorf("failed to create provider in db: %w", err)
	}

	return provider, nil
}

// nolint:cyclop
func (t *ProvidersTableHandler) SaveProvider(provider models.Provider, params models.PutProvidersProviderIDParams) (models.Provider, error) {
	if provider.Id == nil || *provider.Id == "" {
		return models.Provider{}, &common.BadRequestError{
			Reason: "id is required to save provider",
		}
	}

	if provider.Status.State == "" || provider.Status.Reason == "" || provider.Status.LastTransitionTime.IsZero() {
		return models.Provider{}, &common.BadRequestError{
			Reason: "status state, status reason and status last transition time are required to save provider",
		}
	}

	var dbObj Provider
	if err := getExistingObjByID(t.DB, providerSchemaName, *provider.Id, &dbObj); err != nil {
		return models.Provider{}, fmt.Errorf("failed to get provider from db: %w", err)
	}

	var dbProvider models.Provider
	err := json.Unmarshal(dbObj.Data, &dbProvider)
	if err != nil {
		return models.Provider{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	if err := checkRevisionEtag(params.IfMatch, dbProvider.Revision); err != nil {
		return models.Provider{}, err
	}

	provider.Revision = bumpRevision(dbProvider.Revision)

	marshaled, err := json.Marshal(provider)
	if err != nil {
		return models.Provider{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	dbObj.Data = marshaled

	if err = t.DB.Save(&dbObj).Error; err != nil {
		return models.Provider{}, fmt.Errorf("failed to save provider in db: %w", err)
	}

	return provider, nil
}

// nolint:cyclop
func (t *ProvidersTableHandler) UpdateProvider(provider models.Provider, params models.PatchProvidersProviderIDParams) (models.Provider, error) {
	if provider.Id == nil || *provider.Id == "" {
		return models.Provider{}, fmt.Errorf("ID is required to update provider in DB")
	}

	if provider.Status.State == "" || provider.Status.Reason == "" || provider.Status.LastTransitionTime.IsZero() {
		return models.Provider{}, &common.BadRequestError{
			Reason: "status state, status reason and status last transition time are required to save provider",
		}
	}

	var dbObj Provider
	if err := getExistingObjByID(t.DB, providerSchemaName, *provider.Id, &dbObj); err != nil {
		return models.Provider{}, err
	}

	var err error
	var dbProvider models.Provider
	err = json.Unmarshal(dbObj.Data, &dbProvider)
	if err != nil {
		return models.Provider{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	if err := checkRevisionEtag(params.IfMatch, dbProvider.Revision); err != nil {
		return models.Provider{}, err
	}

	provider.Revision = bumpRevision(dbProvider.Revision)

	dbObj.Data, err = patchObject(dbObj.Data, provider)
	if err != nil {
		return models.Provider{}, fmt.Errorf("failed to apply patch: %w", err)
	}

	var ret models.Provider
	err = json.Unmarshal(dbObj.Data, &ret)
	if err != nil {
		return models.Provider{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	if err := t.DB.Save(&dbObj).Error; err != nil {
		return models.Provider{}, fmt.Errorf("failed to save provider in db: %w", err)
	}

	return ret, nil
}

func (t *ProvidersTableHandler) DeleteProvider(providerID models.ProviderID) error {
	if err := deleteObjByID(t.DB, providerID, &Provider{}); err != nil {
		return fmt.Errorf("failed to delete provider: %w", err)
	}

	return nil
}
