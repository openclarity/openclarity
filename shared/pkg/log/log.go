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

package log

import (
	"io"
	"os"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/grpclog"
)

const (
	DefaultLogLevel = log.WarnLevel
)

func InitLogger(level string, output io.Writer) {
	logLevel, err := log.ParseLevel(level)
	if err != nil {
		log.Errorf("failed to prase log level, using default(%s): %v", DefaultLogLevel, err)
		logLevel = DefaultLogLevel
	}
	log.SetLevel(logLevel)
	setGrpcLogs(logLevel)

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:          true,
		DisableTimestamp:       false,
		DisableSorting:         true,
		DisableLevelTruncation: true,
		QuoteEmptyFields:       true,
	})

	if logLevel >= log.DebugLevel {
		log.SetReportCaller(true)
	}

	log.SetOutput(output)
}

// Set log level to Trace to see grpc logs. We do this so that they don't mess with our stdio.
func setGrpcLogs(level log.Level) {
	if level < log.TraceLevel {
		grpclog.SetLoggerV2(grpclog.NewLoggerV2(io.Discard, io.Discard, io.Discard))
	} else {
		verboseLevel := 2
		grpclog.SetLoggerV2(grpclog.NewLoggerV2WithVerbosity(os.Stderr, os.Stderr, os.Stderr, verboseLevel))
	}
}
