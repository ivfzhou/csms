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
	"strings"

	tus "gitee.com/ivfzhou/tus_client/v2"
)

// 日志等级。
const (
	LevelDebug Level = iota + 1
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

var (
	// LevelTusdTo Tusd 日志等级与本系统的日志等级的映射。
	LevelTusdTo = map[int]Level{
		tus.Level_Debug:   LevelDebug,
		tus.Level_Info:    LevelInfo,
		tus.Level_Warning: LevelWarn,
		tus.Level_Error:   LevelError,
		tus.Level_Silent:  LevelFatal,
	}
	// LevelTusdFrom Tusd 日志等级与本系统的日志等级的映射。
	LevelTusdFrom = map[Level]int{
		LevelDebug: tus.Level_Debug,
		LevelInfo:  tus.Level_Info,
		LevelWarn:  tus.Level_Warning,
		LevelError: tus.Level_Error,
		LevelFatal: tus.Level_Silent,
	}
)

// RedisLogger Redis 日志打印接口。
type RedisLogger interface {
	Printf(ctx context.Context, format string, args ...any)
}

// Logger 日志打印接口。
type Logger interface {
	// Debug 打印日志。
	Debug(ctx context.Context, args ...any)

	// Debugf 打印日志。
	Debugf(ctx context.Context, format string, args ...any)

	// Info 打印日志。
	Info(ctx context.Context, args ...any)

	// Infof 打印日志。
	Infof(ctx context.Context, format string, args ...any)

	// Warn 打印日志。
	Warn(ctx context.Context, args ...any)

	// Warnf 打印日志。
	Warnf(ctx context.Context, format string, args ...any)

	// Error 打印日志。
	Error(ctx context.Context, args ...any)

	// Errorf 打印日志。
	Errorf(ctx context.Context, format string, args ...any)

	// ErrorIf 打印日志。
	ErrorIf(ctx context.Context, err error, msg string, args ...any)

	// ErrorIff 打印日志。
	ErrorIff(ctx context.Context, err error, format string, args ...any)

	// Fatal 打印日志。
	Fatal(ctx context.Context, exitCode int, args ...any)

	// Fatalf 打印日志。
	Fatalf(ctx context.Context, exitCode int, format string, args ...any)

	// FatalIf 打印日志。
	FatalIf(ctx context.Context, exitCode int, err error, msg string, args ...any)

	// FatalIff 打印日志。
	FatalIff(ctx context.Context, exitCode int, err error, format string, args ...any)
}

// Level 日志等级。
type Level int

// ParseLevel 字符串转成 Level。
func ParseLevel(level string) Level {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case LevelDebug.String():
		return LevelDebug
	case LevelInfo.String():
		return LevelInfo
	case LevelWarn.String():
		return LevelWarn
	case LevelError.String():
		return LevelError
	case LevelFatal.String():
		return LevelFatal
	default:
		return 0
	}
}

// String 日志字符串形式。
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	}
	return "UNKNOWN"
}
