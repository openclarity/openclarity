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

package scan_manager // nolint:revive,stylecheck

import (
	"context"
	"fmt"

	"github.com/openclarity/vmclarity/scanner/families"

	"github.com/sirupsen/logrus"
)

type CreateScannerFunc[ConfigType, ScannerResultType any] func(context.Context, string, ConfigType) (families.Scanner[ScannerResultType], error)

type Factory[ConfigType, ScannerResultType any] struct {
	scanners map[string]CreateScannerFunc[ConfigType, ScannerResultType]
}

func NewFactory[ConfigType, ScannerResultType any]() *Factory[ConfigType, ScannerResultType] {
	return &Factory[ConfigType, ScannerResultType]{
		scanners: make(map[string]CreateScannerFunc[ConfigType, ScannerResultType]),
	}
}

func (f *Factory[CT, RT]) Register(name string, createScannerFunc CreateScannerFunc[CT, RT]) {
	if f.scanners == nil {
		f.scanners = make(map[string]CreateScannerFunc[CT, RT])
	}

	if _, ok := f.scanners[name]; ok {
		logrus.Fatalf("%q already registered", name)
	}

	f.scanners[name] = createScannerFunc
}

func (f *Factory[CT, RT]) createScanner(ctx context.Context, name string, config CT) (families.Scanner[RT], error) {
	createFunc, ok := f.scanners[name]
	if !ok {
		return nil, fmt.Errorf("%v not a registered scanner", name)
	}

	return createFunc(ctx, name, config)
}
