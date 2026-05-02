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

package log

import (
	"context"
	"fmt"
	"os"
	"strings"

	tus "gitee.com/ivfzhou/tus_client/v2"
)

var (
	defaultRedisLoggerImpl RedisLogger = &defaultRedisLogger{}
	defaultTusdLoggerImpl  tus.Logger  = &defaultTusdLogger{}
	defaultLoggerImpl      Logger      = &defaultLogger{}
)

type (
	defaultRedisLogger struct{}
	defaultTusdLogger  struct{}
	defaultLogger      struct{}
)

func (l *defaultRedisLogger) Printf(_ context.Context, format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stdout, format+"\n", args...)
}

func (l *defaultTusdLogger) Printf(_ context.Context, level int, format string, args ...any) {
	level2 := LevelTusdTo[level]
	if level2 >= GetLevel() {
		_, _ = fmt.Fprintf(os.Stdout, format+"\n", args...)
	}
}

func (l *defaultLogger) Debug(_ context.Context, args ...any) {
	if GetLevel() <= LevelDebug {
		_, _ = fmt.Fprintln(os.Stdout, args...)
	}
}

func (l *defaultLogger) Debugf(_ context.Context, format string, args ...any) {
	if GetLevel() <= LevelDebug {
		_, _ = fmt.Fprintf(os.Stdout, format+"\n", args...)
	}
}

func (l *defaultLogger) Info(_ context.Context, args ...any) {
	if GetLevel() <= LevelInfo {
		_, _ = fmt.Fprintln(os.Stdout, args...)
	}
}

func (l *defaultLogger) Infof(_ context.Context, format string, args ...any) {
	if GetLevel() <= LevelInfo {
		_, _ = fmt.Fprintf(os.Stdout, format+"\n", args...)
	}
}

func (l *defaultLogger) Warn(_ context.Context, args ...any) {
	if GetLevel() <= LevelWarn {
		_, _ = fmt.Fprintln(os.Stdout, args...)
	}
}

func (l *defaultLogger) Warnf(_ context.Context, format string, args ...any) {
	if GetLevel() <= LevelWarn {
		_, _ = fmt.Fprintf(os.Stdout, format+"\n", args...)
	}
}

func (l *defaultLogger) Error(_ context.Context, args ...any) {
	if GetLevel() <= LevelError {
		_, _ = fmt.Fprintln(os.Stdout, args...)
	}
}

func (l *defaultLogger) Errorf(_ context.Context, format string, args ...any) {
	if GetLevel() <= LevelError {
		_, _ = fmt.Fprintf(os.Stdout, format+"\n", args...)
	}
}

func (l *defaultLogger) ErrorIf(_ context.Context, err error, msg string, args ...any) {
	if err != nil && GetLevel() <= LevelError {
		arr := make([]any, 1, len(args)+2)
		arr[0] = msg
		arr = append(arr, args...)
		arr = append(arr, err)
		_, _ = fmt.Fprintln(os.Stdout, arr...)
	}
}

func (l *defaultLogger) ErrorIff(_ context.Context, err error, format string, args ...any) {
	if err != nil && GetLevel() <= LevelError {
		format = strings.TrimRight(format, "\n") + " %v\n"
		arr := make([]any, 0, len(args)+1)
		arr = append(arr, args...)
		arr = append(arr, err)
		_, _ = fmt.Fprintf(os.Stdout, format, arr...)
	}
}

func (l *defaultLogger) Fatal(_ context.Context, exitCode int, args ...any) {
	if GetLevel() <= LevelFatal {
		_, _ = fmt.Fprintln(os.Stdout, args...)
		os.Exit(exitCode)
	}
}

func (l *defaultLogger) Fatalf(_ context.Context, exitCode int, format string, args ...any) {
	if GetLevel() <= LevelFatal {
		_, _ = fmt.Fprintf(os.Stdout, format+"\n", args...)
		os.Exit(exitCode)
	}
}

func (l *defaultLogger) FatalIf(_ context.Context, exitCode int, err error, msg string, args ...any) {
	if err != nil && GetLevel() <= LevelFatal {
		arr := make([]any, 1, len(args)+2)
		arr[0] = msg
		arr = append(arr, args...)
		arr = append(arr, err)
		_, _ = fmt.Fprintln(os.Stdout, arr...)
		os.Exit(exitCode)
	}
}

func (l *defaultLogger) FatalIff(_ context.Context, exitCode int, err error, format string, args ...any) {
	if err != nil && GetLevel() <= LevelFatal {
		format = strings.TrimRight(format, "\n") + " %v\n"
		arr := make([]any, 0, len(args)+1)
		arr = append(arr, args...)
		arr = append(arr, err)
		_, _ = fmt.Fprintf(os.Stdout, format, arr...)
		os.Exit(exitCode)
	}
}
