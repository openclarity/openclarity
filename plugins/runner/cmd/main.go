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

package main

import (
	"context"
	"fmt"

	"github.com/openclarity/vmclarity/plugins/runner"
)

// Test start scanner function.
func main() {
	ctx := context.Background()

	config := runner.PluginConfig{
		Name:          "",
		ImageName:     "", // TODO Add image name
		InputDir:      "", // TODO Add input directory
		OutputFile:    "", // TODO Add output file
		ScannerConfig: "",
	}
	runner, err := runner.New(config)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Starting plugin runner\n")
	cleanup, err := runner.Start(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cleanup(ctx) //nolint:errcheck

	fmt.Printf("Waiting for plugin %s to be ready\n", config.Name)
	err = runner.WaitReady(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Running plugin %s\n", config.Name)
	err = runner.Run(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Waiting for plugin %s to finish\n", config.Name)
	err = runner.WaitDone(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
}
