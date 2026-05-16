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
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	gorm "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"

	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/consts"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/log/internal"
	"gitee.com/ivfzhou/csms/comm/util"
)

var (
	noMatchCallerRegexp = regexp.MustCompile(`(?:(?:^comm(@v.*)/model/t_.*$)|(?:^comm(@v.*)/query/t_.*$)|(?:^comm(@v.*)/query/gen.go$)|(?:^comm/model/t_.*$)|(?:^comm/query/t_.*$)|(?:^comm/query/gen.go$))`)
	findVersionRegexp   = regexp.MustCompile(`^.*?(@v.*?)/.*$`)
)

type gormLoggerImpl struct {
	SlowSQLThreshold time.Duration
}

// 新建实例。
func newGormLogger() gorm.Interface {
	l := &gormLoggerImpl{SlowSQLThreshold: cfg.Get().Log().SlowSQLThreshold()}
	cfg.RegisterNotifier(func(c cfg.Configurer) {
		slowSQLThreshold := c.Log().SlowSQLThreshold()
		if l.SlowSQLThreshold == slowSQLThreshold {
			return
		}

		ctx := ctxs.New()
		log.Warn(ctx, "update gorm log slow sql threshold", slowSQLThreshold)
		l.SlowSQLThreshold = slowSQLThreshold
	})
	return l
}

func (l *gormLoggerImpl) LogMode(gorm.LogLevel) gorm.Interface {
	return l
}

func (l *gormLoggerImpl) Info(ctx context.Context, format string, args ...any) {
	if internal.GetLevel() <= log.LevelInfo {
		caller := utils.FileWithLineNum()
		firstIndex := strings.Index(caller, consts.SystemName)
		if firstIndex != -1 {
			caller = caller[firstIndex+len(consts.SystemName)+1:]
		}
		internal.SendBuilderToWriter(internal.CreateBuilderf(ctx, format, args...).SetLevel(log.LevelInfo).
			SetCaller(caller).SetLibrary("gorm"))
	}
}

func (l *gormLoggerImpl) Warn(ctx context.Context, format string, args ...any) {
	if internal.GetLevel() <= log.LevelWarn {
		caller := utils.FileWithLineNum()
		firstIndex := strings.Index(caller, consts.SystemName)
		if firstIndex != -1 {
			caller = caller[firstIndex+len(consts.SystemName)+1:]
		}
		internal.SendBuilderToWriter(internal.CreateBuilderf(ctx, format, args...).SetCaller(caller).
			SetLevel(log.LevelWarn).SetLibrary("gorm"))
	}
}

func (l *gormLoggerImpl) Error(ctx context.Context, format string, args ...any) {
	if internal.GetLevel() <= log.LevelError {
		caller := utils.FileWithLineNum()
		firstIndex := strings.Index(caller, consts.SystemName)
		if firstIndex != -1 {
			caller = caller[firstIndex+len(consts.SystemName)+1:]
		}
		internal.SendBuilderToWriter(internal.CreateBuilderf(ctx, format, args...).SetCaller(caller).
			SetLevel(log.LevelError).SetLibrary("gorm"))
	}
}

func (l *gormLoggerImpl) Trace(ctx context.Context, begin time.Time, fc func() (
	sql string, rowsAffected int64), err error) {

	elapsed := time.Since(begin)

	// 寻找 sql 触发代码位置。
	caller := ""
	callers := util.GetStackCallers()
	for len(callers) > 0 {
		caller = callers[0]
		callers = callers[1:]
		if !noMatchCallerRegexp.MatchString(caller) {
			hits := findVersionRegexp.FindStringSubmatch(caller)
			if len(hits) >= 2 {
				caller = strings.Replace(caller, hits[1], "", 1)
			}
			break
		}
	}
	if len(callers) <= 0 {
		caller = utils.FileWithLineNum()
	}

	var builder *internal.Builder
	switch {
	case err != nil && internal.GetLevel() <= log.LevelError && !errors.Is(err, gorm.ErrRecordNotFound):
		sql, rows := fc()
		msg := ""
		if rows == -1 {
			msg = fmt.Sprintf("%s [%.3fms] %s", err, float64(elapsed.Nanoseconds())/1e6, sql)
		} else {
			msg = fmt.Sprintf("%s [%.3fms] [rows:%v] %s", err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
		builder = internal.CreateBuilder(ctx, msg).SetCaller(caller).SetLibrary("gorm").SetLevel(log.LevelError)
	case elapsed > l.SlowSQLThreshold && l.SlowSQLThreshold != 0 &&
		internal.GetLevel() <= log.LevelWarn:
		sql, rows := fc()
		msg := ""
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowSQLThreshold)
		if rows == -1 {
			msg = fmt.Sprintf("%s [%.3fms] %s", slowLog, float64(elapsed.Nanoseconds())/1e6, sql)
		} else {
			msg = fmt.Sprintf("%s [%.3fms] [rows:%v] %s", slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
		builder = internal.CreateBuilder(ctx, msg).SetCaller(caller).SetLibrary("gorm").SetLevel(log.LevelWarn)
	case internal.GetLevel() <= log.LevelInfo:
		sql, rows := fc()
		msg := ""
		if rows == -1 {
			msg = fmt.Sprintf("[%.3fms] %s", float64(elapsed.Nanoseconds())/1e6, sql)
		} else {
			msg = fmt.Sprintf("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
		builder = internal.CreateBuilder(ctx, msg).SetCaller(caller).SetLibrary("gorm").SetLevel(log.LevelInfo)
	}
	if builder != nil {
		internal.SendBuilderToWriter(builder)
	}
}
