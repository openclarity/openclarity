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

package types

import (
	log "github.com/sirupsen/logrus"

	"wwwin-github.cisco.com/eti/scan-gazr/api/server/models"
)

type ResourceType string

const (

	// ResourceTypeIMAGE captures enum value "IMAGE".
	ResourceTypeIMAGE ResourceType = "IMAGE"

	// ResourceTypeDIRECTORY captures enum value "DIRECTORY".
	ResourceTypeDIRECTORY ResourceType = "DIRECTORY"

	// ResourceTypeFILE captures enum value "FILE".
	ResourceTypeFILE ResourceType = "FILE"
)

func ResourceTypeToModels(typ ResourceType) models.ResourceType {
	switch typ {
	case ResourceTypeIMAGE:
		return models.ResourceTypeIMAGE
	case ResourceTypeDIRECTORY:
		return models.ResourceTypeDIRECTORY
	case ResourceTypeFILE:
		return models.ResourceTypeFILE
	default:
		log.Warnf("Unknown resoure type %v", typ)
		return models.ResourceTypeIMAGE
	}
}
