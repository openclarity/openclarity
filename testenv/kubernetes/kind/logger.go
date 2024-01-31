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

package kind

import (
	"github.com/sirupsen/logrus"
	kindlog "sigs.k8s.io/kind/pkg/log"
)

var (
	_ kindlog.Logger     = &Logger{}
	_ kindlog.InfoLogger = &Logger{}
)

type Logger struct {
	*logrus.Entry
}

func (l *Logger) Warn(message string) {
	l.Entry.Warn(message)
}

func (l *Logger) Error(message string) {
	l.Entry.Error(message)
}

const (
	kindLogLevelInfo  = 0
	kindLogLevelDebug = 1
	kindLogLevelTrace = 2
)

func (l *Logger) V(level kindlog.Level) kindlog.InfoLogger {
	entry := logrus.NewEntry(l.Logger)
	switch {
	case level <= kindLogLevelInfo:
		entry.Level = logrus.InfoLevel
	case level == kindLogLevelDebug:
		entry.Level = logrus.DebugLevel
	case level >= kindLogLevelTrace:
		entry.Level = logrus.TraceLevel
	}

	return &Logger{entry}
}

func (l *Logger) Info(message string) {
	l.Entry.Info(message)
}

func (l *Logger) Enabled() bool {
	return l.Logger.IsLevelEnabled(l.Entry.Level)
}

func NewLogger(logger *logrus.Entry) *Logger {
	if logger == nil {
		logger = logrus.NewEntry(logrus.StandardLogger())
	}
	return &Logger{logger}
}
