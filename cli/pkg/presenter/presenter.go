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

package presenter

import (
	"context"

	"github.com/openclarity/vmclarity/shared/pkg/families"
	"github.com/openclarity/vmclarity/shared/pkg/families/results"
)

type Presenter interface {
	ExportSbomResult(context.Context, *results.Results, families.RunErrors) error
	ExportVulResult(context.Context, *results.Results, families.RunErrors) error
	ExportSecretsResult(context.Context, *results.Results, families.RunErrors) error
	ExportMalwareResult(context.Context, *results.Results, families.RunErrors) error
	ExportExploitsResult(context.Context, *results.Results, families.RunErrors) error
	ExportMisconfigurationResult(context.Context, *results.Results, families.RunErrors) error
	ExportRootkitResult(context.Context, *results.Results, families.RunErrors) error
}
