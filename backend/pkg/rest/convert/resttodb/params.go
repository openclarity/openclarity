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

package resttodb

import (
	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/backend/pkg/database"
)

func ConvertGetTargetsParams(params models.GetTargetsParams) database.GetTargetsParams {
	return database.GetTargetsParams{
		Filter:   params.Filter,
		Page:     params.Page,
		PageSize: params.PageSize,
	}
}

func ConvertGetScanResultsParams(params models.GetScanResultsParams) database.GetScanResultsParams {
	return database.GetScanResultsParams{
		Filter:   params.Filter,
		Select:   params.Select,
		Page:     params.Page,
		PageSize: params.PageSize,
	}
}

func ConvertGetScanResultsScanResultIDParams(params models.GetScanResultsScanResultIDParams) database.GetScanResultsScanResultIDParams {
	return database.GetScanResultsScanResultIDParams{
		Select: params.Select,
	}
}

func ConvertGetScansParams(params models.GetScansParams) database.GetScansParams {
	return database.GetScansParams{
		Filter:   params.Filter,
		Page:     params.Page,
		PageSize: params.PageSize,
	}
}

func ConvertGetScanConfigsParams(params models.GetScanConfigsParams) database.GetScanConfigsParams {
	return database.GetScanConfigsParams{
		Filter:   params.Filter,
		Page:     params.Page,
		PageSize: params.PageSize,
	}
}
