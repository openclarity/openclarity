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
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/openclarity/vmclarity/plugins/runner"
)

// Test start scanner function.
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create plugin runner
	fmt.Printf("Starting plugin runner\n")
	config := LoadConfig()
	runner, err := runner.New(ctx, config)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer runner.Remove(ctx) //nolint:errcheck

	// Start plugin container
	fmt.Printf("Starting scanner plugin...\n")
	if err := runner.Start(ctx); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Waiting for plugin %s to be ready\n", config.Name)
	err = runner.WaitReady(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Stream logs
	go func() {
		logs, err := runner.Logs(ctx)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer logs.Close()

		for r := bufio.NewScanner(logs); r.Scan(); {
			fmt.Println("scanner log line: ", r.Text())
		}
	}()

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

	// Print result
	result, _ := runner.Result(ctx)
	defer result.Close()

	bytes, _ := io.ReadAll(result)
	fmt.Println(string(bytes))
}
