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
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	log "github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"
)

func TestNewTestDB(t *testing.T) {
	logger, _ := logrusTest.NewNullLogger()
	logEntry := logger.WithField("test", "valueToMisconfiguration")

	type args struct {
		logger     *log.Entry
		lynisDBDir string
	}
	tests := []struct {
		name    string
		args    args
		want    *TestDB
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				logger:     logEntry,
				lynisDBDir: "./testdata/simpledb/db",
			},
			want: &TestDB{
				tests: testdb{
					"ACCT-2754": {
						Category:    "security",
						Description: "Check for available FreeBSD accounting information",
					},
					"ACCT-2760": {
						Category:    "security",
						Description: "Check for available OpenBSD accounting information",
					},
					"ACCT-9622": {
						Category:    "security",
						Description: "Check for available Linux accounting information",
					},
					"ACCT-9626": {
						Category:    "security",
						Description: "Check for sysstat accounting data",
					},
					"ACCT-9628": {
						Category:    "security",
						Description: "Check for auditd",
					},
				},
			},
		},
		{
			name: "missing db file",
			args: args{
				logger:     logEntry,
				lynisDBDir: "./testdata/does-not-exist",
			},
			wantErr: true,
		},
		{
			name: "malformed db file",
			args: args{
				logger:     logEntry,
				lynisDBDir: "./testdata/baddb/db",
			},
			want: &TestDB{
				tests: testdb{
					"ACCT-2754": {
						Category:    "security",
						Description: "Check for available FreeBSD accounting information",
					},
					"ACCT-2760": {
						Category:    "security",
						Description: "Check for available OpenBSD accounting information",
					},
					"ACCT-9622": {
						Category:    "security",
						Description: "Check for available Linux accounting information",
					},
					"ACCT-9626": {
						Category:    "security",
						Description: "Check for sysstat accounting data",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTestDB(tt.args.logger, tt.args.lynisDBDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTestDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(TestDB{})); diff != "" {
				t.Errorf("NewTestDB() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTestDB_GetCategoryForTestID(t *testing.T) {
	type fields struct {
		tests testdb
	}
	type args struct {
		testid string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "test id that exists",
			fields: fields{
				tests: testdb{
					"Test1": {Category: "Cat1", Description: "Description1"},
					"Test2": {Category: "Cat2", Description: "Description2"},
					"Test3": {Category: "Cat3", Description: "Description3"},
				},
			},
			args: args{
				testid: "Test1",
			},
			want: "Cat1",
		},
		{
			name: "test id that does not exist",
			fields: fields{
				tests: testdb{
					"Test1": {Category: "Cat1", Description: "Description1"},
					"Test2": {Category: "Cat2", Description: "Description2"},
					"Test3": {Category: "Cat3", Description: "Description3"},
				},
			},
			args: args{
				testid: "Test4",
			},
			want: "unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &TestDB{
				tests: tt.fields.tests,
			}
			got := a.GetCategoryForTestID(tt.args.testid)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("TestDB.GetCategoryForTestID() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTestDB_GetDescriptionForTestID(t *testing.T) {
	type fields struct {
		tests testdb
	}
	type args struct {
		testid string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "test id that exists",
			fields: fields{
				tests: testdb{
					"Test1": {Category: "Cat1", Description: "Description1"},
					"Test2": {Category: "Cat2", Description: "Description2"},
					"Test3": {Category: "Cat3", Description: "Description3"},
				},
			},
			args: args{
				testid: "Test1",
			},
			want: "Description1",
		},
		{
			name: "test id that does not exist",
			fields: fields{
				tests: testdb{
					"Test1": {Category: "Cat1", Description: "Description1"},
					"Test2": {Category: "Cat2", Description: "Description2"},
					"Test3": {Category: "Cat3", Description: "Description3"},
				},
			},
			args: args{
				testid: "Test4",
			},
			want: "unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &TestDB{
				tests: tt.fields.tests,
			}
			got := a.GetDescriptionForTestID(tt.args.testid)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("TestDB.GetDescriptionForTestID() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_parseTestsFromDBPath(t *testing.T) {
	logger, _ := logrusTest.NewNullLogger()
	logEntry := logger.WithField("test", "valueToMisconfiguration")

	type args struct {
		logger      *log.Entry
		lynisDBPath string
	}
	tests := []struct {
		name    string
		args    args
		want    testdb
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				logger:      logEntry,
				lynisDBPath: "./testdata/simpledb/db/tests.db",
			},
			want: testdb{
				"ACCT-2754": {
					Category:    "security",
					Description: "Check for available FreeBSD accounting information",
				},
				"ACCT-2760": {
					Category:    "security",
					Description: "Check for available OpenBSD accounting information",
				},
				"ACCT-9622": {
					Category:    "security",
					Description: "Check for available Linux accounting information",
				},
				"ACCT-9626": {
					Category:    "security",
					Description: "Check for sysstat accounting data",
				},
				"ACCT-9628": {
					Category:    "security",
					Description: "Check for auditd",
				},
			},
		},
		{
			name: "missing db file",
			args: args{
				logger:      logEntry,
				lynisDBPath: "./testdata/does-not-exist",
			},
			wantErr: true,
		},
		{
			name: "malformed db file",
			args: args{
				logger:      logEntry,
				lynisDBPath: "./testdata/baddb/db/tests.db",
			},
			want: testdb{
				"ACCT-2754": {
					Category:    "security",
					Description: "Check for available FreeBSD accounting information",
				},
				"ACCT-2760": {
					Category:    "security",
					Description: "Check for available OpenBSD accounting information",
				},
				"ACCT-9622": {
					Category:    "security",
					Description: "Check for available Linux accounting information",
				},
				"ACCT-9626": {
					Category:    "security",
					Description: "Check for sysstat accounting data",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTestsFromDBPath(tt.args.logger, tt.args.lynisDBPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTestsFromDBPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("parseTestsFromDBPath() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_parseTestsFromFileScanner(t *testing.T) {
	logger, _ := logrusTest.NewNullLogger()
	logEntry := logger.WithField("test", "valueToMisconfiguration")

	type args struct {
		logger  *log.Entry
		scanner FileScanner
	}
	tests := []struct {
		name    string
		args    args
		want    testdb
		wantErr bool
	}{
		{
			name: "happy scanner",
			args: args{
				logger: logEntry,
				scanner: &TestScanner{
					lines: []string{"ACCT-9622:test:security:accounting:Linux:Check for available Linux accounting information:"},
				},
			},
			want: testdb{
				"ACCT-9622": {
					Category:    "security",
					Description: "Check for available Linux accounting information",
				},
			},
		},
		{
			name: "scanner error",
			args: args{
				logger: logEntry,
				scanner: &TestScanner{
					lines: []string{"ACCT-9622:test:security:accounting:Linux:Check for available Linux accounting information:"},
					err:   errors.New("scanner error"),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTestsFromFileScanner(tt.args.logger, tt.args.scanner)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTestsFromFileScanner() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("parseTestsFromFileScanner() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
