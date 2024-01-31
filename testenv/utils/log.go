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

package utils

import (
	"context"
	"io"

	log "github.com/sirupsen/logrus"
)

type LoggerContextKeyType string

const LoggerContextKey LoggerContextKeyType = "TestEnvContextLogger"

func GetLoggerFromContextOrDefault(ctx context.Context) *log.Entry {
	logger, ok := GetLoggerFromContext(ctx)
	if !ok {
		logger = log.NewEntry(log.StandardLogger()).WithContext(ctx)
	}
	return logger
}

func GetLoggerFromContextOrDiscard(ctx context.Context) *log.Entry {
	logger, ok := GetLoggerFromContext(ctx)
	if !ok {
		logger = log.NewEntry(&log.Logger{
			Out:   io.Discard,
			Level: 0,
		}).WithContext(ctx)
	}
	return logger
}

func GetLoggerFromContext(ctx context.Context) (*log.Entry, bool) {
	logger, ok := ctx.Value(LoggerContextKey).(*log.Entry)
	if ok && logger != nil {
		logger = logger.WithContext(ctx)
	}
	return logger, ok
}

func SetLoggerForContext(ctx context.Context, l *log.Entry) context.Context {
	return context.WithValue(ctx, LoggerContextKey, l)
}
