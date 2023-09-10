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

package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/pkg/shared/families/rootkits/types"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

// nolint:nonamedreturns
func SplitFuncSeparator(sep string) func(data []byte, atEOF bool) (advance int, token []byte, err error) {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		strData := string(data)

		sepIndex := strings.Index(strData[1:], sep)

		if sepIndex != -1 {
			return sepIndex + 1, data[:sepIndex+1], nil
		}

		if atEOF {
			return len(data), data, nil
		}

		return 0, nil, nil
	}
}

// This list comes from the chkrootkit source code.
// https://github.com/Magentron/chkrootkit/blob/master/chkrootkit#L28-L33
var applicationRootkits = []string{
	"amd", "basename", "biff", "chfn", "chsh", "cron", "crontab", "date", "du", "dirname",
	"echo", "egrep", "env", "find", "fingerd", "gpm", "grep", "hdparm", "su", "ifconfig",
	"inetd", "inetdconf", "identd", "init", "killall", "", "ldsopreload", "login", "ls",
	"lsof", "mail", "mingetty", "netstat", "named", "passwd", "pidof", "pop2", "pop3",
	"ps", "pstree", "rpcinfo", "rlogind", "rshd", "slogin", "sendmail", "sshd", "syslogd",
	"tar", "tcpd", "tcpdump", "top", "telnetd", "timed", "traceroute", "vdir", "w", "write",
}

type Rootkit struct {
	RkType   types.RootkitType
	RkName   string
	Message  string
	Infected bool
}

func ParseChkrootkitOutput(chkrootkitOutput []byte) ([]Rootkit, error) {
	var rootkits []Rootkit

	checkingPrefix := "Checking `"
	checkingPrefixLen := len(checkingPrefix)

	outputScanner := bufio.NewScanner(bytes.NewBuffer(chkrootkitOutput))
	outputScanner.Split(SplitFuncSeparator("Checking"))
	for outputScanner.Scan() {
		line := outputScanner.Text()

		if strings.HasPrefix(line, "ROOTDIR is") {
			// Skipping root dir path message.
			continue
		}

		if !strings.HasPrefix(line, checkingPrefix) {
			// Probably should error.
			log.Warnf("Missing 'Checking' prefix, skipping line %q", line)
			continue
		}

		// Removing checking prefix.
		line = line[checkingPrefixLen:]

		// Splitting test name and result
		testName, result, found := strings.Cut(line, "'... ")
		if !found {
			// Probably should error.
			log.Warnf("Failed to found test name and result, skipping line %q", line)
			continue
		}
		result = strings.TrimSpace(result)

		if utils.Contains(applicationRootkits, testName) {
			rootkits = append(rootkits, Rootkit{
				RkType:   types.APPLICATION,
				RkName:   "UNKNOWN",
				Message:  fmt.Sprintf("Application %q %s", testName, result),
				Infected: result == "INFECTED",
			})
		} else if testName == "aliens" {
			aliensToRootkits, err := processAliensToRootkits(result)
			if err != nil {
				return nil, fmt.Errorf("failed to process aliens to rootkits: %w", err)
			}
			rootkits = append(rootkits, aliensToRootkits...)
		} else {
			// Probably should error.
			log.Warnf("Unknown test name %q, skipping line %q", testName, line)
			continue
		}
	}

	if err := outputScanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan the output: %w", err)
	}

	return rootkits, nil
}

func processAliensToRootkits(aliensResult string) ([]Rootkit, error) {
	outputScanner := bufio.NewScanner(bytes.NewBufferString(aliensResult))
	outputScanner.Split(SplitFuncSeparator("Searching"))

	rootkits := map[string]Rootkit{}
	searchingForPrefix := "Searching for "
	searchingForPrefixLen := len(searchingForPrefix)

	for outputScanner.Scan() {
		line := outputScanner.Text()

		if !strings.HasPrefix(line, searchingForPrefix) {
			// Probably should error.
			log.Warnf("Missing 'Searching for' prefix, skipping line %q", line)
			continue
		}
		line = line[searchingForPrefixLen:]

		name, result, found := strings.Cut(line, "...")
		if !found {
			// Probably should error.
			log.Warnf("Failed to found test name and result, skipping line %q", line)
			continue
		}

		name = strings.TrimSuffix(strings.TrimSpace(name), "default files")
		name = strings.TrimSuffix(strings.TrimSpace(name), "default dir")
		name = strings.TrimSuffix(strings.TrimSpace(name), "default files and dirs")
		name = strings.TrimSuffix(strings.TrimSpace(name), "default files and dir")
		name = strings.TrimSuffix(strings.TrimSpace(name), "files and dirs")
		name = strings.TrimSuffix(strings.TrimSpace(name), "modules")
		name = strings.TrimSuffix(strings.TrimSpace(name), "defaults")
		name = strings.TrimSuffix(strings.TrimSpace(name), ", it may take a while")
		name = strings.TrimSuffix(strings.TrimSpace(name), "logs")
		name = strings.TrimSuffix(strings.TrimSpace(name), "'s")
		name = strings.TrimSpace(name)

		result = strings.TrimSpace(result)

		rkType := types.UNKNOWN
		if strings.Contains(strings.ToLower(name), "lkm") {
			rkType = types.KERNEL
		}

		infected := result != "nothing found" && result != "not tested"

		var rk Rootkit
		var ok bool
		rk, ok = rootkits[name]
		if !ok {
			// Create rootkit info.
			rk = Rootkit{
				RkType:   rkType,
				RkName:   name,
				Message:  result,
				Infected: infected,
			}
		} else {
			if !rk.Infected {
				// Update existing rootkit, if previously found as not infected.
				rk.Infected = infected
			}

			// Append the result to the message
			rk.Message = fmt.Sprintf("%s %s", rk.Message, result)
		}

		rootkits[name] = rk
	}

	if err := outputScanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to get output: %w", err)
	}

	return utils.StringKeyMapToArray(rootkits), nil
}
