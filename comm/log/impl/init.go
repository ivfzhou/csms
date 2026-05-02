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
	"sync"

	"gitee.com/ivfzhou/csms/comm/cfg"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/log/internal"
)

var (
	initializedOnce           sync.Once
	initializedFileLogOnce    sync.Once
	initializedConsoleLogOnce sync.Once
	initializedReportLogOnce  sync.Once
)

// Initialize 初始化日志打印。
func Initialize(_ context.Context) {
	initializedOnce.Do(func() {
		level := log.ParseLevel(cfg.Get().Log().Level())
		if level >= log.LevelDebug && level <= log.LevelFatal {
			internal.SetLevel(level)
		}

		cfg.RegisterNotifier(func(c cfg.Configurer) {
			level = log.ParseLevel(c.Log().Level())
			if level != internal.GetLevel() && level >= log.LevelDebug && level <= log.LevelFatal {
				ctx := ctxs.New()
				log.Warn(ctx, "update logger level", level.String())
				internal.SetLevel(level)
			}
		})

		internal.SetLogger(newLogger())
		internal.SetGormLogger(newGormLogger())
		internal.SetCronLogger(newCronLogger())
		internal.SetRedisLogger(newRedisLogger())
		internal.SetTusdLogger(newTusdLogger())
		internal.SetGinLogger(newGinLogger())
		internal.SetRabbitMQLogger(newRabbitMQLogger())

		log.RegisterImplement(
			internal.GetLogger,
			internal.CloseWriter,
			internal.GetLevel,
			internal.GetGormLogger,
			internal.GetCronLogger,
			internal.GetRedisLogger,
			internal.GetRabbitMQLogger,
			internal.GetTusdLogger,
			internal.GetGinLogger,
		)
	})
}

// InitializeConsoleLog 初始化控制台日志打印。
func InitializeConsoleLog(ctx context.Context) {
	initializedConsoleLogOnce.Do(func() {
		Initialize(ctx)
		internal.RegisterWriter(newConsoleWriter())
	})
}

// InitializeFileLog 初始化日志文件打印。
func InitializeFileLog(ctx context.Context) {
	initializedFileLogOnce.Do(func() {
		Initialize(ctx)
		internal.RegisterWriter(newFileWriter())
	})
}

// InitializeReportLog 日志上报。
func InitializeReportLog(ctx context.Context) {
	initializedReportLogOnce.Do(func() {
		Initialize(ctx)
		internal.RegisterWriter(newReportWriter())
	})
}
