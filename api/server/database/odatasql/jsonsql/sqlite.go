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

type sqlite struct{}

var SQLite Variant = sqlite{}

func (sqlite) JSONObject(parts []string) string {
	return fmt.Sprintf("JSON_OBJECT(%s)", strings.Join(parts, ", "))
}

func (sqlite) JSONArrayAggregate(value string) string {
	return fmt.Sprintf("JSON_GROUP_ARRAY(%s)", value)
}

func (sqlite) CastToDateTime(strTime string) string {
	return fmt.Sprintf("datetime(%s)", strTime)
}

func (sqlite) JSONEach(source string) string {
	return fmt.Sprintf("JSON_EACH(%s)", source)
}

func (sqlite) JSONArray(items []string) string {
	return fmt.Sprintf("JSON_ARRAY(%s)", strings.Join(items, ", "))
}

func (sqlite) JSONExtract(source, path string) string {
	return fmt.Sprintf("%s->'%s'", source, path)
}

func (sqlite) JSONExtractText(source, path string) string {
	return fmt.Sprintf("%s->>'%s'", source, path)
}

func (sqlite) JSONQuote(value string) string {
	return fmt.Sprintf("JSON_QUOTE(%s)", value)
}

func (sqlite) JSONCast(value string) string {
	return fmt.Sprintf("JSON(%s)", value)
}

func (sqlite) JSONArrayLength(value string) string {
	return fmt.Sprintf("JSON_ARRAY_LENGTH(%s)", value)
}
