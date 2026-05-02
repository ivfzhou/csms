/*
 * Copyright (c) 2024 ivfzhou
 * csms is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package impl

import (
	"context"
	"os"
	"strings"

	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/log/internal"
)

type loggerImpl struct{}

// 新建实例。
func newLogger() log.Logger {
	return &loggerImpl{}
}

func (l *loggerImpl) Debug(ctx context.Context, args ...any) {
	if internal.GetLevel() <= log.LevelDebug {
		internal.SendBuilderToWriter(internal.CreateBuilder(ctx, args...).SetLevel(log.LevelDebug))
	}
}

func (l *loggerImpl) Debugf(ctx context.Context, format string, args ...any) {
	if internal.GetLevel() <= log.LevelDebug {
		internal.SendBuilderToWriter(internal.CreateBuilderf(ctx, format, args...).SetLevel(log.LevelDebug))
	}
}

func (l *loggerImpl) Info(ctx context.Context, args ...any) {
	if internal.GetLevel() <= log.LevelInfo {
		internal.SendBuilderToWriter(internal.CreateBuilder(ctx, args...).SetLevel(log.LevelInfo))
	}
}

func (l *loggerImpl) Infof(ctx context.Context, format string, args ...any) {
	if internal.GetLevel() <= log.LevelInfo {
		internal.SendBuilderToWriter(internal.CreateBuilderf(ctx, format, args...).SetLevel(log.LevelInfo))
	}
}

func (l *loggerImpl) Warn(ctx context.Context, args ...any) {
	if internal.GetLevel() <= log.LevelWarn {
		internal.SendBuilderToWriter(internal.CreateBuilder(ctx, args...).SetLevel(log.LevelWarn))
	}
}

func (l *loggerImpl) Warnf(ctx context.Context, format string, args ...any) {
	if internal.GetLevel() <= log.LevelWarn {
		internal.SendBuilderToWriter(internal.CreateBuilderf(ctx, format, args...).SetLevel(log.LevelWarn))
	}
}

func (l *loggerImpl) Error(ctx context.Context, args ...any) {
	if internal.GetLevel() <= log.LevelError {
		internal.SendBuilderToWriter(internal.CreateBuilder(ctx, args...).SetLevel(log.LevelError))
	}
}

func (l *loggerImpl) Errorf(ctx context.Context, format string, args ...any) {
	if internal.GetLevel() <= log.LevelError {
		internal.SendBuilderToWriter(internal.CreateBuilderf(ctx, format, args...).SetLevel(log.LevelError))
	}
}

func (l *loggerImpl) ErrorIf(ctx context.Context, err error, msg string, args ...any) {
	if err != nil && internal.GetLevel() <= log.LevelError {
		arr := make([]any, 1, len(args)+2)
		arr[0] = msg
		arr = append(arr, args...)
		arr = append(arr, err)
		internal.SendBuilderToWriter(internal.CreateBuilder(ctx, arr...).SetLevel(log.LevelError))
	}
}

func (l *loggerImpl) ErrorIff(ctx context.Context, err error, format string, args ...any) {
	if err != nil && internal.GetLevel() <= log.LevelError {
		format = strings.TrimRight(format, "\n") + " %v\n"
		arr := make([]any, 0, len(args)+1)
		arr = append(arr, args...)
		arr = append(arr, err)
		internal.SendBuilderToWriter(internal.CreateBuilderf(ctx, format, arr...).SetLevel(log.LevelError))
	}
}

func (l *loggerImpl) Fatal(ctx context.Context, exitCode int, args ...any) {
	if internal.GetLevel() <= log.LevelFatal {
		internal.SendBuilderToWriter(internal.CreateBuilder(ctx, args...).SetLevel(log.LevelFatal))
		os.Exit(exitCode)
	}
}

func (l *loggerImpl) Fatalf(ctx context.Context, exitCode int, format string, args ...any) {
	if internal.GetLevel() <= log.LevelFatal {
		internal.SendBuilderToWriter(internal.CreateBuilderf(ctx, format, args...).SetLevel(log.LevelFatal))
		os.Exit(exitCode)
	}
}

func (l *loggerImpl) FatalIf(ctx context.Context, exitCode int, err error, msg string, args ...any) {
	if err != nil && internal.GetLevel() <= log.LevelFatal {
		arr := make([]any, 1, len(args)+2)
		arr[0] = msg
		arr = append(arr, args...)
		arr = append(arr, err)
		internal.SendBuilderToWriter(internal.CreateBuilder(ctx, arr...).SetLevel(log.LevelFatal))
		os.Exit(exitCode)
	}
}

func (l *loggerImpl) FatalIff(ctx context.Context, exitCode int, err error, format string, args ...any) {
	if err != nil && internal.GetLevel() <= log.LevelFatal {
		format = strings.TrimRight(format, "\n") + " %v\n"
		arr := make([]any, 0, len(args)+1)
		arr = append(arr, args...)
		arr = append(arr, err)
		internal.SendBuilderToWriter(internal.CreateBuilderf(ctx, format, arr...).SetLevel(log.LevelFatal))
		os.Exit(exitCode)
	}
}
