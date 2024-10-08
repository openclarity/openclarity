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

package plugin

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

func init() {
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(GetLogLevel())); err != nil {
		logLevel = DefaultLogLevel
	}

	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
}

// GetLogger defines JSON logger that outputs to stdout with level loaded fom
// EnvLogLevel.
func GetLogger() *slog.Logger {
	return logger
}
