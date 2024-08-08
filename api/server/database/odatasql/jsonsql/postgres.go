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

package jsonsql

import (
	"fmt"
	"strings"
)

type postgres struct{}

var Postgres Variant = postgres{}

func (postgres) JSONObject(parts []string) string {
	return fmt.Sprintf("JSONB_BUILD_OBJECT(%s)", strings.Join(parts, ", "))
}

func (postgres) JSONArrayAggregate(value string) string {
	return fmt.Sprintf("JSONB_AGG(%s)", value)
}

func (postgres) CastToDateTime(strTime string) string {
	return fmt.Sprintf("(%s)::timestamptz", strTime)
}

func (postgres) CastToInteger(strInt string) string {
	return fmt.Sprintf("(%s)::integer", strInt)
}

func (postgres) JSONEach(source string) string {
	// The postgres function expect the data must be an array, so we need
	// to detect any other types in the SQL statement and switch it to
	// empty array.
	return fmt.Sprintf("JSONB_ARRAY_ELEMENTS(CASE JSONB_TYPEOF(%s) WHEN 'array' THEN (%s) ELSE '[]' END)", source, source)
}

func (postgres) JSONArray(items []string) string {
	return fmt.Sprintf("JSONB_BUILD_ARRAY(%s)", strings.Join(items, ", "))
}

func convertJSONPathToPostgresPath(jsonPath string) string {
	parts := strings.Split(jsonPath, ".")
	newParts := []string{}
	for _, part := range parts {
		if part == "$" {
			continue
		}
		newParts = append(newParts, part)
	}
	return fmt.Sprintf("{%s}", strings.Join(newParts, ","))
}

func (postgres) JSONExtract(source, path string) string {
	path = convertJSONPathToPostgresPath(path)
	return fmt.Sprintf("%s#>'%s'", source, path)
}

func (postgres) JSONExtractText(source, path string) string {
	path = convertJSONPathToPostgresPath(path)
	return fmt.Sprintf("%s#>>'%s'", source, path)
}

func (postgres) JSONQuote(value string) string {
	return fmt.Sprintf("TO_JSONB(%s::text)", value)
}

func (postgres) JSONCast(value string) string {
	return fmt.Sprintf("TO_JSONB(%s)", value)
}

func (postgres) JSONArrayLength(value string) string {
	return fmt.Sprintf("JSONB_ARRAY_LENGTH(%s)", value)
}
