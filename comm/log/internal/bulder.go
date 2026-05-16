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

package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/log"
)

const (
	black = 30 + iota
	read
	green
	yellow
	blue
	magenta
	cyan
	white
)

const (
	timeFormat         = "2006-01-02 15:04:05.000"
	numberOfSkipCaller = 3
)

var (
	logBuilderPool    = sync.Pool{New: func() any { return &Builder{} }}
	findVersionRegexp = regexp.MustCompile(`^.*?(@v.*?)/.*$`)
)

// Builder 日志构建器。
type Builder struct {
	time         time.Time
	level        log.Level
	requestID    string
	requestURI   string
	caller       string
	userID       int
	appID        int
	apiAccountID int
	requestIP    string
	library      string
	format       string
	args         []any
}

// CreateBuilder 构建新的日志格式化器。
func CreateBuilder(ctx context.Context, args ...any) *Builder {
	builder := logBuilderPool.Get().(*Builder)
	builder.time = time.Now()
	builder.requestID = ctxs.RequestID(ctx)
	builder.requestIP = ctxs.RequestIP(ctx)
	userInfo := ctxs.User(ctx)
	if userInfo != nil {
		builder.userID = userInfo.ID
	}
	apiAccountInfo := ctxs.APIAccount(ctx)
	if apiAccountInfo != nil {
		builder.apiAccountID = apiAccountInfo.ID
	}
	appInfo := ctxs.App(ctx)
	if appInfo != nil {
		builder.appID = appInfo.ID
	}
	builder.requestURI = ctxs.RequestURI(ctx)
	builder.args = args
	_, file, line, ok := runtime.Caller(numberOfSkipCaller)
	if ok {
		index := strings.Index(file, consts.SystemName)
		if index >= 0 {
			file = strings.TrimLeft(file[index+len(consts.SystemName):], "/")

			// 再去除其中的版本号。
			hits := findVersionRegexp.FindStringSubmatch(file)
			if len(hits) >= 2 {
				file = strings.Replace(file, hits[1], "", 1)
			}

			builder.caller = fmt.Sprintf("%s:%d", file, line)
		}
	}
	return builder
}

// CreateBuilderf 构建新的日志格式化器。
func CreateBuilderf(ctx context.Context, format string, args ...any) *Builder {
	builder := logBuilderPool.Get().(*Builder)
	builder.time = time.Now()
	builder.requestID = ctxs.RequestID(ctx)
	builder.requestIP = ctxs.RequestIP(ctx)
	userInfo := ctxs.User(ctx)
	if userInfo != nil {
		builder.userID = userInfo.ID
	}
	apiAccountInfo := ctxs.APIAccount(ctx)
	if apiAccountInfo != nil {
		builder.apiAccountID = apiAccountInfo.ID
	}
	appInfo := ctxs.App(ctx)
	if appInfo != nil {
		builder.appID = appInfo.ID
	}
	builder.requestURI = ctxs.RequestURI(ctx)
	builder.args = args
	builder.format = format
	_, file, line, ok := runtime.Caller(numberOfSkipCaller)
	if ok {
		index := strings.Index(file, consts.SystemName)
		if index >= 0 {
			file = strings.TrimLeft(file[index+len(consts.SystemName):], "/")

			// 再去除其中的版本号。
			hits := findVersionRegexp.FindStringSubmatch(file)
			if len(hits) >= 2 {
				file = strings.Replace(file, hits[1], "", 1)
			}

			builder.caller = fmt.Sprintf("%s:%d", file, line)
		}
	}
	return builder
}

// SetLibrary 设置日志打印方。
func (builder *Builder) SetLibrary(library string) *Builder {
	builder.library = library
	return builder
}

// SetLevel 设置日志打印等级。
func (builder *Builder) SetLevel(level log.Level) *Builder {
	builder.level = level
	return builder
}

// SetCaller 设置日志打印调用点。
func (builder *Builder) SetCaller(caller string) *Builder {
	if len(caller) > 0 {
		builder.caller = caller
	}
	return builder
}

// BuildAndReclaim 构建日志并回收。
func (builder *Builder) BuildAndReclaim() string {
	defer builder.reclaim()
	return builder.build(false)
}

// BuildWithColorAndReclaim 构建彩色日志并回收。
func (builder *Builder) BuildWithColorAndReclaim() string {
	defer builder.reclaim()
	return builder.build(true)
}

// BuildTwiceAndReclaim 构建日志并回收。
func (builder *Builder) BuildTwiceAndReclaim() (log string, withColorlog string) {
	defer builder.reclaim()
	return builder.build(false), builder.build(true)
}

// 回收。
func (builder *Builder) reclaim() {
	if builder == nil {
		return
	}
	builder.time = time.Time{}
	builder.level = 0
	builder.requestID = ""
	builder.requestURI = ""
	builder.caller = ""
	builder.userID = 0
	builder.appID = 0
	builder.apiAccountID = 0
	builder.requestIP = ""
	builder.library = ""
	builder.args = nil
	builder.format = ""
	logBuilderPool.Put(builder)
}

// 生成字符串。
func (builder *Builder) build(colorful bool) string {
	elems := make([]string, 0, 12)
	if !builder.time.IsZero() {
		elems = append(elems, builder.time.Format(timeFormat))
	}
	if colorful {
		switch builder.level {
		case log.LevelDebug:
			elems = append(elems, builder.wrapperColor(builder.level.String(), black))
		case log.LevelInfo:
			elems = append(elems, builder.wrapperColor(builder.level.String(), green))
		case log.LevelWarn:
			elems = append(elems, builder.wrapperColor(builder.level.String(), yellow))
		case log.LevelError:
			elems = append(elems, builder.wrapperColor(builder.level.String(), read))
		case log.LevelFatal:
			elems = append(elems, builder.wrapperColor(builder.level.String(), magenta))
		}
	} else {
		elems = append(elems, builder.level.String())
	}
	if len(builder.requestID) > 0 {
		if colorful {
			elems = append(elems, builder.wrapperColor(builder.requestID, blue))
		} else {
			elems = append(elems, builder.requestID)
		}
	}
	if len(builder.requestURI) > 0 {
		elems = append(elems, builder.requestURI)
	}
	if len(builder.caller) > 0 {
		elems = append(elems, builder.caller)
	}
	if builder.userID > 0 {
		if colorful {
			elems = append(elems, builder.wrapperColor(strconv.Itoa(builder.userID), cyan))
		} else {
			elems = append(elems, strconv.Itoa(builder.userID))
		}
	}
	if builder.appID > 0 {
		elems = append(elems, strconv.Itoa(builder.appID))
	}
	if builder.apiAccountID > 0 {
		elems = append(elems, strconv.Itoa(builder.apiAccountID))
	}
	if len(builder.requestIP) > 0 {
		elems = append(elems, builder.requestIP)
	}
	if len(builder.library) > 0 {
		elems = append(elems, fmt.Sprintf("[%s]", builder.library))
	}
	if len(builder.format) > 0 {
		elems = append(elems, "--", builder.toOneLine(fmt.Sprintf(builder.format, builder.args...)))
	} else {
		elems = append(elems, "--", builder.toOneLine(builder.stringifyArgs(builder.args...)))
	}
	return strings.Join(elems, " ")
}

// 包裹颜色代码。
func (builder *Builder) wrapperColor(s string, colorCode int) string {
	return fmt.Sprintf("\033[1;%dm%s\033[0m", colorCode, s)
}

// 字符串化。
func (builder *Builder) stringifyArgs(args ...any) string {
	sb := strings.Builder{}
	for i := 0; i < len(args)-1; i++ {
		_, _ = fmt.Fprintf(&sb, "%s ", builder.stringify(args[i]))
	}
	if len(args) > 0 {
		_, _ = fmt.Fprintf(&sb, "%v", builder.stringify(args[len(args)-1]))
	}
	return sb.String()
}

// 整理成一行。
func (builder *Builder) toOneLine(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(
		strings.ReplaceAll(s, "\r\n", `\r\n`),
		"\n", `\n`,
	), "\r", `\r`) + "\n"
}

// 字符串化。
func (builder *Builder) stringify(v any) string {
	switch value := v.(type) {
	case int:
		return strconv.FormatInt(int64(value), 10)
	case int8:
		return strconv.FormatInt(int64(value), 10)
	case int16:
		return strconv.FormatInt(int64(value), 10)
	case int32:
		return strconv.FormatInt(int64(value), 10)
	case int64:
		return strconv.FormatInt(value, 10)
	case uintptr:
		return strconv.FormatUint(uint64(value), 10)
	case uint:
		return strconv.FormatUint(uint64(value), 10)
	case uint8:
		return strconv.FormatUint(uint64(value), 10)
	case uint16:
		return strconv.FormatUint(uint64(value), 10)
	case uint32:
		return strconv.FormatUint(uint64(value), 10)
	case uint64:
		return strconv.FormatUint(value, 10)
	case float32:
		return strconv.FormatFloat(float64(value), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case complex64:
		return fmt.Sprintf("%f+%fi", real(value), imag(value))
	case complex128:
		return fmt.Sprintf("%f+%fi", real(value), imag(value))
	case bool:
		return strconv.FormatBool(value)
	case string:
		return value
	case []byte:
		return string(value)
	case []rune:
		return string(value)
	case json.RawMessage:
		return string(value)
	default:
		return fmt.Sprintf("%+v", value)
	}
}
