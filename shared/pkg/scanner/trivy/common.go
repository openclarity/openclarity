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

package trivy

import (
	log "github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	dlog "github.com/aquasecurity/go-dep-parser/pkg/log"
	flog "github.com/aquasecurity/trivy/pkg/fanal/log"
	trivyLog "github.com/aquasecurity/trivy/pkg/log"

	"github.com/openclarity/kubeclarity/shared/pkg/config"
	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
)

const ScannerName = "trivy"

func New(c job_manager.IsConfig,
	logger *log.Entry,
	resultChan chan job_manager.Result,
) job_manager.Job {
	conf := c.(*config.Config) // nolint:forcetypeassert

	logger = logger.Dup().WithField("scanner", ScannerName)

	// Init trivy's loggers with a hook into our logger
	lc := logrusCore{logger}
	zap := zap.New(lc)
	trivyLog.Logger = zap.Sugar()
	dlog.SetLogger(trivyLog.Logger)
	flog.SetLogger(trivyLog.Logger)

	return &LocalScanner{
		logger:     logger,
		config:     config.CreateLocalScannerTrivyConfigEx(conf.Scanner, conf.Registry),
		resultChan: resultChan,
		localImage: conf.LocalImageScan,
	}
}

type logrusCore struct {
	logger *log.Entry
}

func (lc logrusCore) Enabled(zapcore.Level) bool {
	return true
}

func (lc logrusCore) With(fields []zapcore.Field) zapcore.Core {
	logger := lc.logger.Dup()
	for _, field := range fields {
		logger = lc.logger.WithField(field.Key, field.Interface)
	}
	return logrusCore{logger}
}

func (lc logrusCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return checkedEntry.AddCore(entry, lc)
}

func (lc logrusCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	logger := lc.logger.Dup()
	for _, field := range fields {
		logger = lc.logger.WithField(field.Key, field.Interface)
	}

	switch entry.Level {
	case zapcore.DebugLevel:
		logger.Debug(entry.Message)
	case zapcore.InfoLevel:
		logger.Info(entry.Message)
	case zapcore.WarnLevel:
		logger.Warn(entry.Message)
	case zapcore.ErrorLevel:
		logger.Error(entry.Message)
	case zapcore.DPanicLevel:
		logger.Panic(entry.Message)
	case zapcore.PanicLevel:
		logger.Panic(entry.Message)
	case zapcore.FatalLevel:
		logger.Fatal(entry.Message)
	case zapcore.InvalidLevel:
	}

	return nil
}

func (lc logrusCore) Sync() error {
	return nil
}
