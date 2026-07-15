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
	"io"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/ivfzhou/cron/v3"
	amqp "github.com/rabbitmq/amqp091-go"
	gorm "gorm.io/gorm/logger"

	"gitee.com/ivfzhou/csms/comm/cfg"
	tus "gitee.com/ivfzhou/tus_client/v2"
)

var (
	initializedFlag       atomic.Int32
	closeFunc             = func(context.Context) {}
	getLoggerFunc         = func() Logger { return defaultLoggerImpl }
	getLevelFunc          = func() Level { return ParseLevel(cfg.Get().Log().Level()) }
	getGormLoggerFunc     = func() gorm.Interface { return gorm.Default }
	getCronLoggerFunc     = func() cron.Logger { return cron.DefaultLogger }
	getRedisLoggerFunc    = func() RedisLogger { return defaultRedisLoggerImpl }
	getRabbitMQLoggerFunc = func() amqp.Logging { return amqp.Logger }
	getTusdLoggerFunc     = func() tus.Logger { return defaultTusdLoggerImpl }
	getGinLoggerFunc      = func() io.Writer { return gin.DefaultWriter }
)

// RegisterImplement 设置实现者。
func RegisterImplement(
	getLogger func() Logger,
	close func(context.Context),
	getLevel func() Level,
	getGormLogger func() gorm.Interface,
	getCronLogger func() cron.Logger,
	getRedisLogger func() RedisLogger,
	getRabbitMQLogger func() amqp.Logging,
	getTusdLogger func() tus.Logger,
	getGinLogger func() io.Writer,
) {
	if getLogger == nil || closeFunc == nil || getLevel == nil || getGormLogger == nil || getCronLogger == nil ||
		getRedisLogger == nil || getRabbitMQLogger == nil || getTusdLogger == nil || getGinLogger == nil {
		panic("nil value is not allowed")
	}

	if !initializedFlag.CompareAndSwap(0, 1) {
		panic("function has already been called")
	}

	getLoggerFunc = getLogger
	closeFunc = close
	getLevelFunc = getLevel
	getGormLoggerFunc = getGormLogger
	getCronLoggerFunc = getCronLogger
	getRedisLoggerFunc = getRedisLogger
	getRabbitMQLoggerFunc = getRabbitMQLogger
	getTusdLoggerFunc = getTusdLogger
	getGinLoggerFunc = getGinLogger
}

// Debug 打印日志。
func Debug(ctx context.Context, args ...any) {
	GetLogger().Debug(ctx, args...)
}

// Debugf 打印日志。
func Debugf(ctx context.Context, format string, args ...any) {
	GetLogger().Debugf(ctx, format, args...)
}

// Info 打印日志。
func Info(ctx context.Context, args ...any) {
	GetLogger().Info(ctx, args...)
}

// Infof 打印日志。
func Infof(ctx context.Context, format string, args ...any) {
	GetLogger().Infof(ctx, format, args...)
}

// Warn 打印日志。
func Warn(ctx context.Context, args ...any) {
	GetLogger().Warn(ctx, args...)
}

// Warnf 打印日志。
func Warnf(ctx context.Context, format string, args ...any) {
	GetLogger().Warnf(ctx, format, args...)
}

// Error 打印日志。
func Error(ctx context.Context, args ...any) {
	GetLogger().Error(ctx, args...)
}

// Errorf 打印日志。
func Errorf(ctx context.Context, format string, args ...any) {
	GetLogger().Errorf(ctx, format, args...)
}

// ErrorIf 打印日志。
func ErrorIf(ctx context.Context, err error, msg string, args ...any) {
	GetLogger().ErrorIf(ctx, err, msg, args...)
}

// ErrorIff 打印日志。
func ErrorIff(ctx context.Context, err error, format string, args ...any) {
	GetLogger().ErrorIff(ctx, err, format, args...)
}

// Fatal 打印日志。
func Fatal(ctx context.Context, exitCode int, args ...any) {
	GetLogger().Fatal(ctx, exitCode, args...)
}

// Fatalf 打印日志。
func Fatalf(ctx context.Context, exitCode int, format string, args ...any) {
	GetLogger().Fatalf(ctx, exitCode, format, args...)
}

// FatalIf 打印日志。
func FatalIf(ctx context.Context, exitCode int, err error, msg string, args ...any) {
	GetLogger().FatalIf(ctx, exitCode, err, msg, args...)
}

// FatalIff 打印日志。
func FatalIff(ctx context.Context, exitCode int, err error, format string, args ...any) {
	GetLogger().FatalIff(ctx, exitCode, err, format, args...)
}

// Close 关闭日志打印。
func Close(ctx context.Context) {
	closeFunc(ctx)
}

// GetLevel 获取日志等级。
func GetLevel() Level {
	return getLevelFunc()
}

// GetLogger 获取日志打印对象。
func GetLogger() Logger {
	return getLoggerFunc()
}

// GetGormLogger 获取 GORM 日志打印对象。
func GetGormLogger() gorm.Interface {
	return getGormLoggerFunc()
}

// GetCronLogger 获取 cron 日志打印对象。
func GetCronLogger() cron.Logger {
	return getCronLoggerFunc()
}

// GetRedisLogger 获取 Redis 日志打印对象。
func GetRedisLogger() RedisLogger {
	return getRedisLoggerFunc()
}

// GetRabbitMQLogger 获取 RabbitMQ 日志打印对象。
func GetRabbitMQLogger() amqp.Logging {
	return getRabbitMQLoggerFunc()
}

// GetTusdLogger 获取 Tusd 日志打印对象。
func GetTusdLogger() tus.Logger {
	return getTusdLoggerFunc()
}

// GetGinLogger 获取 Gin 日志打印对象。
func GetGinLogger() io.Writer {
	return getGinLoggerFunc()
}
