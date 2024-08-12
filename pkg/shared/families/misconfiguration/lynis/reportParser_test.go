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
	logrusTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/openclarity/vmclarity/pkg/shared/families/misconfiguration/types"
)

func TestNewReportParser(t *testing.T) {
	logger, _ := logrusTest.NewNullLogger()
	logEntry := logger.WithField("test", "valueToMisconfiguration")
	testdb, err := NewTestDB(logEntry, "./testdata/db")
	if err != nil {
		t.Fatalf("Unable to load test db: %v", err)
	}

	type args struct {
		testdb *TestDB
	}
	tests := []struct {
		name string
		args args
		want *ReportParser
	}{
		{
			name: "sanity",
			args: args{
				testdb: testdb,
			},
			want: &ReportParser{
				testdb: testdb,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewReportParser(tt.args.testdb)
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(ReportParser{}), cmp.AllowUnexported(TestDB{})); diff != "" {
				t.Errorf("NewReportParser() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestReportParser_ParseLynisReport(t *testing.T) {
	logger, _ := logrusTest.NewNullLogger()
	logEntry := logger.WithField("test", "valueToMisconfiguration")
	testdb, err := NewTestDB(logEntry, "./testdata/db")
	if err != nil {
		t.Fatalf("Unable to load test db: %v", err)
	}

	type fields struct {
		testdb *TestDB
	}
	type args struct {
		scanPath   string
		reportPath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []types.Misconfiguration
		wantErr bool
	}{
		{
			name: "sanity",
			fields: fields{
				testdb: testdb,
			},
			args: args{
				scanPath:   "scanPath",
				reportPath: "./testdata/lynis-report.dat",
			},
			want: testdataMisconfigurations,
		},
		{
			name: "missing report",
			fields: fields{
				testdb: testdb,
			},
			args: args{
				scanPath:   "scanPath",
				reportPath: "./testdata/does-not-exist.dat",
			},
			wantErr: true,
		},
		{
			name: "invalid lynis report",
			fields: fields{
				testdb: testdb,
			},
			args: args{
				scanPath:   "scanPath",
				reportPath: "./testdata/lynis-report-bad.txt",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &ReportParser{
				testdb: tt.fields.testdb,
			}
			got, err := a.ParseLynisReport(tt.args.scanPath, tt.args.reportPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReportParser.ParseLynisReport() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ReportParser.ParseLynisReport() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestReportParser_parseLynisReportLine(t *testing.T) {
	logger, _ := logrusTest.NewNullLogger()
	logEntry := logger.WithField("test", "valueToMisconfiguration")
	testdb, err := NewTestDB(logEntry, "./testdata/db")
	if err != nil {
		t.Fatalf("Unable to load test db: %v", err)
	}

	type fields struct {
		testdb *TestDB
	}
	type args struct {
		scanPath string
		line     string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		want1   types.Misconfiguration
		wantErr bool
	}{
		{
			name: "suggestion",
			fields: fields{
				testdb: testdb,
			},
			args: args{
				scanPath: "scanPath",
				line:     "suggestion[]=FILE-6362|Set the sticky bit on /home/ubuntu/debian11/tmp, to prevent users deleting (by other owned) files in the /tmp directory.|/tmp|text:Set sticky bit|",
			},
			want: true,
			want1: types.Misconfiguration{
				ScannedPath:     "scanPath",
				TestCategory:    "security",
				TestID:          "FILE-6362",
				TestDescription: "Checking /tmp sticky bit",
				Severity:        "LowSeverity",
				Message:         "Set the sticky bit on /home/ubuntu/debian11/tmp, to prevent users deleting (by other owned) files in the /tmp directory. Details: /tmp",
				Remediation:     "text:Set sticky bit",
			},
		},
		{
			name: "warning",
			fields: fields{
				testdb: testdb,
			},
			args: args{
				scanPath: "scanPath",
				line:     "warning[]=FILE-6362|Set the sticky bit on /home/ubuntu/debian11/tmp, to prevent users deleting (by other owned) files in the /tmp directory.|/tmp|text:Set sticky bit|",
			},
			want: true,
			want1: types.Misconfiguration{
				ScannedPath:     "scanPath",
				TestCategory:    "security",
				TestID:          "FILE-6362",
				TestDescription: "Checking /tmp sticky bit",
				Severity:        "HighSeverity",
				Message:         "Set the sticky bit on /home/ubuntu/debian11/tmp, to prevent users deleting (by other owned) files in the /tmp directory. Details: /tmp",
				Remediation:     "text:Set sticky bit",
			},
		},
		{
			name: "other option",
			fields: fields{
				testdb: testdb,
			},
			args: args{
				scanPath: "scanPath",
				line:     "pam_module[]=FILE-6362|Set the sticky bit on /home/ubuntu/debian11/tmp, to prevent users deleting (by other owned) files in the /tmp directory.|/tmp|text:Set sticky bit|",
			},
			want:  false,
			want1: types.Misconfiguration{},
		},
		{
			name: "comment",
			fields: fields{
				testdb: testdb,
			},
			args: args{
				scanPath: "scanPath",
				line:     "#comments start with a pound symbol",
			},
			want:  false,
			want1: types.Misconfiguration{},
		},
		{
			name: "section",
			fields: fields{
				testdb: testdb,
			},
			args: args{
				scanPath: "scanPath",
				line:     "[sections are described in square brackets]",
			},
			want:  false,
			want1: types.Misconfiguration{},
		},
		{
			name: "missing option",
			fields: fields{
				testdb: testdb,
			},
			args: args{
				scanPath: "scanPath",
				line:     "FILE-6362|Set the sticky bit on /home/ubuntu/debian11/tmp, to prevent users deleting (by other owned) files in the /tmp directory.|/tmp|text:Set sticky bit|",
			},
			wantErr: true,
		},
		{
			name: "suggestion but test is a LYNIS sggestion",
			fields: fields{
				testdb: testdb,
			},
			args: args{
				scanPath: "scanPath",
				line:     "suggestion[]=LYNIS|This release is more than 4 months old. Check the website or GitHub to see if there is an update available.|-|-|",
			},
			want:  false,
			want1: types.Misconfiguration{},
		},
		{
			name: "suggestion short line",
			fields: fields{
				testdb: testdb,
			},
			args: args{
				scanPath: "scanPath",
				line:     "suggestion[]=FILE-6362|Set the sticky bit on /home/ubuntu/debian11/tmp, to prevent users deleting (by other owned) files in the /tmp directory.|/tmp|",
			},
			wantErr: true,
		},
		{
			name: "warning short line",
			fields: fields{
				testdb: testdb,
			},
			args: args{
				scanPath: "scanPath",
				line:     "warning[]=FILE-6362|Set the sticky bit on /home/ubuntu/debian11/tmp, to prevent users deleting (by other owned) files in the /tmp directory.|/tmp|",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &ReportParser{
				testdb: tt.fields.testdb,
			}
			got, got1, err := a.parseLynisReportLine(tt.args.scanPath, tt.args.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReportParser.parseLynisReportLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ReportParser.parseLynisReportLine() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want1, got1); diff != "" {
				t.Errorf("ReportParser.parseLynisReportLine() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestReportParser_valueToMisconfiguration(t *testing.T) {
	logger, _ := logrusTest.NewNullLogger()
	logEntry := logger.WithField("test", "valueToMisconfiguration")
	testdb, err := NewTestDB(logEntry, "./testdata/db")
	if err != nil {
		t.Fatalf("Unable to load test db: %v", err)
	}

	type fields struct {
		testdb *TestDB
	}
	type args struct {
		scanPath string
		value    string
		severity types.Severity
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    types.Misconfiguration
		wantErr bool
	}{
		{
			name: "good value",
			fields: fields{
				testdb: testdb,
			},
			args: args{
				scanPath: "scanPath",
				value:    "FILE-6362|Set the sticky bit on /home/ubuntu/debian11/tmp, to prevent users deleting (by other owned) files in the /tmp directory.|/tmp|text:Set sticky bit|",
				severity: types.LowSeverity,
			},
			want: types.Misconfiguration{
				ScannedPath:     "scanPath",
				TestCategory:    "security",
				TestID:          "FILE-6362",
				TestDescription: "Checking /tmp sticky bit",
				Severity:        "LowSeverity",
				Message:         "Set the sticky bit on /home/ubuntu/debian11/tmp, to prevent users deleting (by other owned) files in the /tmp directory. Details: /tmp",
				Remediation:     "text:Set sticky bit",
			},
		},
		{
			name: "different severity",
			fields: fields{
				testdb: testdb,
			},
			args: args{
				scanPath: "scanPath",
				value:    "FILE-6362|Set the sticky bit on /home/ubuntu/debian11/tmp, to prevent users deleting (by other owned) files in the /tmp directory.|/tmp|text:Set sticky bit|",
				severity: types.HighSeverity,
			},
			want: types.Misconfiguration{
				ScannedPath:     "scanPath",
				TestCategory:    "security",
				TestID:          "FILE-6362",
				TestDescription: "Checking /tmp sticky bit",
				Severity:        "HighSeverity",
				Message:         "Set the sticky bit on /home/ubuntu/debian11/tmp, to prevent users deleting (by other owned) files in the /tmp directory. Details: /tmp",
				Remediation:     "text:Set sticky bit",
			},
		},
		{
			name: "different scanPath",
			fields: fields{
				testdb: testdb,
			},
			args: args{
				scanPath: "scanPath2",
				value:    "FILE-6362|Set the sticky bit on /home/ubuntu/debian11/tmp, to prevent users deleting (by other owned) files in the /tmp directory.|/tmp|text:Set sticky bit|",
				severity: types.HighSeverity,
			},
			want: types.Misconfiguration{
				ScannedPath:     "scanPath2",
				TestCategory:    "security",
				TestID:          "FILE-6362",
				TestDescription: "Checking /tmp sticky bit",
				Severity:        "HighSeverity",
				Message:         "Set the sticky bit on /home/ubuntu/debian11/tmp, to prevent users deleting (by other owned) files in the /tmp directory. Details: /tmp",
				Remediation:     "text:Set sticky bit",
			},
		},
		{
			name: "not enough fields value",
			fields: fields{
				testdb: testdb,
			},
			args: args{
				scanPath: "scanPath",
				value:    "FILE-6362|Set the sticky bit on /home/ubuntu/debian11/tmp, to prevent users deleting (by other owned) files in the /tmp directory.|/tmp|",
				severity: types.LowSeverity,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &ReportParser{
				testdb: tt.fields.testdb,
			}
			got, err := a.valueToMisconfiguration(tt.args.scanPath, tt.args.value, tt.args.severity)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReportParser.valueToMisconfiguration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ReportParser.valueToMisconfiguration() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

type TestScanner struct {
	lines []string
	text  string
	err   error
}

func (t *TestScanner) Scan() bool {
	if len(t.lines) < 1 {
		return false
	}
	t.text = t.lines[0]
	t.lines = t.lines[1:]
	return true
}

func (t *TestScanner) Text() string {
	return t.text
}

func (t *TestScanner) Err() error {
	return t.err
}

func TestReportParser_scanLynisReportFile(t *testing.T) {
	logger, _ := logrusTest.NewNullLogger()
	logEntry := logger.WithField("test", "valueToMisconfiguration")
	testdb, err := NewTestDB(logEntry, "./testdata/db")
	if err != nil {
		t.Fatalf("Unable to load test db: %v", err)
	}

	type fields struct {
		testdb *TestDB
	}
	type args struct {
		scanPath string
		scanner  FileScanner
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []types.Misconfiguration
		wantErr bool
	}{
		{
			name: "happy scanner",
			fields: fields{
				testdb: testdb,
			},
			args: args{
				scanPath: "scanPath",
				scanner: &TestScanner{
					lines: []string{"suggestion[]=FILE-6362|Set the sticky bit on /home/ubuntu/debian11/tmp, to prevent users deleting (by other owned) files in the /tmp directory.|/tmp|text:Set sticky bit|"},
				},
			},
			want: []types.Misconfiguration{
				{
					ScannedPath:     "scanPath",
					TestCategory:    "security",
					TestID:          "FILE-6362",
					TestDescription: "Checking /tmp sticky bit",
					Severity:        "LowSeverity",
					Message:         "Set the sticky bit on /home/ubuntu/debian11/tmp, to prevent users deleting (by other owned) files in the /tmp directory. Details: /tmp",
					Remediation:     "text:Set sticky bit",
				},
			},
		},
		{
			name: "invalid lines",
			fields: fields{
				testdb: testdb,
			},
			args: args{
				scanPath: "scanPath",
				scanner: &TestScanner{
					lines: []string{"FILE-6362|Set the sticky bit on /home/ubuntu/debian11/tmp, to prevent users deleting (by other owned) files in the /tmp directory.|/tmp|text:Set sticky bit|"},
				},
			},
			wantErr: true,
		},
		{
			name: "scanner error",
			fields: fields{
				testdb: testdb,
			},
			args: args{
				scanPath: "scanPath",
				scanner: &TestScanner{
					lines: []string{"suggestion[]=FILE-6362|Set the sticky bit on /home/ubuntu/debian11/tmp, to prevent users deleting (by other owned) files in the /tmp directory.|/tmp|text:Set sticky bit|"},
					err:   errors.New("Scanner error"),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &ReportParser{
				testdb: tt.fields.testdb,
			}
			got, err := a.scanLynisReportFile(tt.args.scanPath, tt.args.scanner)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReportParser.scanLynisReportFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ReportParser.scanLynisReportFile() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

var testdataMisconfigurations []types.Misconfiguration = []types.Misconfiguration{
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "BOOT-5122",
		TestDescription: "Check for GRUB boot password",
		Severity:        "LowSeverity",
		Message:         "Set a password on GRUB boot loader to prevent altering boot configuration (e.g. boot in single user mode without password) Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "BOOT-5264",
		TestDescription: "Run systemd-analyze security",
		Severity:        "LowSeverity",
		Message:         "Consider hardening system services Details: Run '/usr/bin/systemd-analyze security SERVICE' for each service",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "KRNL-5788",
		TestDescription: "Checking availability new Linux kernel",
		Severity:        "LowSeverity",
		Message:         "Determine why /home/ubuntu/debian11/vmlinuz or /home/ubuntu/debian11/boot/vmlinuz is missing on this Debian/Ubuntu system. Details: /vmlinuz or /boot/vmlinuz",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "KRNL-5820",
		TestDescription: "Checking core dumps configuration",
		Severity:        "LowSeverity",
		Message:         "If not required, consider explicit disabling of core dump in /etc/security/limits.conf file Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "KRNL-5830",
		TestDescription: "Checking if system is running on the latest installed kernel",
		Severity:        "HighSeverity",
		Message:         "Reboot of system is most likely needed Details: ",
		Remediation:     "text:reboot",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "AUTH-9230",
		TestDescription: "Check group password hashing rounds",
		Severity:        "LowSeverity",
		Message:         "Configure password hashing rounds in /etc/login.defs Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "AUTH-9262",
		TestDescription: "Checking presence password strength testing tools (PAM)",
		Severity:        "LowSeverity",
		Message:         "Install a PAM module for password strength testing like pam_cracklib or pam_passwdqc Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "AUTH-9286",
		TestDescription: "Checking user password aging",
		Severity:        "LowSeverity",
		Message:         "Configure minimum password age in /etc/login.defs Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "AUTH-9286",
		TestDescription: "Checking user password aging",
		Severity:        "LowSeverity",
		Message:         "Configure maximum password age in /etc/login.defs Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "AUTH-9328",
		TestDescription: "Default umask values",
		Severity:        "LowSeverity",
		Message:         "Default umask in /etc/login.defs could be more strict like 027 Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "FILE-6310",
		TestDescription: "Checking /tmp, /home and /var directory",
		Severity:        "LowSeverity",
		Message:         "To decrease the impact of a full /home file system, place /home on a separate partition Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "FILE-6310",
		TestDescription: "Checking /tmp, /home and /var directory",
		Severity:        "LowSeverity",
		Message:         "To decrease the impact of a full /tmp file system, place /tmp on a separate partition Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "FILE-6310",
		TestDescription: "Checking /tmp, /home and /var directory",
		Severity:        "LowSeverity",
		Message:         "To decrease the impact of a full /var file system, place /var on a separate partition Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "FILE-6362",
		TestDescription: "Checking /tmp sticky bit",
		Severity:        "LowSeverity",
		Message:         "Set the sticky bit on /home/ubuntu/debian11/tmp, to prevent users deleting (by other owned) files in the /tmp directory. Details: /tmp",
		Remediation:     "text:Set sticky bit",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "FILE-6363",
		TestDescription: "Checking /var/tmp sticky bit",
		Severity:        "LowSeverity",
		Message:         "Set the sticky bit on /home/ubuntu/debian11/var/tmp, to prevent users deleting (by other owned) files in the /var/tmp directory. Details: /var/tmp",
		Remediation:     "text:Set sticky bit",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "USB-1000",
		TestDescription: "Check if USB storage is disabled",
		Severity:        "LowSeverity",
		Message:         "Disable drivers like USB storage when not used, to prevent unauthorized storage or data theft Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "STRG-1846",
		TestDescription: "Check if firewire storage is disabled",
		Severity:        "LowSeverity",
		Message:         "Disable drivers like firewire storage when not used, to prevent unauthorized storage or data theft Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "NAME-4404",
		TestDescription: "Check /etc/hosts contains an entry for this server name",
		Severity:        "LowSeverity",
		Message:         "Add the IP name and FQDN to /etc/hosts for proper name resolving Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "PKGS-7346",
		TestDescription: "Search unpurged packages on system",
		Severity:        "LowSeverity",
		Message:         "Purge old/removed packages (1 found) with aptitude purge or dpkg --purge command. This will cleanup old configuration files, cron jobs and startup scripts. Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "PKGS-7370",
		TestDescription: "Checking for debsums utility",
		Severity:        "LowSeverity",
		Message:         "Install debsums utility for the verification of packages with known good database. Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "PKGS-7390",
		TestDescription: "Check Ubuntu database consistency",
		Severity:        "HighSeverity",
		Message:         "apt-get check returned a non successful exit code. Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "PKGS-7390",
		TestDescription: "Check Ubuntu database consistency",
		Severity:        "LowSeverity",
		Message:         "Run apt-get to perform a manual package database consistency check. Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "PKGS-7394",
		TestDescription: "Check for Ubuntu updates",
		Severity:        "LowSeverity",
		Message:         "Install package apt-show-versions for patch management purposes Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "PKGS-7420",
		TestDescription: "Detect toolkit to automatically download and apply upgrades",
		Severity:        "LowSeverity",
		Message:         "Consider using a tool to automatically apply upgrades Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "NETW-3200",
		TestDescription: "Determine available network protocols",
		Severity:        "LowSeverity",
		Message:         "Determine if protocol 'dccp' is really needed on this system Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "NETW-3200",
		TestDescription: "Determine available network protocols",
		Severity:        "LowSeverity",
		Message:         "Determine if protocol 'sctp' is really needed on this system Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "NETW-3200",
		TestDescription: "Determine available network protocols",
		Severity:        "LowSeverity",
		Message:         "Determine if protocol 'rds' is really needed on this system Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "NETW-3200",
		TestDescription: "Determine available network protocols",
		Severity:        "LowSeverity",
		Message:         "Determine if protocol 'tipc' is really needed on this system Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "FIRE-4513",
		TestDescription: "Check iptables for unused rules",
		Severity:        "LowSeverity",
		Message:         "Check iptables rules to see which rules are currently not used Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "SSH-7408",
		TestDescription: "Check SSH specific defined options",
		Severity:        "LowSeverity",
		Message:         "Consider hardening SSH configuration Details: AllowTcpForwarding (set YES to NO)",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "SSH-7408",
		TestDescription: "Check SSH specific defined options",
		Severity:        "LowSeverity",
		Message:         "Consider hardening SSH configuration Details: ClientAliveCountMax (set 3 to 2)",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "SSH-7408",
		TestDescription: "Check SSH specific defined options",
		Severity:        "LowSeverity",
		Message:         "Consider hardening SSH configuration Details: Compression (set YES to NO)",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "SSH-7408",
		TestDescription: "Check SSH specific defined options",
		Severity:        "LowSeverity",
		Message:         "Consider hardening SSH configuration Details: LogLevel (set INFO to VERBOSE)",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "SSH-7408",
		TestDescription: "Check SSH specific defined options",
		Severity:        "LowSeverity",
		Message:         "Consider hardening SSH configuration Details: MaxAuthTries (set 6 to 3)",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "SSH-7408",
		TestDescription: "Check SSH specific defined options",
		Severity:        "LowSeverity",
		Message:         "Consider hardening SSH configuration Details: MaxSessions (set 10 to 2)",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "SSH-7408",
		TestDescription: "Check SSH specific defined options",
		Severity:        "LowSeverity",
		Message:         "Consider hardening SSH configuration Details: Port (set 22 to )",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "SSH-7408",
		TestDescription: "Check SSH specific defined options",
		Severity:        "LowSeverity",
		Message:         "Consider hardening SSH configuration Details: TCPKeepAlive (set YES to NO)",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "SSH-7408",
		TestDescription: "Check SSH specific defined options",
		Severity:        "LowSeverity",
		Message:         "Consider hardening SSH configuration Details: X11Forwarding (set YES to NO)",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "SSH-7408",
		TestDescription: "Check SSH specific defined options",
		Severity:        "LowSeverity",
		Message:         "Consider hardening SSH configuration Details: AllowAgentForwarding (set YES to NO)",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "LOGG-2154",
		TestDescription: "Checking syslog configuration file",
		Severity:        "LowSeverity",
		Message:         "Enable logging to an external logging host for archiving purposes and additional protection Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "LOGG-2190",
		TestDescription: "Checking for deleted files in use",
		Severity:        "LowSeverity",
		Message:         "Check what deleted files are still in use and why. Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "BANN-7126",
		TestDescription: "Check issue banner file contents",
		Severity:        "LowSeverity",
		Message:         "Add a legal banner to /home/ubuntu/debian11/etc/issue, to warn unauthorized users Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "BANN-7130",
		TestDescription: "Check issue.net banner file contents",
		Severity:        "LowSeverity",
		Message:         "Add legal banner to /etc/issue.net, to warn unauthorized users Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "SCHD-7704",
		TestDescription: "Check crontab/cronjobs",
		Severity:        "HighSeverity",
		Message:         "Found one or more cronjob files with incorrect ownership (see log for details) Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "ACCT-9622",
		TestDescription: "Check for available Linux accounting information",
		Severity:        "LowSeverity",
		Message:         "Enable process accounting Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "ACCT-9626",
		TestDescription: "Check for sysstat accounting data",
		Severity:        "LowSeverity",
		Message:         "Enable sysstat to collect accounting (no results) Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "ACCT-9628",
		TestDescription: "Check for auditd",
		Severity:        "LowSeverity",
		Message:         "Enable auditd to collect audit information Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "FINT-4350",
		TestDescription: "File integrity software installed",
		Severity:        "LowSeverity",
		Message:         "Install a file integrity tool to monitor changes to critical and sensitive files Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "TOOL-5002",
		TestDescription: "Checking for automation tools",
		Severity:        "LowSeverity",
		Message:         "Determine if automation tools are present for system management Details: -",
		Remediation:     "-",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "FILE-7524",
		TestDescription: "Perform file permissions check",
		Severity:        "LowSeverity",
		Message:         "Consider restricting file permissions Details: See screen output or log file",
		Remediation:     "text:Use chmod to change file permissions",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "KRNL-6000",
		TestDescription: "Check sysctl key pairs in scan profile",
		Severity:        "LowSeverity",
		Message:         "One or more sysctl values differ from the scan profile and could be tweaked Details: ",
		Remediation:     "Change sysctl value or disable test (skip-test=KRNL-6000:<sysctl-key>)",
	},
	{
		ScannedPath:     "scanPath",
		TestCategory:    "security",
		TestID:          "HRDN-7230",
		TestDescription: "Check for malware scanner",
		Severity:        "LowSeverity",
		Message:         "Harden the system by installing at least one malware scanner, to perform periodic file system scans Details: -",
		Remediation:     "Install a tool like rkhunter, chkrootkit, OSSEC",
	},
}
