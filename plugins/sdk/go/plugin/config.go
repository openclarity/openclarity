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
	"os"
)

const (
	EnvLogLevel     = "PLUGIN_SERVER_LOG_LEVEL"
	DefaultLogLevel = "info"

	EnvListenAddress     = "PLUGIN_SERVER_LISTEN_ADDRESS"
	DefaultListenAddress = "http://0.0.0.0:8080"
)

func getLogLevel() string {
	if logLevel := os.Getenv(EnvLogLevel); logLevel != "" {
		return logLevel
	}

	return DefaultLogLevel
}

func getListenAddress() string {
	if listenAddress := os.Getenv(EnvListenAddress); listenAddress != "" {
		return listenAddress
	}

	return DefaultListenAddress
}
