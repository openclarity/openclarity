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

package database

import (
	"fmt"
	"sync"

	"github.com/openclarity/vmclarity/pkg/apiserver/database/gorm"
	"github.com/openclarity/vmclarity/pkg/apiserver/database/types"
)

type DBDriver func(config types.DBConfig) (types.Database, error)

var (
	DBDrivers           map[string]DBDriver
	RegisterDriversOnce sync.Once
)

func RegisterDrivers() {
	RegisterDriversOnce.Do(func() {
		// If DBDrivers is initialised before this function is called don't
		// reset it, this is useful for testing.
		if DBDrivers == nil {
			DBDrivers = map[string]DBDriver{}
		}
		DBDrivers[types.DBDriverTypeLocal] = gorm.NewDatabase
		DBDrivers[types.DBDriverTypePostgres] = gorm.NewDatabase
	})
}

func InitializeDatabase(config types.DBConfig) (types.Database, error) {
	RegisterDrivers()

	if driver, ok := DBDrivers[config.DriverType]; ok {
		return driver(config)
	}
	return nil, fmt.Errorf("unknown DB driver %s", config.DriverType)
}
