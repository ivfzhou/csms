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
	"io"

	"github.com/ivfzhou/cron/v3"
	amqp "github.com/rabbitmq/amqp091-go"
	gorm "gorm.io/gorm/logger"

	"gitee.com/ivfzhou/csms/comm/log"
	tus "gitee.com/ivfzhou/tus_client/v2"
)

type (
	loggerWrapper struct {
		log.Logger
	}
	gormLoggerWrapper struct {
		gorm.Interface
	}
	cronLoggerWrapper struct {
		cron.Logger
	}
	redisLoggerWrapper struct {
		log.RedisLogger
	}
	rabbitmqLoggerWrapper struct {
		amqp.Logging
	}
	tusdLoggerWrapper struct {
		tus.Logger
	}
	ginLoggerWrapper struct {
		io.Writer
	}
)

var (
	level          log.Level
	logger         = &loggerWrapper{log.GetLogger()}
	gormLogger     = &gormLoggerWrapper{log.GetGormLogger()}
	cronLogger     = &cronLoggerWrapper{log.GetCronLogger()}
	redisLogger    = &redisLoggerWrapper{log.GetRedisLogger()}
	rabbitmqLogger = &rabbitmqLoggerWrapper{log.GetRabbitMQLogger()}
	tusdLogger     = &tusdLoggerWrapper{log.GetTusdLogger()}
	ginLogger      = &ginLoggerWrapper{log.GetGinLogger()}
)

// SetLevel 设置日志级别。
func SetLevel(l log.Level) {
	level = l
}

// GetLevel 获取日志级别。
func GetLevel() log.Level {
	return level
}

// GetLogger 获取日志打印对象。
func GetLogger() log.Logger {
	return logger
}

// SetLogger 设置日志实现。
func SetLogger(l log.Logger) {
	if l == nil {
		return
	}
	logger.Logger = l
}

// GetGormLogger 获取日志打印对象。
func GetGormLogger() gorm.Interface {
	return gormLogger
}

// SetGormLogger 设置日志实现。
func SetGormLogger(l gorm.Interface) {
	if l == nil {
		return
	}
	gormLogger.Interface = l
}

// GetRedisLogger 获取日志打印对象。
func GetRedisLogger() log.RedisLogger {
	return redisLogger
}

// SetRedisLogger 设置日志实现。
func SetRedisLogger(l log.RedisLogger) {
	if l == nil {
		return
	}
	redisLogger.RedisLogger = l
}

// GetCronLogger 获取日志打印对象。
func GetCronLogger() cron.Logger {
	return cronLogger
}

// SetCronLogger 设置日志实现。
func SetCronLogger(l cron.Logger) {
	if l == nil {
		return
	}
	cronLogger.Logger = l
}

// GetGinLogger 获取日志打印对象。
func GetGinLogger() io.Writer {
	return ginLogger
}

// SetGinLogger 设置日志实现。
func SetGinLogger(l io.Writer) {
	if l == nil {
		return
	}
	ginLogger.Writer = l
}

// GetRabbitMQLogger 获取日志打印对象。
func GetRabbitMQLogger() amqp.Logging {
	return rabbitmqLogger
}

// SetRabbitMQLogger 设置日志实现。
func SetRabbitMQLogger(l amqp.Logging) {
	if l == nil {
		return
	}
	rabbitmqLogger.Logging = l
}

// GetTusdLogger 获取日志打印对象。
func GetTusdLogger() tus.Logger {
	return tusdLogger
}

// SetTusdLogger 设置日志实现。
func SetTusdLogger(l tus.Logger) {
	if l == nil {
		return
	}
	tusdLogger.Logger = l
}
