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

	"github.com/openclarity/vmclarity/api/server/common"
	dbtypes "github.com/openclarity/vmclarity/api/server/database/types"
	"github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
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

func (db *Handler) ProvidersTable() dbtypes.ProvidersTable {
	return &ProvidersTableHandler{
		DB: db.DB,
	}
}

func (t *ProvidersTableHandler) GetProviders(params types.GetProvidersParams) (types.Providers, error) {
	var providers []Provider
	err := ODataQuery(t.DB, providerSchemaName, params.Filter, params.Select, params.Expand, params.OrderBy, params.Top, params.Skip, true, &providers)
	if err != nil {
		return types.Providers{}, err
	}

	items := make([]types.Provider, len(providers))
	for i, pr := range providers {
		var provider types.Provider
		err = json.Unmarshal(pr.Data, &provider)
		if err != nil {
			return types.Providers{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		items[i] = provider
	}

	output := types.Providers{Items: &items}

	if params.Count != nil && *params.Count {
		count, err := ODataCount(t.DB, providerSchemaName, params.Filter)
		if err != nil {
			return types.Providers{}, fmt.Errorf("failed to count records: %w", err)
		}
		output.Count = &count
	}

	return output, nil
}

func (t *ProvidersTableHandler) GetProvider(providerID types.ProviderID, params types.GetProvidersProviderIDParams) (types.Provider, error) {
	var dbProvider Provider
	filter := fmt.Sprintf("id eq '%s'", providerID)
	err := ODataQuery(t.DB, providerSchemaName, &filter, params.Select, params.Expand, nil, nil, nil, false, &dbProvider)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.Provider{}, dbtypes.ErrNotFound
		}
		return types.Provider{}, err
	}

	var apiProvider types.Provider
	err = json.Unmarshal(dbProvider.Data, &apiProvider)
	if err != nil {
		return types.Provider{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return apiProvider, nil
}

func (t *ProvidersTableHandler) CreateProvider(provider types.Provider) (types.Provider, error) {
	// Check the user didn't provide an ID
	if provider.Id != nil {
		return types.Provider{}, &common.BadRequestError{
			Reason: "can not specify id field when creating a new Provider",
		}
	}

	if provider.Status != nil && (provider.Status.State == "" || provider.Status.Reason == "" || provider.Status.LastTransitionTime.IsZero()) {
		return types.Provider{}, &common.BadRequestError{
			Reason: "status state, status reason and status last transition time are required to create provider",
		}
	}

	// Generate a new UUID
	provider.Id = to.Ptr(uuid.New().String())

	// Initialise revision
	provider.Revision = to.Ptr(1)

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
		return types.Provider{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	newProvider := Provider{}
	newProvider.Data = marshaled

	if err = t.DB.Create(&newProvider).Error; err != nil {
		return types.Provider{}, fmt.Errorf("failed to create provider in db: %w", err)
	}

	return provider, nil
}

// nolint:cyclop
func (t *ProvidersTableHandler) SaveProvider(provider types.Provider, params types.PutProvidersProviderIDParams) (types.Provider, error) {
	if provider.Id == nil || *provider.Id == "" {
		return types.Provider{}, &common.BadRequestError{
			Reason: "id is required to save provider",
		}
	}

	if provider.Status.State == "" || provider.Status.Reason == "" || provider.Status.LastTransitionTime.IsZero() {
		return types.Provider{}, &common.BadRequestError{
			Reason: "status state, status reason and status last transition time are required to save provider",
		}
	}

	var dbObj Provider
	if err := getExistingObjByID(t.DB, providerSchemaName, *provider.Id, &dbObj); err != nil {
		return types.Provider{}, fmt.Errorf("failed to get provider from db: %w", err)
	}

	var dbProvider types.Provider
	err := json.Unmarshal(dbObj.Data, &dbProvider)
	if err != nil {
		return types.Provider{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	if err := checkRevisionEtag(params.IfMatch, dbProvider.Revision); err != nil {
		return types.Provider{}, err
	}

	provider.Revision = bumpRevision(dbProvider.Revision)

	marshaled, err := json.Marshal(provider)
	if err != nil {
		return types.Provider{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	dbObj.Data = marshaled

	if err = t.DB.Save(&dbObj).Error; err != nil {
		return types.Provider{}, fmt.Errorf("failed to save provider in db: %w", err)
	}

	return provider, nil
}

// nolint:cyclop
func (t *ProvidersTableHandler) UpdateProvider(provider types.Provider, params types.PatchProvidersProviderIDParams) (types.Provider, error) {
	if provider.Id == nil || *provider.Id == "" {
		return types.Provider{}, errors.New("ID is required to update provider in DB")
	}

	if provider.Status.State == "" || provider.Status.Reason == "" || provider.Status.LastTransitionTime.IsZero() {
		return types.Provider{}, &common.BadRequestError{
			Reason: "status state, status reason and status last transition time are required to save provider",
		}
	}

	var dbObj Provider
	if err := getExistingObjByID(t.DB, providerSchemaName, *provider.Id, &dbObj); err != nil {
		return types.Provider{}, err
	}

	var err error
	var dbProvider types.Provider
	err = json.Unmarshal(dbObj.Data, &dbProvider)
	if err != nil {
		return types.Provider{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	if err := checkRevisionEtag(params.IfMatch, dbProvider.Revision); err != nil {
		return types.Provider{}, err
	}

	provider.Revision = bumpRevision(dbProvider.Revision)

	dbObj.Data, err = patchObject(dbObj.Data, provider)
	if err != nil {
		return types.Provider{}, fmt.Errorf("failed to apply patch: %w", err)
	}

	var ret types.Provider
	err = json.Unmarshal(dbObj.Data, &ret)
	if err != nil {
		return types.Provider{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	if err := t.DB.Save(&dbObj).Error; err != nil {
		return types.Provider{}, fmt.Errorf("failed to save provider in db: %w", err)
	}

	return ret, nil
}

func (t *ProvidersTableHandler) DeleteProvider(providerID types.ProviderID) error {
	if err := deleteObjByID(t.DB, providerID, &Provider{}); err != nil {
		return fmt.Errorf("failed to delete provider: %w", err)
	}

	return nil
}
