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

	amqp "github.com/rabbitmq/amqp091-go"

	"gitee.com/ivfzhou/csms/comm/log/internal"
)

type rabbitMQLogger struct{}

// 新建实例。
func newRabbitMQLogger() amqp.Logging {
	return &rabbitMQLogger{}
}

func (l *rabbitMQLogger) Printf(format string, args ...any) {
	internal.SendBuilderToWriter(internal.CreateBuilderf(context.TODO(), format, args...).SetLibrary("rabbitmq"))
}
