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

package odatasql

import (
	"context"
	"testing"

	"github.com/CiscoM31/godata"
)

func Test_SelectTree(t *testing.T) {
	basicExpand, _ := godata.ParseExpandString(context.TODO(), "Manufacturer")
	expandWithNestedFilter, _ := godata.ParseExpandString(context.TODO(), "Manufacturer($filter=foo eq 'bar')")
	expandWithNestedSelect, _ := godata.ParseExpandString(context.TODO(), "Manufacturer($select=Name)")
	expandWithNestedSelectAndFilter, _ := godata.ParseExpandString(context.TODO(), "Manufacturer($select=Names($filter=foo eq 'bar'))")

	tests := []struct {
		name        string
		selectQuery *godata.GoDataSelectQuery
		expandQuery *godata.GoDataExpandQuery
		wantErr     bool
	}{
		{
			name:        "Basic Select",
			selectQuery: &godata.GoDataSelectQuery{RawValue: "ModelName,Engine/Options"},
		},
		{
			name:        "Basic Expand",
			expandQuery: basicExpand,
		},
		{
			name:        "Basic Expand and Select",
			selectQuery: &godata.GoDataSelectQuery{RawValue: "ModelName,Engine/Options"},
			expandQuery: basicExpand,
		},
		{
			name: "Select with filter in two places",
			selectQuery: &godata.GoDataSelectQuery{
				RawValue: "Engine/Options($filter=Foo eq 'bar'),Engine($select=Options($filter=Id eq 'id'))",
			},
			expandQuery: nil,
			wantErr:     true,
		},
		{
			name:        "Select with empty string",
			selectQuery: &godata.GoDataSelectQuery{RawValue: ""},
			wantErr:     true,
		},
		{
			name:        "Expand with nested filter",
			expandQuery: expandWithNestedFilter,
		},
		{
			name:        "Expand with nested select",
			expandQuery: expandWithNestedSelect,
		},
		{
			name:        "Expand with nested select and filter",
			expandQuery: expandWithNestedSelectAndFilter,
		},
		{
			name:        "Selection specified for same field in multiple forms",
			selectQuery: &godata.GoDataSelectQuery{RawValue: "Engine/Options,Engine($select=Options)"},
			wantErr:     true,
		},
		{
			name:        "Expand is not allowed inside $select",
			selectQuery: &godata.GoDataSelectQuery{RawValue: "Engine($expand=Foo)"},
			wantErr:     true,
		},
	}

	for _, test := range tests {
		selectTree := newSelectTree()
		err := selectTree.insert(nil, nil, nil, test.selectQuery, test.expandQuery, false)
		if err != nil && !test.wantErr {
			t.Errorf("unexpected error for %v: %v", test.name, err)
		} else if err == nil && test.wantErr {
			t.Errorf("expected error for %v but got nil", test.name)
		}
	}
}
