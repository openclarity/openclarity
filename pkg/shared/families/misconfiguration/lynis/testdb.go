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

package lynis

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	lynisDBEntryParts            = 6
	lynisDBEntryDescriptionIndex = 5
	lynisDBEntryCategoryIndex    = 2
	lynisDBEntryTestNameIndex    = 0

	unknown = "unknown"
)

type DBTestEntry struct {
	Category    string
	Description string
}

type testdb map[string]DBTestEntry

type TestDB struct {
	tests testdb
}

func NewTestDB(logger *log.Entry, lynisDBDir string) (*TestDB, error) {
	// Comes from the Lynis install:
	// https://github.com/CISOfy/lynis/blob/master/db/tests.db
	lynisTestDBPath := filepath.Join(lynisDBDir, "tests.db")
	if _, err := os.Stat(lynisTestDBPath); err != nil {
		return nil, fmt.Errorf("failed to find DB @ %v: %w", lynisTestDBPath, err)
	}

	tests, err := parseTestsFromDBPath(logger, lynisTestDBPath)
	if err != nil {
		return nil, fmt.Errorf("unable to initialise lynis test DB: %w", err)
	}

	return &TestDB{tests: tests}, nil
}

func (a *TestDB) GetCategoryForTestID(testid string) string {
	if entry, ok := a.tests[testid]; ok {
		return entry.Category
	}
	return unknown
}

func (a *TestDB) GetDescriptionForTestID(testid string) string {
	if entry, ok := a.tests[testid]; ok {
		return entry.Description
	}
	return unknown
}

func parseTestsFromDBPath(logger *log.Entry, lynisDBPath string) (testdb, error) {
	db, err := os.Open(lynisDBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB from path %v: %w", lynisDBPath, err)
	}
	defer db.Close()
	scanner := bufio.NewScanner(db)

	return parseTestsFromFileScanner(logger, scanner)
}

func parseTestsFromFileScanner(logger *log.Entry, scanner FileScanner) (testdb, error) {
	output := testdb{}
	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments which start with #
		if line[0] == '#' {
			continue
		}

		// Sometimes there is an extra ":" on the end and sometimes
		// there isn't and that screws up the split below so trim them
		// off.
		line = strings.Trim(line, ":")

		parts := strings.Split(line, ":")
		if len(parts) != lynisDBEntryParts {
			logger.Warnf("Ignoring malformed line in lynis DB: %v", line)
			continue
		}

		entry := DBTestEntry{
			Category:    parts[lynisDBEntryCategoryIndex],
			Description: parts[lynisDBEntryDescriptionIndex],
		}
		output[parts[lynisDBEntryTestNameIndex]] = entry
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read lines from DB file: %w", err)
	}

	return output, nil
}
