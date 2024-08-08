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
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/openclarity/vmclarity/plugins/sdk-go/types"
)

// Run starts Plugin HTTP Server and uses provided Scanner to respond to
// requests. Run logs data to standard output using GetLogger. This operation
// blocks until exit. It handles graceful termination. Server listens on address
// loaded from EnvListenAddress. It exists with exit code 1 on error. Run
// simplifies the main logic. Can only be called once.
//
// Usage example:
//
// import (
//
//	"github.com/openclarity/vmclarity/plugins/sdk-go/plugin"
//	"github.com/openclarity/vmclarity/plugins/sdk-go/types"
//	)
//
//	func main() {
//	     var myScanner types.Scanner    // your implementation of Scanner interface
//	     plugin.Run(myScanner)          // start server until exit
//	}
func Run(scanner types.Scanner) {
	// Start server
	server, err := newServer(scanner)
	if err != nil {
		logger.Error("failed to create HTTP server", slog.Any("error", err))
		os.Exit(1)
	}

	go func() {
		logger.Info("Plugin HTTP server starting...")

		listenAddress := getListenAddress()
		listenAddress = strings.TrimPrefix(listenAddress, "http://")
		listenAddress = strings.TrimPrefix(listenAddress, "https://")

		if err = server.Start(listenAddress); err != nil {
			logger.Error("failed to start HTTP server", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	defer func() {
		logger.Info("Plugin HTTP server stopping...")
		if err = server.Stop(); err != nil {
			logger.Error("failed to stop HTTP server", slog.Any("error", err))
			return
		}
		logger.Info("Plugin HTTP server stopped")
	}()

	// Wait until termination
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	s := <-sig
	logger.Warn(fmt.Sprintf("Received a termination signal: %v", s))
}
