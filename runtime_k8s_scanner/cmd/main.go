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

package main

import (
	"fmt"
	"os"

	logutils "github.com/Portshift/go-utils/log"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/urfave/cli"

	"wwwin-github.cisco.com/eti/scan-gazr/runtime_k8s_scanner/pkg/scanner"
	"wwwin-github.cisco.com/eti/scan-gazr/runtime_k8s_scanner/pkg/version"
	shared "wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/config"
)

func run(c *cli.Context) {
	logutils.InitLogs(c, os.Stdout)
	scanner.Run()
}

func versionPrint(_ *cli.Context) {
	fmt.Printf("Version: %s \nCommit: %s\nBuild Time: %s",
		version.Version, version.CommitHash, version.BuildTimestamp)
}

func main() {
	viper.SetDefault(shared.ResultServiceAddress, "kubeclarity.kubeclarity:8888")
	viper.AutomaticEnv()
	app := cli.NewApp()
	app.Usage = ""
	app.Name = "KubeClarity Scanner Job"
	app.Version = version.Version

	runCommand := cli.Command{
		Name:   "scan",
		Usage:  "Starts KubeClarity Scanner Job",
		Action: run,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  logutils.LogLevelFlag,
				Value: logutils.LogLevelDefaultValue,
				Usage: logutils.LogLevelFlagUsage,
			},
		},
	}
	runCommand.UsageText = runCommand.Name

	versionCommand := cli.Command{
		Name:   "version",
		Usage:  "KubeClarity Scanner Job Version Details",
		Action: versionPrint,
	}
	versionCommand.UsageText = versionCommand.Name

	app.Commands = []cli.Command{
		runCommand,
		versionCommand,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
