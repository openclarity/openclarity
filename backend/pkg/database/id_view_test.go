// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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
	"testing"

	"github.com/openclarity/kubeclarity/api/server/restapi/operations"
)

func TestIdView(t *testing.T) {
	// prepare database
	db := Init(&DBConfig{
		EnableInfoLogs:            false,
		DriverType:                DBDriverTypeLocal,
		ViewRefreshIntervalSecond: 1,
	})
	db.CreateFakeData()

	// fetch first application
	sortDir := string("ASC")
	apps, _, err := db.ApplicationTable().GetApplicationsAndTotal(GetApplicationsParams{
		GetApplicationsParams: operations.GetApplicationsParams{
			SortDir:  &sortDir,
			SortKey:  "applicationName",
			Page:     1,
			PageSize: 1,
		},
	})
	if err != nil || len(apps) != 1 {
		t.Fatalf("expected one application")
	}

	// set application ID
	appID := apps[0].ID

	// fetch application from view table
	appViewIDs, err := db.IDsView().GetIDs(
		GetIDsParams{
			FilterIDs:    []string{appID},
			FilterIDType: ApplicationIDType,
			LookupIDType: ApplicationIDType,
		},
		true,
	)
	if err != nil || len(appViewIDs) != 1 {
		t.Fatalf("expected one application from IDsView table")
	}
	if appViewIDs[0] != appID {
		t.Fatalf("expected application ID from IDsView table to be same as ID from Applications table")
	}
}
